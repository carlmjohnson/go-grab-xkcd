package grabxkcd

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

// CLI runs the go-grab-xkcd command line app and returns its exit status.
func CLI(args []string) int {
	var app appEnv
	err := app.fromArgs(args)
	if err != nil {
		return 2
	}
	if err = app.run(); err != nil {
		fmt.Fprintf(os.Stderr, "Runtime error: %v\n", err)
		return 1
	}
	return 0
}

type appEnv struct {
	hc         http.Client
	comicNo    int
	saveImage  bool
	outputJSON bool
}

func (app *appEnv) fromArgs(args []string) error {
	// Shallow copy of default client
	app.hc = *http.DefaultClient
	fl := flag.NewFlagSet("xkcd-grab", flag.ContinueOnError)
	fl.IntVar(
		&app.comicNo, "n", LatestComic, "Comic number to fetch (default latest)",
	)
	fl.DurationVar(&app.hc.Timeout, "t", 30*time.Second, "Client timeout")
	fl.BoolVar(
		&app.saveImage, "s", false, "Save image to current directory",
	)
	outputType := fl.String(
		"o", "text", "Print output in format: text/json",
	)
	if err := fl.Parse(args); err != nil {
		return err
	}
	if *outputType != "text" && *outputType != "json" {
		fmt.Fprintf(os.Stderr, "got bad output type: %q\n", *outputType)
		fl.Usage()
		return flag.ErrHelp
	}
	app.outputJSON = *outputType == "json"
	return nil
}

func (app *appEnv) run() error {
	u := BuildURL(app.comicNo)

	var resp APIResponse
	if err := app.fetchJSON(u, &resp); err != nil {
		return err
	}
	if resp.Date() == nil {
		return fmt.Errorf("could not parse date of comic: %q/%q/%q",
			resp.Year, resp.Month, resp.Day)
	}
	if app.saveImage {
		if err := app.fetchAndSave(resp.Image, resp.Filename()); err != nil {
			return err
		}
		fmt.Fprintf(os.Stderr, "Saved: %q\n", resp.Filename())
	}
	if app.outputJSON {
		return printJSON(resp)
	}

	return prettyPrint(resp)
}

func (app *appEnv) fetchJSON(url string, data interface{}) error {
	resp, err := app.hc.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return json.NewDecoder(resp.Body).Decode(data)
}

func (app *appEnv) fetchAndSave(url, destPath string) error {
	resp, err := app.hc.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	f, err := os.Create(destPath)
	if err != nil {
		return err
	}

	_, err = io.Copy(f, resp.Body)
	return err
}

// Output is the JSON output of this app
type Output struct {
	Title       string `json:"title"`
	Number      int    `json:"number"`
	Date        string `json:"date"`
	Description string `json:"description"`
	Image       string `json:"image"`
}

func printJSON(ar APIResponse) error {
	o := Output{
		Title:       ar.Title,
		Number:      ar.Number,
		Date:        ar.Date().Format("2006-01-02"),
		Description: ar.Description,
		Image:       ar.Image,
	}
	b, err := json.MarshalIndent(&o, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(b))
	return nil
}

func prettyPrint(ar APIResponse) error {
	_, err := fmt.Printf(
		"Title: %s\nComic No: %d\nDate: %s\nDescription: %s\nImage: %s\n",
		ar.Title,
		ar.Number,
		ar.Date().Format("02-01-2006"),
		ar.Description,
		ar.Image,
	)
	return err
}
