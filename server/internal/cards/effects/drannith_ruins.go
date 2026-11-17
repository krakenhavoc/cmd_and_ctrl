package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Drannith Ruins — Land (March of the Machine: The Aftermath):
//
//	"{T}: Add {C}.
//	 {2}, {T}: Put two +1/+1 counters on target non-Human creature that
//	 entered this turn."
//
// The first is a mana ability. The second is an ordinary activation whose
// target is judged at announce and again as it resolves (CR 608.2b): the
// per-object entered-this-turn tally (b06EnteredThisTurn) and the creature's
// effective subtypes, so a changeling, which is a Human, is not a legal
// target, and a creature that left in response gets nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:      "d1f10cca-8dfa-4ea5-b227-4446cd8514a8",
		Name:          "Drannith Ruins",
		Completeness:  CompletenessFull,
		ManaAbilities: []ManaAbility{colorlessTapRow()},
		Activated: []ActivatedAbility{{
			Label: "{2}, {T}: Put two +1/+1 counters on target non-Human creature that entered this turn.",
			Cost:  Plus(ManaCost("{2}"), TapCost()),
			Targets: TargetCreature("target non-Human creature that entered this turn",
				Not(HasSubtype("Human")),
				func(g *game.Game, _ uuid.UUID, c game.Card) bool { return b06EnteredThisTurn(g, c.InstanceID) }),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				target, ok := b16FirstLegalTargetCard(ctx)
				if !ok {
					return nil
				}
				return AddCounter{Target: target, Kind: game.CounterPlusOne, N: 2}.Apply(ctx)
			},
		}},
	})
}
