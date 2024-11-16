# Nsmurder

![GitHub go.mod Go version (subdirectory of monorepo)](https://img.shields.io/github/go-mod/go-version/achetronic/nsmurder)
![GitHub](https://img.shields.io/github/license/achetronic/nsmurder)

![YouTube Channel Subscribers](https://img.shields.io/youtube/channel/subscribers/UCeSb3yfsPNNVr13YsYNvCAw?label=achetronic&link=http%3A%2F%2Fyoutube.com%2Fachetronic)
![GitHub followers](https://img.shields.io/github/followers/achetronic?label=achetronic&link=http%3A%2F%2Fgithub.com%2Fachetronic)
![X (formerly Twitter) Follow](https://img.shields.io/twitter/follow/achetronic?style=flat&logo=twitter&link=https%3A%2F%2Ftwitter.com%2Fachetronic)

## Description

A CLI to assassinate Kubernetes namespaces following different strategies without mercy

## Motivation

Some companies manage Kubernetes clusters dynamically using pipelines.
One of the problems of this approach is not on creation time, but on destruction.

Kubernetes' distributions, like EKS or GKE, can create cloud resources in the provider,
such as DNS registries, LoadBalancers or Volumes dynamically, when creating Ingress,
Service or PVC resources.

The reason for this is that they are attached to the provider that is running it
(under the hoods). This is convenient for the customers, but need a cleaning
process when destroying a cluster, in order not to leave orphan resources
created on the cloud (which means money)

Thinking about this problem and best practise, we created this CLI as a single
binary, which can be integrated on the pipelines to do the cleaning trick.

## Strategies

The process followed to assassinate namespaces is described in the following steps:

> There is a time between steps that can be configured using the flag `--duration-between-strategies`

1. Schedule namespace deletion for all the namespaces introduced by using the flags.
   This step is intended to identify which ones are stuck in `Terminating` state later.

2. Get a list with every resource type that can be created into a namespace.
   Then loop over namespaces which are in `Terminating` status.
   For each namespace, delete all resources inside to clean it.

3. Loop over namespaces which are still in `Terminating` state and remove their finalizers

## Flags

There are several flags that can be configured to change the behaviour of the
application. They are described in the following table:

| Name                            | Description                                           |     Default      | Example                               |
|:--------------------------------|:------------------------------------------------------|:----------------:|:--------------------------------------|
| `--include-all`                 | Include all the namespaces. This override `--include` |     `false`      | `--include-all`                       |
| `--include`                     | Include one namespace to be deleted                   |        -         | `--include default`                   |
| `--ignore`                      | Ignore deletion of one namespace                      |        -         | `--ignore kube-system`                |
| `--duration-between-strategies` | Duration between strategies perform                   |       `1m`       | `--duration-between-strategies "30s"` |
| `--kubeconfig`                  | Path to the kubeconfig file                           | `~/.kube/config` | `--kubeconfig "~/.kube/config"`       |
| `--help`                        | Show this help message                                |        -         | -                                     |

## Examples

To delete all the namespaces, ignoring some of them, execute the command as follows:

```sh
nsmurder --include-all \
         --ignore "kube-system,kube-public,external-dns,calico-system" \
         --duration-between-strategies "5m" \
         --kubeconfig "~/.kube/config"
```

To delete only some namespaces, execute it as in the following example:

```sh
nsmurder --include "app-develop,app-staging,app-production" \
         --ignore "kube-system,kube-public,external-dns,calico-system" \
         --duration-between-strategies "50s" \
         --kubeconfig "~/.kube/config"
```

> ATTENTION:
> If you execute this CLI, and have other namespaces in terminating state, they will be processed too.
> The reason for this is that the project pretends to do the best job when is part of a cleaning pipeline, etc.

## How to collaborate

We are open to external collaborations for this project: improvements, bugfixes, whatever.

For doing it, open an issue to discuss the need of the changes, then:

- Fork the repository
- Make your changes to the code
- Open a PR and wait for review

The code will be reviewed and tested (always)

> We are developers and hate bad code. For that reason we ask you the highest quality
> on each line of code to improve this project on each iteration.
