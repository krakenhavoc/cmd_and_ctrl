package cardsdi

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Dulcet Sirens — Creature — Siren {2}{U}, 1/3 (Commander 2014, #5553):
//
//	"{U}, {T}: Target creature attacks target opponent this turn if able.
//	 Morph {U} (You may cast this card face down as a 2/2 creature for {3}.
//	 Turn it face up any time for its morph cost.)"
//
// Two target clauses, the creature and then the opponent (CR 115.1), each
// re-checked as the ability resolves; with either one gone it does nothing.
// The requirement is Knight Rampager's AttackPlayerIfAbleMod on the creature
// until end of turn, so it binds every combat left this turn and CR 508.1d
// lets the creature stay home when it can't attack that player. Face down the
// Sirens is a 2/2 with no abilities (CR 708.2) until it is turned up for {U}.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:         "37c3cd4d-8e9a-46d4-a9c2-55f70a321eb6",
		Name:             "Dulcet Sirens",
		Completeness:     effects.CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{effects.Morph("{U}")},
		Activated: []effects.ActivatedAbility{{
			Label:   "{U}, {T}: Target creature attacks target opponent this turn if able.",
			Purpose: game.Purpose{Answers: game.AnswerCombatGrant},
			Cost:    effects.Plus(effects.ManaCost("{U}"), effects.TapCost()),
			Targets: effects.Clauses(
				effects.TargetCreature("target creature"),
				effects.TargetPlayer("target opponent", effects.Opponent()),
			),
			Effect: dulcetSirensForceAttack,
		}},
	})
}

// dulcetSirensForceAttack makes the target creature attack the target
// opponent this turn if able.
func dulcetSirensForceAttack(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	var creature, opp uuid.UUID
	for _, t := range ctx.LegalTargets() {
		switch t.Kind {
		case game.TargetCard:
			creature = t.ID
		case game.TargetPlayer:
			opp = t.ID
		}
	}
	if creature == uuid.Nil || opp == uuid.Nil {
		return nil
	}
	return effects.UntilEndOfTurn(ctx, creature, nil, "Dulcet Sirens — attacks that opponent this turn if able",
		game.AttackPlayerIfAbleMod(opp))
}
