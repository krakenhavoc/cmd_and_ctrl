package effectstest

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// granted_mana_spent_readers_test.go — ADR 0109 §11 (#1552), the card
// half: the printed sunburst cards on the keyword, the readers that one
// permanent grants over another's cast (Lux Artillery, Coin of Mastery,
// Kalain, Solar Array), "no mana was spent" (Satoru, Freestrider
// Commando), the spend riders with a duration or a count (Generator
// Servant, Carnelian Orb, Opal Palace) and Tyvar Kell's emblem. The
// engine contract is pinned in game/granted_mana_spent_readers_test.go.

const (
	GmrLuxArtilleryOracle      = "cbd76b22-d04e-48e7-bcab-c7f5466d67d8"
	GmrCoinOfMasteryOracle     = "d78518ee-df79-48d1-b9d5-4f968b441899"
	GmrSatoruOracle            = "7555c429-5f2d-4171-b6b0-8e3c8da7f314"
	GmrOpalPalaceOracle        = "aa6723a2-75da-49f5-a1ba-cbfa82c55301"
	GmrKalainOracle            = "4fca09ac-8134-43d9-a84b-686db5e2bf69"
	GmrGeneratorServantOracle  = "68ac061a-e4a3-46df-9169-62f6a8481584"
	GmrCarnelianOrbOracle      = "651c967c-8f71-4eb5-b22f-545e55ea050e"
	GmrTyvarKellOracle         = "14c8a590-88ec-4982-ad7b-36a219ff6de7"
	GmrFreestriderOracle       = "09117016-5fd1-4590-9867-aad65d3097e3"
	GmrSolarArrayOracle        = "e8e2f273-5e74-4f16-8d49-2e86e9c9f2dc"
	GmrEngineeredExplosives    = "95fd897e-9086-42c4-8e0d-61bb02333c5f"
	GmrPentadPrismOracle       = "72e8d67a-41e6-4074-bcbd-b54ead995237"
	GmrHeliophialOracle        = "6c3be36d-2461-4e38-9ab4-4394cd4996e7"
	GmrInfusedArrowsOracle     = "942216f7-6803-4ab3-9239-463002590f6f"
	GmrSpinalParasiteOracle    = "159427f7-27fe-490d-ac78-fed092952f51"
	GmrArcboundWandererOracle  = "03436524-197a-4941-a2a1-c7c4b71c4709"
	GmrClearwaterGobletOracle  = "ffecd3e8-aa28-41a1-8154-a54ae3ce8016"
	GmrOpalineBracersOracle    = "a4d6cb66-698d-4b76-b922-f799ad9d8805"
	GmrLunarAvengerOracle      = "ab42fb8b-c463-45db-b71d-ed847c9abf1f"
	GmrSawtoothThresherOracle  = "3a8f382e-0bc6-4e79-87e2-9f99d907d234"
	GmrSuncrusherOracle        = "4e81ead0-df8b-4d9d-b1cb-d1d9dc503d4c"
	GmrSolarionOracle          = "4022e540-8fa0-4eee-bcf1-df5046289f2b"
	GmrBatonOfCourageOracle    = "a20bb8b3-b724-4fec-b179-ca9e6adc4718"
	GmrSkyreachMantaOracle     = "58eab745-9afa-4db7-9d8a-f5eb0b9dca6c"
	GmrSuntouchedMyrOracle     = "75d4f61a-8276-43ad-99e6-0d756a597e4b"
	GmrHalfTreasureSourceKinds = game.ManaSourceTreasure | game.ManaSourceArtifact
)

// gmrFloatFrom floats one mana of each colour in `spec`, made by a
// source of `kinds` — a Treasure's mana is artifact mana too.
func GmrFloatFrom(g *game.Game, p *game.Player, spec string, kinds game.ManaSourceKinds) {
	g.WithWriteLock(func() {
		for _, r := range spec {
			p.ManaPool.AddMana(game.ManaToken{Color: string(r), Source: uuid.New(), SourceKinds: kinds})
		}
	})
}

// gmrPoolCount is how many mana of `color` is floating in p's pool.
func GmrPoolCount(p *game.Player, color string) int {
	n := 0
	for _, t := range p.ManaPool.Units() {
		if t.Color == color {
			n++
		}
	}
	return n
}
