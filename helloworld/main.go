package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/schema"

	"github.com/cloudwego/eino-ext/components/model/openai"

	"github.com/joho/godotenv"
)

func main() {
	demo := flag.String("demo", "", "run a scripted demo scenario non-interactively (scenario name or \"all\"); omitted = interactive chat")
	flag.Parse()
	// load .env file
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	ctx := context.Background()

	// Khởi tạo model
	model, err := openai.NewChatModel(ctx, &openai.ChatModelConfig{
		APIKey:  os.Getenv("OPENAI_API_KEY"),
		Model:   os.Getenv("OPENAI_MODEL"),
		BaseURL: os.Getenv("OPENAI_BASE_URL"),
		ByAzure: func() bool {
			return os.Getenv("OPENAI_BY_AZURE") == "true"
		}(),
	})
	if err != nil {
		log.Fatal(err)
	}

	// Khởi tạo tools (logic deterministic nằm ở đây, không nằm ở model).
	searchCRM, err := NewSearchCRMDatabaseTool()
	if err != nil {
		log.Fatal(err)
	}
	checkReq, err := NewCheckLeadRequirementsTool()
	if err != nil {
		log.Fatal(err)
	}
	recPlan, err := NewRecommendCRMPlanTool()
	if err != nil {
		log.Fatal(err)
	}

	// Tạo ReAct Agent
	agent, err := NewCRMAgent(ctx, model, searchCRM, checkReq, recPlan)
	if err != nil {
		log.Fatal(err)
	}

	// Runner: EnableStreaming=false → one complete assistant message per turn.
	runner := adk.NewRunner(ctx, adk.RunnerConfig{
		Agent:           agent,
		EnableStreaming: false,
	})

	if *demo != "" {
		runDemo(ctx, runner, *demo)
		return
	}

	// Multi-turn loop. App owns history (User + Assistant text only), re-sent each turn.
	history := []adk.Message{}
	reader := bufio.NewReader(os.Stdin)

	fmt.Println("Agent: Dạ em là tư vấn viên CRM. Anh/chị cần hỗ trợ gì ạ? (gõ 'exit' để thoát)")

	for {
		fmt.Print("\nKhách: ")

		line, err := reader.ReadString('\n')
		text := strings.TrimSpace(line)

		if text == "exit" {
			fmt.Println()
			break
		}

		// Dòng rỗng: bỏ qua để Enter không tốn một lượt gọi model.
		if text != "" {
			history = append(history, schema.UserMessage(text))

			reply := runTurn(ctx, runner, history)
			fmt.Printf("Agent: %s\n", reply)

			history = append(history, schema.AssistantMessage(reply, nil))
		}

		// EOF (pipe đóng): xử lý nốt dòng còn lại rồi thoát.
		if err != nil {
			break
		}
	}
}

// runTurn drains one turn's event stream and returns the last non-empty assistant text.
func runTurn(ctx context.Context, runner *adk.Runner, history []adk.Message) string {
	var reply string
	events := runner.Run(ctx, history)
	for {
		ev, ok := events.Next()
		if !ok {
			break
		}
		if ev.Err != nil {
			log.Printf("loi: %v", ev.Err)
			break
		}
		if ev.Output == nil || ev.Output.MessageOutput == nil {
			continue
		}
		msg, gErr := ev.Output.MessageOutput.GetMessage()
		if gErr != nil || msg == nil {
			continue
		}
		// Chỉ giữ lại văn bản trả lời cuối cùng của assistant; bỏ qua tool-call.
		if msg.Role == schema.Assistant && len(msg.ToolCalls) == 0 && strings.TrimSpace(msg.Content) != "" {
			reply = msg.Content
		}
	}
	return reply
}
