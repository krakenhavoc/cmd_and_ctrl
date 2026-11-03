package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Great Hall of the Biblioplex — Land (Secrets of Strixhaven promo, #4834):
//
//	"{T}: Add {C}.
//	 {T}, Pay 1 life: Add one mana of any color. Spend this mana only to
//	 cast an instant or sorcery spell.
//	 {5}: If this land isn't a creature, it becomes a 2/4 Wizard creature
//	 with 'Whenever you cast an instant or sorcery spell, this creature
//	 gets +1/+0 until end of turn.' It's still a land."
//
// Two mana abilities: the colourless tap, and the life-paying one whose
// mana is restricted to casting instants and sorceries (Ore-Rich
// Stalactite's restriction, Boseiju's life cost).
//
// The animation is Raging Ravine's shape, with one difference the text
// makes: it has no duration, so it is permanent (CR 611.2a). It is one
// effect at one timestamp, pinned to this object (CR 611.2c): the Creature
// type and Wizard subtype (layer 4, "it's still a land"), the quoted
// trigger as a granted bundle (layer 6, ADR 0093) and base 2/4 (layer 7b),
// indefinitely (IndefiniteDuration pinned to the land). The land then stays
// a creature, so the activation does nothing the second time ("if this land
// isn't a creature", read as the ability resolves). Its tap-for-mana
// abilities can't be activated while it is a creature that has not been
// under its controller's control since the turn began (CR 302.6, the
// ruling), which is the engine's.
//
// The trigger belongs to the land (ADR 0093 decision 4): "this creature"
// is the host, and it fires on its controller's casts.
//
// No simplification.
const greatHallOfTheBiblioplexGrant = "great-hall-of-the-biblioplex/cast-pump"

func init() {
	Register(Spec{
		OracleID:     "a8c70dab-1e27-4a9c-bd2d-910d5720d02d",
		Name:         "Great Hall of the Biblioplex",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{
			{
				Cost:     ManaAbilityCost{Tap: true},
				Produced: "{C}",
				Label:    "Add {C}",
			},
			{
				Cost:         ManaAbilityCost{Tap: true, Life: 1},
				Produced:     "{W|U|B|R|G}",
				Label:        "{T}, Pay 1 life: Add one mana of any color. Spend this mana only to cast an instant or sorcery spell",
				Restrictions: []string{ManaRestrictCast, game.ManaRestrictAnyType("Instant", "Sorcery")},
			},
		},
		Grants: []AbilityGrant{{
			Key: greatHallOfTheBiblioplexGrant,
			Triggered: []game.TriggeredAbility{
				WheneverYouCast(Or(Instant(), Sorcery()),
					"Great Hall of the Biblioplex — this creature gets +1/+0 until end of turn",
					func(g *game.Game, item *game.StackItem) error {
						return BoostUntilEOT{
							Target: item.SourceCardID,
							Power:  1,
							Label:  "Great Hall of the Biblioplex — +1/+0 until end of turn",
						}.Apply(NewContext(g, item))
					}),
			},
			Text: "Whenever you cast an instant or sorcery spell, this creature gets +1/+0 until end of turn.",
		}},
		Activated: []ActivatedAbility{{
			Label:   "{5}: If this land isn't a creature, it becomes a 2/4 Wizard creature with \"Whenever you cast an instant or sorcery spell, this creature gets +1/+0 until end of turn.\" It's still a land.",
			Purpose: game.Purpose{Answers: game.AnswerAnimate},
			Cost:    ManaCost("{5}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				if !sourceIsStillThisPermanent(g, item) {
					return nil
				}
				if land, ok := g.LookupCardForEffect(item.SourceCardID); !ok || land.IsCreature() {
					return nil
				}
				return GrantAbilitiesFor{
					Target: item.SourceCardID,
					Keys:   []string{greatHallOfTheBiblioplexGrant},
					Also: []game.Mod{
						game.AddTypesMod("Creature"),
						game.AddSubtypesMod(g.TextReaderForEffect(item).TextSubtypes("Wizard")...),
						game.SetBasePowerMod(2),
						game.SetBaseToughnessMod(4),
					},
					Duration: g.PinnedTo(game.IndefiniteDuration(), item.SourceCardID),
					Label:    "Great Hall of the Biblioplex — a 2/4 Wizard creature",
				}.Apply(NewContext(g, item))
			},
		}},
	})
}
