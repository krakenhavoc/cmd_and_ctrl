package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Pongify — Instant {U}:
//
//	"Destroy target creature. It can't be regenerated. Its controller
//	creates a 3/3 green Ape creature token."
//
// Blue's one-mana unconditional creature answer, and mechanically
// Beast Within's little brother: destroy, then hand the VICTIM a
// vanilla 3/3. The controller is read BEFORE the destroy for the same
// reason Beast Within reads it first — afterwards the card is in a
// graveyard and its controller field is stale for a permanent that
// had changed hands.
//
// "It can't be regenerated" is ENFORCED as of #667 (CR 701.19c).
// The
// clause is a no-op today rather than a simplification — if a
// regeneration shield ever lands, this card has to be revisited, so
// the note stays.
func init() {
	Register(Spec{
		OracleID:     "05849bd6-8f38-4031-be2b-e2aa03beb8cc",
		Name:         "Pongify",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
				return nil
			}
			target := item.Targets[0].ID
			controller, ok := controllerOfTarget(ctx, target)
			if !ok {
				return nil
			}
			if err := (DestroyTarget{Target: target, CantBeRegenerated: true}).Apply(ctx); err != nil {
				return err
			}
			return CreateToken{
				Controller: controller,
				Template:   TokenCard("3/3 colorless Ape"),
				N:          1,
			}.Apply(ctx)
		},
	})
}
