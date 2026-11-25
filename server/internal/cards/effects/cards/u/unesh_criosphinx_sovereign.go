package cardsu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Unesh, Criosphinx Sovereign — {4}{U}{U} Legendary Creature — Sphinx 4/4 (Hour of Devastation):
//
//	"Flying
//	 Sphinx spells you cast cost {2} less to cast.
//	 Whenever Unesh or another Sphinx you control enters, reveal the top four
//	 cards of your library. An opponent separates those cards into two piles.
//	 Put one pile into your hand and the other into your graveyard."
//
// Each Sphinx entering is its own trigger. The piles are Fact or Fiction's
// split, the opponent chosen by the controller.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "7492ea06-24d5-4a1f-a6ac-c33743e91aa7",
		Name:            "Unesh, Criosphinx Sovereign",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Purpose:         game.Purpose{Answers: game.AnswerValue},
		CostModifiers: []game.CostModifier{
			effects.CostsLess(2, "Sphinx spells you cast cost {2} less to cast.",
				effects.YourSpell(), effects.SpellOfCreatureType("Sphinx")),
		},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventETB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				if ev.Kind != game.EventETB {
					return false
				}
				if ev.CardID == source.InstanceID {
					return true
				}
				c, ok := effects.EnteredUnderYourControl(ev, source, g, true)
				return ok && c.HasSubtype(source.TextSubtype("Sphinx"))
			}, "Unesh, Criosphinx Sovereign — reveal the top four cards; an opponent separates them into two piles; one pile to your hand, the other to your graveyard",
				func(g *game.Game, item *game.StackItem) error {
					return effects.OldB240FactOrFictionReveal(effects.NewContext(g, item), "Unesh, Criosphinx Sovereign", 4)
				}),
		},
	})
}
