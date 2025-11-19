package config

import (
	"os"
	"testing"

	"github.com/gymshark/go-easy-config/loader/generic"
)

// BenchmarkAnalysisPhase benchmarks the analysis phase overhead
func BenchmarkAnalysisPhase(b *testing.B) {
	type Config struct {
		Env    string `env:"BENCH_ENV" config:"availableAs=ENV"`
		Region string `env:"BENCH_REGION" config:"availableAs=REGION"`
		DBHost string `env:"BENCH_DB_HOST"`
		APIKey string `env:"BENCH_API_KEY"`
	}

	os.Setenv("BENCH_ENV", "production")
	os.Setenv("BENCH_REGION", "us-east-1")
	os.Setenv("BENCH_DB_HOST", "db.example.com")
	os.Setenv("BENCH_API_KEY", "key-12345")
	defer func() {
		os.Unsetenv("BENCH_ENV")
		os.Unsetenv("BENCH_REGION")
		os.Unsetenv("BENCH_DB_HOST")
		os.Unsetenv("BENCH_API_KEY")
	}()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		engine := NewInterpolationEngine[Config]()
		var cfg Config
		_ = engine.Analyze(&cfg)
	}
}

// BenchmarkInterpolation benchmarks the interpolation performance
func BenchmarkInterpolation(b *testing.B) {
	type Config struct {
		Env    string `env:"INTERP_ENV" config:"availableAs=ENV"`
		Region string `env:"INTERP_REGION" config:"availableAs=REGION"`
		DBHost string `env:"INTERP_DB_HOST"`
		APIKey string `env:"INTERP_API_KEY"`
	}

	os.Setenv("INTERP_ENV", "production")
	os.Setenv("INTERP_REGION", "us-east-1")
	os.Setenv("INTERP_DB_HOST", "db.example.com")
	os.Setenv("INTERP_API_KEY", "key-12345")
	defer func() {
		os.Unsetenv("INTERP_ENV")
		os.Unsetenv("INTERP_REGION")
		os.Unsetenv("INTERP_DB_HOST")
		os.Unsetenv("INTERP_API_KEY")
	}()

	engine := NewInterpolationEngine[Config]()
	var cfg Config
	_ = engine.Analyze(&cfg)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		stages := engine.GetDependencyStages()
		for _, stage := range stages {
			_ = engine.InterpolateTags(stage)
		}
	}
}

// BenchmarkWithInterpolation benchmarks loading with interpolation infrastructure
func BenchmarkWithInterpolation(b *testing.B) {
	type Config struct {
		Env    string `env:"WITH_ENV" config:"availableAs=ENV"`
		Region string `env:"WITH_REGION" config:"availableAs=REGION"`
		DBHost string `env:"WITH_DB_HOST"`
		APIKey string `env:"WITH_API_KEY"`
		Port   int    `env:"WITH_PORT"`
	}

	os.Setenv("WITH_ENV", "production")
	os.Setenv("WITH_REGION", "us-east-1")
	os.Setenv("WITH_DB_HOST", "db.example.com")
	os.Setenv("WITH_API_KEY", "key-12345")
	os.Setenv("WITH_PORT", "8080")
	defer func() {
		os.Unsetenv("WITH_ENV")
		os.Unsetenv("WITH_REGION")
		os.Unsetenv("WITH_DB_HOST")
		os.Unsetenv("WITH_API_KEY")
		os.Unsetenv("WITH_PORT")
	}()

	handler := NewConfigHandler[Config](
		WithLoaders[Config](&generic.EnvironmentLoader[Config]{}),
	)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var cfg Config
		_ = handler.Load(&cfg)
	}
}

// BenchmarkWithoutInterpolation benchmarks loading without interpolation (fast path)
func BenchmarkWithoutInterpolation(b *testing.B) {
	type Config struct {
		Env    string `env:"WITHOUT_ENV"`
		Region string `env:"WITHOUT_REGION"`
		DBHost string `env:"WITHOUT_DB_HOST"`
		APIKey string `env:"WITHOUT_API_KEY"`
		Port   int    `env:"WITHOUT_PORT"`
	}

	os.Setenv("WITHOUT_ENV", "production")
	os.Setenv("WITHOUT_REGION", "us-east-1")
	os.Setenv("WITHOUT_DB_HOST", "db.example.com")
	os.Setenv("WITHOUT_API_KEY", "key-12345")
	os.Setenv("WITHOUT_PORT", "8080")
	defer func() {
		os.Unsetenv("WITHOUT_ENV")
		os.Unsetenv("WITHOUT_REGION")
		os.Unsetenv("WITHOUT_DB_HOST")
		os.Unsetenv("WITHOUT_API_KEY")
		os.Unsetenv("WITHOUT_PORT")
	}()

	handler := NewConfigHandler[Config](
		WithLoaders[Config](&generic.EnvironmentLoader[Config]{}),
	)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var cfg Config
		_ = handler.Load(&cfg)
	}
}

