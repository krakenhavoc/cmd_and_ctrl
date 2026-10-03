package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Arcbound Wanderer — Artifact Creature — Golem {6}, 0/0:
//
//	"Modular—Sunburst (This creature enters with a +1/+1 counter on it
//	 for each color of mana spent to cast it. When it dies, you may put
//	 its +1/+1 counters on target artifact creature.)"
//
// Modular (CR 702.43a) is two abilities: "this permanent enters with N
// +1/+1 counters on it", and a dies trigger. Here N is sunburst's count
// (CR 702.44c, sunburst setting the number for another ability), and on
// a creature that is exactly sunburst's own +1/+1 counter per colour —
// so the first half is the sunburst keyword the engine reads off the
// resolving spell (ADR 0109 §11, #1552).
//
// The second half is an optional, targeted dies trigger that puts one
// +1/+1 counter on the target for each +1/+1 counter the Wanderer had as
// it died, read from its last-known information (CR 603.10a).
func init() {
	Register(Spec{
		OracleID:            "03436524-197a-4941-a2a1-c7c4b71c4709",
		Name:                "Arcbound Wanderer",
		Completeness:        CompletenessCaveats,
		Caveats:             []string{"With strict mana off, the game doesn't track which mana you spent, so it enters with no counters at all."},
		WantsDistinctColors: true,
		PrintedKeywords:     []string{game.KeywordSunburst},
		Triggered: []game.TriggeredAbility{
			Optional(Targeting(WhenThisDies("Arcbound Wanderer — modular: its +1/+1 counters on target artifact creature",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					info, ok := ctx.SourcePermanent()
					if !ok || info.Counters[game.CounterPlusOne] <= 0 {
						return nil
					}
					for _, t := range ctx.LegalTargets() {
						if t.Kind == game.TargetCard {
							return AddCounter{Target: t.ID, Kind: game.CounterPlusOne, N: info.Counters[game.CounterPlusOne]}.Apply(ctx)
						}
					}
					return nil
				}), TargetCreature("target artifact creature", Artifact())),
				"Put Arcbound Wanderer's +1/+1 counters on target artifact creature?"),
		},
	})
}
