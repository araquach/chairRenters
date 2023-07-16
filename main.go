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
	"sync"
	"time"
)

type InvoiceFigures struct {
	Date           string
	Stylist        string
	Invoice        string
	DateFrom       string
	DateTo         string
	Weeks          string
	Turnover       string
	RetailRevenue  string
	ServicePercent string
	ServiceCharge  string
	WklyCharge     string
	ServiceVAT     string
	Tips           string
	Charges        string
	ChargesVAT     string
	TotalCharge    string
	ServiceRel     string
	Commission     string
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
		products, _ := strconv.ParseFloat(col[7], 32)
		tips, _ := strconv.ParseFloat(col[8], 8)
		extra, _ := strconv.ParseFloat(col[9], 8)
		servicePercent, _ := strconv.ParseFloat(col[10], 8)

		serviceCharge := services * servicePercent
		wklyCharge := float64(weeks) * 5.00
		servVAT := (serviceCharge + wklyCharge) * .2
		serviceRel := services - serviceCharge - wklyCharge - servVAT

		charges := serviceCharge + wklyCharge
		chargesVAT := servVAT
		totalCharge := charges + chargesVAT
		retailRel := ((products / 2) / 100) * 45
		totalRel := serviceRel + retailRel + tips + extra

		sp := fmt.Sprintf("%v%%", servicePercent*100)
		sc := fmt.Sprintf("£%.2f", serviceCharge)
		wc := fmt.Sprintf("£%.2f", wklyCharge)
		sv := fmt.Sprintf("£%.2f", servVAT)
		sr := fmt.Sprintf("£%.2f", serviceRel)
		c := fmt.Sprintf("£%.2f", charges)
		cv := fmt.Sprintf("£%.2f", chargesVAT)
		tc := fmt.Sprintf("£%.2f", totalCharge)
		rr := fmt.Sprintf("£%.2f", retailRel)
		total := fmt.Sprintf("£%.2f", totalRel)

		results = append(results, InvoiceFigures{
			Date:           col[2],
			Stylist:        col[0],
			Invoice:        col[1],
			DateFrom:       col[3],
			DateTo:         col[4],
			Weeks:          col[5],
			Turnover:       col[6],
			RetailRevenue:  col[7],
			ServicePercent: sp,
			ServiceCharge:  sc,
			WklyCharge:     wc,
			ServiceVAT:     sv,
			Tips:           col[8],
			Charges:        c,
			ChargesVAT:     cv,
			TotalCharge:    tc,
			ServiceRel:     sr,
			Commission:     rr,
			Extra:          col[9],
			TotalRel:       total,
		})
	}

	done := make(chan bool, len(results))

	// Process each invoice concurrently using goroutines
	for _, v := range results {
		go processInvoice(v, done)
		sendInvoice(v)
	}
	// Wait for all goroutines to finish
	for range results {
		<-done
	}
}

func processInvoice(v InvoiceFigures, done chan<- bool) {
	createPDF(v)
	time.Sleep(2 * time.Second)
	fmt.Println(v.Stylist, v.Invoice, v.TotalRel)
	// Signal that this goroutine is done
	done <- true
}

