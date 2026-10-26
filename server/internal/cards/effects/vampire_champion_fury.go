package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const (
	vampireChampionFuryHellbent = "vampire-champion-fury/hellbent"
	vampireChampionFuryDamage   = "vampire-champion-fury/sacrifice-divided-damage"
)

// Vampire Champion Fury — Stickers (an Unfinity sticker sheet, CR 123.2; ADR 0208, #4307):
//
//	Name stickers "Vampire", "Champion", "Fury"; three art stickers.
//	{TK}{TK} — Hellbent — This creature gets +3/+3 as long as you have no cards in hand.
//	{TK}{TK}{TK}{TK} — {2}, Sacrifice this creature: It deals X damage divided as you choose among any number of target creatures, where X is its power.
//	{TK}{TK} — 1/5
//	{TK}{TK}{TK} — 6/3
//
// The first sticker gives hellbent (CR 207.2c, an ability word) as a static in a bundle: +3/+3 (layer 7c) as long as its controller has no cards in hand. The bundle's static declares that it depends on hand sizes, so a draw, a discard or a card put into the hand re-reads it. The second gives an ability whose cost sacrifices the creature: it deals X damage, X its power, divided as the controller chooses among any number of target creatures. X is read as the ability is announced (CR 601.2d), while the creature still has the power it will be sacrificed with, and the damage is dealt by the creature as it last existed (CR 609.7).
//
// The name, power and toughness, and art stickers are the rules' (CR 123.6, 123.8, 123.9).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "be19d5ef-15ed-422d-8799-63eabfdbaae3",
		Name:         "Vampire Champion Fury",
		Completeness: CompletenessFull,
		Grants: []AbilityGrant{
			{
				Key: vampireChampionFuryHellbent,
				Static: []game.StaticAbility{{
					Label:    "Hellbent — This creature gets +3/+3 as long as you have no cards in hand.",
					Layer:    game.Layer7PT,
					SubLayer: game.SubLayer7C_Modify,
					AppliesTo: func(target *game.Card, g *game.Game, source *game.Card) bool {
						return source != nil && target.InstanceID == source.InstanceID && stickerHandSize(g, source.Controller) == 0
					},
					Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
						c.Power += 3
						c.Toughness += 3
					},
					DependsOnHandSize: true,
				}},
				Text: "Hellbent — This creature gets +3/+3 as long as you have no cards in your hand.",
			},
			{
				Key: vampireChampionFuryDamage,
				Activated: []ActivatedAbility{{
					Label: "{2}, Sacrifice this creature: It deals X damage divided as you choose among any number of target creatures, where X is its power.",
					Cost:  Plus(ManaCost("{2}"), SacrificeThis()),
					Targets: TargetCreature("any number of target creatures").WithCount(0, 0).
						Dividing(DivideBy(DivideSourcePower)),
					Effect: func(g *game.Game, item *game.StackItem) error {
						return DealDividedDamage(NewContext(g, item))
					},
				}},
				Text: "{2}, Sacrifice this creature: It deals X damage divided as you choose among any number of target creatures, where X is its power.",
			},
		},
		StickerSheet: &StickerSheetSpec{Abilities: []StickerAbilitySpec{
			{Grant: vampireChampionFuryHellbent},
			{Grant: vampireChampionFuryDamage},
		}},
	})
}

// stickerHandSize is the number of cards in `you`'s hand, walking the seats
// directly because a layer pass may hold only the read lock.
func stickerHandSize(g *game.Game, you uuid.UUID) int {
	for _, p := range g.Seats {
		if p != nil && p.ID == you && p.Hand != nil {
			return p.Hand.Size()
		}
	}
	return 0
}
