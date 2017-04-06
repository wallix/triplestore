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
var outFormatFlag string

var useRdfPrefixesFlag bool
var prefixesFlag arrayFlags
var baseFlag string

func init() {
	flag.StringVar(&outFormatFlag, "out-format", "ntriples", "output format (ntriples, bin)")
	flag.Var(&filesFlag, "in", "input file paths")
	flag.BoolVar(&useRdfPrefixesFlag, "rdf-prefixes", false, "use default RDF prefixes (rdf, rdfs, xsd)")
	flag.Var(&prefixesFlag, "prefix", "RDF custom prefixes (format: \"prefix:http://my.uri\"")
	flag.StringVar(&baseFlag, "base", "", "RDF custom base prefix")
}

func main() {
	flag.Parse()
	if len(filesFlag) == 0 {
		log.Fatal("need at list an argument `-in INPUT_FILE`")
	}
	context, err := buildContext(useRdfPrefixesFlag, prefixesFlag, baseFlag)
	if err != nil {
		log.Fatal(err)
	}
	err = convert(filesFlag, outFormatFlag, context)
	if err != nil {
		log.Fatal(err)
	}
}

func buildContext(useRdfPrefixes bool, prefixes []string, base string) (*tstore.Context, error) {
	var context *tstore.Context
	if useRdfPrefixes {
		context = tstore.RDFContext
	} else {
		context = tstore.NewContext()
	}
	for _, prefix := range prefixes {
		splits := strings.SplitN(prefix, ":", 2)
		if splits[0] == "" || splits[1] == "" {
			return context, fmt.Errorf("invalid prefix format: '%s'. expected \"prefix:http://my.uri\"", prefix)
		}
		context.Prefixes[splits[0]] = splits[1]
	}
	context.Base = base
	return context, nil
}

func convert(inFilePaths []string, outFormatFlag string, context *tstore.Context) error {
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
	switch outFormatFlag {
	case "ntriples":
		encoder = tstore.NewNTriplesEncoderWithContext(os.Stdout, context)
	case "bin":
		encoder = tstore.NewBinaryEncoder(os.Stdout)
	default:
		return fmt.Errorf("unknown format %s, expected either 'ntriples' or 'bin'", outFormatFlag)
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
