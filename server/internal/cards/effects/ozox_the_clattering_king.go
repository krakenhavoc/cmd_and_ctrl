package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Ozox, the Clattering King — Legendary Creature — Skeleton Noble {2}{B},
// 3/2 (Foundations Jumpstart):
//
//	"Ozox can't block.
//	 When Ozox dies, create Jumblebones, a legendary 2/1 black Skeleton
//	 creature token with "Jumblebones can't block" and "When Jumblebones
//	 leaves the battlefield, return target card named Ozox, the Clattering
//	 King from your graveyard to your hand.""
//
// A self restriction and a dies trigger that makes the token. Jumblebones
// is a catalog token (token_catalog.go) with its own can't-block static and
// a leaves trigger whose clause is a card NAMED Ozox in YOUR graveyard, as
// Rekindling Phoenix's Elemental names its Phoenix. Ozox is already in the
// graveyard when Jumblebones arrives, so a later Jumblebones death brings
// it back; if Ozox is no longer there the trigger has no target and is
// removed (CR 603.3d).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "04e575d1-f843-4330-972f-e77b264073e0",
		Name:         "Ozox, the Clattering King",
		Completeness: CompletenessFull,
		Static:       []game.StaticAbility{RestrictSelf(game.CantBlock)},
		Triggered: []game.TriggeredAbility{
			WhenThisDies("Ozox, the Clattering King — create Jumblebones",
				Do(CreateToken{Template: s82BnsB51JumblebonesToken(), N: 1})),
		},
	})
}

func s82BnsB51JumblebonesToken() game.Card { return TokenFromCatalog(printedBnsB51JumblebonesToken) }

// printedBnsB51JumblebonesToken is Jumblebones as PRINTED.
func printedBnsB51JumblebonesToken() TokenTemplate {
	return TokenTemplate{
		Slug: "jumblebones",
		Card: game.Card{
			Name:      "Jumblebones",
			TypeLine:  "Token Legendary Creature — Skeleton",
			Power:     2,
			Toughness: 1,
			Colors:    []string{"B"},
		},
		Static: []game.StaticAbility{RestrictSelf(game.CantBlock)},
		Triggered: []game.TriggeredAbility{
			Targeting(WhenThisLeaves("Jumblebones — return target card named Ozox, the Clattering King from your graveyard to your hand",
				ReturnFirstLegalGraveyardTargetToHand),
				TargetCardInGraveyard("target card named Ozox, the Clattering King from your graveyard", YouOwn(),
					func(_ *game.Game, _ uuid.UUID, c game.Card) bool { return c.Name == "Ozox, the Clattering King" })),
		},
		Text: "Jumblebones can't block.\nWhen Jumblebones leaves the battlefield, return target card named Ozox, the Clattering King from your graveyard to your hand.",
	}
}
