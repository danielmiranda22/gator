-- name: CreatePostLike :one
INSERT INTO
    post_likes (
        id,
        created_at,
        updated_at,
        post_id,
        user_id
    )
VALUES ($1, $2, $3, $4, $5)
RETURNING
    *;

-- name: DeletePostLike :exec
DELETE FROM post_likes WHERE post_id = $1 AND user_id = $2;

-- name: GetPostLikeByPostAndUser :one
SELECT * FROM post_likes WHERE post_id = $1 AND user_id = $2;

-- name: GetLikedPostsForUser :many
SELECT posts.*, feeds.name AS feed_name
FROM post_likes
    JOIN posts ON post_likes.post_id = posts.id
    JOIN feeds ON posts.feed_id = feeds.id
WHERE
    post_likes.user_id = $1
ORDER BY post_likes.created_at DESC;

-- name: CountLikesForPost :one
SELECT COUNT(*)::BIGINT FROM post_likes WHERE post_id = $1;
