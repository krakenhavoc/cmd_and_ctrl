package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Thran Vigil — Enchantment {1}{B} (The Brothers' War, #5482):
//
//	"Whenever one or more artifact and/or creature cards leave your
//	 graveyard during your turn, put a +1/+1 counter on target creature you
//	 control."
//
// Attuned Hunter's trigger (left_graveyard.go) over artifact and/or creature
// cards: every route a card leaves your graveyard by counts (an unearth, an
// exile, a return, a cast from it), cards leaving together trigger it once
// (OncePerBatch, CR 603.2c), and the type is the card's as it was in the
// graveyard (CR 603.10a). The target is chosen as the trigger goes on the
// stack and is re-checked as it resolves (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "01c179aa-533e-47e1-b6e5-c35587a39680",
		Name:         "Thran Vigil",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(OncePerBatch(OnAny(leftGraveyardEvents, s82Bro13ArtifactOrCreatureCardLeftYourGraveyard,
				"Thran Vigil — put a +1/+1 counter on target creature you control",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					target, ok := ctx.ClauseTarget(0)
					if !ok {
						return nil
					}
					return AddCounter{Target: target.ID, Kind: game.CounterPlusOne, N: 1}.Apply(ctx)
				})),
				TargetCreature("target creature you control", YouControl())),
		},
	})
}
