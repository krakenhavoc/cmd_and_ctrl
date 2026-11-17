package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Tawnos's Tinkering — Instant {3}{G} (The Brothers' War, #5482):
//
//	"Put two +1/+1 counters on target artifact, creature, or land you
//	 control. Untap that permanent. If it isn't a creature, it becomes a
//	 0/0 creature in addition to its other types."
//
// In printed order (CR 608.2c): the counters go on through the CR 614
// placement window, then (from the counters' keyed tail, so a CR 616
// ordering prompt holds the rest) the untap, then the type change, which
// prints no duration and so lasts until the end of the game (CR 611.2a),
// pinned to this object (CR 400.7). "Isn't a creature" is asked after
// the counters and the untap. A 0/0 with two counters is a 2/2.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "bd101a9b-4c1e-44c8-b9ef-cd79c4d28f36",
		Name:         "Tawnos's Tinkering",
		Completeness: CompletenessFull,
		Targets:      TargetPermanent("target artifact, creature, or land you control", And(YouControl(), Or(Artifact(), Creature(), Land()))),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			t, ok := ctx.ClauseTarget(0)
			if !ok || t.Kind != game.TargetCard || !onBattlefield(ctx.Game, t.ID) {
				return nil
			}
			return ctx.Game.AddCounterByKeyedForEffect(item.Controller, t.ID, game.CounterPlusOne, 2, ctx.ItemTail(tawnossTinkeringAfter))
		},
	})
}

// tawnossTinkeringAfter is the untap and the animation once the counters
// have landed. An on-disk key: never renamed, never reused.
var tawnossTinkeringAfter = CardTail("s82-bro-b11/tawnoss-tinkering-after", func(ctx *Context, _ game.TailOutcome) error {
	t, ok := ctx.ClauseTarget(0)
	if !ok || t.Kind != game.TargetCard || !onBattlefield(ctx.Game, t.ID) {
		return nil
	}
	if err := (UntapTarget{Target: t.ID}).Apply(ctx); err != nil {
		return err
	}
	c, ok := ctx.Game.LookupCardForEffect(t.ID)
	if !ok || c.IsCreature() {
		return nil
	}
	return ScopedEffectFor{
		Target: t.ID,
		Mods: append([]game.Mod{game.AddTypesMod("Creature")},
			game.SetBasePTMods(0, 0)...),
		Duration: ctx.Game.PinnedTo(game.IndefiniteDuration(), t.ID),
		Label:    "Tawnos's Tinkering — a 0/0 creature in addition to its other types",
	}.Apply(ctx)
})
