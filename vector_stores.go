package glytos

import "context"

// VectorStoresService manages vector stores over knowledge-base documents.
type VectorStoresService struct{ client *Client }

// List returns your vector stores.
func (s *VectorStoresService) List(ctx context.Context) ([]VectorStore, error) {
	var out []VectorStore
	err := s.client.do(ctx, "GET", "/vector-stores", nil, nil, &out)
	return out, err
}

// Create creates a vector store.
func (s *VectorStoresService) Create(ctx context.Context, name string) (*VectorStore, error) {
	var out VectorStore
	err := s.client.do(ctx, "POST", "/vector-stores", map[string]any{"name": name}, nil, &out)
	return &out, err
}

// Retrieve returns a vector store by uuid.
func (s *VectorStoresService) Retrieve(ctx context.Context, vectorStoreUUID string) (*VectorStore, error) {
	var out VectorStore
	err := s.client.do(ctx, "GET", "/vector-stores/"+esc(vectorStoreUUID), nil, nil, &out)
	return &out, err
}

// Delete deletes a vector store.
func (s *VectorStoresService) Delete(ctx context.Context, vectorStoreUUID string) error {
	return s.client.do(ctx, "DELETE", "/vector-stores/"+esc(vectorStoreUUID), nil, nil, nil)
}

// UploadDocument adds a document file to a vector store, so an agent can search it.
func (s *VectorStoresService) UploadDocument(ctx context.Context, vectorStoreUUID, filename string, content []byte) (*Document, error) {
	var out Document
	err := s.client.UploadFile(ctx, "/vector-stores/"+esc(vectorStoreUUID)+"/documents/upload", nil, filename, content, &out)
	return &out, err
}
