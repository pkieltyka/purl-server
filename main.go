package main

import (
	"github.com/codegangsta/martini"
	"github.com/mreiferson/go-httpclient"
	"github.com/ugorji/go/codec"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"reflect"
	"strconv"
	"sync"
	"time"
	// "runtime"
)

const (
	VERSION   = "v0.9"
	MAXTIME   = 30             // in seconds
	CACHE_TTL = 60 * 60 * 24   // 1 day, in seconds
	THRUPUT   = 30             // max number of goroutines per fetch request
	ADDR      = "0.0.0.0:9333" // server address
)

var (
	reqId uint64 = 0 // Request counter
)

type Response struct {
	Url    string
	Status int
	Data   []byte
}

func main() {
	// runtime.GOMAXPROCS(runtime.NumCPU())
	// fmt.Printf("GOMAXPROCS is %d\n", runtime.GOMAXPROCS(0))

	// Server and middleware
	m := martini.New()
	m.Use(martini.Logger())
	m.Use(martini.Recovery())

	r := martini.NewRouter()
	m.Action(r.Handle)

	r.Get("/", func() string {
		return "."
	})

	r.Any("/fetch", func(req *http.Request, res http.ResponseWriter) {
		err := req.ParseForm()
		if err != nil {
			renderMsg(res, 422, "Unable to parse parameters")
			return
		}

		// Grab params and set defaults
		urls := req.Form["url"]
		if len(urls) == 0 {
			urls = req.Form["url[]"]
		}
		maxtime, _ := strconv.Atoi(req.Form.Get("maxtime"))
		if maxtime == 0 {
			maxtime = MAXTIME
		}
		ttl, _ := strconv.Atoi(req.Form.Get("ttl"))
		if ttl == 0 {
			ttl = CACHE_TTL
		}
		if len(urls) == 0 {
			renderMsg(res, 422, "Url parameter required")
			return
		}

		responses := httpFetch(urls, maxtime)

		if err != nil {
			renderMsg(res, 500, err.Error())
		} else {
			renderMsg(res, 200, responses)
		}
	})

	// Boot the server
	log.Println("** Purl", VERSION, "http server listening on", ADDR)
	if err := http.ListenAndServe(ADDR, m); err != nil {
		log.Fatal(err)
	}
}

func httpFetch(urls []string, maxtime int) []*Response {
	n := len(urls)
	if n == 0 {
		return nil
	}

	log.Println("Purl req", reqId)
	reqId++

	responses := make([]*Response, n)

	in := make(chan int)
	go func() {
		for i, url := range urls {
			responses[i] = &Response{Url: url}
			in <- i
		}
		close(in)
	}()

	timeout := time.Duration(time.Duration(maxtime) * time.Second)
	// timeout := time.Duration(1000 * time.Millisecond)
	transport := &httpclient.Transport{RequestTimeout: timeout} //, DisableKeepAlives: true}
	client := &http.Client{Transport: transport}
	defer transport.Close()

	// hrmm.. problem with the thruput is that we need to add more
	// logic around overall timeouts.. could be maxtime+maxtime
	var wg sync.WaitGroup
	thruput := int(math.Min(float64(THRUPUT), float64(n)))

	for i := 0; i < thruput; i++ {
		// log.Println("SCALE", i)
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := range in {
				url := urls[j]
				log.Println("Fetching", j, url)

				req, _ := http.NewRequest("GET", url, nil)
				resp, err := client.Do(req)
				if err != nil {
					log.Println("Http connect error for", url, "because:", err.Error())
					break
				}
				defer resp.Body.Close()

				body, err := ioutil.ReadAll(resp.Body)
				if err != nil {
					log.Println("Http GET error reading body for", url, "because:", err.Error())
					break
				}
				responses[j].Status = resp.StatusCode
				responses[j].Data = body
			}
		}()
	}
	wg.Wait()
	return responses
}

func renderMsg(res http.ResponseWriter, status int, data interface{}) {
	var out []byte
	var mh codec.MsgpackHandle
	enc := codec.NewEncoderBytes(&out, &mh)

	// Typecast ahead of encoding
	var err error
	switch data.(type) {
	case string:
		err = enc.Encode(data.(string))
	default:
		err = enc.Encode(data)
	}

	if err != nil {
		log.Println("Encoding error:", err.Error())
		renderMsg(res, 500, err.Error())
		return
	}

	res.Header().Set("Content-Type", "application/x-msgpack")
	res.WriteHeader(status)
	res.Write(out)
}

//
// * Debug helpers
//

func dTypeOf(x interface{}) {
	log.Println(reflect.TypeOf(x))
}
