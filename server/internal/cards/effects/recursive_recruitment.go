package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Recursive Recruitment — Sorcery {2}{U}{B} (Reality Fracture, tracker #2795):
//
//	"Create two 2/2 colorless Wizard Soldier creature tokens named
//	 Cadet. If this spell was cast from a graveyard, put a +1/+1 counter
//	 on each of them for every three cards in your graveyard.
//	 Flashback {6}{U}{B}"
//
// The counters ride the tokens' entry (as printed they are put on after,
// but nothing can act between the two). The graveyard is counted as the
// spell resolves, and the flashed-back card is on the stack, not in it.
// "Cast from a graveyard" reads CastFromZone, as Increasing Vengeance does.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:         "b73e35b0-b8c9-471a-bdff-f7dc8f453d7a",
		Name:             "Recursive Recruitment",
		Completeness:     CompletenessFull,
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{Flashback("{6}{U}{B}")},
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			opts := game.TokenEntryOptions{}
			if item.CastFromZone == game.ZoneGraveyard {
				if p := ctx.PlayerByID(ctx.Controller()); p != nil && p.Graveyard != nil {
					if n := len(p.Graveyard.Cards) / 3; n > 0 {
						opts.Counters = map[string]int{"+1/+1": n}
					}
				}
			}
			_, err := ctx.Game.CreateTokensForEffect(ctx.Controller(), TokenCard("2/2 colorless Wizard Soldier named Cadet"), 2, opts)
			return err
		},
	})
}
