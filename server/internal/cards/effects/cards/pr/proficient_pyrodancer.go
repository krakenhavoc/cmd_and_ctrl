package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Proficient Pyrodancer — Creature — Human Performer {2}{R}, 2/3:
//
//	"When this creature enters, you may put an art sticker on a nonland
//	 permanent you own.
//	 {2}{R}: Another target creature with an art sticker on it gets +2/+0
//	 and gains menace until end of turn."
//
// The art sticker is the engine's keyed questions with only art stickers
// offered (ADR 0208 §4); it does nothing but mark the object (CR 123.9),
// which the target clause reads (WithSticker).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ddd3f530-1ea4-4b1e-9f9b-5d34fa626529",
		Name:         "Proficient Pyrodancer",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Proficient Pyrodancer — you may put an art sticker on a nonland permanent you own",
				effects.Do(effects.PutASticker{Kinds: []game.StickerKind{game.StickerArt}})),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}{R}: Another target creature with an art sticker on it gets +2/+0 and gains menace until end of turn.",
			Purpose: effects.ForTargets(effects.BenefitUntilEOT(0, 2, 0, "menace")),
			Cost:    effects.ManaCost("{2}{R}"),
			Targets: effects.Another(effects.TargetCreature("another target creature with an art sticker on it", effects.WithSticker(game.StickerArt))),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				id, ok := effects.B16FirstLegalTargetCard(ctx)
				if !ok {
					return nil
				}
				if err := (effects.BoostUntilEOT{Target: id, Power: 2, Label: "Proficient Pyrodancer — +2/+0"}).Apply(ctx); err != nil {
					return err
				}
				return effects.GrantKeywordUntilEOT{Target: id, Keywords: []string{"menace"}, Label: "Proficient Pyrodancer — menace"}.Apply(ctx)
			},
		}},
	})
}
