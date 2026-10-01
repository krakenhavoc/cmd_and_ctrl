package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Dinosaur Egg — Creature — Dinosaur Egg {1}{G}, 0/3:
//
//	"Evolve (Whenever a creature you control enters, if that creature
//	 has greater power or toughness than this creature, put a +1/+1
//	 counter on this creature.)
//	 When this creature dies, you may discover X, where X is its
//	 toughness."
//
// X is the Egg's toughness as it last existed on the battlefield,
// counters included (CR 603.10a, ctx.TriggeringPermanent), so every
// +1/+1 counter evolve put on it raises the discover. Discover is
// ADR 0099's (game/discover.go). Evolve is the engine's keyword trigger
// (game/evolve.go, #1805), declared here as a printed keyword.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "ac4d5a97-6177-4eac-b0f2-cf10531ad879",
		Name:            "Dinosaur Egg",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{game.KeywordEvolve},
		Discovers:       true,
		Triggered: []game.TriggeredAbility{
			Optional(WhenThisDies("Dinosaur Egg — discover X", func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				info, ok := ctx.TriggeringPermanent()
				if !ok {
					return nil
				}
				return Discover{N: info.Toughness}.Apply(ctx)
			}), "Dinosaur Egg — discover X, where X is its toughness?"),
		},
	})
}
