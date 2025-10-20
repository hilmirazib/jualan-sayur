package config

import "github.com/spf13/viper"

type App struct {
	AppPort string `json:"app_port"`
	AppEnv  string `json:"app_env"`

	JwtSecretKey string `json:"jwt_secret_key"`
	JwtIssuer    string `json:"jwt_issuer"`
}

type PsqlDB struct {
	Host      string `json:"host"`
	Port      string `json:"port"`
	User      string `json:"user"`
	Password  string `json:"password"`
	DBName    string `json:"db_name"`
	DBMaxOpen int    `json:"db_max_open"`
	DBMaxIdle int    `json:"db_max_idle"`
}

type Supabase struct {
	ProjectURL string `json:"project_url"`
	APIKey     string `json:"api_key"`
	BucketName string `json:"bucket_name"`
}

type Config struct {
	App      App      `json:"app"`
	PsqlDB   PsqlDB   `json:"psql_db"`
	Redis    RedisConfig `json:"redis"`
	RabbitMQ RabbitMQ `json:"rabbitmq"`
	Supabase Supabase `json:"supabase"`
}

func NewConfig() *Config {
	viper.SetConfigFile(".env")
	if err := viper.ReadInConfig(); err != nil {
		panic(err)
	}

	return &Config{
		App: App{
			AppPort: viper.GetString("APP_PORT"),
			AppEnv:  viper.GetString("APP_ENV"),

			JwtSecretKey: viper.GetString("JWT_SECRET_KEY"),
			JwtIssuer:    viper.GetString("JWT_ISSUER"),
		},
		PsqlDB: PsqlDB{
			Host:      viper.GetString("DATABASE_HOST"),
			Port:      viper.GetString("DATABASE_PORT"),
			User:      viper.GetString("DATABASE_USER"),
			Password:  viper.GetString("DATABASE_PASSWORD"),
			DBName:    viper.GetString("DATABASE_NAME"),
			DBMaxOpen: viper.GetInt("DATABASE_MAX_OPEN_CONNECTION"),
			DBMaxIdle: viper.GetInt("DATABASE_MAX_IDLE_CONNECTION"),
		},
		Redis: RedisConfig{
			Host:     viper.GetString("REDIS_HOST"),
			Port:     viper.GetString("REDIS_PORT"),
			Password: viper.GetString("REDIS_PASSWORD"),
			DB:       viper.GetInt("REDIS_DB"),
		},
		RabbitMQ: RabbitMQ{
			Host:     viper.GetString("RABBITMQ_HOST"),
			Port:     viper.GetString("RABBITMQ_PORT"),
			User:     viper.GetString("RABBITMQ_USER"),
			Password: viper.GetString("RABBITMQ_PASSWORD"),
			VHost:    viper.GetString("RABBITMQ_VHOST"),
		},
		Supabase: Supabase{
			ProjectURL: viper.GetString("SUPABASE_PROJECT_URL"),
			APIKey:     viper.GetString("SUPABASE_API_KEY"),
			BucketName: viper.GetString("SUPABASE_BUCKET_NAME"),
		},
	}
}
