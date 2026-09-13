package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Tomb of the Spirit Dragon — Land (EDHREC rank 1759):
//
//	"{T}: Add {C}.
//	 {2}, {T}: You gain 1 life for each colorless creature you
//	 control."
//
// The Eldrazi deck's utility land. A painless {C} and a mana-plus-tap
// activated ability that counts the colorless creatures its
// controller controls when it resolves — effective colours, so a
// creature something painted a colour stops counting and a
// colourless token counts.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:      "22f6391e-2634-440f-af1b-9581d1bff818",
		Name:          "Tomb of the Spirit Dragon",
		Completeness:  CompletenessFull,
		ManaAbilities: []ManaAbility{painlessColorless()},
		Activated: []ActivatedAbility{{
			Label: "{2}, {T}: You gain 1 life for each colorless creature you control.",
			Cost:  Plus(ManaCost("{2}"), TapCost()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return GainLife{
					Player: item.Controller,
					Amount: b16ColorlessCreaturesControlled(g, item.Controller),
				}.Apply(NewContext(g, item))
			},
		}},
	})
}
