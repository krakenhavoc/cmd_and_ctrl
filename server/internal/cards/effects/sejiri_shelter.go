package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Sejiri Shelter // Sejiri Glacier — Modal double-faced card (EDHREC
// top 1,000, #3654). The front face:
//
//	"Target creature you control gains protection from the color of your
//	 choice until end of turn." {1}{W} Instant
//
// Mother of Runes' colour choice as a spell: the chooser names the
// colour at resolution (CR 608.2d), so a creature that left in response
// asks nothing. The land back face is the shared mdfc_lands.go cycle.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "d54e4e37-042b-44a5-918d-757308545d4d",
		Name:         "Sejiri Shelter",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature you control", YouControl()),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			id, ok := b16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			chooseProtectionColorFor(ctx, item.Controller, id, "Sejiri Shelter", "Sejiri Shelter — choose a color")
			return nil
		},
	})
}
