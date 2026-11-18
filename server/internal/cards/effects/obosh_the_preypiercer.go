package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Obosh, the Preypiercer — Legendary Creature — Hellion Horror
// {3}{B/R}{B/R}, 3/5:
//
//	"Companion — Your starting deck contains only cards with odd mana
//	 values and land cards. (...)
//	 If a source you control with an odd mana value would deal damage to a
//	 permanent or player, it deals double that damage to that permanent or
//	 player instead."
//
// A CR 614 damage replacement. The source is read as the damage event
// snapshotted it (a creature that has left deals its damage as the source
// it last was, CR 608.2h), so its controller and its mana value are the
// last-known ones. A source whose mana value is not known (an emblem) is
// not odd. Each of Obosh's own replacements applies once to a given
// event, so two Oboshes double twice (CR 616.1). Companion is the deck
// check's business (internal/deck companion registry, ADR 0172).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "6985c8a9-9299-461b-8bef-bd2b2c739c11",
		Name:         "Obosh, the Preypiercer",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{{
			Watches:  []game.EventKind{game.EventDealDamage},
			RowLabel: "If a source you control with an odd mana value would deal damage to a permanent or player, it deals double that damage to that permanent or player instead.",
			Label:    "Obosh, the Preypiercer — double that damage",
			AppliesTo: func(ev *game.ReplacementEvent, g *game.Game, src *game.Card) bool {
				if src == nil || ev.Kind != game.RepEventDamage || ev.DamageAmount <= 0 || ev.DamageSource == uuid.Nil {
					return false
				}
				ch, ok := damageSourceCharacteristics(ev, g)
				return ok && ch.Controller == src.Controller && ch.SourceManaValueKnown && ch.SourceManaValue%2 == 1
			},
			Replace: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) error {
				ev.DamageAmount *= 2
				return nil
			},
			Controller: func(_ *game.ReplacementEvent, _ *game.Game, src *game.Card) uuid.UUID { return src.Controller },
		}},
	})
}
