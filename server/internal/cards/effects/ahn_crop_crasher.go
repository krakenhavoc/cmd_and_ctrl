package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ahn-Crop Crasher — Creature — Minotaur Warrior {2}{R}, 3/2:
//
//	"Haste
//	 You may exert this creature as it attacks. When you do, target
//	 creature can't block this turn. (An exerted creature won't untap
//	 during your next untap step.)"
//
// ADR 0130 §11: exert as it attacks (CR 701.43d) with a targeted linked
// trigger (CR 607.2h). Any creature is a legal target, not only an
// opponent's. The restriction is Clamor Shaman's: the target can't be
// declared as a blocker until the turn ends, and the trigger resolves
// before blockers. With no creature to target the trigger is removed
// and the Crasher stays exerted (CR 603.3d).
//
// No simplification.
func init() {
	const label = "Ahn-Crop Crasher — target creature can't block this turn"
	Register(Spec{
		OracleID:        "5b3ef91a-5804-43cc-a947-f25f33b7b88c",
		Name:            "Ahn-Crop Crasher",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"haste"},
		ExertOnAttack:   ExertAsItAttacks(),
		Triggered: []game.TriggeredAbility{
			Targeting(
				WhenExerted(label, func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					id, ok := b16FirstLegalTargetCard(ctx)
					if !ok {
						return nil
					}
					return RestrictUntilEOT{Target: id, Restrictions: game.CantBlock, Label: label}.Apply(ctx)
				}),
				TargetCreature("target creature")),
		},
	})
}
