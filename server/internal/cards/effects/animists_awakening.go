package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Animist's Awakening — Sorcery {X}{G}:
//
//	"Reveal the top X cards of your library. Put all land cards from
//	 among them onto the battlefield tapped and the rest on the bottom
//	 of your library in a random order.
//	 Spell mastery — If there are two or more instant and/or sorcery
//	 cards in your graveyard, untap those lands."
//
// "All land cards" is not a choice, so there is no prompt: every land
// revealed enters in one simultaneous batch (#745), tapped as printed,
// and the rest go to the bottom through the public random-order bottom
// on the game's seeded RNG.
//
// Spell mastery is read at resolution, while Animist's Awakening is
// still on the stack — it does not count itself (CR 608.2h reads the
// graveyard as it is when the effect needs the information). "Those
// lands" are the ones that actually entered, so a land a replacement
// kept off the battlefield is not untapped by accident.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "6f1bfe50-b61a-4fdd-963c-40d59117bdf4",
		Name:         "Animist's Awakening",
		Completeness: CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			revealed := ctx.Game.RevealTopOfLibraryForEffect(item.Controller, ctx.Source(), ctx.X(),
				"Animist's Awakening — revealed from the top of the library")
			return PutFromLibraryOntoBattlefield{
				Player: item.Controller,
				Cards:  revealed,
				Match:  Land(),
				All:    true,
				Tapped: true,
				Then: func(g *game.Game, res PutFromLibraryResult) error {
					if err := PutRestOnBottomInRandomOrder(g, res); err != nil {
						return err
					}
					if !spellMastery(g, res.Player) {
						return nil
					}
					for _, id := range res.Entered {
						if err := g.UntapTargetForEffect(id); err != nil {
							return err
						}
					}
					return nil
				},
			}.Apply(ctx)
		},
	})
}

// spellMastery is "if there are two or more instant and/or sorcery
// cards in your graveyard". Spell mastery is an ability word (CR 207.2c)
// with no rules meaning of its own; the condition is the whole of it.
func spellMastery(g *game.Game, player uuid.UUID) bool {
	p := g.PlayerByIDForEffect(player)
	if p == nil || p.Graveyard == nil {
		return false
	}
	n := 0
	for _, c := range p.Graveyard.Cards {
		if c.IsInstant() || c.IsSorcery() {
			n++
		}
	}
	return n >= 2
}
