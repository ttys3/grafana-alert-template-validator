# grafana-alert-template-validator

## faq


The `define` tag in the **Content** section assigns **the template name**.

This tag is **optional**, and when omitted, the template name is derived from the **Name** field.

When **both** are specified, it is a best practice to ensure that they are the same.


## Refs

ref https://grafana.com/docs/grafana/latest/alerting/unified-alerting/message-templating/

default template https://github.com/grafana/grafana/blob/main/pkg/services/ngalert/notifier/channels/default_template.go
