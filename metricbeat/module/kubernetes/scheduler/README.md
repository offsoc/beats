# Scheduler Stats

## Version history

- December 2022, `v1.25.x`

## Resources

- [Process metrics](https://github.com/kubernetes/kubernetes/blob/master/vendor/github.com/prometheus/client_golang/prometheus/process_collector.go)
- [Scheduler metrics](https://github.com/kubernetes/kubernetes/blob/master/pkg/scheduler/metrics/metrics.go)
- [Rest client metrics](https://github.com/kubernetes/component-base/blob/master/metrics/prometheus/restclient/metrics.go)
- [Workqueue metrics](https://github.com/kubernetes/kubernetes/blob/master/staging/src/k8s.io/component-base/metrics/prometheus/workqueue/metrics.go)
- [Metrics general information](https://kubernetes.io/docs/reference/instrumentation/metrics/)


## Metrics insight

Metrics used are either stable (not explicit) or alpha (explicit).

- process_cpu_seconds_total
- process_resident_memory_bytes
- process_virtual_memory_bytes
- process_open_fds
- process_start_time_seconds
- process_max_fds


- rest_client_requests_total (alpha)
  - code
  - host
  - method
- rest_client_response_size_bytes (alpha)
  - host
  - verb
- rest_client_request_size_bytes (alpha)
  - host
  - verb
- rest_client_request_duration_seconds (alpha)
  - host
  - verb


- workqueue_longest_running_processor_seconds (alpha)
  - name
- workqueue_unfinished_work_seconds (alpha)
  - name
- workqueue_adds_total (alpha)
  - name
- workqueue_depth (alpha)
  - name
- workqueue_retries_total (alpha)
  - name
- workqueue_work_duration_seconds (alpha)
  - name


- scheduler_pending_pods
  - queue
- scheduler_preemption_victims
- scheduler_preemption_attempts_total
- scheduler_scheduling_attempt_duration_seconds
  - profile
  - result


- leader_election_master_status (alpha)
  - name

## Setup environment for manual tests

Kubernetes scheduler will usually run at every master node, but that might not be the case. It could be executed as a host process or an in-cluster pod.

- If host process (for example, systemd), metricbeat should be running at that same node gathering data from the scheduler.
- If executing as a pod:
    - A metricbeat instance can be also executed using the same affinity and deployment object (deployment, daemonset, ...) as the kubernetes scheduler.
    - A metricbeat instance can be launched as a sidecar container

## Generating expectation files

In order to support a new Kubernetes release you'll have to generate new expectation files for this module in `_meta/test`. For that, start by deploying a new kubernetes cluster on the required Kubernetes version, for example:

```bash
kind create cluster --image kindest/node:v1.32.0
```

After that, you can apply the `kubernetes.yml` file from the root of the kubernetes module:

```bash
kubectl apply -f kubernetes.yml
```

This is required since accessing the scheduler metrics requires additional permissions.

Then you can expose the scheduler api:

```bash
kubectl -n kube-system port-forward pod/kube-scheduler-kind-control-plane 10259
```

Save the metrics output from `https://localhost:10259/metrics` to a new `_meta/test/metrics.x.xx` file.

```bash
curl -k https://localhost:10259/metrics > _meta/test/metrics.x.xx
```

Run the following commands to generate and test the expected files:

```bash
cd metricbeat/module/kubernetes/scheduler
# generate the expected files
go test ./... --data
# test the expected files
go test ./...
```











