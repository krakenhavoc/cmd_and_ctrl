package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Angelic Benediction — Enchantment {3}{W} (Shards of Alara, #5588):
//
//	"Exalted
//	 Whenever a creature you control attacks alone, you may tap target
//	 creature."
//
// Exalted is the canonical keyword (CR 702.83a); an enchantment's
// exalted pumps the lone attacker like a creature's. The tap is the
// attacks-alone trigger (CR 506.5) with a target chosen as it goes on the
// stack (CR 603.3d) and a "you may" asked as it resolves (CR 603.5).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "6b19377c-f178-49cb-95f8-fd3451acc7cf",
		Name:            "Angelic Benediction",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{game.KeywordExalted},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.Optional(
					effects.On(game.EventAttack, effects.MshB1YouAttackAlone,
						"Angelic Benediction — tap target creature", effects.TapEachLegalTarget),
					"Angelic Benediction — tap that creature?"),
				effects.TargetCreature("target creature")),
		},
	})
}
