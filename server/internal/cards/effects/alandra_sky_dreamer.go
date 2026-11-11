package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Alandra, Sky Dreamer — Legendary Creature — Merfolk Wizard {2}{U}{U},
// 2/4 (Murders at Karlov Manor Commander, #5368):
//
//	"Whenever you draw your second card each turn, create a 2/2 blue
//	 Drake creature token with flying.
//	 Whenever you draw your fifth card each turn, Alandra and Drakes you
//	 control each get +X/+X until end of turn, where X is the number of
//	 cards in your hand."
//
// Two draw-tally triggers: the tally is read after the draw, so each
// fires on its own draw and no other. The pump reads X as it resolves
// and locks the set it applies to then (CR 611.2c): Alandra and the
// Drakes you control at that moment.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "97bb1589-2941-4a75-a5f6-916871e51e9e",
		Name:         "Alandra, Sky Dreamer",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventDrawCard, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return b41DrewYourSecondCardThisTurn(ev, source, g)
			}, "Alandra — create a 2/2 blue Drake creature token with flying",
				Do(CreateToken{Template: TokenCard("2/2 blue Drake with flying"), N: 1})),
			On(game.EventDrawCard, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return cmdr2S22DrewYourFifthCardThisTurn(ev, source, g)
			}, "Alandra — Alandra and Drakes you control each get +X/+X until end of turn, where X is the number of cards in your hand",
				cmdr2S22AlandraAndDrakesGetCardsInHand),
		},
	})
}
