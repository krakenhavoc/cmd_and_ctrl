package cardskr

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Lobotomy — Sorcery {2}{U}{B} (Invasion, S82 older sets batch 662, #5616):
//
//	"Target player reveals their hand, then you choose a card other than a
//	 basic land card from it. Search that player's graveyard, hand, and
//	 library for all cards with the same name as the chosen card and exile
//	 them. Then that player shuffles."
//
// Reap Intellect's reveal and pick, for one card that is not a basic land card,
// left in the hand for the search to find (Keep), and Extirpate's "all"
// extraction over its name. A hand with nothing to choose still shuffles.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a7ebe748-c3de-402f-8a6e-236a6413c0f8",
		Name:         "Lobotomy",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPlayer("target player"),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			victim := effects.TargetedPlayer(ctx)
			if victim == uuid.Nil {
				return nil
			}
			return effects.ChooseFromRevealedHand{
				Player: victim,
				Filter: effects.NotABasicLandCard(),
				Label:  "card other than a basic land card",
				Keep:   true,
				Then:   lobotomySearch,
			}.Apply(ctx)
		},
	})
}

// lobotomySearch is the clause after the pick: search the victim's graveyard,
// hand and library for every card sharing the chosen card's name and exile
// them; the library is shuffled even when nothing was chosen. An on-disk key.
var lobotomySearch = effects.RevealedPickThen("revealed-pick/s82-old-b662-lobotomy-search", func(ctx *effects.Context, pick game.RevealedPick) error {
	var names []string
	for _, c := range pick.Chosen {
		names = append(names, effects.ExtractionNames(c)...)
	}
	if len(names) == 0 {
		return effects.ShuffleLibrary{Player: pick.FromPlayer}.Apply(ctx)
	}
	return effects.ExtractionSearch(ctx, "Lobotomy", pick.FromPlayer, names, true)
})
