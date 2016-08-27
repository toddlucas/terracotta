package pre

import "path"

func removeFileExtension(filename string) string {
	var extension = path.Ext(filename)
	var name = filename[0 : len(filename)-len(extension)]
	return name
}
