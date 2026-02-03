package interactive

import (
	"fmt"

	"github.com/eiannone/keyboard"
)

type MenuItem struct {
	Label string
	Value string
}

func SelectMenu(title string, items []MenuItem, currentIndex int) (int, error) {
	if err := keyboard.Open(); err != nil {
		return -1, err
	}
	defer keyboard.Close()

	selected := currentIndex
	if selected < 0 {
		selected = 0
	}

	for {
		// Clear screen and redraw
		fmt.Print("\033[H\033[2J")
		fmt.Printf("%s%s▸ %s%s\n\n", bold, orange, title, reset)

		// Display items
		for i, item := range items {
			if i == selected {
				fmt.Printf("%s▸ %s%s\n", orange, item.Label, reset)
			} else {
				fmt.Printf("  %s\n", item.Label)
			}
		}

		fmt.Printf("\n%sUse ↑↓ arrows, Enter/→ to select, ←/Esc to go back%s\n", orange, reset)

		// Read key
		_, key, err := keyboard.GetKey()
		if err != nil {
			return -1, err
		}

		switch key {
		case keyboard.KeyArrowUp:
			if selected > 0 {
				selected--
			}
		case keyboard.KeyArrowDown:
			if selected < len(items)-1 {
				selected++
			}
		case keyboard.KeyEnter, keyboard.KeyArrowRight:
			return selected, nil
		case keyboard.KeyEsc, keyboard.KeyArrowLeft:
			return -1, nil
		}
	}
}
