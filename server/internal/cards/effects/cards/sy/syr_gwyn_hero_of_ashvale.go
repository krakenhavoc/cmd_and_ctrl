package cardssy

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Syr Gwyn, Hero of Ashvale — Legendary Creature — Human Knight
// {3}{R}{W}{B}, 5/5 (Throne of Eldraine, #5516):
//
//	"Vigilance, menace
//	 Whenever an equipped creature you control attacks, you draw a card and
//	 you lose 1 life.
//	 Equipment you control have equip Knight {0}."
//
// The attack trigger is Swordsman's (one trigger per equipped attacker you
// control, judged as the attack is declared). The grant adds an equip
// ability, limited to a Knight you control, to every Equipment you control
// (Puresteel Paladin's shape); the Equipment keeps its printed equip (CR
// 702.6d) and the granted one is an ordinary equip, sorcery speed.
//
// No simplification.
const syrGwynEquipGrant = "syr-gwyn-hero-of-ashvale/equip-knight-zero"

func init() {
	effects.Register(effects.Spec{
		OracleID:        "ca411184-3413-492c-ab87-08c9fa1d4806",
		Name:            "Syr Gwyn, Hero of Ashvale",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"vigilance", "menace"},
		Grants: []effects.AbilityGrant{{
			Key: syrGwynEquipGrant,
			Activated: []effects.ActivatedAbility{effects.EquipOnlyAbility("Equip Knight {0}", "{0}",
				effects.TargetCreature("target Knight you control", effects.YouControl(), effects.Subtype("Knight")))},
			Text: "Equip Knight {0}",
		}},
		Static: []game.StaticAbility{effects.GrantAbilities(effects.EquipmentYouControl, syrGwynEquipGrant)},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventAttack, effects.EquippedCreatureYouControlAttacks, "Syr Gwyn — you draw a card and you lose 1 life",
				effects.Do(effects.DrawCards{N: 1}, effects.YouLoseLife(1))),
		},
	})
}
