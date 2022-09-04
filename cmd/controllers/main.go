package main

import (
	"flag"
	"k8s.io/client-go/tools/clientcmd/api"
	melodyiov1alpha1 "melody/api/v1alpha1"
	"os"

	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	"melody/controllers"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func init() {
	_ = clientgoscheme.AddToScheme(scheme)
	_ = melodyiov1alpha1.AddToScheme(scheme)
}

func main() {
	var (
		ctrlMetricsAddr string
		//metricsAddr          string
		enableLeaderElection bool
	)

	flag.StringVar(&ctrlMetricsAddr, "controller-metrics-addr", ":8080", "The address the controller metric endpoint binds to.")
	flag.BoolVar(&enableLeaderElection, "enable-leader-election", false, "Enable leader election for controller manager. Enabling this will ensure there is only one active controller manager.")
	flag.Parse()
	ctrl.SetLogger(zap.New(zap.UseDevMode(true)))

	opt := ctrl.Options{
		Scheme:             scheme,
		MetricsBindAddress: ctrlMetricsAddr,
		Port:               9443,
		LeaderElection:     enableLeaderElection,
		//LeaderElectionID:   "tuning-kubedl-morphling",
	}

	// Create manager to provide shared dependencies and start components
	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), opt)
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	// Setup Scheme for all resources
	setupLog.Info("Setting up scheme")
	if err := api.AddToScheme(mgr.GetScheme()); err != nil {
		setupLog.Error(err, "unable to add APIs to scheme")
		os.Exit(1)
	}

	// Setup all Controllers
	setupLog.Info("Setting up controller")
	if err := controllers.AddToManager(mgr); err != nil {
		setupLog.Error(err, "unable to register controllers to the manager")
		os.Exit(1)
	}

	// Start the Cmd
	setupLog.Info("Starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}
