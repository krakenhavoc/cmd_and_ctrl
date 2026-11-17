package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Teething Wurmlet — Creature — Wurm {G}, 1/1 (The Brothers' War, #5482):
//
//	"This creature has deathtouch as long as you control three or more
//	 artifacts.
//	 Whenever an artifact you control enters, you gain 1 life. If this is
//	 the first time this ability has resolved this turn, put a +1/+1
//	 counter on this creature."
//
// The deathtouch is a layer 6 grant to itself counting artifacts on the
// post-layer type line (the inputs are all permanents, so no recompute
// flag). The life is gained on every resolution; the counter only on the
// first this turn, which is Belladonna Took's per-object resolution tally
// (Game.ResolvedThisTurn, which includes the one resolving).
//
// No simplification.
func init() {
	deathtouch := KeywordGrant(func(target *game.Card, g *game.Game, source *game.Card) bool {
		return target.InstanceID == source.InstanceID && b15ArtifactsControlled(g, source.Controller) >= 3
	}, "deathtouch")
	deathtouch.Label = "Has deathtouch as long as you control three or more artifacts."
	Register(Spec{
		OracleID:     "a44ceede-9934-4d4e-8b8c-e087b3ef74d4",
		Name:         "Teething Wurmlet",
		Completeness: CompletenessFull,
		Static:       []game.StaticAbility{deathtouch},
		Triggered: []game.TriggeredAbility{
			On(game.EventETB, ArtifactEnteredUnderYourControl, teethingWurmletLabel,
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					if err := (GainLife{Player: item.Controller, Amount: 1}).Apply(ctx); err != nil {
						return err
					}
					if g.ResolvedThisTurn(item.SourceCardID, teethingWurmletLabel) != 1 {
						return nil
					}
					return CounterOnThis{Kind: game.CounterPlusOne, N: 1}.Apply(ctx)
				}),
		},
	})
}

const teethingWurmletLabel = "Teething Wurmlet — gain 1 life; if this is the first time this ability has resolved this turn, put a +1/+1 counter on this creature"
