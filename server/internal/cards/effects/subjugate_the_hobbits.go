package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Subjugate the Hobbits — Sorcery {5}{U}{U} (Tales of Middle-earth Commander):
//
//	"Gain control of each noncommander creature with mana value 3 or
//	 less."
//
// A table-wide control change with no stated duration (CR 611.2a,
// indefinite), run once per creature on the battlefield as the spell
// resolves, snapshotted before anything moves (CR 611.2c: the set is
// locked in). "Noncommander" reads Card.IsCommander (CR 903.3), so a
// commander of any player is left alone; the caster's own creatures
// are taken too, a no-op as in paper.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "eb56141a-eb43-40e9-8ff1-e0ef5424f6a4",
		Name:         "Subjugate the Hobbits",
		Completeness: CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			for _, c := range MatchingBattlefield(ctx, And(Creature(), Not(ACommander()), ManaValueLE(3))) {
				if err := (GainControl{
					Target:   c.InstanceID,
					Duration: game.IndefiniteDuration(),
					Label:    "Subjugate the Hobbits — gain control (no stated duration)",
				}).Apply(ctx); err != nil {
					return err
				}
			}
			return nil
		},
	})
}
