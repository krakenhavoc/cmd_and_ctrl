package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Mirrorweave — Instant {2}{W/U}{W/U}:
//
//	"Each other creature becomes a copy of target nonlegendary creature
//	 until end of turn."
//
// One duration copy over many permanents (#1593, become_copy.go): the
// affected set is every creature other than the target as the spell
// resolves, locked then (CR 611.2c), so a creature that enters later in
// the turn is itself. The copy is one record with one timestamp, and it
// ends for every creature at the same cleanup.
//
// Only the target has to be nonlegendary. A legendary creature among
// the others becomes a copy like any of them.
func init() {
	Register(Spec{
		OracleID:     "026b7221-0caf-4c8b-8c1b-e7de836797b7",
		Name:         "Mirrorweave",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target nonlegendary creature", Not(Legendary())),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			targets := ctx.LegalTargets()
			if len(targets) == 0 {
				return nil
			}
			model := targets[0].ID
			var others []uuid.UUID
			for _, c := range ctx.Game.BattlefieldCardsForEffect() {
				if c.IsCreature() && c.InstanceID != model {
					others = append(others, c.InstanceID)
				}
			}
			return BecomeCopy{
				Targets: others,
				Of:      model,
				Label:   "Mirrorweave — becomes a copy until end of turn",
			}.Apply(ctx)
		},
	})
}
