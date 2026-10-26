package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

const (
	trainedBlessedMindExile     = "trained-blessed-mind/tap-exile"
	trainedBlessedMindThreshold = "trained-blessed-mind/threshold"
)

// Trained Blessed Mind — Stickers (an Unfinity sticker sheet, CR 123.2; ADR 0208, #4307):
//
//	Name stickers "Trained", "Blessed", "Mind"; three art stickers.
//	{TK}{TK} — {T}: Exile target card from a graveyard.
//	{TK}{TK}{TK} — Threshold — As long as seven or more cards are in your graveyard, this creature gets +4/+0 and has trample.
//	{TK}{TK} — 4/2
//	{TK}{TK}{TK}{TK}{TK} — 6/9
//
// The first sticker gives a tap ability that exiles a target card from a graveyard. The second gives threshold (CR 207.2c, an ability word) as two statics in one bundle, both read with the permanent the sticker is on as their source: as long as seven or more cards are in its controller's graveyard it gets +4/+0 (layer 7c) and has trample (layer 6). The graveyard's size is read every recompute, and a card crossing a graveyard boundary already invalidates the layers.
//
// The name, power and toughness, and art stickers are the rules' (CR 123.6, 123.8, 123.9).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "7a1ff7f3-0b3e-47fe-9e89-522c04f53869",
		Name:         "Trained Blessed Mind",
		Completeness: CompletenessFull,
		Grants: []AbilityGrant{
			{
				Key: trainedBlessedMindExile,
				Activated: []ActivatedAbility{{
					Label:   "{T}: Exile target card from a graveyard.",
					Cost:    TapCost(),
					Targets: TargetCardInGraveyard("target card in a graveyard"),
					Effect: func(g *game.Game, item *game.StackItem) error {
						ctx := NewContext(g, item)
						id, ok := stickerFirstLegalTarget(ctx)
						if !ok {
							return nil
						}
						return ExileTarget{Target: id}.Apply(ctx)
					},
				}},
				Text: "{T}: Exile target card from a graveyard.",
			},
			{
				Key: trainedBlessedMindThreshold,
				Static: []game.StaticAbility{
					{
						Label:     "Threshold — This creature gets +4/+0 as long as seven or more cards are in your graveyard.",
						Layer:     game.Layer7PT,
						SubLayer:  game.SubLayer7C_Modify,
						AppliesTo: thisWithSevenCardsInYourGraveyard,
						Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
							c.Power += 4
						},
					},
					{
						Label:     "Threshold — This creature has trample as long as seven or more cards are in your graveyard.",
						Layer:     game.Layer6Ability,
						AppliesTo: thisWithSevenCardsInYourGraveyard,
						Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
							c.Abilities = game.AppendKeywordAbility(c.Abilities, "trample")
						},
					},
				},
				Text: "Threshold — As long as seven or more cards are in your graveyard, this creature gets +4/+0 and has trample.",
			},
		},
		StickerSheet: &StickerSheetSpec{Abilities: []StickerAbilitySpec{
			{Grant: trainedBlessedMindExile},
			{Grant: trainedBlessedMindThreshold},
		}},
	})
}

// thisWithSevenCardsInYourGraveyard is threshold for the permanent the
// bundle is on (its source): its controller has seven or more cards in
// their graveyard (CR 207.2c). It walks the seats directly, because a layer
// pass may hold only the read lock.
func thisWithSevenCardsInYourGraveyard(target *game.Card, g *game.Game, source *game.Card) bool {
	if source == nil || target.InstanceID != source.InstanceID {
		return false
	}
	for _, p := range g.Seats {
		if p != nil && p.ID == source.Controller && p.Graveyard != nil {
			return p.Graveyard.Size() >= 7
		}
	}
	return false
}
