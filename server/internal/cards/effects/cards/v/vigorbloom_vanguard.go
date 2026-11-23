package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Vigorbloom Vanguard // Seed Suture — Creature — Troll Druid {1}{G/W},
// 2/2 // Sorcery {G/W} (preparation card, CR 722):
//
//	"This creature enters prepared. (While it's prepared, you may cast a
//	 copy of its spell. Doing so unprepares it.)
//	 Each creature you control with a +1/+1 counter on it has vigilance."
//
//	Seed Suture — "Put a +1/+1 counter on target creature. You gain 1
//	 life."
//
// The vigilance grant is a layer 6 static read live, so a creature
// gains it the moment the counter lands and loses it with the last one.
// The Vanguard counts itself when it has a counter.
//
// No simplification.
func init() {
	const id = "bd416426-d037-45e3-9647-816bf982cbea"
	effects.Register(effects.Spec{
		OracleID:     id,
		Name:         "Vigorbloom Vanguard",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{effects.SelfEntersPrepared()},
		Static: []game.StaticAbility{
			effects.B16GrantKeywords(func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.IsCreature() && target.Controller == source.Controller && effects.B34HasPlusCounter(*target)
			}, "vigilance"),
		},
	})
	effects.Register(effects.Spec{
		OracleID:     id + "#1",
		Name:         "Seed Suture",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		OnResolve:    effects.SeedSutureResolve,
	})
}
