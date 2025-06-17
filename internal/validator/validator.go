package val

import (
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"regexp"
)

const emailPattern = `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
const phonePattern = `^\+\d{5,}$`

func CheckUsername(username string) error {
	if username == "" {
		return status.Error(codes.InvalidArgument, "username is empty")
	}
	if (len(username) < 5) || (len(username) > 30) {
		return status.Error(codes.InvalidArgument, "invalid username length")
	}

	return nil
}

func CheckEmail(email string) error {
	if email == "" {
		return status.Error(codes.InvalidArgument, "email is empty")
	}
	if len(email) > 100 {
		return status.Error(codes.InvalidArgument, "invalid email length")
	}
	regx := regexp.MustCompile(emailPattern)

	res := regx.MatchString(email)
	if res == false {
		return status.Error(codes.InvalidArgument, "invalid email")
	}

	return nil
}

func CheckFIO(FIO string) error {
	if FIO == "" {
		return status.Error(codes.InvalidArgument, "FIO is empty")
	}
	if len(FIO) > 100 {
		return status.Error(codes.InvalidArgument, "invalid FIO length")
	}

	return nil
}

func CheckPhoneNumber(phoneNumber string) error {
	if phoneNumber == "" {
		return status.Error(codes.InvalidArgument, "phone number is empty")
	}
	if len(phoneNumber) > 15 || len(phoneNumber) < 7 {
		return status.Error(codes.InvalidArgument, "invalid phone number length")
	}
	rgx := regexp.MustCompile(phonePattern)

	res := rgx.MatchString(phoneNumber)
	if res == false {
		return status.Error(codes.InvalidArgument, "invalid phone number")
	}

	return nil
}

func CheckPassword(password string) error {
	if password == "" {
		return status.Error(codes.InvalidArgument, "password is empty")
	}
	if (len(password) < 6) || (len(password) > 24) {
		return status.Error(codes.InvalidArgument, "invalid password length")
	}

	return nil
}
