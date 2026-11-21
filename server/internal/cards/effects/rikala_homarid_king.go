package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Rikala, Homarid King — {1}{G}{U} Legendary Creature — Lobster Noble 0/4
// (Mystery Booster Commander, S82 GAP batch 3, #5948):
//
//	"At the beginning of your upkeep, put a tide counter on Rikala. Then
//	 if there are four or more tide counters on Rikala, remove them.
//	 Creatures you control get +1/+0 for each tide counter on Rikala.
//	 Whenever you tap an Island for mana, if there are three or more tide
//	 counters on Rikala, add an additional {U}."
//
// The upkeep trigger is one ability whose resolution puts the counter on
// and then, reading the count as it stands (a counter replacement is
// respected), removes every tide counter at four or more. The anthem is
// a layer-7c effect read off the count on every recompute and reaches
// Rikala herself. The extra {U} is a triggered mana ability (CR 605.1b,
// Spec.ManaTriggers): it resolves with the Island's mana, and its
// intervening "if" is checked when the Island is tapped (CR 603.4); the
// Island type is a word of the source's text (WheneverYouTapALandOfSubtypeForMana).
//
// No simplification.
func init() {
	const tide = "tide"
	atLeastThree := func(source *game.Card) bool { return source.Counters[tide] >= 3 }
	manaTrigger := WheneverYouTapALandOfSubtypeForMana(
		"Rikala, Homarid King — add an additional {U}", "Island", AddsFixedMana("{U}"))
	island := manaTrigger.AppliesTo
	manaTrigger.AppliesTo = func(prod game.ManaProduced, source *game.Card, g *game.Game) bool {
		return atLeastThree(source) && island(prod, source, g)
	}
	Register(Spec{
		OracleID:     "45284633-8aab-4078-ae5c-7b0d75efe2f5",
		Name:         "Rikala, Homarid King",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			AtYourUpkeep("Rikala, Homarid King — put a tide counter on it, then remove them if there are four or more",
				func(g *game.Game, item *game.StackItem) error {
					if err := putACounterOnThis(tide)(g, item); err != nil {
						return err
					}
					c, ok := g.LookupCardForEffect(item.SourceCardID)
					if !ok || c.Counters[tide] < 4 {
						return nil
					}
					return removeAllCountersFromThis(tide)(g, item)
				}),
		},
		Static: []game.StaticAbility{{
			Label:    "Creatures you control get +1/+0 for each tide counter on Rikala.",
			Layer:    game.Layer7PT,
			SubLayer: game.SubLayer7C_Modify,
			AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.IsCreature() && target.Controller == source.Controller
			},
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, source *game.Card) {
				c.Power += source.Counters[tide]
			},
		}},
		ManaTriggers: []game.ManaTrigger{manaTrigger},
	})
}
