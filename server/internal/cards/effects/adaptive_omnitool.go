package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Adaptive Omnitool — Artifact — Equipment {2}:
//
//	"Equipped creature gets +1/+1 for each artifact you control.
//	 Whenever equipped creature attacks, look at the top six cards of your
//	 library. You may reveal an artifact card from among them and put it into
//	 your hand. Put the rest on the bottom of your library in a random order.
//	 Equip {3}"
//
// The pump counts the artifacts its controller controls, the Omnitool itself
// and the equipped creature if it is one included, read live
// (PumpAttachedPer). The attack trigger is the look-and-take of Monumental
// Henge (takeRevealRestBottomRandom): only the artifact taken is revealed.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "e91240bc-4a13-4131-9379-5ddac6555f4a",
		Name:         "Adaptive Omnitool",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			PumpAttachedPer(1, 1, func(g *game.Game, source *game.Card) int {
				return permanentsYouControl(g, source, func(c *game.Card) bool { return c.IsArtifact() })
			}),
		},
		Triggered: []game.TriggeredAbility{
			On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return attachedCreatureAttacked(ev, source)
			}, "Adaptive Omnitool — look at the top six cards, you may reveal an artifact card and put it into your hand, the rest on the bottom in a random order",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					return TakeFromLibraryToHand{
						Player:   item.Controller,
						Cards:    g.LookAtTopOfLibraryForEffect(item.Controller, 6),
						Match:    Artifact(),
						Max:      1,
						Optional: true,
						Label:    "Adaptive Omnitool — you may reveal an artifact card from among them and put it into your hand",
						Key:      takeRevealRestBottomRandom,
					}.Apply(ctx)
				}),
		},
		Activated: []ActivatedAbility{EquipAbility("{3}")},
	})
}
