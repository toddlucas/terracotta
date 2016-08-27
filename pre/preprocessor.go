package pre

import (
	"io/ioutil"
	"log"
	"os"
	"path"
	"sort"
	"strings"
)

const terraformDirectory = ".terraform"
const tfdefsFilename = "terraform.tfdefs"
const terraformExtension = ".tf"
const templateExtension = ".tft"

// Preprocessor encapsulates file parsing and code generation.
type Preprocessor struct {
	parser Parser
}

// ProcessDirectory enumerates and parses relevant files in the source
// directory and generates corresponding files in the output directory.
// After the current directory is complete, subdirectories are processed.
func (p *Preprocessor) ProcessDirectory(source string, output string, defines []string, undefs []string) {
	if _, err := os.Stat(source); os.IsNotExist(err) {
		log.Fatalf("Directory '%s' does not exist", source)
	}

	if _, err := os.Stat(output); os.IsNotExist(err) {
		log.Fatalf("Directory '%s' does not exist", output)
	}

	dirs, files, tfdefs, err := p.getDirectoryContents(source)
	if err != nil {
		log.Fatal(err)
	}

	p.parser.Enter()

	// If there's a file called 'terraform.tfdefs', load it.
	if tfdefs {
		err = p.processDefines(path.Join(source, tfdefsFilename))
		if err != nil {
			log.Fatal(err)
		}
	}

	// Apply any command-line overrides after the file-based defs.
	err = p.applyDefines(defines, undefs)
	if err != nil {
		log.Fatal(err)
	}

	for _, templateFilename := range files {
		baseName := removeFileExtension(templateFilename)
		generatedFilename := baseName + terraformExtension

		err := p.processFile(path.Join(source, templateFilename), path.Join(output, generatedFilename))
		if err != nil {
			log.Fatal(templateFilename + err.Error())
			return
		}
	}

	// Preprocess subdirectories.
	for _, dir := range dirs {
		p.ProcessDirectory(path.Join(source, dir), path.Join(output, dir), nil, nil)
	}

	p.parser.Leave()
}

func (p *Preprocessor) getDirectoryContents(path string) ([]string, []string, bool, error) {
	list, err := ioutil.ReadDir(path)
	if err != nil {
		return nil, nil, false, err
	}

	var hasDefines = false
	var files []string
	var dirs []string
	for _, file := range list {
		if file.IsDir() {
			// Read all directories except '.terraform'
			if !strings.EqualFold(file.Name(), terraformDirectory) {
				dirs = append(dirs, file.Name())
			}
		} else if strings.HasSuffix(file.Name(), templateExtension) {
			files = append(files, file.Name())
		} else if strings.EqualFold(file.Name(), tfdefsFilename) {
			hasDefines = true
		}
	}

	sort.Strings(dirs)
	sort.Strings(files)

	return dirs, files, hasDefines, nil
}

func (p *Preprocessor) processFile(input string, output string) error {
	f, err := os.Create(output)
	if err != nil {
		return err
	}

	defer f.Close()

	p.parser.SetFile(input)
	p.parser.Enter()

	err = p.parser.Parse(func(line string) {
		f.WriteString(line + eol)
	})
	if err != nil {
		return err
	}

	p.parser.Leave()
	return nil
}

func (p *Preprocessor) processDefines(filename string) error {
	p.parser.SetFile(filename)

	return p.parser.ParseDefines()
}

func (p *Preprocessor) applyDefines(defines []string, undefs []string) error {
	if defines != nil {
		for _, define := range defines {
			err := p.parser.Define(define)
			if err != nil {
				return err
			}
		}
	}

	if undefs != nil {
		for _, undef := range undefs {
			err := p.parser.Undef(undef)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
