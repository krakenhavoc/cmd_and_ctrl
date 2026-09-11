package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Damn — Sorcery for {B}{B}:
//
//	"Destroy target creature. A creature destroyed this way can't be
//	 regenerated."
//	"Overload {2}{W}{W} (You may cast this spell for its overload
//	 cost. If you do, change 'target' in its text to 'each'.)"
//
// Two cards in one: a two-mana Doom Blade that never misses, and a
// four-mana Wrath. Overload is a real alternative cost now (CR
// 118.9), so both halves are live and the choice is made at announce
// the way it is in paper.
//
// The batch-02 triage (#295) filed this under "would ship STRONGER
// than printed — the drawback half has no seam", and that reading
// does not survive contact with the engine. The clause it meant is
// "can't be regenerated", and regeneration does not exist here: no
// card grants it, no shield counter stands in for it, and nothing in
// the destroy path checks for it. A clause that forbids something
// impossible is VACUOUS, not a missing drawback — Damn resolves
// identically with and without it. Indestructible is a separate
// thing and Damn never beat it: "can't be regenerated" does not
// touch indestructible in paper either.
//
// So the card ships whole, with the text that has an effect
// implemented and the text that cannot have one noted as inert. If
// regeneration ever lands, this comment is the thing that has to
// change and the destroy path is where the clause would attach.
//
// The overload sweep snapshots the battlefield before destroying
// anything: DestroyPermanentForEffect mutates the slice and a dies
// trigger can add to it mid-sweep. Same shape as Cyclonic Rift's.
// "Each creature", not "each creature you don't control" — Damn
// kills your board too, and a player casting it for four knew that.
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
