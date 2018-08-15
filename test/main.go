package main

import (
	"fmt"
	"log"
	"regexp"
	"strings"

	"../parser" //TODO
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	sen1 := `LOOK (indexname1'typename, indexname2'typename2, indexname3'typename3):
                 CONDITION  [ indexName1.field1 GT 100, indexName1.field1 NEQ "a32bd", indexName2.field3 NEQ 123.123 ,indexName2.field2 LT 100, index2.field2 EQ 123, index3.field4 SF "ab2c32", index3.field2 GTE 1000, index4.file4 LTE 120]
                 AT [ 2018.14.23:12.23.45 - 2018.12.13:12.12.12]`
	//sen2 := `RECENT (indexname1'typename, indexname2'typename2) :
	//TOTAL [100]
	//AT [2018.14.23:12.23.45 - 2018.12.13.12.12.12]`

	r1 := strings.NewReader(sen1)

	p1 := parser.NewParser(r1)
	resu1, err := p1.Parse()
	fmt.Printf("%+v\n", resu1.IndexToTypeSet)
	for _, v := range resu1.IndexToFieldSet {
		for k, va := range v {
			fmt.Println(k, va)
		}
	}
	fmt.Println(resu1.TimeBegin, resu1.TimeEnd)
	log.Println("[ERROR] => ", err)

}

func TestSymbol() {
	reg := regexp.MustCompile(`[\:\.\'\(\)\{\}]`)
	fmt.Println("regexp: ", reg.MatchString("'"))
}
