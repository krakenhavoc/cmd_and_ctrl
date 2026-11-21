package cardsto

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Tor Wauki the Younger — Legendary Creature — Human Archer {3}{B}{R},
// 3/3:
//
//	"Reach, lifelink
//	 If another source you control would deal noncombat damage to a
//	 permanent or player, it deals that much damage plus 1 to that
//	 permanent or player instead.
//	 Whenever you cast an instant or sorcery spell, Tor Wauki deals 2
//	 damage to any target."
//
// The first is an additive damage replacement (CR 614.1a, CR 616.1), the
// shape of Thor, Asgard's Avenger and Aether Revolt without their
// recipient bound: any player or any permanent, whoever controls it. It
// is another source, so Tor Wauki's own damage, including the trigger's
// two, is not increased; it is noncombat only. The source is read for its
// controller as it is (damageSourceControlledBy, last-known when it has
// gone). The trigger is Gev's spell-cast shape aimed at any target, and
// Tor Wauki is the source of its damage, so lifelink gains the life.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "b44d5457-b4d6-4521-a25a-8ef96adba461",
		Name:            "Tor Wauki the Younger",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"reach", "lifelink"},
		Replacements: []game.ReplacementEffect{{
			Watches: []game.EventKind{game.EventDealDamage},
			AppliesTo: func(ev *game.ReplacementEvent, g *game.Game, src *game.Card) bool {
				if ev.Kind != game.RepEventDamage || ev.DamageAmount <= 0 || ev.IsCombatDamage ||
					ev.DamageSource == uuid.Nil || ev.DamageSource == src.InstanceID {
					return false
				}
				return effects.DamageSourceControlledBy(ev, g, src.Controller)
			},
			Replace: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) error {
				ev.DamageAmount++
				return nil
			},
			Controller: func(_ *game.ReplacementEvent, _ *game.Game, src *game.Card) uuid.UUID {
				return src.Controller
			},
			Label: "Tor Wauki the Younger: noncombat damage from another source plus 1",
		}},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WheneverYouCast(effects.Or(effects.Instant(), effects.Sorcery()), "Tor Wauki the Younger — 2 damage to any target",
				effects.DamageToFirstTarget(2)),
				effects.TargetAny()),
		},
	})
}
