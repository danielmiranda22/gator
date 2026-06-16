-- name: CreatePost :one
INSERT INTO
    posts (
        id,
        created_at,
        updated_at,
        title,
        url,
        description,
        published_at,
        feed_id
    )
VALUES (
        $1,
        $2,
        $3,
        $4,
        $5,
        $6,
        $7,
        $8
    )
RETURNING
    *;
--

-- name: GetPostsForUser :many
SELECT posts.*, feeds.name AS feed_name
FROM
    posts
    JOIN feed_follows ON feed_follows.feed_id = posts.feed_id
    JOIN feeds ON posts.feed_id = feeds.id
WHERE
    feed_follows.user_id = $1
ORDER BY posts.published_at DESC
LIMIT $2;
--

-- name: GetPostsForUserOldest :many
SELECT posts.*, feeds.name AS feed_name
FROM
    posts
    JOIN feed_follows ON feed_follows.feed_id = posts.feed_id
    JOIN feeds ON posts.feed_id = feeds.id
WHERE
    feed_follows.user_id = $1
ORDER BY posts.published_at ASC
LIMIT $2;
--

-- name: GetPostsForUserFilterByTitle :many
SELECT posts.*, feeds.name AS feed_name
FROM
    posts
    JOIN feed_follows ON feed_follows.feed_id = posts.feed_id
    JOIN feeds ON posts.feed_id = feeds.id
WHERE
    feed_follows.user_id = $1
    AND (
        posts.title = $2
        OR posts.description = $2
    )
ORDER BY posts.published_at DESC
LIMIT $3;
--

-- name: GetPostsForUserWithPagination :many
SELECT posts.*, feeds.name AS feed_name
FROM
    posts
    JOIN feed_follows ON feed_follows.feed_id = posts.feed_id
    JOIN feeds ON posts.feed_id = feeds.id
WHERE
    feed_follows.user_id = $1
ORDER BY posts.published_at DESC
LIMIT $2
OFFSET
    $3;
--

-- name: GetPostsForUserOldestWithPagination :many
SELECT posts.*, feeds.name AS feed_name
FROM
    posts
    JOIN feed_follows ON feed_follows.feed_id = posts.feed_id
    JOIN feeds ON posts.feed_id = feeds.id
WHERE
    feed_follows.user_id = $1
ORDER BY posts.published_at ASC
LIMIT $2
OFFSET
    $3;
--

-- name: SearchPostsForUser :many
SELECT posts.*, feeds.name AS feed_name
FROM
    posts
    JOIN feed_follows ON feed_follows.feed_id = posts.feed_id
    JOIN feeds ON posts.feed_id = feeds.id
WHERE
    feed_follows.user_id = sqlc.arg (user_id)
    AND (
        posts.title ILIKE '%' || sqlc.arg (term) || '%'
        OR posts.description ILIKE '%' || sqlc.arg (term) || '%'
    )
ORDER BY posts.published_at DESC
LIMIT sqlc.arg (limit_count)
OFFSET
    sqlc.arg (offset_count);

-- name: GetPostByID :one
SELECT posts.*, feeds.name AS feed_name
FROM posts
    JOIN feeds ON posts.feed_id = feeds.id
WHERE
    posts.id = $1;
