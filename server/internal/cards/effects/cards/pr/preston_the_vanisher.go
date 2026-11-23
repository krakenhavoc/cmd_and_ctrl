package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Preston, the Vanisher — Legendary Creature — Rabbit Wizard {3}{W}, 2/5
// (Jumpstart 2022):
//
//	"Whenever another nontoken creature you control enters, if it wasn't
//	 cast, create a token that's a copy of that creature, except it's a 0/1
//	 white Illusion.
//	 {1}{W}, Sacrifice five Illusions: Exile target nonland permanent."
//
// The trigger is Rapid Augmenter's "another creature you control enters, if it
// wasn't cast" (the entry's own history) narrowed to nontoken creatures, so a
// token copy never begets another. The copy's exception (CR 707.9b) sets
// 0/1, white, and Illusion in place of its creature types, keeping the rest
// of its copiable values. The Illusions are a cost (CR 118), so they are gone
// before the exile resolves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "2e00ccc9-6d5f-4de3-bbec-48b49d363244",
		Name:         "Preston, the Vanisher",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventETB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				c, ok := effects.Cmdr2B21AnotherCreatureYouControlEntered(ev, source, g)
				return ok && !effects.IsToken(c) && effects.Cmdr2B21EnteredWithoutBeingCast(g, ev.CardID)
			}, "Preston, the Vanisher — create a token copy of that creature, except it's a 0/1 white Illusion",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					return effects.CreateTokenCopy{
						Controller: ctx.Controller(),
						Copy:       item.Trigger.Event.CardID,
						N:          1,
						Except: func(tok *game.Card) {
							tok.Power, tok.Toughness = 0, 1
							tok.VariableToughness = false
							tok.PrintedPTKnown = true
							tok.SetCopyExceptionColors("W")
							tok.TypeLine = effects.S82MidB11TypeLineWithCreatureType(tok.TypeLine, "Illusion")
						},
					}.Apply(ctx)
				}),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{W}, Sacrifice five Illusions: Exile target nonland permanent.",
			Purpose: game.Purpose{Answers: game.AnswerRemove | game.AnswerSacOutlet},
			Cost:    effects.Plus(effects.ManaCost("{1}{W}"), effects.SacrificeN(5, "five Illusions", effects.Subtype("Illusion"))),
			Targets: effects.TargetPermanent("target nonland permanent", effects.Nonland()),
			Effect:  effects.ExileFirstTarget,
		}},
	})
}
