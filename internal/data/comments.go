package data

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/martinezmoises/comments/internal/validator"
)

type Comment struct {
	ID        int64     `json:"id"`
	Content   string    `json:"content"`
	Author    string    `json:"author"`
	CreatedAt time.Time `json:"-"`
	Version   int32     `json:"version"`
}

type CommentModel struct {
	DB *sql.DB
}

func ValidateComment(v *validator.Validator, comment *Comment) {
	// check if the Content field is empty
	v.Check(comment.Content != "", "content", "must be provided")
	// check if the Author field is empty
	v.Check(comment.Author != "", "author", "must be provided")
	// check if the Content field is empty
	v.Check(len(comment.Content) <= 100, "content", "must not be more than 100 bytes long")
	// check if the Author field is empty
	v.Check(len(comment.Author) <= 25, "author", "must not be more than 25 bytes long")

}

// Insert a new row in the comments table
// Expects a pointer to the actual comment
func (c CommentModel) Insert(comment *Comment) error {
	// the SQL query to be executed against the database table
	query := `
		 INSERT INTO comments (content, author)
		 VALUES ($1, $2)
		 RETURNING id, created_at, version
		 `
	// the actual values to replace $1, and $2
	args := []any{comment.Content, comment.Author}

	// Create a context with a 3-second timeout. No database
	// operation should take more than 3 seconds or we will quit it
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	// execute the query against the comments database table. We ask for the the
	// id, created_at, and version to be sent back to us which we will use
	// to update the Comment struct later on
	return c.DB.QueryRowContext(ctx, query, args...).Scan(&comment.ID, &comment.CreatedAt, &comment.Version)
}

// Get a specific Comment from the comments table
func (c CommentModel) Get(id int64) (*Comment, error) {
	// check if the id is valid
	if id < 1 {
		return nil, ErrRecordNotFound
	}
	// the SQL query to be executed against the database table
	query := `
		 SELECT id, created_at, content, author, version
		 FROM comments
		 WHERE id = $1
	   `
	// declare a variable of type Comment to store the returned comment
	var comment Comment

	// Set a 3-second context/timer
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := c.DB.QueryRowContext(ctx, query, id).Scan(&comment.ID, &comment.CreatedAt, &comment.Content, &comment.Author, &comment.Version)

	// check for which type of error
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}
	return &comment, nil

}

// Update a specific Comment from the comments table
func (c CommentModel) Update(comment *Comment) error {
	// The SQL query to be executed against the database table
	// Every time we make an update, we increment the version number
	query := `
			UPDATE comments
			SET content = $1, author = $2, version = version + 1
			WHERE id = $3
			RETURNING version
		  `

	args := []any{comment.Content, comment.Author, comment.ID}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	return c.DB.QueryRowContext(ctx, query, args...).Scan(&comment.Version)
}

// Delete a specific Comment from the comments table
func (c CommentModel) Delete(id int64) error {

	// check if the id is valid
	if id < 1 {
		return ErrRecordNotFound
	}
	// the SQL query to be executed against the database table
	query := `
        DELETE FROM comments
        WHERE id = $1
      `
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// ExecContext does not return any rows unlike QueryRowContext.
	// It only returns  information about the the query execution
	// such as how many rows were affected
	result, err := c.DB.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}
	// Were any rows  delete?
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	// Probably a wrong id was provided or the client is trying to
	// delete an already deleted comment
	if rowsAffected == 0 {
		return ErrRecordNotFound
	}

	return nil
}

// Delete a specific Comment from the comments table
func (c CommentModel) GetAll(content string, author string, filters Filters) ([]*Comment, Metadata, error) {

	// the SQL query to be executed against the database table
	query := fmt.Sprintf(`
		SELECT COUNT(*) OVER(), id, created_at, content, author, version
		FROM comments
		WHERE (to_tsvector('simple', content) @@
			plainto_tsquery('simple', $1) OR $1 = '') 
		AND (to_tsvector('simple', author) @@ 
			plainto_tsquery('simple', $2) OR $2 = '') 
			ORDER BY %s %s, id ASC 
			LIMIT $3 OFFSET $4`, filters.sortColumn(), filters.sortDirection())

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// QueryContext returns multiple rows.
	rows, err := c.DB.QueryContext(ctx, query, content, author, filters.limit(), filters.offset())
	if err != nil {
		return nil, Metadata{}, err
	}

	// clean up the memory that was used
	defer rows.Close()
	totalRecords := 0
	// we will store the address of each comment in our slice
	comments := []*Comment{}

	// process each row that is in rows
	for rows.Next() {
		var comment Comment
		err := rows.Scan(&totalRecords, &comment.ID, &comment.CreatedAt, &comment.Content, &comment.Author, &comment.Version)
		if err != nil {
			return nil, Metadata{}, err
		}
		// add the row to our slice
		comments = append(comments, &comment)
	} // end of for loop
	// after we exit the loop we need to check if it generated any errors
	err = rows.Err()
	if err != nil {
		return nil, Metadata{}, err
	}

	// Create the metadata
	metadata := calculateMetaData(totalRecords, filters.Page, filters.PageSize)

	return comments, metadata, nil

}
