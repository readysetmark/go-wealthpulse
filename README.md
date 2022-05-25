# WealthPulse

WealthPulse rewrite in go... I mean, why not, right?

... currently working on price parser... execute tests with `go test ./pkg/parse`


An observation about the lexer stateFn return type... it is inconvenient that each lexer needs to know the next state, as it seems to make the individual lexers hard to compose? Maybe I'm just missing something though