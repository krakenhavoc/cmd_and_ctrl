package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Laboratory Maniac — Creature — Human Wizard {2}{U}, 2/2:
//
//	"If you would draw a card while your library has no cards in it,
//	 you win the game instead."
//
// ADR 0057's draw-replacement win (#749, CR 104.2b, CR 614). A CR 614
// replacement on game.RepEventDraw, the same event the draw doublers
// and Notion Thief rewrite: when its controller would draw with an
// empty library, the draw is CANCELLED and they win the game instead.
//
// The order inside Replace is the ruling (2021-03-19): cancel the draw
// first, then win. "If for some reason you can't win the game (because
// your opponent has Platinum Angel on the battlefield, for example),
// you won't lose for having tried to draw a card from a library with
// no cards in it. The draw was still replaced." So a prevented win
// leaves no CR 704.5b flag behind — the next check has nothing to read
// — and the game goes on.
//
// The library is read at the moment of each draw, so "draw three" off
// a two-card library draws two and wins on the third (CR 121.2 makes
// it three separate draws). A draw an opponent's Notion Thief
// redirects to its controller is no longer this player's draw.
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:     "aa286dd5-aa19-446d-9003-684d81eb57ca",
		Name:         "Laboratory Maniac",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{WinInsteadOfDrawingFromAnEmptyLibrary("Laboratory Maniac")},
	})
}

// WinInsteadOfDrawingFromAnEmptyLibrary is "If you would draw a card
// while your library has no cards in it, you win the game instead" —
// Laboratory Maniac, and Jace, Wielder of Mysteries' static.
//
// The draw is cancelled BEFORE the win is attempted, so a win a "can't
// win" gate prevents still replaced the draw (the 2019 and 2021
// rulings) and the player does not lose for it. Replace returns nil
// either way: a replacement is not a resolution, and the draw path
// carries on — after a win the game is over and every later draw in
// the same instruction is replaced into a no-op the same way.
func WinInsteadOfDrawingFromAnEmptyLibrary(name string) game.ReplacementEffect {
	return game.ReplacementEffect{
		Watches: []game.EventKind{game.EventDrawCard},
		AppliesTo: func(ev *game.ReplacementEvent, g *game.Game, src *game.Card) bool {
			if ev.Kind != game.RepEventDraw || ev.DrawCount <= 0 || src == nil {
				return false
			}
			if ev.DrawPlayer != src.Controller {
				return false
			}
			p := g.PlayerByIDForEffect(ev.DrawPlayer)
			return p != nil && !p.Eliminated && p.Library.Size() == 0
		},
		Replace: func(ev *game.ReplacementEvent, g *game.Game, src *game.Card) error {
			ev.Cancel()
			_, err := g.WinTheGameForEffect(ev.DrawPlayer, src.InstanceID)
			return err
		},
		Controller: func(_ *game.ReplacementEvent, _ *game.Game, src *game.Card) uuid.UUID {
			return src.Controller
		},
		Label: name + ": win the game instead of drawing from an empty library",
	}
}
