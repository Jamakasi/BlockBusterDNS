package dmatcher

type IStorage interface {
	AddDomain(string) error
	DelDomain(string) error
	ContainDomain(string) (bool, error)
	GetDomainList(string) ([]string, error)
	Load()
	Save()
}
