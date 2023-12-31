package gani

import (
	"bufio"
	"bytes"
	"strings"
)

// Need some examples before I can really do anything w. these.
// TODO: SCRIPT block(s)? [Anything else we are missing?]
// TODO: COLOREFFECT?

// Gani defines the Gani file format
type Gani struct {
	Sprites  []*Sprite
	Settings *Settings
	Frames   []*Frame

	header string
}

func (g *Gani) String() string {
	builder := strings.Builder{}

	builder.WriteString(g.header)
	builder.WriteString("\n")

	for _, sprite := range g.Sprites {
		builder.WriteString(sprite.String())
		builder.WriteString("\n")
	}
	builder.WriteString("\n")

	builder.WriteString(g.Settings.String())
	builder.WriteString("\n")

	builder.WriteString("ANI\n")

	for i, frame := range g.Frames {
		builder.WriteString(frame.String())

		if i != len(g.Frames)-1 {
			builder.WriteString("\n")
		}
	}

	builder.WriteString("ANIEND\n")

	return builder.String()
}

// Parse extracts the Gani data from the incoming buffer
func (g *Gani) Parse(b []byte) error {
	scanner := bufio.NewScanner(bytes.NewReader(b))
	scanner.Scan()

	// SPRITE lines
	for scanner.Scan() {
		if strings.EqualFold(scanner.Text(), "") {
			break
		}

		sprite, err := NewSprite(scanner.Text())
		if err != nil {
			return err
		}

		g.Sprites = append(g.Sprites, sprite)
	}

	// Settings lines
	for scanner.Scan() {
		if strings.EqualFold(scanner.Text(), "") {
			break
		}

		if err := g.Settings.Parse(scanner.Text()); err != nil {
			return err
		}
	}

	// ANI - ANIEND
	for scanner.Scan() {
		if strings.EqualFold(scanner.Text(), "ANI") {
			continue
		}

		// Populate the Frame data
		frame := NewFrame(g.Settings.SingleDirection)
		for i := 0; i < 4; i++ {
			err := frame.AppendPlacedSprites(scanner.Text(), FrameDirection(i))
			if err != nil {
				return err
			}

			if !scanner.Scan() {
				break
			}

			// Break if we new-line or hit a wait/playsound
			if strings.EqualFold(scanner.Text(), "") ||
				strings.HasPrefix(scanner.Text(), "WAIT") ||
				strings.HasPrefix(scanner.Text(), "PLAYSOUND") {
				break
			}
		}

		// Populate the WAIT & PLAYSOUND(s)
		for !strings.EqualFold(scanner.Text(), "") && !strings.EqualFold(scanner.Text(), "ANIEND") {
			if err := frame.ParseWaitOrSound(scanner.Text()); err != nil {
				return err
			}

			if !scanner.Scan() {
				break
			}
		}
		g.Frames = append(g.Frames, frame)

		// Reached the end
		if strings.EqualFold(scanner.Text(), "ANIEND") {
			break
		}
	}

	return nil
}

// NewGani creates a new empty Gani
func NewGani() *Gani {
	return &Gani{
		Sprites:  make([]*Sprite, 0),
		Settings: NewSettings(),
		Frames:   make([]*Frame, 0),
		header:   "GANI0001",
	}
}
