# WealthPulse

WealthPulse rewrite in go... I mean, why not, right?

... currently working on price parser... execute tests with `go test ./pkg/parse`


- [x] Update main so it actually parses prices and prints some info for the first few
- [ ] Parsing/lexing improvements
    - [ ] An observation about the lexer stateFn return type... it is inconvenient that each lexer needs to know the next state, as it seems to make the individual lexers hard to compose? Maybe I'm just missing something though
    - [ ] Add line numbers?
    - [ ] Good errors?
    - [ ] make the parser/lexer stream from file, maybe with an `io.Reader` ... tests can use `strings.NewReader`
        - Not sure io.Reader alone will cut it... maybe need a RuneReader, but might also want seeking abilities for the line? Not sure.
        - Maybe use buffered io stuff in bufio package? Is buffered io... better? I don't even know!
        - (there's probably a blog post here!)