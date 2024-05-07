# lfgw-config-operator

Sync ACLs from different Namespaces to common ConfigMap used by [LFGW](https://github.com/weisdd/lfgw).

## Description

[LFGW](https://github.com/weisdd/lfgw) - is a simple reverse proxy designed for filtering PromQL / MetricsQL metrics based on OIDC roles. It utilizes VictoriaMetrics/metricsql to manipulate label filters in metric expressions according to an Access Control List (ACL) before forwarding a request to Prometheus/VictoriaMetrics. 

To configure metric filtering, you need to describe a configuration file, for example: acl.yaml

```yaml
admin: .*
wallet-stage-ro: wallet-stage
wallet-stage-rw: wallet-stage
```
LFGW will read this file and apply filtering according to the user roles received from the OIDC provider. 

__lfgw-config-operator__ allows you not to describe all ACL rules in a single ConfigMap, but to deploy them in different namespaces as CustorResource

```yaml
apiVersion: controls.lfgw.io/v1alpha1
kind: ACL
metadata:
  name: example-acl
  namespace: test
spec:
  rules:
    - roleName: "admin"
      namespaceFilter: ".*"
    - roleName: "bots-dev-ro"
      namespaceFilter: "bots-dev"
```
The operator monitors CustomResource ACLs and adds ACL-rules to the target ConfigMap, which is mounted to LFGW. This allows us to manage LFGW configuration more flexibly.

## Getting Started

**Install the CRDs into the cluster:**

```shell
kubectl apply -f config/crd/bases
```

**Deploy operator**

Install CRD:
```bash
kubectl apply -f https://raw.githubusercontent.com/MadEngineX/lfgw-config-operator/main/config/crd/bases/controls.lfgw.io_acls.yaml
```

Deploy as Helm release:
```bash 
helm repo add m8x https://MadEngineX.github.io/helm-charts/
helm repo update

helm upgrade --install lfgw-operator m8x/lfgw-operator-chart 
```

See all possible [values](https://github.com/MadEngineX/helm-charts/blob/main/charts/lfgw-operator-chart/values.yaml).

You can also deploy __lfgw-config-operator__ + __LFGW__ from one Helm Chart: -
- https://github.com/MadEngineX/helm-charts/tree/main/charts/lfgw

## Docker images

Docker images are published on Dockerhub: [ksxack/lfgw-config-operator](https://hub.docker.com/r/ksxack/lfgw-config-operator/tags)

## Configuration

Environment variables:

| Name         | Type   | Description                                                                                                 |
|--------------|--------|-------------------------------------------------------------------------------------------------------------|
| CM_NAMESPACE | string | Namespace in which ConfigMap containing the ACL file for LFGW must be deployed, default: "infra-monitoring" |
| CM_NAME      | string | Name of ConfigMap, default: "lfgw-config"                                                                   |
| CM_FILENAME  | string | Name of file inside ConfigMap, default: "acl.yaml"                                                          |
| LOG_LEVEL    | string | info/warn/debug/trace, default:"info"                                                                       |

## ToDo

1. Current version of the lfgw-config-operator doesn't support managing the LFGW instance. Therefore, when the operator updates ConfigMap with LFGW ACLs, nothing happens. To automatically trigger LFGW to re-read ACLs from ConfigMap, external tools such as [stakater/Reloader](https://github.com/stakater/Reloader) need to be used. It is necessary to add the capability to the operator to manage LFGW-instance in order to simplify the stack installation.

