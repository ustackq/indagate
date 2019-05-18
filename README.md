# Indagate

Indagate is a fully opensoure product, which construct on mircoservices but has an opinion on the best way to hold webcast. This project not only helps you build a public discuz which has supported  automation install ,but It also provide admin to manage and troubleshoot them.

Goals of the project:

- Install automation
- Secure by default (uses TLS, RBAC by default, OIDC AuthN, etcd、AMI)
- Web discuz
- Run on such OS: Centos, Ubuntu
- k8s supported

## Getting Started

**To use a tested release** on a supported platform, follow the links below.

**To hack or modify** the templates or add a new platform, use the scripts in this repo to boot and tear down clusters.

### Architecture overview
![](dcos/architecture.png)
See the architecture below:



### Library

#### Microservices standard library

- [go-kit/kit](https://github.com/go-kit/kit)

#### Logging

- [uber-go/zap](https://github.com/uber-go/zap)

#### Store

- [etcd-io/bblot](https://github.com/etcd-io/bbolt)

#### Auth

- [jwt](https://github.com/dgrijalva/jwt-go)

#### Monitoring

- [prometheus](https://github.com/prometheus/prometheus)

#### Tracing

- [opencensus](https://github.com/census-instrumentation/opencensus-go)


## Documentation

Dashboard documentation can be found on [Wiki](https://github.com/ustackq/indagate/wiki) pages, it includes:

* Common: Entry-level overview

* Install Guide: [Installation](https://github.com/ustackq/indagate/docs/Installation), [Istaller Dashboard](
https://github.com/ustackq/yunus/docs/Accessing-dashboard) and more for users

* Manager Guide: [Management](https://github.com/ustackq/indagate/docs/management), [Management Dashboard](
https://github.com/ustackq/indagate/docs/Accessing-dashboard) and more for users

* Developer Guide: [Getting Started](https://github.com/ustackq/indagate/docs/Getting-started), [Dependency
Management](https://github.com/ustackq/indagate/docs/Dependency-management) and more for anyone interested in contributing

## License

The work done has been licensed under Apache License 2.0. The license file can be found [here](LICENSE). You can find
out more about the license at [www.apache.org/licenses/LICENSE-2.0](//www.apache.org/licenses/LICENSE-2.0).



