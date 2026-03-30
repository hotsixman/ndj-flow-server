package server

type ServerException struct {
	code    string
	message string
}

func (this ServerException) Error() string {
	return "Server exception occured: " + this.code + "\n" + this.message
}
