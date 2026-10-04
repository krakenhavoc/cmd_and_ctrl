package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Folio of Fancies — Artifact — Book {1}{U}:
//
//	"Players have no maximum hand size.
//	 {X}{X}, {T}: Each player draws X cards.
//	 {2}{U}, {T}: Each opponent mills cards equal to the number of cards
//	 in their hand."
//
// "Players have no maximum hand size" reaches every player (ADR 0113
// §3, #2074), folded in CR 613.11 timestamp order. {X}{X} is paid as
// twice X (the 2019-10-04 ruling) and read back with ctx.X(); each
// player draws in turn order from the active player, and a player who
// can't draw them all loses as usual. The mill counts each opponent's
// hand as the ability resolves.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "9030d508-5cfe-46ed-954b-21d970bc3b5a",
		Name:         "Folio of Fancies",
		Completeness: CompletenessFull,
		HandSize:     []game.HandSizeStatic{PlayersHaveNoMaxHandSize()},
		Activated: []ActivatedAbility{
			{
				Label: "{X}{X}, {T}: Each player draws X cards.",
				Cost:  Plus(ManaCost("{X}{X}"), TapCost()),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return b05EachPlayerDraws(g, item, NewContext(g, item).X())
				},
			},
			{
				Label: "{2}{U}, {T}: Each opponent mills cards equal to the number of cards in their hand.",
				Cost:  Plus(ManaCost("{2}{U}"), TapCost()),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					for _, p := range livePlayers(g) {
						if p.ID == item.Controller {
							continue
						}
						if err := (MillCards{Player: p.ID, N: b14HandSize(g, p.ID)}).Apply(ctx); err != nil {
							return err
						}
					}
					return nil
				},
			},
		},
	})
}
