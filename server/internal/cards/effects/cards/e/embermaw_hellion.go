package cardse

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Embermaw Hellion — Creature — Hellion {3}{R}{R}, 4/5 (Magic Origins):
//
//	"Trample
//	 If another red source you control would deal damage to a permanent
//	 or player, it deals that much damage plus 1 to that permanent or
//	 player instead."
//
// Torbran's replacement with +1 in place of +2, no opponent gate (any
// permanent or player, yours included) and "another": the Hellion's own
// damage is untouched. The source is read as the damage event
// snapshotted it (#1417, CR 608.2h). Two Hellions each add 1 to the
// other's damage, as printed.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "01c39670-3677-49df-a4f6-e5062e0909d5",
		Name:            "Embermaw Hellion",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Replacements: []game.ReplacementEffect{{
			Watches: []game.EventKind{game.EventDealDamage},
			AppliesTo: func(ev *game.ReplacementEvent, g *game.Game, src *game.Card) bool {
				if ev.Kind != game.RepEventDamage || ev.DamageAmount <= 0 || ev.DamageSource == uuid.Nil || ev.DamageSource == src.InstanceID {
					return false
				}
				return effects.DamageSourceIsRedControlledBy(ev, g, src.Controller, false)
			},
			Replace: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) error {
				ev.DamageAmount++
				return nil
			},
			Controller: func(_ *game.ReplacementEvent, _ *game.Game, src *game.Card) uuid.UUID {
				return src.Controller
			},
			Label: "Embermaw Hellion: +1 damage",
		}},
	})
}
