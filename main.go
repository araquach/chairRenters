package main

import (
	"bytes"
	"context"
	"encoding/csv"
	"fmt"
	"github.com/mailgun/mailgun-go/v3"
	"github.com/signintech/pdft"
	gopdf "github.com/signintech/pdft/minigopdf"
	"html/template"
	"io"
	"io/ioutil"
	"log"
	"os/user"
	"strconv"
	"strings"
	"time"
)

type InvoiceFigures struct {
	Date           string
	Stylist        string
	Invoice        string
	DateFrom       string
	DateTo         string
	Weeks          string
	Services       string
	Products       string
	TotalRev       string
	ServiceCharge  string
	WklyCharge     string
	ServVAT        string
	Tips           string
	RetailPurchase string
	RetailProfit   string
	RetailVAT      string
	Charges        string
	ChargesVAT     string
	TotalCharge    string
	ServiceRel     string
	RetailRel      string
	Extra          string
	TotalRel       string
}

func main() {
	var results []InvoiceFigures

	now := time.Now()
	today := now.Format("02-01-2006")

	content, _ := ioutil.ReadFile("figures/" + today + ".csv")

	reader := csv.NewReader(bytes.NewBuffer(content))
	_, err := reader.Read() // skip first line
	if err != nil {
		if err != io.EOF {
			log.Fatalln(err)
		}
	}
	for {
		col, error := reader.Read()
		if error == io.EOF {
			break
		} else if error != nil {
			log.Fatal(error)
		}

		weeks, _ := strconv.Atoi(col[5])
		services, _ := strconv.ParseFloat(col[6], 32)
		products, _ := strconv.ParseFloat(col[7], 8)
		tips, _ := strconv.ParseFloat(col[8], 8)
		extra, _ := strconv.ParseFloat(col[9], 8)

		totalRev := services + products
		serviceCharge := (services - tips) * .45
		wklyCharge := float64(weeks) * 5.00
		servVAT := (serviceCharge + wklyCharge) * .2
		serviceRel := services - serviceCharge - wklyCharge - servVAT

		retailPurchase := products * .5
		retailProfit := (products - retailPurchase) * .4
		retailVAT := retailProfit * .2

		charges := serviceCharge + retailPurchase + retailProfit + wklyCharge
		chargesVAT := servVAT + retailVAT
		totalCharge := charges + chargesVAT
		retailRel := products - retailPurchase - retailProfit - retailVAT
		totalRel := serviceRel + retailRel + tips + extra

		tr := fmt.Sprintf("£%.2f", totalRev)
		sc := fmt.Sprintf("£%.2f", serviceCharge)
		wc := fmt.Sprintf("£%.2f", wklyCharge)
		sv := fmt.Sprintf("£%.2f", servVAT)
		sr := fmt.Sprintf("£%.2f", serviceRel)
		rp := fmt.Sprintf("£%.2f", retailPurchase)
		rpft := fmt.Sprintf("£%.2f", retailProfit)
		rv := fmt.Sprintf("£%.2f", retailVAT)
		c := fmt.Sprintf("£%.2f", wklyCharge)
		cv := fmt.Sprintf("£%.2f", chargesVAT)
		tc := fmt.Sprintf("£%.2f", totalCharge)
		rr := fmt.Sprintf("£%.2f", retailRel)
		total := fmt.Sprintf("£%.2f", totalRel)

		results = append(results, InvoiceFigures{
			Stylist:        col[0],
			Invoice:        col[1],
			Date:           col[2],
			DateFrom:       col[3],
			DateTo:         col[4],
			Weeks:          col[5],
			Services:       col[6],
			Products:       col[7],
			Tips:           col[8],
			Extra:          col[9],
			TotalRev:       tr,
			ServiceCharge:  sc,
			WklyCharge:     wc,
			ServVAT:        sv,
			ServiceRel:     sr,
			RetailPurchase: rp,
			RetailProfit:   rpft,
			RetailVAT:      rv,
			Charges:        c,
			ChargesVAT:     cv,
			TotalCharge:    tc,
			RetailRel:      rr,
			TotalRel:       total,
		})
	}
	for _, v := range results {
		createPDF(v)
		sendInvoice(v)
		fmt.Println(v.Stylist, v.Invoice, v.TotalRel)
	}
}

