package models

import "github.com/google/uuid"

type Employee struct {
	ID        uuid.UUID
	Identity  string
	Email     string
	Name      string
	IsActive  bool
	RoleTitle string
}
