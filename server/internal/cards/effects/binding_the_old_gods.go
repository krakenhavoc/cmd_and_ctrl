package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID: "4050e5b3-07c2-461c-9852-a2c081c0198d",
		Name:     "Binding the Old Gods",
		Triggered: []game.TriggeredAbility{
			ChapterTriggerTargeting(1, "Binding the Old Gods — I: destroy target nonland permanent",
				TargetPermanent("target nonland permanent an opponent controls",
					Nonland(), OpponentControls()),
				bindingOldGodsDestroy),
			ChapterTrigger(2, "Binding the Old Gods — II: search for a Forest", bindingOldGodsFetchForest),
			ChapterTrigger(3, "Binding the Old Gods — III: creatures gain deathtouch", bindingOldGodsDeathtouch),
		},
	})
}

func bindingOldGodsDestroy(g *game.Game, item *game.StackItem) error {
	if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
		return nil
	}
	return DestroyTarget{Target: item.Targets[0].ID}.Apply(NewContext(g, item))
}

func bindingOldGodsFetchForest(g *game.Game, item *game.StackItem) error {
	return SearchLibrary{
		Player:        item.Controller,
		Predicate:     IsLandWithSubtype("Forest"),
		Dest:          game.ZoneBattlefield,
		Limit:         1,
		Shuffle:       true,
		TappedOnEntry: true,
		Reason:        "Binding the Old Gods — search for a Forest card",
	}.Apply(NewContext(g, item))
}

func bindingOldGodsDeathtouch(g *game.Game, item *game.StackItem) error {
	return GrantKeywordUntilEOT{
		Match:    And(Creature(), YouControl()),
		Keywords: []string{"deathtouch"},
		Label:    "Binding the Old Gods — creatures you control gain deathtouch",
	}.Apply(NewContext(g, item))
}
