package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Dwarven Sea Clan — Creature — Dwarf {2}{R}, 1/1 (Homelands, S82 OLD b824,
// #5635):
//
//	"{T}: Choose target attacking or blocking creature whose controller controls an
//	Island. This creature deals 2 damage to that creature at end of combat.
//	Activate only before the end of combat step."
//
// The target clause (attacking or blocking, controller has an Island) is judged
// as the ability is announced and again as it resolves; the damage is a delayed
// trigger (CR 603.7) at the end of combat step, dealt by the Clan as it last
// existed (CR 608.2h), to the creature only if it is still on the battlefield.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "55ae9711-65ac-4166-8a51-1deeb10a8904",
		Name:         "Dwarven Sea Clan",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:     "{T}: Choose target attacking or blocking creature whose controller controls an Island. This creature deals 2 damage to that creature at end of combat. Activate only before the end of combat step.",
			Purpose:   game.Purpose{Answers: game.AnswerRemove},
			Cost:      effects.TapCost(),
			Targets:   effects.TargetCreature("target attacking or blocking creature whose controller controls an Island", effects.Or(effects.AttackingCreature(), effects.BlockingCreature()), effects.OldB824ControllerHasAnIsland()),
			Condition: effects.OldB824BeforeTheEndOfCombatStep(),
			Effect:    effects.OldB824DealsTwoAtEndOfCombat("Dwarven Sea Clan"),
		}},
	})
}
