package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Arachnus Spinner — Creature — Spider {5}{G}, 5/7 (S82 GAP2, #5948,
// #6090):
//
//	"Reach
//	 Tap an untapped Spider you control: Search your graveyard and/or
//	 library for a card named Arachnus Web and put it onto the battlefield
//	 attached to target creature. If you search your library this way,
//	 shuffle."
//
// The cost is Honor-Worn Shaku's TapAnUntapped over Spiders: not the {T}
// symbol, so the Spinner itself, or a Spider that arrived this turn, may pay
// it. The search is the engine's "library and/or graveyard" search
// (SearchLibrarySpec.AlsoGraveyard): the player says which zones they
// search before anything is shown, and only a search of the library
// shuffles it and counts as a library search. The card found enters
// attached to the target creature (SearchLibrarySpec.EntersAttachedTo); an
// Arachnus Web that can't legally enchant it stays where it was (CR
// 303.4i). A target that is gone by resolution makes the ability do
// nothing (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "8fea91ee-5849-4c46-8198-fd383d9a55e8",
		Name:            "Arachnus Spinner",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"reach"},
		Activated: []effects.ActivatedAbility{{
			Label:   "Tap an untapped Spider you control: Search your graveyard and/or library for a card named Arachnus Web and put it onto the battlefield attached to target creature. If you search your library this way, shuffle.",
			Cost:    effects.TapAnUntapped("an untapped Spider you control", effects.OfCreatureType("Spider")),
			Targets: effects.TargetCreature("target creature"),
			Effect:  arachnusSpinnerSearch,
		}},
	})
}

const arachnusWebName = "Arachnus Web"

// arachnusSpinnerSearch searches the graveyard and/or library for Arachnus
// Web and puts it onto the battlefield attached to the still-legal target.
func arachnusSpinnerSearch(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	targets := ctx.LegalTargets()
	if len(targets) == 0 || targets[0].Kind != game.TargetCard {
		return nil
	}
	return g.SearchLibraryThenForEffect(game.SearchLibrarySpec{
		Player:           item.Controller,
		Source:           ctx.Source(),
		Pred:             func(c game.Card) bool { return c.Name == arachnusWebName },
		Dest:             game.ZoneBattlefield,
		Limit:            1,
		Shuffle:          true,
		AlsoGraveyard:    true,
		EntersAttachedTo: game.TargetRef{Kind: game.TargetCard, ID: targets[0].ID},
		Reason:           "Arachnus Spinner — search your graveyard and/or library for a card named Arachnus Web",
	})
}
