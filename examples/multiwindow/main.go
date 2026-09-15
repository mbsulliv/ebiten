// Copyright 2026 The Ebitengine Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Several windows: the key that types N opens another window, Escape closes the window the key was pressed in,
// the close button closes a window, and the program ends with the last one. Each window counts its own ticks
// and shows which window has the focus, the last key typed and where the cursor is, so the per-window input and
// the current-window resolution can be seen.
package main

import (
	"fmt"
	"image/color"
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

type game struct {
	n     int // which window this game runs in, for the caption
	ticks int
	x     float64
	hue   color.RGBA

	keys    []ebiten.Key
	lastKey string
}

var opened int

func (g *game) Update() error {
	g.ticks++
	// Key constants are physical positions on a US layout; the letter a key types on the user's layout comes
	// from KeyName, so "N" is the key labelled N on a Dvorak keyboard too. The last key typed is shown.
	g.keys = inpututil.AppendJustPressedKeys(g.keys[:0])
	pressedN := false
	for _, k := range g.keys {
		name := ebiten.KeyName(k)
		if name != "" { // a modifier types nothing: it does not replace the last letter
			g.lastKey = name
		}
		if name == "n" || (name == "" && k == ebiten.KeyN) {
			pressedN = true
		}
	}
	g.x += 2
	if g.x > 400 {
		g.x = 0
	}
	if pressedN {
		opened++
		w, err := ebiten.OpenWindow(&game{n: opened + 1, hue: color.RGBA{uint8(60 * opened), 120, 200, 255}}, &ebiten.WindowOptions{
			Title:        fmt.Sprintf("Window %d", opened+1),
			Width:        420,
			Height:       300,
			ResizingMode: ebiten.WindowResizingModeEnabled,
		})
		if err != nil {
			log.Println("OpenWindow:", err)
		} else {
			log.Println("opened window", w.ID())
		}
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		return ebiten.Termination // closes this window; the last one ends the program
	}
	return nil
}

func (g *game) Draw(screen *ebiten.Image) {
	screen.Fill(color.RGBA{30, 30, 36, 255})
	vector.DrawFilledRect(screen, float32(g.x), 60, 40, 40, g.hue, false)
	cx, cy := ebiten.CursorPosition()
	w := ebiten.CurrentWindow()
	id := 0
	if w != nil {
		id = w.ID()
	}
	mods := ""
	for _, m := range []struct {
		k ebiten.Key
		s string
	}{{ebiten.KeyShift, "shift"}, {ebiten.KeyControl, "ctrl"}, {ebiten.KeyAlt, "alt"}, {ebiten.KeyMeta, "meta"}} {
		if ebiten.IsKeyPressed(m.k) {
			mods += " " + m.s
		}
	}
	msg := fmt.Sprintf("window %d (id %d)\nticks %d  focused %v  last key %q  held:%s\ncursor %d,%d  screen %v\nN: open a window  Esc: close this one",
		g.n, id, g.ticks, ebiten.IsFocused(), g.lastKey, mods, cx, cy, func() string { a, b := ebiten.ScreenSize(); return fmt.Sprintf("%dx%d", a, b) }())
	ebitenutil.DebugPrint(screen, msg)
}

func (g *game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return outsideWidth, outsideHeight
}

func main() {
	ebiten.SetWindowTitle("Window 1")
	ebiten.SetWindowSize(420, 300)
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	if err := ebiten.RunGame(&game{n: 1, hue: color.RGBA{220, 120, 60, 255}}); err != nil {
		log.Fatal(err)
	}
}
