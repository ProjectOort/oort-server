package conf

import "github.com/spf13/viper"

type App struct {
	Name     string   `mapstructure:"name"`
	Version  string   `mapstructure:"version"`
	Endpoint Endpoint `mapstructure:"endpoint"`
	Logger   Logger   `mapstructure:"logger"`
	Repo     Repo     `mapstructure:"repo"`
	Biz      Biz      `mapstructure:"biz"`
}

type Endpoint struct {
	HTTP HTTP `mapstructure:"http"`
}

type HTTP struct {
	URL string `mapstructure:"url"`
}

type Logger struct {
	Level     string `mapstructure:"level"`
	Path      string `mapstructure:"path"`
	MaxSizeMB int    `mapstructure:"max_size_mb"`
	MaxAgeDay int    `mapstructure:"max_age_day"`
	Compress  bool   `mapstructure:"compress"`
}

type Repo struct {
	Mongo         Mongo         `mapstructure:"mongo"`
	Neo4j         Neo4j         `mapstructure:"neo4j"`
	Elasticsearch Elasticsearch `mapstructure:"elasticsearch"`
}

type Mongo struct {
	URL string `mapstructure:"url"`
}

type Neo4j struct {
	URL      string `mapstructure:"url"`
	Realm    string `mapstructure:"realm"`
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
}

type Elasticsearch struct {
	URL      string `mapstructure:"url"`
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
}

type Biz struct {
	Account Account `mapstructure:"account"`
}

type Account struct {
	TokenKey       string `mapstructure:"token_key"`
	TokenExpireSec int    `mapstructure:"token_expire_sec"`

	GiteeClientID     string `mapstructure:"gitee_client_id"`
	GiteeClientSecret string `mapstructure:"gitee_client_secret"`
	GiteeRedirectURI  string `mapstructure:"gitee_redirect_uri"`
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
