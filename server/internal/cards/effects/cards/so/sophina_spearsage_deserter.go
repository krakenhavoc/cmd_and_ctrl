package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sophina, Spearsage Deserter — Legendary Creature — Human Soldier {2}{R}{W},
// 4/4 (Secret Lair Drop):
//
//	"Menace
//	 Whenever Sophina, Spearsage Deserter attacks, investigate once for each
//	 nontoken attacking creature.
//	 Partner—Friends forever (You can have two commanders if both have this
//	 ability.)"
//
// Wojek Investigator's InvestigateForEach, counted as the trigger resolves over
// every nontoken creature attacking, whoever controls it (a creature put onto
// the battlefield attacking is one too). Partner—Friends forever is a
// deck-construction rule read off the oracle text.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "095dabda-4570-4764-8d62-4a29eb19d3e6",
		Name:            "Sophina, Spearsage Deserter",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"menace"},
		Purpose:         game.Purpose{Tokens: 1},
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisAttacks("Sophina, Spearsage Deserter — investigate once for each nontoken attacking creature",
				effects.Do(effects.InvestigateForEach{
					Name: "nontoken attacking creatures",
					Count: func(ctx *effects.Context) int {
						n := 0
						for _, c := range effects.B26AttackingCreatures(ctx.Game) {
							if !c.IsToken() {
								n++
							}
						}
						return n
					},
				})),
		},
	})
}
