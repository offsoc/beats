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

package postgresql

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/elastic/beats/v7/metricbeat/mb"

	// Register postgresql database/sql driver
	_ "github.com/lib/pq"
)

type MetricSet struct {
	mb.BaseMetricSet

	db *sql.DB
}

// NewMetricSet creates a PostgreSQL metricset with a pool of connections
func NewMetricSet(base mb.BaseMetricSet) (*MetricSet, error) {
	return &MetricSet{BaseMetricSet: base}, nil
}

// DB creates a database connection, it must be freed after use with `Close()`
func (ms *MetricSet) DB(ctx context.Context) (*sql.Conn, error) {
	if ms.db == nil {
		db, err := sql.Open("postgres", ms.HostData().URI)
		if err != nil {
			return nil, fmt.Errorf("failed to open connection: %w", err)
		}
		ms.db = db
	}
	return ms.db.Conn(ctx)
}

// QueryStats makes the database call for a given metric
func (ms *MetricSet) QueryStats(ctx context.Context, query string) ([]map[string]interface{}, error) {
	db, err := ms.DB(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to obtain a connection with the database: %w", err)
	}
	defer db.Close()

	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query database: %w", err)
	}

	columns, err := rows.Columns()
	if err != nil {
		return nil, fmt.Errorf("scanning columns: %w", err)
	}
	vals := make([][]byte, len(columns))
	valPointers := make([]interface{}, len(columns))
	for i := range vals {
		valPointers[i] = &vals[i]
	}

	results := []map[string]interface{}{}

	for rows.Next() {
		err = rows.Scan(valPointers...)
		if err != nil {
			return nil, fmt.Errorf("scanning row: %w", err)
		}

		result := map[string]interface{}{}
		for i, col := range columns {
			result[col] = string(vals[i])
		}

		ms.Logger().Named("postgresql").Debugf("Result: %v", result)
		results = append(results, result)
	}
	return results, nil
}

// Close closes the metricset and its connections
func (ms *MetricSet) Close() error {
	if ms.db == nil {
		return nil
	}

	if err := ms.db.Close(); err != nil {
		return fmt.Errorf("failed to close connection: %w", err)
	}
	return nil
}
