# common-healthcheck [![Build Status][ci-img]][ci] [![Coverage Status][cov-img]][cov] [![GoDoc][godoc-img]][godoc] [![Go Report Card][report-img]][report]

All Cloudtrust microservices send metrics to InfluxDB, traces to Jaeger, technical logs to Redis, and errors to Sentry. The health of all those components is monitored with health checks. Because they are used in every microservice, this library with all the common health checks was created.

[ci-img]: https://travis-ci.org/cloudtrust/common-healthcheck.svg?branch=master
[ci]: https://travis-ci.org/cloudtrust/common-healthcheck
[cov-img]: https://coveralls.io/repos/github/cloudtrust/common-healthcheck/badge.svg?branch=master
[cov]: https://coveralls.io/github/cloudtrust/common-healthcheck?branch=master
[godoc-img]: https://godoc.org/github.com/cloudtrust/common-healthcheck?status.svg
[godoc]: https://godoc.org/github.com/cloudtrust/common-healthcheck
[report-img]: https://goreportcard.com/badge/github.com/cloudtrust/common-healthcheck
[report]: https://goreportcard.com/report/github.com/cloudtrust/common-healthcheck
