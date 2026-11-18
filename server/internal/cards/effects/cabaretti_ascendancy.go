package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Cabaretti Ascendancy — Enchantment {R}{G}{W} (Streets of New Capenna, #5491):
//
//	"At the beginning of your upkeep, look at the top card of your library.
//	 If it's a creature or planeswalker card, you may reveal it and put it
//	 into your hand. If you don't put the card into your hand, you may put it
//	 on the bottom of your library."
//
// Bucolic Ranch's look: the look is private and the answer is one pick over
// what the rules allow — reveal it into your hand (only when it is a creature
// or planeswalker card, read as it sits in the library), put it on the
// bottom, or leave it on top (CR 608.2d).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "c35c39be-6db4-4a80-ace1-07932d1bb460",
		Name:         "Cabaretti Ascendancy",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventBeginUpkeep, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return ev.Actor == source.Controller
			}, "Cabaretti Ascendancy — look at the top card of your library; if it's a creature or planeswalker card, you may reveal it and put it into your hand; if you don't, you may put it on the bottom",
				cabarettiAscendancyLook),
		},
	})
}

const (
	cabarettiAscendancyTake   = "Reveal it and put it into your hand"
	cabarettiAscendancyBottom = "Put it on the bottom of your library"
	cabarettiAscendancyLeave  = "Leave it on top"
)

// cabarettiAscendancyLook looks at the top card and asks what to do with it.
func cabarettiAscendancyLook(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	you := ctx.Controller()
	top := g.LookAtTopOfLibraryForEffect(you, 1)
	if len(top) == 0 {
		return nil
	}
	options := []game.ChoiceOption{{Label: cabarettiAscendancyBottom}, {Label: cabarettiAscendancyLeave}}
	if c, ok := g.LookupCardForEffect(top[0]); ok && !c.IsToken() && (c.IsCreature() || c.IsPlaneswalker()) {
		options = append([]game.ChoiceOption{{Label: cabarettiAscendancyTake}}, options...)
	}
	return PickOption{
		Question: "Cabaretti Ascendancy — what do you do with the top card of your library?",
		Options:  options,
		ThenKey:  cabarettiAscendancyAnswer,
		Carry:    []uuid.UUID{top[0]},
	}.Apply(ctx)
}

// cabarettiAscendancyAnswer carries out the pick. An on-disk key: never
// renamed, never reused.
var cabarettiAscendancyAnswer = OptionPickThen("option-pick/cabaretti-ascendancy-top-card", func(ctx *Context, r game.OptionPicked) error {
	if r.Option == nil || len(r.Carry) == 0 {
		return nil
	}
	g, id := ctx.Game, r.Carry[0]
	switch r.Option.Label {
	case cabarettiAscendancyTake:
		g.RevealForEffect(game.RevealSpec{Player: r.Chooser, Source: r.Source, Reason: "Cabaretti Ascendancy — reveal the creature or planeswalker card", Cards: []uuid.UUID{id}})
		g.TakeFromLibraryToHandForEffect(r.Chooser, []uuid.UUID{id})
	case cabarettiAscendancyBottom:
		return g.PutOnBottomInRandomOrderForEffect(r.Chooser, game.ZoneLibrary, []uuid.UUID{id})
	}
	return nil
})
