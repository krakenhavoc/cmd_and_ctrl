package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Breath of the Sleepless — Enchantment {3}{U} (Innistrad: Crimson Vow
// Commander):
//
//	"You may cast Spirit spells as though they had flash.
//	 Whenever you cast a creature spell during an opponent's turn, tap up to
//	 one target creature."
//
// The timing grant is Blur of Heroism's CastKindAsThoughFlash narrowed to
// the Spirit type. The trigger reads the active player when the spell is
// cast, and takes up to one target (zero taps nothing).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "eb887067-4466-45da-8387-4444fdc723f8",
		Name:         "Breath of the Sleepless",
		Completeness: CompletenessFull,
		CastTimings: []game.CastTimingRule{
			CastKindAsThoughFlash(game.PermissionFilter{CreatureType: "Spirit"},
				"You may cast Spirit spells as though they had flash."),
		},
		Triggered: []game.TriggeredAbility{
			Targeting(
				On(game.EventCast, vowB12YouCastACreatureSpellOnAnOpponentsTurn,
					"Breath of the Sleepless — tap up to one target creature", Do(Instruction(vowB12TapUpToOneTarget))),
				vowB12UpToOneCreature()),
		},
	})
}
