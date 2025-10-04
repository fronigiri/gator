package main

import (
	"context"
	"fmt"
	"time"

	"github.com/fronigiri/gator/internal/database"
	"github.com/google/uuid"
)

func handlerAddFeed(s *state, cmd command, user database.User) error {
	if len(cmd.Args) != 2 {
		return fmt.Errorf("usage: %s <name> <url>", cmd.Name)
	}
	parms := database.CreateFeedParams{
		ID:        uuid.New(),
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
		Name:      cmd.Args[0],
		Url:       cmd.Args[1],
		UserID:    user.ID,
	}

	feed, err := s.db.CreateFeed(context.Background(), parms)
	if err != nil {
		return fmt.Errorf("couldn't create feed: %w", err)
	}

	params2 := database.CreateFeedFollowParams{
		ID:        uuid.New(),
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
		UserID:    user.ID,
		FeedID:    feed.ID,
	}

	_, err2 := s.db.CreateFeedFollow(context.Background(), params2)
	if err2 != nil {
		return fmt.Errorf("unable to create follow feed: %v", err)
	}

	fmt.Println("Feed created successfully:")
	printFeed(feed)
	fmt.Println()
	fmt.Println("=====================================")

	return nil
}

func printFeed(feed database.Feed) {
	fmt.Printf("* ID:            %s\n", feed.ID)
	fmt.Printf("* Created:       %v\n", feed.CreatedAt)
	fmt.Printf("* Updated:       %v\n", feed.UpdatedAt)
	fmt.Printf("* Name:          %s\n", feed.Name)
	fmt.Printf("* URL:           %s\n", feed.Url)
	fmt.Printf("* UserID:        %s\n", feed.UserID)
}

func handlerListFeeds(s *state, cmd command) error {
	feedList, err := s.db.GetFeeds(context.Background())
	if err != nil {
		return fmt.Errorf("couldn't find feeds: %v", err)
	}
	for _, feed := range feedList {
		name, err := s.db.GetUserID(context.Background(), feed.UserID)
		if err != nil {
			return fmt.Errorf("name not found: %v", err)
		}
		fmt.Printf("* Name:          %s\n", feed.Name)
		fmt.Printf("* URL:           %s\n", feed.Url)
		fmt.Printf("* User:          %s\n", name)
	}
	return nil
}

func handlerFollow(s *state, cmd command, user database.User) error {
	if len(cmd.Args) != 1 {
		return fmt.Errorf("usage: %s <url>", cmd.Name)
	}
	url := cmd.Args[0]
	feed, err := s.db.GetFeedFromURL(context.Background(), url)
	if err != nil {
		return fmt.Errorf("unable to get feed: %v", err)
	}

	params := database.CreateFeedFollowParams{
		ID:        uuid.New(),
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
		UserID:    user.ID,
		FeedID:    feed.ID,
	}

	rows, err := s.db.CreateFeedFollow(context.Background(), params)
	if err != nil {
		return fmt.Errorf("unable to create follow feed: %v", err)
	}

	if len(rows) == 0 {
		return fmt.Errorf("no follow row returned")
	}
	follow := rows[0]

	fmt.Println("Feed follow added successfully")

	fmt.Printf("* Name:          %s\n", follow.FeedName)
	fmt.Printf("* User:          %s\n", follow.UserName)
	return nil
}

func handlerFollowing(s *state, cmd command, user database.User) error {
	followingList, err := s.db.GetFeedFollowsForUser(context.Background(), user.ID)

	if err != nil {
		return fmt.Errorf("unable to retreive feeds: %v", err)
	}

	for _, name := range followingList {
		fmt.Printf("%s\n", name)
	}
	return nil
}

func handlerUnfollow(s *state, cmd command, user database.User) error {
	feedURL := cmd.Args[0]
	feed, err := s.db.GetFeedFromURL(context.Background(), feedURL)
	if err != nil {
		return fmt.Errorf("unable to retieve feed: %w", err)
	}

	params := database.DelFeedFollowParams{
		UserID: user.ID,
		FeedID: feed.ID,
	}

	_, err2 := s.db.DelFeedFollow(context.Background(), params)
	if err2 != nil {
		return fmt.Errorf("unable to unfollow current feed: %w", err2)
	}
	fmt.Println("unfollowed feed successfully")
	return nil
}
