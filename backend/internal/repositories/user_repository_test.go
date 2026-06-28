package repositories

import (
	"testing"

	"recipe-app/internal/models"
)

func sampleUser() *models.User {
	return &models.User{
		Email:     "chef@example.com",
		Username:  "chef",
		FirstName: "Remy",
		LastName:  "Ratatouille",
		Password:  "hashed-password",
	}
}

func TestUserRepository_CreateAndLookup(t *testing.T) {
	db := newTestDB(t)
	repo := NewUserRepository(db)

	u := sampleUser()
	if err := repo.CreateUser(u); err != nil {
		t.Fatalf("CreateUser: %v", err)
	}
	if u.ID == "" {
		t.Fatal("expected generated user ID")
	}

	byID, err := repo.GetUserByID(u.ID)
	if err != nil {
		t.Fatalf("GetUserByID: %v", err)
	}
	if byID.Email != u.Email {
		t.Errorf("email = %q, want %q", byID.Email, u.Email)
	}

	byEmail, err := repo.GetUserByEmail(u.Email)
	if err != nil {
		t.Fatalf("GetUserByEmail: %v", err)
	}
	if byEmail.ID != u.ID {
		t.Errorf("id = %q, want %q", byEmail.ID, u.ID)
	}

	byUsername, err := repo.GetUserByUsername(u.Username)
	if err != nil {
		t.Fatalf("GetUserByUsername: %v", err)
	}
	if byUsername.ID != u.ID {
		t.Errorf("id = %q, want %q", byUsername.ID, u.ID)
	}
}

func TestUserRepository_ExistenceChecks(t *testing.T) {
	db := newTestDB(t)
	repo := NewUserRepository(db)

	u := sampleUser()
	if err := repo.CreateUser(u); err != nil {
		t.Fatalf("CreateUser: %v", err)
	}

	tests := []struct {
		name  string
		check func() (bool, error)
		want  bool
	}{
		{"existing email", func() (bool, error) { return repo.EmailExists("chef@example.com") }, true},
		{"missing email", func() (bool, error) { return repo.EmailExists("nobody@example.com") }, false},
		{"existing username", func() (bool, error) { return repo.UsernameExists("chef") }, true},
		{"missing username", func() (bool, error) { return repo.UsernameExists("ghost") }, false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := tc.check()
			if err != nil {
				t.Fatalf("check: %v", err)
			}
			if got != tc.want {
				t.Errorf("got %v, want %v", got, tc.want)
			}
		})
	}
}

func TestUserRepository_DuplicateEmailFails(t *testing.T) {
	db := newTestDB(t)
	repo := NewUserRepository(db)

	if err := repo.CreateUser(sampleUser()); err != nil {
		t.Fatalf("first CreateUser: %v", err)
	}

	dup := sampleUser()
	dup.Username = "different"
	if err := repo.CreateUser(dup); err == nil {
		t.Fatal("expected duplicate email to fail the UNIQUE constraint")
	}
}

func TestUserRepository_GetByID_NotFound(t *testing.T) {
	db := newTestDB(t)
	repo := NewUserRepository(db)

	if _, err := repo.GetUserByID("missing"); err == nil {
		t.Fatal("expected error for missing user")
	}
}

func TestUserRepository_Update(t *testing.T) {
	db := newTestDB(t)
	repo := NewUserRepository(db)

	u := sampleUser()
	if err := repo.CreateUser(u); err != nil {
		t.Fatalf("CreateUser: %v", err)
	}

	u.FirstName = "Alfredo"
	u.AvatarURL = "/img/avatar.png"
	if err := repo.UpdateUser(u); err != nil {
		t.Fatalf("UpdateUser: %v", err)
	}

	got, err := repo.GetUserByID(u.ID)
	if err != nil {
		t.Fatalf("GetUserByID: %v", err)
	}
	if got.FirstName != "Alfredo" || got.AvatarURL != "/img/avatar.png" {
		t.Errorf("update not persisted: %+v", got)
	}
}

func TestUserRepository_UpdatePassword(t *testing.T) {
	db := newTestDB(t)
	repo := NewUserRepository(db)

	u := sampleUser()
	if err := repo.CreateUser(u); err != nil {
		t.Fatalf("CreateUser: %v", err)
	}

	if err := repo.UpdatePassword(u.ID, "new-hash"); err != nil {
		t.Fatalf("UpdatePassword: %v", err)
	}

	got, err := repo.GetUserByEmail(u.Email)
	if err != nil {
		t.Fatalf("GetUserByEmail: %v", err)
	}
	if got.Password != "new-hash" {
		t.Errorf("password hash = %q, want new-hash", got.Password)
	}
}

func TestUserRepository_Delete(t *testing.T) {
	db := newTestDB(t)
	repo := NewUserRepository(db)

	u := sampleUser()
	if err := repo.CreateUser(u); err != nil {
		t.Fatalf("CreateUser: %v", err)
	}

	if err := repo.DeleteUser(u.ID); err != nil {
		t.Fatalf("DeleteUser: %v", err)
	}

	if _, err := repo.GetUserByID(u.ID); err == nil {
		t.Fatal("expected user to be gone")
	}
}
