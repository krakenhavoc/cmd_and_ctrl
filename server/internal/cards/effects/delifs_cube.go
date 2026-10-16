package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Delif's Cube — Artifact {1}:
//
//	"{2}, {T}: This turn, when target creature you control attacks and
//	 isn't blocked, it assigns no combat damage this turn and you put a
//	 cube counter on this artifact.
//	 {2}, Remove a cube counter from this artifact: Regenerate target
//	 creature."
//
// #3531, ADR 0187 amendment 2026-10-16: the first ability creates a
// one-shot delayed trigger about the creature it targeted, as Delif's
// Cone does (whenThatCreatureAttacksUnblocked). Its source is the Cube
// (CR 603.7d), so "this artifact" is the Cube object that made it: a Cube
// that has left the battlefield, or come back as a new object, gets no
// counter, and the creature is still marked.
//
// No simplification.
const delifsCubeCounter = "cube"

func init() {
	Register(Spec{
		OracleID:     "3106ba15-08d7-474c-a384-6848b713ba8d",
		Name:         "Delif's Cube",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{
			{
				Label:   "{2}, {T}: This turn, when target creature you control attacks and isn't blocked, it assigns no combat damage this turn and you put a cube counter on this artifact.",
				Cost:    Plus(ManaCost("{2}"), TapCost()),
				Targets: TargetCreature("target creature you control", YouControl()),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					return whenThatCreatureAttacksUnblocked(ctx, FirstLegalBattlefieldTarget(ctx),
						"Delif's Cube — it assigns no combat damage this turn and you put a cube counter on Delif's Cube", delifsCubeBody)
				},
			},
			{
				Label:   "{2}, Remove a cube counter from this artifact: Regenerate target creature.",
				Cost:    Plus(ManaCost("{2}"), RemoveCountersFromThis(delifsCubeCounter, 1)),
				Targets: TargetCreature("target creature"),
				Effect:  regenerateTheTargetPermanent,
			},
		},
	})
}

// delifsCube is the delayed trigger's body: mark the creature, then put
// the counter on the Cube that made the trigger.
func delifsCube(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	if it := firstPayloadCard(ctx); it != uuid.Nil {
		g.AssignsNoCombatDamageThisTurnForEffect(ctx.Source(), it, "Delif's Cube — assigns no combat damage this turn")
	}
	return AddCounter{Target: ctx.Source(), Kind: delifsCubeCounter, N: 1}.Apply(ctx)
}
