package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/signintech/gopdf"
)

type Resume struct {
	Name       string         `json:"name"`
	Title      string         `json:"title"`
	Picture    string         `json:"picture"`
	Contacts   []IconText     `json:"contacts"`
	Socials    []IconTextLink `json:"socials"`
	Summary    IconText       `json:"summary"`
	Experience []Job          `json:"experience"`
	Projects   []Project      `json:"projects"`
	Education  []School       `json:"education"`
	Skills     []Skill        `json:"skills"`
	lastX      float64
	lastY      float64
}

type IconText struct {
	Value string `json:"value"`
	Icon  string `json:"icon"`
}

type IconTextLink struct {
	Name string `json:"name"`
	Link string `json:"link"`
	Icon string `json:"icon"`
}

type Project struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Link        string `json:"link"`
}

type Job struct {
	Title       string `json:"title"`
	Company     string `json:"company"`
	Location    string `json:"location"`
	Dates       string `json:"dates"`
	Description string `json:"description"`
}

type School struct {
	Degree      string `json:"degree"`
	SchoolName  string `json:"school"`
	Dates       string `json:"dates"`
	Description string `json:"description"`
}

type SkillItem struct {
	Name  string `json:"name"`
	Level string `json:"level"`
	Icon  string `json:"icon"`
}

type Skill struct {
	Category string      `json:"category"`
	Items    []SkillItem `json:"items"`
}

type TextColor struct {
	R uint8
	G uint8
	B uint8
}

var (
	BLACK     = TextColor{R: 0, G: 0, B: 0}
	DARK_GREY = TextColor{R: 66, G: 66, B: 66}
	BLUE      = TextColor{R: 0, G: 122, B: 255}
)

func (r *Resume) Load(filename string) error {
	b, err := os.ReadFile(filename)
	if err != nil {
		panic(err)
	}
	return json.Unmarshal(b, r)
}

func (r *Resume) createResumePdf() {
	pdf := gopdf.GoPdf{}
	pdf.Start(gopdf.Config{PageSize: *gopdf.PageSizeA4})
	pdf.AddPage()

	fontFamilyMap := map[string]string{
		"Roboto":              "fonts/Roboto-Regular.ttf",
		"Roboto-Bold":         "fonts/Roboto-Bold.ttf",
		"Roboto-Italic":       "fonts/Roboto-Italic.ttf",
		"Roboto-BoldItalic":   "fonts/Roboto-BoldItalic.ttf",
		"Roboto-Black":        "fonts/Roboto-Black.ttf",
		"Roboto-BlackItalic":  "fonts/Roboto-BlackItalic.ttf",
		"Roboto-Light":        "fonts/Roboto-Light.ttf",
		"Roboto-LightItalic":  "fonts/Roboto-LightItalic.ttf",
		"Roboto-Medium":       "fonts/Roboto-Medium.ttf",
		"Roboto-MediumItalic": "fonts/Roboto-MediumItalic.ttf",
		"Roboto-Thin":         "fonts/Roboto-Thin.ttf",
		"Roboto-ThinItalic":   "fonts/Roboto-ThinItalic.ttf",
	}

	for fontFamily, fontPath := range fontFamilyMap {
		err := pdf.AddTTFFontWithOption(fontFamily, fontPath, gopdf.TtfOption{
			UseKerning: true,
		})
		if err != nil {
			panic(err)
		}
	}

	r.addAllDetailsToResumePdf(&pdf)
	pdf.WritePdf("resume.pdf")
}

func (r *Resume) addText(pdf *gopdf.GoPdf, x float64, y float64, fontFamily string, size int, text string, textColor TextColor) (float64, float64) {

	pdf.SetXY(x, y)
	pdf.SetTextColor(textColor.R, textColor.G, textColor.B)
	pdf.SetLineWidth(0.5)
	pdf.SetFillColor(0, 0, 0)
	pdf.SetFont(fontFamily, "", size)
	pdf.Cell(nil, text)
	lastX := pdf.GetX()
	lastY := pdf.GetY()
	pdf.Br(float64(size))
	return lastX, lastY
}

func (r *Resume) addImage(pdf *gopdf.GoPdf, x float64, y float64, width float64, height float64, image string) {
	fmt.Println(image)
	pdf.Image(image, x, y, &gopdf.Rect{W: width, H: height})
	pdf.Br(20)
}

