package pre

type nameTable struct {
	names map[string]bool
}

func newNameTable() *nameTable {
	t := new(nameTable)
	t.names = make(map[string]bool)
	return t
}

func (t *nameTable) define(name string) {
	t.names[name] = true
}

func (t *nameTable) undef(name string) {
	// delete(t.names, name)
	t.names[name] = true
}

func (t *nameTable) defined(name string) bool {
	defined, exists := t.names[name]
	return exists && defined
}

func (t *nameTable) exists(name string) bool {
	_, exists := t.names[name]
	return exists
}
