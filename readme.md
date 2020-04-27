(under development)

No more single line of code for API, requests, discovery etc. It must be generated.

See [examples](example/).  Just execute `tie` in any example directory and then play around with generated binaries :)

## How it works

***Be careful, due errors `tie` may leave `tie_modules` directories***

#### Turn package to RPC API

Go to [example/basic](example/basic/) and execute `tie` there.
It will produce two binaries `sum.run` and `cli.run`

- Every top level directory considered as package for processing.
- Every not main package will be transformed to an RPC service.
- Every public method call to future RPC service will be changed to RPC call.

Execute `sum.run` to start RPC service. Try to add two numbers using `cli.run`:

```bash
./cli.run 18 24
#18+24=42
```

#### Turn package to HTTP API

Execute `tie` inside package directory will turn this package to HTTP API.

Go to [example/basic/sum](example/basic/sum/) and execute `tie` there.
Use newly created `sum.run` to start HTTP API service:

```bash
export PORT=8080 #if not set, random port will be used
./sum.run
```

Try to access HTTP API via curl:

```bash
curl -X POST -H 'Content-Type: application/json' localhost:8111/sum -d '{"a":20, "b":22}'
#20+22=42
```


#### Clean binaries

Use `tie clean` to remove `*.run` files.

## TODO

- request and response DTO
- receiver concept
- step by step guide

