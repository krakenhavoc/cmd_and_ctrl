package effects

import (
	"sort"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b242_helpers.go — shared vocabulary of Amonkhet batch 242 (S82 older
// sets, #5535): "-1/-1 counters on a creature you control" enters triggers,
// and "for each kind of counter on target permanent, put another counter of
// that kind on it or remove one from it" (Dramatist's Puppet, Quarry Hauler).

// OldB242WhenThisEntersMinusCountersOnYourCreature is "When this creature
// enters, put N -1/-1 counters on target creature you control": the target is
// chosen as the ability goes on the stack, and may be the creature itself.
func OldB242WhenThisEntersMinusCountersOnYourCreature(name string, n int) game.TriggeredAbility {
	t := WhenThisEnters(name+" — put "+counterWord(n)+" -1/-1 on target creature you control",
		func(g *game.Game, item *game.StackItem) error {
			ctx := NewContext(g, item)
			target, ok := ctx.ClauseTarget(0)
			if !ok || target.Kind != game.TargetCard {
				return nil
			}
			return AddCounter{Target: target.ID, Kind: game.CounterMinusOne, N: n}.Apply(ctx)
		})
	t.Targets = TargetCreatureYouControl()
	return t
}

func counterWord(n int) string {
	switch n {
	case 1:
		return "a"
	case 2:
		return "two"
	case 3:
		return "three"
	}
	return "some"
}

// OldB242ForEachKindPutOrRemove is the enters effect "for each kind of counter
// on target permanent, put another counter of that kind on it or remove one
// from it". The kinds are fixed as the ability resolves: every kind the target
// has a counter of then, in a stable order. Its controller is asked once per
// kind (the remaining kinds and the target ride the prompt as data,
// PickOption.Words and Carry), and that kind is then raised or lowered by one.
// A target that has left stops the chain; a kind whose counter is gone cannot
// be removed and is skipped. The target is clause 0.
func OldB242ForEachKindPutOrRemove() Effect {
	return func(g *game.Game, item *game.StackItem) error {
		ctx := NewContext(g, item)
		id, ok := B16FirstLegalTargetCard(ctx)
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
		return oldB242AskKind(ctx, id, kinds)
	}
}

// oldB242Next is oldB242AskKind, reached through a variable so the answer's
// body does not close an initialisation cycle with the prompt it answers.
var oldB242Next func(ctx *Context, target uuid.UUID, kinds []string) error

func init() { oldB242Next = oldB242AskKind }

func oldB242AskKind(ctx *Context, target uuid.UUID, kinds []string) error {
	if len(kinds) == 0 || !OnBattlefield(ctx.Game, target) {
		return nil
	}
	kind := kinds[0]
	name := "Counters"
	if src, ok := ctx.Game.LookupCardForEffect(ctx.Source()); ok {
		name = src.Name
	}
	return PickOption{
		Question: name + " — " + kind + " counters: put another one on it or remove one?",
		Options: []game.ChoiceOption{
			{Label: "Put another " + kind + " counter on it"},
			{Label: "Remove a " + kind + " counter from it"},
		},
		ThenKey: oldB242KindAnswer,
		Carry:   []uuid.UUID{target},
		Words:   kinds,
	}.Apply(ctx)
}

// oldB242KindAnswer keeps the on-disk key Dramatist's Puppet persisted
// (effect keys are never renamed); Words carries the kinds still to ask about.
var oldB242KindAnswer = OptionPickThen("option-pick/dramatists-puppet-put-or-remove", func(ctx *Context, r game.OptionPicked) error {
	if len(r.Carry) == 0 || len(r.Words) == 0 {
		return nil
	}
	target, kind := r.Carry[0], r.Words[0]
	if r.Index == game.NoChoiceIndex {
		return nil
	}
	if c, ok := ctx.Game.LookupCardForEffect(target); ok && OnBattlefield(ctx.Game, target) {
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
	return oldB242Next(ctx, target, r.Words[1:])
})

// OldB242DecimatorTargets is the target clause set of Decimator Beetle's attack
// trigger: "target creature you control" and "up to one target creature
// defending player controls". The defending player is a fact about this
// attack (CR 508.5), read off the source's own attack target.
func OldB242DecimatorTargets(tc game.TriggerContext, source *game.Card, g *game.Game) *game.TargetSpec {
	mine := TargetCreatureYouControl()
	theirs := TargetCreatureDefendingPlayerControls(tc, source, g)
	if theirs == nil {
		return mine
	}
	theirs.Label = "up to one target creature defending player controls"
	theirs.Min = 0
	return Clauses(mine, theirs)
}
