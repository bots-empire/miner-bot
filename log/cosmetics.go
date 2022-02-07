package log

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"runtime"

	"github.com/mbndr/figlet4go"
)

func ClearTerminal() {
	switch runtime.GOOS {
	case "darwin", "linux":
		cmd := exec.Command("clear")
		cmd.Stdout = os.Stdout
		_ = cmd.Run()

	case "windows":
		panic("Fucking windows...")
	}
}

func PrintLogo(text string, colors []string) {
	parsedColors := make([]figlet4go.Color, len(colors))

	for i, color := range colors {
		var err error
		if parsedColors[i], err = figlet4go.NewTrueColorFromHexString(color); err != nil {
			panic(err.Error())
		}
	}

	render := figlet4go.NewAsciiRender()
	renderOptions := figlet4go.NewRenderOptions()

	if err := render.LoadFont("assets/larry3d.flf"); err != nil {
		panic(err.Error())
	}

	renderOptions.FontName = "larry3d"
	renderOptions.FontColor = parsedColors

	renderedString, err := render.RenderOpts(text, renderOptions)
	if err != nil {
		panic(err.Error())
	}

	fmt.Println(renderedString)
	for i := 0; i < 3; i++ {
		fmt.Println()
	}
}

func FormatData(data interface{}) string {
	result, err := json.MarshalIndent(data, "", "\t")
	if err != nil {
		panic("Unexpected error occurred: %s" + err.Error())
	}

	return string(result)
}
