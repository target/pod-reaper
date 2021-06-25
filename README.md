# pod-reaper: kills pods dead

[![license](https://img.shields.io/github/license/mashape/apistatus.svg)](LICENSE)
[![release](https://img.shields.io/github/release/target/pod-reaper.svg)](https://github.com/target/pod-reaper/releases/latest)
[![docker](https://img.shields.io/docker/automated/target/pod-reaper.svg)](https://hub.docker.com/r/target/pod-reaper)

A rules based pod killing container. Pod-Reaper was designed to kill pods that meet specific conditions. See the "Implemented Rules" section below for details on specific rules.

## Configuring Pod Reaper

Pod-Reaper is configurable through environment variables. The pod-reaper specific environment variables are:

- `NAMESPACE` the kubernetes namespace where pod-reaper should look for pods
- `GRACE_PERIOD` duration that pods should be given to shut down before hard killing the pod
- `SCHEDULE` schedule for when pod-reaper should look for pods to reap
- `RUN_DURATION` how long pod-reaper should run before exiting
- `EVICT` try to evict pods instead of deleting them
- `EXCLUDE_LABEL_KEY` pod metadata label (of key-value pair) that pod-reaper should exclude
- `EXCLUDE_LABEL_VALUES` comma-separated list of metadata label values (of key-value pair) that pod-reaper should exclude
- `REQUIRE_LABEL_KEY` pod metadata label (of key-value pair) that pod-reaper should require
- `REQUIRE_LABEL_VALUES` comma-separated list of metadata label values (of key-value pair) that pod-reaper should require
- `REQUIRE_ANNOTATION_KEY` pod metadata annotation (of key-value pair) that pod-reaper should require
- `REQUIRE_ANNOTATION_VALUES` comma-separated list of metadata annotation values (of key-value pair) that pod-reaper should require
- `RULES` comma-separated list of rules to load regardless of default

Additionally, at least one rule must be enabled, or the pod-reaper will error and exit. See the Rules section below for configuring and enabling rules.

Example environment variables:

```sh
# pod-reaper configuration
NAMESPACE=test
SCHEDULE=@every 30s
RUN_DURATION=15m
EXCLUDE_LABEL_KEY=pod-reaper
EXCLUDE_LABEL_VALUES=disabled,false

# enable at least one rule
CHAOS_CHANCE=.001
```

#### Annotations

Rule configuration may be overridden by annotations on individual pods. For single-value rules, the configured rule value will be replaced by the annotation value. For multi-value rules, annotations will be added to the configured rule values. See [Implemented Rules](#implemented-rules) for available annotations.

Example environment variables with annotations:

```sh
# pod-reaper configuration
NAMESPACE=test
SCHEDULE=@every 30s

# enable at least one rule
MAX_UNREADY=5m
RULES=duration,unready
```

Pods

```yaml
apiVersion: v1
kind: Pod
metadata:
  name: test
  annotations:
    pod-reaper/max-duration: 1h
spec:
  containers:
  - name: nginx
    image: nginx:1.14.2
    ports:
    - containerPort: 80
```

In this configuration, the Duration, and Unready rules will be loaded. The pod will be reaped if it is older than 1 hour and unready for 5 minutes.

### `NAMESPACE`

Default value: "" (which will look at ALL namespaces)

Controls which kubernetes namespace the pod-reaper is in scope for the pod-reaper. Note that the pod-reaper uses an `InClusterConfig` which makes use of the service account that kubernetes gives to its pods. Only pods (and namespaces) accessible to this service account will be visible to the pod-reaper.

### `GRACE_PERIOD`

Default value: nil (indicates to the use the default specified for pods)

Controls the grace period between a soft pod termination and a hard termination. This will determine the time between when the pod's containers are send a `SIGTERM` signal and when they are sent a `SIGKILL` signal. The format follows the go-lang `time.duration` format (example: "1h15m30s"). A duration of `0s` can be considered a hard kill of the pod.

### `SCHEDULE`

Default value: "@every 1m"

Controls how frequently pod-reaper queries kubernetes for pods. The format follows the upstream cron library https://godoc.org/github.com/robfig/cron. For most use cases, the interval format `@every 1h2m3s` is sufficient. But more complex use cases can make use of the `* * * * *` notation. The cron parser used can optionally support seconds if a sixth parameter is add. `12 * * * * *` for example will run on the 12th second of every minute.

### `RUN_DURATION`

Default value: "0s" (which corresponds to running indefinitely)

Controls the minimum duration that pod-reaper will run before intentionally exiting. The value of "0s" (or anything equivalent such as the empty string) will be interpreted as an indefinite run duration. The format follows the go-lang `time.duration` format (example: "1h15m30s"). Pod-Reaper will not wait for reap-cycles to finishing waiting and will exit immediately (with exit code 0) after the duration has elapsed.

Warnings about `RUN_DURATION`

- pod-rescheduling: if the reaper completes, even successfully, it may be restarted depending on the pod-spec.
- self-reaping: the pod-reaper can reap itself if configured to do so, this can cause the reaper to not run for the expected duration.

Recommendations:

One time run:

- create a pod spec and apply it to kubernetes
- make the pod spec has `restartPolicy: Never`
- add an exclusion label and key using `EXCLUDE_LABEL_KEY` and `EXCLUDE_LABEL_VALUES`
- make the pod spec for the reaper match an excluded label and key to prevent it from reaping itself

Sustained running:

- do not use `RUN_DURATION`
- manage the pod reaper via a deployment

### `EVICT`

Use the [Eviction API](https://kubernetes.io/docs/tasks/administer-cluster/safely-drain-node/#eviction-api) instead of pod deletion when reaping pods.  The Eviction API will honor the [disruption budget](https://kubernetes.io/docs/tasks/run-application/configure-pdb/) assigned to pods, and can for example be useful when reaping pods by duration to ensure that you don't reap all the pods of a specific deployment simultaneously, interrupting a published service.  When a pod cannot be reaped due to a disruption budget, the reason will be logged as a warning.

### `EXCLUDE_LABEL_KEY` and `EXCLUDE_LABEL_VALUES`

These environment variables are used to build a label selector to exclude pods from reaping. The key must be a properly formed kubernetes label key. Values are a comma-separated (without whitespace) list of kubernetes label values. Setting exactly one of the key or values environment variables will result in an error.

A pod will be excluded from the pod-reaper if the pod has a metadata label has a key corresponding to the pod-reaper's exclude label key, and that same metadata label has a value in the pod-reaper's list of excluded label values. This means that exclusion requires both the pod-reaper and pod to be configured in a compatible way.

### `REQUIRE_LABEL_KEY` and `REQUIRE_LABEL_VALUES`

These environment variables build a label selector that pods must match in order to be reaped. Use them the same way as you would `EXCLUDE_LABEL_KEY` and `EXCLUDE_LABEL_VALUES`.

### `REQUIRE_ANNOTATION_KEY` and `REQUIRE_ANNOTATION_VALUES`

These environment variables build a annotation selector that pods must match in order to be reaped. Use them the same way as you would `EXCLUDE_LABEL_KEY` and `EXCLUDE_LABEL_VALUES`.

### `DRY_RUN`

Deafult value: unset (which will behave as if it were set to "false")

Acceptable values are 1, t, T, TRUE, true, True, 0, f, F, FALSE, false, False. Any other values will error. If the provided value is one of the "true" values then pod reaper will do select pods for reaper but will not actually kill any pods. Logging messages will reflect that a pod was selected for reaping and that pod was not killed because the reaper is in dry-run mode.

## Logging

Pod reaper logs in JSON format using a logrus (https://github.com/sirupsen/logrus). 

- rule load: customer messages for each rule are logged when the pod-reaper is starting
- reap cycle: a message is logged each time the reaper starts a cycle.
- pod reap: a message is logged (with a reason for each rule) when a pod is flag for reaping.
- exit: a message is logged when the reaper exits successfully (only is `RUN_DURATION` is specified)

### `LOG_LEVEL`

Default value: Info

Messages this level and above will be logged. Available logging levels: Debug, Info, Warning, Error, Fatal and Panic

### Example Log

```json
{"level":"info","msg":"loaded rule: chaos chance .3","time":"2017-10-18T17:09:25Z"}
{"level":"info","msg":"loaded rule: maximum run duration 2m","time":"2017-10-18T17:09:25Z"}
{"level":"info","msg":"executing reap cycle","time":"2017-10-18T17:09:55Z"}
{"level":"info","msg":"reaping pod","pod":"hello-cloud-deployment-3026746346-bj65k","reasons":["was flagged for chaos","has been running for 3m6.257891269s"],"time":"2017-10-18T17:09:55Z"}
{"level":"info","msg":"reaping pod","pod":"example-pod-deployment-125971999cgsws","reasons":["was flagged for chaos","has been running for 2m55.269615797s"],"time":"2017-10-18T17:09:55Z"}
{"level":"info","msg":"executing reap cycle","time":"2017-10-18T17:10:25Z"}
{"level":"info","msg":"reaping pod","pod":"hello-cloud-deployment-3026746346-grw12","reasons":["was flagged for chaos","has been running for 3m36.054164005s"],"time":"2017-10-18T17:10:25Z"}
{"level":"info","msg":"pod reaper is exiting","time":"2017-10-18T17:10:46Z"}
```

### `LOG_FORMAT`

Default value: Logrus

This environment variable modifies the structured log format for easy ingestion into different logging systems, including Stackdriver via the Fluentd format. Available formats: Logrus, Fluentd

### `RULES`

This is an optional, comma-separated list of rules which should be loaded. If a rule is specified here, it will be loaded even if it does not have a configuration defined in an environment variable. This is used to load rules which only operate on annotations.

Available rules: chaos, container_status, duration, pod_status, unready

## Implemented Rules

### Chaos Chance

Annotation: `pod-reaper/chaos-chance`

Flags a pod for reaping based on a random number generator.

Enabled and configured by setting the environment variable `CHAOS_CHANCE` with a floating point value. A random number generator will generate a value in range `[0,1)` and if the the generated value is below the configured chaos chance, the pod will be flagged for reaping.

Example:

```sh
# every 30 seconds kill 1/100 pods found (based on random chance)
SCHEDULE=@every 30s
CHAOS_CHANCE=.01
```

Remember that pods can be excluded from reaping if the pod has a label matching the pod-reaper's configuration. See the `EXCLUDE_LABEL_KEY` and `EXCLUDE_LABEL_VALUES` section above for more details.

### Container Status

Annotation: `pod-reaper/container-statuses`

Flags a pod for reaping based on a container within a pod having a specific container status.

Enabled and configured by setting the environment variable `CONTAINER_STATUSES` with a coma separated list (no whitespace) of statuses. If a pod is in either a waiting or terminated state with a status in the specified list of status, the pod will be flagged for reaping.

Example:

```sh
# every 10 minutes, kill all pods with a container with a status ImagePullBackOff, ErrImagePull, or Error
SCHEDULE=@every 10m
CONTAINER_STATUSES=ImagePullBackOff,ErrImagePull,Error
```
Note that this will not catch statuses that are describing the entire pod like the `Evicted` status.

### Pod Status

Annotation: `pod-reaper/pod-statuses`

Flags a pod for reaping based on the pod status. 

Enabled and configured by setting the environment variable `POD_STATUSES` with a coma separated list (no whitespace) of statuses. If the pod status in the specified list of status, the pod will be flagged for reaping.

Example:

```sh
# every 10 minutes, kill all pods with status ImagePullBackOff, ErrImagePull, or Error
SCHEDULE=@every 10m
POD_STATUSES=Evicted,Unknown
```
Note that pod status is different than container statuses as it checks the status of the overall pod rather than teh status of containers in the pod. The most obvious use case of this if dealing with `Evicted` pods.

### Duration

Annotation: `pod-reaper/max-duration`

Flags a pod for reaping based on the pods current run duration.

Enabled and configured by setting the environment variable `MAX_DURATION` with a valid go-lang `time.duration` format (example: "1h15m30s"). If a pod has been running longer than the specified duration, the pod will be flagged for reaping.

### Unready

Annotation: `pod-reaper/max-unready`

Flags a pod for reaping based on the time the pod has been unready.

Enabled and configured by setting the environment variable `MAX_UNREADY` with a valid go-lang `time.duration` format (example: "10m"). If a pod has been unready longer than the specified duration, the pod will be flagged for reaping.

## Running Pod-Reapers

### Service Accounts

Pod reaper uses the permissions of the pod's service account to list and delete pods. Unless specified, the service account used will be the default service account in the pod's namespace. By default, and in most cases, the default service account will not have the neccessary permissions to list and delete pods.

- Cluster Wide Permissions: [example](https://github.com/target/pod-reaper/blob/master/examples/cluster-permissions.yml)
- Namespace Specific Permissions: [example](https://github.com/target/pod-reaper/blob/master/examples/namespace-permissions.yml)

### Combining Rules

A pod will only be reaped if ALL rules flag the pod for reaping, but you can achieve reaping on OR logic by simply running another pod-reaper.

For example, in the same pod-reaper container:

```sh
CHAOS_CHANCE=.01
RUN_DURATION=2h
```

Means that 1/100 pods that also have a run duration of over 2 hours will be reaped. If you want 1/100 pods reaped regardless of duration and also want all pods with a run duration of over hours to be reaped, run two pod-reapers. one with: `CHAOS_CHANCE=.01` and another with `RUN_DURATION=2h`.

### Deployments

Multiple pod-reapers can be easily managed and configured with kubernetes deployments. It is encouraged that if you are using deployments, that you leave the `RUN_DURATION` environment variable unset (or "0s") to let the reaper run forever, since the deployment will reschedule it anyway. Note that the pod-reaper can and will reap itself if it is not excluded.

### One Time Runs

You can run run pod-reaper as a one time, limited duration container by usable the `RUN_DURATION` environment variable. An example use case might be wanting to introduce a high degree of chaos into your kubernetes environment for a short duration:

```sh
# 30% chaos chance every 1 minute for 15 minutes
SCHEDULE=@every 1m
RUN_DURATION=15m
CHAOS_CHANCE=.3
```
