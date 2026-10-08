package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Aether Revolt — Enchantment {2}{R}{R}:
//
//	"Revolt — As long as a permanent left the battlefield under your
//	 control this turn, if a source you control would deal noncombat
//	 damage to an opponent or a permanent an opponent controls, it deals
//	 that much damage plus 2 instead.
//	 Whenever you get one or more {E}, this enchantment deals that much
//	 damage to any target."
//
// ADR 0129 §6 (#1995). The replacement is Solphim, Mayhem Dominus's
// with an addition in place of a doubling (CR 614.1a), switched on by
// revolt: the turn tally's record of a permanent that left the
// battlefield under its controller's control (Revolt). The trigger is
// one per placement of energy (CR 603.2c), and "that much" is the
// energy that landed, after any replacement such as Izzet
// Generatorium's. Its damage is noncombat, so with revolt on it is 2
// more against an opponent or their permanent.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "b73661a6-d136-4c65-804c-461e83484f4b",
		Name:         "Aether Revolt",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{{
			Watches: []game.EventKind{game.EventDealDamage},
			AppliesTo: func(ev *game.ReplacementEvent, g *game.Game, src *game.Card) bool {
				if ev.Kind != game.RepEventDamage || ev.DamageAmount <= 0 || ev.IsCombatDamage {
					return false
				}
				if !Revolt(g, src.Controller) || !damageSourceControlledBy(ev, g, src.Controller) {
					return false
				}
				return damageHitsAnOpponentOf(ev, g, src.Controller)
			},
			Replace: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) error {
				ev.DamageAmount += 2
				return nil
			},
			Controller: func(_ *game.ReplacementEvent, _ *game.Game, src *game.Card) uuid.UUID {
				return src.Controller
			},
			Label: "Aether Revolt: +2 noncombat damage",
		}},
		Triggered: []game.TriggeredAbility{
			Targeting(WheneverYouGetEnergy("Aether Revolt — that much damage to any target",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					legal := ctx.LegalTargets()
					if len(legal) == 0 {
						return nil
					}
					return DealDamage{Source: ctx.Source(), Target: legal[0].ID, Amount: EnergyGotten(item)}.Apply(ctx)
				}), TargetAny()),
		},
	})
}
