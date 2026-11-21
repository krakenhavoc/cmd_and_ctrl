package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Most Decrepit Old Bird // Speak Secrets — an Adventure card (CR 715,
// The Hobbit, #4454), oracle 34d34f2c…:
//
//	Most Decrepit Old Bird — Creature — Bird {U}, 1/1
//	  "Flying
//	   Threshold — This creature gets +1/+1 as long as there are seven or
//	   more cards in your graveyard."
//	Speak Secrets — Sorcery — Adventure {1}{U}
//	  "Mill four cards, then put an instant or sorcery card from among
//	   them into your hand. (Then exile this card. You may cast the
//	   creature later from exile.)"
//
// Threshold is the layer 7c self-modifier read live on every recompute
// (fraThresholdSelfPT). The pick is not a "may": with an instant or
// sorcery among the four, the player must take one (the prompt floors at
// one). "From among them" is the cards that really reached the graveyard.
//
// No simplification.

// mostDecrepitOldBirdOracleID is shared by both faces of the card.
const MostDecrepitOldBirdOracleID = "34d34f2c-06ff-4acf-8f2f-ff867774b777"

func init() {
	effects.Register(effects.Spec{
		OracleID:        MostDecrepitOldBirdOracleID,
		Name:            "Most Decrepit Old Bird",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Static:          []game.StaticAbility{effects.FraThresholdSelfPT(1, 1)},
	})
	effects.Register(effects.Spec{
		OracleID:     MostDecrepitOldBirdOracleID + "#1",
		Name:         "Speak Secrets",
		Completeness: effects.CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.HobB9MillFour(ctx, speakSecretsMilled)
		},
	})
}

// speakSecretsMilled is the clause after the mill: take an instant or sorcery
// from among the milled cards.
var speakSecretsMilled = effects.CardTail("speak-secrets/after-mill", func(ctx *effects.Context, o game.TailOutcome) error {
	return effects.HobB9TakeFromMilled(ctx, o.IDs, effects.Or(effects.Instant(), effects.Sorcery()),
		"Speak Secrets — put an instant or sorcery card from among them into your hand", 1, 1)
})
