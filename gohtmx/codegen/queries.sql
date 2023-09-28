-- name: Users :many
SELECT * FROM users;

-- name: GetUser :one
SELECT * FROM users WHERE username=$1 LIMIT 1;

-- name: GetThread :one
SELECT * FROM threads WHERE id=$1 LIMIT 1;

-- name: CreateUser :one
INSERT INTO users (
    username,
    password
) VALUES ($1, $2)
RETURNING *;

-- name: CreatePost :one 
INSERT INTO threads (
    id,
    root_id,
    author_id,
    title,
    content
) VALUES ($1, $2, $3, $4, $5) 
RETURNING *;

-- name: CreateComment :one 
INSERT INTO threads (
    root_id,
    parent_id,
    author_id,
    content
) VALUES ($1, $2, $3, $4) 
RETURNING *;

-- name: CreateVote :one
INSERT INTO votes (
    voter_id,
    thread_id,
    upvote
) VALUES ($1, $2, $3)
ON CONFLICT ON CONSTRAINT votes_voter_id_thread_id_key DO UPDATE SET upvote = $3 WHERE votes.voter_id = $1 AND votes.thread_id = $2
RETURNING *;

-- name: DeleteVote :exec
DELETE FROM votes WHERE voter_id=$1 AND thread_id=$2;

-- name: GetThreads :many
WITH selected_threads AS (
	SELECT * FROM threads thread 
    WHERE thread.author_id = coalesce(sqlc.narg('author_id'), thread.author_id) -- filter threads from one author
    AND thread.root_id = coalesce(sqlc.narg('root_id'), thread.root_id) -- filter threads from one root conversation
    AND (NOT @root_threads_only::bool OR thread.parent_id IS NULL) -- if root threads only then thread.parent_id must be null
    AND (NOT @child_threads_only::bool OR thread.parent_id IS NOT NULL) -- if child threads only then thread.parent_id must be not null
	ORDER BY thread.created_at DESC
),
reply_counts AS (
	SELECT parent_id as thread_id, COUNT(*) as reply_count
	FROM threads thread
	WHERE thread.parent_id IN (SELECT id FROM selected_threads)
	GROUP BY thread.parent_id
),
vote_counts AS (
	SELECT v.thread_id as thread_id,
	COUNT(CASE WHEN upvote THEN 1 ELSE NULL END) as upvote_count, 
	COUNT(CASE WHEN NOT upvote THEN 1 ELSE NULL END) as downvote_count 
	FROM votes v
	WHERE v.thread_id IN (SELECT id FROM selected_threads)
	GROUP BY v.thread_id
)
SELECT 
    -- thread info
	st.*,
    -- extended info
	u.username,
	coalesce(rc.reply_count, 0) as reply_count,
	coalesce(vc.upvote_count, 0) as upvote_count,
	coalesce(vc.downvote_count, 0) as downvote_count,
	EXISTS (SELECT 1 FROM votes v1 WHERE v1.thread_id = st.id AND v1.voter_id = sqlc.narg('reader_id') AND v1.upvote = TRUE) as has_upvoted,
    EXISTS (SELECT 1 FROM votes v2 WHERE v2.thread_id = st.id AND v2.voter_id = sqlc.narg('reader_id') AND v2.upvote = FALSE) as has_downvoted
FROM selected_threads st
LEFT JOIN reply_counts rc ON st.id = rc.thread_id
LEFT JOIN vote_counts vc ON st.id = vc.thread_id
INNER JOIN users u ON st.author_id = u.id;


-- name: GetThreadExtended :one
WITH selected_threads AS (
	SELECT * FROM threads thread 
    WHERE thread.id = sqlc.arg('id')
    LIMIT 1
),
reply_counts AS (
	SELECT parent_id as thread_id, COUNT(*) as reply_count
	FROM threads thread
	WHERE thread.parent_id IN (SELECT id FROM selected_threads)
	GROUP BY thread.parent_id
),
vote_counts AS (
	SELECT v.thread_id as thread_id,
	COUNT(CASE WHEN upvote THEN 1 ELSE NULL END) as upvote_count, 
	COUNT(CASE WHEN NOT upvote THEN 1 ELSE NULL END) as downvote_count 
	FROM votes v
	WHERE v.thread_id IN (SELECT id FROM selected_threads)
	GROUP BY v.thread_id
)
SELECT 
    -- thread info
	st.*,
    -- extended info
	u.username,
	coalesce(rc.reply_count, 0) as reply_count,
	coalesce(vc.upvote_count, 0) as upvote_count,
	coalesce(vc.downvote_count, 0) as downvote_count,
	EXISTS (SELECT 1 FROM votes v1 WHERE v1.thread_id = st.id AND v1.voter_id = sqlc.narg('reader_id') AND v1.upvote = TRUE) as has_upvoted,
    EXISTS (SELECT 1 FROM votes v2 WHERE v2.thread_id = st.id AND v2.voter_id = sqlc.narg('reader_id') AND v2.upvote = FALSE) as has_downvoted
FROM selected_threads st
LEFT JOIN reply_counts rc ON st.id = rc.thread_id
LEFT JOIN vote_counts vc ON st.id = vc.thread_id
INNER JOIN users u ON st.author_id = u.id;