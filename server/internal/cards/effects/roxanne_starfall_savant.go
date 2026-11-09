package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Roxanne, Starfall Savant — Legendary Creature — Cat Druid {3}{R}{G}, 4/4
// (Outlaws of Thunder Junction, #4468):
//
//	"Whenever Roxanne enters or attacks, create a tapped colorless artifact
//	 token named Meteorite with 'When this token enters, it deals 2 damage
//	 to any target' and '{T}: Add one mana of any color.'
//	 Whenever you tap an artifact token for mana, add one mana of any type
//	 that artifact token produced."
//
// The token is std_b35_meteorite.go's template, created tapped. The second
// ability is a triggered mana ability (CR 605.1b) over any mana its
// controller's artifact token produces (Kinnan's shape, narrowed to artifact
// tokens): it adds one mana of a type that production made.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "11765eaf-87a1-48bd-90c1-7a5f21be5918",
		Name:         "Roxanne, Starfall Savant",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Tokens: 1},
		Triggered: []game.TriggeredAbility{
			WhenThisEntersOrAttacks("Roxanne, Starfall Savant — create a tapped Meteorite token",
				Do(CreateTokenAdvanced{Spec: Token(MeteoriteToken()).EntersTapped(), N: 1})),
		},
		ManaTriggers: []game.ManaTrigger{{
			Label: "Roxanne, Starfall Savant — add one mana of any type that artifact token produced",
			AppliesTo: func(prod game.ManaProduced, source *game.Card, _ *game.Game) bool {
				return prod.Source.IsToken() && prod.Source.IsArtifact() && prod.Controller == source.Controller
			},
			Produced: AddsOneManaOfAnyTypeProduced(),
		}},
	})
}
