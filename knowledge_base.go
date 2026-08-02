package glytos

import "context"

// KnowledgeBaseService manages knowledge-base documents and hybrid retrieval search.
type KnowledgeBaseService struct{ client *Client }

// DocumentCreateParams are the fields for KnowledgeBaseService.CreateDocument.
type DocumentCreateParams struct {
	// Name is the document name (required).
	Name string
	// Content is the document text (required); it is chunked and embedded.
	Content string
	// ChunkSize overrides the default chunk size.
	ChunkSize *int
	// ChunkOverlap overrides the default chunk overlap.
	ChunkOverlap *int
}

// SearchParams are the fields for KnowledgeBaseService.Search.
type SearchParams struct {
	// Query is the search text (required).
	Query string
	// TopK caps the number of results.
	TopK *int
	// DocumentIDs scopes the search to specific documents.
	DocumentIDs []int
	// MinScore sets a similarity floor for results.
	MinScore *float64
}

// ListDocuments returns your knowledge-base documents.
func (s *KnowledgeBaseService) ListDocuments(ctx context.Context) ([]Document, error) {
	var out []Document
	err := s.client.do(ctx, "GET", "/knowledge-base/documents", nil, nil, &out)
	return out, err
}

// CreateDocument adds a document (it is chunked and embedded for retrieval).
func (s *KnowledgeBaseService) CreateDocument(ctx context.Context, params DocumentCreateParams) (*Document, error) {
	body := map[string]any{"name": params.Name, "content": params.Content}
	if params.ChunkSize != nil {
		body["chunk_size"] = *params.ChunkSize
	}
	if params.ChunkOverlap != nil {
		body["chunk_overlap"] = *params.ChunkOverlap
	}
	var out Document
	err := s.client.do(ctx, "POST", "/knowledge-base/documents", body, nil, &out)
	return &out, err
}

// Search runs hybrid (vector + full-text) search over your documents.
func (s *KnowledgeBaseService) Search(ctx context.Context, params SearchParams) ([]map[string]any, error) {
	body := map[string]any{"query": params.Query}
	if params.TopK != nil {
		body["top_k"] = *params.TopK
	}
	if params.DocumentIDs != nil {
		body["document_ids"] = params.DocumentIDs
	}
	if params.MinScore != nil {
		body["min_score"] = *params.MinScore
	}
	var out []map[string]any
	err := s.client.do(ctx, "POST", "/knowledge-base/search", body, nil, &out)
	return out, err
}

// UploadDocument uploads a document file (txt, md, pdf) instead of pasting its text.
func (s *KnowledgeBaseService) UploadDocument(ctx context.Context, filename string, content []byte) (*Document, error) {
	var out Document
	err := s.client.UploadFile(ctx, "/knowledge-base/documents/upload", nil, filename, content, &out)
	return &out, err
}
