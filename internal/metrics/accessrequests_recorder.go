// Copyright 2024-2025 Peak Scale
// SPDX-License-Identifier: Apache-2.0

package metrics

import (
	sopsv1alpha1 "github.com/peak-scale/sops-operator/api/v1alpha1"
	"github.com/peak-scale/sops-operator/internal/meta"
	"github.com/prometheus/client_golang/prometheus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	crtlmetrics "sigs.k8s.io/controller-runtime/pkg/metrics"
)

type AccessRequestsRecorder struct {
	requestConditionGauge *prometheus.GaugeVec
}

func MustMakeRecorder() *AccessRequestsRecorder {
	metricsRecorder := NewRecorder()
	crtlmetrics.Registry.MustRegister(metricsRecorder.Collectors()...)

	return metricsRecorder
}

func NewRecorder() *AccessRequestsRecorder {
	namespace := "access_requests"

	return &AccessRequestsRecorder{
		requestConditionGauge: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "phase",
				Help:      "The current phase of the AccessRequest.",
			},
			[]string{"name", "target_namespace", "status"},
		),
	}
}

func (r *AccessRequestsRecorder) Collectors() []prometheus.Collector {
	return []prometheus.Collector{
		r.requestConditionGauge,
	}
}

// RecordCondition records the condition as given for the ref.
func (r *AccessRequestsRecorder) RecordRequestCondition(provider *sopsv1alpha1.SopsProvider) {
	for _, status := range []string{meta.ReadyCondition, meta.NotReadyCondition} {
		var value float64
		if provider.Status.Condition.Status == metav1.ConditionTrue {
			value = 1
		}

		r.requestConditionGauge.WithLabelValues(provider.Name, status).Set(value)
	}
}

// DeleteCondition deletes the condition metrics for the ref.
func (r *AccessRequestsRecorder) DeleteRequestMetrics(secret *sopsv1alpha1.SopsSecret) {
	for _, status := range []string{meta.ReadyCondition, meta.NotReadyCondition} {
		r.requestConditionGauge.DeleteLabelValues(secret.Name, secret.Namespace, status)
	}
}
