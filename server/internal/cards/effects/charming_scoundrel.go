package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Charming Scoundrel — Creature — Human Rogue {1}{R}, 1/1:
//
//	"Haste
//	 When this creature enters, choose one —
//	 • Discard a card, then draw a card.
//	 • Create a Treasure token.
//	 • Create a Wicked Role token attached to target creature you
//	 control. (Enchanted creature gets +1/+0. When this token is put
//	 into a graveyard, each opponent loses 1 life.)"
//
// A modal ETB (CR 603.3c) like Charming Prince's: the third bullet's
// target clause rides its own ModeOption, so choosing another bullet
// asks for no target. The first bullet discards before it draws, and
// the draw happens even from an empty hand (nothing in "then" is
// conditional on the discard).
//
// No simplifications.
func init() {
	scoundrel := WhenThisEnters("Charming Scoundrel — choose one",
		func(*game.Game, *game.StackItem) error { return nil })
	scoundrel.Modes = ChooseOne(
		ModeDoing("Discard a card, then draw a card.", nil, func(item *game.StackItem, ctx *Context, _ int) error {
			player := item.Controller
			return ctx.Game.PlayerDiscardsThenForEffect(game.DiscardPrompt{
				Player: player,
				Source: item.SourceCardID,
				N:      1,
			}, func(g *game.Game, _ game.PromptedDiscards) error {
				return g.DrawNForEffect(player, 1)
			})
		}),
		ModeDoing("Create a Treasure token.", nil, func(item *game.StackItem, ctx *Context, _ int) error {
			return CreateToken{Controller: item.Controller, Template: TreasureToken(), N: 1}.Apply(ctx)
		}),
		ModeDoing("Create a Wicked Role token attached to target creature you control.",
			TargetCreature("target creature you control", YouControl()),
			func(item *game.StackItem, ctx *Context, occ int) error {
				t, ok := ModeTarget(ctx, occ)
				if !ok {
					return nil
				}
				return CreateRoleToken{Role: RoleWicked, Host: t.ID}.Apply(ctx)
			}),
	)
	Register(Spec{
		OracleID:        "9eec304d-ffcb-4287-b877-280ac1e4f496",
		Name:            "Charming Scoundrel",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"haste"},
		Triggered:       []game.TriggeredAbility{scoundrel},
	})
}
