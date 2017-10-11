package schema

// Tags defined the tags for struct fields and union branches.
type Tags map[string]string

// Deprecated returns true, if the deprecated tag is set. Otherwise
// false will be returned.
func (t Tags) Deprecated() bool {
	_, has := t["deprecated"]
	return has
}
