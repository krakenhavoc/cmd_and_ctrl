package cardss

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Search the City — Enchantment {4}{U} (Return to Ravnica, #5566):
//
//	"When this enchantment enters, exile the top five cards of your library.
//	 Whenever you play a card with the same name as one of the exiled cards,
//	 you may put one of those cards with that name into its owner's hand. Then
//	 if there are no cards exiled with this enchantment, sacrifice it. If you
//	 do, take an extra turn after this one."
//
// The exile is a plain move of the top five (MillToZone to exile, so it is not
// a mill), recorded against the enchantment by the label of the trigger that
// made it (Pit of Offerings' record, CR 400.7: an enchantment that left and
// came back remembers nothing). The second ability fires on a land played or a
// spell cast by you (WheneverYouPlayACard) whose name matches an exiled card.
// The "may" is the prompt's floor of zero, one card at most; the "then" half
// runs whatever was answered, and finds the enchantment's cards as they are
// then: with none left it is sacrificed, and only if it was (a sacrifice that
// can't happen is not "if you do") its controller takes the extra turn.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "fe1b0b19-0994-4bc7-ab6c-d1419d274f40",
		Name:         "Search the City",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters(searchTheCityLabel, func(g *game.Game, item *game.StackItem) error {
				return effects.MillToZone{Player: item.Controller, N: 5, To: game.ZoneExile}.Apply(effects.NewContext(g, item))
			}),
			effects.OnAny([]game.EventKind{game.EventCast, game.EventZoneMove},
				func(ev game.Event, source *game.Card, ch game.Characteristic, g *game.Game) bool {
					if !effects.YouPlayedACard(ev, source, ch, g) {
						return false
					}
					played, ok := g.LookupCardForEffect(ev.CardID)
					if !ok {
						return false
					}
					return len(searchTheCityMatches(g, source.InstanceID, played)) > 0
				},
				"Search the City — you may put one of the exiled cards with that name into its owner's hand; then if none are left, sacrifice it, and if you do, take an extra turn",
				searchTheCityPlayed),
		},
	})
}

const searchTheCityLabel = "Search the City — exile the top five cards of your library"

// searchTheCityMatches are the cards exiled with the enchantment that share the
// played card's name, in the order they were exiled.
func searchTheCityMatches(g *game.Game, enchantment uuid.UUID, played game.Card) []uuid.UUID {
	var out []uuid.UUID
	for _, id := range effects.B27ExiledWith(g, enchantment, searchTheCityLabel) {
		if c, ok := g.LookupCardForEffect(id); ok && c.Name != "" && game.CardNameMatches(played, c.Name) {
			out = append(out, id)
		}
	}
	return out
}

// searchTheCityPlayed is the second ability's body: offer the matching cards,
// or go straight to the "then" half when none is left to offer.
func searchTheCityPlayed(g *game.Game, item *game.StackItem) error {
	if item.Trigger == nil || !effects.SourceIsStillThisPermanent(g, item) {
		return nil
	}
	played, ok := g.LookupCardForEffect(item.Trigger.Event.CardID)
	if !ok {
		return nil
	}
	candidates := searchTheCityMatches(g, item.SourceCardID, played)
	if len(candidates) == 0 {
		return searchTheCityThen(g, item.Controller, item.SourceCardID)
	}
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:  item.Controller,
		Source:   item.SourceCardID,
		Question: "Search the City — you may put one of those cards into its owner's hand",
		Cards:    candidates,
		Min:      0,
		Max:      1,
		LiveZone: game.ZoneExile,
		ThenKey:  searchTheCityPick,
	})
	return nil
}

// searchTheCityPick puts the chosen card (if any) into its owner's hand and
// then runs the sacrifice half. Keyed: the pick holds no closure, so a table
// waiting on it is a restore point. An on-disk key: never renamed, never reused.
var searchTheCityPick = game.RegisterChooseCardsThen("choose-cards/s82-old-b426-search-the-city", func(g *game.Game, r game.CardsChosen) error {
	for _, id := range r.Picked {
		if err := g.BounceToHandForEffect(id); err != nil {
			return err
		}
	}
	return searchTheCityThen(g, r.Chooser, r.Source)
})

// searchTheCityThen is "Then if there are no cards exiled with this enchantment,
// sacrifice it. If you do, take an extra turn after this one."
func searchTheCityThen(g *game.Game, controller, enchantment uuid.UUID) error {
	if !effects.OnBattlefield(g, enchantment) || len(effects.B27ExiledWith(g, enchantment, searchTheCityLabel)) > 0 {
		return nil
	}
	if err := g.SacrificePermanentForEffect(enchantment); err != nil {
		return err
	}
	if effects.OnBattlefield(g, enchantment) {
		return nil
	}
	g.TakeExtraTurnsForEffect(controller, enchantment, 1)
	return nil
}
