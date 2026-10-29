package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Elven Passage — Land (The Hobbit):
//
//	"{T}, Pay 1 life, Sacrifice this land: Search your library for a basic
//	 land card, put it onto the battlefield tapped, then shuffle. You may
//	 behold an Elf. If you do, untap that land. (To behold an Elf, choose
//	 an Elf you control or reveal an Elf card from your hand.)"
//
// #4454, S76. Fabled Passage's search with its untap told the land that was
// found (a registered tail), and the fetchlands' {T}, pay 1 life, sacrifice
// cost. Behold is "choose an Elf you control or reveal an Elf card from
// your hand" (the same action BeholdOrPay costs name): choosing an Elf you
// control is free and untapping the land is never worse, so a controlled Elf
// untaps it with no question; otherwise a player holding Elf cards is asked
// which one to reveal, and may answer none.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "97a2cd39-6b54-496b-b3ac-dab9dfed7edc",
		Name:         "Elven Passage",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{T}, Pay 1 life, Sacrifice this land: Search your library for a basic land card, put it onto the battlefield tapped, then shuffle. You may behold an Elf. If you do, untap that land.",
			Purpose: game.Purpose{Answers: game.AnswerValue, Lands: 1},
			Cost:    fetchlandCost(),
			Effect:  elvenPassageFetch,
		}},
	})
}

func elvenPassageFetch(g *game.Game, item *game.StackItem) error {
	controller := item.Controller
	ctx := NewContext(g, item)
	return SearchLibrary{
		Player:        controller,
		Predicate:     IsBasicLand,
		Dest:          game.ZoneBattlefield,
		Limit:         1,
		Reveal:        true,
		Shuffle:       true,
		TappedOnEntry: true,
		Reason:        "Elven Passage — a basic land",
		ThenTail:      ctx.ItemTail(elvenPassageBehold),
	}.Apply(ctx)
}

// elvenPassageBehold is "you may behold an Elf. If you do, untap that land",
// told the land that was found. An on-disk key.
var elvenPassageBehold = CardTail("elven-passage/behold-an-elf", func(ctx *Context, o game.TailOutcome) error {
	g := ctx.Game
	// A whiffed or declined search still cost the land; there is nothing
	// to untap.
	if len(o.IDs) == 0 {
		return nil
	}
	land, player := o.IDs[0], o.Tail.Player
	// "Elf" is a word of this card's text (CR 612.2).
	elf := ctx.TextSubtype("Elf")
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == player && c.HasSubtype(elf) {
			return g.UntapTargetForEffect(land)
		}
	}
	p := g.PlayerByIDForEffect(player)
	if p == nil {
		return nil
	}
	var elves []uuid.UUID
	for _, c := range p.Hand.Cards {
		if c.HasSubtype(elf) {
			elves = append(elves, c.InstanceID)
		}
	}
	if len(elves) == 0 {
		return nil
	}
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:  player,
		Source:   o.Tail.Source,
		Question: "Elven Passage — reveal an Elf card from your hand to untap the land?",
		Cards:    elves,
		Min:      0,
		Max:      1,
		ThenKey:  elvenPassageReveal,
		LiveZone: game.ZoneHand,
		Carry:    []uuid.UUID{land},
	})
	return nil
})

// elvenPassageReveal is the Elf card revealed to behold: the reveal, then the
// untap of the land carried. Keyed (#4253).
var elvenPassageReveal = game.RegisterChooseCardsThen("choose-cards/elven-passage-behold", func(g *game.Game, r game.CardsChosen) error {
	if len(r.Picked) == 0 || len(r.Carry) == 0 {
		return nil
	}
	g.RevealForEffect(game.RevealSpec{
		Player: r.Chooser,
		Source: r.Source,
		Reason: "Elven Passage — behold an Elf",
		Cards:  r.Picked,
	})
	return g.UntapTargetForEffect(r.Carry[0])
})
