package main

import (
	"fmt"
	"math/rand/v2"
	"os"
	"strconv"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

type tile struct {
	isBomb        bool
	isCleared     bool
	isFlagged     bool
	adjacentBombs int // Predetermined before the game starts
}

type vector struct {
	x, y int
}

type model struct {
	grid     [][]tile // grid and whether or not there's a friggin bomb on it
	pointer  vector   // which to-do items are selected
	gameOver bool
}

func initialModel(gridSize int) model {
	// Make da grid
	// 1. Allocate the outer slice (array of rows)
	grid := make([][]tile, gridSize)

	// 2. Loop through and allocate each individual row slice
	for i := range grid {
		grid[i] = make([]tile, gridSize)
	}

	// 3. Randomly place bombs
	numBombs := gridSize * gridSize / 7
	threshold := 0.15 // likelihood of each pos placing a bomb
	for numBombs > 0 {
		for i := 0; i < len(grid); i++ {
			for j := 0; j < len(grid[0]); j++ {
				if numBombs > 0 && rand.Float64() < threshold {
					grid[i][j].isBomb = true
					numBombs--
				}
			}
		}
	}

	// 4. Figure out how many bombs each jit has around
	for i := 0; i < len(grid); i++ {
		for j := 0; j < len(grid[0]); j++ {
			if grid[i][j].isBomb {
				// i-1, j-1
				if i > 0 && j > 0 {
					grid[i-1][j-1].adjacentBombs += 1
				}
				// i-1, j
				if i > 0 {
					grid[i-1][j].adjacentBombs += 1
				}
				// i-1, j+1
				if i > 0 && j < len(grid[0])-1 {
					grid[i-1][j+1].adjacentBombs += 1
				}
				// i, j-1
				if j > 0 {
					grid[i][j-1].adjacentBombs += 1
				}
				// i, j+1
				if j < len(grid[0])-1 {
					grid[i][j+1].adjacentBombs += 1
				}
				// i+1, j-1
				if i < len(grid)-1 && j > 0 {
					grid[i+1][j-1].adjacentBombs += 1
				}
				// i+1, j
				if i < len(grid)-1 {
					grid[i+1][j].adjacentBombs += 1
				}

				// i+1, j+1
				if i < len(grid)-1 && j < len(grid[0])-1 {
					grid[i+1][j+1].adjacentBombs += 1
				}
			}
		}
	}

	return model{
		grid,
		vector{0, 0},
		false,
	}
}

func (m model) Init() tea.Cmd {
	// Just return `nil`, which means "no I/O right now, please."
	return nil
}

func (m *model) clearAtCursor() {
	tileAtCursor := &m.grid[m.pointer.y][m.pointer.x]
	// If there's a bomb here it's game ova
	if tileAtCursor.isBomb {
		m.gameOver = true
	}

	// If there's a flag here we should probably send a message to the screen but oh well
	if tileAtCursor.isBomb {
		m.gameOver = true
	} else {
		// Otherwise, algorithm to clear everything that isn't bordering a bomb
		var clearThisAndSurrounding func(x, y int)
		clearThisAndSurrounding = func(x, y int) {
			// If this is a 0, we keep spreading
			// If this is >0, then we just reveal this then stop
			m.grid[y][x].isCleared = true
			if m.grid[y][x].adjacentBombs == 0 {
				if y-1 >= 0 && !m.grid[y-1][x].isCleared {
					clearThisAndSurrounding(x, y-1)
				}
				if y+1 < len(m.grid) && !m.grid[y+1][x].isCleared {
					clearThisAndSurrounding(x, y+1)
				}
				if x-1 >= 0 && !m.grid[y][x-1].isCleared {
					clearThisAndSurrounding(x-1, y)
				}
				if x+1 < len(m.grid[0]) && !m.grid[y][x+1].isCleared {
					clearThisAndSurrounding(x+1, y)
				}
			}
		}
		clearThisAndSurrounding(m.pointer.x, m.pointer.y)
	}
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	// Is it a key press?
	case tea.KeyPressMsg:

		// Cool, what was the actual key pressed?
		switch msg.String() {

		// These keys should exit the program.
		case "ctrl+c", "q":
			return m, tea.Quit

		case "up", "k":
			if m.pointer.y > 0 {
				m.pointer.y--
			} else {
				m.pointer.y = len(m.grid) - 1
			}

		case "down", "j":
			if m.pointer.y < len(m.grid)-1 {
				m.pointer.y++
			} else {
				m.pointer.y = 0
			}

		case "left", "h":
			if m.pointer.x > 0 {
				m.pointer.x--
			} else {
				m.pointer.x = len(m.grid[0]) - 1
			}

		case "right", "l":
			if m.pointer.x < len(m.grid[0])-1 {
				m.pointer.x++
			} else {
				m.pointer.x = 0
			}

		case "f":
			m.grid[m.pointer.y][m.pointer.x].isFlagged = !m.grid[m.pointer.y][m.pointer.x].isFlagged

		case "space":
			m.clearAtCursor()
		}
	}

	// Return the updated model to the Bubble Tea runtime for processing.
	// Note that we're not returning a command.
	return m, nil
}

func (m model) View() tea.View {
	var b strings.Builder

	for y, row := range m.grid {
		for x, t := range row {
			glyph := " "
			style := lipgloss.NewStyle().Background(lipgloss.Color("240")) // hidden

			switch {
			case m.gameOver && t.isBomb:
				glyph = "*"
				style = style.Background(lipgloss.Color("196")).Foreground(lipgloss.Color("231"))
			case t.isFlagged:
				glyph = "f"
				style = style.Foreground(lipgloss.Color("208")).Bold(true)
			case t.isCleared:
				style = style.Background(lipgloss.Color("236"))
				glyph = strconv.Itoa(t.adjacentBombs)
			}

			l, r := " ", " "
			if m.pointer == (vector{x, y}) {
				l, r = "[", "]"
			}
			b.WriteString(style.Render(l + glyph + r))
		}
		b.WriteByte('\n')
	}

	if m.gameOver {
		b.WriteString("\nBOOM. Press q to quit.\n")
	} else {
		b.WriteString("\narrows move • space reveal • f flag • q quit\n")
	}

	v := tea.NewView(lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		Padding(1, 2).
		Margin(1, 2).
		Render(b.String()))
	v.AltScreen = true
	v.WindowTitle = "Minesweeper"
	return v
}

func main() {
	p := tea.NewProgram(initialModel(20))
	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}
