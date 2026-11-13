package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Polygoyf — Creature — Lhurgoyf {2}{G}, */1+* (Modern Horizons 3
// Commander, #5468):
//
//	"Trample, myriad
//	 Polygoyf's power is equal to the number of card types among cards in
//	 all graveyards and its toughness is equal to that number plus 1."
//
// The power and toughness are Tarmogoyf's characteristic-defining
// ability (layer 7a): the union of card types over every graveyard,
// recomputed whenever a card enters or leaves one. Myriad is the shared
// trigger (myriad.go): one "you may" per opponent other than the
// defending player, the copies exiled at end of combat. Each copy is a
// Polygoyf, so it has the same power and toughness and trample.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "6021f2e3-f915-428c-8289-2aa675b5bc78",
		Name:            "Polygoyf",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Static: []game.StaticAbility{{
			Label:    "Power is equal to the number of card types among cards in all graveyards, and its toughness is that number plus 1.",
			Layer:    game.Layer7PT,
			SubLayer: game.SubLayer7A_CDA,
			AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.InstanceID == source.InstanceID
			},
			Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, _ *game.Card) {
				n := game.DistinctCardTypesInAllGraveyards(g)
				c.Power = n
				c.Toughness = n + 1
			},
		}},
		Triggered: []game.TriggeredAbility{
			GrantedMyriad("Polygoyf"),
		},
	})
}
