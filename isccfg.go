package isccfg

import (
	"encoding/json"
	"github.com/timtadh/lexmachine"
	lexmachines "github.com/timtadh/lexmachine/machines"
	"io"
	"io/ioutil"
)

type Value interface{}
type Config map[string]Value

func NewLexer() *lexmachine.Lexer {
	literals := []string{
		"{",
		"}",
		";",
		",",
	}
	tokens := []string{
		"VALUE",
	}
	tokens = append(tokens, literals...)
	tokenIds := map[string]int{}
	for i, tok := range tokens {
		tokenIds[tok] = i
	}
	lex := lexmachine.NewLexer()

	skip := func(*lexmachine.Scanner, *lexmachines.Match) (interface{}, error) {
		return nil, nil
	}
	token := func(name string) lexmachine.Action {
		return func(s *lexmachine.Scanner, m *lexmachines.Match) (interface{}, error) {
			return s.Token(tokenIds[name], string(m.Bytes), m), nil
		}
	}
	stripToken := func(name string) lexmachine.Action {
		return func(s *lexmachine.Scanner, m *lexmachines.Match) (interface{}, error) {
			m.Bytes = m.Bytes[1:len(m.Bytes)-1]
			return s.Token(tokenIds[name], string(m.Bytes), m), nil
		}
	}

	lex.Add([]byte(`#[^\n]*\n?`), skip)  // comments as in ISC-DHCP-Server
	lex.Add([]byte(`//[^\n]*\n?`), skip) // comments as in ISC-BIND9
	lex.Add([]byte(`([a-z]|[A-Z]|[0-9]|_|\-|\.)*`), token("VALUE"))
	lex.Add([]byte(`"([^\\"]|(\\.))*"`), stripToken("VALUE"))
	lex.Add([]byte("[\n \t]"), skip)
	for _, lit := range literals {
		lex.Add([]byte(lit), token(lit))
	}

	err := lex.Compile()
	if err != nil {
		panic(err)
	}

	return lex
}

func Parse(reader io.Reader) (cfg Config, err error) {
	var notParsedConfig []byte
	notParsedConfig, err = ioutil.ReadAll(reader)
	if err != nil {
		return
	}

	lex := NewLexer()

	var scanner *lexmachine.Scanner
	scanner, err = lex.Scanner(notParsedConfig)
	if err != nil {
		return
	}
	cfg = Config{}
	stack := []*Config{&cfg}
	ptr := stack[0]

	accumulated := []string{}

	addValue := func(value string) {
		if (*ptr)["_value"] == nil {
			(*ptr)["_value"] = []string{}
		}
		(*ptr)["_value"] = append((*ptr)["_value"].([]string), value)
	}
	goThrough := func(path []string) {
		for _, step := range path {
			if (*ptr)[step] == nil {
				(*ptr)[step] = &Config{}
			}
			ptr = (*ptr)[step].(*Config)
		}
	}

	for tok, err, eof := scanner.Next(); !eof; tok, err, eof = scanner.Next() {
		if err != nil {
			return Config{}, err
		}
		token := tok.(*lexmachine.Token)

		switch token.Type {
		case 0: // some word or quoted value
			accumulated = append(accumulated, string(token.Lexeme))
		case 1: // "{"
			goThrough(accumulated)
			accumulated = []string{}
			stack = append(stack, ptr)
		case 2: // "}"
			ptr, stack = stack[0], stack[1:]
		case 3: // ";"
			if len(accumulated) > 0 {
				goThrough(accumulated[:len(accumulated)-1])
				addValue(accumulated[len(accumulated)-1])
				accumulated = []string{}
			}
			ptr = stack[len(stack)-1]
		case 4: // ","
			goThrough(accumulated[:len(accumulated)-1])
			addValue(accumulated[len(accumulated)-1])
			for tok, err, eof := scanner.Next(); !eof; tok, err, eof = scanner.Next() {
				if err != nil {
					return Config{}, err
				}
				token := tok.(*lexmachine.Token)

				if token.Type == 3 {
					break
				}
				if token.Type == 4 {
					continue
				}

				addValue(string(token.Lexeme))
			}
			accumulated = []string{}
			ptr = stack[len(stack)-1]
		}
	}

	return
}

func (cfg Config) WriteJsonTo(writer io.Writer) (err error) {
	jsonEncoder := json.NewEncoder(writer)
	jsonEncoder.SetIndent("", "  ")
	return jsonEncoder.Encode(cfg)
}
