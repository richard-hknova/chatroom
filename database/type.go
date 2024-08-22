package database

type User struct {
	Username string
	Avatar   int
	Hash     string
}

type Message struct {
	Received bool
	Sender   string
	Receiver string
	Content  string
}

type Friend struct {
	UserOne  string
	UserTwo  string
	Accepted bool
}
