package main

import (
	"archive/zip"
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
)

func main() {
	dataDir := os.Getenv("DATA_DIR")
	outDir := os.Getenv("OUT_DIR")

	// open zip file
	zipFile, err := zip.OpenReader(fmt.Sprintf("%s/ucd.all.grouped.zip", dataDir))
	if err != nil {
		log.Fatal(err)
	}
	defer zipFile.Close()

	// open output file
	outFile, err := os.Create(fmt.Sprintf("%s/data.go", outDir))
	if err != nil {
		log.Fatal(err)
	}
	defer outFile.Close()

	fmt.Fprintf(outFile, "package unicode\n\n\n")
	fmt.Fprintf(outFile, "// this file is auto generated use 'make generate' to regenerate\n\n")
	fmt.Fprintf(outFile, "// Unicode rune enumeration\n")
	fmt.Fprintf(outFile, "const (\n")

	for _, f := range zipFile.File {
		if f.Name != "ucd.all.grouped.xml" {
			continue
		}

		if err := process(f, outFile); err != nil {
			log.Fatal(err)
		}
	}

	fmt.Fprintf(outFile, ")\n\n")
}

type ucdDocument struct {
	XMLName    xml.Name      `xml:"ucd"`
	Repertoire ucdRepertoire `xml:"repertoire"`
}

type ucdRepertoire struct {
	Groups []ucdGroup `xml:"group"`
}

type ucdGroup struct {
	Script     string         `xml:"scx,attr"`
	Block      string         `xml:"blk,attr"`
	Characters []ucdCharacter `xml:"char"`
}

type ucdCharacter struct {
	CodePoint   string         `xml:"cp,attr"`
	Name        string         `xml:"na,attr"`
	NameAliases []ucdNameAlias `xml:"name-alias"`
}

type ucdNameAlias struct {
	Alias string `xml:"alias,attr"`
}

func process(file *zip.File, writer io.Writer) error {
	rc, err := file.Open()
	if err != nil {
		return err
	}

	defer rc.Close()

	var document ucdDocument
	if err := xml.NewDecoder(rc).Decode(&document); err != nil {
		return err
	}

	uniqueEntry := make(map[string]struct{})
	for _, group := range document.Repertoire.Groups {
		if group.Script == "" {
			continue
		}

		if len(group.Characters) == 0 {
			continue
		}

		log.Printf("processing script %s block %s\n", group.Script, group.Block)
		for _, character := range group.Characters {
			codePoint := character.CodePoint
			name := pascalCase(character.Name)

			if name != "" {
				if _, found := uniqueEntry[name]; !found {
					uniqueEntry[name] = struct{}{}
					fmt.Fprintf(writer, "    %s rune = 0x%s\n", name, codePoint)
				}
			}

			for _, aliasInfo := range character.NameAliases {
				alias := pascalCase(aliasInfo.Alias)
				if _, found := uniqueEntry[alias]; !found {
					uniqueEntry[alias] = struct{}{}
					fmt.Fprintf(writer, "    %s rune = 0x%s\n", alias, codePoint)
				}
			}
		}
	}

	return nil
}

func pascalCase(v string) string {
	v = strings.Replace(v, "-", " ", -1)
	parts := strings.Fields(v)
	if len(parts) == 1 {
		return v
	}

	var sb strings.Builder
	for _, part := range parts {
		sb.WriteString(strings.Title(strings.ToLower(part)))
	}

	return sb.String()
}
