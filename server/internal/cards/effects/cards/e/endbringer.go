package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Endbringer — Creature — Eldrazi {5}{C}, 5/5:
//
//	"Untap this creature during each other player's untap step.
//	 {T}: This creature deals 1 damage to any target.
//	 {C}, {T}: Target creature can't attack or block this turn.
//	 {C}{C}, {T}: Draw a card."
//
// The untap is Thousand Moons Infantry's (an UntapStepPermission aimed at
// the source). The {C} in the costs is colourless mana only (CR 107.4c).
// The three tap abilities are a ping (Apprentice Sorcerer's), Off Balance's
// restriction on a target creature until end of turn, and Snapping
// Voidcraw's draw.
func init() {
	const name = "Endbringer"
	effects.Register(effects.Spec{
		OracleID:     "db1ea6e0-4076-49dd-92f7-7e6d7eb65a27",
		Name:         name,
		Completeness: effects.CompletenessFull,
		UntapStep: []game.UntapStepPermission{
			effects.UntapSelfDuringEachOtherPlayersUntapStep(name + " — untap this creature"),
		},
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{T}: This creature deals 1 damage to any target.",
				Cost:    effects.TapCost(),
				Targets: effects.TargetAny(),
				Purpose: effects.ForTargets(effects.DamageToTarget(0, 1)),
				Effect:  effects.SourceDealsDamageToEachLegalTarget(1),
			},
			{
				Label:   "{C}, {T}: Target creature can't attack or block this turn.",
				Cost:    effects.Plus(effects.ManaCost("{C}"), effects.TapCost()),
				Targets: effects.TargetCreature("target creature"),
				Effect:  effects.NemTargetRestrictedThisTurn(game.CantAttackOrBlock, name+" — can't attack or block this turn"),
			},
			{
				Label:   "{C}{C}, {T}: Draw a card.",
				Purpose: game.Purpose{Answers: game.AnswerValue, Draws: 1},
				Cost:    effects.Plus(effects.ManaCost("{C}{C}"), effects.TapCost()),
				Effect:  effects.B36DrawOne,
			},
		},
	})
}
