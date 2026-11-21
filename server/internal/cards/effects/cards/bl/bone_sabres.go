package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Bone Sabres — Artifact — Equipment {2}{G} (Warhammer 40,000 Commander,
// S82 GAP batch 186, #5948):
//
//	"Whenever equipped creature attacks, put four +1/+1 counters on it.
//	 Equip {3}"
//
// The counters go on the creature that attacked, read off the attack event.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "7ef45067-74bd-4df8-9f95-621761a0b4a0",
		Name:         "Bone Sabres",
		Completeness: effects.CompletenessFull,
		Activated:    []effects.ActivatedAbility{effects.EquipAbility("{3}")},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return effects.AttachedCreatureAttacked(ev, source)
			}, "Bone Sabres — put four +1/+1 counters on it",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					id := ctx.Trigger().Event.CardID
					if !effects.OnBattlefield(g, id) {
						return nil
					}
					return effects.AddCounter{Target: id, Kind: game.CounterPlusOne, N: 4}.Apply(ctx)
				}),
		},
	})
}
