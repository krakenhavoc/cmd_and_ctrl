package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Balthor the Defiled — Legendary Creature — Zombie Dwarf {2}{B}{B}, 2/2
// (Secret Lair Drop):
//
//	"Minion creatures get +1/+1.
//	 {B}{B}{B}, Exile Balthor: Each player returns all black and all red
//	 creature cards from their graveyard to the battlefield."
//
// The anthem is Bad Moon's, over every Minion creature at the table (Balthor
// is a Zombie Dwarf, so not himself). The return is the one simultaneous entry
// of ReturnFromGraveyardTogether over every graveyard, each card under its
// owner's control: the ability's cost exiles Balthor, so he is not among them
// (and a Balthor already in a graveyard is not either, being exiled).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "dfd8e6c2-e425-41fa-af55-985271f771b3",
		Name:         "Balthor the Defiled",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.B16Anthem(func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.IsCreature() && target.HasSubtype(source.TextSubtype("Minion"))
			}, 1, 1),
		},
		Activated: []effects.ActivatedAbility{{
			Label: "{B}{B}{B}, Exile Balthor: Each player returns all black and all red creature cards from their graveyard to the battlefield.",
			// The creatures it returns can block, at instant speed.
			Purpose: game.Purpose{Answers: game.AnswerMakesBlocker},
			Cost:    effects.Plus(effects.ManaCost("{B}{B}{B}"), effects.ExileThis()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				black, red := ctx.TextColor("B"), ctx.TextColor("R")
				ids := effects.AllGraveyardsCardIDs(ctx, func(c game.Card) bool {
					return c.IsCreature() && (c.HasColor(black) || c.HasColor(red))
				})
				if len(ids) == 0 {
					return nil
				}
				return effects.ReturnFromGraveyardTogether{Targets: ids}.Apply(ctx)
			},
		}},
	})
}
