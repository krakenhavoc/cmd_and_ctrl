package cardsap

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Assembly Hall — Artifact {5} (Mercadian Masques, S82 OLD, #5619):
//
//	"{4}, {T}: Reveal a creature card in your hand. Search your library for a
//	 card with the same name as that card, reveal it, put it into your hand,
//	 then shuffle."
//
// The reveal is the first instruction of the effect, not a cost: the controller
// names a creature card in hand (nothing to name means nothing to reveal, and no
// search), it is shown to the table, and the search finds a card with that name,
// which may be nothing (CR 701.23b). The name is the revealed card's own.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8da9f975-519d-455a-80c1-2b468cad79a1",
		Name:         "Assembly Hall",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{4}, {T}: Reveal a creature card in your hand. Search your library for a card with the same name as that card, reveal it, put it into your hand, then shuffle.",
			Purpose: game.Purpose{Tutors: 1, Answers: game.AnswerValue},
			Cost:    effects.Plus(effects.ManaCost("{4}"), effects.TapCost()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return assemblyHallReveal(g, item)
			},
		}},
	})
}

// assemblyHallReveal asks for a creature card in the controller's hand.
func assemblyHallReveal(g *game.Game, item *game.StackItem) error {
	p := g.PlayerByIDForEffect(item.Controller)
	if p == nil || p.Hand == nil {
		return nil
	}
	var creatures []uuid.UUID
	for _, c := range p.Hand.Cards {
		if c.IsCreature() {
			creatures = append(creatures, c.InstanceID)
		}
	}
	if len(creatures) == 0 {
		return nil
	}
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:  item.Controller,
		Source:   item.SourceCardID,
		Question: "Assembly Hall — reveal a creature card in your hand",
		Cards:    creatures,
		Min:      1,
		Max:      1,
		LiveZone: game.ZoneHand,
		ThenKey:  assemblyHallAssemblyHallRevealed,
		Asked:    game.NewAskedItem(item),
	})
	return nil
}

// assemblyHallAssemblyHallRevealed shows the chosen card, then searches for one with
// its name. An on-disk key: never renamed, never reused.
var assemblyHallAssemblyHallRevealed = game.RegisterChooseCardsThen("choose-cards/assembly-hall-reveal", func(g *game.Game, r game.CardsChosen) error {
	if len(r.Picked) == 0 || r.Asked == nil {
		return nil
	}
	c, ok := g.LookupCardForEffect(r.Picked[0])
	if !ok {
		return nil
	}
	name := c.Name
	g.RevealForEffect(game.RevealSpec{
		Player: r.Chooser, Source: r.Source,
		Reason: "Assembly Hall — reveal a creature card in your hand",
		Cards:  []uuid.UUID{c.InstanceID},
	})
	return effects.SearchLibrary{
		Player:    r.Chooser,
		Predicate: func(l game.Card) bool { return l.Name == name },
		Dest:      game.ZoneHand,
		Limit:     1,
		Reveal:    true,
		Shuffle:   true,
		Reason:    "Assembly Hall — a card named " + name,
	}.Apply(effects.NewContext(g, r.Asked))
})
