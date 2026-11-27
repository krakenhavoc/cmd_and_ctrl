package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Giltspire Avenger — Creature — Human Soldier {G}{W}{U}, 2/2 (Conflux):
//
//	"Exalted (Whenever a creature you control attacks alone, that creature
//	 gets +1/+1 until end of turn.)
//	 {T}: Destroy target creature that dealt damage to you this turn."
//
// Exalted is the canonical keyword, the trigger derived by the engine
// (CR 702.83a). The target clause reads the per-turn record of which creature
// objects dealt damage to this ability's controller (Reciprocate, #2149): a
// creature that left and came back is a new object and not a legal target
// (CR 400.7), and one that damaged someone else is not either. Legality is
// checked again as the ability resolves (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "e29350cd-c6ff-454d-a1fc-0c1bb60047b1",
		Name:            "Giltspire Avenger",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{game.KeywordExalted},
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}: Destroy target creature that dealt damage to you this turn.",
			Purpose: game.Purpose{Answers: game.AnswerRemove},
			Cost:    effects.TapCost(),
			Targets: effects.TargetCreature("target creature that dealt damage to you this turn", effects.DealtDamageToYouThisTurn()),
			Effect:  effects.DestroyFirstLegalTarget,
		}},
	})
}
