package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// std_b27_helpers.go — shared pieces of the Standard leftovers redo set
// (feat/std-redo-27).

// CondYourGraveyardAtLeast is "if there are N or more cards in your
// graveyard" (threshold, N = 7) declared as an intervening if
// (InterveningIf): asked as the ability triggers and again as it
// resolves (CR 603.4). The graveyard is the resolving item's
// controller's, which for an ability that works from a graveyard is the
// card's owner (CR 108.4a).
type CondYourGraveyardAtLeast struct{ N int }

func (c CondYourGraveyardAtLeast) Holds(ctx *Context) bool {
	return b31GraveyardSize(ctx.Game, ctx.Controller()) >= c.N
}

func (CondYourGraveyardAtLeast) condition() (game.FootprintStep, bool) {
	return game.FootprintStep{}, false
}

// --- Vren, the Relentless's Rat ------------------------------------

// VrenRatToken is the 1/1 black Rat creature token with "This token gets
// +1/+1 for each other Rat you control" (Vren, the Relentless; the printed
// Rat token of Bloomburrow, tblb #13).
func VrenRatToken() game.Card { return tokenFromCatalog(printedVrenRatToken) }

// printedVrenRatToken is that Rat as PRINTED. Its slug is an on-disk
// identity and is never renamed.
func printedVrenRatToken() tokenTemplate {
	return tokenTemplate{
		Slug: "rat-pumped-by-other-rats",
		Card: game.Card{
			Name:      "Rat",
			TypeLine:  "Token Creature — Rat",
			Power:     1,
			Toughness: 1,
			Colors:    []string{"B"},
		},
		Static: []game.StaticAbility{{
			Label:    "This token gets +1/+1 for each other Rat you control.",
			Layer:    game.Layer7PT,
			SubLayer: game.SubLayer7C_Modify,
			AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.InstanceID == source.InstanceID
			},
			Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
				n := hobA9OtherSubtypeYouControl(g, source, source.TextSubtype("Rat"))
				c.Power += n
				c.Toughness += n
			},
		}},
		Text: "This token gets +1/+1 for each other Rat you control.",
	}
}
