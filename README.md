# A GoLang Playground

This serves two purposes:

1. A place where I can learn "real world" GoLang (though my heart tells this is a mistake. GoLang is a disaster programming language.)
2. The outcome being something useful to me, which in this case is an API for recipes.

## Note on Personal Preference

The number of design decisions in GoLang make learning the language
unnecessarily steep and unrewarding: 

- Library hell. Importing `strconv.Atoi` to convert a string to an `int` :eyeroll:.
- Error handling. The garbage `if err != nil` liters the codebase, making it
hard to read. Even with error checks all over the place, my program `panic` on
first actual usage, due to some dereferenced pointers in libraries I don't control.
- Reference handling. Someone call the 1970's, because here we are back in manual
memory management. What an unnecessary overhead to "business logic" development!
- Structural typing by default. Packages, interfaces, structural typing done in the
way that it is implemented in GoLang makes refactoring a nightmare. Trying to quickly
investigate the "structure" of a thing is far more difficult than it needs to be.
- Build tools. Makefiles!? Seriously!?

GoLang is one of the worst programming languages I ever used, both from a build tool
and a syntax perspective. Given a choice, I think learning "effective C" is a
better use of my time. Understanding how to author and read *effective* GoLang is a
postition where life has taken me at the moment, so I press on diligently.

## Building

Thankfully, the *real* build is shoved in a Docker container. You will have to touch
one of the poisonous built-in `go build` for a simple action:

```
go build -o app cmd/recipes/main.go
```

Take a deep breath, because that's all you need to do:

```
docker build -t go-apigw-local .
```

## Running

Once you have the container image created, you can run it with:

```
docker run -it --rm --name go-lambda -p 9000:8080 go-apigw-local
```

## Testing

You can now send "fake" API Gateway requests to the running container using cURL:

```
curl -XPOST "http://localhost:9000/2015-03-31/functions/function/invocations" -d @example.json
```

Where `example.json` is a request modified from [the public documentation][1].

[1]: https://docs.aws.amazon.com/apigateway/latest/developerguide/http-api-develop-integrations-lambda.html#http-api-develop-integrations-lambda.proxy-format