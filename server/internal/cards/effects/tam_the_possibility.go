package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Tam, the Possibility — Legendary Creature — Gorgon Wizard {1}{G}{U}, 2/4
// (Reality Fracture, tracker #2795):
//
//	"Planeswalker spells you cast cost {1} less to cast.
//	 {W}{U}{B}{R}{G}, {T}: Proliferate X times, where X is the number of
//	 planeswalker types among planeswalkers you control."
//
// The discount is a cost modifier over the controller's own planeswalker
// spells. X is counted as the ability resolves (it is not a cost), from the
// post-layer subtypes of the planeswalkers you control, each distinct type
// counted once. Each proliferate is its own prompt (CR 701.34), so the
// repetitions are chained through Proliferate's continuation: the next
// begins only when the last has been answered, and a planeswalker that
// gained a counter on the first pass can be chosen again on the second.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "c92c7744-0c92-4a36-9ada-4ffca812009e",
		Name:         "Tam, the Possibility",
		Completeness: CompletenessFull,
		CostModifiers: []game.CostModifier{
			CostsLess(1, "Planeswalker spells you cast cost {1} less to cast.",
				YourSpell(), tamPlaneswalkerSpell),
		},
		Activated: []ActivatedAbility{{
			Label: "{W}{U}{B}{R}{G}, {T}: Proliferate X times, where X is the number of planeswalker types among planeswalkers you control.",
			Cost:  Plus(ManaCost("{W}{U}{B}{R}{G}"), TapCost()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				return tamProliferate(g, ctx.Controller(), ctx.Source(), planeswalkerTypesAmong(g, ctx.Controller()))
			},
		}},
	})
}

// tamPlaneswalkerSpell is "planeswalker spells" for the cost query.
func tamPlaneswalkerSpell(q game.CostQuery) bool { return q.Card.IsPlaneswalker() }

// planeswalkerTypesAmong is "the number of planeswalker types among
// planeswalkers you control": distinct subtypes across `player`'s
// planeswalker permanents. Caller must hold g.mu.
func planeswalkerTypesAmong(g *game.Game, player uuid.UUID) int {
	g.RecomputeLayersIfStaleLocked()
	seen := map[string]bool{}
	for _, c := range g.Battlefield.Cards {
		if c.Controller != player || !c.IsPlaneswalker() {
			continue
		}
		for _, st := range c.Effective().Subtypes {
			seen[st] = true
		}
	}
	return len(seen)
}

// tamProliferate proliferates `n` times in sequence: each repetition waits
// for the previous prompt to be answered. Captures scalars only.
func tamProliferate(g *game.Game, player, source uuid.UUID, n int) error {
	if n <= 0 {
		return nil
	}
	return g.ProliferateChoosingForEffect(player, source, func(g *game.Game) error {
		return tamProliferate(g, player, source, n-1)
	})
}
