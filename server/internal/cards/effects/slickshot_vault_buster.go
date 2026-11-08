package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Slickshot Vault-Buster — Creature — Human Rogue {2}{U}, 1/4 (Outlaws
// of Thunder Junction, #4468):
//
//	"Vigilance
//	 This creature gets +2/+0 as long as you've committed a crime this
//	 turn. (Targeting opponents, anything they control, and/or cards in
//	 their graveyards is a crime.)"
//
// Omenport Vigilante's condition on a layer 7c modification (CR 613.4c)
// of the Vault-Buster alone: read live (CR 611.3a) off this turn's
// event log (YouCommittedACrimeThisTurn), it switches on with the
// controller's first crime, stays on for the rest of the turn even if
// that spell is countered, and is off again on the next turn.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "2085642e-29be-4e13-81dc-1f2ff6c095eb",
		Name:            "Slickshot Vault-Buster",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Static: []game.StaticAbility{{
			Label:    "Gets +2/+0 as long as you've committed a crime this turn.",
			Layer:    game.Layer7PT,
			SubLayer: game.SubLayer7C_Modify,
			AppliesTo: func(target *game.Card, g *game.Game, source *game.Card) bool {
				return selfCreature(target, g, source) && YouCommittedACrimeThisTurn(g, source.Controller)
			},
			Apply: pumpBy(2, 0),
		}},
	})
}
