package main

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"github.com/signintech/pdft"
	gopdf "github.com/signintech/pdft/minigopdf"
	"io"
	"log"
	"os"
)

type InvoiceFigures struct {
	Date     string
	Stylist  string
	Invoice  string
	DateFrom string
	DateTo   string
	Services string
	Products string
	Tips     string
	Released string
}

func main() {
	var results []InvoiceFigures

	file := "figures/22-05-21.csv"

	csvFile, _ := os.Open(file)
	reader := csv.NewReader(bufio.NewReader(csvFile))
	for {
		col, error := reader.Read()
		if error == io.EOF {
			break
		} else if error != nil {
			log.Fatal(error)
		}
		results = append(results, InvoiceFigures{
			Stylist:  col[0],
			Invoice:  col[1],
			Date:     col[2],
			DateFrom: col[3],
			DateTo:   col[4],
			Services: col[5],
			Products: col[6],
			Tips:     col[7],
			Released: "2,000,000",
		})
	}
	for _, v := range results {
		createPDF(&v)
		fmt.Println(v.Stylist, v.Date)
	}
}

func createPDF(r *InvoiceFigures) {
	var pt pdft.PDFt
	err := pt.Open("template/" + r.Stylist + ".pdf")
	if err != nil {
		panic("Couldn't open pdf.")
	}

	err = pt.AddFont("helvetica", "fonts/Helvetica.ttf")
	if err != nil {
		log.Fatal(err)
		return
	}

	err = pt.AddFont("helvetica-bold", "fonts/Helvetica-Bold.ttf")
	if err != nil {
		log.Fatal(err)
		return
	}

	err = pt.SetFont("helvetica", "", 10)
	if err != nil {
		panic(err)
	}

	//insert text to pdf
	// Invoice:
	err = pt.Insert(r.Invoice, 1, 78, 198, 100, 100, gopdf.Left|gopdf.Top)
	if err != nil {
		panic(err)
	}

	// Date:
	err = pt.Insert(r.Date, 1, 78, 221, 100, 100, gopdf.Left|gopdf.Top)
	if err != nil {
		panic(err)
	}

	// Service Revenue:
	err = pt.Insert(r.Services, 1, 200, 281.5, 100, 100, gopdf.Center|gopdf.Top)
	if err != nil {
		panic(err)
	}

	// Product Revenue:
	err = pt.Insert(r.Products, 1, 200, 305.5, 100, 100, gopdf.Center|gopdf.Top)
	if err != nil {
		panic(err)
	}

	err = pt.SetFont("helvetica-bold", "", 10)
	if err != nil {
		panic(err)
	}

	// Total Revenue:
	err = pt.Insert("Total", 1, 200, 332, 100, 100, gopdf.Center|gopdf.Top)
	if err != nil {
		panic(err)
	}

	err = pt.SetFont("helvetica", "", 10)
	if err != nil {
		panic(err)
	}

	// 45% Service Charge
	err = pt.Insert("£1000", 1, 200, 406.5, 100, 100, gopdf.Center|gopdf.Top)
	if err != nil {
		panic(err)
	}

	// Weekly Charge
	err = pt.Insert("£10.00", 1, 200, 433, 100, 100, gopdf.Center|gopdf.Top)
	if err != nil {
		panic(err)
	}

	// VAT
	err = pt.Insert("£100.00", 1, 200, 458.5, 100, 100, gopdf.Center|gopdf.Top)
	if err != nil {
		panic(err)
	}

	// Service Credit Release
	err = pt.Insert("£2000.00", 1, 200, 522.5, 100, 100, gopdf.Center|gopdf.Top)
	if err != nil {
		panic(err)
	}

	// Retail Credit Release
	err = pt.Insert("£200.00", 1, 200, 549.5, 100, 100, gopdf.Center|gopdf.Top)
	if err != nil {
		panic(err)
	}

	// Other Service Payments
	err = pt.Insert("£0.00", 1, 200, 574.5, 100, 100, gopdf.Center|gopdf.Top)
	if err != nil {
		panic(err)
	}

	// Tips
	err = pt.Insert(r.Tips, 1, 200, 600.5, 100, 100, gopdf.Center|gopdf.Top)
	if err != nil {
		panic(err)
	}

	err = pt.SetFont("helvetica-bold", "", 10)
	if err != nil {
		panic(err)
	}

	// Total Credit Released
	err = pt.Insert("£2220.00", 1, 200, 627.5, 100, 100, gopdf.Center|gopdf.Top)
	if err != nil {
		panic(err)
	}

	// Total Remaining
	err = pt.Insert("£0.00", 1, 200, 653, 100, 100, gopdf.Center|gopdf.Top)
	if err != nil {
		panic(err)
	}

	err = pt.SetFont("helvetica", "", 10)
	if err != nil {
		panic(err)
	}

	//
	err = pt.Insert(r.Date, 1, 200, 717.5, 100, 100, gopdf.Center|gopdf.Top)
	if err != nil {
		panic(err)
	}

	err = pt.SetFont("helvetica", "", 10)
	if err != nil {
		panic(err)
	}

	// 50% Retail Charge on retail
	err = pt.Insert("£1000", 1, 465, 406.5, 100, 100, gopdf.Center|gopdf.Top)
	if err != nil {
		panic(err)
	}

	// 40% Charge on retail
	err = pt.Insert("£100.00", 1, 465, 433, 100, 100, gopdf.Center|gopdf.Top)
	if err != nil {
		panic(err)
	}

	// VAT
	err = pt.Insert("£100.00", 1, 465, 458.5, 100, 100, gopdf.Center|gopdf.Top)
	if err != nil {
		panic(err)
	}

	// Service Credit Release
	err = pt.Insert("£2000.00", 1, 465, 522.5, 100, 100, gopdf.Center|gopdf.Top)
	if err != nil {
		panic(err)
	}

	// Retail Credit Release
	err = pt.Insert("£400.00", 1, 465, 549.5, 100, 100, gopdf.Center|gopdf.Top)
	if err != nil {
		panic(err)
	}

	err = pt.SetFont("helvetica-bold", "", 10)
	if err != nil {
		panic(err)
	}

	// Total Charge
	err = pt.Insert("£1000.00", 1, 465, 574.5, 100, 100, gopdf.Center|gopdf.Top)
	if err != nil {
		panic(err)
	}

	err = pt.Save("output/" + r.Stylist + ".pdf")
	if err != nil {
		panic("Couldn't save pdf.")
	}
}
