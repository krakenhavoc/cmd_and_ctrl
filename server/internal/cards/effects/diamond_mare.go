package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Diamond Mare — Artifact Creature — Horse {2}, 1/3 (Foundations):
//
//	"As this creature enters, choose a color.
//	 Whenever you cast a spell of the chosen color, you gain 1 life."
//
// The colour is the stored "as this enters" choice (ChooseColorAsEnters)
// and the trigger reads it off the Mare itself. Before a colour is
// chosen, or when the spell is colourless, nothing triggers. A spell's
// colour is its own (CR 202.2); it is not a word of the Mare's text.
func init() {
	Register(Spec{
		OracleID:     "9a440122-a015-4af2-b270-37f019884458",
		Name:         "Diamond Mare",
		Completeness: CompletenessFull,
		AsEnters:     ChooseColorAsEnters(game.ColorForBenefit, "Diamond Mare"),
		Triggered: []game.TriggeredAbility{
			On(game.EventCast, youCastASpellOfTheChosenColor,
				"Diamond Mare — you gain 1 life", Do(GainLife{Amount: 1})),
		},
	})
}

// youCastASpellOfTheChosenColor — you cast a spell whose colours include
// the colour stored on the source ("Whenever you cast a spell of the
// chosen color").
func youCastASpellOfTheChosenColor(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.Kind != game.EventCast || ev.Actor != source.Controller || source.ChosenColor == "" {
		return false
	}
	spell, ok := g.LookupCardForEffect(ev.CardID)
	return ok && spell.HasColor(source.ChosenColor)
}
