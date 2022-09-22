package main

import (
	"os"
	"fmt"
	"strings"
	"strconv"
	// "regexp"
	"flag"
	// "time"
	"github.com/bitfield/script"
	"github.com/spakin/awk"
)

var billPath string;
var registryFilename string;
var clientsFilename string;
var registryFormatString string = "%-8.8s|%-9.9s|%-11.11s|%-14.14s|%-14.14s|%5s|%5s|%8s|%7s|%7s|%8s|%-3.3s\n"
var registryHeaderString string = fmt.Sprintf(registryFormatString, "status", "num", "date", "client", "project", "iva", "irpf", "base", "ivaamt", "irpfamt", "total", "roi")

type Item struct {
	description string
	quant float32
	unitPrice float32
	iva float32
	irpf float32
}

type Entry struct {
	status string
	num string
	date string
	project string // YYYYMM-name
	client string
	items []Item
	roi string
}

func client( client string ) {
	fmt.Printf("client: %s\n", client)
}

func job( client string, project string ) {
	fmt.Printf("client: %s\n", client)
	fmt.Printf("project: %s\n", project)
}

func invoice( projectDirName string ) {
	fmt.Printf("invoice project: %s\n", projectDirName)
}

func exportToPdf( invoiceNum string ) {
	fmt.Printf("export invoice: %s\n", invoiceNum)
}

func status( status string ) {
	fmt.Printf("status: %s\n", status)
}

func getItemSubtotal( item Item ) float32 {
	return item.quant * item.unitPrice
}

func getBase ( items []Item ) float32 {
	var base float32 = 0.0;
	for _, item := range items {
		base += getItemSubtotal(item)
	}
	return base
}

func asItem ( str string ) Item {
	var item Item
	str = strings.Replace(str, "@", ":", -1)
	tokens := strings.Split(str, ":")
	item.description = tokens[1]

	quant := strings.TrimSpace(tokens[2])
	if q, err := strconv.ParseFloat(quant, 32); err == nil {
		item.quant = float32(q)
		// fmt.Printf("quant: %s (%T) %.2f (%T)\n", quant, quant, q, q)
	}

	unitprice := strings.TrimSpace(tokens[3])
	if p, err := strconv.ParseFloat(unitprice, 32); err == nil {
		item.unitPrice = float32(p)
		// fmt.Printf("p/u: %s (%T) %.2f (%T)\n", unitprice, unitprice, p, p)
	}

	return item
}

func printSlice( x []Item ) {
	fmt.Printf("len=%d cap=%d slice=%v\n", len(x), cap(x), x)
}

func printItem( item Item ) {
	fmt.Printf("item: %-20s quant: %6.2f p/u: %6.2f iva: %6.2f\n", item.description, item.quant, item.unitPrice, item.iva)
}

func getColIndex( str string, strArray []string ) int {
	for i, s := range strArray {
		str = strings.TrimSpace(str)
		s = strings.TrimSpace(s)
		// fmt.Printf("%d: %-8s %-8s %b\n", i, str, s, (str == s))
		if str == s {
			return i
		}
	}

	return -1
}


/// \brief	convert a registry entry to a row
func entryAsColumns( entry string ) string {
	// awk input to one row with columns
	// fmt.Println(entry)
	var str string
	var status string
	var num string
	var date string
	var client string
	var project string
	var items []Item
	// var itemstr string
	var iva float64
	var irpf float64
	var base float64
	var ivaamt float64
	var irpfamt float64
	var total float64
	var roi string
	aw := awk.NewScript()
	aw.Begin = func(s *awk.Script) { aw.SetFS(":") }
	aw.AppendStmt(
		func(aw *awk.Script) bool 	{ return aw.F(1).Match("status") },
		func(aw *awk.Script) 		{ status = aw.F(2).String() })
	aw.AppendStmt(
		func(aw *awk.Script) bool 	{ return aw.F(1).Match("num") },
		func(aw *awk.Script) 		{ num = aw.F(2).String() })
	aw.AppendStmt(
		func(aw *awk.Script) bool 	{ return aw.F(1).Match("date") },
		func(aw *awk.Script) 		{ date = aw.F(2).String() })
	aw.AppendStmt(
		func(aw *awk.Script) bool 	{ return aw.F(1).Match("client") },
		func(aw *awk.Script) 		{ client = aw.F(2).String() })
	aw.AppendStmt(
		func(aw *awk.Script) bool 	{ return aw.F(1).Match("project") },
		func(aw *awk.Script) 		{ project = aw.F(2).String() })
	aw.AppendStmt(
		func(aw *awk.Script) bool 	{ return aw.F(1).Match("iva") },
		func(aw *awk.Script) 		{ iva = aw.F(2).Float64() })
	aw.AppendStmt(
		func(aw *awk.Script) bool 	{ return aw.F(1).Match("irpf") },
		func(aw *awk.Script) 		{ irpf = aw.F(2).Float64() })
	aw.AppendStmt(
		func(aw *awk.Script) bool 	{ return aw.F(1).Match("item") },
		func(aw *awk.Script) 		{
			item := asItem(aw.F(0).String())
			items = append(items, item)
			base = float64(getBase(items))
		})

	// fmt.Println(entry)


	aw.End = func(aw *awk.Script) {
		ivaamt = base * iva / 100
		irpfamt = base * irpf / 100
		total = base + ivaamt - irpfamt
		if( iva == 0.0 && irpf > 0.0 ) {
			roi = "exempt"
		} else if ( iva == 0.0 && irpf == 0.0 ) {
			roi = "roi"
		} else {
			roi = "-"
		}
		registryFormatString := "%-8.8s|%-9.9s|%-11.11s|%-14.14s|%-14.14s|%5.2f|%5.2f|%8.2f|%7.2f|%7.2f|%8.2f|%-3.3s\n"
		str = fmt.Sprintf(registryFormatString, status, num, date, client, project, iva, irpf, base, ivaamt, irpfamt, total, roi)
	}

	areader := strings.NewReader(entry)
	aw.Run(areader)
	return str
}

