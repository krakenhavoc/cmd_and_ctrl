package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Battle Cry Goblin — Creature — Goblin {1}{R}, 2/2 (Adventures in the
// Forgotten Realms):
//
//	"{1}{R}: Goblins you control get +1/+0 and gain haste until end of turn.
//	 Pack tactics — Whenever this creature attacks, if you attacked with
//	 creatures with total power 6 or greater this combat, create a 1/1 red
//	 Goblin creature token that's tapped and attacking."
//
// The Goblins are the ones you control as the ability resolves (CR 611.2c). The
// token is Adeline's: it enters tapped and attacking the player the Goblin
// attacked, and was never declared as an attacker (CR 508.4), so it is not
// counted by pack tactics.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e74c4001-d958-4d2b-b57d-c3601f01580c",
		Name:         "Battle Cry Goblin",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.S82AfrB09PackTactics("Battle Cry Goblin — pack tactics: a tapped and attacking Goblin", func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				tmpl := effects.TokenCard("1/1 red Goblin")
				tmpl.Tapped = true
				tmpl.AttackingTarget = ctx.Trigger().Event.Target
				return effects.CreateToken{Controller: item.Controller, Template: tmpl, N: 1}.Apply(ctx)
			}),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{R}: Goblins you control get +1/+0 and gain haste until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    effects.ManaCost("{1}{R}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				goblins := effects.And(effects.Creature(), effects.YouControl(), effects.Subtype("Goblin"))
				if err := (effects.BoostUntilEOT{Match: goblins, Power: 1, Label: "Battle Cry Goblin — Goblins get +1/+0 until end of turn"}).Apply(ctx); err != nil {
					return err
				}
				return effects.GrantKeywordUntilEOT{Match: goblins, Keywords: []string{"haste"},
					Label: "Battle Cry Goblin — Goblins gain haste until end of turn"}.Apply(ctx)
			},
		}},
	})
}
