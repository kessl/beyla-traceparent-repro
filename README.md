# OBI traceparent duplication

OBI duplicates the `traceparent` header when a Go `httputil.ReverseProxy` forwards requests within the same host. The fix from open-telemetry/opentelemetry-ebpf-instrumentation#1162 does not prevent it.

`main.go` builds a single binary that runs as either a `httputil.ReverseProxy` on :80 or a backend on :3000 that prints incoming `Traceparent` header values. They run as separate processes. Direct requests to :3000 (bypassing the proxy) do not duplicate.

## Prerequisites

Pull the OBI image:

```bash
docker pull ghcr.io/open-telemetry/opentelemetry-ebpf-instrumentation/ebpf-instrument:v0.6.0
```

Build the proxy:

```bash
go build -o proxy .
```

## Reproduce

Start OBI:

```bash
sudo docker run --rm --privileged --pid=host --network=host \
  -e OTEL_EXPORTER_OTLP_ENDPOINT=http://localhost:4318 \
  -e OTEL_EBPF_OPEN_PORT=80,3000 \
  -e OTEL_EBPF_BPF_CONTEXT_PROPAGATION=headers \
  -e OTEL_EBPF_BPF_TRACK_REQUEST_HEADERS=true \
  -e OTEL_EBPF_TRACE_PRINTER=text \
  ghcr.io/open-telemetry/opentelemetry-ebpf-instrumentation/ebpf-instrument:v0.6.0
```

In separate terminals, start the backend and proxy processes:

```bash
sudo ./proxy backend
sudo ./proxy
```

Wait for OBI to attach, then:

```bash
for i in 1 2 3 4 5; do
  curl -H 'traceparent: 00-aaaabbbbccccddddeeeeffffaaaabbbb-1111222233334444-01' http://localhost:80/
done
```

## Expected

OBI should replace the `traceparent` with its own span ID (the client span for the :80 → :3000 hop):

```
traceparent count=1 values=00-aaaabbbbccccddddeeeeffffaaaabbbb-<obi-span>-01
```

The original `1111222233334444` becomes the parent of OBI's :80 server span, and the backend receives only the propagated client span.

## Actual

```
traceparent count=2 values=00-aaaabbbbccccddddeeeeffffaaaabbbb-bb2b6a109723a171-01 | 00-aaaabbbbccccddddeeeeffffaaaabbbb-1111222233334444-01
```

Two `traceparent` values: one injected by OBI plus the foworiginal forwarded by `httputil.ReverseProxy`.
