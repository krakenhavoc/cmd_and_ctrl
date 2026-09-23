package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Bolt Bend — Instant {3}{R}:
//
//	"This spell costs {3} less to cast if you control a creature with
//	 power 4 or greater.
//	 Change the target of target spell or ability with a single
//	 target."
//
// Deflecting Swat's poor relation, and the other half of the CR 115.7
// family: where the Swat says "choose new targets" (CR 115.7c, every
// slot may move), this says "change THE target … with a single
// target" (CR 115.7b) — exactly one slot, on a spell that has exactly
// one. The printed "with a single target" is the clause predicate
// (HasASingleTarget), so a two-target spell is not offered in the
// picker at all rather than refused after the click.
//
// The change is MANDATORY. There is no "you may" printed, so the
// prompt's minimum is one and declining is not an answer — but a
// target with nowhere else legal to go is unchanged all the same, and
// the prompt is never opened for it (CR 115.7a).
//
// The cost reduction is a plain SelfCostModifiers clause: {3} less
// while a creature you control has power 4 or greater, which makes
// the card a one-mana redirect in the deck that wants it and a
// four-mana one everywhere else. Power is read post-layer
// (CurrentPower), so a pumped 2/2 switches it on.
//
// DECLARED CAVEAT — SPELLS ONLY, the same one Deflecting Swat ships:
// the engine cannot target an ABILITY on the stack (ADR 0065's open
// item), so "or ability" is out. Strictly narrower than printed
// (#259).
func init() {
	Register(Spec{
		OracleID:     "c20a96f7-aa5a-4c15-b8b1-806685c99b27",
		Name:         "Bolt Bend",
		Completeness: CompletenessCaveats,
		Caveats: []string{
			"Only a SPELL can be chosen, not an activated or triggered ability on the stack — the engine cannot target an ability item.",
		},
		Targets: TargetSpell("target spell with a single target", HasASingleTarget()),
		SelfCostModifiers: []game.CostModifier{
			CostsLess(3, "This spell costs {3} less to cast if you control a creature with power 4 or greater.",
				youControlAPowerFourCreature),
		},
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
				return nil
			}
			return ChangeTargets{
				StackID: item.Targets[0].ID,
				Policy:  game.RetargetChangeOne,
				Reason:  "Bolt Bend — change the target",
			}.Apply(ctx)
		},
	})
}

// youControlAPowerFourCreature is Bolt Bend's discount condition. The
// SPELL ITSELF is skipped for the reason every self-cost modifier
// skips it: it is on the stack while the price is being worked out,
// and a creature spell would otherwise count its own printed power.
func youControlAPowerFourCreature(q game.CostQuery) bool {
	if q.Game == nil || q.Game.Battlefield == nil {
		return false
	}
	for _, c := range q.Game.Battlefield.Cards {
		if c.InstanceID == q.Card.InstanceID || c.Controller != q.Controller || !c.IsCreature() {
			continue
		}
		if c.CurrentPower() >= 4 {
			return true
		}
	}
	return false
}
