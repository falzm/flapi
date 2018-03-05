# FLAPI – A FLapping API

## Introduction

FLAPI is a proof-of-concept web application back-end exposing a configurable RESTful API. It is intended for educational purposes, where users can simulate a *microservice* architecture and experiment different forms of "chaos" by injecting arbitraty latency and errors on targeted API endpoints. Additionally, the backend is instrumented with HTTP access logs and metrics (Prometheus format).

## Build

Build requirements:

* [Go language](https://golang.org/) compiler >= 1.9
* [gb](https://getgb.io/) utility

To compile FLAPI, execute `make`. The resulting binary should appear in the `bin` directory.

## Configuration

The `flapi` executable reads its configuration from a YAML-formatted file.

### API Endpoints

API endpoints are declared in the `api_endpoints` top-level section. An API endpoint is defined by an HTTP method (e.g. `GET`, `POST`...) and a URL path relative to `/api` (i.e. `/a`, `/x/y/z`...).

There are 2 types of API endpoints: local and chained:

* A *local* endpoint returns a static response, configurable by specifying a `response_status` integer between `100` and `599` and an optional `response_body` string.
* A *chained* endpoint performs HTTP sub-requests to a *chain* of targets, and returns the responses received. A chain target is defined by a `method` string parameter describing the target HTTP method to use, and a `url` string parameter describing the target URL.

Example endpoints definition:

```yaml
---
api_endpoints:
### POST /api/a
- method: POST
  route: /a
  response_status: 201
### GET /api/a
- method: GET
  route: /a
  response_status: 200
  response_body: A
### PUT /api/b
- method: PUT
  route: /b
  response_status: 200
  response_body: B
### GET /api/c
- method: GET
  route: /c
  response_status: 200
  response_body: C
### GET /api/x
- method: GET
  route: /x
  chain:
  - method: GET
    url: http://localhost:8001/api/y
  - method: GET
    url: http://localhost:8002/api/z
```

## API documentation

Note: you can inject both delay and error to an endpoint.

### Endpoints Management

#### `GET /`

Retrieve the list of configured endpoints (in JSON format).

### Delay Injection

The delay injection management API has 2 different contexts:

* `base` allowing to set a delay that will be injected for all requests
* `endpoint` for targeting specific API endpoints

#### `GET /delay/base`

Retrieve base delay injection specification.

#### `PUT /delay/base`

Set base delay injection specification.

URL parameters:

* `method`: target API endpoint HTTP method (e.g. `GET`, `POST`...)
* `route`: target API endpoint URL route (e.g. `/api/a`)
* `duration`: delay duration in milliseconds (e.g. `100`)

#### `GET /delay/endpoint`

Retrieve delay injection specification for a specific API endpoint.

URL parameters:

* `method`: target API endpoint HTTP method
* `route`: target API endpoint URL route

#### `PUT /delay/endpoint`

URL parameters:

* `method`: target API endpoint HTTP method
* `route`: target API endpoint URL route
* `duration`: delay duration in milliseconds
* `probability` (optional, default `1`): delay injection probability between 0 and 1 (e.g. `0.5`)

#### `DELETE /delay/endpoint`

URL parameters:

* `method`: target API endpoint HTTP method
* `route`: target API endpoint URL route

### Error Injection

#### `GET /error/endpoint`

Retrieve error injection specification for a specific API endpoint.

URL parameters:

* `method`: target API endpoint HTTP method
* `route`: target API endpoint URL route

#### `PUT /error/endpoint`

URL parameters:

* `method`: target API endpoint HTTP method
* `route`: target API endpoint URL route
* `status_code` (optional, default `500`): HTTP status code to return
* `message` (optional): HTTP response body to return (optional)
* `probability` (optional, default `1`): delay injection probability between 0 and 1

#### `DELETE /error/endpoint`

URL parameters:

* `method`: target API endpoint HTTP method
* `route`: target API endpoint URL route

### Usage

`flapi` command usage:

```
$ flapi -h
Usage of flapi:
  -bind-addr string
    	network [address]:port to bind to (default ":8000")
  -config string
    	path to configuration file (default "flapi.yaml")
  -help
    	display this help and exit
  -log-level string
    	logging level (default "info")
  -version
    	display version and exit
```

For the following configuration:

```yaml
api_endpoints:
- method: POST
  route: /a
  response_status: 201
- method: GET
  route: /a
  response_status: 200
  response_body: A
- method: GET
  route: /b
  response_status: 200
  response_body: B
- method: PUT
  route: /c
  response_status: 202
- method: GET
  route: /c
  response_status: 200
  response_body: C
```

List configured endpoints:

```json
$ curl localhost:8000/ | jq .
[
  {
    "method": "POST",
    "response_body": "",
    "response_status": 201,
    "route": "/api/a"
  },
  {
    "method": "GET",
    "response_body": "A",
    "response_status": 200,
    "route": "/api/a"
  },
  {
    "method": "GET",
    "response_body": "B",
    "response_status": 200,
    "route": "/api/b"
  },
  {
    "method": "PUT",
    "response_body": "",
    "response_status": 202,
    "route": "/api/c"
  },
  {
    "method": "GET",
    "response_body": "C",
    "response_status": 200,
    "route": "/api/c"
  }
]
```

Test configured endpoints requests results:

```
$ curl -i -X POST localhost:8000/api/a
HTTP/1.1 201 Created
X-Flapi-Version: 0.1.0
Date: Mon, 05 Mar 2018 10:26:28 GMT
Content-Length: 1
Content-Type: text/plain; charset=utf-8

$ curl -i localhost:8000/api/a
HTTP/1.1 200 OK
X-Flapi-Version: 0.1.0
Date: Mon, 05 Mar 2018 10:26:34 GMT
Content-Length: 2
Content-Type: text/plain; charset=utf-8

A

$ curl -i localhost:8000/api/b
HTTP/1.1 200 OK
X-Flapi-Version: 0.1.0
Date: Mon, 05 Mar 2018 10:27:24 GMT
Content-Length: 2
Content-Type: text/plain; charset=utf-8

B

$ curl -i -X PUT localhost:8000/api/c
HTTP/1.1 202 Accepted
X-Flapi-Version: 0.1.0
Date: Mon, 05 Mar 2018 10:28:16 GMT
Content-Length: 1
Content-Type: text/plain; charset=utf-8
```

#### Delay injection example

Inject a 500ms base latency:

```
$ curl -i -X PUT localhost:8000/delay/base?duration=500
HTTP/1.1 204 No Content
Date: Mon, 05 Mar 2018 10:11:45 GMT

$ curl -i localhost:8000/delay/base
HTTP/1.1 200 OK
Date: Mon, 05 Mar 2018 10:11:46 GMT
Content-Length: 5
Content-Type: text/plain; charset=utf-8

500ms

$ curl -i -w '---\ntime_total=%{time_total}s\n' localhost:8000/api/a
HTTP/1.1 200 OK
X-Flapi-Version: 0.1.0
Date: Mon, 05 Mar 2018 10:13:36 GMT
Content-Length: 2
Content-Type: text/plain; charset=utf-8

A
---
time_total=0.510756s
```

Inject an additional 300ms delay specific to endpoint `/api/a`:

```
$ curl -i -X PUT 'localhost:8000/delay/endpoint?method=GET&route=/api/a&duration=300'
HTTP/1.1 204 No Content
Date: Mon, 05 Mar 2018 10:15:58 GMT

$ curl -i -w '---\ntime_total=%{time_total}s\n' localhost:8000/api/a
HTTP/1.1 200 OK
X-Flapi-Version: 0.1.0
Date: Mon, 05 Mar 2018 10:16:03 GMT
Content-Length: 2
Content-Type: text/plain; charset=utf-8

A
---
time_total=0.816943s
```

Reset base latency to zero:

```
$ curl -i -X PUT localhost:8000/delay/base?duration=0
HTTP/1.1 204 No Content
Date: Mon, 05 Mar 2018 10:17:23 GMT

$ curl -i localhost:8000/delay/base
HTTP/1.1 200 OK
Date: Mon, 05 Mar 2018 10:18:43 GMT
Content-Length: 3
Content-Type: text/plain; charset=utf-8

0s

$ curl -i -w '---\ntime_total=%{time_total}s\n' localhost:8000/api/a
HTTP/1.1 200 OK
X-Flapi-Version: 0.1.0
Date: Mon, 05 Mar 2018 10:17:25 GMT
Content-Length: 2
Content-Type: text/plain; charset=utf-8

A
---
time_total=0.306930s
```

Reset endpoint `/api/a` latency to zero:

```
$ curl -i -X DELETE 'localhost:8000/delay/endpoint?method=GET&route=/api/a'
HTTP/1.1 204 No Content
Date: Mon, 05 Mar 2018 10:19:59 GMT

$ curl -i -w '---\ntime_total=%{time_total}s\n' localhost:8000/api/a
HTTP/1.1 200 OK
X-Flapi-Version: 0.1.0
Date: Mon, 05 Mar 2018 10:20:02 GMT
Content-Length: 2
Content-Type: text/plain; charset=utf-8

A
---
time_total=0.005079s
```

#### Error injection example

Inject a `HTTP 504` error to the `/api/a` endpoint:

```
$ curl -i -X PUT 'localhost:8000/error?method=GET&route=/api/a&status_code=504'
HTTP/1.1 204 No Content
Date: Mon, 05 Mar 2018 11:42:41 GMT

$ curl -i localhost:8000/api/a
HTTP/1.1 504 Gateway Timeout
Content-Type: text/plain; charset=utf-8
X-Content-Type-Options: nosniff
Date: Mon, 05 Mar 2018 11:43:07 GMT
Content-Length: 1
```

Remove error injection:

```
$ curl -i -X DELETE 'localhost:8000/error?method=GET&route=/api/a'
HTTP/1.1 204 No Content
Date: Mon, 05 Mar 2018 11:43:26 GMT

$ curl -i localhost:8000/api/a
HTTP/1.1 200 OK
X-Flapi-Version: 0.1.0
Date: Mon, 05 Mar 2018 11:43:28 GMT
Content-Length: 2
Content-Type: text/plain; charset=utf-8

A
```