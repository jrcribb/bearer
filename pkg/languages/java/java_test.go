package java_test

import (
	_ "embed"
	"testing"

	"github.com/bradleyjkemp/cupaloy"

	"github.com/bearer/bearer/pkg/languages/java"
	"github.com/bearer/bearer/pkg/languages/testhelper"
	patternquerybuilder "github.com/bearer/bearer/pkg/scanner/detectors/customrule/patternquery/builder"
)

//go:embed testdata/import.yml
var importRule []byte

//go:embed testdata/logger.yml
var loggerRule []byte

//go:embed testdata/scope_rule.yml
var scopeRule []byte

func TestImport(t *testing.T) {
	testhelper.GetRunner(t, importRule, java.Get()).RunTest(t, "./testdata/import", ".snapshots/")
}

func TestFlow(t *testing.T) {
	testhelper.GetRunner(t, loggerRule, java.Get()).RunTest(t, "./testdata/testcases/flow", ".snapshots/flow/")
}

func TestScope(t *testing.T) {
	testhelper.GetRunner(t, scopeRule, java.Get()).RunTest(t, "./testdata/scope", ".snapshots/")
}

func TestPattern(t *testing.T) {
	for _, test := range []struct{ name, pattern string }{
		{"method params is a container type", `
				class $<_> {
					void main($<!>$<_>) {}
				}
		`},
		{"catch types is a container type", `
				class $<_> {
					void main() {
						try {} catch ($<!>$<_> e) {}
					}
				}
		`},
	} {
		t.Run(test.name, func(tt *testing.T) {
			result, err := patternquerybuilder.Build(java.Get(), test.pattern, "")
			if err != nil {
				tt.Fatalf("failed to build pattern: %s", err)
			}

			cupaloy.SnapshotT(tt, result)
		})
	}
}
