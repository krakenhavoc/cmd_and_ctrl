package cardsst

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Sway of the Stars — Sorcery {8}{U}{U} (Secret Lair Drop):
//
//	"Each player shuffles their hand, graveyard, and all permanents they own
//	 into their library, then draws seven cards. Each player's life total
//	 becomes 7."
//
// Every card in each hand and graveyard, and every permanent on the
// battlefield, goes to its OWNER's library as one simultaneous exit (a
// permanent leaves under whoever controls it; a commander offers CR 903.9's
// replacement and the sweep resumes once it is answered; tokens cease to
// exist). When all of them have gone, each player in turn order from the
// caster shuffles and draws seven, and then each life total becomes 7: a loss
// or a gain of the difference (CR 119.5), so a life-total replacement or
// "can't gain life" sees it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c4fccdd5-5b51-46b9-9937-9167c108743d",
		Name:         "Sway of the Stars",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Draws: 7},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			g := ctx.Game
			var ids []uuid.UUID
			for _, p := range g.Seats {
				if p == nil || p.Eliminated {
					continue
				}
				for _, c := range p.Hand.Cards {
					ids = append(ids, c.InstanceID)
				}
				for _, c := range p.Graveyard.Cards {
					ids = append(ids, c.InstanceID)
				}
			}
			for _, c := range g.BattlefieldCardsForEffect() {
				if owner := g.PlayerByIDForEffect(c.Owner); owner != nil && !owner.Eliminated {
					ids = append(ids, c.InstanceID)
				}
			}
			return g.TuckCardsToLibraryKeyedForEffect(ids, game.TuckOptions{}, ctx.Tail(swayOfTheStarsAfter))
		},
	})
}

// swayOfTheStarsAfter is the rest of the spell once everything has been put
// into the libraries: each player shuffles and draws seven, then each life
// total becomes 7. An on-disk key: never renamed, never reused.
var swayOfTheStarsAfter = game.RegisterTail("then/sway-of-the-stars-shuffle-draw-life", func(g *game.Game, o game.TailOutcome) error {
	ctx := effects.TailContext(g, o.Tail)
	for _, id := range effects.SeatsFromActive(g) {
		if err := g.ShuffleLibraryForEffect(id); err != nil {
			return err
		}
		if err := g.DrawNForEffect(id, 7); err != nil {
			return err
		}
	}
	for _, id := range effects.SeatsFromActive(g) {
		p := g.PlayerByIDForEffect(id)
		if p == nil || p.Eliminated || p.Life == 7 {
			continue
		}
		if err := g.ChangePlayerLifeForEffect(ctx.Source(), id, 7-p.Life); err != nil {
			return err
		}
	}
	return nil
})
