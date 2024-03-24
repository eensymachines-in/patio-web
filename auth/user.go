package auth

type UserRole uint8

const (
	SuperUser UserRole = iota
	Admin
	EndUser
	Guest
)

// User : any user in the system, can be authenticated against database
type User struct {
	Name    string   `bson:"name"`
	Email   string   `bson:"email"`
	Role    UserRole `bson:"role"`
	TelegID int64    `bson:"telegid"`
	Auth    string   `bson:"auth"`
}
