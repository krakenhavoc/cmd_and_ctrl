package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Starting Column — Artifact {3}:
//
//	"Start your engines!
//	 {T}: Add one mana of any color.
//	 Max speed — {T}, Sacrifice this artifact: Draw two cards, then
//	 discard a card."
//
// ADR 0136 (#2122). The loot is an ordinary activated ability,
// activatable only while its controller has max speed (CR 702.178a).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "d608a3fa-faee-44a4-9ab6-be701d3b7a49",
		Name:            "Starting Column",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{StartYourEngines},
		ManaAbilities: []ManaAbility{
			{Cost: ManaAbilityCost{Tap: true}, Produced: "{W|U|B|R|G}", Label: "{T}: Add one mana of any color."},
		},
		Activated: []ActivatedAbility{
			MaxSpeedActivated(ActivatedAbility{
				Label: "Max speed — {T}, Sacrifice this artifact: Draw two cards, then discard a card.",
				Cost:  Plus(TapCost(), SacrificeThis()),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return b16DrawThenDiscard(g, item, 2, 1)
				},
			}),
		},
	})
}
