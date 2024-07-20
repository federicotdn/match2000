# match2000
match2000 is a library for performing fuzzy text search in Go. It is a modified version of a subset of the [go-diff](https://github.com/sergi/go-diff) library.

The main modification made was that the `pattern` argument can now have any length. Previously having too long of a pattern would cause an overflow due to the code attempting to left-shift an `int` more than 32 times. The match2000 implementation uses the `Int` type from `math/big`, gaining flexibility but with a performance penalty.

Other smaller modifications applied were:
- Removed assumptions about strings being UTF-8 encoded.
- Changed all usages of `math.Min` and `math.Max` to `min` and `max`.
- Removed all unnecessary `int` -> `float64` -> `int` conversions (after above change was applied).
- Removed dependency on `github.com/stretchr/testify/assert` package.
- Changed `DiffMatchPatch` struct name to just `Match`.
- Unexported `MatchAlphabet`.
- Simplified code slightly where possible.
- Adapted tests to changes described above, plus added more test cases.

## License

Same license as the original go-diff library.
