package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Staff of Domination — Artifact {3} (EDHREC rank 1033):
//
//	"{1}: Untap this artifact.
//	 {2}, {T}: You gain 1 life.
//	 {3}, {T}: Untap target creature.
//	 {4}, {T}: Tap target creature.
//	 {5}, {T}: Draw a card."
//
// The Swiss-army combo piece: with a creature that taps for five or
// more it is infinite life, infinite cards, or an infinite untap.
// Five ordinary CR 602 activated abilities, each on the stack with a
// response window, in printed order. The two targeted ones use the
// ordinary "target creature" clause, so a hexproof creature can be
// neither tapped nor untapped, as printed.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "d7888719-647d-4022-a211-822fa09f0791",
		Name:         "Staff of Domination",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{
			{
				Label: "{1}: Untap this artifact.",
				Cost:  ManaCost("{1}"),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return UntapTarget{Target: item.SourceCardID}.Apply(NewContext(g, item))
				},
			},
			{
				Label: "{2}, {T}: You gain 1 life.",
				Cost:  Plus(ManaCost("{2}"), TapCost()),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return GainLife{Player: item.Controller, Amount: 1}.Apply(NewContext(g, item))
				},
			},
			{
				Label:   "{3}, {T}: Untap target creature.",
				Cost:    Plus(ManaCost("{3}"), TapCost()),
				Targets: TargetCreature("target creature"),
				Effect: func(g *game.Game, item *game.StackItem) error {
					if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
						return nil
					}
					return UntapTarget{Target: item.Targets[0].ID}.Apply(NewContext(g, item))
				},
			},
			{
				Label:   "{4}, {T}: Tap target creature.",
				Cost:    Plus(ManaCost("{4}"), TapCost()),
				Targets: TargetCreature("target creature"),
				Effect:  tapChosenPermanent,
			},
			{
				Label: "{5}, {T}: Draw a card.",
				Cost:  Plus(ManaCost("{5}"), TapCost()),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return DrawCards{Player: item.Controller, N: 1}.Apply(NewContext(g, item))
				},
			},
		},
	})
}
