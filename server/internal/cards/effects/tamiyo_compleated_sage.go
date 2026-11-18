package effects

import (
	"github.com/google/uuid"

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
	Register(Spec{
		OracleID:        "021b4745-4a15-462c-a6b0-2dee1dc9d196",
		Name:            "Tamiyo, Compleated Sage",
		Completeness:    CompletenessFull,
		StartingLoyalty: 5,
		PrintedKeywords: []string{"compleated"},
		Replacements:    []game.ReplacementEffect{Compleated()},
		Activated: []ActivatedAbility{
			{
				Label:   "+1: Tap up to one target artifact or creature. It doesn't untap during its controller's next untap step.",
				Cost:    LoyaltyCost(1),
				Targets: TargetPermanent("up to one target artifact or creature", Or(Artifact(), Creature())).WithCount(0, 1),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					id, ok := b16FirstLegalTargetCard(ctx)
					if !ok {
						return nil
					}
					return TapAndFreeze{Targets: []uuid.UUID{id}, Label: "Tamiyo, Compleated Sage — doesn't untap during its controller's next untap step"}.Apply(ctx)
				},
			},
			{
				Label: "−X: Exile target nonland permanent card with mana value X from your graveyard. Create a token that's a copy of that card.",
				Cost:  LoyaltyMinusX(),
				Targets: TargetCardInGraveyard("target nonland permanent card with mana value X from your graveyard",
					Permanent(), Nonland(), YouOwn()).WithManaValueEqualsX(),
				Purpose: game.Purpose{Tokens: 1},
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					id, ok := b16FirstLegalTargetCard(ctx)
					if !ok {
						return nil
					}
					return ExileTarget{
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
				Cost:    LoyaltyCost(-7),
				Purpose: game.Purpose{Tokens: 1},
				Effect:  Do(CreateToken{Template: TamiyosNotebookToken(), N: 1}),
			},
		},
	})
}

// tamiyoCompleatedSageCopy is the −X's second sentence: a token that's a
// copy of the exiled card. An on-disk key: never renamed, never reused.
var tamiyoCompleatedSageCopy = CardTail("then/tamiyo-compleated-sage-copy", func(ctx *Context, o game.TailOutcome) error {
	if len(o.Tail.Carry) == 0 {
		return nil
	}
	return CreateTokenCopy{Controller: o.Tail.Player, Copy: o.Tail.Carry[0], N: 1}.Apply(ctx)
})
