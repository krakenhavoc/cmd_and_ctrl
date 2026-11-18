package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Eruth, Tormented Prophet — Legendary Creature — Human Wizard {1}{U}{R},
// 2/4 (Innistrad: Crimson Vow):
//
//	"If you would draw a card, exile the top two cards of your library
//	 instead. You may play those cards this turn."
//
// A CR 614 draw replacement whose substitute is an effect: the draw is
// cancelled (nothing is drawn, so no "whenever you draw" ability triggers,
// CR 614.6) and the top two cards of the library are exiled with a
// this-turn permission to play them (a land counts against the land drop).
// "Draw three" is three draws (CR 121.2), each replaced on its own, so it
// exiles six. The rest of the draw instruction waits for the exile to
// settle, as Ormos's does. With fewer than two cards, as many as exist are
// exiled and no card is drawn, so the empty library is not drawn from
// (CR 704.5b needs a draw).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "a0380b63-58ef-4545-beec-6ad307bbc21b",
		Name:         "Eruth, Tormented Prophet",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{{
			Watches: []game.EventKind{game.EventDrawCard},
			AppliesTo: func(ev *game.ReplacementEvent, g *game.Game, src *game.Card) bool {
				if ev.Kind != game.RepEventDraw || ev.DrawCount <= 0 || src == nil || ev.DrawPlayer != src.Controller {
					return false
				}
				p := g.PlayerByIDForEffect(ev.DrawPlayer)
				return p != nil && !p.Eliminated
			},
			DrawInstead: game.RegisterDrawInstead("eruth-tormented-prophet", eruthExileTwoInstead),
			Controller: func(_ *game.ReplacementEvent, _ *game.Game, src *game.Card) uuid.UUID {
				return src.Controller
			},
			Label: "Eruth, Tormented Prophet: exile the top two cards of your library instead of drawing",
		}},
	})
}

// eruthExileTwoInstead is the replaced draw: the drawer exiles the top two
// cards of their library and may play them this turn, then the rest of the
// draw instruction (the parked DrawDone) goes on once they have settled.
func eruthExileTwoInstead(g *game.Game, drawer, _ uuid.UUID, done game.DrawDone) error {
	then := game.KeyedTail{Ref: eruthThenDraw, Player: drawer, Carry: []uuid.UUID{done.Park(g)}}
	return g.ExileTopWithPermissionKeyedForEffect(drawer, drawer, 2, game.CastPermission{}, then)
}

// eruthThenDraw pays the rest of the draw instruction once the two cards
// have settled. An on-disk key.
var eruthThenDraw = game.RegisterTail("eruth-tormented-prophet/then-draw", func(g *game.Game, o game.TailOutcome) error {
	if len(o.Tail.Carry) == 0 {
		return nil
	}
	return game.DrawDoneByID(o.Tail.Carry[0]).Run(g)
})
