// Licensed to Elasticsearch B.V. under one or more contributor
// license agreements. See the NOTICE file distributed with
// this work for additional information regarding copyright
// ownership. Elasticsearch B.V. licenses this file to you under
// the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

//go:build !requirefips

/*
Package mysql is Metricbeat module for MySQL server.
*/
package mysql

import (
	"crypto/tls"
	"database/sql"
	"fmt"

	"github.com/elastic/beats/v7/metricbeat/mb"
	"github.com/elastic/elastic-agent-libs/transport/tlscommon"

	"github.com/go-sql-driver/mysql"
)

const TLSConfigKey = "custom"

func init() {
	// Register the ModuleFactory function for the "mysql" module.
	if err := mb.Registry.AddModule("mysql", NewModule); err != nil {
		panic(err)
	}
}

func NewModule(base mb.BaseModule) (mb.Module, error) {
	// Validate that at least one host has been specified.
	var c Config
	if err := base.UnpackConfig(&c); err != nil {
		return nil, err
	}

	return &base, nil
}

type Metricset struct {
	mb.BaseMetricSet
	Config Config
}

func NewMetricset(base mb.BaseMetricSet) (*Metricset, error) {
	var c Config
	if err := base.Module().UnpackConfig(&c); err != nil {
		return nil, fmt.Errorf("could not read config: %w", err)
	}

	if c.TLS.IsEnabled() {
		tlsConfig, err := tlscommon.LoadTLSConfig(c.TLS)
		if err != nil {
			return nil, fmt.Errorf("could not load provided TLS configuration: %w", err)
		}

		c.TLSConfig = tlsConfig.ToConfig()
	}

	return &Metricset{Config: c, BaseMetricSet: base}, nil
}

// ParseDSN creates a DSN (data source name) string by parsing the host.
// It validates the resulting DSN and returns an error if the DSN is invalid.
//
//	Format:  [username[:password]@][protocol[(address)]]/
//	Example: root:test@tcp(127.0.0.1:3306)/
func ParseDSN(mod mb.Module, host string) (mb.HostData, error) {
	c := struct {
		Username string            `config:"username"`
		Password string            `config:"password"`
		TLS      *tlscommon.Config `config:"ssl"`
	}{}

	if err := mod.UnpackConfig(&c); err != nil {
		return mb.HostData{}, err
	}

	config, err := mysql.ParseDSN(host)
	if err != nil {
		return mb.HostData{}, fmt.Errorf("error parsing mysql host: %w", err)
	}

	if config.User == "" {
		config.User = c.Username
	}

	if config.Passwd == "" {
		config.Passwd = c.Password
	}

	// Add connection timeouts to the DSN.
	if timeout := mod.Config().Timeout; timeout > 0 {
		config.Timeout = timeout
		config.ReadTimeout = timeout
		config.WriteTimeout = timeout
	}

	noCredentialsConfig := *config
	noCredentialsConfig.User = ""
	noCredentialsConfig.Passwd = ""

	if c.TLS.IsEnabled() {
		config.TLSConfig = TLSConfigKey
	}

	return mb.HostData{
		URI:          config.FormatDSN(),
		SanitizedURI: noCredentialsConfig.FormatDSN(),
		Host:         config.Addr,
		User:         config.User,
		Password:     config.Passwd,
	}, nil
}

// NewDB returns a new mysql database handle. The dsn value (data source name)
// must be valid, otherwise an error will be returned.
//
//	DSN Format: [username[:password]@][protocol[(address)]]/
func NewDB(dsn string, tlsConfig *tls.Config) (*sql.DB, error) {
	if tlsConfig != nil {
		err := mysql.RegisterTLSConfig(TLSConfigKey, tlsConfig)
		if err != nil {
			return nil, fmt.Errorf("registering custom tls config failed: %w", err)
		}
	}

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("sql open failed: %w", err)
	}

	return db, nil
}
