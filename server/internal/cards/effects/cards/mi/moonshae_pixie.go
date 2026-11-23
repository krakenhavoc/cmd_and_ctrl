package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Moonshae Pixie // Pixie Dust — an Adventure creature (Commander Legends:
// Battle for Baldur's Gate, #5948):
//
//	Moonshae Pixie {3}{U}, Creature — Faerie, 2/2
//	  "Flying
//	   When this creature enters, draw cards equal to the number of opponents
//	   who were dealt combat damage this turn."
//	Pixie Dust {1}{U}, Instant — Adventure
//	  "Up to three target creatures gain flying until end of turn."
//
// The count is Tymna the Weaver's: opponents of the controller that any
// creature dealt combat damage to this turn (CR 120.2), read as the trigger
// resolves. Pixie Dust grants flying to each target still legal.
//
// No simplification.
func init() {
	const oracleID = "89f1aefd-feb6-45ce-a992-707d027714e7"
	effects.Register(effects.Spec{
		OracleID:        oracleID,
		Name:            "Moonshae Pixie",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Moonshae Pixie — draw cards equal to the number of opponents who were dealt combat damage this turn",
				func(g *game.Game, item *game.StackItem) error {
					n := 0
					for _, p := range g.Seats {
						if p != nil && p.ID != item.Controller && len(g.CreaturesThatDealtCombatDamageToThisTurn(p.ID)) > 0 {
							n++
						}
					}
					if n == 0 {
						return nil
					}
					return effects.DrawCards{Player: item.Controller, N: n}.Apply(effects.NewContext(g, item))
				}),
		},
	})
	effects.Register(effects.Spec{
		OracleID:     oracleID + "#1",
		Name:         "Pixie Dust",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("up to three target creatures").WithCount(0, 3),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			for _, t := range ctx.LegalTargets() {
				if t.Kind != game.TargetCard {
					continue
				}
				if err := (effects.GrantKeywordUntilEOT{Target: t.ID, Keywords: []string{"flying"}, Label: "Pixie Dust — flying"}).Apply(ctx); err != nil {
					return err
				}
			}
			return nil
		},
	})
}
