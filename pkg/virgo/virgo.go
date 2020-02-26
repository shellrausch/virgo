package virgo

import (
	"bytes"
	"io"
	"io/ioutil"
	"net/http"
	"sync"
	"time"

	"github.com/shellrausch/virgo/pkg/options"
)

// Client is just used for the selectors.
type Client struct{}

// Result is a wrapper around a plain HTTP response.
type Result struct {
	Response *http.Response

	// Typically the body is a part of the HTTP response.
	// We need to save it here explicitly otherwise it will wiped after the connection is closed.
	Body []byte
	Err  error
}

// request contains all information needed to make a plain HTTP request.
type request struct {
	url    string
	method string
	body   []byte
	header map[string]string
}

var httpClient http.Client
var o *options.Options

// New initializes the client.
func New() *Client {
	o = options.New()

	httpClient = initHTTPClient()
	return &Client{}
}

// SetClient sets a custom http client (e.g. proxy, socks5, ...).
func (*Client) SetClient(newClient http.Client) {
	httpClient = newClient
}

// SetOptions sets custom Options for a request.
func (*Client) SetOptions(newOptions *options.Options) {
	o = newOptions

	// The timeout is set for client and not for a request.
	httpClient.Timeout = time.Duration(newOptions.TimeoutMs) * time.Millisecond
}

// SetTimeout sets a global timeout which affects all http requests.
func (*Client) SetTimeout(timeoutMs int) {
	// The timeout is set for client and not for a request.
	httpClient.Timeout = time.Duration(timeoutMs) * time.Millisecond
}

// GetOptions returns the currently set Options.
func (*Client) GetOptions() *options.Options {
	return o
}

// Start starts to execute the HTTP requests which are provided as a list of urls.
// resultCh sends a result when a HTTP response is available.
func (*Client) Start(urls []string, resultCh chan *Result) {
	// Number of global Go routines.
	concurrencyWg := new(sync.WaitGroup)

	// Queued request.
	requestQueueCh := produceRequests(urls)

	for i := 0; i < o.Concurrency; i++ {
		concurrencyWg.Add(1)

		// Starts a worker
		go func() {
			for {
				req, open := <-requestQueueCh
				if !open {
					concurrencyWg.Done()
					return
				}
				consumeRequest(req, resultCh)
			}
		}()
	}

	// Order matters for a proper termination of all Go routines.
	close(requestQueueCh)
	concurrencyWg.Wait()
	close(resultCh)
}

// produceRequests produces HTTP request structs and puts them in a queue.
func produceRequests(urls []string) chan *request {
	requestQueueCh := make(chan *request, len(urls))

	for _, url := range urls {
		requestQueueCh <- &request{
			url:    url,
			body:   o.Body,
			method: o.Method,
			header: o.Header,
		}
	}

	return requestQueueCh
}

// consumeRequest takes a given request and invokes the plain HTTP request.
func consumeRequest(r *request, resultCh chan *Result) {
	res, body, err := invokeRequest(r)

	resultCh <- &Result{
		Response: res,
		Body:     body,
		Err:      err,
	}
}

// invokeRequest does the raw HTTP request with the provided set of options.
func invokeRequest(r *request) (*http.Response, []byte, error) {
	var req *http.Request
	var err error

	req, err = http.NewRequest(r.method, r.url, bytes.NewReader(r.body))
	if err != nil {
		return nil, nil, err
	}

	if o.UserAgent != "" {
		req.Header.Set("User-Agent", o.UserAgent)
	}

	if o.Cookie != "" {
		req.Header.Set("Cookie", o.Cookie)
	}

	for k, v := range r.header {
		req.Header.Set(k, v)
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, nil, err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, err
	}

	defer resp.Body.Close()

	_, err = io.Copy(ioutil.Discard, resp.Body)
	if err != nil {
		return nil, nil, err
	}

	return resp, body, err
}

// initHTTPClient initialises the default HTTP client with fundamental
// connection options for every request.
func initHTTPClient() http.Client {
	return http.Client{
		Timeout: time.Duration(o.TimeoutMs) * time.Millisecond,
		// Do not follow redirects (HTTP status codes 30x).
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if !o.FollowRedirects {
				return http.ErrUseLastResponse
			}
			return nil
		},
	}
}
