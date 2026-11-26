package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Destructive Force — Sorcery {5}{R}{R} (Magic 2011, #5579):
//
//	"Each player sacrifices five lands of their choice. Destructive Force
//	 deals 5 damage to each creature."
//
// Every living player is asked, in APNAP order, for five lands (fewer when
// they control fewer; a player with none is skipped), and the choices are
// sacrificed together (CR 101.4, 701.21a). The damage waits for the
// sacrifices (a keyed run continuation, so a land that is also a creature
// is gone before the damage, CR 608.2c) and is dealt to every creature as
// one instance, the spell being the source.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b0967525-554b-41eb-b639-75fe36e0fe57",
		Name:         "Destructive Force",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDamage, Amount: 5}},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return ctx.PlayersSacrificeThen(effects.S82OldB475SeatsOwing(ctx.Game, 5),
				effects.SacrificeSpec("a land", effects.Land()), "Destructive Force — sacrifice five lands",
				destructiveForceDamage, ctx.RunData(nil))
		},
	})
}

// destructiveForceDamage is "Destructive Force deals 5 damage to each
// creature", run once the lands have left. An on-disk key: never renamed,
// never reused.
var destructiveForceDamage = effects.RunThen("prompt-run/destructive-force-damage", func(ctx *effects.Context, _ game.RunDone) error {
	return effects.DamageEachMatching(ctx, effects.Creature(), 5)
})