func (r *Resume) addTextWithIcon(pdf *gopdf.GoPdf, x float64, y float64, fontFamily string, size int, text string, icon string, link string, textColor TextColor) (float64, float64) {
	var x_delta float64 = 0
	if icon != "" {
		r.addImage(pdf, x, y, float64(size), float64(size), icon)
		x_delta = 20
	}
	lastX, lastY := r.addText(pdf, x+15, y, fontFamily, size, text, textColor)
	if link != "" {
		link_x := x + x_delta + float64(len(text)*size)*0.415
		link_width := float64(len(link)*size) * 0.5
		link_height := float64(size + 2)

		lastX, lastY = r.addText(pdf, link_x, y, fontFamily, size, link, BLUE)
		pdf.AddExternalLink(link, link_x, y, link_width, link_height)
	}
	return lastX, lastY
}

func (r *Resume) addHeader(pdf *gopdf.GoPdf) {
	const (
		IMAGE_WIDTH  = 90
		IMAGE_HEIGHT = 90
		IMAGE_X      = 10
		IMAGE_Y      = 5

		Name_X = IMAGE_X + IMAGE_WIDTH + 15
		Name_Y = IMAGE_Y

		Title_X = Name_X
		Title_Y = Name_Y + 15
	)

	r.addImage(pdf, IMAGE_X, IMAGE_Y, IMAGE_WIDTH, IMAGE_HEIGHT, r.Picture)
	r.addText(pdf, Name_X, Name_Y, "Roboto-Bold", 14, r.Name, BLACK)
	r.addText(pdf, Title_X, Title_Y, "Roboto-Medium", 14, r.Title, BLACK)

	var contactX float64 = Name_X - 10
	for _, contact := range r.Contacts {
		contactX, _ = r.addTextWithIcon(pdf, contactX+10, Title_Y+15, "Roboto", 12, contact.Value, contact.Icon, "", BLACK)
	}

	socialCounter := 1
	for _, social := range r.Socials {
		socialY := 5 + float64(socialCounter*14) + Title_Y + 15
		r.addTextWithIcon(pdf, Name_X, socialY, "Roboto-Light", 10, social.Name+" - ", social.Icon, social.Link, BLACK)
		socialCounter++
	}
}

func (r *Resume) drawLine(pdf *gopdf.GoPdf, x float64, y float64, width float64) {
	pdf.SetLineWidth(0.75)
	pdf.SetLineType("solid")
	pdf.Line(x, y, x+width, y)
}

func (r *Resume) addSummary(pdf *gopdf.GoPdf) {
	const (
		Summary_X = 10
		Summary_Y = 105
	)

	r.addTextWithIcon(pdf, Summary_X, Summary_Y, "Roboto-Bold", 14, "Summary", r.Summary.Icon, "", DARK_GREY)
	r.drawLine(pdf, Summary_X, Summary_Y+15, gopdf.PageSizeA4.W-Summary_X*2)
	var lineNumber int = 1
	for _, line := range strings.Split(r.Summary.Value, "\n") {
		r.addText(pdf, Summary_X, Summary_Y+float64(lineNumber*16)+8, "Roboto", 12, line, BLACK)
		lineNumber++
	}
}

func (r *Resume) addSkills(pdf *gopdf.GoPdf) {
	const Skills_X float64 = 10
	var Skills_Y float64 = float64(105+24*len(strings.Split(r.Summary.Value, "\n"))) + 10

	r.addTextWithIcon(pdf, Skills_X, Skills_Y, "Roboto-Bold", 14, "Skills", r.Summary.Icon, "", DARK_GREY)
	r.drawLine(pdf, Skills_X, Skills_Y+15, 90)

	var skillCatergoryNumber int = 1
	var skillItemNumber int = 1
	for _, skill := range r.Skills {
		r.addText(pdf, Skills_X, Skills_Y+float64((skillCatergoryNumber+skillItemNumber)*16)-6, "Roboto-Bold", 12, skill.Category, BLACK)
		for _, item := range skill.Items {
			r.addText(pdf, Skills_X+5, Skills_Y+float64(skillCatergoryNumber*16+skillItemNumber*16)+12, "Roboto", 12, item.Name, BLACK)
			skillItemNumber++
		}
		skillCatergoryNumber++
	}
}

