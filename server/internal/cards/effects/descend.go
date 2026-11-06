package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// descend.go — the shared shape of the descend ability words (CR 207.2c,
// "descend 4", "descend 8") on a triggered ability: "Descend N — Whenever
// <event>, if there are N or more permanent cards in your graveyard, …".
// The ability word has no rules meaning of its own; the "if" after the
// trigger event is an intervening if (CR 603.4), checked as the event
// happens and again as the ability resolves. The count is
// permanentCardsInGraveyard (CR 110.4), Wail of the Forgotten's.

// CondDescend is "if there are N or more permanent cards in your
// graveyard" as an intervening-if Condition, for InterveningIf and If.
// "Your" is the controller of the resolving ability. It declares no
// footprint: the graveyard's contents are not a state the
// trigger-independence analysis models, so a row that carries it is
// simply not reordered (CondVoid's reason).
type CondDescend struct{ N int }

func (c CondDescend) Holds(ctx *Context) bool {
	return permanentCardsInGraveyard(ctx.Game, ctx.Controller()) >= c.N
}

func (CondDescend) condition() (game.FootprintStep, bool) { return game.FootprintStep{}, false }

// DescendHolds is the When for "if there are N or more permanent cards in
// your graveyard", read for the source's controller as the event happens,
// for composing with AllOf.
func DescendHolds(n int) When {
	return func(_ game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
		return permanentCardsInGraveyard(g, source.Controller) >= n
	}
}

// WheneverThisAttacksIfDescend is "Descend N — Whenever ~ attacks, if
// there are N or more permanent cards in your graveyard, …" (Stinging Cave
// Crawler).
func WheneverThisAttacksIfDescend(n int, label string, effect Effect) game.TriggeredAbility {
	return InterveningIf(CondDescend{N: n}, On(game.EventAttack, AllOf(ThisAttacked, DescendHolds(n)), label, effect))
}