func createPDF(r InvoiceFigures) {
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
	err = pt.Insert(string(r.Invoice), 1, 78, 198, 100, 100, gopdf.Left|gopdf.Top)
	if err != nil {
		panic(err)
	}

	// Date:
	err = pt.Insert(string(r.Date), 1, 78, 221, 100, 100, gopdf.Left|gopdf.Top)
	if err != nil {
		panic(err)
	}

	// Date Range:
	err = pt.Insert(string(r.DateFrom) + " to " + string(r.DateTo), 1, 465, 221, 100, 100, gopdf.Left|gopdf.Top)
	if err != nil {
		panic(err)
	}

	// Service Revenue:
	err = pt.Insert(string("£" + r.Services), 1, 200, 281.5, 100, 100, gopdf.Center|gopdf.Top)
	if err != nil {
		panic(err)
	}

	// Product Revenue:
	err = pt.Insert("£" + r.Products, 1, 200, 305.5, 100, 100, gopdf.Center|gopdf.Top)
	if err != nil {
		panic(err)
	}

	err = pt.SetFont("helvetica-bold", "", 10)
	if err != nil {
		panic(err)
	}

	// Total Revenue:
	err = pt.Insert(r.TotalRev, 1, 200, 332, 100, 100, gopdf.Center|gopdf.Top)
	if err != nil {
		panic(err)
	}

	err = pt.SetFont("helvetica", "", 10)
	if err != nil {
		panic(err)
	}

	// 45% Service Charge
	err = pt.Insert(r.ServiceCharge, 1, 200, 406.5, 100, 100, gopdf.Center|gopdf.Top)
	if err != nil {
		panic(err)
	}

	// Weekly Charge
	err = pt.Insert(r.WklyCharge, 1, 200, 433, 100, 100, gopdf.Center|gopdf.Top)
	if err != nil {
		panic(err)
	}

	// VAT
	err = pt.Insert(r.ServVAT, 1, 200, 458.5, 100, 100, gopdf.Center|gopdf.Top)
	if err != nil {
		panic(err)
	}

	// Service Credit Release
	err = pt.Insert(r.ServiceRel, 1, 200, 522.5, 100, 100, gopdf.Center|gopdf.Top)
	if err != nil {
		panic(err)
	}

	// Retail Credit Release
	err = pt.Insert(r.RetailRel, 1, 200, 549.5, 100, 100, gopdf.Center|gopdf.Top)
	if err != nil {
		panic(err)
	}

	// Other Service Payments
	err = pt.Insert("£" + r.Extra, 1, 200, 574.5, 100, 100, gopdf.Center|gopdf.Top)
	if err != nil {
		panic(err)
	}

	// Tips
	err = pt.Insert("£" + r.Tips, 1, 200, 600.5, 100, 100, gopdf.Center|gopdf.Top)
	if err != nil {
		panic(err)
	}

	err = pt.SetFont("helvetica-bold", "", 10)
	if err != nil {
		panic(err)
	}

	// Total Credit Released
	err = pt.Insert(r.TotalRel, 1, 200, 627.5, 100, 100, gopdf.Center|gopdf.Top)
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
	err = pt.Insert(r.RetailPurchase, 1, 465, 406.5, 100, 100, gopdf.Center|gopdf.Top)
	if err != nil {
		panic(err)
	}

	// 40% Charge on retail
	err = pt.Insert(r.RetailProfit, 1, 465, 433, 100, 100, gopdf.Center|gopdf.Top)
	if err != nil {
		panic(err)
	}

	// VAT
	err = pt.Insert(r.RetailVAT, 1, 465, 458.5, 100, 100, gopdf.Center|gopdf.Top)
	if err != nil {
		panic(err)
	}

	// Service Credit Release
	err = pt.Insert(r.ServiceRel, 1, 465, 522.5, 100, 100, gopdf.Center|gopdf.Top)
	if err != nil {
		panic(err)
	}

	// Retail Credit Release
	err = pt.Insert(r.RetailRel, 1, 465, 549.5, 100, 100, gopdf.Center|gopdf.Top)
	if err != nil {
		panic(err)
	}

	err = pt.SetFont("helvetica-bold", "", 10)
	if err != nil {
		panic(err)
	}

	// Total Charge
	err = pt.Insert(r.TotalCharge, 1, 465, 574.5, 100, 100, gopdf.Center|gopdf.Top)
	if err != nil {
		panic(err)
	}

	//get home directory
	myself, err := user.Current()
	if err != nil {
		panic(err)
	}
	homedir := myself.HomeDir

	dir1 := homedir + "/Dropbox (Personal)/chair renters/" + r.Stylist + "/Invoices/"
	//dir1 := homedir + "/Dropbox/invoice_test/"
	fn1 := "invoice " + r.Invoice + " - " + dateFormat(r.Date) +  ".pdf"

	//get year for directory
	d := dateFormat(r.Date)
	m := strings.Split(d, "-")[1]
	y := strings.Split(d, "-")[2]

	dir2 := homedir + "/Dropbox (Personal)/Salon Accounts/Invoices/20" + y + "/" + m + y + "/"
	fn2 := r.Stylist + " - inv " + r.Invoice + " - " + dateFormat(r.Date) +  ".pdf"

	//save within the apps output folder
	err = pt.Save("output/" + r.Stylist + "/invoice " + r.Invoice + " - " + dateFormat(r.Date) +  ".pdf")
	if err != nil {
		log.Fatalf("Couldn't save output pdf of %v", r.Stylist)
	}
	time.Sleep(120 * time.Millisecond)
	//save to chair renters dropbox folder
	err = pt.Save(dir1 + fn1)
	if err != nil {
		log.Fatalf("Couldn't save dropbox pdf of %v", r.Stylist)
	}
	time.Sleep(120 * time.Millisecond)
	//save to salon accounts folder
	err = pt.Save(dir2 + fn2)
	if err != nil {
		log.Fatalf("Couldn't save accounts pdf of %v", r.Stylist)
	}
	time.Sleep(120 * time.Millisecond)
}

