package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Airbender's Reversal — Instant — Lesson {1}{W} (EDHREC rank 4306):
//
//	"Choose one —
//	 • Destroy target attacking creature.
//	 • Airbend target creature you control. (Exile it. While it's
//	   exiled, its owner may cast it for {2} rather than its mana
//	   cost.)"
//
// Two mana of combat trick that is a removal spell when you are being
// attacked and a rescue when you are the one losing a creature. The
// second mode is the interesting one: airbending your own creature in
// response to a removal spell saves it, and the {2} recast refunds
// most of what it cost you.
//
// Both modes target and only one may be chosen, which is the shape
// Register allows — per-mode target slots are unsupported for a spell
// that can choose several bullets, and "choose one" never does.
//
// "Target ATTACKING creature" is the combat-state read, not a
// keyword: a creature is attacking while it has an attack target
// stamped on it, which is true from the declaration until combat ends.
// Any attacking creature qualifies, including your own — attacking
// somebody else's creature is not a requirement the card states.
//
// Airbend grants the recast to the card's OWNER, not to the caster,
// which is why the second mode says "you control" rather than "target
// creature": used on your own creature the refund comes back to you,
// and the primitive would hand an opponent's creature back to them.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "64e3791b-de99-4eed-af49-cbec218c28aa",
		Name:         "Airbender's Reversal",
		Completeness: CompletenessFull,
		Modes: ChooseOne(
			Mode("Destroy target attacking creature.",
				TargetCreature("target attacking creature", b41Attacking())),
			Mode("Airbend target creature you control.",
				TargetCreature("target creature you control", YouControl())),
		),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			target, ok := b16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			if ctx.HasMode(0) {
				return DestroyTarget{Target: target}.Apply(ctx)
			}
			return Airbend{Target: target}.Apply(ctx)
		},
	})
}

// b41Attacking is "target attacking creature": a creature with an
// attack target stamped on it. The flag is set at declaration and
// cleared when combat ends, which is exactly the window in which a
// creature "is attacking" (CR 506.4).
//
// Kept beside the card because it is one card's clause.
func b41Attacking() CardPredicate {
	return func(_ *game.Game, _ uuid.UUID, c game.Card) bool {
		return c.AttackingTarget != uuid.Nil
	}
}
