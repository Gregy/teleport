/*
Copyright 2022 Gravitational, Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package config

import (
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"

	"github.com/gravitational/teleport"
	"github.com/gravitational/teleport/api/types"
	"github.com/gravitational/trace"
)

const (
	DefaultCertificateTTL = 60 * time.Minute
	DefaultRenewInterval  = 20 * time.Minute
	DefaultJoinMethod     = "token"
)

var log = logrus.WithFields(logrus.Fields{
	trace.Component: teleport.ComponentTBot,
})

// CLIConf is configuration from the CLI.
type CLIConf struct {
	ConfigPath string

	Debug      bool
	AuthServer string

	// DataDir stores the bot's internal data.
	DataDir string

	// DestinationDir stores the generated end-user certificates.
	DestinationDir string

	// CAPins is a list of pinned SKPI hashes of trusted auth server CAs, used
	// only on first connect.
	CAPins []string

	// Token is a bot join token.
	Token string

	// RenewalInterval is the interval at which certificates are renewed, as a
	// time.ParseDuration() string. It must be less than the certificate TTL.
	RenewalInterval time.Duration

	// CertificateTTL is the requested TTL of certificates. It should be some
	// multiple of the renewal interval to allow for failed renewals.
	CertificateTTL time.Duration

	// JoinMethod is the method the bot should use to exchange a token for the
	// initial certificate
	JoinMethod string

	// Oneshot controls whether the bot quits after a single renewal.
	Oneshot bool

	// InitDir specifies which destination to initialize if multiple are
	// configured.
	InitDir string

	// BotUser is a Unix username that should be given permission to write
	BotUser string

	// ReaderUser is the Unix username that will be reading the files
	ReaderUser string

	// Owner is the user:group that will own the destination files. Due to SSH
	// restrictions on key permissions, it cannot be the same as the reader
	// user. If ACL support is unused or unavailable, the reader user will own
	// files directly.
	Owner string

	// Clean is a flag that, if set, instructs `tbot init` to remove existing
	// unexpected files.
	Clean bool
}

// OnboardingConfig contains values only required on first connect.
type OnboardingConfig struct {
	// Token is a bot join token.
	Token string `yaml:"token"`

	// CAPath is an optional path to a CA certificate.
	CAPath string

	// CAPins is a list of certificate authority pins, used to validate the
	// connection to the Teleport auth server.
	CAPins []string `yaml:"ca_pins"`

	// JoinMethod is the method the bot should use to exchange a token for the
	// initial certificate
	JoinMethod types.JoinMethod `yaml:"join_method"`
}

// BotConfig is the bot's root config object.
type BotConfig struct {
	Onboarding   *OnboardingConfig    `yaml:"onboarding,omitempty"`
	Storage      *StorageConfig       `yaml:"storage,omitempty"`
	Destinations []*DestinationConfig `yaml:"destinations,omitempty"`

	Debug           bool          `yaml:"debug"`
	AuthServer      string        `yaml:"auth_server"`
	CertificateTTL  time.Duration `yaml:"certificate_ttl"`
	RenewalInterval time.Duration `yaml:"renewal_interval"`
	Oneshot         bool          `yaml:"oneshot"`
}

func (conf *BotConfig) CheckAndSetDefaults() error {
	if conf.Storage == nil {
		conf.Storage = &StorageConfig{}
	}

	if err := conf.Storage.CheckAndSetDefaults(); err != nil {
		return trace.Wrap(err)
	}

	for _, dest := range conf.Destinations {
		if err := dest.CheckAndSetDefaults(); err != nil {
			return trace.Wrap(err)
		}
	}

	if conf.CertificateTTL == 0 {
		conf.CertificateTTL = DefaultCertificateTTL
	}

	if conf.RenewalInterval == 0 {
		conf.RenewalInterval = DefaultRenewInterval
	}

	return nil
}

// GetDestinationByPath attempts to fetch a destination by its filesystem path.
// Only valid for filesystem destinations; returns nil if no matching
// destination exists.
func (conf *BotConfig) GetDestinationByPath(path string) (*DestinationConfig, error) {
	for _, dest := range conf.Destinations {
		destImpl, err := dest.GetDestination()
		if err != nil {
			return nil, trace.Wrap(err)
		}

		destDir, ok := destImpl.(*DestinationDirectory)
		if !ok {
			continue
		}

		// Note: this compares only paths as written in the config file. We
		// might want to compare .Abs() if that proves to be confusing (though
		// this may have its own problems)
		if destDir.Path == path {
			return dest, nil
		}
	}

	return nil, nil
}

// NewDefaultConfig creates a new minimal bot configuration from defaults.
// CheckAndSetDefaults() will be called.
func NewDefaultConfig(authServer string) (*BotConfig, error) {
	// Note: we need authServer for CheckAndSetDefaults to succeed.
	cfg := BotConfig{
		AuthServer: authServer,
	}
	if err := cfg.CheckAndSetDefaults(); err != nil {
		return nil, trace.Wrap(err)
	}

	return &cfg, nil
}

// FromCLIConf loads bot config from CLI parameters, potentially loading and
// merging a configuration file if specified. CheckAndSetDefaults() will
// be called. Note that CLI flags, if specified, will override file values.
func FromCLIConf(cf *CLIConf) (*BotConfig, error) {
	var config *BotConfig
	var err error

	if cf.ConfigPath != "" {
		config, err = ReadConfigFromFile(cf.ConfigPath)

		if err != nil {
			return nil, trace.Wrap(err, "loading bot config from path %s", cf.ConfigPath)
		}
	} else {
		config = &BotConfig{}
	}

	if cf.Debug {
		config.Debug = true
	}

	if cf.Oneshot {
		config.Oneshot = true
	}

	if cf.AuthServer != "" {
		if config.AuthServer != "" {
			log.Warnf("CLI parameters are overriding auth server configured in %s", cf.ConfigPath)
		}
		config.AuthServer = cf.AuthServer
	}

	if cf.CertificateTTL != 0 {
		if config.CertificateTTL != 0 {
			log.Warnf("CLI parameters are overriding certificate TTL configured in %s", cf.ConfigPath)
		}
		config.CertificateTTL = cf.CertificateTTL
	}

	if cf.RenewalInterval != 0 {
		if config.RenewalInterval != 0 {
			log.Warnf("CLI parameters are overriding renewal interval configured in %s", cf.ConfigPath)
		}
		config.RenewalInterval = cf.RenewalInterval
	}

	// DataDir overrides any previously-configured storage config
	if cf.DataDir != "" {
		if config.Storage != nil {
			if _, err := config.Storage.GetDestination(); err != nil {
				log.Warnf("CLI parameters are overriding storage location from %s", cf.ConfigPath)
			}
		}

		config.Storage = &StorageConfig{
			DestinationMixin: DestinationMixin{
				Directory: &DestinationDirectory{
					Path: cf.DataDir,
				},
			},
		}
	}

	if cf.DestinationDir != "" {
		// CLI only supports a single filesystem destination with SSH client config
		// and all roles.
		if len(config.Destinations) > 0 {
			log.Warnf("CLI parameters are overriding destinations from %s", cf.ConfigPath)
		}

		// CheckAndSetDefaults() will configure default kinds and templates
		config.Destinations = []*DestinationConfig{{
			DestinationMixin: DestinationMixin{
				Directory: &DestinationDirectory{
					Path: cf.DestinationDir,
				},
			},
		}}
	}

	// If any onboarding flags are set, override the whole section.
	// (CAPath, CAPins, etc follow different codepaths so we don't want a
	// situation where different fields become set weirdly due to struct
	// merging)
	if cf.Token != "" || len(cf.CAPins) > 0 || cf.JoinMethod != "" {
		onboarding := config.Onboarding
		if onboarding != nil && (onboarding.Token != "" || onboarding.CAPath != "" || len(onboarding.CAPins) > 0) || cf.JoinMethod != DefaultJoinMethod {
			// To be safe, warn about possible confusion.
			log.Warnf("CLI parameters are overriding onboarding config from %s", cf.ConfigPath)
		}

		config.Onboarding = &OnboardingConfig{
			Token:      cf.Token,
			CAPins:     cf.CAPins,
			JoinMethod: types.JoinMethod(cf.JoinMethod),
		}
	}

	if err := config.CheckAndSetDefaults(); err != nil {
		return nil, trace.Wrap(err, "validing merged bot config")
	}

	return config, nil
}

// ReadFromFile reads and parses a YAML config from a file.
func ReadConfigFromFile(filePath string) (*BotConfig, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, trace.Wrap(err, fmt.Sprintf("failed to open file: %v", filePath))
	}

	defer f.Close()
	return ReadConfig(f)
}

// ReadConfig parses a YAML config file from a Reader.
func ReadConfig(reader io.Reader) (*BotConfig, error) {
	var config BotConfig

	decoder := yaml.NewDecoder(reader)
	decoder.KnownFields(true)
	if err := decoder.Decode(&config); err != nil {
		return nil, trace.BadParameter("failed parsing config file: %s", strings.Replace(err.Error(), "\n", "", -1))
	}

	return &config, nil
}
