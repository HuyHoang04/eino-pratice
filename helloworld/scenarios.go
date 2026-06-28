package main

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/schema"
)

// scenario is one scripted customer conversation (one Turn = one customer message).
type scenario struct {
	Name        string
	Description string
	Turns       []string
}

// scenarios is the demo catalog.
var scenarios = []scenario{
	{
		Name:        "happy_vi",
		Description: "Vietnamese happy path: context up front, then contact, then consult.",
		Turns: []string{
			"chào shop ạ, em là Hoàng bên công ty bán lẻ, bên em tầm 20 nhân viên thôi, dạo này quản lý khách hàng toàn bằng excel nên hay bị lộn xộn, lỡ mất dữ liệu, giờ em muốn tìm cái CRM nào gọn nhẹ xài thử",
			"hoang@retailco.vn",
			"dạ vâng em cảm ơn shop tư vấn nha",
		},
	},
	{
		Name:        "happy_en",
		Description: "English input -> agent should reply in English (language matching).",
		Turns: []string{
			"hi there, I'm Hoang from a 20-person retail team, we're drowning in Excel spreadsheets and keep losing customer data, looking for a CRM to clean things up",
			"hoang@retailco.vn",
			"great, thanks!",
		},
	},
	{
		Name:        "autofill_vi",
		Description: "Seeded customer (Lan) -> looked up by phone (unique id), company/team/current tool auto-filled from CRM, no re-ask.",
		Turns: []string{
			"chào shop ạ, em là Lan,  Bên em đang muốn đổi sang CRM mới cho dễ dùng, shop tư vấn giúp em nha",
			"số điện thoại em là 0909 112 233",
			"dạ vâng, vậy mấy gói khác nhau gì không ạ",
			"oke cảm mơn shop",
		},
	},
	{
		Name:        "refuse_vi",
		Description: "Customer refuses phone -> pivot to email, no handoff.",
		Turns: []string{
			"chào, mình là Hùng, bên công ty mình có team sales cần phần mềm quản lý khách hàng",
			"mình không muốm, bạn cứ tư trước đi",
			"team mình có khoảng 12 người thôi",
			"team mình chưa có gì hết hiện tại nên mình muốn tìm cái CRM nào gọn nhẹ xài thử",
			"oke bạn có thể gửi thời gmail cty mình supertalent@gmail.com",
		},
	},
	{
		Name:        "hard_vi",
		Description: "Evasive customer keeps dodging contact and asks about price -> answer helpfully, don't repeat the ask, let them come around.",
		Turns: []string{
			"chào, mình là Tuấn bên công ty vận tải, mình cần tìm CRM quản lý khách hàng",
			"khoan đi nhà mình gói rẻ nhất giá bao nhiêu, có dùng thử miễn phí không",
			"mình cảm giác gói đó không đủ cho team mình đâu",
			"ok liên lạc với mình qua email tuan@transportinglumia.vn",
			"team mình có khoảng 10 người thôi",
			"ok được rồi",
		},
	},
	{
		Name:        "hurry_vi",
		Description: "Customer is in a hurry and complains about price/service -> answer helpfully, don't repeat the ask, let them come around.",
		Turns: []string{
			"chào mình đang vội, vào thẳng bên bạn có gì tư vấn luôn được không",
			"sao đắt thế và dịch vụ kém quá vậy !",
		},
	},
	{
		Name:        "handoff_vi",
		Description: "Customer refuses all contact channels -> recommend human handoff.",
		Turns: []string{
			"chào, mình là Thanh, mình muốn tìm hiểu về CRM bên mình",
			"mình không muốn để lại thông tin liên lạc gì cả, tư vấn qua đây luôn được không",
		},
	},
	{
		Name:        "sparse_vi",
		Description: "Terse customer -> priority order name -> need -> contact over 3 turns.",
		Turns: []string{
			"cho mình hỏi về crm",
			"mình là Đức, bên mình cần quản lý khách hàng",
			"duc@gmail.com",
		},
	},
}

// runDemo replays scenario(s) by name (or "all"); unknown names exit with a list.
func runDemo(ctx context.Context, runner *adk.Runner, selector string) {
	pick := scenarios
	if selector != "all" {
		s, ok := findScenario(selector)
		if !ok {
			log.Fatalf("unknown scenario %q. Available: %s", selector, scenarioList())
		}
		pick = []scenario{s}
	}
	for _, s := range pick {
		fmt.Printf("══ SCENARIO: %s — %s ══\n", s.Name, s.Description)
		// Fresh history per scenario so each starts from a blank lead.
		history := []adk.Message{}
		for _, turn := range s.Turns {
			fmt.Printf("\nKhách: %s\n", turn)
			history = append(history, schema.UserMessage(turn))
			reply := runTurn(ctx, runner, history)
			fmt.Printf("Agent: %s\n", reply)
			history = append(history, schema.AssistantMessage(reply, nil))
		}
		fmt.Println()
	}
}

func findScenario(name string) (scenario, bool) {
	for _, s := range scenarios {
		if s.Name == name {
			return s, true
		}
	}
	return scenario{}, false
}

func scenarioList() string {
	names := make([]string, len(scenarios))
	for i, s := range scenarios {
		names[i] = s.Name
	}
	return strings.Join(names, ", ")
}
