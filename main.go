package main

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"runtime"
	"sync"
	"sync/atomic"
	"time"
)

const (
	tokenLength = 15
	alphabet    = "0123456789abcdef"
	workers     = 8
	baseURL     = "https://127.0.0.1:8443/index.php?auth="
)

var (
	totalGenerated uint64
	totalMatched   uint64
)

/* =======================
   HTTPS MOCK SERVER
   ======================= */

func startMockServer() {
	mux := http.NewServeMux()

	mux.HandleFunc("/index.php", func(w http.ResponseWriter, r *http.Request) {
		auth := r.URL.Query().Get("auth")

		// Simulierter "gültiger" Token
		if auth == "5bb4ba51a4cb8cf" {
			body := make([]byte, 1234)
			w.Header().Set("Content-Length", "1234")
			w.WriteHeader(http.StatusOK)
			w.Write(body)
			return
		}

		// Standardantwort
		body := make([]byte, 5465)
		w.Header().Set("Content-Length", "5465")
		w.WriteHeader(http.StatusOK)
		w.Write(body)
	})

	server := &http.Server{
		Addr:    ":8443",
		Handler: mux,
	}

	go func() {
		err := server.ListenAndServeTLS("cert.pem", "key.pem")
		if err != nil {
			panic(err)
		}
	}()
}

/* =======================
   HTTPS CLIENT
   ======================= */

var client = &http.Client{
	Timeout: 2 * time.Second,
	Transport: &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true, // ⚠️ nur localhost
		},
		MaxIdleConns:        2000,
		MaxIdleConnsPerHost: 2000,
		IdleConnTimeout:    30 * time.Second,
	},
}

func checkToken(token string) bool {
	resp, err := client.Get(baseURL + token)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return false
	}

	// TRUE wenn Content-Length ≠ 5465
	return resp.Header.Get("Content-Length") != "5465"
}

/* =======================
   TOKEN GENERATOR
   ======================= */

func validNext(token []byte, pos int, c byte) bool {
	return !(pos >= 2 && token[pos-1] == c && token[pos-2] == c)
}

func generate(pos int, token []byte, jobs chan<- string) {
	if pos == tokenLength {
		jobs <- string(token)
		return
	}

	for i := 0; i < len(alphabet); i++ {
		c := alphabet[i]
		if validNext(token, pos, c) {
			token[pos] = c
			generate(pos+1, token, jobs)
		}
	}
}

/* =======================
   WORKER
   ======================= */

func worker(jobs <-chan string, wg *sync.WaitGroup) {
	defer wg.Done()

	for token := range jobs {
		atomic.AddUint64(&totalGenerated, 1)

		if checkToken(token) {
			atomic.AddUint64(&totalMatched, 1)
			fmt.Println("MATCH:", token)
		}
	}
}

/* =======================
   MAIN
   ======================= */

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	fmt.Println("Starting HTTPS mock server...")
	startMockServer()

	time.Sleep(500 * time.Millisecond) // Server hochfahren lassen

	jobs := make(chan string, 10_000)
	var wg sync.WaitGroup

	start := time.Now()

	for i := 0; i < workers; i++ {
		wg.Add(1)
		go worker(jobs, &wg)
	}

	go func() {
		generate(0, make([]byte, tokenLength), jobs)
		close(jobs)
	}()

	wg.Wait()
	elapsed := time.Since(start)

	gen := atomic.LoadUint64(&totalGenerated)
	match := atomic.LoadUint64(&totalMatched)

	fmt.Println("===== RESULT =====")
	fmt.Println("CPU cores:", runtime.NumCPU())
	fmt.Println("Workers:", workers)
	fmt.Println("Tokens generated:", gen)
	fmt.Println("Matches:", match)
	fmt.Println("Elapsed time:", elapsed)

	rate := float64(gen) / elapsed.Seconds()
	fmt.Printf("Rate: %.2f tokens/sec\n", rate)
}
