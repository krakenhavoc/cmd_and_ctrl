package cardse

import (
	"strings"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Energy Vortex — Enchantment {3}{U}{U} (Mirage, S82 OLD b805):
//
//	"As this enchantment enters, choose an opponent.
//	 At the beginning of your upkeep, remove all vortex counters from this
//	 enchantment.
//	 At the beginning of the chosen player's upkeep, this enchantment deals 3
//	 damage to that player unless they pay {1} for each vortex counter on this
//	 enchantment.
//	 {X}: Put X vortex counters on this enchantment. Activate only during
//	 your upkeep."
//
// The choice is made as it enters (CR 614.12) and stored on the permanent
// (Black Vise). The chosen player's upkeep asks that player to pay as the
// trigger resolves (CR 118.12), the cost read then; with no counters the
// payment is {0} and the damage is simply dealt. Declining, or being unable
// to pay, deals the damage from the enchantment as it last existed.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "bcb2f489-9d58-424d-96c1-a124a8d25a14",
		Name:         "Energy Vortex",
		Completeness: effects.CompletenessFull,
		XMatters:     true,
		AsEnters:     effects.ChoosePlayerAsEnters("Energy Vortex", effects.Opponents),
		Triggered: []game.TriggeredAbility{
			effects.AtYourUpkeep("Energy Vortex — remove all vortex counters from this enchantment",
				effects.RemoveAllCountersFromThisEffect("vortex")),
			effects.On(game.EventBeginUpkeep, effects.TheChosenPlayersTurn,
				"Energy Vortex — 3 damage to the chosen player unless they pay {1} for each vortex counter",
				func(g *game.Game, item *game.StackItem) error {
					player := effects.TriggeringActor(item)
					if player == uuid.Nil {
						return nil
					}
					ctx := effects.NewContext(g, item)
					n := effects.TrekB5CountersOnThis(g, item.SourceCardID, "vortex")
					if n <= 0 {
						return effects.DealDamage{Source: item.SourceCardID, Target: player, Amount: 3}.Apply(ctx)
					}
					cost := strings.Repeat("{1}", n)
					return effects.PayUnless{
						Chooser:  player,
						Cost:     cost,
						Question: "Energy Vortex — pay " + cost + " or take 3 damage?",
						Else:     energyVortexDeclined,
					}.Apply(ctx)
				}),
		},
		Activated: []effects.ActivatedAbility{{
			Label:     "{X}: Put X vortex counters on this enchantment. Activate only during your upkeep.",
			Purpose:   game.Purpose{Answers: game.AnswerValue},
			Cost:      effects.ManaCost("{X}"),
			Condition: effects.DuringYourUpkeep(),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				return effects.AddCounter{Target: item.SourceCardID, Kind: "vortex", N: ctx.X()}.Apply(ctx)
			},
		}},
	})
}

// energyVortexDeclined is the unpaid branch: 3 damage to the player who was
// asked. An on-disk key: never renamed, never reused.
var energyVortexDeclined = effects.PayUnlessElse("pay-unless/s82-old-b805-energy-vortex-damage",
	func(ctx *effects.Context, r game.PayUnlessAnswer) error {
		return effects.DealDamage{Source: ctx.Item.SourceCardID, Target: r.Chooser, Amount: 3}.Apply(ctx)
	})
