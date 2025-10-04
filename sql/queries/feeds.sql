-- name: CreateFeed :one
INSERT INTO feeds (id, created_at, updated_at, name, url, user_id)
VALUES ( 
    $1,
    $2,
    $3,
    $4,
    $5,
    $6
)
RETURNING *;


-- name: GetFeeds :many
SELECT name, url, user_id FROM feeds;


-- name: CreateFeedFollow :many

WITH inserted_feed_follow AS (
    INSERT INTO feed_follows (id, created_at, updated_at, user_id, feed_id)
    VALUES (
        $1,
        $2,
        $3,
        $4,
        $5
    )
    ON CONFLICT (user_id, feed_id) DO NOTHING
    RETURNING *
)
SELECT
    inserted_feed_follow.*,
    feeds.name AS feed_name,
    users.name AS user_name
FROM inserted_feed_follow
INNER JOIN users
ON inserted_feed_follow.user_id = users.id
INNER JOIN feeds
ON inserted_feed_follow.feed_id = feeds.id;

-- name: GetFeedFromURL :one
SELECT * FROM feeds
WHERE url = $1;


-- name: GetFeedFollowsForUser :many
SELECT feeds.name FROM feed_follows
JOIN feeds
ON feed_follows.feed_id = feeds.id
WHERE feed_follows.user_id = $1;


-- name: DelFeedFollow :one
DELETE FROM feed_follows
WHERE user_id = $1 AND feed_id = $2
RETURNING *;


-- name: MarkFeedFetched :one
UPDATE feeds
SET last_fetched_at = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP
WHERE id = $1
RETURNING *;

-- name: GetNextFeedToFetch :one

SELECT id, url FROM feeds
ORDER BY last_fetched_at ASC NULLS FIRST, 
updated_at ASC,
id ASC
LIMIT 1;
