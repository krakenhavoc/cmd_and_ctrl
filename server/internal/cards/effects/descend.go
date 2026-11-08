package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

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

// WhenThisEntersIfDescend is "Descend N — When this creature enters, if
// there are N or more permanent cards in your graveyard, …" (Coati
// Scavenger, Council of Echoes): checked as it enters and again as the
// trigger resolves (CR 603.4).
func WhenThisEntersIfDescend(n int, label string, effect Effect) game.TriggeredAbility {
	return InterveningIf(CondDescend{N: n}, On(game.EventETB, AllOf(Self, DescendHolds(n)), label, effect))
}

// CondDescended is "if you descended this turn" (CR 700.11: a permanent
// card was put into your graveyard from anywhere this turn) as an
// intervening-if Condition, read for the resolving ability's controller.
// It declares a read of the controller's graveyard
// (game.FootprintCondGraveyard): the answer only changes when a card is
// put into that graveyard, and the check conflicts that read with every
// item that moves a card between zones or changes a permanent, which
// covers every way one can get there.
type CondDescended struct{}

func (CondDescended) Holds(ctx *Context) bool { return descendedThisTurn(ctx.Game, ctx.Controller()) }

func (CondDescended) condition() (game.FootprintStep, bool) {
	return condStep(game.FootprintCondGraveyard, 0), true
}

// YouDescendedThisTurn is the When for "if you descended this turn", read
// for the source's controller as the event happens, for composing with
// AllOf.
func YouDescendedThisTurn(_ game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	return descendedThisTurn(g, source.Controller)
}

// AtYourEndStepIfDescended is "At the beginning of your end step, if you
// descended this turn, …" (Deep Goblin Skulltaker, Enterprising
// Scallywag): an intervening if (CR 603.4), checked as the end step
// begins and again as the trigger resolves.
func AtYourEndStepIfDescended(label string, effect Effect) game.TriggeredAbility {
	return InterveningIf(CondDescended{}, On(game.EventBeginEndStep, AllOf(ByYou, YouDescendedThisTurn), label, effect))
}

// descendHas is "there are N or more permanent cards in your graveyard"
// for a player, the shape whileYouHave reads.
func descendHas(n int) func(g *game.Game, you uuid.UUID) bool {
	return func(g *game.Game, you uuid.UUID) bool { return permanentCardsInGraveyard(g, you) >= n }
}

// SelfPumpWhileDescend is "Descend N — This creature gets +p/+t as long
// as there are N or more permanent cards in your graveyard" (Frilled
// Cave-Wurm, Echo of Dusk): a layer 7c modification (CR 613.4c) of the
// source alone. The graveyard is read on every recompute (CR 611.3a); a
// card entering or leaving it is a zone move, which drops the layer
// cache.
func SelfPumpWhileDescend(n, power, toughness int) game.StaticAbility {
	return game.StaticAbility{
		Label:     "Gets " + signedPT(power, toughness) + " as long as there are " + numberWord(n) + " or more permanent cards in your graveyard.",
		Layer:     game.Layer7PT,
		SubLayer:  game.SubLayer7C_Modify,
		AppliesTo: whileYouHave(descendHas(n), selfCreature),
		Apply:     pumpBy(power, toughness),
	}
}

// SelfKeywordWhileDescend is "Descend N — This creature has <keyword> as
// long as there are N or more permanent cards in your graveyard" (Didact
// Echo's flying, Echo of Dusk's lifelink): a layer 6 grant (CR 613.1f) on
// the same live condition.
func SelfKeywordWhileDescend(n int, keyword string) game.StaticAbility {
	ab := KeywordGrant(whileYouHave(descendHas(n), selfCreature), keyword)
	ab.Label = "Has " + keyword + " as long as there are " + numberWord(n) + " or more permanent cards in your graveyard."
	return ab
}
