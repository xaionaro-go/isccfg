package main

import (
	"bufio"
	"fmt"
	"github.com/xaionaro-go/isccfg"
	"os"
)

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}
func main() {
	if len(os.Args) < 2 {
		panic(fmt.Errorf("Path to the config file is not defined. Syntax: %v path_to_the_config", os.Args[0]))
	}

	filePath := os.Args[1]

	file, err := os.Open(filePath)
	checkErr(err)

	cfgReader := bufio.NewReader(file)

	cfg, err := isccfg.Parse(cfgReader)
	checkErr(err)

	cfg.WriteJsonTo(os.Stdout)
	//cfg.WriteTo(os.Stdout)

	return
}
