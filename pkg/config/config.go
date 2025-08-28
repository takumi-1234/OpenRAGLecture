// OpenRAGLecture/pkg/config/config.go

package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

// Config stores all configuration of the application.
// The values are read by viper from a config file or environment variables.
type Config struct {
	Server    ServerConfig    `mapstructure:"server"`
	Database  DatabaseConfig  `mapstructure:"database"`
	VectorDB  VectorDBConfig  `mapstructure:"vector_db"`
	Cache     CacheConfig     `mapstructure:"cache"`
	Auth      AuthConfig      `mapstructure:"auth"`
	Storage   StorageConfig   `mapstructure:"storage"`
	Google    GoogleConfig    `mapstructure:"google"`
	Logging   LoggingConfig   `mapstructure:"logging"`
	Telemetry TelemetryConfig `mapstructure:"telemetry"`
}

type ServerConfig struct {
	Port                string `mapstructure:"port"`
	Mode                string `mapstructure:"mode"`
	ReadTimeoutSeconds  int    `mapstructure:"read_timeout_seconds"`
	WriteTimeoutSeconds int    `mapstructure:"write_timeout_seconds"`
}

type DatabaseConfig struct {
	MySQL MySQLConfig `mapstructure:"mysql"`
}

type MySQLConfig struct {
	User                   string `mapstructure:"user"`
	Password               string `mapstructure:"password"`
	Host                   string `mapstructure:"host"`
	Port                   string `mapstructure:"port"`
	DBName                 string `mapstructure:"dbname"`
	MaxOpenConns           int    `mapstructure:"max_open_conns"`
	MaxIdleConns           int    `mapstructure:"max_idle_conns"`
	ConnMaxLifetimeMinutes int    `mapstructure:"conn_max_lifetime_minutes"`
}

type VectorDBConfig struct {
	Qdrant QdrantConfig `mapstructure:"qdrant"`
}

type QdrantConfig struct {
	Host           string `mapstructure:"host"`
	Port           int    `mapstructure:"port"`
	GrpcPort       int    `mapstructure:"grpc_port"`
	APIKey         string `mapstructure:"api_key"`
	UseTLS         bool   `mapstructure:"use_tls"`
	CollectionName string `mapstructure:"collection_name"`
	VectorSize     uint64 `mapstructure:"vector_size"`
}

type CacheConfig struct {
	Redis RedisConfig `mapstructure:"redis"`
}

type RedisConfig struct {
	Host     string `mapstructure:"host"`
	Port     string `mapstructure:"port"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}

type AuthConfig struct {
	JWT JWTConfig `mapstructure:"jwt"`
}

type JWTConfig struct {
	SecretKey              string `mapstructure:"secret_key"`
	AccessTokenExpiryHours int    `mapstructure:"access_token_expiry_hours"`
	RefreshTokenExpiryDays int    `mapstructure:"refresh_token_expiry_days"`
}

type StorageConfig struct {
	Type  string             `mapstructure:"type"`
	Local LocalStorageConfig `mapstructure:"local"`
	S3    S3StorageConfig    `mapstructure:"s3"`
}

type LocalStorageConfig struct {
	Path string `mapstructure:"path"`
}

type S3StorageConfig struct {
	Bucket    string `mapstructure:"bucket"`
	Region    string `mapstructure:"region"`
	Endpoint  string `mapstructure:"endpoint"`
	AccessKey string `mapstructure:"access_key"`
	SecretKey string `mapstructure:"secret_key"`
	UseSSL    bool   `mapstructure:"use_ssl"`
}

type GoogleConfig struct {
	APIKey         string `mapstructure:"api_key"`
	ProjectID      string `mapstructure:"project_id"` // ★★★ 追加 ★★★
	Location       string `mapstructure:"location"`   // ★★★ 追加 ★★★
	EmbeddingModel string `mapstructure:"embedding_model"`
	LLMModel       string `mapstructure:"llm_model"`
}

type LoggingConfig struct {
	Level    string `mapstructure:"level"`
	Encoding string `mapstructure:"encoding"`
}

type TelemetryConfig struct {
	Enabled          bool   `mapstructure:"enabled"`
	ServiceName      string `mapstructure:"service_name"`
	ExporterEndpoint string `mapstructure:"exporter_endpoint"`
	Insecure         bool   `mapstructure:"insecure"`
}

// LoadConfig reads configuration from file or environment variables.
func LoadConfig(path string) (config Config, err error) {
	// Priority 1: Environment Variables
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// Priority 2: Default Config File (config.defaults.yaml)
	// This is the base configuration.
	viper.AddConfigPath(path)
	viper.SetConfigName("config.defaults")
	viper.SetConfigType("yaml")

	if err = viper.ReadInConfig(); err != nil {
		// The default config file must exist.
		return
	}

	// Priority 3: Environment-specific Config File (e.g., config.debug.yaml)
	// This will override the defaults. Environment variables will override this.
	env := viper.GetString("server.mode")
	if env == "" {
		// Default to "debug" if SERVER_MODE env var is not set.
		// We check viper again in case it was set in the defaults file.
		viper.SetDefault("server.mode", "debug")
		env = viper.GetString("server.mode")
	}

	if env != "" {
		viper.AddConfigPath(path)
		viper.SetConfigName("config." + env)
		viper.SetConfigType("yaml")
		// Merge the environment-specific config on top of the defaults.
		// Errors are ignored if the file doesn't exist.
		_ = viper.MergeInConfig()
	}

	// Unmarshal the final merged config into the struct
	// Viper has already handled the priority: Env Vars > Specific Config > Default Config
	err = viper.Unmarshal(&config)
	if err != nil {
		return
	}

	fmt.Printf("Configuration loaded for environment: %s\n", config.Server.Mode)
	return
}
