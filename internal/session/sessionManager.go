package session

import (
	"github.com/alexedwards/scs/postgresstore"
	"time"
)

type postgresManager struct {
	 *postgresstore.PostgresStore
}

//GetTokenData get and return a specific token from the database
func (m *postgresManager) GetTokenData(token string) (b []byte, exists bool, err error) {
	return m.PostgresStore.Find(token)
}

//Savetoken is used to save the token in the db
func (m *postgresManager) Savetoken (token string, b []byte, expiry time.Time) error {
	return m.Commit(token, b, expiry)
}

//DeleteToken removes token from the db
func (m *postgresManager) DeleteToken(token string) error {
	return m.Delete(token)
}

//CleanUp uses to remove all token from the db and terminates all goroutines
//working in the db
func (m *postgresManager) CleanUp() {
	m.StopCleanup()
}