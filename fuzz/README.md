## Fuzzing decoders

For context have a look at [go-fuzz](https://github.com/dvyukov/go-fuzz)

Corpus sample data are in `fuzz/corpus/samples.*`

Build with

```sh
go-fuzz-build github.com/wallix/triplestore/fuzz
```

Then

```sh
go-fuzz -bin=tstore-fuzz.zip -workdir=fuzz/corpus
```

Stop when enough and look at the results. Fix the bugs and clean up the generated unneeded data (`rm -r fuzz/corpus/{corpus,crashers,suppressions}`)
