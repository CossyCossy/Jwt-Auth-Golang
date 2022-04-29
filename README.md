# How to create Authentication in Go with JWT and Rest APIs using postgres with Golang

## Authentication in Go with JWT

A simple project on how to build a JWT token based authentication with REST APIs using Golang and Postgresql.

## Installation

If using Linux/WSL:

```
sudo apt update
sudo apt install postgresql postgresql-contrib
```

If using macOS:

```
brew tap homebrew/services
brew install postgresql
initdb /user/local/var/postgres
```

## To start the server

```
go run main.go
```
Replace the .env with your own values.

## Test EndPoints

* SignUp. 

request
```
localhost:8080/signup
```
body:

```
{
    "username": "johndoe",
    "password": "abc123",
    "email" : "johndoe@example.com"
}
```

* Login. 

request
```
localhost:8080/login
```
body:

```
{
    "username": "johndoe",
    "password": "abc123"
}
```

## Contributing
Pull requests are always welcome! Feel free to open a new GitHub issue for any changes that can be made.
## Author
Cosmas Mbuvi | [https://crunchgarage.com](https://crunchgarage.com)
## License
[MIT](./LICENSE)