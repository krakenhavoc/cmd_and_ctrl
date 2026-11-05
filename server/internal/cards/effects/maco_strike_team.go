package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// MACO Strike Team — {G} Creature — Human Soldier, 2/1 (Star Trek,
// #4973):
//
//	"Whenever a creature you control becomes the target of a spell or
//	 ability an opponent controls, create a tapped 1/1 red Officer
//	 creature token."
//
// EventBecomesTarget fires once per target slot as the targets are chosen
// (CR 115.3, 601.2c), so a spell that targets two of your creatures makes
// two Officers, and the trigger goes on the stack above the spell that
// targeted. The creature is read through the layers as it is targeted;
// "an opponent" is any other seated player (Leovold's read).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "0d60cfb1-dd77-4013-b862-41e95fdb53ae",
		Name:         "MACO Strike Team",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventBecomesTarget, macoYourCreatureTargetedByAnOpponent,
				"MACO Strike Team — create a tapped 1/1 red Officer creature token",
				Do(CreateTokenAdvanced{Spec: TokenSpec{Template: TokenCard("1/1 red Officer"), Tapped: true}, N: 1})),
		},
	})
}

// macoYourCreatureTargetedByAnOpponent is "a creature you control becomes
// the target of a spell or ability an opponent controls".
func macoYourCreatureTargetedByAnOpponent(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.Kind != game.EventBecomesTarget || source == nil || ev.CardID == uuid.Nil {
		return false
	}
	you := source.Controller
	if you == uuid.Nil || ev.Actor == uuid.Nil || ev.Actor == you || g.PlayerByIDForEffect(ev.Actor) == nil {
		return false
	}
	if !onBattlefield(g, ev.CardID) {
		return false
	}
	c, ok := g.LookupCardForEffect(ev.CardID)
	return ok && c.Controller == you && c.IsCreature()
}
