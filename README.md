# ParkLeaks API

The ParkLeaks API is used to interact between any application (including the official app) to interact with ParkLeaks.nl, the following repository contains the code powering the API.

## Dependencies
- [gorilla/mux](https://github.com/gorilla/mux)
- [Go-MySQL-Driver](https://github.com/go-sql-driver/mysql)
- [GoDotEnv](https://github.com/joho/godotenv)

## Deployment

Rename `example.env` to `.env` and enter your MySQL DSN and reCAPTCHA key, if you'd like to test only use [this key](https://developers.google.com/recaptcha/docs/faq#id-like-to-run-automated-tests-with-recaptcha.-what-should-i-do).
Now run the follow command
```bash
go build -i
```
This will will give you a single binary  which you can now run

## License

This project is licensed under the Apache License 2.0. see the [LICENSE](LICENSE) file for more information
