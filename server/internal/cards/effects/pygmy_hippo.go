package effects

import (
	"strconv"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Pygmy Hippo — Creature — Hippo {G}{U}, 2/2:
//
//	"Whenever this creature attacks and isn't blocked, you may have
//	 defending player activate a mana ability of each land they control
//	 and lose all unspent mana. If you do, this creature assigns no combat
//	 damage this turn and at the beginning of your next main phase this
//	 turn, you add an amount of {C} equal to the amount of mana that player
//	 lost this way."
//
// #3952, ADR 0200: the defending player is the one whose block declaration
// fired the trigger. Saying yes marks the Hippo (CR 510.1a), then the
// defending player activates a mana ability of each land they control,
// choosing which one and its colours (ActivateAManaAbilityOfEachLand), and
// loses their whole pool, mana that was already there included. The amount
// lost is the {C} a delayed trigger adds at the beginning of the next main
// phase still to come this turn (CR 603.7), bound to that phase; a turn with
// no main phase left adds nothing.
//
// No simplifications.
func init() {
	const label = "Pygmy Hippo — you may have defending player activate a mana ability of each land they control and lose all unspent mana"
	Register(Spec{
		OracleID:     "0d39513a-1f05-4d9a-a772-745f4f1a3471",
		Name:         "Pygmy Hippo",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenAttacksAndIsNotBlockedEffect(label, func(g *game.Game, item *game.StackItem, defender uuid.UUID) error {
				return MayChoice{
					Question: label + "? (it then assigns no combat damage this turn)",
					OnYes: func(ctx *Context) error {
						assignsNoCombatDamageThisTurn(ctx, "Pygmy Hippo — assigns no combat damage this turn")
						if !defendingPlayerStillIn(ctx.Game, defender) {
							return nil
						}
						return ActivateAManaAbilityOfEachLand{
							Player: defender,
							Card:   "Pygmy Hippo",
							Then: func(ctx *Context, lost []game.ManaToken) error {
								n := game.TotalMana(lost)
								if n <= 0 {
									return nil
								}
								step, phase, ok := ctx.Game.NextMainPhaseThisTurnForEffect()
								if !ok {
									return nil
								}
								return ScheduleDelayedTrigger{
									At:      step,
									OnPhase: phase,
									Label:   "Pygmy Hippo — add {C} × " + strconv.Itoa(n),
									Body:    pygmyHippoManaBody,
									Params:  game.EffectParams{Amount: n},
								}.Apply(ctx)
							},
						}.Apply(ctx)
					},
				}.Apply(NewContext(g, item))
			}),
		},
	})
}
