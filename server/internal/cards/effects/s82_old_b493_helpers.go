package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b493_helpers.go — shared pieces of S82 OLD batch 493 (Zendikar
// quest enchantments, #5583).

// Old493SearchEquipmentOntoBattlefieldAttach is Quest for the Holy Relic's
// effect: "Search your library for an Equipment card, put it onto the
// battlefield, attach it to a creature you control, then shuffle." The search
// puts the card onto the battlefield; the tail asks which creature the
// controller attaches it to.
func Old493SearchEquipmentOntoBattlefieldAttach(g *game.Game, item *game.StackItem) error {
	return SearchLibrary{
		Player:    item.Controller,
		Source:    item.SourceCardID,
		Predicate: func(c game.Card) bool { return c.HasSubtype("Equipment") },
		Dest:      game.ZoneBattlefield,
		Limit:     1,
		Shuffle:   true,
		Reason:    "Quest for the Holy Relic — an Equipment card",
		ThenTail:  game.KeyedTail{Ref: old493AttachFound, Player: item.Controller, Source: item.SourceCardID},
	}.Apply(NewContext(g, item))
}

// old493AttachFound asks which creature the found Equipment is attached to.
// An on-disk key: never renamed, never reused.
var old493AttachFound = game.RegisterTail("search/old-b493-attach-found-equipment", func(g *game.Game, o game.TailOutcome) error {
	if len(o.IDs) == 0 {
		return nil
	}
	eq := o.IDs[0]
	if !OnBattlefield(g, eq) {
		return nil
	}
	var hosts []uuid.UUID
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == o.Tail.Player && c.IsCreature() && c.InstanceID != eq && g.AttachLegalForEffect(eq, c.InstanceID) {
			hosts = append(hosts, c.InstanceID)
		}
	}
	if len(hosts) == 0 {
		return nil
	}
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:  o.Tail.Player,
		Source:   o.Tail.Source,
		Question: "Quest for the Holy Relic — choose a creature you control to attach the Equipment to",
		Cards:    hosts,
		Min:      1,
		Max:      1,
		LiveZone: game.ZoneBattlefield,
		ThenKey:  old493AttachChosen,
		Carry:    []uuid.UUID{eq},
	})
	return nil
})

var old493AttachChosen = game.RegisterChooseCardsThen("choose-cards/old-b493-attach-equipment", func(g *game.Game, r game.CardsChosen) error {
	if len(r.Picked) == 0 || len(r.Carry) == 0 || !OnBattlefield(g, r.Carry[0]) {
		return nil
	}
	return g.AttachForEffect(r.Carry[0], game.TargetRef{Kind: game.TargetCard, ID: r.Picked[0]})
})

// Old493DrewTwoOrMoreCards is Archmage Ascension's intervening "if you drew
// two or more cards this turn", read as the ability triggers and again as it
// resolves (CR 603.4).
func Old493DrewTwoOrMoreCards(g *game.Game, player uuid.UUID) bool {
	return len(g.DrawnThisTurn[player]) >= 2
}

// Old493ArchmageSearchInstead is Archmage Ascension's replacement body: "you
// may instead search your library for a card, put that card into your hand,
// then shuffle". The rest of the draw instruction goes on once it is done.
func Old493ArchmageSearchInstead(g *game.Game, drawer, source uuid.UUID, done game.DrawDone) error {
	return SearchLibrary{
		Player:  drawer,
		Source:  source,
		Dest:    game.ZoneHand,
		Limit:   1,
		Shuffle: true,
		Reason:  "Archmage Ascension — search your library for a card instead of drawing",
		ThenTail: game.KeyedTail{Ref: old493ArchmageThenDraw, Player: drawer, Source: source,
			Carry: []uuid.UUID{done.Park(g)}},
	}.Apply(KeyedContext(g, drawer, source))
}

var old493ArchmageThenDraw = game.RegisterTail("search/old-b493-archmage-then-rest-of-draw", func(g *game.Game, o game.TailOutcome) error {
	if len(o.Tail.Carry) == 0 {
		return nil
	}
	return game.DrawDoneByID(o.Tail.Carry[0]).Run(g)
})

// Old493SpellSharesNameWithYourGraveyard is Pyromancer Ascension's counter
// condition: "you cast an instant or sorcery spell that has the same name as a
// card in your graveyard".
func Old493SpellSharesNameWithYourGraveyard(ev game.Event, source *game.Card, ch game.Characteristic, g *game.Game) bool {
	if !YouCast(Or(Instant(), Sorcery()))(ev, source, ch, g) {
		return false
	}
	spell, ok := g.LookupCardForEffect(ev.CardID)
	p := g.PlayerByIDForEffect(source.Controller)
	if !ok || p == nil {
		return false
	}
	for _, c := range p.Graveyard.Cards {
		if c.Name == spell.Name {
			return true
		}
	}
	return false
}

// Old493CopyThatSpell is the effect of "you may copy that spell. You may
// choose new targets for the copy."; the "may" is the trigger's.
func Old493CopyThatSpell(g *game.Game, item *game.StackItem) error {
	if item.Trigger == nil {
		return nil
	}
	return CopySpell{
		StackID:          item.Trigger.Event.CardID,
		Controller:       item.Controller,
		FromLastKnown:    true,
		ChooseNewTargets: true,
	}.Apply(NewContext(g, item))
}
