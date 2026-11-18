package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Awakened Awareness — Enchantment — Aura {X}{U}{U} (Neon Dynasty, #5493):
//
//	"Enchant artifact or creature
//	 When this Aura enters, put X +1/+1 counters on enchanted permanent.
//	 As long as enchanted permanent is a creature, it has base power and
//	 toughness 1/1."
//
// X is the value announced for the Aura spell (CR 107.3m), carried on the
// enters trigger as it is built (WhenThisEntersWithCastX), so it is exact
// even if the Aura has left by the time the trigger resolves. The counters
// land on the permanent the Aura enchants as the trigger resolves, or the
// one it last enchanted (CR 608.2h), if it is still on the battlefield. The
// base 1/1 is a layer 7b set, so the counters (and any pump) still add.
// Counters go on an artifact that is not a creature too, and wait there.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "6629e5f3-ac59-48f8-b09c-3b98f59b9922",
		Name:         "Awakened Awareness",
		Completeness: CompletenessFull,
		Targets:      TargetPermanent("enchant artifact or creature", Or(Artifact(), Creature())),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Static:       []game.StaticAbility{neoB12BasePTWhenCreature(1, 1)},
		Triggered: []game.TriggeredAbility{
			WhenThisEntersWithCastX("Awakened Awareness — put X +1/+1 counters on enchanted permanent",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					host, ok := enchantedCreatureAsItResolves(ctx)
					if !ok || !onBattlefield(g, host) || item.Params.Amount <= 0 {
						return nil
					}
					return AddCounter{Target: host, Kind: game.CounterPlusOne, N: item.Params.Amount}.Apply(ctx)
				}),
		},
	})
}
