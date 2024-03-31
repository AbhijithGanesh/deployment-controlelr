# Go Controller for Deployments

This is a custom controller Proof of Concept (PoC) for accessing resources in a Kubernetes cluster. The controller is designed to manage Deployments in a Kubernetes cluster, providing functionalities such as synchronization, updating, and monitoring of Deployment resources.

## Purpose

The purpose of this controller is to demonstrate how to implement a custom controller in Go for managing Deployments in Kubernetes. It serves as a starting point for developers looking to build their own controllers or extend the functionality provided here for their specific use cases.

## Features

- **Synchronization**: The controller synchronizes Deployment resources with the desired state, ensuring that the cluster state matches the declared configuration.
- **Updating**: It facilitates updating Deployments based on changes in the desired state or external triggers.
- **Monitoring**: The controller continuously monitors Deployment resources for any changes or events and takes appropriate actions as needed.

## Usage

To use this controller, follow these steps:

1. Clone the repository or download the source code.
2. Build the controller using the provided build script or by running `go build`.
3. Deploy the controller to your Kubernetes cluster.
4. Monitor the logs and observe the controller's behavior as it manages Deployment resources.

## Contributions

Contributions to this project are welcome. Feel free to submit issues, feature requests, or pull requests on the GitHub repository.

## Note

Set Kubeconfig variable before running

### Windows

```
set KUBECONFIG="/../../"
```

### Linux

```sh
export KUBECONFIG="../../"
```
