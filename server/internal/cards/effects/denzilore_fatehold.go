package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Denzilore Fatehold — Legendary Creature — Elder Sphinx {1}{W}{U}{U}, 3/4:
//
//	"Flash
//	 Flying
//	 Whenever you scry or surveil, put a +1/+1 counter on each
//	 creature you control."
//
// One ability watching two events (EventScry, EventSurveil), each
// emitted once the scry or surveil is finished. The creatures are
// those you control as the trigger resolves, Denzilore included.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "0f8de7f0-c61d-415d-b4db-4959374343da",
		Name:            "Denzilore Fatehold",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash", "flying"},
		Triggered: []game.TriggeredAbility{
			OnAny([]game.EventKind{game.EventScry, game.EventSurveil}, ByYou,
				"Denzilore Fatehold — put a +1/+1 counter on each creature you control",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					g.RecomputeLayersIfStaleLocked()
					var ids []game.TargetRef
					for _, c := range g.Battlefield.Cards {
						if c.Controller == item.Controller && c.IsCreature() {
							ids = append(ids, game.TargetRef{Kind: game.TargetCard, ID: c.InstanceID})
						}
					}
					for _, t := range ids {
						if err := (AddCounter{Target: t.ID, Kind: game.CounterPlusOne, N: 1}).Apply(ctx.asGroupMember()); err != nil {
							return err
						}
					}
					return nil
				}),
		},
	})
}
