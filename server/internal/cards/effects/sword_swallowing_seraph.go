package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Sword-Swallowing Seraph — Creature — Angel Performer {4}{W}{W}, 4/4:
//
//	"Flying, vigilance
//	 When this creature enters, you may put a name sticker on a nonland
//	 permanent you own.
//	 {1}{W}, {T}: Put a +1/+1 counter on another target creature with a
//	 name sticker on it. Activate only as a sorcery."
//
// The name sticker is the engine's keyed questions with only name stickers
// offered (ADR 0208 §4). The target clause reads the sticker data
// (WithSticker).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "01efde9a-a20d-4972-b17f-c2d9eb024b0b",
		Name:            "Sword-Swallowing Seraph",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "vigilance"},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Sword-Swallowing Seraph — you may put a name sticker on a nonland permanent you own",
				Do(PutASticker{Kinds: []game.StickerKind{game.StickerName}})),
		},
		Activated: []ActivatedAbility{{
			Label:        "{1}{W}, {T}: Put a +1/+1 counter on another target creature with a name sticker on it. Activate only as a sorcery.",
			Cost:         Plus(ManaCost("{1}{W}"), TapCost()),
			SorcerySpeed: true,
			Targets:      Another(TargetCreature("another target creature with a name sticker on it", WithSticker(game.StickerName))),
			// A +1/+1 counter on the announced creature, if still legal:
			// Animal Sanctuary's body.
			Effect: b36CounterOnChosenAnimal,
		}},
	})
}
