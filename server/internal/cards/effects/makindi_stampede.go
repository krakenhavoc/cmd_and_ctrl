package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Makindi Stampede // Makindi Mesas — Modal DFC. Front: Sorcery
// {3}{W}{W}: "Creatures you control get +2/+2 until end of turn." Back:
// Land (mdfc_lands.go).
//
// The set is the creatures you control as it resolves, pinned there
// (CR 611.2c): one that enters later this turn does not get the bonus.
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "342e08f9-d4d0-4408-8621-66e087058616",
		Name:         "Makindi Stampede",
		Completeness: CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return BoostUntilEOT{
				Match:     And(Creature(), YouControl()),
				Power:     2,
				Toughness: 2,
				Label:     "Makindi Stampede — +2/+2",
			}.Apply(ctx)
		},
	})
}
