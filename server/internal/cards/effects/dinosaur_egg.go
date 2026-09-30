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
// counters included (CR 603.10a, ctx.TriggeringPermanent). Discover is
// ADR 0099's (game/discover.go).
//
// Declared simplification, weaker than printed: evolve (CR 702.100) has
// no engine support, so the Egg never grows and dies as the 0/3 it was
// printed as, discovering 3 unless something else changed its
// toughness. The roadmap's evolve row lists it.
func init() {
	Register(Spec{
		OracleID:     "ac4d5a97-6177-4eac-b0f2-cf10531ad879",
		Name:         "Dinosaur Egg",
		Completeness: CompletenessCaveats,
		Caveats:      []string{"Evolve isn't implemented, so the Egg never gets +1/+1 counters from creatures entering."},
		Discovers:    true,
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
