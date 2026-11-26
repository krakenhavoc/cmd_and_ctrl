package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b134_helpers.go — Theros Beyond Death batch 134 (S82 OLD, #5514).

// CastTheNamedSpellForTheFirstTimeCondition is the event condition of "when you
// cast a spell with the chosen name for the first time this turn" (Medomai's
// Prophecy, chapter III): an EventCast by the trigger's controller whose spell
// has the name in CondParams.Name (every face, CR 201.2b), and which is the first
// such cast by that player this turn. The turn's earlier casts are read off the
// event log back to the turn's start, so a spell with that name cast before the
// chapter resolved, in the same turn, has used up the "first time" (CR 603.4).
var CastTheNamedSpellForTheFirstTimeCondition = game.DelayedCondition("cast/the-named-spell-first-time-this-turn", castTheNamedSpellForTheFirstTime)

// castTheNamedSpellForTheFirstTime is the condition's body.
func castTheNamedSpellForTheFirstTime(ev game.Event, dt *game.DelayedTrigger, g *game.Game, p game.EffectParams) bool {
	if ev.Kind != game.EventCast || ev.Actor != dt.Controller || p.Name == "" {
		return false
	}
	c, ok := g.LookupCardForEffect(ev.CardID)
	if !ok || !game.CardNameMatches(c, p.Name) {
		return false
	}
	for i := len(g.Events) - 1; i >= 0; i-- {
		prior := g.Events[i]
		if prior.Kind == game.EventTurnBegan {
			break
		}
		if prior.Seq >= ev.Seq || prior.Kind != game.EventCast || prior.Actor != dt.Controller {
			continue
		}
		if pc, ok := g.LookupCardForEffect(prior.CardID); ok && game.CardNameMatches(pc, p.Name) {
			return false
		}
	}
	return true
}

// DrawTwoBody is "draw two cards" as a delayed trigger's body.
var DrawTwoBody = game.SimpleDelayedBody("draw/two-cards-s82-old-b134", func(g *game.Game, item *game.StackItem) error {
	return DrawCards{Player: item.Controller, N: 2}.Apply(NewContext(g, item))
})

// WheneverYouCastFirstSpellOnAnOpponentsTurn is "whenever you cast your first
// spell during each opponent's turn, <effect>": the controller's cast on a turn
// that is not theirs which the per-turn cast tally counts as their first
// (B22FirstSpellOnAnOpponentsTurn). The caller adds the target clause, if any.
func WheneverYouCastFirstSpellOnAnOpponentsTurn(label string, effect Effect) game.TriggeredAbility {
	return On(game.EventCast, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
		return B22FirstSpellOnAnOpponentsTurn(ev, source, g)
	}, label, effect)
}

// PurphorossInterventionElemental is the X/1 red Elemental token with trample and
// haste of Purphoros's Intervention, sized X as it is created. The printed token
// card is */1; the template is 0/1 and the power is set here (the Horror of Profane
// Transfusion, S82OldB083HorrorToken, is the same shape). It returns the made ids.
func PurphorossInterventionElemental(ctx *Context, x int) ([]uuid.UUID, error) {
	tok := ctx.TextToken(TokenCard("0/1 red Elemental with trample and haste"))
	tok.Power = max(x, 0)
	return ctx.Game.CreateTokensForEffect(ctx.Controller(), tok, 1, game.TokenEntryOptions{})
}
