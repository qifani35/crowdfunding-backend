package user

import (
	"errors"
	"testing"

	"golang.org/x/crypto/bcrypt"
)

// MockRepository implements Repository interface for testing
type MockRepository struct {
	users      map[int]User
	nextID     int
	shouldFail bool
}

func NewMockRepository() *MockRepository {
	return &MockRepository{
		users:  make(map[int]User),
		nextID: 1,
	}
}

func (m *MockRepository) Save(user User) (User, error) {
	if m.shouldFail {
		return User{}, errors.New("mock save error")
	}
	user.ID = m.nextID
	m.nextID++
	m.users[user.ID] = user
	return user, nil
}

func (m *MockRepository) FindByEmail(email string) (User, error) {
	if m.shouldFail {
		return User{}, errors.New("mock find error")
	}
	for _, u := range m.users {
		if u.Email == email {
			return u, nil
		}
	}
	return User{}, nil
}

func (m *MockRepository) FindByID(ID int) (User, error) {
	if m.shouldFail {
		return User{}, errors.New("mock find error")
	}
	user, exists := m.users[ID]
	if !exists {
		return User{}, errors.New("user not found")
	}
	return user, nil
}

func (m *MockRepository) UpdateUser(user User) (User, error) {
	if m.shouldFail {
		return User{}, errors.New("mock update error")
	}
	m.users[user.ID] = user
	return user, nil
}

func (m *MockRepository) FindAll() ([]User, error) {
	if m.shouldFail {
		return nil, errors.New("mock find all error")
	}
	users := make([]User, 0, len(m.users))
	for _, u := range m.users {
		users = append(users, u)
	}
	return users, nil
}

func (m *MockRepository) DeleteUser(ID int) error {
	if m.shouldFail {
		return errors.New("mock delete error")
	}
	delete(m.users, ID)
	return nil
}

func (m *MockRepository) SeedUser(user User) {
	user.ID = m.nextID
	m.nextID++
	m.users[user.ID] = user
}

// Tests

func TestRegisterUser(t *testing.T) {
	repo := NewMockRepository()
	service := NewService(repo)

	input := RegisterUserInput{
		Name:       "John Doe",
		Email:      "john@example.com",
		Occupation: "Developer",
		Password:   "password123",
	}

	user, err := service.RegisterUser(input)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if user.ID == 0 {
		t.Error("Expected non-zero user ID")
	}
	if user.Name != input.Name {
		t.Errorf("Expected name '%s', got '%s'", input.Name, user.Name)
	}
	if user.Email != input.Email {
		t.Errorf("Expected email '%s', got '%s'", input.Email, user.Email)
	}
	if user.Role != "user" {
		t.Errorf("Expected role 'user', got '%s'", user.Role)
	}

	// Verify password is hashed
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(input.Password))
	if err != nil {
		t.Error("Password should be properly hashed")
	}
}

func TestRegisterUser_DBError(t *testing.T) {
	repo := NewMockRepository()
	repo.shouldFail = true
	service := NewService(repo)

	input := RegisterUserInput{
		Name:       "John Doe",
		Email:      "john@example.com",
		Occupation: "Developer",
		Password:   "password123",
	}

	_, err := service.RegisterUser(input)
	if err == nil {
		t.Error("Expected error from failing repository")
	}
}

func TestLoginUser_Success(t *testing.T) {
	repo := NewMockRepository()
	service := NewService(repo)

	// Register first
	registerInput := RegisterUserInput{
		Name:       "John Doe",
		Email:      "john@example.com",
		Occupation: "Developer",
		Password:   "password123",
	}
	registeredUser, _ := service.RegisterUser(registerInput)

	// Login
	loginInput := LoginUserInput{
		Email:    "john@example.com",
		Password: "password123",
	}

	user, err := service.LoginUser(loginInput)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if user.ID != registeredUser.ID {
		t.Errorf("Expected user ID %d, got %d", registeredUser.ID, user.ID)
	}
}

func TestLoginUser_WrongPassword(t *testing.T) {
	repo := NewMockRepository()
	service := NewService(repo)

	// Register first
	service.RegisterUser(RegisterUserInput{
		Name:       "John Doe",
		Email:      "john@example.com",
		Occupation: "Developer",
		Password:   "password123",
	})

	// Login with wrong password
	loginInput := LoginUserInput{
		Email:    "john@example.com",
		Password: "wrongpassword",
	}

	_, err := service.LoginUser(loginInput)
	if err == nil {
		t.Error("Expected error for wrong password")
	}
}

func TestLoginUser_EmailNotFound(t *testing.T) {
	repo := NewMockRepository()
	service := NewService(repo)

	loginInput := LoginUserInput{
		Email:    "nonexistent@example.com",
		Password: "password123",
	}

	_, err := service.LoginUser(loginInput)
	if err == nil {
		t.Error("Expected error for non-existent email")
	}
}

func TestIsEmailAvailable_Available(t *testing.T) {
	repo := NewMockRepository()
	service := NewService(repo)

	input := CheckEmailInput{Email: "new@example.com"}
	available, err := service.IsEmailAvailable(input)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if !available {
		t.Error("Expected email to be available")
	}
}

func TestIsEmailAvailable_Taken(t *testing.T) {
	repo := NewMockRepository()
	service := NewService(repo)

	// Register a user first
	service.RegisterUser(RegisterUserInput{
		Name:       "John Doe",
		Email:      "john@example.com",
		Occupation: "Developer",
		Password:   "password123",
	})

	input := CheckEmailInput{Email: "john@example.com"}
	available, err := service.IsEmailAvailable(input)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if available {
		t.Error("Expected email to be unavailable")
	}
}

func TestGetUserByID(t *testing.T) {
	repo := NewMockRepository()
	service := NewService(repo)

	// Register a user
	registered, _ := service.RegisterUser(RegisterUserInput{
		Name:       "John Doe",
		Email:      "john@example.com",
		Occupation: "Developer",
		Password:   "password123",
	})

	user, err := service.GetUserByID(registered.ID)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if user.ID != registered.ID {
		t.Errorf("Expected user ID %d, got %d", registered.ID, user.ID)
	}
}

func TestFormatUser(t *testing.T) {
	user := User{
		ID:         1,
		Name:       "John Doe",
		Occupation: "Developer",
		Email:      "john@example.com",
	}

	formatter := FormatUser(user, "test-token")
	if formatter.ID != user.ID {
		t.Errorf("Expected ID %d, got %d", user.ID, formatter.ID)
	}
	if formatter.Name != user.Name {
		t.Errorf("Expected name '%s', got '%s'", user.Name, formatter.Name)
	}
	if formatter.Token != "test-token" {
		t.Errorf("Expected token 'test-token', got '%s'", formatter.Token)
	}
}
