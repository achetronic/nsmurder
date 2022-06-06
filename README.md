# Nsmurder

> This CLI is intended to be used into a pipeline

## Description

A CLI to assassinate namespaces following different strategies sequentially

## Motivation

Some companies manage Kubernetes clusters dinamically using pipelines.
One of the problems of this approach is not on creation time, but on destruction.

Kubernetes distributions, like EKS or GKE, can create cloud resources in the provider,
such as DNS registries, LoadBalancers or Volumes dynamically, when creating Ingress,
Service or PVC resources.

The reason for this is that they are attached to the provider that is running it
(under the hoods). This is convenient for the customers, but need a cleanning
process when destroying a cluster, in order not to leave orphan resources
created on the cloud (which means money)

Thinking about this problem and best practise, we created this CLI as a single
binary, which can be integrated on the pipelines to do the cleaning trick.

## Strategies

The process followed to assassinate namespaces is described in the following steps:

> There is a time between steps that can be configured using the flag `--duration-between-strategies`

1. Schedule namespace deletion. The intention for this step is to identify later
   which ones are stuck in `Terminating` status.

2. Get a list with every resource type that can be created into a namespace.
   Then loop over namespaces which are in `Terminating` status. On each cycle,
   delete all resources inside for all possible types to clean it.

3. Loop over namespaces which are still in `Terminating` status and remove the
   finalizers

## Flags

There are several flags that can be configured to change the behaviour of the
application. They are described in the following table:

| Name                            | Description                                           | Default | Example                               |
| :------------------------------ | :---------------------------------------------------- | :-----: | :------------------------------------ |
| `--include-all`                 | Include all the namespaces. This override `--include` |    -    | -                                     |
| `--include`                     | Include one namespace                                 |    -    | `--include default`                   |
| `--exclude`                     | Exclude one namespace                                 |    -    | `--exclude kube-system`               |
| `--duration-between-strategies` | Duration between strategies perform                   |    -    | `--duration-between-strategies "30s"` |
| `--kubeconfig`                  | Path to the kubeconfig file                           |    -    | `--kubeconfig "~/.kube/config"`       |

## Example

```sh
nsmurder --include-all \
         --exclude kube-system \
         --exclude kube-public \
         --exclude kube-node-lease \
         --exclude external-dns \
         --exclude calico-system \
         --duration-between-strategies "5m" \
         --kubeconfig "~/.kube/config"
```

## How to collaborate

We are open to external collaborations for this project: improvements, bugfixes, whatever.
For doing it you must fork the repository, make your changes to the code and open a PR.
The code will be reviewed and tested (always)

> We are developers and hate bad code. For that reason we ask you the highest quality
> on each line of code to improve this project on each iteration.
