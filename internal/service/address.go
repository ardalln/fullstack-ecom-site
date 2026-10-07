package service

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"shop-api/internal/domain"
)

var postalCodePattern = regexp.MustCompile(`^\d{10}$`)

// AddressInput is the data needed to create or replace a saved address.
type AddressInput struct {
	ReceiverName  string
	ReceiverPhone string
	Province      string
	City          string
	AddressLine   string
	PostalCode    string
}

func (in AddressInput) normalize() (domain.Address, error) {
	phone, err := domain.NormalizePhone(in.ReceiverPhone)
	if err != nil {
		return domain.Address{}, err
	}

	postal := strings.TrimSpace(in.PostalCode)
	if !postalCodePattern.MatchString(postal) {
		return domain.Address{}, fmt.Errorf("%w: postal code must be exactly 10 digits", domain.ErrInvalidInput)
	}

	name := strings.TrimSpace(in.ReceiverName)
	province := strings.TrimSpace(in.Province)
	city := strings.TrimSpace(in.City)
	line := strings.TrimSpace(in.AddressLine)
	if name == "" || province == "" || city == "" || line == "" {
		return domain.Address{}, fmt.Errorf("%w: receiver name, province, city and address are required", domain.ErrInvalidInput)
	}

	return domain.Address{
		ReceiverName: name, ReceiverPhone: phone, Province: province, City: city,
		AddressLine: line, PostalCode: postal,
	}, nil
}

type AddressService struct {
	addresses domain.AddressRepository
}

func NewAddressService(addresses domain.AddressRepository) *AddressService {
	return &AddressService{addresses: addresses}
}

func (s *AddressService) List(ctx context.Context, userID int64) ([]domain.Address, error) {
	return s.addresses.ListByUserID(ctx, userID)
}

func (s *AddressService) Create(ctx context.Context, userID int64, in AddressInput) (*domain.Address, error) {
	a, err := in.normalize()
	if err != nil {
		return nil, err
	}
	a.UserID = userID
	if err := s.addresses.Create(ctx, &a); err != nil {
		return nil, err
	}
	return &a, nil
}

func (s *AddressService) Update(ctx context.Context, userID, id int64, in AddressInput) (*domain.Address, error) {
	a, err := in.normalize()
	if err != nil {
		return nil, err
	}
	a.ID = id
	a.UserID = userID
	if err := s.addresses.Update(ctx, &a); err != nil {
		return nil, err
	}
	return &a, nil
}

func (s *AddressService) Delete(ctx context.Context, userID, id int64) error {
	return s.addresses.Delete(ctx, userID, id)
}

func (s *AddressService) SetDefault(ctx context.Context, userID, id int64) error {
	return s.addresses.SetDefault(ctx, userID, id)
}
