package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// The fixed-count mana-ability users that forced #758. They share one
// TapOthers component but deliberately cover its different printed shapes:
// a source that also pays {T}, a creature source whose {T} excludes itself,
// a count above one, and a source that does not tap itself at all.
func init() {
	Register(Spec{
		OracleID:     "bdeb440d-a714-40e8-9038-f651e6ae45fb",
		Name:         "Springleaf Drum",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost: ManaAbilityCost{
				Tap: true,
				TapOthers: &game.TapOthersCost{
					Count:  1,
					Filter: TargetPermanent("an untapped creature you control", Creature()),
					Label:  "an untapped creature you control",
				},
			},
			Produced: "{W|U|B|R|G}",
			Label:    "Add one mana of any color",
		}},
	})

	Register(Spec{
		OracleID:        "e2f98813-4f39-4135-b05b-407e9eb797f6",
		Name:            "Jaspera Sentinel",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"reach"},
		ManaAbilities: []ManaAbility{{
			Cost: ManaAbilityCost{
				Tap: true,
				TapOthers: &game.TapOthersCost{
					Count:  1,
					Filter: TargetPermanent("an untapped creature you control", Creature()),
					Label:  "an untapped creature you control",
				},
			},
			Produced: "{W|U|B|R|G}",
			Label:    "Add one mana of any color",
		}},
	})

	Register(Spec{
		OracleID:     "0b7f9c71-6f3c-4056-9bed-04b9f2296f3b",
		Name:         "Heritage Druid",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost: ManaAbilityCost{TapOthers: &game.TapOthersCost{
				Count:  3,
				Filter: TargetPermanent("three untapped Elves you control", OfCreatureType("Elf")),
				Label:  "three untapped Elves you control",
			}},
			Produced: "{G}{G}{G}",
			Label:    "Add {G}{G}{G}",
		}},
	})

	Register(Spec{
		OracleID:     "4d9554de-c005-41ab-a941-39328ebff9de",
		Name:         "Relic of Legends",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{
			{
				Cost:     ManaAbilityCost{Tap: true},
				Produced: "{W|U|B|R|G}",
				Label:    "Add one mana of any color",
			},
			{
				Cost: ManaAbilityCost{TapOthers: &game.TapOthersCost{
					Count:  1,
					Filter: TargetPermanent("an untapped legendary creature you control", Creature(), Legendary()),
					Label:  "an untapped legendary creature you control",
				}},
				Produced: "{W|U|B|R|G}",
				Label:    "Add one mana of any color",
			},
		},
	})

	Register(Spec{
		OracleID:     "e6b77545-de5c-4f4a-b7ea-83498fb33ba8",
		Name:         "Holdout Settlement",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{
			{
				Cost:     ManaAbilityCost{Tap: true},
				Produced: "{C}",
				Label:    "Add {C}",
			},
			{
				Cost: ManaAbilityCost{
					Tap: true,
					TapOthers: &game.TapOthersCost{
						Count:  1,
						Filter: TargetPermanent("an untapped creature you control", Creature()),
						Label:  "an untapped creature you control",
					},
				},
				Produced: "{W|U|B|R|G}",
				Label:    "Add one mana of any color",
			},
		},
	})
}
