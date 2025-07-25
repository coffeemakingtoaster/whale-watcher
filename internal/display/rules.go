package display

import (
	_ "embed"
	"fmt"
	"log"
	"net/http"
	"text/template"

	"github.com/coffeemakingtoaster/whale-watcher/pkg/rules"
)

//go:embed site.tmpl
var siteTemplate string

func ServeRules(ruleSet rules.RuleSet) {

	serve := func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("Rendering docs")
		tmpl, err := template.New("site").Parse(siteTemplate)
		if err != nil {
			panic(err)
		}
		err = tmpl.ExecuteTemplate(w, "site", ruleSet)
		if err != nil {
			panic(err)
		}
	}

	http.HandleFunc("/", serve)
	fmt.Println("See the docs: http://localhost:3000")
	err := http.ListenAndServe(":3000", nil)
	if err != nil {
		log.Fatal(err)
	}
}
