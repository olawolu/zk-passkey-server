package data

type Datastore struct {
	// database
	// session store
}

func (ds *Datastore) GetUser() User { return User{} }

func (ds *Datastore) SaveUser(user User) {}
