package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Isengard Unleashed — Sorcery {2}{R}{R}{R}:
//
//	"Damage can't be prevented this turn. If a source you control would
//	 deal damage this turn to an opponent or a permanent an opponent
//	 controls, it deals triple that damage instead.
//	 Flashback {4}{R}{R}{R}"
//
// Insult's shape (ADR 0107 §5's turn grant, ADR 0108 §3's multiplier)
// with the recipient narrowed: an opponent, or a permanent an opponent
// controls, read as the damage would be dealt (CR 611.2c). Damage to you
// or to your own permanents is dealt as printed. Flashback is the
// engine's (CR 702.34a).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:         "2b947703-751a-4d95-b5de-e2d6b1fcb502",
		Name:             "Isengard Unleashed",
		Completeness:     CompletenessFull,
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{Flashback("{4}{R}{R}{R}")},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			if err := (DamageCantBePreventedThisTurn{}).Apply(ctx); err != nil {
				return err
			}
			return MultiplyDamage{Factor: 3, Sources: game.DamageSourcesYours,
				Recipients: game.DamageRecipientsOpponentsAndTheirPermanents,
				Label:      "Isengard Unleashed — your sources deal triple damage to opponents"}.Apply(ctx)
		},
	})
}
