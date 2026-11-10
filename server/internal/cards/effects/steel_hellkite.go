package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Steel Hellkite — Artifact Creature — Dragon {6}, 5/5 (Foundations):
//
//	"Flying
//	 {2}: This creature gets +1/+0 until end of turn.
//	 {X}: Destroy each nonland permanent with mana value X whose
//	 controller was dealt combat damage by this creature this turn.
//	 Activate only once each turn."
//
// The sweep reads the turn tally's record of which OBJECT dealt combat
// damage to whom (ObjectRefDealtCombatDamageToPlayerThisTurn), by the
// Hellkite's stamped ref: a Hellkite that died or was exiled in response
// still counts the damage it dealt (CR 608.2h), and one that left and came
// back is a stranger (CR 400.7). "Whose controller" is read as each
// permanent is considered, so a permanent whose controller changed is
// judged by its controller now. A token's mana value is 0.
func init() {
	const sweepLabel = "{X}: Destroy each nonland permanent with mana value X whose controller was dealt combat damage by this creature this turn. Activate only once each turn."
	Register(Spec{
		OracleID:        "6e64f1e0-ddd6-4ee7-8797-086e234b02e8",
		Name:            "Steel Hellkite",
		Completeness:    CompletenessFull,
		XMatters:        true,
		PrintedKeywords: []string{"flying"},
		Activated: []ActivatedAbility{
			{
				Label:   "{2}: This creature gets +1/+0 until end of turn.",
				Purpose: game.Purpose{Answers: game.AnswerPump},
				Cost:    ManaCost("{2}"),
				Effect: func(g *game.Game, item *game.StackItem) error {
					if !sourceIsStillThisPermanent(g, item) {
						return nil
					}
					return BoostUntilEOT{Target: item.SourceCardID, Power: 1,
						Label: "Steel Hellkite — +1/+0 until end of turn"}.Apply(NewContext(g, item))
				},
			},
			{
				Label:     sweepLabel,
				Purpose:   game.Purpose{Answers: game.AnswerRemove, Sweep: game.Sweep{Matches: game.SweepNonlandPermanents, How: game.SweepDestroy, Partial: true}},
				Cost:      ManaCost("{X}"),
				Condition: OncePerTurnActivation(sweepLabel),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					return steelHellkiteSweep(g, item, ctx.X())
				},
			},
		},
	})
}

// steelHellkiteSweep destroys each nonland permanent with mana value X
// whose controller was dealt combat damage by the Hellkite this turn.
func steelHellkiteSweep(g *game.Game, item *game.StackItem, x int) error {
	ctx := NewContext(g, item)
	ref, ok := ctx.SourceRef()
	if !ok {
		return nil
	}
	return DestroyAllMatching{Match: func(g *game.Game, caster uuid.UUID, c game.Card) bool {
		return Nonland()(g, caster, c) && b03ManaValueIs(x)(g, caster, c) &&
			g.ObjectRefDealtCombatDamageToPlayerThisTurn(ref, c.Controller)
	}}.Apply(ctx)
}
