package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Tawnos, Solemn Survivor — Legendary Creature — Human Artificer {1}{U},
// 1/3 (The Brothers' War Commander):
//
//	"{2}, {T}: Create a token that's a copy of up to one target artifact
//	 token you control. Mill two cards.
//	 {1}{W}{U}{B}, {T}, Sacrifice two artifact tokens, Exile an artifact
//	 or creature card from your graveyard: Create a token that's a copy of
//	 the exiled card, except it's an artifact in addition to its other
//	 types. Activate only as a sorcery."
//
// The first ability's target is optional ("up to one"), and the mill
// happens whether or not there was a copy to make: with the target gone or
// never chosen it mills at once, otherwise it mills after the token has
// entered (the copy's tail), so a copy that asks a question first still
// gets its mill (CR 608.2c).
//
// The second pays everything as costs (CR 602.2b): the two artifact tokens
// are sacrificed and the card is exiled from the graveyard as the ability
// is activated, so the copy is of the exiled card, read where it now sits
// in exile (CR 707.2) with the artifact type added (CR 707.9b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "871fbbd6-76a0-4f79-b89f-afee6ebe5a59",
		Name:         "Tawnos, Solemn Survivor",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{
			{
				Label:   "{2}, {T}: Create a token that's a copy of up to one target artifact token you control. Mill two cards.",
				Purpose: game.Purpose{Answers: game.AnswerValue, Tokens: 1},
				Cost:    Plus(ManaCost("{2}"), TapCost()),
				Targets: TargetPermanent("up to one target artifact token you control",
					Artifact(), IsTokenPredicate(), YouControl()).WithCount(0, 1),
				Effect: tawnosSolemnCopyAndMill,
			},
			{
				Label: "{1}{W}{U}{B}, {T}, Sacrifice two artifact tokens, Exile an artifact or creature card from your graveyard: Create a token that's a copy of the exiled card, except it's an artifact in addition to its other types. Activate only as a sorcery.",
				Cost: Plus(ManaCost("{1}{W}{U}{B}"), TapCost(),
					SacrificeN(2, "two artifact tokens", Artifact(), IsTokenPredicate()),
					ExileFromGraveyard(1, "an artifact or creature card", func(c game.Card) bool {
						return c.IsArtifact() || c.IsCreature()
					})),
				SorcerySpeed: true,
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					exiled := ctx.Exiled()
					if len(exiled) == 0 {
						return nil
					}
					return CreateTokenCopy{
						Controller: item.Controller,
						Copy:       exiled[0],
						N:          1,
						Except:     tokenCopyIsAlsoAnArtifact,
					}.Apply(ctx)
				},
			},
		},
	})
}

// tawnosSolemnCopyAndMill is the first ability's resolution.
func tawnosSolemnCopyAndMill(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	id := FirstLegalBattlefieldTarget(ctx)
	if id == uuid.Nil {
		return MillCards{N: 2}.Apply(ctx)
	}
	return CreateTokenCopiesTogether{
		Controller: item.Controller,
		Copies:     []uuid.UUID{id},
		Then:       ctx.ItemTail(tawnosSolemnMill),
	}.Apply(ctx)
}

// tawnosSolemnMill is "Mill two cards." after the copy. An on-disk key.
var tawnosSolemnMill = CardTail("tawnos-solemn-survivor/mill-two", func(ctx *Context, _ game.TailOutcome) error {
	return MillCards{N: 2}.Apply(ctx)
})
