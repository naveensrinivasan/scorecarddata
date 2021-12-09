package deps

type Deps interface {
	FetchDependecies(directory string) ([]string, error)
}
