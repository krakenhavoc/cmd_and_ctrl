package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Abyssal Nocturnus — Creature — Horror {1}{B}{B}, 2/2 (Guildpact, #5599):
//
//	"Whenever an opponent discards a card, this creature gets +2/+2 and
//	 gains fear until end of turn. (It can't be blocked except by
//	 artifact creatures and/or black creatures.)"
//
// Liliana's Caress's trigger (b18OpponentDiscarded): once per card
// discarded, by any player other than the controller, whatever made them
// discard. Each resolution is one until-end-of-turn effect on the
// Nocturnus itself, +2/+2 in layer 7c and fear in layer 6, so two
// discards are +4/+4 and fear twice, which is redundant (CR 702.36c). A
// Nocturnus that left and came back is a new object and gets nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "b79c4066-66a9-4a53-8716-23f7a340a85f",
		Name:         "Abyssal Nocturnus",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventDiscardCard, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return b18OpponentDiscarded(ev, source)
			}, "Abyssal Nocturnus — +2/+2 and fear until end of turn", func(g *game.Game, item *game.StackItem) error {
				return untilEndOfTurn(NewContext(g, item), item.SourceCardID, nil,
					"Abyssal Nocturnus — +2/+2 and fear", game.ModifyPTMod(2, 2), game.AddKeywordsMod("fear"))
			}),
		},
	})
}
