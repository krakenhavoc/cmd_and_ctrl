package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Bramblefort Fink — Creature — Ouphe {1}{G}, 2/2 (Throne of Eldraine, S82 old
// sets):
//
//	"{8}: This creature has base power and toughness 10/10 until end of turn.
//	 Activate only if you control an Oko planeswalker."
//
// Minsc's base-P/T-until-end-of-turn (layer 7b, CR 613.4b: counters and pumps
// still apply on top), on the Fink only while it is still the permanent that
// activated. The activation is refused unless its controller controls an Oko
// planeswalker (a planeswalker with the subtype Oko), checked as it is
// announced and not again as it resolves, as printed.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e56a9b7b-ce3f-4f5c-87d0-6f1d50d98ac3",
		Name:         "Bramblefort Fink",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:     "{8}: This creature has base power and toughness 10/10 until end of turn. Activate only if you control an Oko planeswalker.",
			Purpose:   game.Purpose{Answers: game.AnswerValue},
			Cost:      effects.ManaCost("{8}"),
			Condition: effects.OldB148YouControlAnOkoPlaneswalker,
			Effect: func(g *game.Game, item *game.StackItem) error {
				if !effects.SourceIsStillThisPermanent(g, item) {
					return nil
				}
				return effects.UntilEndOfTurn(effects.NewContext(g, item), item.SourceCardID, nil,
					"Bramblefort Fink — base power and toughness 10/10 until end of turn",
					game.SetBasePowerMod(10), game.SetBaseToughnessMod(10))
			},
		}},
	})
}
