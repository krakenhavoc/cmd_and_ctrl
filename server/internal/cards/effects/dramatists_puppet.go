package effects

import (
	"sort"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Dramatist's Puppet — Artifact Creature — Construct {4}, 2/4 (Kamigawa:
// Neon Dynasty):
//
//	"When this creature enters, for each kind of counter on target permanent,
//	 put another counter of that kind on it or remove one from it."
//
// The kinds are fixed as the ability resolves: every kind the target
// permanent has a counter of then, in a stable order. Its controller is asked
// once per kind (Plague Boiler's put-or-remove prompt, chained: the
// remaining kinds and the target ride the prompt as data, PickOption.Words
// and Carry), and that kind is then raised or lowered by one. A target that
// has left, or become illegal, stops the chain; a kind whose counter is gone
// cannot be removed and is skipped.
//
// No simplification.
func init() {
	dramatistsPuppetNext = dramatistsPuppetAsk
	Register(Spec{
		OracleID:     "e3473ed5-f2b2-4a41-aed9-ee37caa9d7be",
		Name:         "Dramatist's Puppet",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(WhenThisEnters("Dramatist's Puppet — for each kind of counter on target permanent, put another counter of that kind on it or remove one from it",
				dramatistsPuppetEnters),
				TargetPermanent("target permanent")),
		},
	})
}

func dramatistsPuppetEnters(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	id, ok := b16FirstLegalTargetCard(ctx)
	if !ok {
		return nil
	}
	c, ok := g.LookupCardForEffect(id)
	if !ok {
		return nil
	}
	var kinds []string
	for kind, n := range c.Counters {
		if n > 0 {
			kinds = append(kinds, kind)
		}
	}
	sort.Strings(kinds)
	return dramatistsPuppetAsk(ctx, id, kinds)
}

// dramatistsPuppetAsk asks about the first of kinds, if any remain and the
// target is still on the battlefield.
func dramatistsPuppetAsk(ctx *Context, target uuid.UUID, kinds []string) error {
	if len(kinds) == 0 || !onBattlefield(ctx.Game, target) {
		return nil
	}
	kind := kinds[0]
	return PickOption{
		Question: "Dramatist's Puppet — " + kind + " counters: put another one on it or remove one?",
		Options: []game.ChoiceOption{
			{Label: "Put another " + kind + " counter on it"},
			{Label: "Remove a " + kind + " counter from it"},
		},
		ThenKey: dramatistsPuppetAnswer,
		Carry:   []uuid.UUID{target},
		Words:   kinds,
	}.Apply(ctx)
}

var dramatistsPuppetAnswer = OptionPickThen("option-pick/dramatists-puppet-put-or-remove", func(ctx *Context, r game.OptionPicked) error {
	if len(r.Carry) == 0 || len(r.Words) == 0 {
		return nil
	}
	target, kind := r.Carry[0], r.Words[0]
	if r.Index == game.NoChoiceIndex {
		return nil
	}
	if c, ok := ctx.Game.LookupCardForEffect(target); ok && onBattlefield(ctx.Game, target) {
		switch r.Index {
		case 0:
			if err := (AddCounter{Target: target, Kind: kind, N: 1}).Apply(ctx); err != nil {
				return err
			}
		case 1:
			if c.Counters[kind] > 0 {
				if err := (AddCounter{Target: target, Kind: kind, N: -1}).Apply(ctx); err != nil {
					return err
				}
			}
		}
	}
	return dramatistsPuppetNext(ctx, target, r.Words[1:])
})

// dramatistsPuppetNext is dramatistsPuppetAsk, reached through a variable so
// the answer's body does not close an initialisation cycle with the prompt
// it answers.
var dramatistsPuppetNext func(ctx *Context, target uuid.UUID, kinds []string) error
