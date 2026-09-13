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
// Sandbox simplifications, both declared and both weaker:
//
//   - The sacrificed creature's power is its printed power plus its
//     +1/+1 and -1/-1 counters as they were when it left. A bonus
//     from another permanent's static ability (an anthem) is not in
//     it, because the trigger has no last-known characteristic to
//     read for a creature that left as a COST rather than dying to
//     an effect.
//   - The third ability is not implemented: an ability activated from
//     the graveyard has no shape (the Gravecrawler gap), and neither
//     does "sacrifice a Swamp and a Forest" as a two-permanent cost.
//     Jarad has to be recast from the command zone or reanimated the
//     ordinary way.
//
// Engine gap it shares with Tarmogoyf, not the card's: the layer
// cache is invalidated by battlefield motion, counters, taps and
// turn changes, not by a card reaching a graveyard from a hand or a
// library, so a mill shows on Jarad's size at the next recompute.
func init() {
	Register(Spec{
		OracleID:     "87e65e36-9483-49fe-b644-2caca092107f",
		Name:         "Jarad, Golgari Lich Lord",
		Completeness: CompletenessCaveats,
		Caveats: []string{
			"The sacrificed creature's power counts its +1/+1 and -1/-1 counters but not a bonus from another permanent, such as an anthem.",
			"The last ability isn't implemented — Jarad can't be returned from your graveyard by sacrificing a Swamp and a Forest.",
		},
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
				SacrificeOther: sacrificeSpec("another creature", Creature(), b03NotNamed("Jarad, Golgari Lich Lord")),
			}),
			Effect: func(g *game.Game, item *game.StackItem) error {
				fed, ok := b17PermanentSacrificedToPay(g, item)
				if !ok {
					return nil
				}
				power := b17LastKnownPowerOffBattlefield(g, fed)
				if power <= 0 {
					return nil
				}
				return eachOpponentLosesLife(g, item, power)
			},
		}},
	})
}
