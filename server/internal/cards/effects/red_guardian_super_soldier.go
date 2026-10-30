package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Red Guardian, Super-Soldier — Legendary Creature — Human Soldier
// Villain {2}{W}, 2/2 (Marvel Super Heroes, #4455):
//
//	"Flash
//	 When Red Guardian enters, destroy target creature an opponent controls
//	 that dealt damage this turn."
//
// Covert Cutpurse's targeted enter trigger with the other direction of
// damage: the target is the DEALER, to a player or to a permanent, combat
// or not. The record is the turn's damage events (Game.EventsThisTurn),
// each naming the damage actually dealt (after prevention) and its source
// object; a creature that entered after its damage is a new object that
// dealt none (CR 400.7). The clause is a target restriction, judged as the
// trigger goes on the stack (CR 603.3d) and again on resolution
// (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "e85567e8-1102-4d44-a15b-d5f1fdef8465",
		Name:            "Red Guardian, Super-Soldier",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Triggered: []game.TriggeredAbility{
			Targeting(WhenThisEnters("Red Guardian — destroy target creature an opponent controls that dealt damage this turn", destroyChosenPermanent),
				TargetCreature("target creature an opponent controls that dealt damage this turn", OpponentControls(), mshA6DealtDamageThisTurn())),
		},
	})
}

// mshA6DealtDamageThisTurn passes for a permanent that has been the source
// of damage this turn, to anything.
func mshA6DealtDamageThisTurn() CardPredicate {
	return func(g *game.Game, _ uuid.UUID, c game.Card) bool {
		dealt := false
		for _, ev := range g.EventsThisTurn() {
			switch {
			case ev.Kind == game.EventZoneMove && ev.CardID == c.InstanceID && ev.NewZone == game.ZoneBattlefield:
				dealt = false
			case ev.Kind == game.EventDealDamage && ev.Source == c.InstanceID && ev.Amount > 0:
				dealt = true
			}
		}
		return dealt
	}
}
