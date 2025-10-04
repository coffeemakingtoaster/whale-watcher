package violations

import (
	"bytes"
	_ "embed"
	"net/url"
	"text/template"

	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

type Violations struct {
	CheckedCount   int
	ViolationCount int
	FixableCount   int
	Violations     []Violation
}

type Violation struct {
	RuleId      string
	Description string
	Fix         string
	AutoFixed   bool
}

type templateViolation struct {
	RuleId      string
	Description string
	URL         string
}

type templateContent struct {
	Fixed    []templateViolation
	Detected []templateViolation
	DocUrl   string
}

//go:embed pr_content.tmpl
var prTemplate string

func (v *Violations) BuildDescriptionMarkdown() string {
	var fixed []templateViolation
	var detected []templateViolation

	for _, violation := range v.Violations {
		if violation.AutoFixed {
			fixed = append(fixed, violationToTemplate(violation, viper.GetString("docsurl")))
		} else {
			detected = append(detected, violationToTemplate(violation, viper.GetString("docsurl")))
		}
	}
	tmpl, err := template.New("site").Parse(prTemplate)
	if err != nil {
		panic(err)
	}
	var writer bytes.Buffer
	err = tmpl.ExecuteTemplate(&writer, "site", templateContent{
		Fixed:    fixed,
		Detected: detected,
		DocUrl:   viper.GetString("docsurl"),
	})
	if err != nil {
		panic(err)
	}

	return writer.String()
}

func violationToTemplate(violation Violation, docBaseURL string) templateViolation {
	res := templateViolation{
		RuleId:      violation.RuleId,
		Description: violation.Description,
	}

	if len(docBaseURL) > 0 {
		u, err := url.Parse(docBaseURL)
		if err != nil {
			log.Warn().Err(err).Msg("Could not parse the docs url")
			return res
		}
		u.Fragment = url.PathEscape(res.RuleId)
		res.URL = u.String()
	}

	return res
}
