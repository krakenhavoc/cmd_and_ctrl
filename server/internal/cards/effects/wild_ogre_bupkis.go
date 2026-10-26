package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const (
	wildOgreBupkisCounter    = "wild-ogre-bupkis/attacks-counter"
	wildOgreBupkisMetalcraft = "wild-ogre-bupkis/metalcraft-protection"
)

// Wild Ogre Bupkis — Stickers (an Unfinity sticker sheet, CR 123.2; ADR 0208, #4307):
//
//	Name stickers "Wild", "Ogre", "Bupkis"; three art stickers.
//	{TK}{TK} — Whenever this creature attacks, put a +1/+1 counter on it.
//	{TK}{TK}{TK} — Metalcraft — This permanent has protection from noncreature permanents as long as you control three or more artifacts.
//	{TK}{TK} — 5/1
//	{TK}{TK}{TK}{TK} — 7/4
//
// The first sticker gives an attack trigger that puts a +1/+1 counter on the creature. The second gives metalcraft (CR 207.2c, an ability word) as a layer-6 static in a bundle: as long as its controller controls three or more artifacts it has protection from noncreature permanents. The artifacts are counted every recompute, and a permanent entering or leaving the battlefield already invalidates the layers.
//
// The name, power and toughness, and art stickers are the rules' (CR 123.6, 123.8, 123.9).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "578c23d2-225f-4488-be7f-4abe38297bde",
		Name:         "Wild Ogre Bupkis",
		Completeness: CompletenessFull,
		Grants: []AbilityGrant{
			{
				Key: wildOgreBupkisCounter,
				Triggered: []game.TriggeredAbility{
					WheneverThisAttacks("Wild Ogre Bupkis sticker — put a +1/+1 counter on it", thisGetsAPlusOneCounter),
				},
				Text: "Whenever this creature attacks, put a +1/+1 counter on it.",
			},
			{
				Key: wildOgreBupkisMetalcraft,
				Static: []game.StaticAbility{{
					Label: "Metalcraft — This permanent has protection from noncreature permanents as long as you control three or more artifacts.",
					Layer: game.Layer6Ability,
					AppliesTo: func(target *game.Card, g *game.Game, source *game.Card) bool {
						return source != nil && target.InstanceID == source.InstanceID && stickerArtifactsControlledBy(g, source.Controller) >= 3
					},
					Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
						c.Abilities = game.AppendKeywordAbility(c.Abilities, "protection from noncreature permanents")
					},
				}},
				Text: "Metalcraft — This permanent has protection from noncreature permanents as long as you control three or more artifacts.",
			},
		},
		StickerSheet: &StickerSheetSpec{Abilities: []StickerAbilitySpec{
			{Grant: wildOgreBupkisCounter},
			{Grant: wildOgreBupkisMetalcraft},
		}},
	})
}

// thisGetsAPlusOneCounter is "put a +1/+1 counter on it" for the source of
// an attack trigger, nothing if the source is a new object (CR 400.7).
func thisGetsAPlusOneCounter(g *game.Game, item *game.StackItem) error {
	if sourceIsNewObject(g, item) || !onBattlefield(g, item.SourceCardID) {
		return nil
	}
	return AddCounter{Target: item.SourceCardID, Kind: game.CounterPlusOne, N: 1}.Apply(NewContext(g, item))
}

// stickerArtifactsControlledBy counts the artifacts `you` control, walking the
// battlefield directly because a layer pass may hold only the read lock.
func stickerArtifactsControlledBy(g *game.Game, you uuid.UUID) int {
	n := 0
	if g.Battlefield == nil {
		return 0
	}
	for i := range g.Battlefield.Cards {
		if c := &g.Battlefield.Cards[i]; c.Controller == you && c.IsArtifact() {
			n++
		}
	}
	return n
}
