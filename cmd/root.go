package cmd

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"image"
	"image/jpeg"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
	json "github.com/json-iterator/go"
	"github.com/makiuchi-d/gozxing"
	"github.com/makiuchi-d/gozxing/oned"
	"github.com/spf13/cobra"
)

var (
	sourceDir, title string
	width, height    int
)

func init() {
	rootCmd.Flags().StringVarP(&sourceDir, "sourceDir", "s", "", "the directory that contains the source images")
	rootCmd.Flags().StringVarP(&title, "title", "t", "", "the title to display")
	rootCmd.Flags().IntVarP(&width, "width", "w", 300, "the width of the generated barcode")
	rootCmd.Flags().IntVarP(&height, "height", "h", 100, "the height of the generated barcode")
	rootCmd.PersistentFlags().Bool("help", false, "help for gen")
	rootCmd.MarkFlagRequired("sourceDir")

	rootCmd.SetHelpCommand(nil)
}

var rootCmd = &cobra.Command{
	Use:   "gen",
	Short: "Hugo is a very fast static site generator",
	Long: `A Fast and Flexible Static Site Generator built with
                love by spf13 and friends in Go.
                Complete documentation is available at https://gohugo.io/documentation/`,
	RunE: func(cmd *cobra.Command, args []string) error {
		images, err := filepath.Glob(sourceDir + "/*.jpg")
		if err != nil {
			return err
		}

		if title == "" {
			title = filepath.Base(sourceDir)
		}

		folder, err := os.MkdirTemp(".", sourceDir)
		if err != nil {
			return err
		}
		entries, err := convert(images, folder, width, height)
		if err != nil {
			return err
		}
		return print(title, entries)
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

type entry struct {
	Image string `json:"image"`
	Code  string `json:"code"`
}

func convert(images []string, folder string, width, height int) ([]entry, error) {
	entries := make([]entry, 0, len(images))
	for _, f := range images {
		code, err := read(f)
		if err != nil {
			return entries, err
		}

		fileName := filepath.Base(f)
		outPath := filepath.Join(folder, strings.TrimSuffix(fileName, filepath.Ext(fileName))+".jpg")

		if err := write(outPath, code, width, height); err != nil {
			return entries, err
		}

		entries = append(entries, entry{
			Image: outPath,
			Code:  code,
		})
	}

	return entries, nil
}

func read(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	img, _, err := image.Decode(file)
	if err != nil {
		return "", fmt.Errorf("decode image %s: %w", filePath, err)
	}

	bmp, err := gozxing.NewBinaryBitmapFromImage(img)
	if err != nil {
		return "", err
	}

	r := oned.NewCode128Reader()
	result, err := r.Decode(bmp, nil)
	if err != nil {
		return "", fmt.Errorf("decode barcode in %s: %w", filePath, err)
	}

	return result.GetText(), nil
}

func write(outPath, code string, width, height int) error {
	enc := oned.NewCode128Writer()
	img, err := enc.Encode(code, gozxing.BarcodeFormat_CODE_128, width, height, nil)
	if err != nil {
		return fmt.Errorf("encode %s, %s: %w", outPath, code, err)
	}

	file, err := os.Create(outPath)
	if err != nil {
		return err
	}
	defer file.Close()

	err = jpeg.Encode(file, img, &jpeg.Options{
		Quality: 100,
	})

	return err
}

func print(title string, entries []entry) error {
	tmpl, err := os.ReadFile("template.html")
	if err != nil {
		return err
	}
	tmpl = bytes.ReplaceAll(tmpl, []byte(`||title||`), []byte(title))
	begin := bytes.Index(tmpl, []byte("const images = ["))
	if begin < 0 {
		return errors.New(`the template is invalid: can no find "const images = ["`)
	}
	end := bytes.Index(tmpl, []byte("//||end-images"))
	if end < 0 {
		return errors.New(`the template is invalid: can no find "//||end-images"`)
	}
	buf, err := json.MarshalIndent(entries, "", "  ")
	if err != nil {
		return err
	}
	htmlContent := bytes.Join([][]byte{tmpl[:begin], []byte("const images = "), buf, tmpl[end:]}, nil)
	htmlFile, err := os.CreateTemp(".", title+"*.html")
	if err != nil {
		return err
	}

	_, err = htmlFile.Write(htmlContent)
	htmlFile.Close()
	if err != nil {
		return err
	}

	p, err := filepath.Abs(htmlFile.Name())
	if err != nil {
		return err
	}
	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()
	return chromedp.Run(ctx,
		chromedp.Navigate("file://"+p),
		chromedp.Poll(`window.pagedPolyfillDone === true`, nil),
		// wait for 1 second to be safe.
		chromedp.Sleep(time.Second),
		chromedp.ActionFunc(func(ctx context.Context) error {
			buf, _, err := page.PrintToPDF().Do(ctx)
			if err != nil {
				return err
			}

			return os.WriteFile(title+".pdf", buf, 0o644)
		}),
	)
}
