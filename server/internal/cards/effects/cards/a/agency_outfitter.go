package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

var agencyOutfitterNames = []string{"Magnifying Glass", "Thinking Cap"}

// Agency Outfitter — Creature — Sphinx Detective {4}{U}{U}, 4/3 (Murders
// at Karlov Manor, #4469):
//
//	"Flying
//	 When this creature enters, you may search your graveyard, hand
//	 and/or library for a card named Magnifying Glass and/or a card named
//	 Thinking Cap and put them onto the battlefield. If you search your
//	 library this way, shuffle."
//
// search_all_zones_named.go: "search your library too?" (the shuffle
// follows only a library search), then one pick over the named cards in
// the graveyard, hand and, if asked for, library, one per name, put
// onto the battlefield together.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "23b2de61-a29c-4a32-b38c-408c723727ce",
		Name:            "Agency Outfitter",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Agency Outfitter — search your graveyard, hand and/or library for Magnifying Glass and/or Thinking Cap",
				func(g *game.Game, item *game.StackItem) error {
					return effects.SearchAllZonesNamed(effects.NewContext(g, item),
						"Agency Outfitter — search your library as well as your graveyard and hand? (you shuffle if you do)",
						agencyOutfitterLibrary)
				}),
		},
	})
}

// agencyOutfitterLibrary is the "search your library?" answer, keyed.
var agencyOutfitterLibrary = effects.MayAnswer("confirm/agency-outfitter-library", func(ctx *effects.Context, r game.ConfirmAnswer) error {
	word := ""
	if r.Accepted {
		word = "library"
	}
	cands := effects.SearchAllZonesCandidates(ctx.Game, ctx.Controller(), agencyOutfitterNames, r.Accepted)
	if len(cands) == 0 {
		return effects.SearchAllZonesPut(ctx, nil, r.Accepted)
	}
	ctx.Game.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:   ctx.Controller(),
		Source:    ctx.Source(),
		Question:  "Agency Outfitter — choose a card named Magnifying Glass and/or a card named Thinking Cap to put onto the battlefield",
		Cards:     cands,
		Min:       0,
		Max:       2,
		ThenKey:   agencyOutfitterPick,
		CarryWord: word,
		Asked:     game.NewAskedItem(ctx.Item),
	})
	return nil
})

// agencyOutfitterPick puts the chosen cards onto the battlefield, keyed.
var agencyOutfitterPick = effects.AskedPickRule("choose-cards/agency-outfitter-search", effects.OneCardPerName,
	func(ctx *effects.Context, r game.CardsChosen) error {
		return effects.SearchAllZonesPut(ctx, r.Picked, r.Word == "library")
	})
