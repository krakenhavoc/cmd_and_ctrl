package cardssh

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Shadows' Verdict — Sorcery {3}{B}{B} (Zendikar Rising, #5508):
//
//	"Exile all creatures and planeswalkers with mana value 3 or less from
//	 the battlefield and all creature and planeswalker cards with mana value
//	 3 or less from all graveyards."
//
// Two sweeps in one resolution (CR 608.2c): ExileAllMatching over the
// battlefield (a token has mana value 0 and goes too), then every
// graveyard's creature and planeswalker cards with mana value 3 or less.
// Mana value reads the card as it is (X as 0 off the stack, a double-faced
// card's front face, CR 202.3). Exile is not destruction, so indestructible
// does not save anything.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "83e99691-bc25-4b3e-b828-670ef2027592",
		Name:         "Shadows' Verdict",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreaturesManaValue3OrLess, How: game.SweepExile, Partial: true}},
		OnResolve: effects.ResolveSteps(
			effects.Instruction(func(ctx *effects.Context) error {
				return effects.ExileAllMatching{Match: shadowsVerdictMatch()}.Apply(ctx)
			}),
			effects.Instruction(shadowsVerdictExileGraveyards),
		),
	})
}

// shadowsVerdictMatch is "creature or planeswalker with mana value 3 or less".
func shadowsVerdictMatch() effects.CardPredicate {
	return effects.And(effects.Or(effects.Creature(), effects.Planeswalker()), effects.ManaValueLE(3))
}

// shadowsVerdictExileGraveyards is the second half: every graveyard's creature
// and planeswalker cards with mana value 3 or less, exiled together.
func shadowsVerdictExileGraveyards(ctx *effects.Context) error {
	g, match := ctx.Game, shadowsVerdictMatch()
	var ids []uuid.UUID
	for _, p := range g.Seats {
		if p == nil || p.Graveyard == nil {
			continue
		}
		for _, c := range p.Graveyard.Cards {
			if (c.IsCreature() || c.IsPlaneswalker()) && match(g, ctx.Controller(), c) {
				ids = append(ids, c.InstanceID)
			}
		}
	}
	if len(ids) == 0 {
		return nil
	}
	g.ExileCardsForEffect(ids)
	return nil
}
