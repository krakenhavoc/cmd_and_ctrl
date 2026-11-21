package cardskr

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Light of the Legion — Creature — Angel {4}{W}{W}, 5/5:
//
//	"Flying
//	 Mentor (Whenever this creature attacks, put a +1/+1 counter on target
//	 attacking creature with lesser power.)
//	 When this creature dies, put a +1/+1 counter on each white creature
//	 you control."
//
// Mentor is the shared Mentor() trigger. The dies trigger snapshots the
// white creatures its controller controls at resolution (the Light has
// already left, so it is not among them) and places one counter on each,
// Ajani the Greathearted's group walk.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "79fc0d30-9b9b-400c-a614-c3a1f90a5380",
		Name:            "Light of the Legion",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", "mentor"},
		Triggered: []game.TriggeredAbility{
			effects.Mentor("Light of the Legion — mentor"),
			effects.WhenThisDies("Light of the Legion — put a +1/+1 counter on each white creature you control",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					var ids []uuid.UUID
					for _, c := range g.BattlefieldCardsForEffect() {
						if c.Controller == item.Controller && c.IsCreature() && c.HasColorWord("W") {
							ids = append(ids, c.InstanceID)
						}
					}
					for _, id := range ids {
						if err := (effects.AddCounter{Target: id, Kind: game.CounterPlusOne, N: 1}).Apply(ctx.AsGroupMember()); err != nil {
							return err
						}
					}
					return nil
				}),
		},
	})
}
