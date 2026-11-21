package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Binding the Old Gods — Enchantment — Saga for {2}{B}{G}:
//
//	"I — Destroy target nonland permanent an opponent controls.
//	 II — Search your library for a Forest card, put it onto the
//	      battlefield tapped, then shuffle.
//	 III — Creatures you control gain deathtouch until end of turn."
//
// Chapter II fetches a FOREST, not a basic Forest — the printed text
// says "a Forest card", which is any card with the Forest land type,
// so a Bayou and an Overgrown Tomb both qualify. IsLandWithSubtype
// is the predicate that reads the land type rather than the Basic
// supertype; using IsBasicLand here would ship the card weaker than
// printed in exactly the decks that play it.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "4050e5b3-07c2-461c-9852-a2c081c0198d",
		Name:         "Binding the Old Gods",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.ChapterTriggerTargeting(1, "Binding the Old Gods — I: destroy target nonland permanent",
				effects.TargetPermanent("target nonland permanent an opponent controls",
					effects.Nonland(), effects.OpponentControls()),
				effects.DestroyChosenPermanent),
			effects.ChapterTrigger(2, "Binding the Old Gods — II: search for a Forest", bindingOldGodsFetchForest),
			effects.ChapterTrigger(3, "Binding the Old Gods — III: creatures gain deathtouch", bindingOldGodsDeathtouch),
		},
	})
}

func bindingOldGodsFetchForest(g *game.Game, item *game.StackItem) error {
	return effects.SearchLibrary{
		Player:        item.Controller,
		Predicate:     effects.IsLandWithSubtype("Forest"),
		Dest:          game.ZoneBattlefield,
		Limit:         1,
		Shuffle:       true,
		TappedOnEntry: true,
		Reason:        "Binding the Old Gods — search for a Forest card",
	}.Apply(effects.NewContext(g, item))
}

func bindingOldGodsDeathtouch(g *game.Game, item *game.StackItem) error {
	return effects.GrantKeywordUntilEOT{
		Match:    effects.And(effects.Creature(), effects.YouControl()),
		Keywords: []string{"deathtouch"},
		Label:    "Binding the Old Gods — creatures you control gain deathtouch",
	}.Apply(effects.NewContext(g, item))
}
