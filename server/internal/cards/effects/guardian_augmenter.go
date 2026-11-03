package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Guardian Augmenter — Creature — Troll Wizard {2}{G}, 2/2 (Secrets of
// Strixhaven Commander, #4834):
//
//	"Flash
//	 Commander creatures you control get +2/+2.
//	 Commanders you control have hexproof."
//
// Bastion Protector's pump (layer 7c) over the commander creatures you
// control, and a layer-6 hexproof grant over every commander you control,
// creature or not, read off Card.IsCommander. Any commander you control
// counts, not just your own; a copy of a commander is not one, since
// being a commander is not a copiable value (the card's rulings).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "d925e560-faa1-40a9-b25a-47065ee891e6",
		Name:            "Guardian Augmenter",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Static: []game.StaticAbility{
			{
				Label:     "Commander creatures you control get +2/+2.",
				Layer:     game.Layer7PT,
				SubLayer:  game.SubLayer7C_Modify,
				AppliesTo: msc7CommanderCreatureYouControl,
				Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
					c.Power += 2
					c.Toughness += 2
				},
			},
			{
				Label: "Commanders you control have hexproof.",
				Layer: game.Layer6Ability,
				AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
					return target.IsCommander && target.Controller == source.Controller
				},
				Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
					c.Abilities = game.AppendKeywordAbility(c.Abilities, "hexproof")
				},
			},
		},
	})
}
