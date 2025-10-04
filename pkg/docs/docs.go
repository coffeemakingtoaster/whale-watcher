package docs

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

func serveRules(ruleSet rules.RuleSet, onlyExport bool, exportPath string, servePort int64) {
	if onlyExport {
		generateHTML(ruleSet, exportPath)
		return
	}

	serve := func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("Rendering docs")
		render(w, ruleSet)
	}

	http.HandleFunc("/", serve)
	fmt.Printf("See the docs: http://localhost:%d", servePort)
	err := http.ListenAndServe(fmt.Sprintf(":%d", servePort), nil)
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

func generateHTML(rulesSet rules.RuleSet, exportPath string) {
	f, err := os.Create(exportPath)
	if err != nil {
		panic(err)
	}
	render(f, rulesSet)
}
