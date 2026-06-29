package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/cloudwego/eino/compose"
	"github.com/joho/godotenv"
)

// defaultQuestion là câu hỏi mẫu khi chạy CLI không truyền tham số.
const defaultQuestion = "Giải thích cho mình về ý nghĩa của Áo tấc."

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}

	ctx := context.Background()

	chatModel, err := buildChatModel(ctx)
	if err != nil {
		log.Fatal(err)
	}
	agent, err := buildAgent(ctx, chatModel)
	if err != nil {
		log.Fatal(err)
	}
	runnable, err := buildChatChain(ctx, agent, chatModel)
	if err != nil {
		log.Fatal(err)
	}

	// Chế độ chạy:
	//   go run .               → CLI với câu hỏi mẫu
	//   go run . "<câu hỏi>"   → CLI với câu hỏi truyền vào
	//   go run . eval          → chạy 5 test case, ghi kết quả ra evaluation.md
	if len(os.Args) >= 2 && os.Args[1] == "eval" {
		if err := runEval(ctx, runnable); err != nil {
			log.Fatal(err)
		}
		return
	}

	question := defaultQuestion
	if len(os.Args) >= 2 {
		question = strings.Join(os.Args[1:], " ")
	}
	runCLI(ctx, runnable, question)
}

// runCLI chạy chuỗi chat cho một câu hỏi và in câu trả lời + gợi ý ra terminal.
func runCLI(ctx context.Context, runnable compose.Runnable[string, *ChatResponse], question string) {
	fmt.Printf("🟣 Câu hỏi: %s\n", question)
	fmt.Println("⏳ Đang xử lý (ReAct trả lời → sinh gợi ý)...")
	fmt.Println(strings.Repeat("─", 60))

	resp, err := runnable.Invoke(ctx, question)
	if err != nil {
		log.Fatalf("lỗi: %v", err)
	}

	fmt.Println(resp.Answer)
	fmt.Println(strings.Repeat("─", 60))
	if len(resp.Suggestions) == 0 {
		fmt.Println("💡 (không có gợi ý)")
		return
	}
	fmt.Println("💡 Gợi ý câu hỏi tiếp:")
	for i, s := range resp.Suggestions {
		fmt.Printf("  %d. %s\n", i+1, s)
	}
}
