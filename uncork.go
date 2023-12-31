// uncork: Tunnel TCP connections through HTTP proxies, mostly a drop-in
// replacement for corkscrew.
//
// usage: uncork <proxyhost> <proxyport> <desthost> <destport>
package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

var (
	Version = "0.1.2"
	usage   = fmt.Sprintf(`uncork %s (martin.czygan@gmail.com)
usage: uncork <proxyhost> <proxyport> <desthost> <destport>`, Version)
	timeout = flag.Duration("T", 30*time.Second, "connect timeout")
)

// CloseWriter implements CloseWrite, as implemented by e.g. net.TCPConn.
type CloseWriter interface {
	CloseWrite() error
}

// CloseReader implements CloseRead, as implements by e.g. net.TCPConn.
type CloseReader interface {
	CloseRead() error
}

// stickyErrWriter can keep errors around, cf. https://youtu.be/yG-UaBJXZ80?t=33m50s
type stickyErrWriter struct {
	w   io.Writer
	err *error
}

// Write implements the io.Writer interface.
func (sew *stickyErrWriter) Write(p []byte) (n int, err error) {
	if *sew.err != nil {
		return 0, *sew.err
	}
	n, err = sew.w.Write(p)
	*sew.err = err
	return
}

// copyClose copies src to dst, then half closes src for reading and dst for
// writing. If any copy of close fails, this function will fail, albeit with a
// single flat error. To avoid resource leaks, try to put some (long) timeout
// on copy. TODO: allow cancellation.
func copyClose(dst io.WriteCloser, src io.ReadCloser, wg *sync.WaitGroup) (n int64, err error) {
	defer wg.Done()
	var serr, derr error
	n, err = io.Copy(dst, src)
	if v, ok := src.(CloseReader); ok {
		serr = v.CloseRead()
	} else {
		serr = src.Close()
	}
	if v, ok := dst.(CloseWriter); ok {
		derr = v.CloseWrite()
	} else {
		derr = dst.Close()
	}
	if err == nil && serr == nil && derr == nil {
		return n, nil
	}
	return n, fmt.Errorf("copyClose failed: copy: %v, s: %s, d: %v", err, serr, derr)
}

func main() {
	flag.Usage = func() {
		fmt.Println(usage)
	}
	flag.Parse()
	if flag.NArg() != 4 {
		fmt.Println(usage)
		os.Exit(1)
	}
	var (
		proxyHost     = flag.Arg(0)
		proxyPort     = flag.Arg(1)
		destHost      = flag.Arg(2)
		destPort      = flag.Arg(3)
		proxyHostPort = net.JoinHostPort(proxyHost, proxyPort)
		destHostPort  = net.JoinHostPort(destHost, destPort)
	)
	pconn, err := net.DialTimeout("tcp", proxyHostPort, *timeout)
	if err != nil {
		log.Fatal(err)
	}
	w := &stickyErrWriter{w: pconn, err: &err}
	fmt.Fprintf(w, "CONNECT %s HTTP/1.1\r\n", destHostPort)
	fmt.Fprintf(w, "Host: %s\r\n", destHostPort)
	fmt.Fprintf(w, "Proxy-Connection: keep-alive\r\n")
	fmt.Fprintf(w, "\r\n")
	if *w.err != nil {
		log.Fatal(*w.err)
	}
	var buf = make([]byte, 12) // == len("HTTP/1.1 200")
	_, err = pconn.Read(buf)
	if err != nil {
		log.Fatal(err)
	}
	fields := strings.Fields(string(buf))
	if len(fields) != 2 {
		log.Fatal("error: invalid response from proxy")
	}
	status, err := strconv.Atoi(fields[1])
	if err != nil {
		log.Fatal("error: could not parse proxy http status")
	}
	if status >= 300 {
		log.Fatalf("error: got http %v from proxy", status)
	}
	var wg sync.WaitGroup
	wg.Add(2)
	go copyClose(pconn, os.Stdin, &wg)
	go copyClose(os.Stdout, pconn, &wg)
	wg.Wait()
}
