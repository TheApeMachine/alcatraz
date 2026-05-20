package config

import "github.com/theapemachine/errnie"

var errnieRootKey = "errnie"

/*
ErrnieConfig holds logging settings for the alcatraz process.
*/
type ErrnieConfig struct {
	Level         string
	DisableCaller bool
	File          FileConfig
	Elasticsearch ElasticsearchConfig
}

/*
FileConfig controls optional file logging output.
*/
type FileConfig struct {
	Active bool
	Path   string
}

/*
ElasticsearchConfig controls optional Elasticsearch log indexing.
*/
type ElasticsearchConfig struct {
	Active   bool
	URL      string
	Index    string
	Username string
	Password string
}

/*
NewErrnieConfig reads errnie settings from viper-loaded config.yml.
*/
func NewErrnieConfig() *ErrnieConfig {
	level := WithDefault(errnieRootKey+".level", "")

	if level == "" {
		level = WithDefault(errnieRootKey+".loglevel", "info")
	}

	fileRootKey := errnieRootKey + ".file"
	elasticsearchRootKey := errnieRootKey + ".elasticsearch"

	return &ErrnieConfig{
		Level:         level,
		DisableCaller: WithDefault(errnieRootKey+".disable_caller", false),
		File: FileConfig{
			Active: WithDefault(fileRootKey+".active", false),
			Path:   WithDefault(fileRootKey+".path", ""),
		},
		Elasticsearch: ElasticsearchConfig{
			Active:   WithDefault(elasticsearchRootKey+".active", false),
			URL:      WithDefault(elasticsearchRootKey+".url", ""),
			Index:    WithDefault(elasticsearchRootKey+".index", ""),
			Username: WithDefault(elasticsearchRootKey+".username", ""),
			Password: WithDefault(elasticsearchRootKey+".password", ""),
		},
	}
}

/*
ToLibraryConfig converts alcatraz config into the errnie library config shape.
*/
func (errnieConfig *ErrnieConfig) ToLibraryConfig() *errnie.Config {
	libraryConfig := &errnie.Config{
		Level:         errnieConfig.Level,
		DisableCaller: errnieConfig.DisableCaller,
	}
	libraryConfig.File.Active = errnieConfig.File.Active
	libraryConfig.File.Path = errnieConfig.File.Path
	libraryConfig.Elasticsearch.Active = errnieConfig.Elasticsearch.Active
	libraryConfig.Elasticsearch.URL = errnieConfig.Elasticsearch.URL
	libraryConfig.Elasticsearch.Index = errnieConfig.Elasticsearch.Index
	libraryConfig.Elasticsearch.Username = errnieConfig.Elasticsearch.Username
	libraryConfig.Elasticsearch.Password = errnieConfig.Elasticsearch.Password

	return libraryConfig
}
