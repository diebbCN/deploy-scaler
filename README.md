# deploy-scaler
deploy-scaler is a kubernetes operator aiming to autoscale deployment in specified time period.

## Description
Autoscales the pods of deployments that you specified in YAML in a specified time period.


### Prerequisites
- go version v1.23.0+
- docker version 17.03+.
- kubectl version v1.11.3+.
- Access to a Kubernetes v1.11.3+ cluster.