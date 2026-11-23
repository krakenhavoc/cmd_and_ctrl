package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// lrw_b1_helpers.go — the shapes Lorwyn batch 1 (#5593) prints more
// than once: the Harbinger cycle's enters tutor, the tribal reveal-lands'
// card test, and the "a <tribe> put into a graveyard from the
// battlefield" condition two Kindred enchantments read. Its own file, as
// #231 asks of a batch, so a concurrent batch never collides here.
//
// Append-only: add a builder, never change what one means.

// lrwB1Harbinger is the Harbinger cycle's "When this creature enters,
// you may search your library for a <what>, reveal it, then shuffle and
// put that card on top." The "you may" is the trigger's prompt; a yes
// searches, and a search for a card with a stated quality may find
// nothing (CR 701.23b). The found card is shown to the table and goes on
// top after the shuffle, so the controller draws it next.
func LrwB1Harbinger(cardName, what string, pred func(game.Card) bool) game.TriggeredAbility {
	return Optional(
		WhenThisEnters(cardName+" — search for "+what+", reveal it, then shuffle and put it on top",
			func(g *game.Game, item *game.StackItem) error {
				return TutorToTop(NewContext(g, item), cardName+" — "+what, true, pred)
			}),
		cardName+" — search your library for "+what+" and put it on top?")
}

// lrwB1CardOfType is "a <subtype> card": a card of any type whose type
// line carries the subtype (a Kindred card counts as well as a creature),
// a changeling card counting as every creature type (CR 702.73a). It is
// the test on a card in a library or a hand, where no layer applies.
func LrwB1CardOfType(subtype string) func(game.Card) bool {
	return func(c game.Card) bool { return c.HasSubtypeWord(subtype) }
}

// lrwB1PutIntoAGraveyard reports the permanent an EventLTB names when it
// went from the battlefield to a graveyard and was a <subtype> as it last
// existed there (CR 603.10a), whatever its card types: a Kindred
// enchantment of the tribe counts as well as a creature.
func LrwB1PutIntoAGraveyard(ev game.Event, g *game.Game, subtype string) (game.Card, bool) {
	if ev.Kind != game.EventLTB || ev.NewZone != game.ZoneGraveyard {
		return game.Card{}, false
	}
	c, ok := g.LookupCardForEffect(ev.CardID)
	if !ok || !LeftAsSubtype(ev, c, subtype) {
		return game.Card{}, false
	}
	return c, true
}

// lrwB1RegenerateTargetRow is "<cost>: Regenerate target <what>." over a
// permanent clause: the shield goes on the target as the ability
// resolves (CR 701.19a), and a target gone by then gets nothing.
func LrwB1RegenerateTargetRow(label string, cost game.AbilityCost, target *game.TargetSpec) ActivatedAbility {
	return ActivatedAbility{
		Label:   label,
		Purpose: game.Purpose{Answers: game.AnswerSurvive},
		Cost:    cost,
		Targets: target,
		Effect:  RegenerateTheTargetPermanent,
	}
}

// lrwB1RegenerateThisRow is "<cost>: Regenerate this creature."
func LrwB1RegenerateThisRow(cost string) ActivatedAbility {
	return ActivatedAbility{
		Label:   cost + ": Regenerate this creature.",
		Purpose: game.Purpose{Answers: game.AnswerSurvive},
		Cost:    ManaCost(cost),
		Effect: func(g *game.Game, item *game.StackItem) error {
			return Regenerate{Target: item.SourceCardID}.Apply(NewContext(g, item))
		},
	}
}

// lrwB1EachOtherYouControl is "other <subtype>s you control", any
// permanent type: the source is never one of them, and the type is read
// through the layers on the battlefield.
func LrwB1EachOtherYouControl(subtype string) func(target *game.Card, g *game.Game, source *game.Card) bool {
	return func(target *game.Card, _ *game.Game, source *game.Card) bool {
		return target.InstanceID != source.InstanceID && target.Controller == source.Controller &&
			target.HasSubtype(source.TextSubtype(subtype))
	}
}

// eachPlayerDrawsACard is "Each player draws a card." (Mikokoro, Center of
// the Sea; Jace Beleren's +2): one card each, round the table from the
// controller (CR 121.2).
func EachPlayerDrawsACard(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	for _, id := range TablePlayers(ctx) {
		if err := (DrawCards{Player: id, N: 1}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
}

// putAllCreatureCardsFromAllGraveyardsOntoTheBattlefield is "Put all
// creature cards from all graveyards onto the battlefield under your
// control." (Liliana's Talent's granted −8, Liliana Vess's −8): every
// creature card in every graveyard enters at once under the controller.
func PutAllCreatureCardsFromAllGraveyardsOntoTheBattlefield(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	return ReturnFromGraveyardTogether{
		Targets:    AllGraveyardsCardIDs(ctx, game.Card.IsCreature),
		Controller: ctx.Controller(),
	}.Apply(ctx)
}
