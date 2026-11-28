package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Minion of Tevesh Szat — Creature — Demon Minion {4}{B}{B}{B}, 4/4 (Ice Age, S82
// OLD batch 840, #5636):
//
//	"At the beginning of your upkeep, this creature deals 2 damage to you unless you
//	 pay {B}{B}.
//	 {T}: Target creature gets +3/-2 until end of turn."
//
// The upkeep is Fettergeist's pay-or-else: its controller is asked as the trigger
// resolves, and a decline has the Minion deal 2 damage to them. The boost is a
// layer 7c modification that ends at cleanup (CR 514.2), on any creature, the
// Minion included.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "93fbf4a7-4283-4eec-afde-6780149b56df",
		Name:         "Minion of Tevesh Szat",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.AtYourUpkeep("Minion of Tevesh Szat — 2 damage to you unless you pay {B}{B}",
				func(g *game.Game, item *game.StackItem) error {
					return effects.UpkeepPayUnless{
						Chooser:  item.Controller,
						Cost:     "{B}{B}",
						Question: "Minion of Tevesh Szat — pay {B}{B} or take 2 damage?",
						Else:     minionOfTeveshSzatDamage,
					}.Apply(effects.NewContext(g, item))
				}),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}: Target creature gets +3/-2 until end of turn.",
			Cost:    effects.TapCost(),
			Targets: effects.TargetCreature("target creature"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.BoostTheTargetUntilEOT(3, -2, "Minion of Tevesh Szat — +3/-2 until end of turn")(item, effects.NewContext(g, item))
			},
		}},
	})
}

// minionOfTeveshSzatDamage is the decline: the Minion deals 2 damage to the player
// who was asked. Keyed (#4264); an on-disk identity.
var minionOfTeveshSzatDamage = effects.PayUnlessElse("pay-unless/minion-of-tevesh-szat-damage", func(ctx *effects.Context, r game.PayUnlessAnswer) error {
	return effects.DealDamage{Source: r.Source, Target: r.Chooser, Amount: 2}.Apply(ctx)
})
