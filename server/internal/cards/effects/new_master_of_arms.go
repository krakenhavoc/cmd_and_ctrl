package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// New Master of Arms — Creature — Human Soldier {2}{W}, 2/2:
//
//	"First strike
//	 Prevent all combat damage that would be dealt by blocking creatures
//	 that are tapped.
//	 {1}{W}: Tap target creature blocking New Master of Arms."
//
// The prevention is a static replacement (CR 615.1a, like Candletrap's):
// it applies to combat damage dealt by ANY creature that is blocking and
// tapped, whoever controls it. Whether the creature is tapped and blocking
// is read as the damage is dealt (a blocker is tapped by the second
// ability, or by an effect that taps it after blockers are declared; a
// tapped creature still deals combat damage, CR 506.4 / 510.1). The
// activated ability targets a creature blocking this one
// (BlockingSource: this creature must itself be attacking).
func init() {
	Register(Spec{
		OracleID:        "d44b7b43-4674-47cd-b346-7db93a14b74c",
		Name:            "New Master of Arms",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"first strike"},
		Replacements: []game.ReplacementEffect{{
			Watches:    []game.EventKind{game.EventDealDamage},
			Prevention: true,
			AppliesTo: func(ev *game.ReplacementEvent, g *game.Game, _ *game.Card) bool {
				if ev.Kind != game.RepEventDamage || !ev.IsCombatDamage || ev.DamageAmount <= 0 || ev.DamageSource == uuid.Nil {
					return false
				}
				src, ok := g.LookupCardForEffect(ev.DamageSource)
				return ok && src.Tapped && src.BlockingTarget != uuid.Nil
			},
			Replace: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) error {
				ev.Cancel()
				return nil
			},
			Controller: func(_ *game.ReplacementEvent, _ *game.Game, src *game.Card) uuid.UUID {
				return src.Controller
			},
			Label: "New Master of Arms — prevent all combat damage dealt by blocking creatures that are tapped",
		}},
		Activated: []ActivatedAbility{{
			Label:   "{1}{W}: Tap target creature blocking New Master of Arms.",
			Cost:    ManaCost("{1}{W}"),
			Targets: BlockingSource(TargetCreature("target creature blocking New Master of Arms")),
			Purpose: game.Purpose{Answers: game.AnswerRemove},
			Effect:  thousandMoonsCrackshotTap, // tap the one legal target
		}},
	})
}
