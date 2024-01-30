package logger

type Config struct {
	Debug           bool
	Pretty          bool
	SlackUsername   string `split_words:"true"`
	SlackWebhookURL string `split_words:"true"`
}
