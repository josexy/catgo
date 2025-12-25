package util

import (
	"fmt"

	"github.com/fatih/color"
)

type ColorPrinter struct {
	Green     *color.Color
	Red       *color.Color
	Yellow    *color.Color
	Cyan      *color.Color
	Bold      *color.Color
	BoldGreen *color.Color
}

var Printer = &ColorPrinter{
	Green:     color.New(color.FgGreen),
	Red:       color.New(color.FgRed),
	Yellow:    color.New(color.FgYellow),
	Cyan:      color.New(color.FgCyan),
	Bold:      color.New(color.Bold),
	BoldGreen: color.New(color.FgGreen, color.Bold),
}

func (p *ColorPrinter) PrintCompiling(target string) {
	p.BoldGreen.Print("   Compiling")
	fmt.Printf(" %s\n", target)
}

func (p *ColorPrinter) PrintFinished(profile string, duration string) {
	p.BoldGreen.Print("    Finished")
	fmt.Printf(" `%s` target(s) in %s\n", profile, duration)
}

func (p *ColorPrinter) PrintRunning(target string) {
	p.BoldGreen.Print("     Running")
	fmt.Printf(" `%s`\n", target)
}

func (p *ColorPrinter) PrintCreated(item string) {
	p.BoldGreen.Print("     Created")
	fmt.Printf(" %s\n", item)
}

func (p *ColorPrinter) PrintAdding(pkg string) {
	p.BoldGreen.Print("      Adding")
	fmt.Printf(" %s\n", pkg)
}

func (p *ColorPrinter) PrintRemoving(pkg string) {
	p.BoldGreen.Print("    Removing")
	fmt.Printf(" %s\n", pkg)
}

func (p *ColorPrinter) PrintRemoved(items []string) {
	p.BoldGreen.Print("    Removed")
	fmt.Printf(" %d files\n", len(items))
	for _, file := range items {
		fmt.Printf("       - %s\n", file)
	}
}

func (p *ColorPrinter) PrintUpdating(item string) {
	p.BoldGreen.Print("    Updating")
	fmt.Printf(" %s\n", item)
}

func (p *ColorPrinter) PrintVendoring(item string) {
	p.BoldGreen.Print("    Vendoring")
	fmt.Printf(" %s\n", item)
}

func (p *ColorPrinter) PrintSuccess(msg string) {
	p.Green.Print("success")
	fmt.Printf(": %s\n", msg)
}

func (p *ColorPrinter) PrintError(msg string) {
	p.Red.Print("error")
	fmt.Printf(": %s\n", msg)
}

func (p *ColorPrinter) PrintWarning(msg string) {
	p.Yellow.Print("warning")
	fmt.Printf(": %s\n", msg)
}
