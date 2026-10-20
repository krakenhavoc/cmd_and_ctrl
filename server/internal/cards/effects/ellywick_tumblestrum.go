package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Ellywick Tumblestrum — Legendary Planeswalker — Ellywick {2}{G}{G},
// loyalty 4:
//
//	"+1: Venture into the dungeon.
//	 −2: Look at the top six cards of your library. You may reveal a
//	 creature card from among them and put it into your hand. If it's
//	 legendary, you gain 3 life. Put the rest on the bottom of your
//	 library in a random order.
//	 −7: You get an emblem with "Creatures you control have trample and
//	 haste and get +2/+2 for each differently named dungeon you've
//	 completed.""
//
// The −2 is Duskwatch Recruiter's take (TakeFromLibraryToHand) with the
// legendary rider and the random bottom in its Then. The emblem counts
// the distinct dungeons in Player.CompletedDungeons: a dungeon completed
// twice counts once. The layer pass bumps on EventDungeonCompleted, so
// the bonus grows the moment another dungeon is completed.
//
// No simplification.
func init() {
	const name = "Ellywick Tumblestrum"
	Register(Spec{
		OracleID:        "d0375311-b0ea-4870-bc2c-4894da088dc3",
		Name:            name,
		Completeness:    CompletenessFull,
		StartingLoyalty: 4,
		Emblem: &EmblemSpec{
			Label: "Ellywick Tumblestrum emblem",
			Text:  "Creatures you control have trample and haste and get +2/+2 for each differently named dungeon you've completed.",
			Static: []game.StaticAbility{
				KeywordGrant(ellywickYourCreatures, "trample"),
				KeywordGrant(ellywickYourCreatures, "haste"),
				{
					Label:     "Creatures you control get +2/+2 for each differently named dungeon you've completed.",
					Layer:     game.Layer7PT,
					SubLayer:  game.SubLayer7C_Modify,
					AppliesTo: ellywickYourCreatures,
					Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
						n := 2 * differentlyNamedDungeonsCompleted(g, source.Controller)
						c.Power += n
						c.Toughness += n
					},
				},
			},
		},
		Activated: []ActivatedAbility{
			{
				Label:  "+1: Venture into the dungeon.",
				Cost:   LoyaltyCost(1),
				Effect: Venture,
			},
			{
				Label:  "−2: Look at the top six cards of your library. You may reveal a creature card from among them and put it into your hand. If it's legendary, you gain 3 life. Put the rest on the bottom of your library in a random order.",
				Cost:   LoyaltyCost(-2),
				Effect: ellywickLook,
			},
			{
				Label:  "−7: You get an emblem with \"Creatures you control have trample and haste and get +2/+2 for each differently named dungeon you've completed.\"",
				Cost:   LoyaltyCost(-7),
				Effect: Do(CreateEmblem{}),
			},
		},
	})
}

// ellywickYourCreatures is "creatures you control", from the emblem
// (whose controller is its owner, CR 114.2).
func ellywickYourCreatures(target *game.Card, _ *game.Game, source *game.Card) bool {
	return target.IsCreature() && target.Controller == source.Controller
}

// differentlyNamedDungeonsCompleted counts the distinct dungeons
// `player` has completed.
func differentlyNamedDungeonsCompleted(g *game.Game, player uuid.UUID) int {
	p := g.PlayerByIDForEffect(player)
	if p == nil {
		return 0
	}
	seen := map[string]bool{}
	for _, k := range p.CompletedDungeons {
		seen[k] = true
	}
	return len(seen)
}

// ellywickLook is the −2.
func ellywickLook(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	you := item.Controller
	return TakeFromLibraryToHand{
		Player:   you,
		Cards:    g.LookAtTopOfLibraryForEffect(you, 6),
		Match:    Creature(),
		Max:      1,
		Optional: true,
		Reveal:   true,
		Label:    "Ellywick Tumblestrum — you may reveal a creature card from among them and put it into your hand",
		Then: func(g *game.Game, res TakeFromLibraryResult) error {
			for _, id := range res.Taken {
				if c, ok := g.LookupCardForEffect(id); ok && Legendary()(g, res.Player, c) {
					if err := g.ChangePlayerLifeForEffect(res.Source, res.Player, 3); err != nil {
						return err
					}
				}
			}
			return TakeRestOnBottomInRandomOrder(g, res)
		},
	}.Apply(ctx)
}
