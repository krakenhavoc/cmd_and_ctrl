package cardssy

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Szat's Will — Instant {4}{B} (Commander Legends):
//
//	"Choose one. If you control a commander as you cast this spell, you may
//	 choose both instead.
//	 • Each opponent sacrifices a creature they control with the greatest power.
//	 • Exile all opponents' graveyards, then create X 0/1 black Thrull creature
//	 tokens, where X is the greatest power among creature cards exiled this way."
//
// The mode count is the Wills' (OrUpToIf with YouControlACommander). The edict is
// Shadowgrange Archfiend's: each opponent chooses among the creatures tied for
// the greatest power they control, and everything chosen leaves together. With
// both modes the second runs from the edict's continuation, so the sacrificed
// creature cards are already in the graveyards it exiles and count toward X.
// X reads the printed power of the creature cards that really reached exile.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f0d8bde1-76db-4bd2-ae90-3e3a3418de0b",
		Name:         "Szat's Will",
		Completeness: effects.CompletenessFull,
		Modes: effects.ChooseOne(
			effects.Mode("Each opponent sacrifices a creature they control with the greatest power."),
			effects.Mode("Exile all opponents' graveyards, then create X 0/1 black Thrull creature tokens, where X is the greatest power among creature cards exiled this way."),
		).OrUpToIf(2, effects.YouControlACommander),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			if ctx.HasMode(0) {
				opponents := ctx.Opponents()
				if len(opponents) > 0 {
					return ctx.PlayersSacrificeThen(opponents,
						effects.SacrificeSpec("a creature with the greatest power among creatures you control", effects.GreatestPowerYouControl()),
						"Szat's Will — sacrifice a creature with the greatest power", szatsWillSacrificed, ctx.RunData(nil))
				}
			}
			return szatsWillExileAndMake(ctx)
		},
	})
}

// szatsWillSacrificed is the edict's continuation: the second bullet, if it was
// chosen. An on-disk key.
var szatsWillSacrificed = effects.RunThen("prompt-run/szats-will-sacrificed", func(ctx *effects.Context, _ game.RunDone) error {
	return szatsWillExileAndMake(ctx)
})

func szatsWillExileAndMake(ctx *effects.Context) error {
	if !ctx.HasMode(1) {
		return nil
	}
	var ids []uuid.UUID
	for _, opp := range ctx.Opponents() {
		if p := ctx.PlayerByID(opp); p != nil {
			for _, c := range p.Graveyard.Cards {
				ids = append(ids, c.InstanceID)
			}
		}
	}
	return ctx.Game.ExileCardsKeyedForEffect(ids, ctx.ItemTail(szatsWillExiled))
}

// szatsWillExiled makes the Thrulls: X is the greatest power among the creature
// cards that reached exile. An on-disk key.
var szatsWillExiled = effects.CardTail("szats-will/exiled", func(ctx *effects.Context, o game.TailOutcome) error {
	x := 0
	for _, id := range o.IDs {
		if c, ok := ctx.Game.LookupCardForEffect(id); ok && c.IsCreature() {
			x = max(x, c.PowerForComparison())
		}
	}
	if x <= 0 {
		return nil
	}
	return effects.CreateToken{Template: effects.TokenCard("0/1 black Thrull"), N: x}.Apply(ctx)
})
