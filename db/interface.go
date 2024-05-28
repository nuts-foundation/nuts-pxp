package db

type DB interface {
	Close() error
	Create(data SQLData) error
	Delete(id string) error
	Get(id string) (SQLData, error)
	Query(scope string, verifier string, client string) ([]string, error)
}
