package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Lich-Knights' Conquest — Sorcery {4}{B} (EDHREC rank 2596):
//
//	"Sacrifice any number of artifacts, enchantments, and/or tokens.
//	 Return that many creature cards from your graveyard to the
//	 battlefield."
//
// Treasures and Clues into a mass reanimation. Both choices are made
// as the spell resolves, in the printed order (#2863, ADR 0013
// amendment of 2026-10-09):
//
//  1. "Sacrifice any number" is one own-permanents pick over your
//     artifacts, enchantments and tokens with a floor of zero
//     (ChoosePermanents, Scapeshift's shape), and the chosen ones are
//     sacrificed together (SacrificeAllThenForEffect).
//  2. "That many" is how many really left the battlefield, read off
//     the sacrifice's answer rather than off the clicks (#1019's
//     rule), so a permanent that could not be sacrificed buys nothing.
//  3. "Return that many creature cards" is then one pick from the
//     graveyard as it is NOW (ReturnChosenFromGraveyard): exactly that
//     many, or every creature card there if there are fewer (CR
//     608.2). An artifact creature card sacrificed in step 1 is in the
//     graveyard by then and can come back; a sacrificed creature token
//     has ceased to exist and cannot.
//
// Nothing is targeted, so the spell has no pick for opponents to see
// at cast and cannot fizzle. The returned cards enter together
// (#1867), so each sees the others enter.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "b0e533dd-baf2-482a-b9e1-872eb5e47439",
		Name:         "Lich-Knights' Conquest",
		Completeness: CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return ChoosePermanents{
				Question:   "Lich-Knights' Conquest — sacrifice any number of artifacts, enchantments, and/or tokens",
				Candidates: lichKnightsFodder,
				// Sacrificed here rather than with Sacrifice: true,
				// because "that many" is what LEFT, not what was
				// clicked.
				Then: func(ctx *Context, picked game.PromptedPicks) error {
					if picked.Count() == 0 {
						return nil
					}
					return ctx.Game.SacrificeAllThenForEffect(ctx.Source(), picked.Cards(),
						func(g *game.Game, sacrificed []uuid.UUID) error {
							n := len(sacrificed)
							if n == 0 {
								return nil
							}
							return ReturnChosenFromGraveyard{
								Question: "Lich-Knights' Conquest — return that many creature cards from your graveyard to the battlefield",
								Match:    game.Card.IsCreature,
								Min:      n,
								Max:      n,
							}.Apply(NewContext(g, item))
						})
				},
			}.Apply(ctx)
		},
	})
}

// lichKnightsFodder is every artifact, enchantment and token the
// chooser controls, with a floor of zero and no ceiling but the board —
// "any number".
//
// Caller holds g.mu.
func lichKnightsFodder(g *game.Game, of uuid.UUID) ([]uuid.UUID, int, int) {
	match := Or(Artifact(), Enchantment(), IsTokenPredicate())
	var out []uuid.UUID
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == of && match(g, of, c) {
			out = append(out, c.InstanceID)
		}
	}
	return out, 0, len(out)
}