// BenchmarkTypeConversion benchmarks type conversion overhead
func BenchmarkTypeConversion(b *testing.B) {
	type Config struct {
		IntValue    int     `env:"TYPE_INT" config:"availableAs=INT"`
		FloatValue  float64 `env:"TYPE_FLOAT" config:"availableAs=FLOAT"`
		BoolValue   bool    `env:"TYPE_BOOL" config:"availableAs=BOOL"`
		StringValue string  `env:"TYPE_STRING" config:"availableAs=STRING"`
	}

	os.Setenv("TYPE_INT", "12345")
	os.Setenv("TYPE_FLOAT", "123.45")
	os.Setenv("TYPE_BOOL", "true")
	os.Setenv("TYPE_STRING", "test-value")
	defer func() {
		os.Unsetenv("TYPE_INT")
		os.Unsetenv("TYPE_FLOAT")
		os.Unsetenv("TYPE_BOOL")
		os.Unsetenv("TYPE_STRING")
	}()

	handler := NewConfigHandler[Config](
		WithLoaders[Config](&generic.EnvironmentLoader[Config]{}),
	)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var cfg Config
		_ = handler.Load(&cfg)
	}
}

// BenchmarkComplexConfig benchmarks a complex configuration with multiple stages
func BenchmarkComplexConfig(b *testing.B) {
	type Config struct {
		// Stage 0
		Env     string `env:"COMPLEX_ENV" config:"availableAs=ENV"`
		Region  string `env:"COMPLEX_REGION" config:"availableAs=REGION"`
		Cluster string `env:"COMPLEX_CLUSTER" config:"availableAs=CLUSTER"`

		// Stage 1
		Service string `env:"COMPLEX_SERVICE" config:"availableAs=SERVICE"`
		Version string `env:"COMPLEX_VERSION" config:"availableAs=VERSION"`

		// Regular fields
		DBHost   string `env:"COMPLEX_DB_HOST"`
		APIKey   string `env:"COMPLEX_API_KEY"`
		Port     int    `env:"COMPLEX_PORT"`
		Timeout  int    `env:"COMPLEX_TIMEOUT"`
		Debug    bool   `env:"COMPLEX_DEBUG"`
		LogLevel string `env:"COMPLEX_LOG_LEVEL"`
	}

	os.Setenv("COMPLEX_ENV", "production")
	os.Setenv("COMPLEX_REGION", "us-east-1")
	os.Setenv("COMPLEX_CLUSTER", "main")
	os.Setenv("COMPLEX_SERVICE", "api")
	os.Setenv("COMPLEX_VERSION", "v1.2.3")
	os.Setenv("COMPLEX_DB_HOST", "db.example.com")
	os.Setenv("COMPLEX_API_KEY", "key-12345")
	os.Setenv("COMPLEX_PORT", "8080")
	os.Setenv("COMPLEX_TIMEOUT", "30")
	os.Setenv("COMPLEX_DEBUG", "false")
	os.Setenv("COMPLEX_LOG_LEVEL", "info")
	defer func() {
		os.Unsetenv("COMPLEX_ENV")
		os.Unsetenv("COMPLEX_REGION")
		os.Unsetenv("COMPLEX_CLUSTER")
		os.Unsetenv("COMPLEX_SERVICE")
		os.Unsetenv("COMPLEX_VERSION")
		os.Unsetenv("COMPLEX_DB_HOST")
		os.Unsetenv("COMPLEX_API_KEY")
		os.Unsetenv("COMPLEX_PORT")
		os.Unsetenv("COMPLEX_TIMEOUT")
		os.Unsetenv("COMPLEX_DEBUG")
		os.Unsetenv("COMPLEX_LOG_LEVEL")
	}()

	handler := NewConfigHandler[Config](
		WithLoaders[Config](&generic.EnvironmentLoader[Config]{}),
	)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var cfg Config
		_ = handler.Load(&cfg)
	}
}

