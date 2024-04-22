// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package fishs3exporter // import "github.com/open-telemetry/opentelemetry-collector-contrib/exporter/awss3exporter"

import (
	"context"
	"fmt"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/exporter"
	"go.opentelemetry.io/collector/pdata/plog"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.opentelemetry.io/collector/pdata/ptrace"
	"go.uber.org/zap"
)

type s3Exporter struct {
	config     *Config
	dataWriter dataWriter
	logger     *zap.Logger
	marshaler  marshaler
}

func newS3Exporter(config *Config,
	params exporter.CreateSettings) *s3Exporter {

	s3Exporter := &s3Exporter{
		config:     config,
		dataWriter: &s3Writer{},
		logger:     params.Logger,
	}
	return s3Exporter
}

func (e *s3Exporter) start(_ context.Context, host component.Host) error {

	var m marshaler
	var err error
	if e.config.Encoding != nil {
		if m, err = newMarshalerFromEncoding(e.config.Encoding, e.config.EncodingFileExtension, host, e.logger); err != nil {
			return err
		}
	} else {
		if m, err = newMarshaler(e.config.MarshalerName, e.logger); err != nil {
			return fmt.Errorf("unknown marshaler %q", e.config.MarshalerName)
		}
	}

	e.marshaler = m
	return nil
}

func (e *s3Exporter) Capabilities() consumer.Capabilities {
	return consumer.Capabilities{MutatesData: false}
}

func (e *s3Exporter) ConsumeMetrics(ctx context.Context, md pmetric.Metrics) error {
	buf, err := e.marshaler.MarshalMetrics(md)

	if err != nil {
		return err
	}

	return e.dataWriter.writeBuffer(ctx, buf, e.config, "metrics", e.marshaler.format())
}

func (e *s3Exporter) ConsumeLogs(ctx context.Context, logs plog.Logs) error {
	buf, err := e.marshaler.MarshalLogs(logs)

	if err != nil {
		return err
	}

	return e.dataWriter.writeBuffer(ctx, buf, e.config, "logs", e.marshaler.format())
}

func (e *s3Exporter) ConsumeTraces(ctx context.Context, traces ptrace.Traces) error {
	buf, err := e.marshaler.MarshalTraces(traces)
	if err != nil {
		return err
	}

	return e.dataWriter.writeBuffer(ctx, buf, e.config, "traces", e.marshaler.format())
}