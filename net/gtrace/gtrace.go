// Copyright GoFrame Author(https://goframe.org). All Rights Reserved.
//
// This Source Code Form is subject to the terms of the MIT License.
// If a copy of the MIT was not distributed with this file,
// You can obtain one at https://github.com/gogf/gf.

// Package gtrace provides convenience wrapping functionality for tracing feature using OpenTelemetry.
package gtrace

import (
	"context"
	"os"
	"strings"

	"github.com/gogf/gf/v2/internal/intlog"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
	semconv "go.opentelemetry.io/otel/semconv/v1.7.0"
	"go.opentelemetry.io/otel/trace"

	"github.com/gogf/gf/v2/container/gmap"
	"github.com/gogf/gf/v2/container/gvar"
	"github.com/gogf/gf/v2/internal/command"
	"github.com/gogf/gf/v2/net/gipv4"
	"github.com/gogf/gf/v2/util/gconv"
)

const (
	tracingCommonKeyIpIntranet        = `ip.intranet`
	tracingCommonKeyIpHostname        = `hostname`
	commandEnvKeyForTraceEnabled      = "gf.trace.enabled"               // Main switch for tracing feature.
	commandEnvKeyForMaxContentLogSize = "gf.gtrace.max.content.log.size" // To avoid too big tracing content.
	commandEnvKeyForTracingInternal   = "gf.gtrace.tracing.internal"     // For detailed controlling for tracing content.
)

var (
	intranetIps, _           = gipv4.GetIntranetIpArray()
	intranetIpStr            = strings.Join(intranetIps, ",")
	hostname, _              = os.Hostname()
	traceEnabled             = false      // traceEnabled enables tracing feature for all.
	tracingInternal          = true       // tracingInternal enables tracing for internal type spans.
	tracingMaxContentLogSize = 512 * 1024 // Max log size for request and response body, especially for HTTP/RPC request.
	// defaultTextMapPropagator is the default propagator for context propagation between peers.
	defaultTextMapPropagator = propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	)
)

func init() {
	traceEnabled = gconv.Bool(command.GetOptWithEnv(commandEnvKeyForTraceEnabled, "false"))
	tracingInternal = gconv.Bool(command.GetOptWithEnv(commandEnvKeyForTracingInternal, "true"))
	if maxContentLogSize := gconv.Int(command.GetOptWithEnv(commandEnvKeyForMaxContentLogSize)); maxContentLogSize > 0 {
		tracingMaxContentLogSize = maxContentLogSize
	}
	CheckSetDefaultTextMapPropagator()
	intlog.Printf(context.TODO(), `traceEnabled initialized as: %v`, traceEnabled)
}

// SetEnabled enables or disables the tracing feature.
func SetEnabled(enabled bool) {
	traceEnabled = enabled
	intlog.Printf(context.TODO(), `traceEnabled SetEnabled: %v`, enabled)
}

// IsEnabled checks and returns if tracing feature is configured enabled.
func IsEnabled() bool {
	return traceEnabled
}

// IsActivated checks given context and returns if tracing feature is actually activated in this context.
func IsActivated(ctx context.Context) bool {
	return GetTraceID(ctx) != ""
}

// IsTracingInternal returns whether tracing spans of internal components.
func IsTracingInternal() bool {
	return tracingInternal
}

// MaxContentLogSize returns the max log size for request and response body, especially for HTTP/RPC request.
func MaxContentLogSize() int {
	return tracingMaxContentLogSize
}

// CommonLabels returns common used attribute labels:
// ip.intranet, hostname.
func CommonLabels() []attribute.KeyValue {
	return []attribute.KeyValue{
		attribute.String(tracingCommonKeyIpHostname, hostname),
		attribute.String(tracingCommonKeyIpIntranet, intranetIpStr),
		semconv.HostNameKey.String(hostname),
	}
}

// CheckSetDefaultTextMapPropagator sets the default TextMapPropagator if it is not set previously.
func CheckSetDefaultTextMapPropagator() {
	p := otel.GetTextMapPropagator()
	if len(p.Fields()) == 0 {
		otel.SetTextMapPropagator(GetDefaultTextMapPropagator())
	}
}

// GetDefaultTextMapPropagator returns the default propagator for context propagation between peers.
func GetDefaultTextMapPropagator() propagation.TextMapPropagator {
	return defaultTextMapPropagator
}

// GetTraceID retrieves and returns TraceId from context.
// It returns an empty string is tracing feature is not activated.
func GetTraceID(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	traceID := trace.SpanContextFromContext(ctx).TraceID()
	if traceID.IsValid() {
		return traceID.String()
	}
	return ""
}

// GetSpanID retrieves and returns SpanId from context.
// It returns an empty string is tracing feature is not activated.
func GetSpanID(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	spanID := trace.SpanContextFromContext(ctx).SpanID()
	if spanID.IsValid() {
		return spanID.String()
	}
	return ""
}

// SetBaggageValue is a convenient function for adding one key-value pair to baggage.
// Note that it uses attribute.Any to set the key-value pair.
func SetBaggageValue(ctx context.Context, key string, value interface{}) context.Context {
	return NewBaggage(ctx).SetValue(key, value)
}

// SetBaggageMap is a convenient function for adding map key-value pairs to baggage.
// Note that it uses attribute.Any to set the key-value pair.
func SetBaggageMap(ctx context.Context, data map[string]interface{}) context.Context {
	return NewBaggage(ctx).SetMap(data)
}

// GetBaggageMap retrieves and returns the baggage values as map.
func GetBaggageMap(ctx context.Context) *gmap.StrAnyMap {
	return NewBaggage(ctx).GetMap()
}

// GetBaggageVar retrieves value and returns a *gvar.Var for specified key from baggage.
func GetBaggageVar(ctx context.Context, key string) *gvar.Var {
	return NewBaggage(ctx).GetVar(key)
}