/// \brief	return the contents of registry file
// func registry() *script.Pipe {
// 	fmt.Printf("print registry: %s\n", "")
// 	return script.Exec( "cat " + billPath + "/" + registryFilename )
// }


func registry() string {
	// fmt.Printf("print registry: %s\n", "")
	str := ""
	aw := awk.NewScript()
	aw.Begin = func(s *awk.Script) {
		aw.SetRS("\n\n")
		aw.SetFS("\n")
		// aw.SetOFS(" ")
		// registryFormatString := "%-8.8s|%-9.9s|%-11.11s|%-14.14s|%-14.14s|%5s|%5s|%8s|%7s|%7s|%8s|%-3.3s\n"
		// str = fmt.Sprintf(registryFormatString, "status", "num", "date", "client", "project", "iva", "irpf", "base", "ivaamt", "irpfamt", "total", "roi")
	}
	aw.AppendStmt(nil, func(aw *awk.Script) {
		str += entryAsColumns(aw.F(0).String())
	})
	aw.Run(script.Exec("cat " + billPath + "/" + registryFilename))

	return str
}

func filter( regex string ) string {
	cols := strings.Split(registryHeaderString, "|")

	keyValue := strings.Split(regex, "=")

	fmt.Println("")
	if len(keyValue) != 2 {
		for i, c := range cols {
			cols[i] = strings.TrimSpace(c)
		}
		fmt.Printf("invalid filter: %s\n", regex)
		fmt.Println("usage: bill filter <COLUMN>=<REGEX>[:<COLUMN>=<REGEX>:...]")
		fmt.Printf("COLUMN: %s\n", strings.Join(cols, " | "))
		// os.Exit(0)

		fmt.Println("")
		return "\n-- BAD RESULT --\n"
	}

	key := keyValue[0]
	value := keyValue[1]
	col := getColIndex(key, cols) + 1

	fmt.Printf("filter regex: %s\n", regex)
	fmt.Printf("filter columns: %d\n", len(cols))
	fmt.Printf("key: %s : %d\n", key, col)
	fmt.Printf("value: %s\n", value)
	fmt.Println("")

	str := ""
	aw := awk.NewScript()
	aw.Begin = func(s *awk.Script) {
		// aw.SetRS("\n\n")
		// aw.SetFS("\n")
		aw.SetFS("|")
	}
	// aw.AppendStmt(nil, func(s *awk.Script) {
	// 	fmt.Printf("col: %d\n", col)
	// 	fmt.Printf("regex: %s\n", regex)
	// 	fmt.Printf("key: %s\n", key)
	// 	fmt.Printf("awk: %s\n", aw.F(col))
	// 	// fmt.Printf("match: %s\n", aw.F(col + 2).match(regex))
	// 	fmt.Println()
	// })
	aw.AppendStmt(
		func(aw *awk.Script) bool 	{ return aw.F(col).Match(value) },
		func(aw *awk.Script) 		{
			// fmt.Printf("col: %d - key: %s - val: %s\n", col, key, value)
			str += fmt.Sprintln(aw.F(0).String())
		})

	// convert registry string to Reader in order to use
	// it as awk input
	registry := strings.NewReader(registry())
	aw.Run(registry)

	return str
}

func tax( trimester string ) {
	fmt.Printf("filter: %s\n", trimester)
}

func main() {
	// script.Args().Join().Stdout()
	// script.Stdin().Stdout()
	// fmt.Println("alo")
	// script.FindFiles("../../../*.*").Stdout()

	// bla := "3"
	// f, err := strconv.ParseFloat(bla, 32)
	// fmt.Printf("%s (%T) %.2f (%T)", bla, bla, f, f)
	// os.Exit(0)

	homedir, err := os.UserHomeDir()
	if err != nil { return }

	flag.StringVar( &billPath, "p", homedir + "/.local/share/bill", "base path for all data" )
	flag.StringVar( &registryFilename, "f", "billregistry.dat", "file name for the registry" )
	flag.StringVar( &clientsFilename, "c", "billclients.dat", "file name for the clients data" )
	flag.Parse()

	cmd := flag.Arg(0)
	fmt.Printf("cmd: %s\n", cmd)
	fmt.Printf("path: %s\n", billPath)
	fmt.Printf("registry: %s\n", registryFilename)
	fmt.Printf("clients: %s\n", clientsFilename)
	// t := time.Now()
	// fmt.Printf("date: %4d%02d%02d\n", t.Year(), t.Month(), t.Day())

	switch cmd {
	case "client": client( flag.Arg(1) )
	case "job": job( flag.Arg(1), flag.Arg(2) )
	case "invoice": invoice( flag.Arg(1) )
	case "pdf": exportToPdf( flag.Arg(1) )
	case "status": status( flag.Arg(1) )
	case "filter": fmt.Printf("%s%s", registryHeaderString, filter( flag.Arg(1)) )
	case "registry": fmt.Printf("%s%s", registryHeaderString, registry())
	case "tax": tax( flag.Arg(1) )
	}

}
