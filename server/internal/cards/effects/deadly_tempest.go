package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Deadly Tempest — Sorcery {4}{B}{B}:
//
//	"Destroy all creatures. Each player loses life equal to the
//	 number of creatures they controlled that were destroyed this
//	 way."
//
// Fumigate's mirror: instead of paying you for the wipe, it charges
// everyone for the board they had. At a table where one player has
// gone wide it is a wrath that also removes a third of their life
// total, which is why a black control deck wants this over Damnation
// despite two more mana.
//
// The tally is per CONTROLLER and it must be taken BEFORE anything
// moves — once a creature is in a graveyard its battlefield
// controller is no longer a question you can ask. That is what the
// Then callback's `swept` argument is: the pre-move copies. Counting
// after the sweep, or counting creatures in graveyards, both give
// the wrong answer the moment a token is involved.
//
// Life LOSS, not damage: no prevention, no lifelink, no source
// dealing it, and a player at 3 who controlled five creatures is
// dead on the state-based-action check.
func init() {
	Register(Spec{
		OracleID:     "b5516bc9-ec8d-4323-8748-96c49d7d0622",
		Name:         "Deadly Tempest",
		Completeness: CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return DestroyAllMatching{
				Match: Creature(),
				Then: func(ctx *Context, swept []game.Card, _ int) error {
					tally := controllersOf(swept)
					// Seat order, not map order: the life losses are
					// simultaneous in the rules but sequential in the
					// event log, and a log whose order changes run to
					// run is a replay that does not reproduce.
					for _, p := range ctx.Game.Seats {
						if p == nil || tally[p.ID] <= 0 {
							continue
						}
						if err := ctx.Game.ChangePlayerLifeForEffect(ctx.Source(), p.ID, -tally[p.ID]); err != nil {
							return err
						}
					}
					return nil
				},
			}.Apply(ctx)
		},
	})
}
