package isccfg

import (
	"encoding/json"
	"fmt"
	"github.com/timtadh/lexmachine"
	lexmachines "github.com/timtadh/lexmachine/machines"
	"io"
	"io/ioutil"
	"strings"
)

type Value interface{}
type Config map[string]Value

func (cfg Config) Values() []string {
	return cfg["_value"].([]string)
}
func (cfg Config) UnwrapParam(param string) *Config {
	return cfg[param].(*Config)
}
func (cfg Config) Unwrap() (result *Config, value string) {
	for k := range cfg {
		value = k
		result, _ = cfg[k].(*Config)
		break
	}
	return
}
func (cfg Config) Unroll() (result []string) {
	ptr := &cfg
	for {
		ptr, value := ptr.Unwrap()
		result = append(result, value)
		if ptr == nil {
			return
		}
	}
	return
}

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
	lex.Add([]byte(`([a-z]|[A-Z]|[0-9]|_|\-|\.|=)*`), token("VALUE"))
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
	lex := NewLexer()
	return ParseWithLexer(lex, reader)
}

func ParseWithLexer(lex *lexmachine.Lexer, reader io.Reader) (cfg Config, err error) {
	var notParsedConfig []byte
	notParsedConfig, err = ioutil.ReadAll(reader)
	if err != nil {
		return
	}

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
		//fmt.Println("addValue", value)
		if (*ptr)["_value"] == nil {
			(*ptr)["_value"] = []string{}
		}
		(*ptr)["_value"] = append((*ptr)["_value"].([]string), value)
	}
	goThrough := func(path []string) {
		//fmt.Println("goThrough", path)
		for _, step := range path {
			if (*ptr)[step] == nil {
				(*ptr)[step] = &Config{}
			}
			ptr = (*ptr)[step].(*Config)
		}
	}

	wordsPerValue := 1
	for tok, err, eof := scanner.Next(); !eof; tok, err, eof = scanner.Next() {
		if err != nil {
			return Config{}, err
		}
		token := tok.(*lexmachine.Token)
		//fmt.Println("token", token)

		switch token.Type {
		case 0: // some word or quoted value
			accumulated = append(accumulated, string(token.Lexeme))
		case 1: // "{"
			goThrough(accumulated)
			accumulated = []string{}
			stack = append(stack, ptr)
			wordsPerValue = 1
		case 2: // "}"
			ptr, stack = stack[len(stack)-2], stack[:len(stack)-1]
			wordsPerValue = 1
		case 3: // ";"
			if len(accumulated) < wordsPerValue {
				panic("too multiworded value in a comma separated list")
			}
			if len(accumulated) > 0 {
				goThrough(accumulated[:len(accumulated)-wordsPerValue])
				addValue(accumulated[len(accumulated)-wordsPerValue])
				accumulated = []string{}
			}
			ptr = stack[len(stack)-1]
			wordsPerValue = 1
		case 4: // ","
			localAccumulated := []string{}
			localValues := []string{}
			for tok, err, eof := scanner.Next(); !eof; tok, err, eof = scanner.Next() { // walking until ";"
				if err != nil {
					return Config{}, err
				}
				token := tok.(*lexmachine.Token)

				if token.Type == 3 { // ";"
					localValues = append(localValues, strings.Join(localAccumulated, " "))
					wordsPerValue = len(localAccumulated)
					localAccumulated = []string{}
					break
				}
				if token.Type == 4 { // ","
					localValues = append(localValues, strings.Join(localAccumulated, " "))
					localAccumulated = []string{}
					continue
				}
				if token.Type == 0 {
					localAccumulated = append(localAccumulated, string(token.Lexeme))
					continue
				}
				panic(fmt.Errorf("syntax error, got \"%v\"", string(token.Lexeme)))
			}
			if len(accumulated) < wordsPerValue {
				panic("too multiworded value in a comma separated list")
			}
			goThrough(accumulated[:len(accumulated)-wordsPerValue])
			addValue(strings.Join(accumulated[len(accumulated)-wordsPerValue:], " ")) // adding the first value of the list
			for _, localValue := range localValues { // adding the rest values of the list
				addValue(localValue)
			}
			accumulated = []string{}
			ptr = stack[len(stack)-1]
			wordsPerValue = 1
		}
	}

	return
}

func (cfg Config) WriteJsonTo(writer io.Writer) (err error) {
	jsonEncoder := json.NewEncoder(writer)
	jsonEncoder.SetIndent("", "  ")
	return jsonEncoder.Encode(cfg)
}
