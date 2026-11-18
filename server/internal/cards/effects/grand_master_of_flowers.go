package effects

import (
	"slices"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Grand Master of Flowers — Legendary Planeswalker — Bahamut {2}{W}{W},
// loyalty 3 (Adventures in the Forgotten Realms, S82 batch 15, #5499):
//
//	"As long as Grand Master of Flowers has seven or more loyalty counters
//	 on him, he's a 7/7 Dragon God creature with flying and indestructible.
//	 +1: Target creature without first strike, double strike, or vigilance
//	     can't attack or block until your next turn.
//	 +1: Search your library and/or graveyard for a card named Monk of the
//	     Open Hand, reveal it, and put it into your hand. If you search your
//	     library this way, shuffle."
//
// THE STATIC is Gideon Blackblade's three layered effects on one condition
// read at every layer pass, here "has seven or more loyalty counters": layer
// 4 adds Creature and the Dragon and God subtypes (he stays a planeswalker),
// layer 6 adds flying and indestructible, layer 7b sets 7/7 (CR 613.4b).
// Below seven he is a plain planeswalker again. THE FIRST +1 targets a
// creature that has none of the three keywords (checked as it is targeted and
// as it resolves) and restricts it until your next turn (Kaito, Dancing
// Shadow's duration). THE SECOND +1 is Delivery Moogle's search, mandatory,
// over the library and/or graveyard; the player picks the zones first, and the
// library is shuffled only if it was searched (CR 701.23). It may find
// nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "1f3dbc53-4075-40b8-b70d-ce74d3ff5f6d",
		Name:            "Grand Master of Flowers",
		Completeness:    CompletenessFull,
		StartingLoyalty: 3,
		Static: []game.StaticAbility{
			{
				Label:     "As long as this has seven or more loyalty counters on it, it's a Dragon God creature in addition to its other types.",
				Layer:     game.Layer4Type,
				AppliesTo: grandMasterIsACreature,
				Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, source *game.Card) {
					if !slices.Contains(c.Types, "Creature") {
						c.Types = append(c.Types, "Creature")
					}
					for _, st := range source.TextSubtypes("Dragon", "God") {
						if !slices.Contains(c.Subtypes, st) {
							c.Subtypes = append(c.Subtypes, st)
						}
					}
				},
			},
			Labelled(KeywordGrant(grandMasterIsACreature, "flying"), "As long as this has seven or more loyalty counters on it, it has flying."),
			Labelled(KeywordGrant(grandMasterIsACreature, "indestructible"), "As long as this has seven or more loyalty counters on it, it has indestructible."),
			{
				Label:     "As long as this has seven or more loyalty counters on it, it's a 7/7.",
				Layer:     game.Layer7PT,
				SubLayer:  game.SubLayer7B_Set,
				AppliesTo: grandMasterIsACreature,
				Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
					c.Power, c.Toughness = 7, 7
				},
			},
		},
		Activated: []ActivatedAbility{
			{
				Label: "+1: Target creature without first strike, double strike, or vigilance can't attack or block until your next turn.",
				Cost:  LoyaltyCost(1),
				Targets: TargetCreature("target creature without first strike, double strike, or vigilance",
					WithoutKeyword("first strike"), WithoutKeyword("double strike"), WithoutKeyword("vigilance")),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					t, ok := firstLegalTarget(ctx)
					if !ok {
						return nil
					}
					return ScopedEffectFor{
						Target:   t.ID,
						Mods:     []game.Mod{game.AddRestrictionsMod(game.CantAttackOrBlock)},
						Duration: DurationUntilYourNextTurn(ctx, ctx.Controller()),
						Label:    "Grand Master of Flowers — can't attack or block until your next turn",
					}.Apply(ctx)
				},
			},
			{
				Label: "+1: Search your library and/or graveyard for a card named Monk of the Open Hand, reveal it, and put it into your hand. If you search your library this way, shuffle.",
				Cost:  LoyaltyCost(1),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return SearchLibrary{
						Player:        item.Controller,
						Predicate:     func(c game.Card) bool { return c.Name == "Monk of the Open Hand" },
						Dest:          game.ZoneHand,
						Limit:         1,
						Reveal:        true,
						Shuffle:       true,
						AlsoGraveyard: true,
						Reason:        "Grand Master of Flowers — a card named Monk of the Open Hand",
					}.Apply(NewContext(g, item))
				},
			},
		},
	})
}

// grandMasterIsACreature is the static's condition: this permanent has seven
// or more loyalty counters on it, read at every layer pass.
func grandMasterIsACreature(target *game.Card, _ *game.Game, source *game.Card) bool {
	return target.InstanceID == source.InstanceID && source.Counters[game.CounterLoyalty] >= 7
}
