package cardsgr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Grey Knight Paragon — Creature — Astartes Knight {4}{W}, 4/4 (Warhammer
// 40,000 Commander, S82 GAP batch 196, #5948):
//
//	"Flash
//	 Rites of Banishment — When this creature enters, destroy target attacking
//	 creature. If that creature is a Demon, exile it instead."
//
// The target is an attacking creature chosen as the trigger goes on the stack
// and judged again as it resolves; the Demon check reads it as it is then. A
// creature with regeneration or indestructible is only destroyed as usual;
// the exile ignores both.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "12cc97ec-5d03-4434-a31b-51e77d208466",
		Name:            "Grey Knight Paragon",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisEnters("Grey Knight Paragon — destroy target attacking creature; if it's a Demon, exile it instead",
				effects.GapB196RitesOfBanishment),
				effects.TargetCreature("target attacking creature", effects.AttackingCreature())),
		},
	})
}
