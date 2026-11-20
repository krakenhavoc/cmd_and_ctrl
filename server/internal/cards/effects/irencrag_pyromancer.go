package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Irencrag Pyromancer — Creature — Human Wizard {2}{R}, 0/4 (Throne of Eldraine, #5516):
//
//	"Whenever you draw your second card each turn, this creature deals 3 damage to any target."
//
// Thopter Fabricator's "second card each turn" condition
// (b41DrewYourSecondCardThisTurn): it fires once a turn.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "5abcc262-fbe1-4e81-897c-da5a8fb1f1f4",
		Name:         "Irencrag Pyromancer",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(On(game.EventDrawCard, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return b41DrewYourSecondCardThisTurn(ev, source, g)
			}, "Irencrag Pyromancer — this creature deals 3 damage to any target",
				DamageToFirstTarget(3)), TargetAny()),
		},
	})
}
