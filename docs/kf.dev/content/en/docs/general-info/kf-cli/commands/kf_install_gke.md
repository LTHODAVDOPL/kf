---
title: "kf install gke"
slug: kf-install-gke
url: /docs/general-info/kf-cli/commands/kf-install-gke/
---
## kf install gke

Install kf on GKE with Cloud Run (Note: this will incur GCP costs)

### Synopsis

This interactive installer will walk you through the process of installing kf on GKE with Cloud Run. You MUST have gcloud and kubectl installed and available on the path. Note: running this will incur costs to run GKE. See https://cloud.google.com/products/calculator/ to get an estimate.

 To override the GKE version that's chosen, set the environment variable GKE_VERSION.

```
kf install gke [subcommand] [flags]
```

### Examples

```
  kf install gke
```

### Options

```
  -h, --help      help for gke
  -v, --verbose   Display the gcloud and kubectl commands
```

### Options inherited from parent commands

```
      --config string       Config file (default is $HOME/.kf)
      --kubeconfig string   Kubectl config file (default is $HOME/.kube/config)
      --log-http            Log HTTP requests to stderr
      --namespace string    Kubernetes namespace to target
```

### SEE ALSO

* [kf install](/docs/general-info/kf-cli/commands/kf-install/)	 - Install kf

