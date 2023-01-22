package main

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"os"
	"os/signal"
	"syscall"

	"perf-test/pkg/config"
	"perf-test/pkg/interact"

	log "github.com/sirupsen/logrus"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// retrieve the Kubernetes cluster client from outside of the cluster
func getKubernetesClient() *kubernetes.Clientset {
	// construct the path to kubeconfig
	// kubeConfigPath := os.Getenv("HOME") + "/temp/.kubeconfig"

	// create the config from the path
	// config, err := clientcmd.BuildConfigFromFlags("", kubeConfigPath)
	// if err != nil {
	//         log.Fatalf("getClusterConfig: %v", err)
	// }

	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err.Error())
	}
	// config.QPS = 300
	// config.Burst = 400

	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatalf("getClusterConfig: %v", err)
	}
	return client
}

func readConfigFile() (config.Config, error) {
	var cfg config.Config
	rundir, err := os.Getwd()
	if err != nil {
		return cfg, err
	}
	confFile := rundir + "/config/config"

	raw, err := ioutil.ReadFile(confFile)
	if err != nil {
		return cfg, err
	}

	if err := json.Unmarshal(raw, &cfg); err != nil {
		return cfg, err
	}
	return cfg, nil
}

// main code path
func main() {
	// read config
	cfg, err := readConfigFile()
	if err != nil {
		log.WithError(err).Fatal("could not read config file")
	}
	if cfg.Setup == nil {
		log.Fatal("no setup configured")
	}

	// get the Kubernetes client for connectivity
	client := getKubernetesClient()

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer func() {
		cancel()
	}()

	switch cfg.Setup.Test {
	case "watch":
		if err = interact.Watch(ctx, cfg.Setup, client); err != nil {
			log.WithError(err).Fatal("could not run watch test")
		}
	case "create":
		if err = interact.Create(ctx, cfg.Setup, client); err != nil {
			log.WithError(err).Fatal("could not run create test")
		}
	default:
		log.Fatalf("invalid test name: %s", cfg.Setup.Test)
	}
}
