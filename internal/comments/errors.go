package comments

import "errors"

var ErrInvalidComment = errors.New("comment must contain 1 to 5000 characters")
