package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Norin, Swift Survivalist — Legendary Creature — Human Coward {R}, 2/1
// (Duskmourn: House of Horror, #4466):
//
//	"Norin can't block.
//	 Whenever a creature you control becomes blocked, you may exile it. You
//	 may play that card from exile this turn."
//
// "Can't block" is Bloodghast's self restriction. "Becomes blocked" is
// EventBecomesBlocked, once per blocked attacker however many creatures
// block it (CR 509.3c); the trigger goes on the stack before combat damage,
// so a creature exiled by it deals and receives none (the ruling). The "you
// may exile it" is asked as the ability resolves (CR 603.5), and the exile
// is Airbend's (ExileWithPermission) with the grant given to YOU, Norin's
// controller, whoever owns the card: you may play it from exile this turn,
// paying its costs and following its timing (a creature card only in a main
// phase with the stack empty). The creature is exiled only if it is still
// that object on the battlefield (CR 400.7); a token that is exiled ceases
// to exist and cannot be played.
//
// No simplification.
func init() {
	const label = "Norin, Swift Survivalist — you may exile it; you may play that card from exile this turn"
	Register(Spec{
		OracleID:     "425d755b-7a15-423d-9640-494debc79dd9",
		Name:         "Norin, Swift Survivalist",
		Completeness: CompletenessFull,
		Static:       []game.StaticAbility{RestrictSelf(game.CantBlock)},
		Triggered: []game.TriggeredAbility{
			Optional(On(game.EventBecomesBlocked, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				c, ok := g.LookupCardForEffect(ev.CardID)
				return ok && c.IsCreature() && c.Controller == source.Controller
			}, label, norinExileIt), "Norin, Swift Survivalist — exile the blocked creature? You may play it from exile this turn."),
		},
	})
}

// norinExileIt exiles the blocked creature, if it is still that permanent on
// the battlefield, and lets Norin's controller play it this turn.
func norinExileIt(g *game.Game, item *game.StackItem) error {
	if item.Trigger == nil || item.Trigger.Event.CardID == uuid.Nil {
		return nil
	}
	id := item.Trigger.Event.CardID
	if !onBattlefield(g, id) {
		return nil
	}
	if o := item.Trigger.Object; o != nil && g.ObjectEpochForEffect(id) != o.Ref().Epoch {
		return nil
	}
	return ExileWithPermission{Target: id, GrantTo: item.Controller}.Apply(NewContext(g, item))
}
