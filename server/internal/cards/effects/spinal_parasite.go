package effects

import (
	"sort"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Spinal Parasite — Artifact Creature — Insect {5}, -1/-1:
//
//	"Sunburst (This creature enters with a +1/+1 counter on it for each
//	 color of mana spent to cast it.)
//	 Remove two +1/+1 counters from this creature: Remove a counter
//	 from target permanent."
//
// Sunburst is a keyword the engine reads off the resolving spell (ADR
// 0109 §11, #1552). Cast off one colour it is a 0/0 and dies.
//
// "Remove a counter" leaves the kind to the ability's controller, asked
// as it resolves when the permanent has more than one kind; one kind is
// removed without a question, and a permanent with none loses nothing.
func init() {
	Register(Spec{
		OracleID:            "159427f7-27fe-490d-ac78-fed092952f51",
		Name:                "Spinal Parasite",
		Completeness:        CompletenessCaveats,
		Caveats:             []string{"With strict mana off, the game doesn't track which mana you spent, so it enters with no counters at all."},
		WantsDistinctColors: true,
		PrintedKeywords:     []string{game.KeywordSunburst},
		Activated: []ActivatedAbility{{
			Label:   "Remove two +1/+1 counters from this creature: Remove a counter from target permanent.",
			Cost:    RemoveCountersFromThis(game.CounterPlusOne, 2),
			Targets: TargetPermanent("target permanent"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				for _, t := range ctx.LegalTargets() {
					if t.Kind == game.TargetCard {
						return removeACounterOfYourChoice(ctx, t.ID)
					}
				}
				return nil
			},
		}},
	})
}

// removeACounterOfYourChoice is "remove a counter from <permanent>":
// one counter, of a kind the controller picks when there is more than
// one kind on it. The kinds are offered in name order.
func removeACounterOfYourChoice(ctx *Context, target uuid.UUID) error {
	c, ok := ctx.Game.LookupCardForEffect(target)
	if !ok || !onBattlefield(ctx.Game, target) {
		return nil
	}
	var kinds []string
	for kind, n := range c.Counters {
		if n > 0 {
			kinds = append(kinds, kind)
		}
	}
	sort.Strings(kinds)
	switch len(kinds) {
	case 0:
		return nil
	case 1:
		return AddCounter{Target: target, Kind: kinds[0], N: -1}.Apply(ctx)
	}
	options := make([]game.ChoiceOption, len(kinds))
	for i, kind := range kinds {
		options[i] = game.ChoiceOption{Label: "Remove a " + kind + " counter"}
	}
	return PickOption{
		Question: "Remove which counter from " + c.Name + "?",
		Options:  options,
		Then: func(ctx *Context, index int) error {
			if index < 0 || index >= len(kinds) {
				return nil
			}
			return AddCounter{Target: target, Kind: kinds[index], N: -1}.Apply(ctx)
		},
	}.Apply(ctx)
}
