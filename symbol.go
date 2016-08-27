package main

import "fmt"

type symbols []string

//
// flag.Value interface
//

func (i *symbols) String() string {
	return fmt.Sprintf("%s", *i)
}

func (i *symbols) Set(value string) error {
	fmt.Printf("%s\n", value)
	*i = append(*i, value)
	return nil
}
