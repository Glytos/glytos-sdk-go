package glytos

import "net/url"

// esc percent-encodes a value for safe interpolation into a single URL path
// segment. Like the other Glytos SDKs it encodes "/" (and other reserved
// characters) so a value containing "/", "?", "#" or ".." cannot traverse paths
// or inject query/fragment components.
func esc(s string) string { return url.PathEscape(s) }
