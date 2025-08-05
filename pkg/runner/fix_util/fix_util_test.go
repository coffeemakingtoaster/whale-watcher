package fixutil_test

import (
	"reflect"
	"strings"
	"testing"

	fixutil "github.com/coffeemakingtoaster/whale-watcher/pkg/runner/fix_util"
)

var sampleDockerfile = []string{
	"ARG PREFROMSTATEMENT=true",
	"FROM golang:1.24-bookworm AS build",
	"",
	"WORKDIR /build",
	"# Install deps",
	"RUN apt update && apt install -y python3-pip && \\",
	"python3 -m pip install pybindgen --break-system-packages && \\",
	"go install golang.org/x/tools/cmd/goimports@latest && \\",
	"go install github.com/go-python/gopy@latest",
	"",
	"COPY . .",
	"",
	"# Clean is not necessary here...but better safe than sorry",
	"RUN make clean all verify",
	"",
	"FROM python:3.10-bookworm AS runtime",
	"",
	"WORKDIR /app",
	"",
	"COPY --from=build /build/build/whale-watcher ./whale-watcher",
	"",
	"ENTRYPOINT [\"/app/whale-watcher\"]",
}

func getDiff(expected, actual []string) (string, string) {
	bound := min(len(expected), len(actual))
	for i := range bound {
		if actual[i] != expected[i] {
			return expected[i], actual[i]
		}
	}
	return strings.Join(expected[bound:], "\n"), strings.Join(actual[bound:], "\n")
}

func TestAddParamCommandNotPresent(t *testing.T) {
	fu := fixutil.SetupFromContent(sampleDockerfile)
	expected := fu.GetReconstruct()
	fu.EnsureCommandAlwaysHasParam([]string{"curl"}, "-f")
	actual := fu.GetReconstruct()

	if !reflect.DeepEqual(expected, actual) {
		expectedDiff, actualDiff := getDiff(expected, actual)
		t.Errorf("Run node mismatch: Expected %s Got: %s", expectedDiff, actualDiff)
	}
}

func TestAddParamCommandPresent(t *testing.T) {
	alteredDockerfile := append(sampleDockerfile, "RUN apt-get install vim-btw")
	expected := "RUN [\"apt-get\",\"install\",\"-y\",\"vim-btw\"]"
	fu := fixutil.SetupFromContent(alteredDockerfile)
	fu.EnsureCommandAlwaysHasParam([]string{"apt-get", "install"}, "-y")

	out := fu.GetReconstruct()
	actual := out[len(out)-1]

	if actual != expected {
		t.Errorf("Run node mismatch: Expected %s Got: %s", expected, actual)
	}
}

func TestAddParamCommandPresentMultiple(t *testing.T) {
	alteredDockerfile := append(sampleDockerfile, "RUN apt-get update && apt-get install vim-btw && apt-get install arch-btw")
	expected := "RUN [\"apt-get\",\"update\",\"&&\",\"apt-get\",\"install\",\"-y\",\"vim-btw\",\"&&\",\"apt-get\",\"install\",\"-y\",\"arch-btw\"]"
	fu := fixutil.SetupFromContent(alteredDockerfile)
	fu.EnsureCommandAlwaysHasParam([]string{"apt-get", "install"}, "-y")

	out := fu.GetReconstruct()
	actual := out[len(out)-1]

	if actual != expected {
		t.Errorf("Run node mismatch: Expected %s Got: %s", expected, actual)
	}
}
