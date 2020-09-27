# Loago

Loago is a scaleable webapp loadtest utility based on Chrome/Chomium browser fleets.

## Description

Use real Chromium-based browsers to loadtest your webapp/website.

Loago enables website load tests using a fleet of Chromium browser processes to reliable
recreate user behaviour.  
Real users use real browsers to visit webapps. And instead of measuring the maximum amount
of requests your webapp can handle, it lets you decide how many users you wan't to simulate concurrently.  
This gives devops and SRE's an indicator of how many users their webapp can handle,
much better than simply saying how many HTTP requests it can take without any real-world cohesion.  
To accurately simulate real browser behaviour, a real browser is used: Chromium (or Chromium based browsers implementing the DevTools API).  
This makes sure the applied load behaves just like a real user, including browser caching and parallel asset loading.

Loago provides two modes: `instructor` and `worker`.

In instructor-mode you configure a loadtest, instruct one or more workers to
perform this loadtest and save the results.

In worker-mode Loago actually performs the requests coming from an Loago
instructor instance.

This allows for horizontal, geographically spread load scaling.


## Key features
- Efficient client-server communication via protobuf+gRPC including TLS encryption
- Authentication currently implemented as basic auth token
- Horizontal scaling, since one instructor can handle multiple workers
- Written in Go
- Scales as much as your memory does, though it's not as bad as you might think
- Random, but weighted, HTTP requests on specific URL's
- Every response contains TTFB, HTTP status code and message and will be send to the instructor

## Project status

:warning:

The loago project is in it's (very) early stage and should not be used on real websites or webapps!  
But since there is no instructor mode available yet, this should be pretty inconvenient anyway ;-).

## Installation

Install via `go get`:

```shell script
go get github.com/dkorittki/loago
```

## Usage

## worker mode

Call `loago serve` from command line with these flags, to start Loago in worker mode:

`-adress`: listen address, e.g. `127.0.0.1` or `0.0.0.0` (default)  
`-port`: listen port (default 50051)  
`-cert`: path to TLS certificate  
`-key`: path to TLS private key  
`-secret`: basic auth secret for authentication
