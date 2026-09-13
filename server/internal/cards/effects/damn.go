package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Damn — Sorcery {B}{B} (EDHREC rank 346):
//
//	"Destroy target creature. A creature destroyed this way can't be
//	 regenerated.
//	 Overload {2}{W}{W} (You may cast this spell for its overload
//	 cost. If you do, change "target" in its text to "each.")"
//
// A Doom Blade that is also a Wrath of God. Overload carries both
// halves: the {2}{W}{W} paid instead of the {B}{B}, and the deletion
// of the target clause that turns "target creature" into "each
// creature", announced with no targets and so unfizzleable.
//
// S23: the overloaded half is DestroyAllMatching, so the creatures
// die as one event and the aristocrats payoffs see all of them.
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
				return DestroyAllMatching{Match: Creature()}.Apply(ctx)
			}
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
				return nil
			}
			return DestroyTarget{Target: item.Targets[0].ID}.Apply(ctx)
		},
	})
}