func createPDF(r InvoiceFigures) {
	var pt pdft.PDFt
	err := pt.Open("template/" + r.Stylist + ".pdf")
	if err != nil {
		panic("Couldn't open pdf.")
	}

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done() // Decrease the WaitGroup counter when the operation is done
		// Insert your text here, e.g.:
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

		// Date Range:
		err = pt.Insert(string(r.DateFrom)+" to "+string(r.DateTo), 1, 445, 221, 100, 100, gopdf.Left|gopdf.Top)
		if err != nil {
			panic(err)
		}

		// Turnover:
		err = pt.Insert("£"+r.Turnover, 1, 210, 302, 100, 100, gopdf.Left|gopdf.Top)
		if err != nil {
			panic(err)
		}

		// Retail Revenue:
		err = pt.Insert("£"+r.RetailRevenue, 1, 475, 302, 100, 100, gopdf.Left|gopdf.Top)
		if err != nil {
			panic(err)
		}

		// Retail Commission
		err = pt.Insert(r.Commission, 1, 475, 325, 100, 100, gopdf.Left|gopdf.Top)
		if err != nil {
			panic(err)
		}

		// Service Percent
		err = pt.Insert(r.ServicePercent, 1, 236, 412, 100, 100, gopdf.Left|gopdf.Top)
		if err != nil {
			panic(err)
		}

		// Service Charge
		err = pt.Insert(r.ServiceCharge, 1, 285, 412.5, 100, 100, gopdf.Left|gopdf.Top)
		if err != nil {
			panic(err)
		}

		// Weekly Charge
		err = pt.Insert(r.WklyCharge, 1, 285, 437, 100, 100, gopdf.Left|gopdf.Top)
		if err != nil {
			panic(err)
		}

		// Fees ex VAT
		err = pt.Insert(r.Charges, 1, 508, 435, 100, 100, gopdf.Left|gopdf.Top)
		if err != nil {
			panic(err)
		}

		// VAT
		err = pt.Insert(r.ChargesVAT, 1, 507, 458.5, 100, 100, gopdf.Left|gopdf.Top)
		if err != nil {
			panic(err)
		}

		// Total Fees inc VAT
		err = pt.Insert(r.TotalCharge, 1, 508, 483, 100, 100, gopdf.Left|gopdf.Top)
		if err != nil {
			panic(err)
		}

		// Service Credit Release
		err = pt.Insert(r.ServiceRel, 1, 290, 623, 100, 100, gopdf.Left|gopdf.Top)
		if err != nil {
			panic(err)
		}

		// Product Commission
		err = pt.Insert(r.Commission, 1, 290, 647, 100, 100, gopdf.Left|gopdf.Top)
		if err != nil {
			panic(err)
		}

		// Tips
		err = pt.Insert("£"+r.Tips, 1, 290, 670, 100, 100, gopdf.Left|gopdf.Top)
		if err != nil {
			panic(err)
		}

		// Additional Works Carried out
		err = pt.Insert("£"+r.Extra, 1, 290, 695, 100, 100, gopdf.Left|gopdf.Top)
		if err != nil {
			panic(err)
		}

		// Total Credit Released
		err = pt.Insert(r.TotalRel, 1, 290, 742, 100, 100, gopdf.Left|gopdf.Top)
		if err != nil {
			panic(err)
		}

		// Payment Date
		err = pt.Insert(r.Date, 1, 290, 765, 100, 100, gopdf.Left|gopdf.Top)
		if err != nil {
			panic(err)
		}

		// get home directory
		myself, err := user.Current()
		if err != nil {
			panic(err)
		}

		homedir := myself.HomeDir

		d := dateFormat(r.Date)
		m := strings.Split(d, "-")[1]
		y := strings.Split(d, "-")[2]

		dir1 := homedir + "/Jakata Salon Dropbox/Adam Carter/Salon Stuff/chair renters/" + r.Stylist + "/Invoices/"
		fn1 := "invoice " + r.Invoice + " - " + dateFormat(r.Date) + ".pdf"

		dir2 := homedir + "/Jakata Salon Dropbox/Adam Carter/Salon Stuff/Salon Accounts 2/Invoices//20" + y + "/" + m + y + "/"
		fn2 := r.Stylist + " - inv " + r.Invoice + " - " + dateFormat(r.Date) + ".pdf"

		time.Sleep(1000 * time.Millisecond)
		//save within the apps output folder
		err = pt.Save("output/" + r.Stylist + "/invoice " + r.Invoice + " - " + dateFormat(r.Date) + ".pdf")
		if err != nil {
			log.Fatalf("Couldn't save output pdf of %v", r.Stylist)
		}
		time.Sleep(1000 * time.Millisecond)
		// save to chair renters dropbox folder
		err = pt.Save(dir1 + fn1)
		if err != nil {
			log.Fatalf("Couldn't save dropbox pdf of %v", r.Stylist)
		}
		time.Sleep(1000 * time.Millisecond)
		// save to salon accounts folder
		err = pt.Save(dir2 + fn2)
		if err != nil {
			log.Fatalf("Couldn't save accounts pdf of %v", r.Stylist)
		}
	}()
}

func dateFormat(d string) (f string) {
	s := strings.Split(d, "/")
	f = s[0] + "-" + s[1] + "-" + s[2]
	return
}

func sendInvoice(r InvoiceFigures) {
	email := map[string]string{
		"Natalie Sharpe":   "nsharpe13@yahoo.com",
		"Matthew Lane":     "xmlaneyx@hotmail.co.uk",
		"Michelle Railton": "michellerailton@hotmail.com",
		"Georgia Lutton":   "gl.hairgal@gmail.com",
		"Joanne Birchall":  "joannemahoney84@gmail.com",
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
	m.AddAttachment("output/" + r.Stylist + "/invoice " + r.Invoice + " - " + dateFormat(r.Date) + ".pdf")

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
