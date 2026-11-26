package cardsd

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Dakra Mystic — Creature — Merfolk Wizard {U}, 1/1 (Journey into Nyx, S82 OLD,
// #5558):
//
//	"{U}, {T}: Each player reveals the top card of their library. You may put
//	 the revealed cards into their owners' graveyards. If you don't, each player
//	 draws a card."
//
// Every player reveals in turn order (Cutthroat Negotiator's reveal), then you
// choose once for all of the revealed cards: yes puts each into its owner's
// graveyard, no has each player draw a card. The cards are the ones revealed,
// so a draw never changes which cards a yes would mill. A player with an empty
// library reveals nothing, and the draw from it is the CR 704.5b loss as usual.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "63f05280-74de-4fe3-a507-07fb125685c0",
		Name:         "Dakra Mystic",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{U}, {T}: Each player reveals the top card of their library. You may put the revealed cards into their owners' graveyards. If you don't, each player draws a card.",
			Purpose: game.Purpose{Answers: game.AnswerValue, Draws: 1},
			Cost:    effects.Plus(effects.ManaCost("{U}"), effects.TapCost()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				var revealed []uuid.UUID
				for _, p := range effects.ApnapPlayers(g) {
					revealed = append(revealed, g.RevealTopOfLibraryForEffect(p, item.SourceCardID, 1, dakraMysticLabel)...)
				}
				return effects.MayChoice{
					Question: "Dakra Mystic — put the revealed cards into their owners' graveyards? (If you don't, each player draws a card.)",
					YesLabel: "Put them into graveyards",
					NoLabel:  "Each player draws",
					Then:     dakraMysticAnswer,
					Carry:    revealed,
				}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}

const dakraMysticLabel = "Dakra Mystic — each player reveals the top card of their library"

// dakraMysticAnswer is the choice: yes puts the revealed cards into their
// owners' graveyards, no has each player draw. An on-disk key.
var dakraMysticAnswer = effects.MayYesNo("confirm/s82-old-b371-dakra-mystic",
	func(ctx *effects.Context, r game.ConfirmAnswer) error {
		for _, id := range r.Carry {
			if err := ctx.Game.PutIntoGraveyardForEffect(id); err != nil {
				return err
			}
		}
		return nil
	},
	func(ctx *effects.Context, _ game.ConfirmAnswer) error {
		for _, p := range effects.ApnapPlayers(ctx.Game) {
			if err := ctx.Game.DrawNForEffect(p, 1); err != nil {
				return err
			}
		}
		return nil
	})
