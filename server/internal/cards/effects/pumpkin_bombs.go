package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Pumpkin Bombs — Artifact {1}{R} (Marvel's Spider-Man, #4460):
//
//	"{T}, Discard two cards: Draw three cards, then put a fuse counter
//	 on this artifact. It deals damage equal to the number of fuse
//	 counters on it to target opponent. They gain control of this
//	 artifact."
//
// The tap and the two discards are costs, and the target opponent is
// chosen as the ability goes on the stack. On resolution the three cards
// are drawn, then the fuse counter is put on the artifact through the
// settling form of the counter placement (a CR 616 ordering prompt can
// hold it), and only then does the rest run: the artifact deals damage
// equal to the fuse counters it has (by last-known information if it has
// left, CR 608.2h), and the opponent gains control of it with no stated
// duration (CR 611.2a). That opponent is the ability's target; if they
// are no longer a legal target, the damage and the control change do not
// happen, though the draw and the counter do (CR 608.2b). An artifact that
// has become a new object (CR 400.7) gets no counter and changes hands
// not at all.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "b0695221-a61a-41e4-8195-3eeaaab2d969",
		Name:         "Pumpkin Bombs",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{T}, Discard two cards: Draw three cards, then put a fuse counter on this artifact. It deals damage equal to the number of fuse counters on it to target opponent. They gain control of this artifact.",
			Cost:    Plus(TapCost(), DiscardN(2, "two cards")),
			Targets: TargetPlayer("target opponent", Opponent()),
			Effect:  pumpkinBombsResolve,
		}},
	})
}

const pumpkinBombsFuse = "fuse"

func pumpkinBombsResolve(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	if err := (DrawCards{N: 3}).Apply(ctx); err != nil {
		return err
	}
	tail := ctx.ItemTail(pumpkinBombsAfterFuse)
	if sourceIsNewObject(g, item) || !b15OnBattlefield(g, item.SourceCardID) {
		return tail.Run(g, nil)
	}
	return g.AddCounterByKeyedForEffect(ctx.Controller(), item.SourceCardID, pumpkinBombsFuse, 1, tail)
}

// pumpkinBombsAfterFuse is "it deals damage ... to target opponent. They
// gain control of this artifact", run once the fuse counter has settled.
var pumpkinBombsAfterFuse = CardTail("tail/pumpkin-bombs-after-fuse", func(ctx *Context, _ game.TailOutcome) error {
	g, id := ctx.Game, ctx.Item.SourceCardID
	var opponent game.TargetRef
	found := false
	for _, t := range ctx.LegalTargets() {
		if t.Kind == game.TargetPlayer {
			opponent, found = t, true
			break
		}
	}
	if !found {
		return nil
	}
	fuses := g.LastKnownCountersForEffect(id)[pumpkinBombsFuse]
	if b15OnBattlefield(g, id) && !sourceIsNewObject(g, ctx.Item) {
		if c, ok := g.LookupCardForEffect(id); ok {
			fuses = c.Counters[pumpkinBombsFuse]
		}
	}
	if err := (DealDamage{Source: id, Target: opponent.ID, Amount: fuses}).Apply(ctx); err != nil {
		return err
	}
	return GainControl{
		Target:     id,
		Controller: opponent.ID,
		Duration:   game.IndefiniteDuration(),
		Label:      "Pumpkin Bombs — gains control of this artifact",
	}.Apply(ctx)
})
