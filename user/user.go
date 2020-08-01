package user

type User struct {
	Name      string
	OTPSecret string
}

func NewUser(username string, otpsecret string) *User {
	user := new(User)
	user.OTPSecret = otpsecret
	user.Name = username
	return user
}
