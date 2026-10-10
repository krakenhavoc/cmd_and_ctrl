package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Heroism — Enchantment {2}{W}:
//
//	"Sacrifice a white creature: For each attacking red creature, prevent all combat damage that would be dealt by that creature this turn unless its controller pays {2}{R}."
//
// ADR 0108 §7 (#1904, Delivery PR 7b): the attacking red creatures are
// read as the ability resolves, and each one's controller is asked to pay
// {2}{R} for it, then and there (its ruling). A creature whose controller
// declines, or cannot pay, gets a combat-damage shield with itself as the
// source, pinned now (CR 400.7). The prompts hold the step they are asked
// in (the pay-or-else anchor UpkeepPayUnless carries), so nobody can walk
// past them to the combat damage they are about. Paying stops only
// Heroism's prevention, not any other (its ruling).
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:     "74716642-fc8c-4f62-a556-154ed0b4af0e",
		Name:         "Heroism",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "Sacrifice a white creature: For each attacking red creature, prevent all combat damage that would be dealt by that creature this turn unless its controller pays {2}{R}.",
			Purpose: game.Purpose{Answers: game.AnswerPrevent | game.AnswerSacOutlet},
			Cost:    SacrificeN(1, "a white creature", Creature(), OfColor("W")),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				g.RecomputeLayersIfStaleLocked()
				for _, c := range g.BattlefieldCardsForEffect() {
					if !c.IsCreature() || c.AttackingTarget == uuid.Nil || !c.HasColor("R") {
						continue
					}
					attacker := c.InstanceID
					if err := (UpkeepPayUnless{
						Chooser:  c.Controller,
						Cost:     "{2}{R}",
						Question: "Heroism — pay {2}{R}, or " + c.Name + "'s combat damage this turn is prevented",
						OnDecline: func(ctx *Context) error {
							return PreventDamageFromSource{From: attacker, CombatOnly: true, Protect: ShieldAnything}.Apply(ctx)
						},
					}).Apply(ctx); err != nil {
						return err
					}
				}
				return nil
			},
		}},
	})
}
