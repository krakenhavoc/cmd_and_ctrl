package cardsbl

import (
	"strings"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Bloom Tender — Creature — Elf Druid {1}{G}, 1/1 (EDHREC rank 258):
//
//	"Vivid — {T}: For each color among permanents you control, add
//	 one mana of that color."
//
// "Vivid" here is a bare ability word (no rules meaning of its own,
// same family as "Landfall") and carries no PrintedKeywords entry.
// The mana ability itself is a `ProducedFunc` — the S32 mana-pipeline
// slot for output computed at activation — built once here and read
// off the board: one mana symbol per distinct colour among the
// permanents the controller controls, in WUBRG order so the string is
// deterministic. HasColor reads the post-layer effective colour, so a
// permanent a Song of the Dryads or an Aura turned a colour counts,
// and colourless permanents contribute nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0c23fefe-9891-4dd8-9bb1-eebdb3274e31",
		Name:         "Bloom Tender",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{
			{
				Cost:         effects.ManaAbilityCost{Tap: true},
				ProducedFunc: bloomTenderProduced,
				Label:        "{T}: For each color among permanents you control, add one mana of that color",
			},
		},
	})
}

// bloomTenderProduced is "For each color among permanents you
// control, add one mana of that color" — one slot per distinct
// colour present, WUBRG order.
func bloomTenderProduced(g *game.Game, controller, _ uuid.UUID) string {
	var b strings.Builder
	for _, color := range []string{"W", "U", "B", "R", "G"} {
		for _, c := range g.BattlefieldCardsForEffect() {
			if c.Controller == controller && c.HasColor(color) {
				b.WriteString("{" + color + "}")
				break
			}
		}
	}
	return b.String()
}
