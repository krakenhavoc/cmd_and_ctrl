package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Vorinclex, Voice of Hunger — Legendary Creature — Phyrexian Praetor
// {6}{G}{G}, 7/6:
//
//	"Trample
//	 Whenever you tap a land for mana, add one mana of any type that
//	 land produced.
//	 Whenever an opponent taps a land for mana, that land doesn't untap
//	 during its controller's next untap step."
//
// Two different kinds of trigger, and the difference is the card:
//
//   - Yours is a TRIGGERED MANA ability (CR 605.1b): it fires on the tap,
//     adds mana and never uses the stack (CR 605.4a), so the extra mana
//     is in the pool before the spell it pays for is cast. It is Zendikar
//     Resurgent's and Mirari's Wake's ManaTrigger.
//   - The opponents' is an ordinary stack trigger (it adds no mana, so
//     CR 605.1b does not make it a mana ability) and is Manabarbs's shape
//     with a different consequence: when the opponent's land is tapped
//     for mana, a trigger goes on the stack that, as it resolves, marks
//     THAT land to skip its controller's next untap step.
//
// An opponent's land that is tapped for mana by their own mana ability
// that costs no {T} is not "tapped for mana" and does not fire it;
// b20LandTappedForMana asks the land to be tapped now, as Manabarbs does.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "dbf0ad03-ab31-49d2-89b1-05b45948a61f",
		Name:            "Vorinclex, Voice of Hunger",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample"},
		ManaTriggers: []game.ManaTrigger{
			WheneverYouTapALandForMana(
				"Vorinclex, Voice of Hunger — add one mana of any type that land produced",
				AddsOneManaOfAnyTypeProduced()),
		},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventManaAbilityActivated},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return b20LandTappedForMana(ev, g) && ev.Actor != source.Controller
			},
			Key: "Vorinclex, Voice of Hunger — that land doesn't untap during its controller's next untap step",
			Effect: func(g *game.Game, item *game.StackItem) error {
				land := item.Trigger.Event.Source
				if land == uuid.Nil {
					return nil
				}
				return DoesntUntapNextUntapStep{
					Targets: []uuid.UUID{land},
					Label:   "Vorinclex, Voice of Hunger",
				}.Apply(NewContext(g, item))
			},
		}},
	})
}
