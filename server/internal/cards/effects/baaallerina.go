package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Baaallerina — Creature — Sheep Performer {3}{U}, 3/3:
//
//	"Flying
//	 When this creature enters, you may put a name sticker on a nonland
//	 permanent you own.
//	 {2}{U}: Another target creature with a name sticker on it gains
//	 flying until end of turn."
//
// The name sticker is the engine's keyed questions with only name stickers
// offered (ADR 0208 §4); its word goes where its controller says (CR
// 123.6b). The target clause reads the sticker data (WithSticker).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "3bf3e503-0628-4550-84a9-7c1dd0a97457",
		Name:            "Baaallerina",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Baaallerina — you may put a name sticker on a nonland permanent you own",
				Do(PutASticker{Kinds: []game.StickerKind{game.StickerName}})),
		},
		Activated: []ActivatedAbility{{
			Label:   "{2}{U}: Another target creature with a name sticker on it gains flying until end of turn.",
			Cost:    ManaCost("{2}{U}"),
			Targets: Another(TargetCreature("another target creature with a name sticker on it", WithSticker(game.StickerName))),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				id, ok := b16FirstLegalTargetCard(ctx)
				if !ok {
					return nil
				}
				return GrantKeywordUntilEOT{Target: id, Keywords: []string{"flying"}, Label: "Baaallerina — flying"}.Apply(ctx)
			},
		}},
	})
}
