package service

import (
	"context"
	"errors"
	"fmt"
	"service-nest/interfaces"
	"service-nest/model"
	"service-nest/repository"
	"service-nest/util"
)

type UserService struct {
	userRepo interfaces.UserRepository
	otpRepo  repository.OtpRepository
}

func NewUserService(userRepo interfaces.UserRepository, otpRepo *repository.OtpRepository) interfaces.UserService {
	return &UserService{userRepo: userRepo, otpRepo: *otpRepo}
}

func (s *UserService) ViewProfileByID(ctx context.Context, userID string) (*model.User, error) {
	user, err := s.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("could not find user: %v", err)
	}

	if user == nil {
		return nil, fmt.Errorf("user not found")
	}

	return user, nil
}

func (s *UserService) UpdateUser(ctx context.Context, userID string, newEmail, newPassword, newAddress, newPhone *string) error {
	user, err := s.userRepo.GetUserByID(ctx, userID)
	oldEmail := user.Email
	if err != nil {
		return fmt.Errorf("could not find user: %v", err)
	}

	// Update email
	if newEmail != nil {
		if err := util.ValidateEmail(*newEmail); err != nil {
			return err
		}
		existingUser, err := s.userRepo.GetUserByEmail(ctx, *newEmail)
		if err == nil && existingUser.ID != userID {
			return errors.New("email already in use by another user")
		}
		user.Email = *newEmail
	}

	// Update password
	if newPassword != nil {
		if err := util.ValidatePassword(*newPassword); err != nil {
			return err
		}
		user.Password = *newPassword
	}

	// Update contact
	if newPhone != nil {
		if err := util.ValidatePhoneNumber(*newPhone); err != nil {
			return err
		}
		user.Contact = *newPhone
	}
	// Update address
	if newAddress != nil {
		user.Address = *newAddress
	}

	// Save the updated user back to the repository_test
	if err := s.userRepo.UpdateUser(ctx, user, oldEmail); err != nil {
		return fmt.Errorf("could not update user: %v", err)
	}

	return nil
}

func (s *UserService) CheckUserExists(ctx context.Context, email string) (*model.User, error) {
	user, err := s.userRepo.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, fmt.Errorf("could not find user: %v", err)
	}
	return user, nil
}

func (s *UserService) CreateUser(ctx context.Context, user *model.User) error {
	user.ID = util.GenerateUUID()
	err := s.userRepo.SaveUser(ctx, user)
	if err != nil {
		return fmt.Errorf("could not save user: %v", err)
	}
	return nil
}

func (s *UserService) GenerateOtp(ctx context.Context, email string) error {
	_, err := s.userRepo.GetUserByEmail(ctx, email)
	if err != nil {
		return err
	}
	otp, err := s.otpRepo.GenerateOTP()
	if err != nil {
		return err
	}
	err = s.userRepo.SaveOTP(ctx, email, otp)
	if err != nil {
		return err
	}
	return util.SendOTPEmail(email, otp)
}

func (s *UserService) VerifyAndUpdatePassword(ctx context.Context, email, otp string, password string) error {
	valid, err := s.userRepo.ValidateOTP(ctx, email, otp)
	if err != nil {
		return err
	}
	if valid == false {
		return errors.New("invalid otp")
	}
	user, err := s.userRepo.GetUserByEmail(ctx, email)
	user.Password = password
	err = s.userRepo.UpdateUser(ctx, user, email)
	if err != nil {
		return err
	}

	return nil
}
