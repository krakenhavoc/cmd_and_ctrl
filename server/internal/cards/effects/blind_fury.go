package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Blind Fury — Instant {2}{R}{R}:
//
//	"All creatures lose trample until end of turn. If a creature would
//	 deal combat damage to a creature this turn, it deals double that
//	 damage to that creature instead."
//
// Two effects. Losing trample changes a characteristic, so its set is the
// creatures on the battlefield as the spell resolves (CR 611.2c): a
// creature that enters later keeps its trample. The multiplier changes
// none, so it reads every creature as the combat damage would be dealt
// (ADR 0108 §3): creature to creature, combat damage only, for the rest
// of the turn. Combat damage to a player or a planeswalker that is not a
// creature is dealt as printed.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "5dbfd316-a0a9-4caa-99b1-069931d2aaa4",
		Name:         "Blind Fury",
		Completeness: CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			if err := (ScopedEffectFor{
				Match:    Creature(),
				Mods:     []game.Mod{game.RemoveKeywordsMod("trample")},
				Duration: DurationUntilEndOfTurn(ctx),
				Label:    "Blind Fury — creatures lose trample",
			}).Apply(ctx); err != nil {
				return err
			}
			return MultiplyDamage{Factor: 2, Sources: game.DamageSourcesCreatures,
				Recipients: game.DamageRecipientsCreatures, CombatOnly: true,
				Label: "Blind Fury — creatures deal double combat damage to creatures"}.Apply(ctx)
		},
	})
}
