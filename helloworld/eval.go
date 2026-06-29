package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/cloudwego/eino/compose"
)

// evalCase là một kịch bản đánh giá: câu hỏi + góc nhìn (insight) cần quan sát.
type evalCase struct {
	Name    string // tên ngắn
	Insight string // góc nhìn: khía cạnh hành vi cần soi
	Input   string // câu hỏi đầu vào
}

// evalCases được chọn để phơi bày các khía cạnh khác nhau của hệ thống
// (không phải test functionality "trả ra string", mà soi hành vi thật).
var evalCases = []evalCase{
	{
		Name:    "Onboarding & định tuyến tool",
		Insight: "Có gọi cultural_greeter khi chào? Có cá nhân hóa theo tên? Gợi ý có bám nội dung lời chào (cô lập ngữ cảnh ngay từ lượt đầu)?",
		Input:   "Chào bot, mình là Linh.",
	},
	{
		Name:    "Kiến thức chuyên sâu (so sánh)",
		Insight: "Độ sâu & chính xác khi so sánh hai hạng mục; gợi ý có đào đúng vào điểm khác biệt thay vì hỏi lan man?",
		Input:   "Áo tấc khác gì so với áo ngũ thân?",
	},
	{
		Name:    "Giảng giải biểu tượng",
		Insight: "Khả năng giải thích ý nghĩa biểu tượng (không chỉ mô tả); gợi ý có đi vào chi tiết cụ thể của biểu tượng đó?",
		Input:   "Ý nghĩa của mâm ngũ quả trong ngày Tết?",
	},
	{
		Name:    "Phong tục & nguồn gốc",
		Insight: "Khả năng kể nguồn gốc + ý nghĩa; gợi ý có mở ra góc nhìn mới (hiện đại, biểu tượng) chứ không lặp câu hỏi?",
		Input:   "Tục ăn trầu của người Việt có từ bao giờ và ý nghĩa gì?",
	},
	{
		Name:    "Xử lý ngoài phạm vi văn hóa",
		Insight: "Khi bị hỏi lạc đề, agent xử lý ra sao (chuyển hướng/miễn cưỡng trả lời)? Gợi ý có bị trôi khỏi văn hóa hay neo đúng phạm vi?",
		Input:   "Hôm nay thời tiết Hà Nội thế nào?",
	},
}

