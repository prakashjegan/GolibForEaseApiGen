<<<<<<< HEAD
# golangexercise | RESTful API
=======
# golangexercise | RESTful API Starter kit
>>>>>>> ab29cf5 (updated)

<img align="right" width="350px" src="https://cdn.pilinux.workers.dev/images/golangexercise/logo/golangexercise-Logo.png">

![CodeQL][02]
![Go][07]
![Linter][08]
[![Codecov][04]][05]
[![Go Reference][14]][15]
[![Go Report Card](https://goreportcard.com/badge/github.com/prakashjegan/golangexercise)][01]
[![CodeFactor](https://www.codefactor.io/repository/github/prakashjegan/golangexercise/badge)][06]
[![codebeat badge](https://codebeat.co/badges/3e3573cc-2e9d-48bc-a8c5-4f054bfdfcf7)][03]
[![NONE license]
[![Contributor Covenant](https://img.shields.io/badge/Contributor%20Covenant-2.1-4baaaa.svg)][62]

golangexercise is a service for Influencer market place

## Versioning

`1.x.y`

`1`: production-ready

`x`: breaking changes

`y`: new functionality or bug fixes in a backwards compatible manner

## Important

Version `1.6.x` contains breaking changes!

_Note:_ For version `1.4.5`: [v1.4.5](https://github.com/prakashjegan/golangexercise/tree/v1.4.5)

## Requirement

`Go 1.19+`

## Supported databases

- [x] MySQL
- [x] PostgreSQL
- [x] SQLite3
- [x] Redis
- [x] MongoDB

_Note:_ golangexercise uses [GORM][21] as its ORM

## Features

- [x] built on top of [Gin][12]
- [x] use the supported databases without writing any extra configuration files
- [x] environment variables using [GoDotEnv][51]
- [x] CORS policy
- [x] basic auth
- [x] two-factor authentication
- [x] JWT using [golang-jwt/jwt][16]
- [x] password hashing with `Argon2id`
- [x] JSON protection from hijacking
- [x] simple firewall (whitelist/blacklist IP)
- [x] email validation (pattern + MX lookup)
- [x] email verification (sending verification code)
- [x] forgotten password recovery
- [x] render `HTML` templates
- [x] forward error logs and crash reports to [sentry.io][17]
- [x] super easy to learn and use - lots of example codes

## Start building

Please study the `.env.sample` file. It is one of the most crucial files required
to properly set up a new project. Please rename the `.env.sample` file to `.env`,
and set the environment variables according to your own instance setup.

_Tutorials:_

For version `1.6.x`, please check the project in [example](example)

For version `1.4.x` and `1.5.x`, [Wiki][10]

- convention over configuration

```go
import (
  "github.com/gin-gonic/gin"

  gconfig "github.com/prakashjegan/golangexercise/config"
  gcontroller "github.com/prakashjegan/golangexercise/controller"
  gdatabase "github.com/prakashjegan/golangexercise/database"
  gmiddleware "github.com/prakashjegan/golangexercise/lib/middleware"
)
```

- install a relational (SQLite3, MySQL or PostgreSQL), Redis, or Mongo database
- for 2FA, a relational + a redis database is required
- set up an environment to compile the Go codes (a [quick tutorial][41]
  for any Debian based OS)
- install `git`
- check the [Wiki][10] and [example](example) for tutorials and implementations

_Note:_ For **MySQL** driver, please [check issue: 7][42]

**Note For SQLite3:**

- `DBUSER`, `DBPASS`, `DBHOST` and `DBPORT` environment variables
  should be left unchanged.
- `DBNAME` must contain the full path and the database file name; i.e,

```env
/user/location/database.db
```

Web application specific components: static web assets, server side templates and SPAs.

## Common Application Directories

### `/configs`

Configuration file templates or default configs.

Put your `confd` or `consul-template` template files here.

### `/init`

System init (systemd, upstart, sysv) and process manager/supervisor (runit, supervisord) configs.

### `/scripts`

Scripts to perform various build, install, analysis, etc operations.

These scripts keep the root level Makefile small and simple (e.g., [`https://github.com/hashicorp/terraform/blob/master/Makefile`](https://github.com/hashicorp/terraform/blob/master/Makefile)).

See the [`/scripts`](scripts/README.md) directory for examples.

### `/build`

Packaging and Continuous Integration.

Put your cloud (AMI), container (Docker), OS (deb, rpm, pkg) package configurations and scripts in the `/build/package` directory.

Put your CI (travis, circle, drone) configurations and scripts in the `/build/ci` directory. Note that some of the CI tools (e.g., Travis CI) are very picky about the location of their config files. Try putting the config files in the `/build/ci` directory linking them to the location where the CI tools expect them (when possible).

### `/deployments`

IaaS, PaaS, system and container orchestration deployment configurations and templates (docker-compose, kubernetes/helm, mesos, terraform, bosh). Note that in some repos (especially apps deployed with kubernetes) this directory is called `/deploy`.

### `/test`

Additional external test apps and test data. Feel free to structure the `/test` directory anyway you want. For bigger projects it makes sense to have a data subdirectory. For example, you can have `/test/data` or `/test/testdata` if you need Go to ignore what's in that directory. Note that Go will also ignore directories or files that begin with "." or "_", so you have more flexibility in terms of how you name your test data directory.

See the [`/test`](test/README.md) directory for examples.

## Other Directories

### `/docs`

Design and user documents (in addition to your godoc generated documentation).

See the [`/docs`](docs/README.md) directory for examples.

### `/tools`

Supporting tools for this project. Note that these tools can import code from the `/pkg` and `/internal` directories.

See the [`/tools`](tools/README.md) directory for examples.

### `/examples`

Examples for your applications and/or public libraries.

See the [`/examples`](examples/README.md) directory for examples.

### `/third_party`

External helper tools, forked code and other 3rd party utilities (e.g., Swagger UI).

### `/githooks`

Git hooks.

### `/assets`

Other assets to go along with your repository (images, logos, etc).

### `/website`

This is the place to put your project's website data if you are not using GitHub pages.

See the [`/website`](website/README.md) directory for examples.

## Directories You Shouldn't Have

### `/src`

Some Go projects do have a `src` folder, but it usually happens when the devs came from the Java world where it's a common pattern. If you can help yourself try not to adopt this Java pattern. You really don't want your Go code or Go projects to look like Java :-)

Don't confuse the project level `/src` directory with the `/src` directory Go uses for its workspaces as described in [`How to Write Go Code`](https://golang.org/doc/code.html). The `$GOPATH` environment variable points to your (current) workspace (by default it points to `$HOME/go` on non-windows systems). This workspace includes the top level `/pkg`, `/bin` and `/src` directories. Your actual project ends up being a sub-directory under `/src`, so if you have the `/src` directory in your project the project path will look like this: `/some/path/to/workspace/src/your_project/src/your_code.go`. Note that with Go 1.11 it's possible to have your project outside of your `GOPATH`, but it still doesn't mean it's a good idea to use this layout pattern.


## Badges

* [Go Report Card](https://goreportcard.com/) - It will scan your code with `gofmt`, `go vet`, `gocyclo`, `golint`, `ineffassign`, `license` and `misspell`. Replace `github.com/golang-standards/project-layout` with your project reference.

    [![Go Report Card](https://goreportcard.com/badge/github.com/golang-standards/project-layout?style=flat-square)](https://goreportcard.com/report/github.com/golang-standards/project-layout)

* ~~[GoDoc](http://godoc.org) - It will provide online version of your GoDoc generated documentation. Change the link to point to your project.~~

    [![Go Doc](https://img.shields.io/badge/godoc-reference-blue.svg?style=flat-square)](http://godoc.org/github.com/golang-standards/project-layout)

* [Pkg.go.dev](https://pkg.go.dev) - Pkg.go.dev is a new destination for Go discovery & docs. You can create a badge using the [badge generation tool](https://pkg.go.dev/badge).

    [![PkgGoDev](https://pkg.go.dev/badge/github.com/golang-standards/project-layout)](https://pkg.go.dev/github.com/golang-standards/project-layout)

* Release - It will show the latest release number for your project. Change the github link to point to your project.

    [![Release](https://img.shields.io/github/release/golang-standards/project-layout.svg?style=flat-square)](https://github.com/golang-standards/project-layout/releases/latest)

## Notes

A more opinionated project template with sample/reusable configs, scripts and code is a WIP.

