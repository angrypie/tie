(under development)

No more single line of code for API, requests, discovery etc. It must be generated.

See [examples](example/).  Just call `tie` in any example directory and then play around with generated binaries :)

## How it works

Be careful, due errors `tie` may leave folders prefixed by `tie_`.

Use `tie clean` to remove `*.run` files.

Call `tie` without tie.yml configuration:
- Every top level directory considered as package for processing
- Every not main package will be transformed to an RPC service.
- Every public method call to future RPC service will be changed to RPC call.

Call `tie` with configuaration. Example tie.yml:

```yml
services:
  - name: 'github.com/angrypie/tie/example/basic/sum'
    alias: 'sum'
  - name: 'github.com/angrypie/tie/example/basic/cli'
    alias: 'cli'
```

HTTP API tie.yml configuration: 

```yml
services:
  - name: 'github.com/angrypie/tie/example/basic/sum'
    alias: 'sum'
    type: 'httpOnly'
    port: '8111' 
```

Place this file in any directory and call `tie` than execute `sum.run` to start RPC service. Try to access HTTP API:

```bash
  curl -X POST -H 'Content-Type: application/json' localhost:8111/sum -d '{"a":20, "b":22}'
```


