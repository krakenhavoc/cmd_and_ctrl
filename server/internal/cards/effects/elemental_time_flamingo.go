package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

const (
	elementalTimeFlamingoCast  = "elemental-time-flamingo/exile-cast-from-graveyard"
	elementalTimeFlamingoDrain = "elemental-time-flamingo/creature-dies-drain"
)

// Elemental Time Flamingo — Stickers (an Unfinity sticker sheet, CR 123.2; ADR 0208, #4307):
//
//	Name stickers "Elemental", "Time", "Flamingo"; three art stickers.
//	{TK}{TK} — Exile this permanent: You may cast target nonland card from your graveyard this turn.
//	{TK}{TK}{TK}{TK} — Whenever a creature you control dies, each opponent loses 1 life and you gain 1 life.
//	{TK}{TK} — 1/5
//	{TK}{TK}{TK} — 5/4
//
// The first sticker gives an activated ability whose cost exiles the permanent: "You may cast target nonland card from your graveyard this turn" is a per-card permission until end of turn (CastTargetFromYourGraveyardThisTurn), the card's mana cost paid as for any cast. The second gives a trigger for each creature you control that dies, the creature itself included.
//
// The name, power and toughness, and art stickers are the rules' (CR 123.6, 123.8, 123.9).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "5829ac8e-bd1c-4065-b30a-5307ab11ae79",
		Name:         "Elemental Time Flamingo",
		Completeness: CompletenessFull,
		Grants: []AbilityGrant{
			{
				Key: elementalTimeFlamingoCast,
				Activated: []ActivatedAbility{{
					Label:   "Exile this permanent: You may cast target nonland card from your graveyard this turn.",
					Cost:    ExileThis(),
					Targets: TargetCardInGraveyard("target nonland card in your graveyard", YouOwn(), Nonland()),
					Effect: func(g *game.Game, item *game.StackItem) error {
						ctx := NewContext(g, item)
						id, ok := stickerFirstLegalTarget(ctx)
						if !ok {
							return nil
						}
						return CastTargetFromYourGraveyardThisTurn{Target: id, Label: "Elemental Time Flamingo — cast from your graveyard this turn"}.Apply(ctx)
					},
				}},
				Text: "Exile this permanent: You may cast target nonland card from your graveyard this turn.",
			},
			{
				Key: elementalTimeFlamingoDrain,
				Triggered: []game.TriggeredAbility{
					WheneverACreatureYouControlDies("Elemental Time Flamingo sticker — each opponent loses 1 life and you gain 1 life",
						Do(EachOpponentLosesLife{N: 1}, GainLife{Amount: 1})),
				},
				Text: "Whenever a creature you control dies, each opponent loses 1 life and you gain 1 life.",
			},
		},
		StickerSheet: &StickerSheetSpec{Abilities: []StickerAbilitySpec{
			{Grant: elementalTimeFlamingoCast},
			{Grant: elementalTimeFlamingoDrain},
		}},
	})
}
