package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

const phabineLabel = "Phabine, Boss's Confidant — parley: each player reveals the top card of their library; you create a 1/1 Citizen for each land and your creatures get +1/+1 for each nonland card; then each player draws a card"

// Phabine, Boss's Confidant — Legendary Creature — Cat Advisor
// {3}{R}{G}{W}, 3/6 (Streets of New Capenna Commander, #5491):
//
//	"Creature tokens you control have haste.
//	 Parley — At the beginning of combat on your turn, each player reveals
//	 the top card of their library. For each land card revealed this way,
//	 you create a 1/1 green and white Citizen creature token. Then
//	 creatures you control get +1/+1 until end of turn for each nonland
//	 card revealed this way. Then each player draws a card."
//
// The haste is a layer 6 grant to the creature tokens Phabine's controller
// controls (Echoing Assault's scope). The parley is Storm Fleet
// Negotiator's, with both counts: every player reveals in turn order, each
// land makes a Citizen for Phabine's controller, and then the pump lands on
// every creature that player controls, the new Citizens included, as the
// set is read when it is applied (CR 611.2c). Each player draws a card
// last. A player with an empty library reveals nothing, and a card that
// is both is a land.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "70697ad7-d4bb-4b63-b895-bbb69e9d04e1",
		Name:         "Phabine, Boss's Confidant",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Draws: 1, Tokens: 1, CreatureTokens: 1},
		Static: []game.StaticAbility{
			Labelled(b16GrantKeywords(creatureTokensYouControl, "haste"), "Creature tokens you control have haste."),
		},
		Triggered: []game.TriggeredAbility{
			AtBeginningOfYourCombat(phabineLabel, func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				return snc17Parley(g, ctx.Source(), phabineLabel, func(lands, nonland int) error {
					if lands > 0 {
						if err := (CreateToken{Template: TokenCard("1/1 green and white Citizen"), N: lands}).Apply(ctx); err != nil {
							return err
						}
					}
					if nonland == 0 {
						return nil
					}
					return BoostUntilEOT{
						Match:     And(Creature(), YouControl()),
						Power:     nonland,
						Toughness: nonland,
						Label:     "Phabine, Boss's Confidant — +1/+1 for each nonland card revealed",
					}.Apply(ctx)
				})
			}),
		},
	})
}
