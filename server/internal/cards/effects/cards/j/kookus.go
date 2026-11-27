package cardsj

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Kookus — Creature — Djinn {3}{R}{R}, 3/5 (Visions, S82 OLD batch 786,
// #5631):
//
//	"Trample
//	 At the beginning of your upkeep, if you don't control a creature named
//	 Keeper of Kookus, this creature deals 3 damage to you and attacks this
//	 turn if able.
//	 {R}: This creature gets +1/+0 until end of turn."
//
// The upkeep trigger has an intervening "if" (CR 603.4): it triggers only
// while you control no creature named Keeper of Kookus and is checked again
// as it resolves. The damage is Kookus's, dealt to its controller; "attacks
// this turn if able" is Imps' Taunt's requirement (AddAttackRequirementMod,
// CR 508.1d) until end of turn, set only if Kookus is still on the
// battlefield. The firebreathing is the shared ThisGetsUntilEndOfTurn body.
//
// No simplification.
func init() {
	upkeep := effects.InterveningIfYou("you don't control a creature named Keeper of Kookus",
		func(g *game.Game, you uuid.UUID) bool {
			return !effects.JmpB1YouControlACreatureNamed(g, you, "Keeper of Kookus")
		},
		effects.AtYourUpkeep("Kookus — it deals 3 damage to you and attacks this turn if able",
			func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				if err := (effects.DealDamage{Source: item.SourceCardID, Target: item.Controller, Amount: 3}).Apply(ctx); err != nil {
					return err
				}
				if !effects.OnBattlefield(g, item.SourceCardID) {
					return nil
				}
				return effects.ScopedEffectFor{
					Target:   item.SourceCardID,
					Mods:     []game.Mod{game.AddAttackRequirementMod(uuid.Nil)},
					Duration: effects.DurationUntilEndOfTurn(ctx),
					Label:    "Kookus — attacks this turn if able",
				}.Apply(ctx)
			}))
	effects.Register(effects.Spec{
		OracleID:        "4c8196e0-c392-4402-b77c-e3d009e9899c",
		Name:            "Kookus",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered:       []game.TriggeredAbility{upkeep},
		Activated: []effects.ActivatedAbility{{
			Label:   "{R}: This creature gets +1/+0 until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump, Pump: &game.Pump{Power: 1}},
			Cost:    effects.ManaCost("{R}"),
			Effect:  effects.ThisGetsUntilEndOfTurn(1, 0, "Kookus — +1/+0 until end of turn"),
		}},
	})
}
