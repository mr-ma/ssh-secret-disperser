package user

type Storage interface {
	LoadUsers() (map[string]User, error)
}

type DummyStore struct {
}

func (store DummyStore) LoadUsers() (map[string]User, error) {
	users := make(map[string]User)
	users["user1"] = User{
		Name:      "user1",
		OTPSecret: "L2NJFMNEJRBI2SNZQ2HUJNGRDCZEGTGM",
	}
	users["user2"] = User{
		Name:      "user2",
		OTPSecret: "LQDE6HPJHG55LXAHQ4LNEN2J2G6UIFHC",
	}
	return users, nil
}
