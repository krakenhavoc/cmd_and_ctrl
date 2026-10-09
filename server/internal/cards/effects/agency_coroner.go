package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Agency Coroner — Creature — Ogre Cleric {4}{B}:
//
//	"{2}{B}, Sacrifice another creature: Draw a card. If the sacrificed
//	 creature was suspected, draw two cards instead."
//
// "Was suspected" is a fact about the creature as it left, and the
// designation is cleared the moment it does (CR 400.7), so the effect
// reads the cost's record of it — the sacrificed permanent's last-known
// information (CR 608.2h), which carries Suspected — rather than the
// card now in the graveyard.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "3686fd97-49ff-4a92-9cb2-7d8e9438d945",
		Name:         "Agency Coroner",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label: "{2}{B}, Sacrifice another creature: Draw a card. If the sacrificed creature was suspected, draw two cards instead.",
			Cost:  Plus(ManaCost("{2}{B}"), SacrificeAnotherN(1, "another creature", Creature())),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				n := 1
				if info, ok := ctx.SacrificedPermanent(); ok && info.Suspected {
					n = 2
				}
				return DrawCards{Player: item.Controller, N: n}.Apply(ctx)
			},
		}},
	})
}
