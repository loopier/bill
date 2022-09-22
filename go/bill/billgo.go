package main

import (
	"os"
	"fmt"
	"strings"
	"strconv"
	// "regexp"
	"flag"
	"time"
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


/// \brief	convert a registry entry to a '|'-separated row
func entryAsColumns( entry string ) string {
	var str string
	var status string
	var num string
	var date string
	var client string
	var project string
	var items []Item
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

///	\brief	return the full registry as a '|'-separated table
func registry() string {
	str := ""
	aw := awk.NewScript()
	aw.Begin = func(s *awk.Script) {
		aw.SetRS("\n\n")
		aw.SetFS("\n")
	}
	aw.AppendStmt(nil, func(aw *awk.Script) {
		str += entryAsColumns(aw.F(0).String())
	})
	aw.Run(script.Exec("cat " + billPath + "/" + registryFilename))

	return str
}

/// \param	keyOrValue	Int		0 = key; 1 = value
func getRegexKeysOrValues( regex string, keyOrValue int ) []string {
	var result []string
	res := strings.Split(regex, ":")
	for _, re := range res {
		result = append(result, strings.Split(re, "=")[keyOrValue])
	}
	return result
}

/// \brief	filter input table
///
///	Filter the input table with the given criteria.
///	 Criteria can be searched in one or many columns.
///
/// Examples:
///	 - Show all the entries in the first trimester of 2022:
///	 `bill filter date=0[123]/2022`
///	 - To look for all the entries in the first trimester of 2022 from one client:
///	 `bill filter date=0[123]/2022:client=foolanito`
///
/// \param	regex	String	Criteria to match: <COLUMN>=<REGEX>[:<COLUMN>=<CRITERIA>]
/// \param	input	String	A table to be filtered
/// \returns		String	All the input rows containing the given criteria and
/// 						the sum of their tax amounts and grand totals.
func filter( regex string, input string ) string {
	output := ""
	keys := getRegexKeysOrValues(regex, 0);
	values := getRegexKeysOrValues(regex, 1);
	cols := strings.Split(registryHeaderString, "|")

	totalivaamt := 0.0
	totalirpfamt := 0.0
	totalbase := 0.0
	totaltotal := 0.0
	aw := awk.NewScript()
	aw.Begin = func(s *awk.Script) {
		aw.SetFS("|")
	}
	// aw.AppendStmt(nil, func(s *awk.Script) {
	// 	fmt.Printf("%d: %s: %s : %s : %s\n", col, key, value, aw.F(col).String(), aw.F(col).Match(value))
	// })
	aw.AppendStmt(
		func(aw *awk.Script) bool {
			for i, k := range keys {
				c := getColIndex(k, cols) + 1
				match := aw.F(c).Match(values[i])
				if match == false {
					return false
				}
			}

			return true
		},
		func(aw *awk.Script) {
			// fmt.Printf("col: %d - key: %s - val: %s\n", col, key, value)
			totalivaamt += aw.F(getColIndex("ivaamt", cols) + 1).Float64()
			totalirpfamt += aw.F(getColIndex("irpfamt", cols) + 1).Float64()
			totalbase += aw.F(getColIndex("base", cols) + 1).Float64()
			totaltotal += aw.F(getColIndex("total", cols) + 1).Float64()
			output += fmt.Sprintln(aw.F(0).String())
		})

	aw.End = func(s *awk.Script) {
		output += fmt.Sprintf("total iva: %.2f\n", totalivaamt)
		output += fmt.Sprintf("total irpf: %.2f\n", totalirpfamt)
		output += fmt.Sprintf("total base: %.2f\n", totalbase)
		output += fmt.Sprintf("absolute total: %.2f\n", totaltotal)
	}

	// convert result string to Reader in order to use
	// it as awk input
	registry := strings.NewReader(input)
	aw.Run(registry)

	return output
}

/// \breif	filter by trimester
func tax( trimester string, year string ) string {
	var output string
	fmt.Printf("filter by trimester: %s %s\n", trimester, year)
	tri := strings.TrimSpace(trimester)
	if t, err := strconv.ParseInt(tri, 10, 0); err == nil {
		fmt.Printf("trimester: %d\n", t)
		var dec int
		if dec = 0; (t % 4 == 0) {
			dec = 1
		}

		re := fmt.Sprintf("date=%d[%d%d%d]/%s", dec, (((t-1)*3) + 1) % 10, (((t-1)*3) + 2) % 10, (((t-1)*3) + 3) % 10, year)

		fmt.Printf("trimester: %d : regex: %s\n", t, re)
		output = filter( re, registry() )
	}
	return output
}

func main() {
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
	case "filter": fmt.Printf("%s%s", registryHeaderString, filter( flag.Arg(1), registry()) )
	case "registry": fmt.Printf("%s%s", registryHeaderString, registry())
	case "tax":
		year := flag.Arg(2)
		if len(year) < 2 {
			year = fmt.Sprintf("%d", time.Now().Year())
		}
		fmt.Printf("%s%s", registryHeaderString, tax( flag.Arg(1), year ))
	}

}
