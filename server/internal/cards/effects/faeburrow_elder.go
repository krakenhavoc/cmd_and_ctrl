package effects

import (
	"strings"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Faeburrow Elder — Creature — Treefolk Druid {1}{G}{W}, 0/0:
//
//	"Vigilance
//	 This creature gets +1/+1 for each color among permanents you
//	 control.
//	 {T}: For each color among permanents you control, add one mana
//	 of that color."
//
// Vigilance rides PrintedKeywords. The pump is a self-only Layer 7c
// modify (not a CDA — CR 613.3a reserves layer 7a for "power and
// toughness are each equal to X", and this is a boost on top of a
// printed 0/0 base, not a definition of it), counting colours the
// same way the mana ability does: colorsAmongPermanentsYouControl
// walks EffectiveColors() once and both abilities share the read.
// The mana ability is fixed, not "any color" — it adds exactly the
// colours present, one of each, with no picker — so it is built with
// a plain ProducedFunc rather than OneColorOfAmount / pipe syntax.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "70a6f08e-854d-4e2f-9d8c-c45ec3231157",
		Name:            "Faeburrow Elder",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Static: []game.StaticAbility{{
			Layer:    game.Layer7PT,
			SubLayer: game.SubLayer7C_Modify,
			AppliesTo: func(target *game.Card, g *game.Game, source *game.Card) bool {
				return target.InstanceID == source.InstanceID
			},
			Apply: func(c *game.Characteristic, target *game.Card, g *game.Game, source *game.Card) {
				n := len(colorsAmongPermanentsYouControl(g, source.Controller))
				c.Power += n
				c.Toughness += n
			},
		}},
		ManaAbilities: []ManaAbility{{
			Cost:  ManaAbilityCost{Tap: true},
			Label: "For each color among permanents you control, add one mana of that color",
			ProducedFunc: func(g *game.Game, controller, _ uuid.UUID) string {
				var sb strings.Builder
				for _, col := range colorsAmongPermanentsYouControl(g, controller) {
					sb.WriteString("{" + col + "}")
				}
				return sb.String()
			},
		}},
	})
}

// colorsAmongPermanentsYouControl is "each color among permanents you
// control" (Faeburrow Elder's pump and its mana ability both read it):
// the union of EffectiveColors() over every permanent the controller
// controls, in canonical WUBRG order. A colourless permanent
// contributes nothing, and an empty board answers no colours.
func colorsAmongPermanentsYouControl(g *game.Game, controller uuid.UUID) []string {
	seen := map[string]bool{}
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller != controller {
			continue
		}
		for _, col := range c.EffectiveColors() {
			seen[strings.ToUpper(col)] = true
		}
	}
	var out []string
	for _, col := range []string{"W", "U", "B", "R", "G"} {
		if seen[col] {
			out = append(out, col)
		}
	}
	return out
}
