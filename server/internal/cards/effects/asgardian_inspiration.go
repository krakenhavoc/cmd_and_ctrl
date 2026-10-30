package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Asgardian Inspiration — Sorcery {R} (Marvel Super Heroes Commander, #4455):
//
//	"Exile the top card of your library. You may play it this turn.
//	 Whenever a source you control deals noncombat damage to an opponent, you
//	 may pay {2}. If you do, return this card from your graveyard to your
//	 hand."
//
// The spell is the impulse exile (exileTopCardYouMayPlayThisTurn). The second
// ability works from the graveyard (TriggeredAbility.Zones): it fires for each
// damage event of a source the card's owner controls, noncombat, dealt to an
// opponent of theirs. The {2} is asked at resolution, and the card is returned
// only if it is still the same graveyard object (CR 400.7), Silvan Reveler's
// pay-then-return.
//
// No simplification.
func init() {
	t := InGraveyard(On(game.EventDealDamage, msc16NoncombatDamageToAnOpponentFromYourSource,
		"Asgardian Inspiration — you may pay {2} to return this card from your graveyard to your hand",
		func(g *game.Game, item *game.StackItem) error {
			return mayPayForGraveyardReturnToHand(g, item, "{2}",
				"Asgardian Inspiration — pay {2} to return it from your graveyard to your hand?")
		}))
	Register(Spec{
		OracleID:     "1be4bb28-6ad7-4df9-98b7-fa32061e7a95",
		Name:         "Asgardian Inspiration",
		Completeness: CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return exileTopCardYouMayPlayThisTurn(item, ctx, 0)
		},
		Triggered: []game.TriggeredAbility{t},
	})
}

// msc16NoncombatDamageToAnOpponentFromYourSource is "a source you control
// deals noncombat damage to an opponent": positive noncombat damage dealt to a
// player other than the card's owner by a source its owner controls (an
// ability's source counts, one that has left is read as it was).
func msc16NoncombatDamageToAnOpponentFromYourSource(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.Kind != game.EventDealDamage || ev.Combat || ev.Amount <= 0 {
		return false
	}
	if g.PlayerByIDForEffect(ev.Target) == nil || ev.Target == source.Owner {
		return false
	}
	controller, ok := b34DamageSourceController(ev, g)
	return ok && controller == source.Owner
}
