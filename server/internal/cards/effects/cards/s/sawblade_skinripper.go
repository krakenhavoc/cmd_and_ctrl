package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sawblade Skinripper — Creature — Human Assassin {1}{B}{R}, 3/2
// (Duskmourn, #4466):
//
//	"Menace
//	 {2}, Sacrifice another creature or enchantment: Put a +1/+1 counter
//	 on this creature.
//	 At the beginning of your end step, if you sacrificed one or more
//	 permanents this turn, this creature deals that much damage to any
//	 target."
//
// The sacrifice is a cost (b23SacrificeAnotherCreatureOrEnchantment). The
// end-step trigger is an intervening "if" (CR 603.4) on the turn tally of
// permanents its controller sacrificed (Phoenix Fleet Airship's read: a
// token, a Treasure or a creature all count, and one an opponent
// sacrificed does not), checked as the end step begins, so nothing
// sacrificed in the end step itself can raise it (the ruling), and again as
// it resolves. "That much" is that tally as the trigger resolves, so a
// permanent sacrificed in response counts, and the Skinripper is the
// damage's source as it last existed if it has left (CR 608.2h).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "96fd3ee0-84fb-439b-9e2d-b55afe6850a7",
		Name:            "Sawblade Skinripper",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"menace"},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.On(game.EventBeginEndStep, effects.AllOf(effects.ByYou, func(_ game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return g.TurnTallyFor(source.Controller).PermanentsSacrificed > 0
			}), "Sawblade Skinripper — damage to any target equal to the permanents you sacrificed this turn",
				effects.SourceDealsComputedDamageToEachLegalTarget(effects.B23PermanentsYouSacrificedThisTurn)),
				effects.TargetAny()),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}, Sacrifice another creature or enchantment: Put a +1/+1 counter on this creature.",
			Purpose: game.Purpose{Answers: game.AnswerSacOutlet},
			Cost:    effects.Plus(effects.ManaCost("{2}"), effects.B23SacrificeAnotherCreatureOrEnchantment()),
			Effect:  effects.Do(effects.CounterOnThis{Kind: game.CounterPlusOne, N: 1}),
		}},
	})
}
