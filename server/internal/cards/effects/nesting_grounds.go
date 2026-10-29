package effects

import (
	"sort"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Nesting Grounds — Land:
//
//	"{T}: Add {C}.
//	 {1}, {T}: Move a counter from target permanent you control onto a
//	 second target permanent. Activate only as a sorcery."
//
// Two target clauses, the second Distinct (CR 122.5 says moving a
// counter from an object onto itself is not possible, and "a second
// target" means a different one). At resolution both must still be
// legal targets (CR 608.2b). When the first permanent carries one kind
// of counter that kind moves; when it carries several, the controller
// is asked which (PickOption), since the printed text lets them choose.
//
// The move is "put it on the second permanent, then remove it from the
// first only if it landed" (CR 122.5), the same order Simic Fluxmage
// uses, so a counter the second permanent can't take is not lost.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "d27bb97d-286b-4947-8d7b-443e4df93319",
		Name:         "Nesting Grounds",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{C}",
			Label:    "Add {C}",
		}},
		Activated: []ActivatedAbility{{
			Label: "{1}, {T}: Move a counter from target permanent you control onto a second target permanent. Activate only as a sorcery.",
			Cost:  Plus(ManaCost("{1}"), TapCost()),
			Targets: Clauses(
				TargetPermanent("target permanent you control", YouControl()),
				Distinct(TargetPermanent("a second target permanent")),
			),
			SorcerySpeed: true,
			Effect:       nestingGroundsMove,
		}},
	})
}

// nestingGroundsRemove takes the moved counter off the first permanent, if it
// landed on the second (the outcome's Amount); Carry is the first permanent
// and Word the counter kind. An on-disk key: never renamed, never reused.
var nestingGroundsRemove = game.RegisterTail("nesting-grounds/remove", func(g *game.Game, o game.TailOutcome) error {
	if o.Amount <= 0 || len(o.Tail.Carry) == 0 {
		return nil
	}
	return g.AddCounterForEffect(o.Tail.Carry[0], o.Tail.Word, -1)
})

func nestingGroundsMove(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	from, ok := ctx.ClauseTarget(0)
	if !ok {
		return nil
	}
	to, ok := ctx.ClauseTarget(1)
	if !ok || from.ID == to.ID {
		return nil
	}
	src, ok := g.LookupCardForEffect(from.ID)
	if !ok {
		return nil
	}
	kinds := nestingGroundsKinds(src)
	switch len(kinds) {
	case 0:
		return nil
	case 1:
		return nestingGroundsMoveKind(ctx, from.ID, to.ID, kinds[0])
	}
	options := make([]game.ChoiceOption, len(kinds))
	for i, kind := range kinds {
		options[i] = game.ChoiceOption{Label: "Move a " + kind + " counter"}
	}
	return PickOption{
		Question: "Nesting Grounds — which kind of counter do you move?",
		Options:  options,
		ThenKey:  nestingGroundsKind,
		Carry:    []uuid.UUID{from.ID, to.ID},
	}.Apply(ctx)
}

// nestingGroundsKinds lists the counter kinds on the permanent, in the order
// the options are offered. Nothing can change a counter while the question
// is open, so the answer rebuilds the same list.
func nestingGroundsKinds(src game.Card) []string {
	var kinds []string
	for kind, n := range src.Counters {
		if n > 0 {
			kinds = append(kinds, kind)
		}
	}
	sort.Strings(kinds)
	return kinds
}

// nestingGroundsMoveKind puts a counter of `kind` on the second permanent and
// takes it off the first once it has landed.
func nestingGroundsMoveKind(ctx *Context, fromID, toID uuid.UUID, kind string) error {
	g := ctx.Game
	if _, ok := g.LookupCardForEffect(toID); !ok {
		return nil
	}
	then := ctx.Tail(nestingGroundsRemove)
	then.Carry = []uuid.UUID{fromID}
	then.Word = kind
	return g.AddCounterByKeyedForEffect(ctx.Controller(), toID, kind, 1, then)
}

// nestingGroundsKind is the counter kind chosen, keyed (#4253): the two
// permanents ride the prompt as data.
var nestingGroundsKind = OptionPickThen("option-pick/nesting-grounds-counter-kind", func(ctx *Context, r game.OptionPicked) error {
	if len(r.Carry) < 2 {
		return nil
	}
	src, ok := ctx.Game.LookupCardForEffect(r.Carry[0])
	if !ok {
		return nil
	}
	kinds := nestingGroundsKinds(src)
	if r.Index < 0 || r.Index >= len(kinds) {
		return nil
	}
	return nestingGroundsMoveKind(ctx, r.Carry[0], r.Carry[1], kinds[r.Index])
})
