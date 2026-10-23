package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// keyword_reading_grant.go — the battlefield static "<creatures> with
// <keyword> have <keywords>" (#2663, ADR 0067 amendment 2026-10-23):
// Sephara, Sky's Blade's "Other creatures you control with flying have
// indestructible", Kwende, Pride of Femeref's "Creatures you control with
// first strike have double strike".
//
// The keyword it reads is added and removed in layer 6, the layer it
// grants in, so the static declares game.StaticAbility.ReadsAbilities and
// the engine applies it after every effect that would add or remove that
// keyword on a creature it looks at, whatever the timestamps (CR 613.8a,
// 613.8b). A creature that gains flying from a later grant has
// indestructible, and one that loses flying to a later "loses flying" does
// not. A keyword the creature can't have (the Archetypes, CR 613.1f) does not
// count either, though the strip runs only once the bucket is over
// (game.HasKeywordItCanHave).
//
// Append-only: add a helper below the last one.

// WithKeywordHave is the layer-6 static "<scope> with <reads> have
// <grants>". `scope` is the static's ordinary predicate without the
// keyword clause ("other creatures you control"), read live on every pass
// (CR 611.3a); `reads` and `grants` are canonical lowercase keyword tokens
// (game/keywords.go). `label` is the printed sentence.
func WithKeywordHave(scope func(target *game.Card, g *game.Game, source *game.Card) bool, reads, label string, grants ...string) game.StaticAbility {
	gave := append([]string(nil), grants...)
	return game.StaticAbility{
		Label:          label,
		Layer:          game.Layer6Ability,
		ReadsAbilities: true,
		AppliesTo: func(target *game.Card, g *game.Game, source *game.Card) bool {
			return scope(target, g, source) && game.HasKeywordItCanHave(target, reads)
		},
		Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
			for _, kw := range gave {
				c.Abilities = game.AppendKeywordAbility(c.Abilities, kw)
			}
		},
	}
}

// OtherCreaturesYouControl is the static scope "other creatures you
// control": the source's controller's creatures, the source excluded.
func OtherCreaturesYouControl(target *game.Card, _ *game.Game, source *game.Card) bool {
	return target.InstanceID != source.InstanceID && target.Controller == source.Controller && target.IsCreature()
}

// CreaturesYouControl is the static scope "creatures you control", the
// source included when it is one.
func CreaturesYouControl(target *game.Card, _ *game.Game, source *game.Card) bool {
	return target.Controller == source.Controller && target.IsCreature()
}
