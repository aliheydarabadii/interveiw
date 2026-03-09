package ports

import "context"

type HTTPServer interface {
	Start(ctx context.Context) error
}
