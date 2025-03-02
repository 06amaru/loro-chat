package style

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

var LoroTheme = tview.Theme{
	PrimitiveBackgroundColor:    tcell.ColorBlack,
	ContrastBackgroundColor:     tcell.NewHexColor(0xd2f3d2), // white green
	MoreContrastBackgroundColor: tcell.NewHexColor(0xa0b7a0), // MATRIX TERMINAL DARK
	SecondaryTextColor:          tcell.NewHexColor(0x8FE18F), // MATRIX TERMINAL
	TertiaryTextColor:           tcell.NewHexColor(0x78b078), //
}

var ButtonStyle = tcell.Style{}.
	Background(LoroTheme.TertiaryTextColor).
	Foreground(LoroTheme.PrimitiveBackgroundColor)

var BtnActivatedStyle = tcell.Style{}.
	Background(LoroTheme.SecondaryTextColor).
	Foreground(LoroTheme.PrimitiveBackgroundColor)

var CellSelectedtyle = tcell.Style{}.
	Background(LoroTheme.ContrastBackgroundColor).
	Foreground(LoroTheme.PrimitiveBackgroundColor)
