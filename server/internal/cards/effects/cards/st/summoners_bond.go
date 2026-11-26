package cardsst

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Summoner's Bond — Conspiracy (Conspiracy: Take the Crown, S82 OLD batch
// 267, #5540):
//
//	"Double agenda
//	 Whenever you cast a creature spell with one of the chosen names, you may
//	 search your library for a creature card with the other chosen name,
//	 reveal it, put it into your hand, then shuffle."
//
// Double agenda (CR 702.106f, ADR 0227): two different names. The spell is
// matched against each name on its cast face, and the search looks for a
// creature card with the OTHER name. Optional, and the search may find
// nothing (CR 701.23b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "07b6326a-b73d-422b-bcb5-c508aa3fc558",
		Name:         "Summoner's Bond",
		Completeness: effects.CompletenessFull,
		HiddenAgenda: effects.HiddenAgendaTwoNames,
		Purpose:      game.Purpose{Tutors: 1},
		Triggered: []game.TriggeredAbility{
			effects.Optional(
				effects.On(game.EventCast, effects.YouCastASpellWithTheChosenName(func(_ *game.Game, _ uuid.UUID, c game.Card) bool { return c.IsCreature() }),
					"Summoner's Bond — search for a creature card with the other chosen name",
					summonersBondSearch),
				"Summoner's Bond — search your library for a creature card with the other chosen name?"),
		},
	})
}

func summonersBondSearch(g *game.Game, item *game.StackItem) error {
	var bond *game.Card
	for _, c := range g.ConspiraciesOf(item.Controller) {
		if c.InstanceID == item.SourceCardID {
			c := c
			bond = &c
		}
	}
	spell, ok := g.LookupCardForEffect(item.Trigger.Event.CardID)
	if bond == nil || !ok {
		return nil
	}
	var other []string
	for _, n := range game.ChosenNamesOn(bond) {
		if !game.SpellOrLandHasName(spell, n) {
			other = append(other, n)
		}
	}
	if len(other) == 0 {
		return nil
	}
	return effects.SearchLibrary{
		Player: item.Controller,
		Predicate: func(c game.Card) bool {
			if !c.IsCreature() {
				return false
			}
			for _, n := range other {
				if game.CardNameMatches(c, n) {
					return true
				}
			}
			return false
		},
		Dest:     game.ZoneHand,
		Limit:    1,
		Reveal:   true,
		Shuffle:  true,
		Optional: true,
		Reason:   "Summoner's Bond — a creature card with the other chosen name",
	}.Apply(effects.NewContext(g, item))
}
