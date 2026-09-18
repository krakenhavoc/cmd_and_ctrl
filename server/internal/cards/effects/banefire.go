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
// TWO DECLARED SIMPLIFICATIONS, BOTH WEAKER THAN PRINTED, both in the
// conditional rider and neither in the damage:
//
//   - "CAN'T BE COUNTERED" IS NOT MODELLED. Spec.CantBeCountered is
//     an unconditional, boot-time flag — it is read off the catalog
//     entry before anything is announced, and there is no seam that
//     lets it depend on the X the caster just picked. Setting it
//     unconditionally would make a Banefire for X=1 uncounterable,
//     which is STRONGER than printed and is what the #259 rule
//     forbids. So it is left off entirely: a big Banefire can be
//     countered here, which is weaker and honest.
//   - "THE DAMAGE CAN'T BE PREVENTED" IS NOT MODELLED. Unpreventable
//     damage has no shape in the engine at all — the CR 615
//     prevention shield is a replacement and nothing yet suppresses
//     one — so a fog or a protective shield stops a big Banefire
//     here. Weaker than printed, again.
//
// Both riders come back together the day a cast-time-conditional
// static exists; the damage half needs nothing.
func init() {
	Register(Spec{
		OracleID:     "5eff8a06-e0d6-435a-a7c0-db9f9d98636a",
		Name:         "Banefire",
		Completeness: CompletenessCaveats,
		Caveats: []string{
			"A Banefire for X=5 or more can still be countered — the \"can't be countered\" rider isn't implemented.",
			"Damage prevention still stops it — the \"damage can't be prevented\" rider isn't implemented.",
		},
		Targets: TargetAny(),
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
