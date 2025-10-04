package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/lib/pq"

	"github.com/fronigiri/blog-aggregator/internal/database"
	"github.com/google/uuid"
)

var ErrNoFeeds = errors.New("no feeds")

func handlerAgg(s *state, cmd command) error {
	if len(cmd.Args) != 1 {
		return fmt.Errorf("usage: %s <interval>", cmd.Name)
	}

	interval := cmd.Args[0]
	fmt.Printf("Collecting feeds every %s\n", interval)

	timeBetweenRequests, err := time.ParseDuration(interval)
	if err != nil {
		return fmt.Errorf("unable to parse provided time")
	}

	ctx := context.Background()
	ticker := time.NewTicker(timeBetweenRequests)
	for ; ; <-ticker.C {
		for {
			err := scrapeFeeds(s, ctx)
			if err == nil {
				continue
			}
			if errors.Is(err, ErrNoFeeds) {
				break
			}
			log.Printf("scrape error: %v", err)
			break
		}
	}
}

func scrapeFeeds(s *state, ctx context.Context) error {
	feed, err := s.db.GetNextFeedToFetch(ctx)

	if err == sql.ErrNoRows {
		return ErrNoFeeds
	}

	if err != nil {
		return fmt.Errorf("unable to fetch feed: %w", err)
	}

	if _, err := s.db.MarkFeedFetched(ctx, feed.ID); err != nil {
		return fmt.Errorf("unable to mark feed: %w", err)
	}

	rss, err3 := fetchFeed(ctx, feed.Url)
	if err3 != nil {
		return fmt.Errorf("unable to get feed from url: %w", err3)
	}

	for _, item := range rss.Channel.Item {
		desc := item.Description
		if desc == "" {
			desc = ""
		}
		now := time.Now().UTC()
		pub := now
		if t, err := time.Parse(time.RFC1123, item.PubDate); err == nil {
			pub = t.UTC()
		}

		post := database.CreatePostParams{
			ID:          uuid.New(),
			CreatedAt:   now,
			UpdatedAt:   now,
			Title:       item.Title,
			Url:         item.Link,
			Description: desc,
			PublishedAt: pub,
			FeedID:      feed.ID,
		}
		if _, err := s.db.CreatePost(ctx, post); err != nil {
			var pqErr *pq.Error
			if errors.As(err, &pqErr) && pqErr.Code == "23505" {
				continue
			} else {
				log.Printf("create post failed: %v", err)
				continue
			}
		}
		fmt.Printf("Post Scrapped! Title: %s\n", post.Title)
	}

	return nil
}

func handlerBrowse(s *state, cmd command) error {
	limit := int32(2)
	if len(cmd.Args) >= 1 {
		if n, err := strconv.Atoi(cmd.Args[0]); err == nil {
			limit = int32(n)
		}
	}
	u, err := s.db.GetUser(context.Background(), s.cfg.CurrentUserName)
	if err != nil {
		return err
	}
	userID := u.ID

	params := database.GetPostsFromUserParams{
		UserID: userID,
		Limit:  limit,
	}
	posts, err := s.db.GetPostsFromUser(context.Background(), params)
	if err != nil {
		return fmt.Errorf("unable to get user posts:%v", err)
	}
	printPosts(posts)
	return nil
}

func printPosts(posts []database.GetPostsFromUserRow) {
	for _, p := range posts {
		fmt.Printf("Feed: %s\n", p.FeedName)
		fmt.Printf("Title: %s\n", p.Title)
		fmt.Printf("URL: %s\n", p.Url)
		fmt.Printf("Published: %s\n\n", p.PublishedAt.UTC().Format(time.RFC3339))
	}
}
