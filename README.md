# virgo

virgo is a concurrent and tiny HTTP request client.

## Installation

Get the package.

```shell script
go get github.com/shellrausch/virgo
```

## Usage

Concurrent http requests.

```go
urls := []string{ "https://google.com", "https://example.com", "https://github.com" }
resultCh := make(chan *virgo.Result, 16)

v := virgo.New()
go v.Start(urls, resultCh)

for r := range resultCh {
    if r.Err == nil {
        // do something with the response
    }   
}
```

Concurrent http request and custom options for all requests.

```go
urls := []string{ "https://google.com", "https://example.com", "https://github.com" }
resultCh := make(chan *virgo.Result, 16)

o := options.New()
o.UserAgent = "VirgoBot"
o.TimeoutMs = 5000
o.Concurrency = 16
o.Method = "POST"
o.Body = []byte("{ key : value }")

v := virgo.New()
v.SetOptions(o)
go v.Start(urls, resultCh)

for r := range resultCh {
    if r.Err == nil {
        // do something with the response
    }
}
```
