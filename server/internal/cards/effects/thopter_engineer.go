package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Thopter Engineer — Creature — Human Artificer {2}{R}, 1/3 (Bloomburrow
// Commander):
//
//	"When this creature enters, create a 1/1 colorless Thopter artifact
//	 creature token with flying.
//	 Artifact creatures you control have haste."
//
// The token is the table's Thopter row. The grant is a layer 6 keyword grant to
// every artifact creature its controller controls, itself included when it is
// one, read through the layered view.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "4ac53b34-20d2-4a53-bade-4f998fb6ae67",
		Name:         "Thopter Engineer",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Tokens: 1},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Thopter Engineer — create a 1/1 Thopter with flying",
				Do(CreateToken{Template: TokenCard("1/1 colorless Thopter artifact with flying"), N: 1})),
		},
		Static: []game.StaticAbility{
			Labelled(KeywordGrant(func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.IsCreature() && target.IsArtifact() && target.Controller == source.Controller
			}, "haste"), "Artifact creatures you control have haste."),
		},
	})
}
