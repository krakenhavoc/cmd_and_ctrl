package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Surrak, Elusive Hunter — Legendary Creature — Human Warrior {2}{G}, 4/3
// (Tarkir: Dragonstorm, #4463):
//
//	"This spell can't be countered.
//	 Trample
//	 Whenever a creature you control or a creature spell you control
//	 becomes the target of a spell or ability an opponent controls, draw a
//	 card."
//
// "Can't be countered" is the spec's CantBeCountered, which works on the
// stack (CR 113.6g). Trample is a printed keyword. The trigger is Leovold, Emissary of Trest's
// EventBecomesTarget watch, fired once per target slot as the targets are
// chosen (CR 115.3), so it resolves before the spell or ability that did
// the targeting. Its object is either a creature on the battlefield you
// control, or a creature spell on the stack you control (a counterspell
// pointed at one of your creature spells draws you a card, even though
// Surrak himself can't be countered). A creature card in your graveyard is
// neither, so targeting one draws nothing. Your own spells and abilities
// never trigger it. The draw is not optional.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "7a0fe73f-6e08-4c9a-80e1-c2ead6cebe06",
		Name:            "Surrak, Elusive Hunter",
		Completeness:    CompletenessFull,
		CantBeCountered: true,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{
			On(game.EventBecomesTarget, surrakTargetedByAnOpponent,
				"Surrak, Elusive Hunter — draw a card", Do(DrawCards{N: 1})),
		},
	})
}

// surrakTargetedByAnOpponent is "a creature you control or a creature
// spell you control becomes the target of a spell or ability an opponent
// controls".
func surrakTargetedByAnOpponent(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.Kind != game.EventBecomesTarget || source == nil || ev.CardID == uuid.Nil {
		return false
	}
	you := source.Controller
	if you == uuid.Nil || ev.Actor == uuid.Nil || ev.Actor == you || g.PlayerByIDForEffect(ev.Actor) == nil {
		return false
	}
	c, ok := g.LookupCardForEffect(ev.CardID)
	if !ok || !c.IsCreature() {
		return false
	}
	if onBattlefield(g, ev.CardID) {
		return c.Controller == you
	}
	spell := g.StackItemForEffect(ev.CardID)
	return spell != nil && spell.Kind == game.StackItemSpell && spell.Controller == you
}
