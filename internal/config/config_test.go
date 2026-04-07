package config

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

var testConfig = `jwt:
  signing_secret: fjd3252jrkal;f234fk
  admin:
    signing_secret: 32fj32l3;f032f09f32

logging:
  level: INFO
  request:
    queue_size: 150
    retention_days: 7
  audit:
    queue_size: 20
    retention_days: 180

server:
  port: "8080"

rate_limiting:
  backend: "redis" # or "memory"
  redis:
    url: localhost:6379

db:
  url: localhost
  port: "3306"
  username: root
  password: secret
  db_name: api_proxy`

func TestLoadConfig(t *testing.T) {
	scenarios := []struct {
		name           string
		config         string
		envVars        map[string]string
		expectedConfig Config
	}{
		{
			name:    "default",
			config:  testConfig,
			envVars: map[string]string{},
			expectedConfig: Config{
				JWTConfig: &JWTConfig{
					SigningSecret: "fjd3252jrkal;f234fk",
					Admin: &AdminJWTConfig{
						SigningSecret: "32fj32l3;f032f09f32",
					},
				},
				LoggingConfig: &LoggingConfig{
					Level: "INFO",
					LoggingRequestConfig: &LoggingRequestConfig{
						QueueSize:     new(150),
						RetentionDays: new(7),
					},
					LoggingAuditConfig: &LoggingAuditConfig{
						QueueSize:     new(20),
						RetentionDays: new(180),
					},
				},
				Server: &ServerConfig{
					Port: "8080",
				},
				RateLimitingConfig: &RateLimitingConfig{
					Backend: "redis",
					Redis: &RedisConfig{
						URL: "localhost:6379",
					},
				},
				DB: &DBConfig{
					URL:      "localhost",
					Port:     "3306",
					Username: "root",
					Password: "secret",
					DBName:   "api_proxy",
				},
			},
		},
		{
			name:   "env var overrides",
			config: testConfig,
			envVars: map[string]string{
				"LOG_REQUEST_RETENTION_DAYS": "10",
				"DB_USERNAME":                "injected",
			},
			expectedConfig: Config{
				JWTConfig: &JWTConfig{
					SigningSecret: "fjd3252jrkal;f234fk",
					Admin: &AdminJWTConfig{
						SigningSecret: "32fj32l3;f032f09f32",
					},
				},
				LoggingConfig: &LoggingConfig{
					Level: "INFO",
					LoggingRequestConfig: &LoggingRequestConfig{
						QueueSize:     new(150),
						RetentionDays: new(10),
					},
					LoggingAuditConfig: &LoggingAuditConfig{
						QueueSize:     new(20),
						RetentionDays: new(180),
					},
				},
				Server: &ServerConfig{
					Port: "8080",
				},
				RateLimitingConfig: &RateLimitingConfig{
					Backend: "redis",
					Redis: &RedisConfig{
						URL: "localhost:6379",
					},
				},
				DB: &DBConfig{
					URL:      "localhost",
					Port:     "3306",
					Username: "injected",
					Password: "secret",
					DBName:   "api_proxy",
				},
			},
		},
	}

	for _, scenario := range scenarios {
		t.Run(scenario.name, func(t *testing.T) {
			for key, value := range scenario.envVars {
				_ = os.Setenv(key, value)
			}

			testFile := filepath.Join(t.TempDir(), fmt.Sprintf("test_config_%s.yml", scenario.name))

			writeErr := os.WriteFile(testFile, []byte(scenario.config), 0644)

			if writeErr != nil {
				t.Fatal(writeErr)
			}

			actual, err := LoadConfig(testFile)

			if err != nil {
				t.Fatal(err)
			}

			if !reflect.DeepEqual(scenario.expectedConfig, *actual) {
				t.Fatalf("expected config did not match actual config")
			}

			// Cleanup
			for key := range scenario.envVars {
				_ = os.Unsetenv(key)
			}
		})
	}
}

func BenchmarkLoadConfig(b *testing.B) {
	testFile := filepath.Join(b.TempDir(), "test_config_benchmark.yml")
	writeErr := os.WriteFile(testFile, []byte(testConfig), 0644)

	if writeErr != nil {
		b.Fatal(writeErr)
	}

	b.ReportAllocs()
	b.ResetTimer()

	for b.Loop() {
		LoadConfig(testFile)
	}
}
