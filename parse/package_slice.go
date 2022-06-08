package parse

import (
	"sort"
)

type PackageSorter func([]*Package) sort.Interface
type PackageSlice []*Package

type byCoverage struct{ PackageSlice }
type byElapsed struct{ PackageSlice }

// SortByPackageName sorts packages in ascending alphabetical order.
func SortByPackageName(packages []*Package) sort.Interface { return PackageSlice(packages) }
func (packages PackageSlice) Len() int                     { return len(packages) }
func (packages PackageSlice) Swap(i, j int) {
	packages[i], packages[j] = packages[j], packages[i]
}
func (packages PackageSlice) Less(i, j int) bool {
	return packages[i].Summary.Package < packages[j].Summary.Package
}

// SortByCoverage sorts packages in descending order of code coverage.
func SortByCoverage(packages []*Package) sort.Interface { return byCoverage{packages} }
func (packages byCoverage) Less(i, j int) bool {
	return packages.PackageSlice[i].Coverage > packages.PackageSlice[j].Coverage
}

// SortByCoverage sorts packages in descending order of elapsed time per package.
func SortByElapsed(packages []*Package) sort.Interface { return byElapsed{packages} }
func (packages byElapsed) Less(i, j int) bool {
	return packages.PackageSlice[i].Summary.Elapsed > packages.PackageSlice[j].Summary.Elapsed
}
