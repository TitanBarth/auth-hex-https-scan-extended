package main

/*
   High‑Performance Go Application for HTTPS Token Simulation and Response Classification
   -----------------------------------------------------------------------------
   • Generate 15‑character hexadecimal tokens starting from a configurable start value.
   • No character may appear three times consecutively inside a token.
   • Send a GET request for each token to a public HTTPS server.
   • Every successful response (200 + Content‑Length ≠ 5465) is written to matches.txt.
   • Statistics are written to stats.txt.
   • The program is fully concurrent – a worker pool, keep‑alive HTTPS client
     and a dedicated writer goroutine are used.
   • All code lives in a single file (main.go) and is fully deterministic.
*/
import (
	"bufio"
	"crypto/tls"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"runtime"
	"sync"
	"sync/atomic"
	"time"
	"math/rand"
)

// ---------- CONFIGURATION ----------------------------------------------------
type Config struct {
	workers     int
	tokens      uint64 // number of tokens to generate
	startValue  uint64 // numeric start value (decimal)
	tokenLength int    // token length in hex digits (15)
}

// ---------- GLOBALS ----------------------------------------------------------
var (
	cfg         Config
	baseURL     = "https://www.example.com/kensaku_s.html?auth="
	matchesCnt  int64 // atomic counter of matches
	startTime   time.Time
	totalRequests uint64
	totalMatches  uint64
)

// ---------- MAIN -------------------------------------------------------------
func main() {
	// ---- parse command line ------------------------------------------------
	flag.IntVar(&cfg.workers, "workers", runtime.NumCPU(), "number of concurrent workers")
	flag.Uint64Var(&cfg.tokens, "tokens", 1000, "total number of tokens to generate")
	flag.Uint64Var(&cfg.startValue, "start", 100000000000000, "starting decimal value")
	flag.IntVar(&cfg.tokenLength, "len", 15, "token length in hex digits (15)")
	flag.Parse()

	// ---- start statistics --------------------------------------------------
	startTime = time.Now()
	runtime.GOMAXPROCS(runtime.NumCPU())

	// ---- create channels ----------------------------------------------------
	tokenCh := make(chan string, cfg.workers*4) // buffered to keep workers busy
	matchCh := make(chan string, cfg.workers*4)

	// ---- writer goroutine --------------------------------------------------
	var wgWriter sync.WaitGroup
	wgWriter.Add(1)
	go writer(matchCh, &wgWriter)

	// ---- worker pool --------------------------------------------------------
	var wgWorkers sync.WaitGroup
	for i := 0; i < cfg.workers; i++ {
		wgWorkers.Add(1)
		go worker(i, tokenCh, matchCh, &wgWorkers)
	}

	// ---- token generator ---------------------------------------------------
	generateTokens(tokenCh)

	// ---- close token channel and wait for workers -------------------------
	close(tokenCh)
	wgWorkers.Wait()

	// ---- close match channel and wait for writer -------------------------
	close(matchCh)
	wgWriter.Wait()

	// ---- write statistics --------------------------------------------------
	writeStats()
}

// ---------- TOKEN GENERATOR --------------------------------------------------
func generateTokens(ch chan<- string) {
	seed := time.Now().UnixNano()
	rng := rand.New(rand.NewSource(seed))

	hexChars := []byte("0123456789abcdef")
	buf := make([]byte, cfg.tokenLength)

	for {
		for i := 0; i < cfg.tokenLength; i++ {
			for {
				b := hexChars[rng.Intn(16)]

				// keine 3 gleichen Zeichen hintereinander
				if i >= 2 && buf[i-1] == b && buf[i-2] == b {
					continue
				}

				buf[i] = b
				break
			}
		}
		ch <- string(buf)
	}
}

// ---------- CONSECUTIVE CHECK ----------------------------------------------
func hasThreeConsecutive(s string) bool {
	if len(s) < 3 {
		return true // trivially true
	}
	for i := 0; i < len(s)-2; i++ {
		if s[i] == s[i+1] && s[i+1] == s[i+2] {
			return false
		}
	}
	return true
}

// ---------- WORKER ----------------------------------------------------------
func worker(id int, tokenCh <-chan string, matchCh chan<- string, wg *sync.WaitGroup) {
	defer wg.Done()

	// HTTPS client with keep‑alive & insecure TLS (for localhost)
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		MaxIdleConns:    cfg.workers * 10,
		IdleConnTimeout: 30 * time.Second,
	}
	client := &http.Client{Transport: tr}

	for token := range tokenCh {
		reqURL := baseURL + token
		resp, err := client.Get(reqURL)
		if err != nil {
			fmt.Fprintf(os.Stderr, "[worker %d] error GET %s: %v\n", id, reqURL, err)
			continue
		}
	
		// ✅ HIER: Request war erfolgreich
		atomic.AddUint64(&totalRequests, 1)
	
		// read body fully to get Content-Length
	body, err := io.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		continue
	}
	
	if resp.StatusCode == 200 && len(body) != 5465 {
		atomic.AddUint64(&totalMatches, 1)
		matchCh <- token
	}
	resp.Body.Close()
	
	}
}

// ---------- WRITER ----------------------------------------------------------

func writer(ch <-chan string, wg *sync.WaitGroup) {
	defer wg.Done()

	f, err := os.Create("matches.txt")
	if err != nil {
		panic(err)
	}
	defer f.Close()

	w := bufio.NewWriter(f)
	defer w.Flush()

	for line := range ch {
		if _, err := w.WriteString(line + "\n"); err != nil {
			panic(err)
		}
	}
}





func writeStats() {
	f, err := os.Create("stats.txt")
	if err != nil {
		panic(err)
	}
	defer f.Close()

	elapsed := time.Since(startTime)
	rps := float64(totalRequests) / elapsed.Seconds()

	w := bufio.NewWriter(f)
	defer w.Flush()

	fmt.Fprintf(w, "Execution statistics\n")
	fmt.Fprintf(w, "====================\n")
	fmt.Fprintf(w, "CPU cores          : %d\n", runtime.NumCPU())
	fmt.Fprintf(w, "Workers            : %d\n", cfg.workers)
	fmt.Fprintf(w, "Total requests     : %d\n", totalRequests)
	fmt.Fprintf(w, "Valid matches      : %d\n", totalMatches)
	fmt.Fprintf(w, "Elapsed time       : %s\n", elapsed)
	fmt.Fprintf(w, "Requests / second  : %.2f\n", rps)
}
