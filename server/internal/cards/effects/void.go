package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// void.go — "if a nonland permanent left the battlefield this turn or a
// spell was warped this turn" (void, CR 207.2c; #2128).
//
// One fact about the whole TABLE's turn, read off two cells of the turn
// tally (game.TurnTally.NonlandPermanentsLeft and SpellsWarped) through
// Game.VoidThisTurn: any nonland permanent leaving by any route counts,
// tokens included, whoever controlled it; a spell cast for its warp cost
// (CR 702.185c) counts whether or not it resolved. It is a fact about
// the TURN, not about the source: a Skittermaw that enters after the
// Treasure was sacrificed still sees it. Revolt (revolt.go) is the
// per-player sibling.
//
// Readers, one per place the printed text puts the condition:
//
//   - Void(g) for a spell's or ability's own resolution ("instead if",
//     "if ..., draw another card").
//   - WhenThisEntersIfVoid, WheneverThisAttacksIfVoid and
//     AtYourEndStepIfVoid are the three trigger shapes: an intervening
//     "if" (CR 603.4), checked as the trigger fires and again as it
//     resolves (InterveningIf with CondVoid).
//   - CostsLessIfVoid is the self cost modifier (Temporal Intervention).
//   - AtYourEndStepUnlessVoid is the end-step "do this unless void"
//     (Chorale of the Void), which is a plain resolution-time check, not
//     an intervening if.

// Void reports whether a nonland permanent left the battlefield this
// turn or a spell was warped this turn. Caller must hold g.mu (every
// effect body does).
func Void(g *game.Game) bool {
	return g.VoidThisTurn()
}

// CondVoid is "if a nonland permanent left the battlefield this turn or a
// spell was warped this turn" as an intervening-if Condition, for
// InterveningIf and If. It declares no footprint: the state it reads is a
// table-wide tally that no trigger-independence analysis models yet, so a
// row that carries it is simply not reordered.
type CondVoid struct{}

func (CondVoid) Holds(ctx *Context) bool { return Void(ctx.Game) }

func (CondVoid) condition() (game.FootprintStep, bool) { return game.FootprintStep{}, false }

// unlessVoid wraps an Effect so it runs only when void does NOT hold as
// it resolves.
func unlessVoid(effect Effect) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		if Void(g) {
			return nil
		}
		return effect(g, item)
	}
}

// VoidHolds is the When for "if a nonland permanent left the battlefield
// this turn or a spell was warped this turn", for composing with AllOf.
func VoidHolds(_ game.Event, _ *game.Card, _ game.Characteristic, g *game.Game) bool {
	return Void(g)
}

// WhenThisEntersIfVoid is "Void — When ~ enters, if a nonland permanent
// left the battlefield this turn or a spell was warped this turn, …".
// Set Targets on the returned ability for a targeted body.
func WhenThisEntersIfVoid(label string, effect Effect) game.TriggeredAbility {
	return InterveningIf(CondVoid{}, On(game.EventETB, AllOf(Self, VoidHolds), label, effect))
}

// WheneverThisAttacksIfVoid is "Void — Whenever ~ attacks, if a nonland
// permanent left the battlefield this turn or a spell was warped this
// turn, …".
func WheneverThisAttacksIfVoid(label string, effect Effect) game.TriggeredAbility {
	return InterveningIf(CondVoid{}, On(game.EventAttack, AllOf(ThisAttacked, VoidHolds), label, effect))
}

// AtYourEndStepIfVoid is "Void — At the beginning of your end step, if a
// nonland permanent left the battlefield this turn or a spell was warped
// this turn, …". It triggers once however many left. Set Targets on the
// returned ability for a targeted body.
func AtYourEndStepIfVoid(label string, effect Effect) game.TriggeredAbility {
	return InterveningIf(CondVoid{}, On(game.EventBeginEndStep, AllOf(ByYou, VoidHolds), label, effect))
}

// AtYourEndStepUnlessVoid is "Void — At the beginning of your end step,
// <do this> unless a nonland permanent left the battlefield this turn or
// a spell was warped this turn" (Chorale of the Void). Not an intervening
// if: the ability always triggers and the check is made as it resolves.
func AtYourEndStepUnlessVoid(label string, effect Effect) game.TriggeredAbility {
	return On(game.EventBeginEndStep, ByYou, label, unlessVoid(effect))
}

// youLoseLife is the step "you lose N life" (CR 119.3): a life change that
// names the item's source and deals no damage. A step, not a bare call, so
// the steps after it wait for any prompt it asks (CR 608.2c).
func youLoseLife(n int) Instruction {
	return func(ctx *Context) error {
		return ctx.Game.ChangePlayerLifeForEffect(ctx.Source(), ctx.Controller(), -n)
	}
}

// CostsLessIfVoid is the self cost modifier "This spell costs {n} less to
// cast if a nonland permanent left the battlefield this turn or a spell
// was warped this turn" (Temporal Intervention). Read at CR 601.2f, so
// the answer is the one standing as the spell is announced.
func CostsLessIfVoid(n int, label string) game.CostModifier {
	return CostsLess(n, label, func(q game.CostQuery) bool {
		return q.Game != nil && q.Game.VoidThisTurn()
	})
}
