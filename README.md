# loago-worker

Loago is a scaleable webapp loadtest service based on Chrome/Chomium browser fleets.

## Description

Use real Chromium-based browsers to loadtest your webapp/website.

Loago enables website load tests using a fleet of Chromium browser processes to reliable
recreate user behaviour.  
Real users use real browsers to visit webapps. And instead of measuring the maximum amount
of requests your webapp can handle, it lets you decide how many users you wan't to simulate concurrently.  
This gives devops and site reliability engineers an indicator of how many users their website can handle,
much better than simply saying how many HTTP requests it can make.  
To accurately simulate real browser behaviour, a real browser is used: Chromium (or Chromium based browsers).  
This makes sure the applied load behaves just like a real user, including browser caching and parallel asset loading.

The project is split into two apps: The Loago CLI app called **instructor** and the server-side service called **worker**.  
The instructor coordinates multiple workers via network connection.
This allows for horizontal scaling and geographically spreaded testing.


## Key features
- Efficient client-server communication via protobuf+gRPC including TLS encryption and authentication
- Horizontal scaling, since one instructor can handle multiple workers
- Written in Go
- Scales as much as your memory does, though it's not as bad as you might think
- Random, but weighted, HTTP requests on specific URL's
- Every response contains TTFB, HTTP status code and message and will be send to the instructor

## Project status

:warning:

The loago project is in it's (very) early stage and should not be used on real websites or webapps!  
But since there is no instructor available yet, this should be pretty inconvenient anyway ;-).

## Installation

Right now only with `go get`:

```shell script
go get github.com/dkorittki/loago-worker
```

## Usage

Call `loagoworker` from command line with these flags:

`-adress`: listen address, e.g. `127.0.0.1` or `0.0.0.0` (default)  
`-port`: listen port (default 50051)  
`-cert`: path to TLS certificate  
`-key`: path to TLS private key  
`-secret`: basic auth secret for authentication
