package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// free_commander_cast.go — #6242: "you may cast your commander from the
// command zone without paying its mana cost. (You still pay any
// additional costs.)" (Geode Golem). The engine half is
// game/free_cast_command.go.
//
//	Effect: func(g *game.Game, item *game.StackItem) error {
//		OfferFreeCommanderCast(effects.NewContext(g, item), "Geode Golem")
//		return nil
//	}

// OfferFreeCommanderCast offers the resolving effect's controller the
// free cast of a commander they own from their command zone, during the
// resolution (CR 608.2g), through the resolution_cast prompt. The
// commander tax is still paid (CR 903.8), and the cast counts toward it.
// Nothing is asked when no commander of theirs is in the command zone.
func OfferFreeCommanderCast(ctx *Context, name string) {
	g, controller := ctx.Game, ctx.Controller()
	cards := g.OwnCommandersInCommandZoneForEffect(controller)
	if len(cards) == 0 {
		return
	}
	perm := g.FreeCastFromCommandPermission(controller, ctx.Source(), name)
	g.OfferResolutionCastForEffect(game.ResolutionCastOffer{
		Chooser:      controller,
		Source:       ctx.Source(),
		Reason:       name,
		Cards:        cards,
		Permission:   &perm,
		Keyword:      game.MayCastKeywordFreeCast,
		AcceptLabel:  "Cast it free",
		DeclineLabel: "Leave it in the command zone",
	})
}
