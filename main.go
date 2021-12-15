package main

import (
	"github.com/wayneashleyberry/envhunter/pkg/envhunter"
	"golang.org/x/tools/go/analysis/multichecker"
)

func main() {
	multichecker.Main(
		envhunter.Analyzer(),
	)

}
