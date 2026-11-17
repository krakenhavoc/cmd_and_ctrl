package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// wheneverYouDiscardOneOrMoreArtifactCardsOnce is "Whenever you discard
// one or more artifact cards, <effect>. This ability triggers only once
// each turn." (Urza, Powerstone Prodigy; Mishra, Excavation Prodigy).
//
// The engine emits one discard event per card, so the cards discarded
// together by one effect are one batch (OncePerBatch), and the
// once-each-turn guard declines every later batch of the turn. The
// card is read where it landed (an artifact card in the graveyard).
func wheneverYouDiscardOneOrMoreArtifactCardsOnce(label string, effect Effect) game.TriggeredAbility {
	return TriggersOnlyOnceEachTurn(OncePerBatch(
		On(game.EventDiscardCard, youDiscardedA(game.Card.IsArtifact), label, effect)))
}

// nontokenArtifactYouControlWasPutIntoAGraveyardOrExileFromTheBattlefield is
// Slagstone Refinery's condition: "this artifact or another nontoken
// artifact you control is put into a graveyard from the battlefield or is
// put into exile from the battlefield".
//
// A leaves-the-battlefield ability looks back in time (CR 603.10a), so the
// permanent's type and controller are the ones it last had on the
// battlefield (leftAsType, leftUnderControlOf), not whatever the card is in
// the graveyard or exile. A bounce or a library tuck is neither zone. The
// source itself counts even as a token copy ("this artifact"); another
// artifact must be nontoken.
func nontokenArtifactYouControlWasPutIntoAGraveyardOrExileFromTheBattlefield(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.Kind != game.EventLTB || (ev.NewZone != game.ZoneGraveyard && ev.NewZone != game.ZoneExile) {
		return false
	}
	c, ok := g.LookupCardForEffect(ev.CardID)
	if !ok || !leftAsType(ev, c, "artifact") || leftUnderControlOf(ev, c) != source.Controller {
		return false
	}
	return ev.CardID == source.InstanceID || !c.IsToken()
}

// CardPlayedFromExileThisTurn reports whether `you` played a card from
// exile this turn: cast a spell from exile or played a land from exile
// (game.PlayerTurnTally.CardsPlayedFromExile, #5482). Caller must hold
// g.mu.
func CardPlayedFromExileThisTurn(g *game.Game, you uuid.UUID) bool {
	return you != uuid.Nil && g.TurnTallyFor(you).CardsPlayedFromExile > 0
}

// CondNoCardPlayedFromExile is "if you didn't play a card from exile this
// turn" as an intervening-if Condition (CR 603.4): read for the item's
// controller as the ability resolves, with a footprint condition since it
// can only become false and no step casts a spell or plays a land.
type CondNoCardPlayedFromExile struct{}

func (CondNoCardPlayedFromExile) Holds(ctx *Context) bool {
	return !CardPlayedFromExileThisTurn(ctx.Game, ctx.Controller())
}

func (CondNoCardPlayedFromExile) condition() (game.FootprintStep, bool) {
	return condStep(game.FootprintCondNoCardsPlayedFromExile, 0), true
}

// AtYourEndStepIfNoCardPlayedFromExile is "At the beginning of your end
// step, if you didn't play a card from exile this turn, …", checked as the
// end step begins and again as the ability resolves.
func AtYourEndStepIfNoCardPlayedFromExile(label string, effect Effect) game.TriggeredAbility {
	return InterveningIf(CondNoCardPlayedFromExile{},
		On(game.EventBeginEndStep, AllOf(ByYou, func(_ game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
			return !CardPlayedFromExileThisTurn(g, source.Controller)
		}), label, effect))
}

// loseXLife is "you lose X life" off the resolving spell's announced X
// (Terisiare's Devastation). An X of 0 loses nothing: GainLife declines
// a zero amount.
func loseXLife(ctx *Context) error { return GainLife{Amount: -ctx.X()}.Apply(ctx) }
