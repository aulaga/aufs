package main

import (
	"fmt"
	"github.com/aulaga/cloud/src/domain/storage"
)

func main() {
	st := storage.NewFs("C:\\Users\\raul\\GolandProjects\\cloud\\.attic\\fs")
	fi, err := st.Open("foo.txt")
	if err != nil {
		panic(err.Error())
	}

	b := make([]byte, 10)
	n, err := fi.Read(b)
	if err != nil {
		panic(err.Error())
	}

	fmt.Println("Read bytes", n)
	fmt.Println(b)
}
