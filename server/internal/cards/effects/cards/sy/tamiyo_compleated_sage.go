package cardssy

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Tamiyo, Compleated Sage — {2}{G}{G/U/P}{U} Legendary Planeswalker —
// Tamiyo, loyalty 5 (Kamigawa: Neon Dynasty, S82 NEO batch 20):
//
//	"Compleated ({G/U/P} can be paid with {G}, {U}, or 2 life. If life was
//	 paid, this planeswalker enters with two fewer loyalty counters.)
//	 +1: Tap up to one target artifact or creature. It doesn't untap during
//	 its controller's next untap step.
//	 −X: Exile target nonland permanent card with mana value X from your
//	 graveyard. Create a token that's a copy of that card.
//	 −7: Create Tamiyo's Notebook, a legendary colorless Book artifact token
//	 with "Spells you cast cost {2} less to cast" and "{T}: Draw a card.""
//
// Compleated is the shared replacement (compleated.go). The −X announces X
// with the activation (CR 107.3a) and the target clause reads it
// (WithManaValueEqualsX), checked as it is announced and again as the
// ability resolves. The card is exiled, then a token copy of it is made
// from its copiable values (CR 707.2), whether or not the exile reached
// exile. The Notebook is a catalog token (s82_neo_b20_tokens.go): its cost
// reduction and tap-draw work from the battlefield.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "021b4745-4a15-462c-a6b0-2dee1dc9d196",
		Name:            "Tamiyo, Compleated Sage",
		Completeness:    effects.CompletenessFull,
		StartingLoyalty: 5,
		PrintedKeywords: []string{"compleated"},
		Replacements:    []game.ReplacementEffect{effects.Compleated()},
		Activated: []effects.ActivatedAbility{
			{
				Label:   "+1: Tap up to one target artifact or creature. It doesn't untap during its controller's next untap step.",
				Cost:    effects.LoyaltyCost(1),
				Targets: effects.TargetPermanent("up to one target artifact or creature", effects.Or(effects.Artifact(), effects.Creature())).WithCount(0, 1),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					id, ok := effects.B16FirstLegalTargetCard(ctx)
					if !ok {
						return nil
					}
					return effects.TapAndFreeze{Targets: []uuid.UUID{id}, Label: "Tamiyo, Compleated Sage — doesn't untap during its controller's next untap step"}.Apply(ctx)
				},
			},
			{
				Label: "−X: Exile target nonland permanent card with mana value X from your graveyard. Create a token that's a copy of that card.",
				Cost:  effects.LoyaltyMinusX(),
				Targets: effects.TargetCardInGraveyard("target nonland permanent card with mana value X from your graveyard",
					effects.Permanent(), effects.Nonland(), effects.YouOwn()).WithManaValueEqualsX(),
				Purpose: game.Purpose{Tokens: 1},
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					id, ok := effects.B16FirstLegalTargetCard(ctx)
					if !ok {
						return nil
					}
					return effects.ExileTarget{
						Target: id,
						ThenTail: game.KeyedTail{
							Ref: tamiyoCompleatedSageCopy, Player: item.Controller, Source: item.SourceCardID,
							Carry: []uuid.UUID{id},
						},
					}.Apply(ctx)
				},
			},
			{
				Label:   "−7: Create Tamiyo's Notebook, a legendary colorless Book artifact token with \"Spells you cast cost {2} less to cast\" and \"{T}: Draw a card.\"",
				Cost:    effects.LoyaltyCost(-7),
				Purpose: game.Purpose{Tokens: 1},
				Effect:  effects.Do(effects.CreateToken{Template: effects.TamiyosNotebookToken(), N: 1}),
			},
		},
	})
}

// tamiyoCompleatedSageCopy is the −X's second sentence: a token that's a
// copy of the exiled card. An on-disk key: never renamed, never reused.
var tamiyoCompleatedSageCopy = effects.CardTail("then/tamiyo-compleated-sage-copy", func(ctx *effects.Context, o game.TailOutcome) error {
	if len(o.Tail.Carry) == 0 {
		return nil
	}
	return effects.CreateTokenCopy{Controller: o.Tail.Player, Copy: o.Tail.Carry[0], N: 1}.Apply(ctx)
})
