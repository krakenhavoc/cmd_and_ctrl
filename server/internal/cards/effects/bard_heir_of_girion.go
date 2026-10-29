package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Bard, Heir of Girion — Legendary Creature — Human Archer {2}{W}{U}, 4/4
// (The Hobbit):
//
//	"Reach, vigilance
//	 Other creatures you control get +1/+1.
//	 Whenever you attack, draw a card."
//
// Reach and vigilance are printed keywords; the anthem is Maja's
// (TribalAnthem with no tribe, others only, yours only). "Whenever you
// attack" is Adeline's: one trigger per attack declaration however many
// attackers are declared (OncePerBatch over attackDeclaredByYou), so a
// three-creature attack draws once, and an attack by an opponent does not
// trigger it.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "b0e88bd6-50a3-48d7-a15f-ec51f6bced0d",
		Name:            "Bard, Heir of Girion",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"reach", "vigilance"},
		Static: []game.StaticAbility{
			TribalAnthem(TribeFilter{Others: true, YoursOnly: true}, 1, 1),
		},
		Triggered: []game.TriggeredAbility{
			OncePerBatch(On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return attackDeclaredByYou(ev, source.Controller)
			}, "Bard, Heir of Girion — draw a card", Do(DrawCards{N: 1}))),
		},
	})
}
