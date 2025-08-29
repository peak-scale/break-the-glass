// Copyright 2024-2025 Peak Scale
// SPDX-License-Identifier: Apache-2.0

package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	crtlmetrics "sigs.k8s.io/controller-runtime/pkg/metrics"

	addonsv1alpha1 "github.com/peak-scale/break-the-glass/api/v1alpha1"
)

type BreakRequestsRecorder struct {
	requestConditionGauge *prometheus.GaugeVec
}

func MustMakeBreakRequestsRecorder() *BreakRequestsRecorder {
	metricsRecorder := NewBreakRequestsRecorder()
	crtlmetrics.Registry.MustRegister(metricsRecorder.Collectors()...)

	return metricsRecorder
}

func NewBreakRequestsRecorder() *BreakRequestsRecorder {
	namespace := "break_requests"
	return &BreakRequestsRecorder{
		requestConditionGauge: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "phase",
				Help:      "The current phase of the BreakRequest.",
			},
			[]string{"name", "target_namespace", "status"},
		),
	}
}

func (r *BreakRequestsRecorder) Collectors() []prometheus.Collector {
	return []prometheus.Collector{
		r.requestConditionGauge,
	}
}

// RecordCondition records the condition as given for the ref.
func (r *BreakRequestsRecorder) RecordRequestCondition(_ *addonsv1alpha1.BreakRequest) {}

// DeleteCondition deletes the condition metrics for the ref.
func (r *BreakRequestsRecorder) DeleteRequestMetrics(_ *addonsv1alpha1.BreakRequest) {}
