package cmd

import (
	"io"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/arsham/blush/blush"
)

// These variables are provided to support the tests.
var (
	FatalErr = func(s string) {
		log.Fatal(s)
	}
)

// Main reads the provided arguments from the command line and creates a
// blush.Blush instance. If there is any error, it will terminate the
// application with error code `1`, otherwise it calls the Write() method of
// Blush and exits with `0`.
func Main() {
	b, err := GetBlush(os.Args)
	if err != nil {
		FatalErr(err.Error())
		return
	}
	defer func() {
		if err := b.Close(); err != nil {
			FatalErr(err.Error())
		}
	}()
	if err = b.Write(os.Stdout); err != nil {
		FatalErr(err.Error())
		return
	}
}

// GetBlush returns an error if no arguments are provided or it can't find all
// the passed files. Files should be last arguments, otherwise they are counted
// as matching strings. If there is no file passed, the input should come in
// from Stdin as a pipe. We are not using the usual flag package because it
// cannot handle variables in the args.
func GetBlush(input []string) (b *blush.Blush, err error) {
	var ok bool
	if len(input) == 1 {
		return nil, ErrNoInput
	}
	remaining, r, err := getReader(input[1:])
	if err != nil {
		return nil, err
	}
	b = &blush.Blush{}
	b.Reader = r
	if remaining, ok = hasArg(remaining, "-C"); ok {
		b.NoCut = true
	}
	if remaining, ok = hasArg(remaining, "--colour"); ok {
		b.NoCut = true
	}
	b.Locator = getLocator(remaining)
	return
}

// getReader returns os.Stdin if it is piped to the program, otherwise looks for
// files.
func getReader(input []string) (remaining []string, r io.ReadCloser, err error) {
	var (
		recursive bool
		ok        bool
	)
	remaining, ok = hasArg(input, "-R")
	if ok {
		recursive = true
	}
	stat, _ := os.Stdin.Stat()
	if (stat.Mode() & os.ModeCharDevice) == 0 {
		return remaining, os.Stdin, nil
	}

	remaining, p, err := files(input)
	if err != nil {
		return nil, nil, err
	}
	w, err := blush.NewWalker(p, recursive)
	if err != nil {
		return nil, nil, err
	}
	return remaining, w, nil
}

// files starts from the end of the slice and removes any file it finds and
// returns them in p.
func files(input []string) (remaining []string, p []string, err error) {
	var (
		foundOne bool
		counter  int
		ret      []string
	)
	// going backwards from the end.
	sort.SliceStable(input, func(i, j int) bool {
		return i > j
	})
	for i, t := range input {
		t = strings.Trim(t, " ")
		if t == "" || inStringSlice(t, p) {
			continue
		}
		if m, _ := filepath.Glob(t); len(m) > 0 {
			foundOne = true
			p = append(p, t)
			counter++
			continue
		} else if foundOne {
			// there is already a pattern found so we stop here.
			ret = append(ret, input[i:]...)
			break
		}
		ret = append(ret, t)
	}
	if !foundOne {
		return input, nil, ErrNoFilesFound
	}

	// We have reversed it. We need to return back in the same order.
	sort.SliceStable(ret, func(i, j int) bool {
		return i > j
	})
	// to keep the original user's preference.
	sort.SliceStable(p, func(i, j int) bool {
		return i > j
	})
	return ret, p, nil
}

func inStringSlice(s string, haystack []string) bool {
	for _, a := range haystack {
		if a == s {
			return true
		}
	}
	return false
}

func getLocator(input []string) []blush.Locator {
	var (
		lastColour  string
		ret         []blush.Locator
		insensitive bool
		ok          bool
	)
	if input, ok = hasArg(input, "-i"); ok {
		insensitive = true
	}
	for _, token := range input {
		if strings.HasPrefix(token, "-") {
			lastColour = strings.TrimLeft(token, "-")
			continue
		}
		a := blush.NewLocator(lastColour, token, insensitive)
		ret = append(ret, a)
	}
	return ret
}

// hasArg removes the `arg` argument and returns the remaining []string.
func hasArg(input []string, arg string) ([]string, bool) {
	for i, a := range input {
		if a == arg {
			return append(input[:i], input[i+1:]...), true
		}
	}
	return input, false
}
