# Pod Reaper 
Designed to kill pods running in a kubernetes cluster that have been running for a longer-than-desired duration or that
have a specific status.

The motivation for this is a tool that runs a pod as a one time response to some event. Since the run is created on
demand, a deployment is not appropriate. If the job fails, or is taking to long (as configured) it would otherwise 
require manual intervention to clean pods.

## Deployment

It is recommended that pod-reaper be deployed directly into your kubernetes cluster using a kubernetes deployment 
specification (see `deployment.yml`). The configuration of the plugin is controlled by the following environment 
variables:

- `MAX_POD_DURATION` (default: "2h"): pods with an age greater than this duration will be reaped unless excluded
- `POLL_INTERVAL` (default: "15s"): how often the pod reaper will check for pods to delete
- `CONTAINER_STATUSES` (default: ""): pods with a status included in this comma separated list (with no spaces) will be
 reaped unless exclude
- `EXCLUDE_LABEL_KEY` (default: "pod-reaper"): pods with this key (and corresponding value) as a metadata label will be
 never be deleted by the pod reaper
- `EXCLUDE_LABEL_VALUE` (default: "disabled"): pods with this value (and corresponding key) as a metadata label will 
 never be deleted by the pod reaper
- `NAMESPACE` (default ""): the namespace where pod reaper will look for pods

### Logic

The following is the human readable version of the logic behind pod reaper:

1. get all the pods in the specified `NAMESPACE` that do not have a metadata label with 
 `EXCLUDE_LABEL_KEY: EXCLUDE_LABEL_VALUE`
1. compare the age of the pod to the `MAX_POD_DURATION`
1. check if the pod's status is in the set of statuses supplied to `CONTAINER_STATUSES`
1. if either/both of the cases above are met: delete the pod, logging to `STDOUT`
1. sleep for `POLL_INTERVAL` and run again

While not strictly required, using the `EXCLUDE_LABEL_KEY: EXCLUDE_LABEL_VALUE` metadata label to exclude the pod-reaper
 so that it does not reap itself. There is no guarentee on the order of reaping, so it may reap itself before getting to
 other pods that meet the criteria for deletion.


The `EXCLUDED_LABEL_KEY` and `EXCLUDED_LABEL_VALUE` can be used to exclude pods from reaping. Any pod deployed with the 
specified key-value pair in the labels section of a pod's metadata will be excluded, effectively allowing for opt-out.

For example, in deployment.yml in this repo, this combination prevents the pod-reaper from reaping itself.
```yaml
    metadata:
      labels:
        pod-reaper: disabled
    ...
    env:
    - name: EXCLUDED_LABEL_KEY
      value: pod-reaper
    - name: EXCLUDED_LABEL_VALUE
      value: disabled
```

