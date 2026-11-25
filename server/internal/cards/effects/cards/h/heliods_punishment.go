package cardsh

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Heliod's Punishment — Enchantment — Aura {1}{W}:
//
//	"Enchant creature
//	 This Aura enters with four task counters on it.
//	 Enchanted creature can't attack or block. It loses all abilities and has
//	 "{T}: Remove a task counter from Heliod's Punishment. Then if it has no
//	 task counters on it, destroy Heliod's Punishment.""
//
// The ability is the creature's ({T} is subject to summoning sickness), so
// it comes in layer 6 with the loss of the other abilities as one effect.
// It names the Aura that granted it, which the item reads as its grantor;
// an Aura that has left is not touched again (CR 400.7).
func init() {
	removal := effects.LoseAllAbilities()
	removal.GrantAbilities = []string{heliodsPunishmentGrant}
	effects.Register(effects.Spec{
		OracleID:     "ac5d5133-6f5c-43b7-82da-24802246e5a4",
		Name:         "Heliod's Punishment",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraHarm},
		Replacements: []game.ReplacementEffect{
			effects.B10EntersWithCounters("task", 4, "Heliod's Punishment: enters with four task counters"),
		},
		Grants: []effects.AbilityGrant{{
			Key: heliodsPunishmentGrant,
			Activated: []effects.ActivatedAbility{{
				Label:   "{T}: Remove a task counter from Heliod's Punishment. Then if it has no task counters on it, destroy Heliod's Punishment.",
				Cost:    effects.TapCost(),
				Purpose: game.Purpose{Answers: game.AnswerValue},
				Effect:  heliodsPunishmentTask,
			}},
			Text: "{T}: Remove a task counter from Heliod's Punishment. Then if it has no task counters on it, destroy Heliod's Punishment.",
		}},
		Static: []game.StaticAbility{
			effects.RestrictAttached(game.CantAttackOrBlock),
			removal,
		},
	})
}

const heliodsPunishmentGrant = "heliods-punishment/remove-a-task-counter"

// heliodsPunishmentTask removes one task counter from the Aura that granted
// the ability, then destroys it when none are left (CR 608.2c: the "then"
// reads the Aura as it is after the removal).
func heliodsPunishmentTask(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	aura := ctx.Grantor()
	if aura == uuid.Nil || !effects.OnBattlefield(g, aura) {
		return nil
	}
	if c, ok := g.LookupCardForEffect(aura); ok && c.Counters["task"] > 0 {
		if err := (effects.AddCounter{Target: aura, Kind: "task", N: -1}).Apply(ctx); err != nil {
			return err
		}
	}
	if c, ok := g.LookupCardForEffect(aura); ok && c.Counters["task"] <= 0 {
		return effects.DestroyTarget{Target: aura}.Apply(ctx)
	}
	return nil
}
