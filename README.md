# Pod Reaper 
Designed to kill pods running in a kubernetes cluster that have been running for 
a longer-than-desired duration.

## Deployment

It is recommended that pod-reaper be deployed directly into your kubernetes 
cluster using a kubernetes deployment specification (see `deployment.yml`). The
configuration of the plugin is controlled by the following environment 
variables:

- `MAX_DURATION`: The duration a pod must be alive before it is eligible for reaping. Example format: `6h45m30s`
- `POLL_INTERVAL`: How often the pod-reaper should check for pods to reap. Example format: `10m` 
- `NAMESPACE`: The kubernetes namespace to query pods for (defaults to ALL namespaces)
- `LABEL_KEY`: key of a `key: value` pair to be excluded from pod-reaping (details below) 
- `LABEL_VALUE`: value of a `key: value` pair to be excluded from pod-reaping (details below)

Pods are selected from the specified `NAMESPACE` (or all namespaces if not specified). If, however, the pod is deployed
into a different namespace than what is specified here, the go-client must be able to access a token that has access to
the specified namespace to make the query for pods. The pods are taken from the kubernetes configuration associated 
with the deployment.

The `LABEL_KEY` and `LABEL_VALUE` can be used to exclude pods from reaping. Any pod deployed with the specified 
key-value pair in the labels section of a pod's metadata will be excluded, effectively allowing for opt-out.

For example, in deployment.yml in this repo, this combination prevents the pod-reaper from reaping itself.
```yaml
    metadata:
      labels:
        pod-reaper: disabled
    ...
    env:
    - name: LABEL_KEY
      value: pod-reaper
    - name: LABEL_VALUE
      value: disabled
```

