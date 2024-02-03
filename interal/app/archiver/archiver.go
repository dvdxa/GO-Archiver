package archiver

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"

	"github.com/cheggaaa/pb"
	"github.com/klauspost/compress/zip"
)

func Compress() {
	var (
		inputPath      string
		outputDir      string
		outputFileName string
	)

	flag.StringVar(&inputPath, "input", "", "Input path to archive")
	flag.StringVar(&outputDir, "output", "", "Output directory for the archive")
	flag.StringVar(&outputFileName, "name", "archive.zip", "Output file name")
	flag.Parse()

	if inputPath == "" || outputDir == "" || outputFileName == "" {
		fmt.Println("Usage: go run main.go -input <input_path> -output <output_directory> -name <output_file_name>")
		return
	}

	err := os.MkdirAll(outputDir, os.ModePerm)
	if err != nil {
		fmt.Println("Error creating output directory:", err)
		return
	}

	outputArchivePath := filepath.Join(outputDir, outputFileName)
	archiveFile, err := os.Create(outputArchivePath)
	if err != nil {
		fmt.Println("Error creating output archive file:", err)
		return
	}
	defer archiveFile.Close()

	zipWriter := zip.NewWriter(archiveFile)
	defer zipWriter.Close()

	// Count total number of files for progress bar
	var totalFiles int
	filepath.Walk(inputPath, func(filePath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		totalFiles++
		return nil
	})

	// Create progress bar
	bar := pb.New(totalFiles)
	bar.SetMaxWidth(80)
	bar.Start()

	err = filepath.Walk(inputPath, func(filePath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}

		header.Name, err = filepath.Rel(inputPath, filePath)
		if err != nil {
			return err
		}

		if info.IsDir() {
			header.Name += string(filepath.Separator)
		}

		header.Method = zip.Deflate

		writer, err := zipWriter.CreateHeader(header)
		if err != nil {
			return err
		}

		if !info.IsDir() {
			file, err := os.Open(filePath)
			if err != nil {
				return err
			}
			defer file.Close()

			_, err = writer.Write([]byte("This is a placeholder content. Replace with actual file content."))
			if err != nil {
				return err
			}
		}

		// Update progress bar
		bar.Increment()

		return nil
	})

	bar.Finish()

	if err != nil {
		fmt.Println("Error walking through input path:", err)
		return
	}

	fmt.Println("Archive created successfully:", outputArchivePath)
}

func ShowProgress(fileToArchive *os.File, wg *sync.WaitGroup) {

	fileInfo, err := fileToArchive.Stat()
	if err != nil {
		fmt.Println("Error getting file info:", err)
		return
	}

	progressCh := make(chan int)

	go func() {
		defer wg.Done()
		for progress := range progressCh {
			fmt.Printf("\rProgress: %d%%", progress)
		}
		fmt.Println("\nArchiving complete!")
	}()

	buffer := make([]byte, 1024)

	var totalRead int64 
	for {
		n, err := fileToArchive.Read(buffer)
		if n > 0 {
			totalRead += int64(n)
			progress := int(float64(totalRead) / float64(fileInfo.Size()) * 100)
			progressCh <- progress
		}

		if err == io.EOF {
			break
		}

		if err != nil {
			fmt.Println("Error reading file:", err)
			return
		}
	}

	close(progressCh)
}
