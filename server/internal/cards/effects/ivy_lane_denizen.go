package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ivy Lane Denizen — Creature — Elf Warrior {3}{G}, 2/3 (EDHREC rank
// 2462):
//
//	"Whenever another green creature you control enters, put a +1/+1
//	 counter on target creature."
//
// The Elf deck's counter engine. One targeted trigger per green
// creature that enters under the controller's control — the Denizen
// itself excluded, a token included, and "green" read off the
// creature's effective colours — with "target creature" the S20
// clause: any creature on the battlefield, the entering one and the
// Denizen itself both legal picks, chosen when the trigger goes on
// the stack. The counter goes on through AddCounter, so a doubler
// applies.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "a5da5ad6-4ed2-4041-a983-76a8c87fa109",
		Name:         "Ivy Lane Denizen",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventETB},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return b23AnotherGreenCreatureYouControlEntered(ev, source, g)
			},
			Targets: TargetCreature("target creature"),
			Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				return game.NewTriggeredItem(source, "Ivy Lane Denizen — put a +1/+1 counter on target creature",
					func(g *game.Game, item *game.StackItem) error {
						ctx := NewContext(g, item)
						for _, t := range ctx.LegalTargets() {
							if err := (AddCounter{Target: t.ID, Kind: "+1/+1", N: 1}).Apply(ctx); err != nil {
								return err
							}
						}
						return nil
					})
			},
		}},
	})
}
