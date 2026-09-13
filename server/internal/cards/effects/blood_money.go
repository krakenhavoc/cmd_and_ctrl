package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Blood Money — Sorcery {5}{B}{B} (EDHREC rank 1652):
//
//	"Destroy all creatures. For each nontoken creature destroyed
//	 this way, you create a tapped Treasure token."
//
// A seven-mana Damnation that pays you a Treasure per real creature
// it killed, so the turn after the wipe is the biggest of the game.
//
// Not the batched sweep. The Treasure count has to match what
// ACTUALLY died, and the mass-destroy path does not check
// indestructible (#446) — through it an Avacyn board would die and
// pay out, stronger than printed on both counts. Each creature is
// instead destroyed through the single-permanent verb, which honours
// indestructible, and counted only if it left the battlefield (a
// commander tucked away by CR 903.9 was destroyed and counts, as
// Fumigate notes; a token was destroyed and does not, as printed).
//
// Sandbox simplification: the creatures leave one at a time rather
// than as one simultaneous event, so a "whenever another creature
// dies" watcher that is itself in the wipe sees only the creatures
// destroyed before it. Weaker than printed for that watcher's
// controller, never stronger.
func init() {
	Register(Spec{
		OracleID:     "75f5d372-4ff9-430c-8302-72472439e0d2",
		Name:         "Blood Money",
		Completeness: CompletenessCaveats,
		Caveats:      []string{"The creatures are destroyed one after another rather than all at once, so a creature with a \"whenever another creature dies\" ability that is itself destroyed may miss some of the deaths."},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			paid := 0
			for _, c := range MatchingBattlefield(ctx, Creature()) {
				if err := (DestroyTarget{Target: c.InstanceID}).Apply(ctx); err != nil {
					return err
				}
				if !IsToken(c) && !ctx.Game.Battlefield.Contains(c.InstanceID) {
					paid++
				}
			}
			return b13CreateTappedTreasures(ctx, ctx.Controller(), paid)
		},
	})
}
