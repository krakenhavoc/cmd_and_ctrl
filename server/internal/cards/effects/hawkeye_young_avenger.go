package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Hawkeye, Young Avenger — Legendary Creature — Human Archer Hero
// {3}{R}, 2/4 (Marvel Super Heroes, #4455):
//
//	"Reach
//	 If a source you control would deal noncombat damage to an opponent or
//	 a permanent an opponent controls, instead it deals that much damage
//	 plus X, where X is Hawkeye's power."
//
// Torbran's replacement with three differences. The source is any source
// you control (a spell, an ability or a permanent, read from the damage
// event's snapshot so a creature that has left still counts, CR 608.2h),
// not only a red one. Combat damage is left alone (the Torture Pit
// gate). And the addition is not a constant but Hawkeye's power as the
// damage is replaced, so a pumped Hawkeye adds more and two Hawkeyes each
// add their own. Damage to you or to your permanents is untouched.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "c8f8b4ed-8455-45cd-a24c-e2f42cd5e5b1",
		Name:         "Hawkeye, Young Avenger",
		Completeness: CompletenessFull,

		PrintedKeywords: []string{"reach"},
		Replacements: []game.ReplacementEffect{{
			Watches: []game.EventKind{game.EventDealDamage},
			AppliesTo: func(ev *game.ReplacementEvent, g *game.Game, src *game.Card) bool {
				if ev.Kind != game.RepEventDamage || ev.DamageAmount <= 0 || ev.IsCombatDamage || ev.DamageSource == uuid.Nil {
					return false
				}
				return damageSourceControlledBy(ev, g, src.Controller) && damageHitsAnOpponentOf(ev, g, src.Controller)
			},
			Replace: func(ev *game.ReplacementEvent, _ *game.Game, src *game.Card) error {
				ev.DamageAmount += src.CurrentPower()
				return nil
			},
			Controller: func(_ *game.ReplacementEvent, _ *game.Game, src *game.Card) uuid.UUID {
				return src.Controller
			},
			Label: "Hawkeye, Young Avenger: noncombat damage plus Hawkeye's power",
		}},
	})
}
