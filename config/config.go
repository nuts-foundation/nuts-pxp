package config

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/knadh/koanf"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/providers/posflag"
	"github.com/spf13/pflag"
)

const defaultConfigFile = "./config.yaml"
const configFileFlag = "configfile"

const defaultEnvPrefix = "NUTS_"
const defaultEnvDelimiter = "_"
const defaultDelimiter = "."
const configValueListSeparator = ","

type Config struct {
	configMap *koanf.Koanf
	SQL       SQLConfig `koanf:"sql"`
}

type SQLConfig struct {
	// ConnectionString is the connection string for the SQL database.
	// This string may contain secrets (user:password), so should never be logged.
	ConnectionString string `koanf:"connection"`
}

func (c *Config) Load() error {
	c.configMap = koanf.New(defaultDelimiter)
	if err := c.loadConfigMap(flags()); err != nil {
		return err
	}

	if err := loadConfigIntoStruct(c, c.configMap); err != nil {
		return err
	}

	return nil
}

func (c *Config) loadConfigMap(flags *pflag.FlagSet) error {
	if err := loadFromFile(c.configMap, resolveConfigFilePath(flags)); err != nil {
		return err
	}

	if err := loadFromEnv(c.configMap); err != nil {
		return err
	}

	// Besides CLI, also sets default values for flags not yet set in the configMap.
	if err := loadFromFlagSet(c.configMap, flags); err != nil {
		return err
	}

	return nil
}

func flags() *pflag.FlagSet {
	flags := pflag.NewFlagSet("config", pflag.ExitOnError)
	flags.String("config", "", "config file")
	flags.String("sql.connection", "", "SQL connection string")
	return flags
}

// resolveConfigFilePath resolves the path of the config file using the following sources:
// 1. commandline params (using the given flags)
// 2. environment vars,
// 3. default location.
func resolveConfigFilePath(flags *pflag.FlagSet) string {
	k := koanf.New(defaultDelimiter)

	// load env flags
	e := env.Provider(defaultEnvPrefix, defaultDelimiter, func(s string) string {
		return strings.Replace(strings.ToLower(
			strings.TrimPrefix(s, defaultEnvPrefix)), defaultEnvDelimiter, defaultDelimiter, -1)
	})
	// can't return error
	_ = k.Load(e, nil)

	// load cmd flags, without a parser, no error can be returned
	// this also loads the default flag value of config.yaml. So we need a way to know if it's overiden.
	_ = k.Load(posflag.Provider(flags, defaultDelimiter, k), nil)

	envFilepath := k.String("configfile")
	if envFilepath == "" {
		return defaultConfigFile
	}

	return envFilepath
}

func loadConfigIntoStruct(target interface{}, configMap *koanf.Koanf) error {
	// load into struct
	return configMap.UnmarshalWithConf("", target, koanf.UnmarshalConf{
		FlatPaths: false,
	})
}

func loadFromFile(configMap *koanf.Koanf, filepath string) error {
	if filepath == "" {
		return nil
	}
	configFileProvider := file.Provider(filepath)
	// load file
	if err := configMap.Load(configFileProvider, yaml.Parser()); err != nil {
		// return all errors but ignore the missing of the default config file
		if !errors.Is(err, os.ErrNotExist) || filepath != defaultConfigFile {
			return fmt.Errorf("unable to load config file: %w", err)
		}
	}
	return nil
}

// loadFromEnv loads the values from the environment variables into the configMap
func loadFromEnv(configMap *koanf.Koanf) error {
	e := env.ProviderWithValue(defaultEnvPrefix, defaultDelimiter, func(rawKey string, rawValue string) (string, interface{}) {
		key := strings.Replace(strings.ToLower(strings.TrimPrefix(rawKey, defaultEnvPrefix)), defaultEnvDelimiter, defaultDelimiter, -1)

		// Support multiple values separated by a comma, but let them be escaped with a backslash
		values := splitWithEscaping(rawValue, configValueListSeparator, "\\")
		for i, value := range values {
			values[i] = strings.TrimSpace(value)
		}
		if len(values) == 1 {
			return key, values[0]
		}
		return key, values
	})
	return configMap.Load(e, nil)
}

// loadFromFlagSet loads the config values set in the command line options into the configMap.
// Also sets default value for all flags in the provided pflag.FlagSet if the values do not yet exist in the configMap.
func loadFromFlagSet(configMap *koanf.Koanf, flags *pflag.FlagSet) error {
	// error out if flag name ends with .token or .password (which indicates a secret) and is set on the command line
	var err error
	flags.VisitAll(func(flag *pflag.Flag) {
		if strings.HasSuffix(flag.Name, "token") || strings.HasSuffix(flag.Name, "password") {
			if flag.Changed {
				err = fmt.Errorf("flag %s is a secret, please set it in the config file or environment variable to avoid leaking it", flag.Name)
				return
			}
		}
	})
	if err != nil {
		return err
	}

	return configMap.Load(posflag.Provider(flags, defaultDelimiter, configMap), nil)
}

// splitWithEscaping see https://codereview.stackexchange.com/questions/259270/golang-splitting-a-string-by-a-separator-not-prefixed-by-an-escape-string/259382
func splitWithEscaping(s, separator, escape string) []string {
	s = strings.ReplaceAll(s, escape+separator, "\x00")
	tokens := strings.Split(s, separator)
	for i, token := range tokens {
		tokens[i] = strings.ReplaceAll(token, "\x00", separator)
	}
	return tokens
}
