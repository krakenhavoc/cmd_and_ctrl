package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Jarad, Golgari Lich Lord — Legendary Creature — Zombie Elf
// {B}{B}{G}{G}, 2/2 (EDHREC rank 1827):
//
//	"Jarad gets +1/+1 for each creature card in your graveyard.
//	 {1}{B}{G}, Sacrifice another creature: Each opponent loses life
//	 equal to the sacrificed creature's power.
//	 Sacrifice a Swamp and a Forest: Return this card from your
//	 graveyard to your hand."
//
// The Golgari finisher: fling Lord of Extinction at the whole table.
// The first ability is a layer 7c modify over the controller's
// creature cards in the graveyard (b11CreatureCardsInGraveyard).
// The second is an activated ability with a mana component and a
// sacrifice of another creature (excluded by name, the Warren
// Soultrader shape — Jarad is legendary, so no second copy is ever
// wrongly refused); the sacrificed creature is read back off the
// event log at resolution, since the stack item does not carry the
// cost it paid (b17PermanentSacrificedToPay), and every opponent
// loses that much life.
//
// The sacrificed creature's power is its last-known information
// (CR 608.2h, departedCreaturePower): counters and an anthem's bonus
// both count.
//
// The third ability is an activated ability that functions from the
// graveyard (#660, Zones) whose cost is two permanents of DIFFERENT
// kinds, which is what SacrificeEach (#2526) added: the clause's
// predicate is "a Swamp or a Forest" and TargetSpec.EachOf makes the
// SET fill both parts, so two Swamps do not pay it and a Swamp Forest
// pays one part but not both. The return is
// returnThisCardFromYourGraveyardToHand, which does nothing when the
// card left the graveyard in response.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "87e65e36-9483-49fe-b644-2caca092107f",
		Name:         "Jarad, Golgari Lich Lord",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{{
			Layer:    game.Layer7PT,
			SubLayer: game.SubLayer7C_Modify,
			AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.InstanceID == source.InstanceID
			},
			Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
				n := b11CreatureCardsInGraveyard(g, source.Controller)
				c.Power += n
				c.Toughness += n
			},
		}},
		Activated: []ActivatedAbility{{
			Label: "{1}{B}{G}, Sacrifice another creature: Each opponent loses life equal to the sacrificed creature's power",
			Cost: Plus(ManaCost("{1}{B}{G}"), game.AbilityCost{
				SacrificeOther: Another(sacrificeSpec("another creature", Creature())),
			}),
			Effect: func(g *game.Game, item *game.StackItem) error {
				fed, ok := b17PermanentSacrificedToPay(g, item)
				if !ok {
					return nil
				}
				power := departedCreaturePower(g, fed)
				if power <= 0 {
					return nil
				}
				return eachOpponentLosesLife(g, item, power)
			},
		}, {
			Label: "Sacrifice a Swamp and a Forest: Return this card from your graveyard to your hand",
			Cost: SacrificeEach("a Swamp and a Forest",
				SacrificeSubtype("a Swamp", "Swamp"),
				SacrificeSubtype("a Forest", "Forest")),
			Zones:  []game.ZoneKind{game.ZoneGraveyard},
			Effect: returnThisCardFromYourGraveyardToHand,
		}},
	})
}
