package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Cloud, Ex-SOLDIER — Legendary Creature — Human Soldier Mercenary
// {2}{R}{G}{W}, 4/4 (#3656):
//
//	"Haste
//	 When Cloud enters, attach up to one target Equipment you control
//	 to it.
//	 Whenever Cloud attacks, draw a card for each equipped attacking
//	 creature you control. Then if Cloud has power 7 or greater,
//	 create two Treasure tokens."
//
// The enter trigger is Hammer of Nazahn's attach (AttachForEffect) with
// an "up to one" clause, so no Equipment is a legal choice. The attack
// trigger counts at resolution: the equipped attackers still attacking
// then, Cloud included when he is equipped. His power is read as the
// ability resolves, after the draw, as printed.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "653f0452-5802-4639-86c2-1500c3964fcb",
		Name:            "Cloud, Ex-SOLDIER",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"haste"},
		Triggered: []game.TriggeredAbility{
			Targeting(
				WhenThisEnters("Cloud, Ex-SOLDIER — attach up to one target Equipment you control to Cloud",
					func(g *game.Game, item *game.StackItem) error {
						ctx := NewContext(g, item)
						for _, t := range ctx.LegalTargets() {
							if t.Kind != game.TargetCard {
								continue
							}
							if err := g.AttachForEffect(t.ID, game.TargetRef{Kind: game.TargetCard, ID: item.SourceCardID}); err != nil {
								return err
							}
						}
						return nil
					}),
				TargetPermanent("up to one target Equipment you control", HasSubtype("Equipment"), YouControl()).WithCount(0, 1)),
			WheneverThisAttacks("Cloud, Ex-SOLDIER — draw a card for each equipped attacking creature you control",
				Do(
					// The Treasures wait for a prompt the draws ask (CR 608.2c, #4155).
					Instruction(func(ctx *Context) error {
						n := equippedCreaturesYouControl(ctx.Game, ctx.Item.Controller, true)
						return DrawCards{N: n}.Apply(ctx)
					}),
					Instruction(func(ctx *Context) error {
						item := ctx.Item
						cloud, ok := ctx.Game.LookupCardForEffect(item.SourceCardID)
						power := b31CurrentPowerOf(ctx.Game, item.SourceCardID)
						if !ok || !cloud.IsCreature() || power < 7 {
							return nil
						}
						return CreateToken{Template: TreasureToken(), N: 2}.Apply(ctx)
					}),
				)),
		},
	})
}
