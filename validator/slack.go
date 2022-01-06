package validator

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"github.com/grafana/grafana/pkg/infra/log"
	"io"
	"net"
	"net/http"
	"time"
)

// sendSlackRequest sends a request to the Slack API.
// Stubbable by tests.
var SendSlackRequest = func(request *http.Request, logger log.Logger) error {
	netTransport := &http.Transport{
		TLSClientConfig: &tls.Config{
			Renegotiation: tls.RenegotiateFreelyAsClient,
		},
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout: 30 * time.Second,
		}).DialContext,
		TLSHandshakeTimeout: 5 * time.Second,
	}
	netClient := &http.Client{
		Timeout:   time.Second * 30,
		Transport: netTransport,
	}
	resp, err := netClient.Do(request)
	if err != nil {
		return err
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			logger.Warn("Failed to close response body", "err", err)
		}
	}()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		logger.Error("Slack API request failed", "url", request.URL.String(), "statusCode", resp.Status, "body", string(body))
		return fmt.Errorf("request to Slack API failed with status code %d", resp.StatusCode)
	}

	// Slack responds to some requests with a JSON document, that might contain an error.
	rslt := struct {
		Ok  bool   `json:"ok"`
		Err string `json:"error"`
	}{}

	// Marshaling can fail if Slack's response body is plain text (e.g. "ok").
	if err := json.Unmarshal(body, &rslt); err != nil && json.Valid(body) {
		logger.Error("Failed to unmarshal Slack API response", "url", request.URL.String(), "statusCode", resp.Status,
			"body", string(body))
		return fmt.Errorf("failed to unmarshal Slack API response: %s", err)
	}

	if !rslt.Ok && rslt.Err != "" {
		logger.Error("Sending Slack API request failed", "url", request.URL.String(), "statusCode", resp.Status,
			"err", rslt.Err)
		return fmt.Errorf("failed to make Slack API request: %s", rslt.Err)
	}

	logger.Debug("Sending Slack API request succeeded", "url", request.URL.String(), "statusCode", resp.Status)
	return nil
}
