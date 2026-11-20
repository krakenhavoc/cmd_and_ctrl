package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:        "6b19377c-f178-49cb-95f8-fd3451acc7cf",
		Name:            "Angelic Benediction",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{game.KeywordExalted},
		Triggered: []game.TriggeredAbility{
			Targeting(
				Optional(
					On(game.EventAttack, mshB1YouAttackAlone,
						"Angelic Benediction — tap target creature", tapEachLegalTarget),
					"Angelic Benediction — tap that creature?"),
				TargetCreature("target creature")),
		},
	})
}
