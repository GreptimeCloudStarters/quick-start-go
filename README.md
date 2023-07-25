# Introduction

This is a quick start demo for [GreptimeCloud](https://greptime.cloud/). It collects the system metric data such as CPU and memory usage through Opentelemetry and sends the metrics to GreptimeCloud. You can view the metrics on the GreptimeCloud dashboard.

## Quick Start

Use the following command line to start sending metrics without cloning the project:

```shell
go run github.com/GreptimeCloudStarters/quick-start-go@latest -host=<host> -db=<dbname> -username=<username> -password=<password>
```

Or clone the project and run the following command line:

```shell
go run . -host=<host> -db=<dbname> -username=<username> -password=<password>
```

## Release

1. Update the version in `app.go`.
2. Commit and push code.
3. Create a tag with the version and push it to the remote repository.

```shell
git tag v<major>.<minor>.<patch>
git push origin v<major>.<minor>.<patch>
```

4. Update pkg.go.dev via `GOPROXY=proxy.golang.org go list -m github.com/GreptimeCloudStarters/quick-start-go@v<major>.<minor>.<patch>`
5. Write change log in [Github Release](https://github.com/GreptimeCloudStarters/quick-start-go/tags).
