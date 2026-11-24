package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Runed Crown — Artifact — Equipment {3} (Kaldheim, #5505):
//
//	"When this Equipment enters, you may search your library, hand, and/or
//	 graveyard for a Rune card and put it onto the battlefield attached to
//	 this Equipment. If you search your library this way, shuffle.
//	 Equipped creature gets +1/+1.
//	 Equip {2}"
//
// SearchLibrary with the hand and graveyard beside the library: the
// searcher is asked which zones to search first, and the library is
// shuffled if and only if they chose it. The Rune enters attached to the
// Crown (CR 303.4f); one that can't legally enchant it stays where it was.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b8739fc4-6308-4e49-be2e-d9d92d890ee1",
		Name:         "Runed Crown",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{effects.PumpAttached(1, 1)},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Runed Crown — you may search your library, hand, and/or graveyard for a Rune card and put it onto the battlefield attached to this Equipment",
				runedCrownSearch),
		},
		Activated: []effects.ActivatedAbility{effects.EquipAbility("{2}")},
	})
}

func runedCrownSearch(g *game.Game, item *game.StackItem) error {
	if !effects.OnBattlefield(g, item.SourceCardID) || effects.SourceIsNewObject(g, item) {
		return nil
	}
	return effects.SearchLibrary{
		Player:           item.Controller,
		Predicate:        func(c game.Card) bool { return c.HasSubtype("Rune") },
		Dest:             game.ZoneBattlefield,
		Limit:            1,
		Shuffle:          true,
		Optional:         true,
		AlsoGraveyard:    true,
		AlsoHand:         true,
		Reason:           "Runed Crown — a Rune card",
		EntersAttachedTo: game.TargetRef{Kind: game.TargetCard, ID: item.SourceCardID},
	}.Apply(effects.NewContext(g, item))
}
