package cardss

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Sapling of Colfenor — Legendary Creature — Treefolk Shaman {3}{B/G}{B/G},
// 2/5 (Commander 2021):
//
//	"Indestructible
//	 Whenever Sapling of Colfenor attacks, reveal the top card of your
//	 library. If it's a creature card, you gain life equal to that card's
//	 toughness, lose life equal to its power, then put it into your hand."
//
// The reveal is not a draw (no draw trigger sees it). The card's power and
// toughness are read in the library, where a characteristic-defining
// ability still applies (CR 604.3), before it moves; the two life changes
// happen in printed order (gain, then lose) and the card goes to hand last.
// A noncreature top card is revealed and stays on top. An empty library
// reveals nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "0c1e2bb8-0629-466c-b435-c4c71cdc755d",
		Name:            "Sapling of Colfenor",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"indestructible"},
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisAttacks("Sapling of Colfenor — reveal the top card of your library",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					var revealed []uuid.UUID
					if err := (effects.RevealTopOfLibrary{
						Player:   ctx.Controller(),
						N:        1,
						Reason:   "Sapling of Colfenor — reveal the top card of your library",
						Revealed: &revealed,
					}).Apply(ctx); err != nil {
						return err
					}
					if len(revealed) == 0 {
						return nil
					}
					c, ok := g.LookupCardForEffect(revealed[0])
					if !ok || !c.IsCreature() {
						return nil
					}
					toughness, power := g.ToughnessAnywhereForEffect(c), g.PowerAnywhereForEffect(c)
					if toughness > 0 {
						if err := (effects.GainLife{Player: ctx.Controller(), Amount: toughness}).Apply(ctx); err != nil {
							return err
						}
					}
					if power > 0 {
						if err := g.ChangePlayerLifeForEffect(ctx.Source(), ctx.Controller(), -power); err != nil {
							return err
						}
					}
					return effects.BounceToHand{Target: revealed[0]}.Apply(ctx)
				}),
		},
	})
}
