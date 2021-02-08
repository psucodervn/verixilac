package config

type APIConfig struct {
	ListenAddress string `split_words:"true" default:"0.0.0.0:80"`
}

type PostgresConfig struct {
	Host     string `required:"true" split_words:"true"`
	Port     int    `default:"5432" split_words:"true"`
	User     string `required:"true" split_words:"true"`
	Password string `required:"true" split_words:"true"`
	Database string `required:"true" split_words:"true"`
	SSLMode  string `split_words:"true" default:"disable"`
	Debug    bool   `default:"false" split_words:"true"`
}

type MigrationConfig struct {
	Postgres PostgresConfig `split_words:"true"`
}
