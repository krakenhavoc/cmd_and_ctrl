package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
const mostDecrepitOldBirdOracleID = "34d34f2c-06ff-4acf-8f2f-ff867774b777"

func init() {
	Register(Spec{
		OracleID:        mostDecrepitOldBirdOracleID,
		Name:            "Most Decrepit Old Bird",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Static:          []game.StaticAbility{fraThresholdSelfPT(1, 1)},
	})
	Register(Spec{
		OracleID:     mostDecrepitOldBirdOracleID + "#1",
		Name:         "Speak Secrets",
		Completeness: CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return hobB9MillFour(ctx, speakSecretsMilled)
		},
	})
}

// speakSecretsMilled is the clause after the mill: take an instant or sorcery
// from among the milled cards.
var speakSecretsMilled = CardTail("speak-secrets/after-mill", func(ctx *Context, o game.TailOutcome) error {
	return hobB9TakeFromMilled(ctx, o.IDs, Or(Instant(), Sorcery()),
		"Speak Secrets — put an instant or sorcery card from among them into your hand", 1, 1)
})
