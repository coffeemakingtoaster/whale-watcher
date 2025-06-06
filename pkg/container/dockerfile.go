package container

import (
	"github.com/coffeemakingtoaster/dockerfile-parser/pkg/ast"
	"github.com/coffeemakingtoaster/dockerfile-parser/pkg/lexer"
	"github.com/coffeemakingtoaster/dockerfile-parser/pkg/parser"
)

func GetDockerfileAST(path string) (*ast.StageNode, error) {
	l, err := lexer.NewFromFile(path)
	if err != nil {
		return nil, err
	}
	tokens, err := l.Lex()
	if err != nil {
		return nil, err
	}
	p := parser.NewParser(tokens)
	return p.Parse(), nil
}

func GetDockerfileInputAST(input []string) (*ast.StageNode, error) {
	l := lexer.NewFromInput(input)
	tokens, err := l.Lex()
	if err != nil {
		return nil, err
	}
	p := parser.NewParser(tokens)
	return p.Parse(), nil
}
