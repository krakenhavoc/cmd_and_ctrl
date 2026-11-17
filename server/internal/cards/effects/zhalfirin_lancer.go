package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Zhalfirin Lancer — Creature — Human Knight {2}{W}, 3/3 (March of the
// Machine):
//
//	"Whenever another Knight you control enters, this creature gets +1/+1
//	 and gains vigilance until end of turn."
func init() {
	const label = "Zhalfirin Lancer — +1/+1 and vigilance until end of turn"
	Register(Spec{
		OracleID:     "5adcd0cb-9305-4e77-a117-31e9dca5c96d",
		Name:         "Zhalfirin Lancer",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventETB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return s82Mom09AnotherKnightEntered(ev, source, g)
			}, label, func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				if err := (BoostUntilEOT{
					Target:    ctx.Source(),
					Power:     1,
					Toughness: 1,
					Label:     label,
				}).Apply(ctx); err != nil {
					return err
				}
				return GrantKeywordUntilEOT{
					Target:   ctx.Source(),
					Keywords: []string{"vigilance"},
					Label:    label,
				}.Apply(ctx)
			}),
		},
	})
}
