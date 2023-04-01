package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/charmbracelet/log"
	"github.com/mattn/go-isatty"
	openai "github.com/sashabaranov/go-openai"
)

func printUsage() {
	fmt.Printf("Usage: %s [OPTIONS] [PREFIX TERM]\n", os.Args[0])
	flag.PrintDefaults()
}

func readStdinContent() string {
	if !isatty.IsTerminal(os.Stdin.Fd()) {
		reader := bufio.NewReader(os.Stdin)
		stdinBytes, err := ioutil.ReadAll(reader)
		if err != nil {
			log.Fatal("Error reading standard input: ", err)
		}
		return string(stdinBytes)
	}
	return ""
}

func writeOutput(output, fileName string) {
	file, err := os.Create(fileName)
	if err != nil {
		log.Fatal("Error creating output file: ", err)
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	_, err = writer.WriteString(output)
	if err != nil {
		log.Fatal("Error writing to output file: ", err)
	}
	writer.Flush()
}

func createClient(apiKey string) *openai.Client {
	if apiKey == "" {
		log.Fatal("Error: OPENAI_API_KEY environment variable is required")
	}
	return openai.NewClient(apiKey)
}

func main() {
	modelVersionFlag := flag.String("m", "gpt-4", "OpenAI model flag.")
	formatFlag := flag.Bool("f", false, "Ask GPT to format the output as Markdown.")
	outputFileFlag := flag.String("o", "", "Output file to save response. If not specified, prints to console.")
	flag.Usage = printUsage
	flag.Parse()

	client := createClient(os.Getenv("OPENAI_API_KEY"))
	content := readStdinContent()
	prefix := strings.Join(flag.Args(), " ")
	if prefix == "" && content == "" {
		printUsage()
		os.Exit(0)
	}
	if *formatFlag {
		prefix = fmt.Sprintf("%s Format output as Markdown.", prefix)
	}

	if prefix != "" {
		content = strings.TrimSpace(prefix + "\n\n" + content)
	}

	resp, err := client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: *modelVersionFlag,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleUser,
					Content: content,
				},
			},
		},
	)

	if err != nil {
		log.Fatal("ChatCompletion error: ", err)
	}

	gptContent := resp.Choices[0].Message.Content
	if *outputFileFlag != "" {
		writeOutput(gptContent, *outputFileFlag)
	} else {
		fmt.Println(gptContent)
	}
}
