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

func getBase ( items []Item ) float64 {
	/// TODO: change 'itmes' for Item[] class
	return 100.00
}

func asItem ( str string ) Item {
	var item Item
	fmt.Println(str)
	str = strings.Replace(str, "@", ":", -1)
	fmt.Println(str)
	tokens := strings.Split(str, ":")
	item.description = tokens[1]
	fmt.Printf("quant:%s, pu:%s\n", tokens[2], tokens[3])

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
	fmt.Printf("item: %s quant: %d p/u: %.2f iva: %.2f\n", item.description, item.quant, item.unitPrice, item.iva)
}

/// \brief	convert a registry entry to a row
func entryAsColumns( entry string ) string {
	// awk input to one row with columns
	// fmt.Println(entry)
	var str string;
	var status string;
	var num string;
	var date string;
	var client string;
	var project string;
	// var items []Item;
	// var itemstr string;
	var iva float64;
	var irpf float64;
	var base float64;
	var ivaamt float64;
	var irpfamt float64;
	var total float64;
	var roi string;
	aw := awk.NewScript()
	aw.Begin = func(s *awk.Script) { aw.SetFS(":") }
	aw.AppendStmt(
		func(aw *awk.Script) bool 	{ return aw.F(1).Match("status")},
		func(aw *awk.Script) 		{ status = aw.F(2).String()})
	aw.AppendStmt(
		func(aw *awk.Script) bool 	{ return aw.F(1).Match("num")},
		func(aw *awk.Script) 		{ num = aw.F(2).String()})
	aw.AppendStmt(
		func(aw *awk.Script) bool 	{ return aw.F(1).Match("date")},
		func(aw *awk.Script) 		{ date = aw.F(2).String()})
	aw.AppendStmt(
		func(aw *awk.Script) bool 	{ return aw.F(1).Match("client")},
		func(aw *awk.Script) 		{ client = aw.F(2).String()})
	aw.AppendStmt(
		func(aw *awk.Script) bool 	{ return aw.F(1).Match("project")},
		func(aw *awk.Script) 		{ project = aw.F(2).String()})
	aw.AppendStmt(
		func(aw *awk.Script) bool 	{ return aw.F(1).Match("iva")},
		func(aw *awk.Script) 		{ iva = aw.F(2).Float64()})
	aw.AppendStmt(
		func(aw *awk.Script) bool 	{ return aw.F(1).Match("irpf")},
		func(aw *awk.Script) 		{ irpf = aw.F(2).Float64()})
	aw.AppendStmt(
		func(aw *awk.Script) bool 	{ return aw.F(1).Match("item")},
		func(aw *awk.Script) 		{
			item := asItem(aw.F(0).String())
			// asItem(aw.F(0).String())
			printItem(item)
			// itemstr += fmt.Sprintf("%s\n", aw.F(0).String())
		})

	// printSlice(items)

	// base = getBase( "item1" )
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

	aw.End = func(aw *awk.Script) {
		registryFormatString := "%-8.8s|%-9.9s|%-11.11s|%-14.14s|%-14.14s|%8.2f|%6.2f|%6.2f|%6.2f|%6.2f|%8.2f|%-3.3s\n"
		str = fmt.Sprintf(registryFormatString, status, num, date, client, project, iva, irpf, base, ivaamt, irpfamt, total, roi)
		// str = itemstr
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
	fmt.Printf("print registry: %s\n", "")
	str := ""
	aw := awk.NewScript()
	aw.Begin = func(s *awk.Script) {
		aw.SetRS("\n\n")
		aw.SetFS("\n")
		// aw.SetOFS(" ")
		registryFormatString := "%-8.8s|%-9.9s|%-11.11s|%-14.14s|%-14.14s|%8s|%6s|%6s|%6s|%6s|%8s|%-3.3s\n"
		str = fmt.Sprintf(registryFormatString, "status", "num", "date", "client", "project", "iva", "irpf", "base", "ivaamt", "irpfamt", "total", "roi")
	}
	aw.AppendStmt(nil, func(aw *awk.Script) {
		// status := strings.Split(aw.F(2).String(), ":")[1]
		// num := strings.Split(aw.F(3).String(), ":")[1]
		// date := strings.Split(aw.F(4).String(), ":")[1]
		// client := strings.Split(aw.F(6).String(), ":")[1]
		// project := strings.Split(aw.F(5).String(), ":")[1]
		// // TODO: change 'base' for cacluation on items quant * unit price
		// base := strings.Split(aw.F(12).String(), ":")[1]
		// iva := strings.Split(aw.F(14).String(), ":")[1]
		// irpf := strings.Split(aw.F(16).String(), ":")[1]

		// str += fmt.Sprintf("%-8s|%-8s|%-8s|%-8s|%-8s|\n", status, num, date, client, project)
		// str += fmt.Sprintf("%-8s|%-8s|%-11s|%-20s|%-20s|%.2f|%.2f|%.2f\n", status, num, date, client, project, base, iva, irpf)

		str += entryAsColumns(aw.F(0).String())
	})
	aw.Run(script.Exec("cat " + billPath + "/" + registryFilename))

	return str
}

func filter( regex string ) {
	fmt.Printf("filter: %s\n", regex)
	// re := regexp.MustCompile(regex)
	// registry().MatchRegexp(re).Stdout()
	// registry().Stdout()
	aw := awk.NewScript()
	aw.Begin = func(s *awk.Script) {
		aw.SetRS("\n\n")
		aw.SetFS("\n")
	}
	aw.AppendStmt(nil, func(s *awk.Script) {
		fmt.Printf("%s\n", aw.F(0))
	})

	// convert registry string to Reader in order to use
	// it as awk input
	registry := strings.NewReader(registry())
	aw.Run(registry)
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
	fmt.Printf("filename: %s\n", registryFilename)
	fmt.Printf("clients: %s\n", clientsFilename)
	// t := time.Now()
	// fmt.Printf("date: %4d%02d%02d\n", t.Year(), t.Month(), t.Day())

	switch cmd {
	case "client": client( flag.Arg(1) )
	case "job": job( flag.Arg(1), flag.Arg(2) )
	case "invoice": invoice( flag.Arg(1) )
	case "pdf": exportToPdf( flag.Arg(1) )
	case "status": status( flag.Arg(1) )
	case "filter": filter( flag.Arg(1) )
	case "registry": fmt.Println(registry())
	case "tax": tax( flag.Arg(1) )
	}

}