// BenchmarkMultipleLoaders benchmarks loading with multiple loaders
func BenchmarkMultipleLoaders(b *testing.B) {
	type Config struct {
		Env    string `env:"MULTI_ENV" clap:"--env" config:"availableAs=ENV"`
		Region string `env:"MULTI_REGION" clap:"--region" config:"availableAs=REGION"`
		DBHost string `env:"MULTI_DB_HOST" clap:"--db-host"`
		APIKey string `env:"MULTI_API_KEY" clap:"--api-key"`
	}

	os.Setenv("MULTI_ENV", "production")
	os.Setenv("MULTI_REGION", "us-east-1")
	os.Setenv("MULTI_DB_HOST", "db.example.com")
	os.Setenv("MULTI_API_KEY", "key-12345")
	defer func() {
		os.Unsetenv("MULTI_ENV")
		os.Unsetenv("MULTI_REGION")
		os.Unsetenv("MULTI_DB_HOST")
		os.Unsetenv("MULTI_API_KEY")
	}()

	args := []string{"--env", "staging", "--region", "us-west-2"}

	handler := NewConfigHandler[Config](
		WithLoaders[Config](
			&generic.EnvironmentLoader[Config]{},
			&generic.CommandLineLoader[Config]{Args: args},
		),
	)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var cfg Config
		_ = handler.Load(&cfg)
	}
}

// BenchmarkLoadAndValidate benchmarks the complete load and validate flow
func BenchmarkLoadAndValidate(b *testing.B) {
	type Config struct {
		Env    string `env:"LAV_ENV" config:"availableAs=ENV" validate:"required,oneof=dev staging prod"`
		Region string `env:"LAV_REGION" config:"availableAs=REGION" validate:"required"`
		DBHost string `env:"LAV_DB_HOST" validate:"required,hostname"`
		Port   int    `env:"LAV_PORT" validate:"required,min=1,max=65535"`
	}

	os.Setenv("LAV_ENV", "production")
	os.Setenv("LAV_REGION", "us-east-1")
	os.Setenv("LAV_DB_HOST", "db.example.com")
	os.Setenv("LAV_PORT", "8080")
	defer func() {
		os.Unsetenv("LAV_ENV")
		os.Unsetenv("LAV_REGION")
		os.Unsetenv("LAV_DB_HOST")
		os.Unsetenv("LAV_PORT")
	}()

	handler := NewConfigHandler[Config](
		WithLoaders[Config](&generic.EnvironmentLoader[Config]{}),
	)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var cfg Config
		_ = handler.LoadAndValidate(&cfg)
	}
}

// BenchmarkMemoryAllocations benchmarks memory allocations
func BenchmarkMemoryAllocations(b *testing.B) {
	type Config struct {
		Env    string `env:"MEM_ENV" config:"availableAs=ENV"`
		Region string `env:"MEM_REGION" config:"availableAs=REGION"`
		DBHost string `env:"MEM_DB_HOST"`
		APIKey string `env:"MEM_API_KEY"`
		Port   int    `env:"MEM_PORT"`
	}

	os.Setenv("MEM_ENV", "production")
	os.Setenv("MEM_REGION", "us-east-1")
	os.Setenv("MEM_DB_HOST", "db.example.com")
	os.Setenv("MEM_API_KEY", "key-12345")
	os.Setenv("MEM_PORT", "8080")
	defer func() {
		os.Unsetenv("MEM_ENV")
		os.Unsetenv("MEM_REGION")
		os.Unsetenv("MEM_DB_HOST")
		os.Unsetenv("MEM_API_KEY")
		os.Unsetenv("MEM_PORT")
	}()

	handler := NewConfigHandler[Config](
		WithLoaders[Config](&generic.EnvironmentLoader[Config]{}),
	)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var cfg Config
		_ = handler.Load(&cfg)
	}
}

// BenchmarkFastPathDetection benchmarks the fast path detection overhead
func BenchmarkFastPathDetection(b *testing.B) {
	type Config struct {
		Value1 string `env:"FAST_VALUE1"`
		Value2 int    `env:"FAST_VALUE2"`
		Value3 bool   `env:"FAST_VALUE3"`
		Value4 string `env:"FAST_VALUE4"`
		Value5 int    `env:"FAST_VALUE5"`
	}

	os.Setenv("FAST_VALUE1", "test1")
	os.Setenv("FAST_VALUE2", "42")
	os.Setenv("FAST_VALUE3", "true")
	os.Setenv("FAST_VALUE4", "test4")
	os.Setenv("FAST_VALUE5", "100")
	defer func() {
		os.Unsetenv("FAST_VALUE1")
		os.Unsetenv("FAST_VALUE2")
		os.Unsetenv("FAST_VALUE3")
		os.Unsetenv("FAST_VALUE4")
		os.Unsetenv("FAST_VALUE5")
	}()

	handler := NewConfigHandler[Config](
		WithLoaders[Config](&generic.EnvironmentLoader[Config]{}),
	)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var cfg Config
		_ = handler.Load(&cfg)
	}
}
