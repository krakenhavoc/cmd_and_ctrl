package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Dark Petition — Sorcery {3}{B}{B} (Secret Lair Drop):
//
//	"Search your library for a card, put that card into your hand, then
//	 shuffle.
//	 Spell mastery — If there are two or more instant and/or sorcery cards in
//	 your graveyard, add {B}{B}{B}."
//
// Demonic Tutor's search, then the spell mastery check as the spell
// resolves, while it is still on the stack, so it does not count itself
// (Fiery Impulse's reader): the mana is added after the search has been
// answered and stays until the step or phase ends.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "79acdb37-8d52-4bc8-a7be-a7e92b6166e5",
		Name:         "Dark Petition",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tutors: 1},
		OnResolve: effects.ResolveSteps(
			effects.Instruction(func(ctx *effects.Context) error {
				return effects.SearchLibrary{
					Player:  ctx.Controller(),
					Dest:    game.ZoneHand,
					Limit:   1,
					Shuffle: true,
					Reason:  "Dark Petition — search your library for a card",
				}.Apply(ctx)
			}),
			effects.If{
				Cond: effects.CondBoard{Name: "spell mastery: two or more instant and/or sorcery cards in your graveyard", Test: func(ctx *effects.Context) bool {
					return effects.SpellMastery(ctx.Game, ctx.Controller())
				}},
				Then: []effects.Applier{effects.AddMana{Produced: "{B}{B}{B}"}},
			},
		),
	})
}
