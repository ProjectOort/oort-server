package conf

import "github.com/spf13/viper"

type App struct {
	Name     string   `mapstructure:"name"`
	Version  string   `mapstructure:"version"`
	Endpoint Endpoint `mapstructure:"endpoint"`
}

type Endpoint struct {
	HTTP HTTP `mapstructure:"http"`
}

type HTTP struct {
	URL string `mapstructure:"url"`
}

func Parse(path string) *App {
	var (
		v = viper.New()
		c = new(App)
	)

	v.SetConfigName("conf.yaml")
	v.SetConfigType("yaml")
	v.AddConfigPath(path)

	check(v.ReadInConfig())
	check(v.Unmarshal(c))
	return c
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}
