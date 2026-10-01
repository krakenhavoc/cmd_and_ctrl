package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Merchant Ship — Creature — Human {U}, 0/2:
//
//	"This creature can't attack unless defending player controls an
//	 Island.
//	 Whenever this creature attacks and isn't blocked, you gain
//	 2 life.
//	 When you control no Islands, sacrifice this creature."
//
// ADR 0107 (#1858, #1879). The attack restriction is §2's (CR 508.1c),
// with the defending player worked out per target (CR 508.5, 508.5a), so
// in Commander it may attack only the opponents who control an Island.
// The sacrifice is §1's CR 603.8 state trigger over its own controller's
// Islands: it triggers as soon as the last one is gone and not again while
// it waits or is on the stack. Both halves read one PermanentQuery.
//
// "Attacks and isn't blocked" triggers once blockers are declared (CR
// 509.3g), through the catalog's shared constructor.
//
// No simplification.
func init() {
	q := QuerySubtype("Island")
	Register(Spec{
		OracleID:     "69556f6c-c05b-4902-bac7-012f0ed81b75",
		Name:         "Merchant Ship",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			CantAttackUnlessDefendingPlayerControls(q),
		},
		Triggered: []game.TriggeredAbility{
			WhenAttacksAndIsNotBlockedEffect("Merchant Ship — you gain 2 life",
				func(g *game.Game, item *game.StackItem, _ uuid.UUID) error {
					return GainLife{Amount: 2}.Apply(NewContext(g, item))
				}),
			WhenYouControlNo(q, "Merchant Ship — sacrifice it", SacrificeThisIfStillOnBattlefield),
		},
	})
}
