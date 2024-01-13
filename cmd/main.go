package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"

	mpw "github.com/emiljoha/mpw-go/internal"
	"golang.org/x/term"
)

func main() {
	flags := parseFlags()
	config, err := readConfig()
	if err != nil {
		fmt.Printf("error reading config: %s\n", err.Error())
		os.Exit(1)
	}
	if flags.FullName == "" {
		if config.FullName != "" {
			flags.FullName = config.FullName
		} else {
			fullName, err := input("Full Name: ")
			if err != nil {
				fmt.Println(err.Error())
				os.Exit(1)
			}
			flags.FullName = fullName
		}
	}
	if flags.SiteName == "" {
		siteName, err := input("Site Name: ")
		if err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		}
		flags.SiteName = siteName
	}
	_, ok := mpw.TemplateDictionary[flags.SiteResultType]
	if !ok {
		typeAbbreviations := map[mpw.ResultType]mpw.ResultType{
			"x": "Maximum",
			"l": "Long",
			"m": "Medium",
			"b": "Basic",
			"s": "Short",
			"i": "PIN",
			"n": "Name",
			"p": "Phrase",
		}
		fullSiteResult, ok := typeAbbreviations[flags.SiteResultType]
		if !ok {
			fmt.Printf("Site result type not valid: %s\n", flags.SiteResultType)
			os.Exit(1)
		}
		flags.SiteResultType = mpw.ResultType(fullSiteResult)
	}
	fmt.Print("Password: ")
	pass, err := term.ReadPassword(int(os.Stdin.Fd()))
	fmt.Print("\n")
	if err != nil {
		fmt.Printf("password input error: %s\n", err.Error())
		os.Exit(1)		
	}
	sitePassword, err :=mpw.Password(flags.FullName, string(pass), flags.SiteName, flags.Counter, flags.SiteResultType)
	if err != nil {
		fmt.Printf("password generation error: %s\n", err.Error())
		os.Exit(1)
	}
	fmt.Println(sitePassword)
}

func input(prompt string) (string, error) {
	fmt.Print(prompt)
	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("An error occured while reading input. Please try again: %w", err)
	}
	return strings.TrimSuffix(input, "\n"), nil
}

type Config struct {
	FullName string `json:"FULL_NAME"`
}

func readConfig() (Config, error) {
	b, err := os.ReadFile(os.Getenv("HOME") + "/.config/mpw/config.json")
	if err != nil {
		return Config{}, nil
	}
	var c Config
	err = json.Unmarshal(b, &c)
	if err != nil {
		return Config{}, err
	}
	return c, nil
}

type Flags struct {
	FullName string
	Counter int
	SiteResultType mpw.ResultType
	Verbose bool
	Quiet bool
	SiteName string
}
func parseFlags() Flags {
	fullName := flag.String("full-name", "","Specify the full name of the user")
	fullNameShorthand := flag.String("u", "","Specify the full name of the user")
	counter := flag.Int("counter", 1,"Specify the full name of the user")
	counterShorthand := flag.Int("c", 1,"Specify the full name of the user")
	helpSiteResultType := "Specify the password's template\n"+
         "Defaults to 'long' (-t a)\n"+
         "x, Maximum  | 20 characters, contains symbols.\n"+
         "l, Long     | Copy-friendly, 14 characters, symbols.\n"+
         "m, Medium   | Copy-friendly, 8 characters, symbols.\n"+
         "b, Basic    | 8 characters, no symbols.\n"+
         "s, Short    | Copy-friendly, 4 characters, no symbols.\n"+
         "i, Pin      | 4 numbers.\n"+
         "n, Name     | 9 letter name.\n"+
         "p, Phrase   | 20 character sentence."
	siteResultType := flag.String("site-result-type", "Long", helpSiteResultType)
	siteResultTypeShortHand := flag.String("t", "Long", helpSiteResultType)
	verbose := flag.Bool("verbose", false, "Increase output verbosity")
	verboseShortHand := flag.Bool("v", false, "Increase output verbosity")
	quiet := flag.Bool("quiet", false, "Decrease output verbosity")
	quietShortHand := flag.Bool("q", false, "Decrease output verbosity")

	flag.Parse()

	if *fullNameShorthand != "" {
		fullName = fullNameShorthand
	}
	if *counterShorthand != 1 {
		counter = counterShorthand
	}
	if *siteResultTypeShortHand != "Long" {
		siteResultType = siteResultTypeShortHand
	}
	if *verboseShortHand {
		verbose = verboseShortHand
	}
	if *quietShortHand {
		quiet = quietShortHand
	}
	args := flag.Args()
	if len(args) > 1 {
		_, _ = fmt.Printf("only one non-flagged comman line argument allowed: %s", args)
		os.Exit(1)
	}
	siteName := ""
	if len(args) != 0 {
		siteName = args[0]
	}
	_ = fullName
	_ = counter
	_ = siteResultType
	_ = verbose
	_ = quiet
	_ = siteName
	return Flags{
		FullName: *fullName,
		Counter: *counter,
		SiteResultType: mpw.ResultType(*siteResultType),
		Verbose: *verbose,
		Quiet: *quiet,
		SiteName: siteName,		
	}
}
