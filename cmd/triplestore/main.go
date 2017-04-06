package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	tstore "github.com/wallix/triplestore"
)

var filesFlag arrayFlags
var outFormat string

func init() {
	flag.StringVar(&outFormat, "out-format", "ntriples", "output format (ntriples, bin)")
	flag.Var(&filesFlag, "in", "input file paths")
}

func main() {
	flag.Parse()
	if len(filesFlag) == 0 {
		log.Fatal("need at list an argument `-in INPUT_FILE`")
	}
	err := convert(filesFlag, outFormat)
	if err != nil {
		log.Fatal(err)
	}
}

func convert(inFilePaths []string, outFormat string) error {
	var inFiles []io.Reader
	for _, inFilePath := range inFilePaths {
		in, err := os.Open(inFilePath)
		if err != nil {
			return fmt.Errorf("open input file '%s': %s", inFilePath, err)
		}
		inFiles = append(inFiles, in)
	}

	triples, err := tstore.NewDatasetDecoder(tstore.NewBinaryDecoder, inFiles...).Decode()
	if err != nil {
		return err
	}

	var encoder tstore.Encoder
	switch outFormat {
	case "ntriples":
		encoder = tstore.NewNTriplesEncoder(os.Stdout)
	case "bin":
		encoder = tstore.NewBinaryEncoder(os.Stdout)
	default:
		return fmt.Errorf("unknown format %s, expected either 'ntriples' or 'bin'", outFormat)
	}
	err = encoder.Encode(triples...)
	if err != nil {
		return err
	}

	return nil
}

type arrayFlags []string

func (i *arrayFlags) String() string {
	return strings.Join(*i, ",")
}

func (i *arrayFlags) Set(value string) error {
	*i = append(*i, value)
	return nil
}
