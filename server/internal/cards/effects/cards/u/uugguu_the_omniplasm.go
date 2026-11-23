package cardsu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Uugguu, the Omniplasm — Legendary Creature — Ooze {B}{G}{U}, 2/2
// (Mystery Booster Commander Edition):
//
//	"Whenever another nontoken Ooze creature you control dies, create two
//	 tokens that are copies of that creature, except they're 2/2 and they
//	 aren't legendary."
//
// Ratadrabik of Urborg's dies-then-copy shape over Hofri Ghostforge's
// "another nontoken creature you control" condition, plus an Ooze check
// read as the creature last existed on the battlefield (CR 603.10a). The
// copies are made from last-known information (CR 608.2h) and the
// exception clause is part of the copiable values (CR 707.9b): not
// legendary and a base 2/2.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "df9bb1fc-135d-4efb-b143-81950fc95a27",
		Name:         "Uugguu, the Omniplasm",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventLTB, uugguuAnotherNontokenOozeYouControlDied,
				"Uugguu, the Omniplasm — create two tokens that are copies of that creature, except they're 2/2 and they aren't legendary",
				func(g *game.Game, item *game.StackItem) error {
					if item.Trigger == nil {
						return nil
					}
					return effects.CreateTokenCopy{
						Controller: item.Controller,
						Copy:       item.Trigger.Event.CardID,
						N:          2,
						Except: func(t *game.Card) {
							t.TypeLine = effects.RemoveLegendaryFromTypeLine(t.TypeLine)
							t.Power = 2
							t.Toughness = 2
							t.VariableToughness = false
							t.PrintedPTKnown = true
						},
					}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}

// uugguuAnotherNontokenOozeYouControlDied is the death trigger's condition.
func uugguuAnotherNontokenOozeYouControlDied(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
	if !effects.HofriGhostforgeAnotherNontokenCreatureDied(ev, source, lki, g) {
		return false
	}
	dead, ok := g.LookupCardForEffect(ev.CardID)
	return ok && effects.LeftAsSubtype(ev, dead, source.TextSubtype("Ooze"))
}
