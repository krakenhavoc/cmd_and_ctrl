package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Replication Specialist — Creature — Moonfolk Artificer {4}{U}, 3/4
// (Kamigawa: Neon Dynasty, #5493):
//
//	"Flying
//	 Whenever a nontoken artifact you control enters, you may pay {1}{U}.
//	 If you do, create a token that's a copy of that artifact."
//
// Inalla's "pay, then a token copy of the creature that entered" on an
// artifact: the copy is of the artifact as it entered (its copiable
// values, last-known if it has since left, CR 608.2h) and carries no
// extra keywords. A token artifact entering does not trigger it, so a
// copy never copies itself.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "7ce7decb-1438-4c3d-88d2-26b0b3996b33",
		Name:            "Replication Specialist",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			On(game.EventETB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				c, ok := enteredUnderYourControl(ev, source, g, false)
				return ok && c.IsArtifact() && !IsToken(c)
			}, "Replication Specialist — you may pay {1}{U}; if you do, create a token that's a copy of that artifact",
				func(g *game.Game, item *game.StackItem) error {
					if item.Trigger == nil {
						return nil
					}
					return MayPay{
						Chooser:  item.Controller,
						Cost:     "{1}{U}",
						Question: "Replication Specialist — pay {1}{U} to create a token that's a copy of that artifact?",
						Then:     replicationSpecialistPaid,
					}.Apply(NewContext(g, item))
				}),
		},
	})
}

var replicationSpecialistPaid = PayUnlessBranches("pay-unless/replication-specialist-copy",
	func(ctx *Context, _ game.PayUnlessAnswer) error {
		tmpl, ok := TokenCopyTemplate(ctx.Game, ctx.Trigger().Event.CardID)
		if !ok {
			return nil
		}
		_, err := ctx.Game.CreateTokensForEffect(ctx.Controller(), copiedToken(tmpl), 1, game.TokenEntryOptions{})
		return err
	}, nil)
