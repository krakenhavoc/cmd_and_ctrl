package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Druid of the Spade — Creature — Rabbit Druid {2}{G}, 2/3 (Bloomburrow,
// #4467):
//
//	"As long as you control a token, this creature gets +2/+0 and has
//	 trample."
//
// Goblin Tomb Raider's shape: one condition gating two statics on the
// Druid alone, a layer 7c +2/+0 (CR 613.4c) and a layer 6 trample grant
// (CR 613.1f). "You control a token" reads the battlefield
// (selfWhileYouControlA over IsTokenPredicate), any token permanent, so
// both come with the first token and go with the last.
//
// No simplification.
func init() {
	whileToken := selfWhileYouControlA(IsTokenPredicate())
	trample := KeywordGrant(whileToken, "trample")
	trample.Label = "Has trample as long as you control a token."
	Register(Spec{
		OracleID:     "9740ce6d-ef6b-4d70-839a-8bd5e0e1e0a1",
		Name:         "Druid of the Spade",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			{
				Label:     "Gets +2/+0 as long as you control a token.",
				Layer:     game.Layer7PT,
				SubLayer:  game.SubLayer7C_Modify,
				AppliesTo: whileToken,
				Apply:     pumpBy(2, 0),
			},
			trample,
		},
	})
}
