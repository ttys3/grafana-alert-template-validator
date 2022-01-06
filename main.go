package main

import (
	"context"
	"github.com/grafana/grafana/pkg/components/simplejson"
	"github.com/grafana/grafana/pkg/infra/log"
	"github.com/grafana/grafana/pkg/services/ngalert/notifier/channels"
	"github.com/grafana/grafana/pkg/services/secrets"
	"github.com/grafana/grafana/pkg/services/secrets/fakes"
	secretsManager "github.com/grafana/grafana/pkg/services/secrets/manager"
	"os"
	"testing"

	"github.com/prometheus/alertmanager/notify"
	"github.com/prometheus/alertmanager/types"
	"github.com/prometheus/common/model"
	"github.com/ttys3/grafana-alert-template-validator/validator"
	"io"
	"net/http"
	"net/url"
)

func main() {
	slackWebhookURL := os.Getenv("SLACK_WEBHOOK_URL")

	if slackWebhookURL == "" {
		panic("you need set env var `SLACK_WEBHOOK_URL`")
	}

	tmpl := validator.TemplateForTests(validator.TemplateForTestsString)

	externalURL, err := url.Parse("http://localhost")
	validator.ErrPanic(err)

	tmpl.ExternalURL = externalURL

	cases := []struct {
		name     string
		settings string
		alerts   []*types.Alert
	}{{
		name: "Correct config with one alert",
		settings: `{
				"token": "1234",
				"recipient": "#testchannel",
				"icon_emoji": ":emoji:"
			}`,
		alerts: []*types.Alert{
			{
				Alert: model.Alert{
					Labels:      model.LabelSet{"alertname": "alert1", "lbl1": "val1"},
					Annotations: model.LabelSet{"ann1": "annv1", "__dashboardUid__": "abcd", "__panelId__": "efgh"},
				},
			},
		},
	},
		{
			name: "Correct config with multiple alerts and template",
			settings: `{
				"token": "1234",
				"recipient": "#testchannel",
				"icon_emoji": ":emoji:",
				"title": "🔥firing {{ .Alerts.Firing | len }}, ✅resolved {{ .Alerts.Resolved | len }}"
			}`,
			alerts: []*types.Alert{
				{
					Alert: model.Alert{
						Labels:      model.LabelSet{"alertname": "alert1", "lbl1": "val1"},
						Annotations: model.LabelSet{"ann1": "annv1"},
					},
				},
				{
					Alert: model.Alert{
						Labels:      model.LabelSet{"alertname": "alert2", "lbl1": "val2"},
						Annotations: model.LabelSet{"ann1": "annv2"},
					},
				},
			},
		},
	}

	t := &testing.T{}
	for _, c := range cases {
		settingsJSON, err := simplejson.NewJson([]byte(c.settings))
		validator.ErrPanic(err)

		secretsService := secretsManager.SetupTestService(t, fakes.NewFakeSecretsStore())

		secureSettings, err := secretsService.EncryptJsonData(context.TODO(), map[string]string{
			"url": slackWebhookURL,
		}, secrets.WithoutScope())
		validator.ErrPanic(err)

		m := &channels.NotificationChannelConfig{
			Name:           "slack_testing",
			Type:           "slack",
			Settings:       settingsJSON,
			SecureSettings: secureSettings,
		}

		decryptFn := secretsService.GetDecryptedValue
		pn, err := channels.NewSlackNotifier(m, tmpl, decryptFn)

		validator.ErrPanic(err)

		body := ""
		origSendSlackRequest := validator.SendSlackRequest
		defer func() {
			validator.SendSlackRequest = origSendSlackRequest
		}()
		validator.SendSlackRequest = func(request *http.Request, log log.Logger) error {
			defer func() {
				_ = request.Body.Close()
			}()

			b, err := io.ReadAll(request.Body)
			validator.ErrPanic(err)
			body = string(b)
			return nil
		}

		ctx := notify.WithGroupKey(context.Background(), "alertname")
		ctx = notify.WithGroupLabels(ctx, model.LabelSet{"alertname": ""})
		_, err = pn.Notify(ctx, c.alerts...)
		validator.ErrPanic(err)

	}
}
