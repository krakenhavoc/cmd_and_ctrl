package cardss

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Scion of Halaster — Legendary Enchantment — Background {1}{B}:
//
//	"Commander creatures you own have "The first time you would draw a card
//	 each turn, instead look at the top two cards of your library. Put one of
//	 them into your graveyard and the other back on top of your library. Then
//	 draw a card.""
//
// A granted draw replacement (ADR 0093, AbilityGrant.Replacements; Master Chef's
// shape), so the recipient is its source and "you" is whoever controls the
// commander. It is a CR 614.11 draw-instead (Underrealm Lich's shape): applying it
// cancels the draw and runs the body, whose last step is the "then draw a card": a
// fresh draw, after which the rest of the replaced instruction goes on
// (DrawDone.Run). The turn's one application is marked as
// it applies (CR 614.5, DrawBecomes.FirstTimeEachTurn's mark), so that draw, and
// every later one this turn, is a plain draw; a second commander creature with the
// grant has its own once-a-turn replacement.
//
// The look is private to the controller. A library of two or more offers a pick of
// one of the top two for the graveyard (a mandatory pick, in a card-choice prompt
// that waits on the table as data); the other stays on top, in its place. A library
// of one puts that card into the graveyard; an empty one does nothing, and the draw
// that follows loses the game as usual (CR 704.5b). A token is never a card in a
// library and can't be picked.
//
// "Choose a Background" is deck construction (ADR 0144).
//
// No simplification.
const scionOfHalasterGrant = "scion-of-halaster/first-draw-look"

const scionOfHalasterLabel = "Scion of Halaster: the first time you would draw a card each turn, look at the top two cards, put one into your graveyard and draw a card"

func init() {
	effects.Register(effects.Spec{
		OracleID:     "227ddf6c-d21a-48da-84fb-395c7e096914",
		Name:         "Scion of Halaster",
		Completeness: effects.CompletenessFull,
		Grants: []effects.AbilityGrant{{
			Key: scionOfHalasterGrant,
			Replacements: []game.ReplacementEffect{{
				Watches: []game.EventKind{game.EventDrawCard},
				AppliesTo: func(ev *game.ReplacementEvent, g *game.Game, src *game.Card) bool {
					return ev.Kind == game.RepEventDraw && ev.DrawCount > 0 && src != nil && ev.DrawPlayer == src.Controller &&
						!g.ReplacementAppliedThisTurnForEffect(src, scionOfHalasterLabel)
				},
				DrawInstead: game.RegisterDrawInstead("scion-of-halaster", scionOfHalasterLook),
				Replace: func(_ *game.ReplacementEvent, g *game.Game, src *game.Card) error {
					g.MarkReplacementAppliedThisTurnForEffect(src, scionOfHalasterLabel)
					return nil
				},
				Controller: func(_ *game.ReplacementEvent, _ *game.Game, src *game.Card) uuid.UUID {
					return src.Controller
				},
				Label: scionOfHalasterLabel,
			}},
			Text: "The first time you would draw a card each turn, instead look at the top two cards of your library. Put one of them into your graveyard and the other back on top of your library. Then draw a card.",
		}},
		Static: []game.StaticAbility{effects.GrantToCommanderCreaturesYouOwn(scionOfHalasterGrant)},
	})
}

// scionOfHalasterLook is the replacement's body: look at the top two, pick the
// one for the graveyard, and then pay the replaced draw.
func scionOfHalasterLook(g *game.Game, drawer, source uuid.UUID, done game.DrawDone) error {
	looked := g.LookAtTopOfLibraryForEffect(drawer, 2)
	switch len(looked) {
	case 0:
		return done.Run(g)
	case 1:
		if err := effects.RestIntoGraveyard(g, looked); err != nil {
			return err
		}
		return done.Run(g)
	}
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:  drawer,
		Source:   source,
		Question: "Scion of Halaster — put one of the top two cards of your library into your graveyard (the other stays on top)",
		Cards:    looked,
		Min:      1,
		Max:      1,
		LiveZone: game.ZoneLibrary,
		ThenKey:  scionOfHalasterPick,
		Carry:    append([]uuid.UUID{done.Park(g)}, looked...),
	})
	return nil
}

// scionOfHalasterPick puts the chosen card into the graveyard (nothing moves if
// the prompt was withdrawn unanswered) and pays the parked draw. An on-disk key.
var scionOfHalasterPick = game.RegisterChooseCardsThen("choose-cards/scion-of-halaster-pick", func(g *game.Game, r game.CardsChosen) error {
	if len(r.Carry) == 0 {
		return nil
	}
	if r.Answered {
		if err := effects.RestIntoGraveyard(g, r.Picked); err != nil {
			return err
		}
	}
	return scionOfHalasterThenDraw(g, r.Chooser, game.DrawDoneByID(r.Carry[0]))
})

// scionOfHalasterThenDraw is "then draw a card": an ordinary draw, which opens
// its own replacement window (this effect has been used up for the turn, so
// another one may apply), after which the rest of the replaced instruction goes
// on. The parked draw's handle rides the continuation's Source. A draw that
// pauses on a prompt holds the rest behind it. An on-disk key.
func scionOfHalasterThenDraw(g *game.Game, drawer uuid.UUID, done game.DrawDone) error {
	return g.DrawNThenForEffect(drawer, 1, game.DrawThen{Ref: scionOfHalasterRest, Player: drawer, Source: done.Park(g), N: 1})
}

var scionOfHalasterRest = game.RegisterDrawThen("draw-then/scion-of-halaster-rest", func(g *game.Game, d game.DrawThen) error {
	return game.DrawDoneByID(d.Source).Run(g)
})
