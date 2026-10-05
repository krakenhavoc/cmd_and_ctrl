package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Great Train Heist — Instant {R}:
//
//	"Spree (Choose one or more additional costs.)
//	 + {2}{R} — Untap all creatures you control. If it's your combat
//	   phase, there is an additional combat phase after this phase.
//	 + {2} — Creatures you control get +1/+0 and gain first strike
//	   until end of turn.
//	 + {R} — Choose target opponent. Whenever a creature you control
//	   deals combat damage to that player this turn, create a tapped
//	   Treasure token."
//
// The third bullet is CR 603.7b's repeating delayed trigger (#2169): the
// spell sets it up, it fires once per creature that connects until
// cleanup, and it is not tied to the instant, which is in the graveyard
// by then. The chosen opponent rides in the condition's params, so a
// restore point written mid-turn keeps who it watches. The first bullet's
// extra combat applies only when it resolves in a combat phase of the
// caster's own turn.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "afe2f7a4-9440-4d93-801f-a18b627efb21",
		Name:         "Great Train Heist",
		Completeness: CompletenessFull,
		Modes: Spree(
			SpreeModeDoing("Untap all creatures you control. If it's your combat phase, there is an additional combat phase after this phase.", "{2}{R}", nil,
				func(_ *game.StackItem, ctx *Context, _ int) error {
					if err := (UntapAllCreaturesYouControl{}).Apply(ctx); err != nil {
						return err
					}
					g := ctx.Game
					if game.PhaseOf(g.Turn.Step) == game.PhaseCombat && activePlayerIDOf(g) == ctx.Controller() {
						return ExtraCombatAfterThisPhase().Apply(ctx)
					}
					return nil
				}),
			SpreeModeDoing("Creatures you control get +1/+0 and gain first strike until end of turn.", "{2}", nil,
				func(_ *game.StackItem, ctx *Context, _ int) error {
					yours := And(Creature(), YouControl())
					if err := (BoostUntilEOT{Match: yours, Power: 1, Label: "Great Train Heist — +1/+0"}).Apply(ctx); err != nil {
						return err
					}
					return GrantKeywordUntilEOT{Match: yours, Keywords: []string{"first strike"}, Label: "Great Train Heist — first strike"}.Apply(ctx)
				}),
			SpreeModeDoing("Choose target opponent. Whenever a creature you control deals combat damage to that player this turn, create a tapped Treasure token.", "{R}",
				TargetPlayer("target opponent", Opponent()),
				func(_ *game.StackItem, ctx *Context, occ int) error {
					t, ok := ModeTarget(ctx, occ)
					if !ok || t.Kind != game.TargetPlayer {
						return nil
					}
					return DelayedOnEvent{
						Label:      "Great Train Heist — create a tapped Treasure token",
						On:         []game.EventKind{game.EventDealDamage},
						Condition:  yourCreatureHitThePlayerCondition,
						CondParams: game.EffectParams{Player: t.ID},
						Body:       createTappedTreasureBody,
						Repeats:    true,
					}.Apply(ctx)
				}),
		),
	})
}

var (
	// "Whenever a creature you control deals combat damage to <Player>":
	// an EventDealDamage, combat, from a creature the trigger's
	// controller controls, to CondParams.Player.
	yourCreatureHitThePlayerCondition = game.DelayedCondition("combat/your-creature-hit-the-player", yourCreatureHitThePlayer)

	// "Create a tapped Treasure token."
	createTappedTreasureBody = game.SimpleDelayedBody("treasure/create-tapped", createTappedTreasure)
)

func yourCreatureHitThePlayer(ev game.Event, dt *game.DelayedTrigger, g *game.Game, p game.EffectParams) bool {
	return p.Player != uuid.Nil && ev.Target == p.Player && combatDamageToPlayerBy(ev, dt.Controller, g)
}

func createTappedTreasure(g *game.Game, item *game.StackItem) error {
	return CreateToken{Controller: item.Controller, Template: tappedTreasureToken(), N: 1}.Apply(NewContext(g, item))
}
