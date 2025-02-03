package storage

type StorFunc interface {
	SaveUser(Login string, Password string, UserID string) error
	GetUserID(Login string) (string, string, error)
}
