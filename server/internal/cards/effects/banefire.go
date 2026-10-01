package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Banefire — Sorcery {X}{R} (EDHREC rank 4080):
//
//	"Banefire deals X damage to any target.
//	 If X is 5 or more, this spell can't be countered and the damage
//	 can't be prevented."
//
// Red's finisher. In Commander it is the card a ramp deck holds until
// it can point forty mana at somebody, and the rider is what makes it
// a finisher rather than a Fireball: the player you are killing
// cannot Counterspell it and cannot fog it.
//
// The X damage is live and it is the whole body of the card: "any
// target" is the widest clause — a player, a creature, a planeswalker
// or a battle — and X is the announced value, charged on cast.
//
// BOTH RIDERS ARE THE SPELL'S OWN, UNDER ONE CONDITION (ADR 0107 §5).
// "If X is 5 or more" is judged off the spell's stack item, which holds
// the X the caster announced: Spec.CantBeCounteredIf when something
// tries to counter it, Spec.SpellDamageCantBePrevented as the damage is
// dealt. A Banefire for X=4 is an ordinary, counterable, preventable
// spell; one for X=5 shows both chips on the stack and goes through a
// Fog, a shield and protection alike (CR 615.12).
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:                   "5eff8a06-e0d6-435a-a7c0-db9f9d98636a",
		Name:                       "Banefire",
		XMatters:                   true,
		Completeness:               CompletenessFull,
		CantBeCounteredIf:          SpellXAtLeast(5),
		SpellDamageCantBePrevented: SpellXAtLeast(5),
		Targets:                    TargetAny(),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			x := ctx.X()
			if x <= 0 {
				return nil
			}
			for _, t := range ctx.LegalTargets() {
				return DealDamage{Source: item.SourceCardID, Target: t.ID, Amount: x}.Apply(ctx)
			}
			return nil
		},
	})
}