func (r *Resume) addExperiences(pdf *gopdf.GoPdf) {
	const Experiences_X float64 = 110
	var Experiences_Y float64 = float64(105+24*len(strings.Split(r.Summary.Value, "\n"))) + 10

	r.addTextWithIcon(pdf, Experiences_X, Experiences_Y, "Roboto-Bold", 14, "Experience", r.Summary.Icon, "", DARK_GREY)
	r.drawLine(pdf, Experiences_X, Experiences_Y+15, gopdf.PageSizeA4.W-Experiences_X-10)

	var deltaY float64 = 24
	for _, job := range r.Experience {
		lastX, lastY := r.addText(pdf, Experiences_X, Experiences_Y+deltaY, "Roboto-Bold", 12, job.Title, DARK_GREY)
		lastX, _ = r.addText(pdf, lastX+4, Experiences_Y+deltaY, "Roboto-Bold", 12, " | "+job.Company, DARK_GREY)
		lastX, _ = r.addText(pdf, lastX+4, lastY+2, "Roboto-Light", 8, " | "+job.Dates, BLACK)
		r.addText(pdf, lastX+4, Experiences_Y+deltaY+2, "Roboto-Light", 8, " ("+job.Location+" )", BLACK)
		deltaY += 16

		for _, line := range strings.Split(job.Description, "\n") {
			r.addText(pdf, Experiences_X, Experiences_Y+deltaY, "Roboto", 10, line, BLACK)
			deltaY += 14
		}
		deltaY += 6
	}

	r.lastX = Experiences_X
	r.lastY = Experiences_Y + deltaY
}

func (r *Resume) addProjects(pdf *gopdf.GoPdf) {
	const Projects_X = 110

	r.addTextWithIcon(pdf, Projects_X, r.lastY, "Roboto-Bold", 14, "Projects", r.Summary.Icon, "", DARK_GREY)
	r.drawLine(pdf, Projects_X, r.lastY+15, gopdf.PageSizeA4.W-Projects_X-10)

	var deltaY float64 = 24
	for _, project := range r.Projects {
		r.addText(pdf, Projects_X, r.lastY+deltaY, "Roboto-Bold", 12, project.Name, DARK_GREY)
		r.addTextWithIcon(pdf, Projects_X, r.lastY+deltaY+15, "Roboto-Light", 10, "", "", project.Link, BLACK)
		r.addText(pdf, Projects_X, r.lastY+deltaY+30, "Roboto-Light", 10, project.Description, BLACK)
		deltaY += 45
	}

	r.lastX = Projects_X
	r.lastY += deltaY
}

func (r *Resume) addEducation(pdf *gopdf.GoPdf) {
	var Education_X = r.lastX
	var Education_Y = r.lastY

	r.addTextWithIcon(pdf, Education_X, Education_Y+5, "Roboto-Bold", 14, "Education", r.Summary.Icon, "", DARK_GREY)
	r.drawLine(pdf, Education_X, Education_Y+20, gopdf.PageSizeA4.W-Education_X-10)

	var deltaY float64 = 30
	for _, school := range r.Education {
		r.addText(pdf, Education_X, Education_Y+deltaY, "Roboto-Bold", 12, school.Degree, DARK_GREY)
		deltaY += 15
		r.addText(pdf, Education_X, Education_Y+deltaY, "Roboto", 10, school.SchoolName, BLACK)

		for _, line := range strings.Split(school.Description, "\n") {
			deltaY += 15
			r.addText(pdf, Education_X, Education_Y+deltaY, "Roboto-Light", 10, line, BLACK)
		}
	}

	r.lastX = Education_X
	r.lastY += deltaY
}

func (r *Resume) addAllDetailsToResumePdf(pdf *gopdf.GoPdf) {
	r.addHeader(pdf)
	r.addSummary(pdf)
	r.addSkills(pdf)
	r.addExperiences(pdf)
	r.addProjects(pdf)
	r.addEducation(pdf)
}

func main() {
	var r Resume
	err := r.Load("resume.json")
	if err != nil {
		panic(err)
	}
	r.createResumePdf()
}
