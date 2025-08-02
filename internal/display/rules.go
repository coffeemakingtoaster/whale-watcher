package display

import (
	_ "embed"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"text/template"

	"github.com/coffeemakingtoaster/whale-watcher/pkg/rules"
)

//go:embed site.tmpl
var siteTemplate string

func ServeRules(ruleSet rules.RuleSet, export bool) {
	if export {
		generateHTML(ruleSet)
		return
	}

	serve := func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("Rendering docs")
		render(w, ruleSet)
	}

	http.HandleFunc("/", serve)
	fmt.Println("See the docs: http://localhost:3000")
	err := http.ListenAndServe(":3000", nil)
	if err != nil {
		log.Fatal(err)
	}
}

func render(w io.Writer, ruleSet rules.RuleSet) {
	tmpl, err := template.New("site").Parse(siteTemplate)
	if err != nil {
		panic(err)
	}
	err = tmpl.ExecuteTemplate(w, "site", ruleSet)
	if err != nil {
		panic(err)
	}
}

func generateHTML(rulesSet rules.RuleSet) {
	fmt.Println("Generating docs to ./index.html")
	f, err := os.Create("./index.html")
	if err != nil {
		panic(err)
	}
	render(f, rulesSet)
}
