package parse

import (
	"sort"
)

type PackageSorter func([]*Package) sort.Interface
type PackageSlice []*Package

// SortByPackageName sorts packages in ascending alphabetical order.
func SortByPackageName(packages []*Package) sort.Interface { return PackageSlice(packages) }
func (packages PackageSlice) Len() int                     { return len(packages) }
func (packages PackageSlice) Swap(i, j int) {
	packages[i], packages[j] = packages[j], packages[i]
}
func (packages PackageSlice) Less(i, j int) bool {
	return packages[i].Summary.Package < packages[j].Summary.Package
}
