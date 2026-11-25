package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Mizzium Tank — Artifact — Vehicle {1}{R}{R}, 3/2 (War of the Spark, #5520):
//
//	"Trample
//	 Whenever you cast a noncreature spell, this Vehicle becomes an artifact
//	 creature and gets +1/+1 until end of turn.
//	 Crew 1"
//
// Mole Module's crew ability, and a cast trigger on the noncreature spell
// predicate Sky Theater Strix uses: the Tank becomes an artifact creature
// (the crew effect) and gets +1/+1, both until end of turn, if it is still on
// the battlefield.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "60957bfd-2460-429a-a7a7-d99424946f83",
		Name:            "Mizzium Tank",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Activated: []effects.ActivatedAbility{{
			Label:   "Crew 1",
			Cost:    effects.CrewCost(1),
			Purpose: game.Purpose{Answers: game.AnswerAnimate},
			Effect:  effects.CrewEffect("Mizzium Tank"),
		}},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventCast, effects.B10NoncreatureSpellCastByYou,
				"Mizzium Tank — this Vehicle becomes an artifact creature and gets +1/+1 until end of turn",
				func(g *game.Game, item *game.StackItem) error {
					if !effects.OnBattlefield(g, item.SourceCardID) {
						return nil
					}
					ctx := effects.NewContext(g, item)
					if err := (effects.BecomeCreatureUntilEOT{Target: item.SourceCardID, Label: "Mizzium Tank — becomes an artifact creature"}).Apply(ctx); err != nil {
						return err
					}
					return effects.BoostUntilEOT{Target: item.SourceCardID, Power: 1, Toughness: 1, Label: "Mizzium Tank — +1/+1"}.Apply(ctx)
				}),
		},
	})
}
