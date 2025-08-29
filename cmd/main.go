/*
Copyright 2025.

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

package main

import (
	"crypto/tls"
	"flag"
	"os"
	"path/filepath"

	// Import all Kubernetes client auth plugins (e.g. Azure, GCP, OIDC, etc.)
	// to ensure that exec-entrypoint and run can make use of them.
	_ "k8s.io/client-go/plugin/pkg/client/auth"

	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/certwatcher"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/metrics/filters"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	addonsv1alpha1 "github.com/peak-scale/break-the-glass/api/v1alpha1"
	"github.com/peak-scale/break-the-glass/internal/controller"
	"github.com/peak-scale/break-the-glass/internal/metrics"
	"github.com/peak-scale/break-the-glass/internal/webhooks"
	// +kubebuilder:scaffold:imports
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))

	utilruntime.Must(addonsv1alpha1.AddToScheme(scheme))
	// +kubebuilder:scaffold:scheme
}

// nolint:gocyclo
func main() {
	var metricsAddr string
	var metricsCertPath, metricsCertName, metricsCertKey string
	var webhookCertPath, webhookCertName, webhookCertKey string
	var enableLeaderElection, enablePprof, hooks bool
	var probeAddr string
	var secureMetrics bool
	var enableHTTP2 bool
	var tlsOpts []func(*tls.Config)
	flag.StringVar(
		&metricsAddr,
		"metrics-bind-address",
		"0",
		"The address the metrics endpoint binds to. "+
			"Use :8443 for HTTPS or :8080 for HTTP, or leave as 0 to disable the metrics service.",
	)
	flag.StringVar(
		&probeAddr,
		"health-probe-bind-address",
		":8081",
		"The address the probe endpoint binds to.",
	)
	flag.BoolVar(&enableLeaderElection, "leader-elect", false,
		"Enable leader election for controller manager. "+
			"Enabling this will ensure there is only one active controller manager.")
	flag.BoolVar(
		&secureMetrics,
		"metrics-secure",
		true,
		"If set, the metrics endpoint is served securely via HTTPS. Use --metrics-secure=false to use HTTP instead.",
	)
	flag.StringVar(
		&webhookCertPath,
		"webhook-cert-path",
		"",
		"The directory that contains the webhook certificate.",
	)
	flag.StringVar(
		&webhookCertName,
		"webhook-cert-name",
		"tls.crt",
		"The name of the webhook certificate file.",
	)
	flag.StringVar(
		&webhookCertKey,
		"webhook-cert-key",
		"tls.key",
		"The name of the webhook key file.",
	)
	flag.StringVar(&metricsCertPath, "metrics-cert-path", "",
		"The directory that contains the metrics server certificate.")
	flag.StringVar(
		&metricsCertName,
		"metrics-cert-name",
		"tls.crt",
		"The name of the metrics server certificate file.",
	)
	flag.StringVar(
		&metricsCertKey,
		"metrics-cert-key",
		"tls.key",
		"The name of the metrics server key file.",
	)
	flag.BoolVar(&enableHTTP2, "enable-http2", false,
		"If set, HTTP/2 will be enabled for the metrics and webhook servers")
	flag.BoolVar(
		&enablePprof,
		"enable-pprof",
		false,
		"Enables Pprof endpoint for profiling (not recommend in production)",
	)
	flag.BoolVar(&hooks, "enable-webhooks", false, "Register Mutating Webhooks to be serving")

	opts := zap.Options{
		Development: true,
	}
	opts.BindFlags(flag.CommandLine)
	flag.Parse()

	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&opts)))

	// if the enable-http2 flag is false (the default), http/2 should be disabled
	// due to its vulnerabilities. More specifically, disabling http/2 will
	// prevent from being vulnerable to the HTTP/2 Stream Cancellation and
	// Rapid Reset CVEs. For more information see:
	// - https://github.com/advisories/GHSA-qppj-fm5r-hxr3
	// - https://github.com/advisories/GHSA-4374-p667-p6c8
	disableHTTP2 := func(c *tls.Config) {
		setupLog.Info("disabling http/2")
		c.NextProtos = []string{"http/1.1"}
	}

	if !enableHTTP2 {
		tlsOpts = append(tlsOpts, disableHTTP2)
	}

	// Create watchers for metrics and webhooks certificates
	var metricsCertWatcher, webhookCertWatcher *certwatcher.CertWatcher

	// Initial webhook TLS options
	webhookTLSOpts := tlsOpts

	if len(webhookCertPath) > 0 {
		setupLog.Info(
			"Initializing webhook certificate watcher using provided certificates",
			"webhook-cert-path",
			webhookCertPath,
			"webhook-cert-name",
			webhookCertName,
			"webhook-cert-key",
			webhookCertKey,
		)

		var err error
		webhookCertWatcher, err = certwatcher.New(
			filepath.Join(webhookCertPath, webhookCertName),
			filepath.Join(webhookCertPath, webhookCertKey),
		)
		if err != nil {
			setupLog.Error(err, "Failed to initialize webhook certificate watcher")
			os.Exit(1)
		}

		webhookTLSOpts = append(webhookTLSOpts, func(config *tls.Config) {
			config.GetCertificate = webhookCertWatcher.GetCertificate
		})
	}

	// Metrics endpoint is enabled in 'config/default/kustomization.yaml'. The Metrics options configure the server.
	// More info:
	// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.20.4/pkg/metrics/server
	// - https://book.kubebuilder.io/reference/metrics.html
	metricsServerOptions := metricsserver.Options{
		BindAddress:   metricsAddr,
		SecureServing: secureMetrics,
		TLSOpts:       tlsOpts,
	}

	if secureMetrics {
		// FilterProvider is used to protect the metrics endpoint with authn/authz.
		// These configurations ensure that only authorized users and service accounts
		// can access the metrics endpoint. The RBAC are configured in 'config/rbac/kustomization.yaml'. More info:
		// https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.20.4/pkg/metrics/filters#WithAuthenticationAndAuthorization
		metricsServerOptions.FilterProvider = filters.WithAuthenticationAndAuthorization
	}

	// If the certificate is not specified, controller-runtime will automatically
	// generate self-signed certificates for the metrics server. While convenient for development and testing,
	// this setup is not recommended for production.
	//
	// TODO(user): If you enable certManager, uncomment the following lines:
	// - [METRICS-WITH-CERTS] at config/default/kustomization.yaml to generate and use certificates
	// managed by cert-manager for the metrics server.
	// - [PROMETHEUS-WITH-CERTS] at config/prometheus/kustomization.yaml for TLS certification.
	if len(metricsCertPath) > 0 {
		setupLog.Info(
			"Initializing metrics certificate watcher using provided certificates",
			"metrics-cert-path",
			metricsCertPath,
			"metrics-cert-name",
			metricsCertName,
			"metrics-cert-key",
			metricsCertKey,
		)

		var err error
		metricsCertWatcher, err = certwatcher.New(
			filepath.Join(metricsCertPath, metricsCertName),
			filepath.Join(metricsCertPath, metricsCertKey),
		)
		if err != nil {
			setupLog.Error(err, "to initialize metrics certificate watcher", "error", err)
			os.Exit(1)
		}

		metricsServerOptions.TLSOpts = append(
			metricsServerOptions.TLSOpts,
			func(config *tls.Config) {
				config.GetCertificate = metricsCertWatcher.GetCertificate
			},
		)
	}

	ctrlConfig := ctrl.Options{
		Scheme:                 scheme,
		Metrics:                metricsServerOptions,
		HealthProbeBindAddress: probeAddr,
		LeaderElection:         enableLeaderElection,
		LeaderElectionID:       "872932ca.projectcapsule.dev",
	}

	// Conditional config
	if hooks {
		ctrlConfig.WebhookServer = webhook.NewServer(webhook.Options{
			Port:    9443,
			TLSOpts: webhookTLSOpts,
		})
	}

	if enablePprof {
		ctrlConfig.PprofBindAddress = ":8082"
	}

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrlConfig)
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	if hooks {
		setupLog.Info("registering webhooks")
		mgr.GetWebhookServer().Register("/requests/mutate", &admission.Webhook{
			Handler: &webhooks.AccessRequestMutatingWebhook{
				Decoder: admission.NewDecoder(mgr.GetScheme()),
				Client:  mgr.GetClient(),
				Log:     ctrl.Log.WithName("Webhooks").WithName("BreakRequests"),
			},
		})
	}

	if err = (&controller.BreakRequestReconciler{
		Client:   mgr.GetClient(),
		Log:      ctrl.Log.WithName("Controllers").WithName("BreakRequests"),
		Scheme:   mgr.GetScheme(),
		Recorder: mgr.GetEventRecorderFor("access-request-ctrl"),
		Metrics:  *metrics.MustMakeBreakRequestsRecorder(),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "BreakRequests")
		os.Exit(1)
	}
	if err := (&controller.BreakRequestTemplateReconciler{
		Client: mgr.GetClient(),
		Log:    ctrl.Log.WithName("Controllers").WithName("BreakRequestTemplates"),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "BreakRequestTemplates")
		os.Exit(1)
	}
	// +kubebuilder:scaffold:builder

	if metricsCertWatcher != nil {
		setupLog.Info("Adding metrics certificate watcher to manager")
		if err := mgr.Add(metricsCertWatcher); err != nil {
			setupLog.Error(err, "unable to add metrics certificate watcher to manager")
			os.Exit(1)
		}
	}

	if webhookCertWatcher != nil {
		setupLog.Info("Adding webhook certificate watcher to manager")
		if err := mgr.Add(webhookCertWatcher); err != nil {
			setupLog.Error(err, "unable to add webhook certificate watcher to manager")
			os.Exit(1)
		}
	}

	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up health check")
		os.Exit(1)
	}
	if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up ready check")
		os.Exit(1)
	}

	setupLog.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}
