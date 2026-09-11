package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Damn — Sorcery {B}{B} (EDHREC rank 346):
//
//	"Destroy target creature. A creature destroyed this way can't be
//	 regenerated.
//	 Overload {2}{W}{W} (You may cast this spell for its overload
//	 cost. If you do, change 'target' in its text to 'each.')"
//
// A Doom Blade that is also a Wrath of God — Vandalblast's shape in
// black-white. Overload carries both halves: the {2}{W}{W} paid
// instead of the {B}{B}, and the deletion of the target clause that
// turns "target creature" into "each creature", announced with no
// targets and so unfizzleable.
//
// "Can't be regenerated" is a no-op because regeneration is not
// modelled — see terminate.go for the note to revisit.
func init() {
	Register(Spec{
		OracleID: "b01d61cc-9844-4191-86a0-f2db6d42d6e5",
		Name:     "Damn",
		Targets:  TargetCreature("target creature"),
		AlternativeCosts: []game.AlternativeCost{
			Overload("{2}{W}{W}"),
		},
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if ctx.PaidAltCost("overload") {
				// Snapshot before destroying: DestroyTarget removes
				// cards from the slice we would otherwise range over.
				var doomed []uuid.UUID
				for _, c := range ctx.Game.BattlefieldCardsForEffect() {
					if c.IsCreature() {
						doomed = append(doomed, c.InstanceID)
					}
				}
				for _, id := range doomed {
					if err := (DestroyTarget{Target: id}).Apply(ctx); err != nil {
						return err
					}
				}
				return nil
			}
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
				return nil
			}
			return DestroyTarget{Target: item.Targets[0].ID}.Apply(ctx)
		},
	})
}
