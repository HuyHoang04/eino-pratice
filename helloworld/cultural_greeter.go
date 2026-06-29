package main

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"
)

// greetingCustoms là các phong tục chào hỏi xưa của người Việt.
// Mỗi câu được viết dưới dạng vế nối tiếp ngay sau cụm "Bạn có biết ... không?".
var greetingCustoms = []string{
	"vào thời Nguyễn, người xưa thường chắp tay ngang ngực và cúi đầu khi chào nhau để thể hiện sự tôn trọng",
	"trẻ em người Việt xưa khi chào người lớn thường khoanh tay, cúi gập người và nói 'cháu chào ạ' để bày tỏ sự lễ phép",
	"người Việt có tục chọn cách xưng hô 'anh – chị – chú – bác – cô – dì' theo tuổi tác và vai vế, thay vì gọi thẳng tên như trong tiếng Anh",
	"khi có khách đến nhà, người Việt xưa thường mời ngay một chén nước chè xanh nóng như một lời chào mừng",
	"người xưa coi câu hỏi thăm 'ông khỏe không?', 'cụ đi đâu đấy?' cũng chính là lời chào hỏi đầy ân cần",
	"tục cúi đầu nhún mình khi chào người lớn tuổi là cách người Việt bày tỏ sự khiêm nhường và kính trọng bề trên",
}

// culturalFacts là những sự thật thú vị về văn hóa Việt Nam,
// cũng được viết nối tiếp ngay sau cụm "Bạn có biết ... không?".
var culturalFacts = []string{
	"tục ăn trầu đã theo người Việt hàng ngàn năm, từng là biểu tượng của tình bằng hữu và sự hiếu khách",
	"phong bao lì xì màu đỏ dịp Tết Nguyên Đán mang ý nghĩa mong một năm mới may mắn và bình an",
	"áo dài – quốc phục của Việt Nam – được định hình từ thế kỷ 18 dưới thời chúa Nguyễn Phúc Khoát",
	"mâm ngũ quả ngày Tết tượng trưng cho ngũ hành Kim – Mộc – Thủy – Hỏa – Thổ, cầu mong một năm sung túc",
	"vào ngày 23 tháng Chạp âm lịch, người Việt cúng ông Công ông Táo tiễn ba vị thần về trời báo cáo việc nhà với Ngọc Hoàng",
	"câu đối Tết màu đỏ viết bằng chữ nho là lời cầu mong tài lộc, cát tường cho cả gia đình trong năm mới",
}

// pickCulturalContent chọn ngẫu nhiên (nhưng xác định theo ngày) một câu
// phong tục chào hỏi hoặc sự thật văn hóa. Cùng một ngày luôn ra cùng một nội dung,
// giúp trải nghiệm ổn định và dễ kiểm thử — đúng tinh thần "theo ngày hiện tại".
func pickCulturalContent(now time.Time) string {
	combined := make([]string, 0, len(greetingCustoms)+len(culturalFacts))
	combined = append(combined, greetingCustoms...)
	combined = append(combined, culturalFacts...)

	// Seed bầm theo (năm, ngày trong năm) để nội dung thay đổi theo từng ngày.
	r := rand.New(rand.NewSource(int64(now.Year()*1000 + now.YearDay())))
	return combined[r.Intn(len(combined))]
}

// composeGreeting ghép lời chào mang bản sắc: chào tên + một nét văn hóa + câu mời mở.
func composeGreeting(name, content string) string {
	display := name
	if display == "" {
		display = "bạn"
	}
	return fmt.Sprintf(
		"Chào %s. Bạn có biết %s không? Hôm nay bạn muốn tìm hiểu về nét văn hóa nào?",
		display, content,
	)
}

// CulturalGreetingInput là tham số đầu vào của công cụ cultural_greeter.
type CulturalGreetingInput struct {
	Name string `json:"name" jsonschema:"description=Tên của người dùng đang chào hỏi, ví dụ 'Linh'. Bắt buộc."`
}

// CulturalGreetingOutput là kết quả trả về của cultural_greeter.
type CulturalGreetingOutput struct {
	Greeting string `json:"greeting"`
}

// NewCulturalGreeterTool đóng gói thư viện helloworld thành một Tool tên cultural_greeter:
// tiếp nhận tên người dùng, trả về lời chào mang bản sắc văn hóa Việt
// (một phong tục chào hỏi xưa hoặc một sự thật văn hóa thú vị theo ngày hiện tại).
func NewCulturalGreeterTool() (tool.InvokableTool, error) {
	return utils.InferTool(
		"cultural_greeter",
		"Sinh lời chào tiếp đón mang bản sắc văn hóa Việt Nam cho người dùng mới. "+
			"Gọi công cụ này khi người dùng chào hỏi hoặc tự giới thiệu tên, "+
			"truyền vào trường name là tên của họ. Trả về một câu chào kết hợp "+
			"một phong tục chào hỏi xưa hoặc một sự thật văn hóa thú vị theo ngày hiện tại.",
		func(ctx context.Context, input *CulturalGreetingInput) (*CulturalGreetingOutput, error) {
			content := pickCulturalContent(time.Now())
			return &CulturalGreetingOutput{
				Greeting: composeGreeting(input.Name, content),
			}, nil
		},
	)
}
