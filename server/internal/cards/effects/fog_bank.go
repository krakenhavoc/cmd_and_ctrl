package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Fog Bank — Creature — Wall {1}{U}, 0/2:
//
//	"Defender (This creature can't attack.)
//	 Flying
//	 Prevent all combat damage that would be dealt to and dealt by this
//	 creature."
//
// The last line is a standing CR 615 prevention (Marble Priest's shape),
// read live off the battlefield: combat damage whose recipient or whose
// source is this object is prevented. Noncombat damage is not touched,
// so a burn spell still kills it. A Fog Bank that has lost its abilities
// (Humility) prevents nothing, because the replacement is gathered
// through the catalog key.
//
// No simplification. Guffster's Party deck request (#4184).
func init() {
	Register(Spec{
		OracleID:        "3a2bb47d-228d-499e-a34b-9e10d99e9b2c",
		Name:            "Fog Bank",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"defender", "flying"},
		Replacements: []game.ReplacementEffect{{
			Watches:    []game.EventKind{game.EventDealDamage},
			Prevention: true, // CR 615.1a — "prevent"; CR 615.12 reads it
			AppliesTo: func(ev *game.ReplacementEvent, _ *game.Game, src *game.Card) bool {
				if ev.Kind != game.RepEventDamage || !ev.IsCombatDamage {
					return false
				}
				return ev.DamageTarget == src.InstanceID || ev.DamageSource == src.InstanceID
			},
			Replace: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) error {
				ev.Cancel()
				return nil
			},
			Controller: func(_ *game.ReplacementEvent, _ *game.Game, src *game.Card) uuid.UUID {
				return src.Controller
			},
			Label: "Fog Bank — prevent all combat damage dealt to and by it",
		}},
	})
}
