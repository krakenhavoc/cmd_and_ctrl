package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Rhino, Terrible Trampler — {4}{G}{G} Legendary Creature — Human Villain,
// 7/6 (Marvel Super Heroes, #4455):
//
//	"Trample
//	 When Rhino enters, destroy target artifact or land. Distribute three
//	 +1/+1 counters among up to three other target creatures. They gain
//	 trample until end of turn."
//
// One trigger with two target clauses (CR 115.1, Decimate's Then): slot 0 is
// the artifact or land, slot 1 is the creatures, with the three counters
// divided among them (CR 601.2d; each chosen creature gets at least one). The
// artifact or land is required, so with none on the battlefield the trigger is
// removed (CR 603.3d). The counters and trample wait for the destroy to finish
// (an artifact creature can be both targets), and each part is rechecked on
// resolution (CR 608.2b): a creature that has gone takes no counters and its
// share is not moved.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "4e1c9e10-1cae-4ee5-866e-b7b5bb47bbce",
		Name:            "Rhino, Terrible Trampler",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{{
			Watches:   []game.EventKind{game.EventETB},
			AppliesTo: b06SelfETB,
			Targets: TargetPermanent("target artifact or land", Or(Artifact(), Land())).Then(
				Another(TargetCreature("up to three other target creatures", Creature())).
					WithCount(0, 3).Dividing(Divide(3))),
			Key: "Rhino — destroy target artifact or land; distribute three +1/+1 counters among up to three other target creatures, which gain trample",
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				if land, ok := ctx.ClauseTarget(0); ok {
					// Each chosen creature rides the tail once per counter it
					// was assigned: the announced split is not on the asked item.
					var carry []uuid.UUID
					for _, t := range ctx.ClauseTargets(1) {
						for n := item.Distribution[t.ID]; n > 0; n-- {
							carry = append(carry, t.ID)
						}
					}
					return g.DestroyPermanentsKeyedForEffect([]uuid.UUID{land.ID}, withCarry(ctx.ItemTail(rhinoCountersAndTrample), carry...))
				}
				return rhinoCountersAndTrampleStep(ctx, item.Distribution)
			},
		}},
	})
}

// rhinoCountersAndTrample is the rest of the sentence once the destroy is done.
var rhinoCountersAndTrample = CardTail("rhino-terrible-trampler/counters-and-trample", func(ctx *Context, o game.TailOutcome) error {
	shares := map[uuid.UUID]int{}
	for _, id := range o.Tail.Carry {
		shares[id]++
	}
	return rhinoCountersAndTrampleStep(ctx, shares)
})

// rhinoCountersAndTrampleStep puts each chosen creature's share of the three
// counters on it and gives it trample, for those still legal targets.
func rhinoCountersAndTrampleStep(ctx *Context, shares map[uuid.UUID]int) error {
	for _, t := range ctx.ClauseTargets(1) {
		if !ctx.IsTargetLegal(t) {
			continue
		}
		if n := shares[t.ID]; n > 0 {
			if err := (AddCounter{Target: t.ID, Kind: game.CounterPlusOne, N: n}).Apply(ctx); err != nil {
				return err
			}
		}
		if err := (GrantKeywordUntilEOT{Target: t.ID, Keywords: []string{"trample"},
			Label: "Rhino — trample until end of turn"}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
}
