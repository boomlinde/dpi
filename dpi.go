package dpi

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"os/user"
	"path/filepath"
	"regexp"
	"strings"
)

var exp = regexp.MustCompile("([a-z]+)='((?:''|[^'])+)'")

var keypath string

// noErrWriter is a writer that will never error.
// The first erroring write will instead store the error and further writes
// will be ignored.
type noErrWriter struct {
	w   io.Writer
	err error
}

func (ew *noErrWriter) Write(b []byte) (int, error) {
	if ew.err != nil {
		return 0, nil
	}
	n, err := ew.w.Write(b)
	if err != nil {
		ew.err = err
	}
	return n, nil
}

// ErrorOf returns the error, if any, of the writer passed to a Handler
// If the writer is not the one passed to the Handler, nil is returned
func ErrorOf(w io.Writer) error {
	if errw, ok := w.(*noErrWriter); ok {
		return errw.err
	}

	return nil
}

func init() {
	usr, err := user.Current()
	if err != nil {
		panic(err)
	}

	keypath = filepath.Join(usr.HomeDir, ".dillo", "dpid_comm_keys")
}

// Done is returned by a Handler when it considers itself done
var Done = errors.New("done")

// Handler is the user provided callback for handling tags
// The writer passed to the handler by the functions in this package will
// never return an error
type Handler func(map[string]string, io.Writer) error

// AutoRun will automatically determine from os.Args[0] whether to run as a
// filter (*.filter.dpi) or not
func AutoRun(cb Handler) error {
	if strings.HasSuffix(os.Args[0], ".filter.dpi") {
		return RunFilter(cb)
	}

	return Run(cb)
}

// Run starts an endless loop, accepting DPI connections and calling the
// handler with each new tag
func Run(cb Handler) error {
	listener, err := net.FileListener(os.Stdin)
	if err != nil {
		return err
	}
	defer listener.Close()
	for {
		conn, err := listener.Accept()
		if err != nil {
			return err
		}

		go func() {
			defer conn.Close()
			_ = handle(conn, conn, cb)
		}()
	}
}

func RunFilter(cb Handler) error {
	return handle(os.Stdin, os.Stdout, cb)
}

func handle(r io.Reader, w io.Writer, cb Handler) error {
	br := bufio.NewReader(r)
	for {
		var err error
		tag, err := parseCmd(br)
		if err != nil {
			return fmt.Errorf("failed to parse tag: %w", err)
		}
		if cmd, ok := tag["cmd"]; ok {
			if cmd == "auth" {
				if err := handleAuth(tag["msg"]); err != nil {
					if err == Done {
						return nil
					}
					return fmt.Errorf("auth failed: %w", err)
				}
				continue
			}
		}

		// Pass the message off to the handler
		ew := &noErrWriter{w: w}
		if err := cb(tag, ew); err != nil {
			return err
		}
		if ew.err != nil {
			return err
		}

	}
}

func parseCmd(br *bufio.Reader) (map[string]string, error) {
	prefix, err := br.Peek(1)
	if err != nil {
		return nil, err
	}
	if prefix[0] != byte('<') {
		return nil, errors.New("expected start of tag")
	}
	tag, err := br.ReadBytes(byte('>'))
	if err != nil {
		return nil, err
	}

	attrs := exp.FindAllStringSubmatch(string(tag), -1)
	parsed := map[string]string{}
	for _, attr := range attrs {
		parsed[attr[1]] = strings.Replace(attr[2], "''", "'", -1)
	}

	return parsed, nil
}

func handleAuth(msg string) error {
	var (
		pid int
		key string
	)

	f, err := os.Open(keypath)
	if err != nil {
		return err
	}

	if _, err := fmt.Fscanf(f, "%d %s", &pid, &key); err != nil {
		return err
	}
	if msg != key {
		return fmt.Errorf("wrong DPI key: '%s'", msg)
	}

	return nil
}

func Tag(w io.Writer, fields map[string]string) error {
	msg := []byte("<")
	for k, v := range fields {
		msg = append(msg, []byte(k)...)
		msg = append(msg, []byte("='")...)
		msg = append(msg, []byte(strings.Replace(v, "'", "''", -1))...)
		msg = append(msg, []byte("' ")...)
	}
	msg = append(msg, []byte("'>")...)
	_, err := w.Write(msg)
	return err
}
