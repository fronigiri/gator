package main

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"time"

	"github.com/fronigiri/blog-aggregator/internal/database"
	"github.com/google/uuid"
)

func handlerLogin(s *state, cmd command) error {
	if len(cmd.Args) != 1 {
		return fmt.Errorf("usage: %s <name>", cmd.Name)
	}
	name := cmd.Args[0]

	_, err := s.db.GetUser(context.Background(), name)
	switch err {
	case sql.ErrNoRows:
		os.Exit(1)
	case nil:
		err := s.cfg.SetUser(name)

		if err != nil {
			return fmt.Errorf("couldn't set current user: %w", err)
		}

		fmt.Println("User switched successfully!")
	default:
		return err
	}
	return nil
}

func handlerRegister(s *state, cmd command) error {
	if len(cmd.Args) != 1 {
		return fmt.Errorf("usage: %s <name>", cmd.Name)
	}
	name := cmd.Args[0]
	_, err := s.db.GetUser(context.Background(), name)

	switch err {
	case sql.ErrNoRows:
		params := database.CreateUserParams{
			ID:        uuid.New(),
			CreatedAt: time.Now().UTC(),
			UpdatedAt: time.Now().UTC(),
			Name:      name,
		}
		user, createErr := s.db.CreateUser(context.Background(), params)
		if createErr != nil {
			return createErr
		}
		err := s.cfg.SetUser(name)

		if err != nil {
			return fmt.Errorf("couldn't set current user: %w", err)
		}
		fmt.Printf("User has been created successfully!")
		printUser(user)

	case nil:
		return fmt.Errorf("user with name '%s' already exists", name)
	default:
		return err
	}
	return nil
}

func handlerGetUsers(s *state, cmd command) error {
	userList, err := s.db.GetUsers(context.Background())
	if err != nil {
		return fmt.Errorf("unable to retrieve users: %w", err)
	}
	for _, user := range userList {
		if user == s.cfg.CurrentUserName {
			fmt.Printf("* %s (current)\n", user)
		} else {
			fmt.Printf("* %s\n", user)
		}
	}
	return nil
}

func printUser(user database.User) {
	fmt.Printf(" * ID:      %v\n", user.ID)
	fmt.Printf(" * Name:    %v\n", user.Name)
}