func dateFormat(d string) (f string) {
	s := strings.Split(d, "/")
	f = s[0] + "-" + s[1] + "-" + s[2]
	return
}

func sendInvoice(r InvoiceFigures) {
	email := map[string]string{
		"Natalie Sharpe": "nsharpe13@yahoo.com",
		"Matthew Lane":   "xmlaneyx@hotmail.co.uk",
		"Michelle Railton": "michellerailton@hotmail.com",
	}

	htmlContent, err := ParseEmailTemplate("email/template.gohtml", r)
	if err != nil {
		log.Fatalln(err)
	}

	textContent, err := ParseEmailTemplate("email/template.txt", r)
	if err != nil {
		log.Fatalln(err)
	}

	mg := mailgun.NewMailgun("jakatasalon.co.uk", "key-7bdc914427016c8714ed8ef2108a5a49")

	sender := "adam@jakatasalon.co.uk"
	subject := "Your Latest Invoice"
	body := textContent
	recipient := email[r.Stylist]

	m := mg.NewMessage(sender, subject, body, recipient)

	m.SetHtml(htmlContent)
	m.AddAttachment("output/" + r.Stylist + "/invoice " + r.Invoice + " - " + dateFormat(r.Date) +  ".pdf")

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	// Send the message	with a 10 second timeout
	resp, id, err := mg.Send(ctx, m)

	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("ID: %s Resp: %s\n", id, resp)
}

func ParseEmailTemplate(templateFileName string, data interface{}) (content string, err error) {
	tmpl, err := template.ParseFiles(templateFileName)
	if err != nil {
		return "", err
	}

	buf := new(bytes.Buffer)

	if err := tmpl.Execute(buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}


