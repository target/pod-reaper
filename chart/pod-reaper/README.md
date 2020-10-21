# Pod Reaper Helm Chart

This is a helm templated deployment for pod-reaper.

## Chart details

This chart will install a target/pod-reaper for an arbitrary number of configurations. 

Installing the chart.
```bash
helm upgrade --install <deployment name> . -f values.yaml --namespace <namespace>
```

| Parameter                        | Description                                        | Default                       |
| -------------------------------- | -------------------------------------------------- | ----------------------------- |
| `image.repository`               | Image file location                                | `target/pod-reaper`         |
| `image.tag`                      | Image tag                                          | `2.8.0`                          |
| `resources`                      | Provides resource limits                           | ```
|                                  |                                                    | limits:
|                                  |                                                    |   cpu: 30m
|                                  |                                                    |   memory: 30Mi
|                                  |                                                    | requests:
|                                  |                                                    |   cpu: 20m
|                                  |                                                    |   memory: 20Mi
|                                  |                                                    | ```                         |

...


