package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Dragon's Hoard — Artifact {3} (EDHREC rank 1194):
//
//	"Whenever a Dragon you control enters, put a gold counter on this
//	 artifact.
//	 {T}, Remove a gold counter from this artifact: Draw a card.
//	 {T}: Add one mana of any color."
//
// The Dragon deck's three-mana rock. The trigger is the landfall
// shape with "Dragon" for "land", reading effective subtypes so a
// changeling counts; the mana ability is Commander's Sphere's
// five-way pick with no identity narrowing — "any color" is what the
// card says.
//
// The draw is a CR 602 activation whose cost removes a gold counter
// from the Hoard itself (RemoveCountersFromThis, #625). The counter
// comes off at announce with the tap, so a proliferate in response
// cannot bank a counter the ability has already spent — which is why
// the ability waited for a real cost component rather than shipping
// with the removal deferred to resolution (#259). It shares the tap
// with the mana ability, as in paper.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8cf77dc4-763b-41b7-a5da-0ef2734f08e6",
		Name:         "Dragon's Hoard",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventETB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				c, ok := effects.EnteredUnderYourControl(ev, source, g, false)
				// "Dragon" is a word of the source's text (CR 612.2, #3984).
				return ok && c.HasSubtype(source.TextSubtype("Dragon"))
			}, "Dragon's Hoard — put a gold counter on it", func(g *game.Game, item *game.StackItem) error {
				return effects.AddCounter{Target: item.SourceCardID, Kind: "gold", N: 1}.Apply(effects.NewContext(g, item))
			}),
		},
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{W|U|B|R|G}",
			Label:    "Add one mana of any color",
		}},
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}, Remove a gold counter from this artifact: Draw a card.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.Plus(effects.TapCost(), effects.RemoveCountersFromThis("gold", 1)),
			Effect:  effects.B27DrawOne,
		}},
	})
}
