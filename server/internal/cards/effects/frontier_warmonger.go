package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Frontier Warmonger — Creature — Human Warrior {3}{R}, 4/4:
//
//	"Whenever one or more creatures attack one of your opponents or a
//	 planeswalker they control, those creatures gain menace until end of
//	 turn."
//
// An attack is aimed at a player, a planeswalker or a battle; only the
// first two are "one of your opponents or a planeswalker they control"
// (a battle an opponent protects is neither, CR 508.3a), and the
// attackers may be anyone's, the controller's own included.
//
// The engine announces each declared attacker as its own event, so this is
// one trigger per attacker that meets the condition, each granting menace
// to that creature. The set "those creatures" is therefore exactly the
// creatures declared attacking a qualifying target, and a token put onto
// the battlefield attacking (never declared, CR 506.3c) is not one of
// them. The grant is a layer-6 effect on that one permanent until
// cleanup.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "9cb741ae-14d6-4a8d-bfad-703abfe3682c",
		Name:         "Frontier Warmonger",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventAttack, frontierWarmongerAttackedAnOpponentOrTheirWalker,
				"Frontier Warmonger — the attacking creature gains menace until end of turn",
				frontierWarmongerMenace),
		},
	})
}

// frontierWarmongerAttackedAnOpponentOrTheirWalker is "attack one of your
// opponents or a planeswalker they control".
func frontierWarmongerAttackedAnOpponentOrTheirWalker(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if source == nil || ev.Kind != game.EventAttack || ev.Target == uuid.Nil {
		return false
	}
	defender := b17DefendingPlayer(g, ev)
	if defender == uuid.Nil || defender == source.Controller {
		return false
	}
	if p := g.PlayerByIDForEffect(defender); p == nil || p.Eliminated {
		return false
	}
	return attacksYouOrYourPlaneswalker(g, ev.Target, defender)
}

// frontierWarmongerMenace gives the attacker menace until end of turn, if it
// is still on the battlefield.
func frontierWarmongerMenace(g *game.Game, item *game.StackItem) error {
	attacker := item.Trigger.Event.CardID
	if !onBattlefield(g, attacker) {
		return nil
	}
	return GrantKeywordUntilEOT{Target: attacker, Keywords: []string{"menace"}, Label: "Frontier Warmonger — menace"}.Apply(NewContext(g, item))
}
