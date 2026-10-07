package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Clockwork Droid — Artifact Creature — Robot {2}, 3/1:
//
//	"You may exert this creature as it attacks. When you do, it can't
//	 be blocked this turn and you scry 1. (An exerted creature won't
//	 untap during your next untap step. To scry 1, look at the top card
//	 of your library. You may put that card on the bottom.)"
//
// ADR 0130 §11: exert as it attacks (CR 701.43d) with a linked "when
// you do" (CR 607.2h). "Can't be blocked" is the restriction bit, set
// on this object until end of turn, before blockers; a Droid that
// left and came back is a new object and gets nothing (CR 400.7). The
// scry happens regardless.
//
// No simplification.
func init() {
	const label = "Clockwork Droid — it can't be blocked this turn and you scry 1"
	Register(Spec{
		OracleID:      "0aa875dc-89ae-4e7a-b10d-aadde8e2b95c",
		Name:          "Clockwork Droid",
		Completeness:  CompletenessFull,
		ExertOnAttack: ExertAsItAttacks(),
		Triggered: []game.TriggeredAbility{
			WhenExerted(label, func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				if err := (RestrictUntilEOT{Target: item.SourceCardID, Restrictions: game.CantBeBlocked, Label: label}).Apply(ctx); err != nil {
					return err
				}
				return Scry{Player: item.Controller, N: 1}.Apply(ctx)
			}),
		},
	})
}
