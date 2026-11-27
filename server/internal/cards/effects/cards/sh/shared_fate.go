package cardssh

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Shared Fate — Enchantment {4}{U} (Mirrodin, S82 OLD, #5606):
//
//	"If a player would draw a card, that player exiles the top card of one
//	 of their opponents' libraries face down instead.
//	 Each player may look at cards they exiled with this enchantment, and
//	 they may play lands and cast spells from among those cards."
//
// A draw-instead replacement (draw_instead.go, Hullbreacher's shape) on
// every player's draws, whoever controls the enchantment: the draw is
// cancelled (CR 614.6), the drawing player picks one of their opponents
// whose library has a card (asked only when there are two or more), and
// the top card goes to exile face down with a play permission for the
// drawer alone, which is also what lets only them look at it. The
// permission carries no duration beyond staying exiled and does not end if
// Shared Fate leaves the battlefield, because the printed text has no such
// limit, and it is "play lands and cast spells": the cards are paid for
// normally, a land is a land drop, and nothing is made free or any-colour.
// A draw from a library that is empty is still replaced (CR 614.11), so
// nobody loses to it; with no opponent card to exile, nothing is exiled.
// The rest of the draw instruction waits behind the choice (a parked
// DrawDone) and goes on once the card is exiled.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e23bf5fe-e494-473e-8502-606106ee1820",
		Name:         "Shared Fate",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{{
			Watches: []game.EventKind{game.EventDrawCard},
			AppliesTo: func(ev *game.ReplacementEvent, g *game.Game, _ *game.Card) bool {
				return ev.Kind == game.RepEventDraw && ev.DrawPlayer != uuid.Nil && g.PlayerByIDForEffect(ev.DrawPlayer) != nil
			},
			DrawInstead: sharedFateDraw,
			Controller: func(ev *game.ReplacementEvent, _ *game.Game, src *game.Card) uuid.UUID {
				if ev.DrawPlayer != uuid.Nil {
					return ev.DrawPlayer
				}
				return src.Controller
			},
			Label: "Shared Fate: instead of drawing, exile the top card of an opponent's library face down",
		}},
	})
}

// sharedFateDraw is the replaced draw: pick the library, exile its top card
// face down for the drawer, then pay the rest of the instruction. A registered
// body (an on-disk key).
var sharedFateDraw = game.RegisterDrawInstead("s82-old-b602-shared-fate", func(g *game.Game, drawer, source uuid.UUID, done game.DrawDone) error {
	var libraries []uuid.UUID
	for _, p := range g.Seats {
		if p == nil || p.ID == drawer || p.Eliminated || p.Library == nil || len(p.Library.Cards) == 0 {
			continue
		}
		libraries = append(libraries, p.ID)
	}
	switch len(libraries) {
	case 0:
		return done.Run(g)
	case 1:
		if err := sharedFateExile(g, drawer, source, libraries[0]); err != nil {
			return err
		}
		return done.Run(g)
	}
	g.QueueChoosePlayerForEffect(game.ChoosePlayerPrompt{
		Chooser:  drawer,
		Among:    libraries,
		Source:   source,
		Question: "Shared Fate — exile the top card of which opponent's library face down?",
		ThenKey:  sharedFateChosen,
		Carry:    []uuid.UUID{done.Park(g)},
	})
	return nil
})

// sharedFateChosen exiles from the library chosen, then pays the rest of
// the draw. An on-disk key: never renamed, never reused.
var sharedFateChosen = game.RegisterOptionPickThen("option-pick/s82-old-b602-shared-fate-library", func(g *game.Game, r game.OptionPicked) error {
	if r.Option != nil && r.Option.Player != uuid.Nil {
		if err := sharedFateExile(g, r.Chooser, r.Source, r.Option.Player); err != nil {
			return err
		}
	}
	if len(r.Carry) == 0 {
		return nil
	}
	return game.DrawDoneByID(r.Carry[0]).Run(g)
})

// sharedFateExile exiles the top card of `from`'s library face down, and
// lets `drawer` play it for as long as it stays exiled.
func sharedFateExile(g *game.Game, drawer, source, from uuid.UUID) error {
	return g.ExileTopFaceDownWithPermissionForEffect(from, drawer, 1, game.CastPermission{
		Duration:   game.WhileInZoneDuration(),
		Source:     source,
		SourceName: "Shared Fate",
		Label:      "You may play lands and cast spells from among the cards you exiled with Shared Fate",
	})
}