// runEval chạy tất cả test case, tính chỉ số và ghi kết quả (input + output + insight) ra evaluation.md.
func runEval(ctx context.Context, runnable compose.Runnable[string, *ChatResponse]) error {
	type row struct {
		c                 evalCase
		resp              *ChatResponse
		rel               []int // số từ khóa mỗi gợi ý trùng với câu trả lời
		covHits, covTotal int   // từ khóa đặc trưng của input có mặt trong câu trả lời
		dur               time.Duration
	}
	var rows []row

	for i, c := range evalCases {
		fmt.Printf("[%d/%d] %s ... ", i+1, len(evalCases), c.Name)
		t0 := time.Now()
		resp, err := runnable.Invoke(ctx, c.Input)
		if err != nil {
			fmt.Println("LỖI:", err)
			resp = &ChatResponse{}
		}
		dur := time.Since(t0)
		rel := suggestRelevance(resp.Answer, resp.Suggestions)
		h, t := inputCoverage(resp.Answer, c.Input)
		rows = append(rows, row{c, resp, rel, h, t, dur})
		fmt.Printf("xong (%s, %d gợi ý)\n", dur.Round(time.Second), len(resp.Suggestions))
	}

	var b strings.Builder
	fmt.Fprintf(&b, "# Đánh giá Chatbot Văn hóa Việt\n\n")
	fmt.Fprintf(&b, "- Thời gian: %s\n", time.Now().Format("2006-01-02 15:04:05"))
	fmt.Fprintf(&b, "- Model: %s | BaseURL: %s\n", os.Getenv("OPENAI_MODEL"), os.Getenv("OPENAI_BASE_URL"))
	fmt.Fprintf(&b, "- Số case: %d\n\n", len(evalCases))

	relHit := func(r row) int {
		n := 0
		for _, x := range r.rel {
			if x > 0 {
				n++
			}
		}
		return n
	}

	b.WriteString("## Bảng tổng hợp\n\n")
	b.WriteString("| # | Case | Độ dài TL (ký tự) | Số gợi ý | Gợi ý bám TL | Bám input | Thời gian |\n")
	b.WriteString("|---|------|------------------|----------|--------------|-----------|----------|\n")
	for i, r := range rows {
		fmt.Fprintf(&b, "| %d | %s | %d | %d | %d/%d | %d/%d | %s |\n",
			i+1, r.c.Name, len([]rune(r.resp.Answer)), len(r.resp.Suggestions),
			relHit(r), len(r.resp.Suggestions), r.covHits, r.covTotal, r.dur.Round(time.Second))
	}
	b.WriteString("\n> **Gợi ý bám TL**: số gợi ý chia sẻ ≥1 từ khóa với câu trả lời — thước đo *cô lập ngữ cảnh* " +
		"(gợi ý phải sinh từ câu trả lời, không trôi sang chủ đề khác). **Độ dài TL** phản ánh độ chi tiết. " +
		"**Bám input** = từ khóa đặc trưng của câu hỏi có xuất hiện trong trả lời (có đi đúng trọng tâm không).\n\n---\n\n")

	for i, r := range rows {
		fmt.Fprintf(&b, "## Case %d — %s\n", i+1, r.c.Name)
		fmt.Fprintf(&b, "**Góc nhìn:** %s\n\n", r.c.Insight)
		fmt.Fprintf(&b, "**Input:** %s\n\n", r.c.Input)

		b.WriteString("**Trả lời:**\n\n")
		b.WriteString(r.resp.Answer)
		b.WriteString("\n\n")

		b.WriteString("**Gợi ý:**\n")
		if len(r.resp.Suggestions) == 0 {
			b.WriteString("_(không có)_\n")
		}
		for j, s := range r.resp.Suggestions {
			fmt.Fprintf(&b, "%d. %s _(trùng câu trả lời %d từ khóa)_\n", j+1, s, r.rel[j])
		}
		b.WriteString("\n")

		b.WriteString("**Chỉ số:**\n")
		fmt.Fprintf(&b, "- Độ dài trả lời: %d ký tự\n", len([]rune(r.resp.Answer)))
		fmt.Fprintf(&b, "- Số gợi ý: %d\n", len(r.resp.Suggestions))
		fmt.Fprintf(&b, "- Gợi ý bám câu trả lời: %d/%d\n", relHit(r), len(r.resp.Suggestions))
		fmt.Fprintf(&b, "- Bám input (từ khóa input có trong trả lời): %d/%d\n", r.covHits, r.covTotal)
		fmt.Fprintf(&b, "- Thời gian xử lý: %s\n", r.dur.Round(time.Millisecond))
		b.WriteString("\n---\n\n")
	}

	out := "evaluation.md"
	if err := os.WriteFile(out, []byte(b.String()), 0o644); err != nil {
		return err
	}
	fmt.Printf("\n✅ Đã ghi kết quả vào %s (%d bytes)\n", out, len(b.String()))
	return nil
}

// viStop là các từ dừng (tiếng Việt + một số tiếng Anh) để lọc khi đếm từ khóa.
var viStop = map[string]bool{
	"của": true, "và": true, "là": true, "cho": true, "mình": true, "có": true,
	"không": true, "các": true, "những": true, "về": true, "với": true, "được": true,
	"trong": true, "để": true, "một": true, "cũng": true, "như": true, "này": true,
	"đó": true, "khi": true, "đã": true, "sẽ": true, "vào": true, "từ": true,
	"theo": true, "vậy": true, "thì": true, "mà": true, "cái": true, "nhất": true,
	"nhiều": true, "hơn": true, "ra": true, "lên": true, "xuống": true, "bạn": true,
	"bot": true, "nhé": true, "nha": true, "thế": true, "nào": true, "gì": true,
	"the": true, "and": true, "for": true, "with": true, "that": true, "this": true,
}

// keywords tách tập từ khóa (đã lọc từ dừng, viết thường) từ một chuỗi.
func keywords(s string) map[string]bool {
	m := map[string]bool{}
	for _, w := range strings.Fields(s) {
		w = strings.ToLower(strings.Trim(w, ".,!?;:\"'()[]{}–—“”‘’…/-"))
		if len([]rune(w)) >= 3 && !viStop[w] {
			m[w] = true
		}
	}
	return m
}

// suggestRelevance đếm số từ khóa mỗi gợi ý xuất hiện trong câu trả lời.
// Số càng cao = gợi ý càng bám sát nội dung câu trả lời (cô lập ngữ cảnh tốt).
func suggestRelevance(answer string, suggestions []string) []int {
	ak := keywords(answer)
	out := make([]int, len(suggestions))
	for i, s := range suggestions {
		c := 0
		for w := range keywords(s) {
			if ak[w] {
				c++
			}
		}
		out[i] = c
	}
	return out
}

// inputCoverage đếm bao nhiêu từ khóa đặc trưng của input có mặt trong câu trả lời.
func inputCoverage(answer, input string) (hits, total int) {
	ak := keywords(answer)
	ik := keywords(input)
	for w := range ik {
		total++
		if ak[w] {
			hits++
		}
	}
	return hits, total
}
