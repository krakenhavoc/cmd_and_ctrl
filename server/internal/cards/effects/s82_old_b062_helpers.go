package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b062_helpers.go — shared pieces of Kaldheim batch 062 (#5505, S82).

// ReturnOneOrTwoSharingCreatureCardsModes is Raise the Draugr's modal text:
// "Choose one — Return target creature card from your graveyard to your hand.
// • Return two target creature cards that share a creature type from your
// graveyard to your hand." The second bullet's second clause is
// SharesACreatureTypeWithTargetCard over the first (a changeling shares with
// any creature card), and the cards are returned as the spell resolves, each
// still-legal target to its owner's hand.
func ReturnOneOrTwoSharingCreatureCardsModes() *game.ModeSpec {
	return ChooseOne(
		ModeDoing("Return target creature card from your graveyard to your hand.",
			TargetCardInGraveyard("target creature card from your graveyard", Creature(), YouOwn()),
			EoeABCharmReturn),
		ModeDoing("Return two target creature cards that share a creature type from your graveyard to your hand.",
			Clauses(
				TargetCardInGraveyard("target creature card from your graveyard", Creature(), YouOwn()),
				Distinct(SharesACreatureTypeWithTargetCard(
					TargetCardInGraveyard("another target creature card from your graveyard that shares a creature type with it", Creature(), YouOwn()), 0)),
			),
			EoeABCharmReturn),
	)
}

// --- Tergrid, God of Fright -------------------------------------------------

// S82OldB062OpponentSacrificedNontokenOrDiscardedPermanent is Tergrid's
// trigger condition: "an opponent sacrifices a nontoken permanent or discards a
// permanent card". The sacrifice event fires with the permanent still on the
// battlefield (so its token-ness is read there) and the discard event after the
// card reached the graveyard (so it is judged as the card it now is, CR 108.2).
func S82OldB062OpponentSacrificedNontokenOrDiscardedPermanent(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.Actor == uuid.Nil || ev.Actor == source.Controller {
		return false
	}
	c, ok := g.LookupCardForEffect(ev.CardID)
	if !ok {
		return false
	}
	switch ev.Kind {
	case game.EventSacrifice:
		return !IsToken(c)
	case game.EventDiscardCard:
		return c.IsPermanent() && !IsToken(c)
	}
	return false
}

// S82OldB062PutThatCardOntoTheBattlefield is Tergrid's effect: "you may put
// that card from a graveyard onto the battlefield under your control". The
// card is looked up wherever it is now, and only a card still in a graveyard
// comes (CR 608.2b), under the trigger's controller.
func S82OldB062PutThatCardOntoTheBattlefield(g *game.Game, item *game.StackItem) error {
	if item.Trigger == nil {
		return nil
	}
	id := item.Trigger.Event.CardID
	if z := g.FindCardZoneForEffect(id); z == nil || z.Kind != game.ZoneGraveyard {
		return nil
	}
	return ReturnFromGraveyard{Target: id, Dest: game.ZoneBattlefield, Controller: item.Controller}.Apply(NewContext(g, item))
}

// --- Tergrid's Lantern ---------------------------------------------------------

const (
	s82OldB062LanternLife = iota
	s82OldB062LanternSacrifice
	s82OldB062LanternDiscard
)

// S82OldB062LanternEffect is "Target player loses 3 life unless they sacrifice
// a nonland permanent of their choice or discard a card.": the target player
// chooses at resolution among taking the loss and each alternative they can
// actually do. A target that is gone does nothing (CR 608.2b).
func S82OldB062LanternEffect(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	var victim uuid.UUID
	for _, t := range ctx.LegalTargets() {
		if t.Kind == game.TargetPlayer {
			victim = t.ID
		}
	}
	if victim == uuid.Nil {
		return nil
	}
	options, _ := s82OldB062LanternOptions(g, victim)
	return PickOption{
		Player:   victim,
		Question: "Tergrid's Lantern — lose 3 life, or sacrifice a nonland permanent, or discard a card?",
		Options:  options,
		ThenKey:  s82OldB062LanternAnswered,
		Carry:    []uuid.UUID{victim},
		WithItem: true,
	}.Apply(ctx)
}

// s82OldB062LanternOptions lists what the victim may do, the life loss first
// (a branch they can always take), and the kind of each option.
func s82OldB062LanternOptions(g *game.Game, victim uuid.UUID) ([]game.ChoiceOption, []int) {
	options := []game.ChoiceOption{{Label: "Lose 3 life"}}
	kinds := []int{s82OldB062LanternLife}
	if len(NonlandPermanentsControlledBy(g, victim)) > 0 {
		options = append(options, game.ChoiceOption{Label: "Sacrifice a nonland permanent"})
		kinds = append(kinds, s82OldB062LanternSacrifice)
	}
	if p := g.PlayerByIDForEffect(victim); p != nil && p.Hand != nil && p.Hand.Size() > 0 {
		options = append(options, game.ChoiceOption{Label: "Discard a card"})
		kinds = append(kinds, s82OldB062LanternDiscard)
	}
	return options, kinds
}

// s82OldB062LanternAnswered is the victim's answer, keyed. An on-disk key:
// never renamed, never reused.
var s82OldB062LanternAnswered = OptionPickThen("option-pick/s82-old-b062-tergrids-lantern", func(ctx *Context, r game.OptionPicked) error {
	if len(r.Carry) == 0 {
		return nil
	}
	victim := r.Carry[0]
	_, kinds := s82OldB062LanternOptions(ctx.Game, victim)
	kind := s82OldB062LanternLife
	if r.Index >= 0 && r.Index < len(kinds) {
		kind = kinds[r.Index]
	}
	switch kind {
	case s82OldB062LanternSacrifice:
		return SacrificeChoice{
			Player:     victim,
			Candidates: NonlandPermanentsControlledBy(ctx.Game, victim),
			Question:   "Tergrid's Lantern — sacrifice a nonland permanent",
		}.Apply(ctx)
	case s82OldB062LanternDiscard:
		ctx.Game.QueueDiscardChoiceForEffect(game.DiscardPrompt{Player: victim, Source: ctx.Source(), N: 1})
		return nil
	}
	return ctx.Game.ChangePlayerLifeForEffect(ctx.Source(), victim, -3)
})
