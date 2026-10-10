package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// sacrifice_instead_lands.go — the seven lands that print
//
//	"If this land would enter, sacrifice <n> <kind> instead. If you do,
//	 put this land onto the battlefield. If you don't, put it into its
//	 owner's graveyard."
//
// ADR 0098 Decision 11 (#1744). The first sentence is
// EntersOnlyIfYouSacrifice (enters_only_if.go): a CR 614.1a replacement
// of the land's own entry, from whatever zone it enters (CR 113.6h — the
// 2004-10-04 rulings on Heart of Yavimaya, Lake of the Dead and
// Balduvian Trading Post say "no matter how it is put onto the
// battlefield"). It is not a "may": the controller chooses which
// permanents, never whether, and with fewer than it needs nothing is
// sacrificed and the land goes to its owner's graveyard without
// entering (the Lotus Vale and Scorched Ruins rulings). A land played
// this way has still used the turn's land drop (CR 116.2a, 305.2).
//
// The sacrifice is an effect's (SacrificeAllThenForEffect): every
// EventSacrifice payoff sees it, and "if you do" means every named
// permanent really left the battlefield. The clause reads effective
// characteristics, so a land an Urborg made a Swamp is a Swamp.
//
// No simplification on any of the seven.

func init() {
	// Heart of Yavimaya — "{T}: Add {G}. {T}: Target creature gets
	// +1/+1 until end of turn."
	Register(Spec{
		OracleID:     "6c9a854c-0509-4ed4-9d94-c45b823b65e5",
		Name:         "Heart of Yavimaya",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{
			EntersOnlyIfYouSacrifice("Heart of Yavimaya", 1, "a Forest", IsLandWithSubtype("forest")),
		},
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{G}",
			Label:    "Add {G}",
		}},
		Activated: []ActivatedAbility{{
			Label:   "{T}: Target creature gets +1/+1 until end of turn.",
			Cost:    TapCost(),
			Targets: TargetCreature("target creature"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				for _, t := range ctx.LegalTargets() {
					if err := (BoostUntilEOT{Target: t.ID, Power: 1, Toughness: 1,
						Label: "Heart of Yavimaya — +1/+1"}).Apply(ctx); err != nil {
						return err
					}
				}
				return nil
			},
		}},
	})

	// Kjeldoran Outpost — "{T}: Add {W}. {1}{W}, {T}: Create a 1/1
	// white Soldier creature token."
	Register(Spec{
		OracleID:     "8b370db5-dfb9-4ea0-9017-bae3e767b041",
		Name:         "Kjeldoran Outpost",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{
			EntersOnlyIfYouSacrifice("Kjeldoran Outpost", 1, "a Plains", IsLandWithSubtype("plains")),
		},
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{W}",
			Label:    "Add {W}",
		}},
		Activated: []ActivatedAbility{{
			Label:   "{1}{W}, {T}: Create a 1/1 white Soldier creature token.",
			Purpose: game.Purpose{Answers: game.AnswerMakesBlocker},
			Cost:    Plus(ManaCost("{1}{W}"), TapCost()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return CreateToken{Template: TokenCard("1/1 white Soldier"), N: 1}.Apply(NewContext(g, item))
			},
		}},
	})

	// Lake of the Dead — "{T}: Add {B}. {T}, Sacrifice a Swamp: Add
	// {B}{B}{B}{B}."
	Register(Spec{
		OracleID:     "bdf476e5-1d57-4b17-b45b-d52fd75aadeb",
		Name:         "Lake of the Dead",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{
			EntersOnlyIfYouSacrifice("Lake of the Dead", 1, "a Swamp", IsLandWithSubtype("swamp")),
		},
		ManaAbilities: []ManaAbility{
			{
				Cost:     ManaAbilityCost{Tap: true},
				Produced: "{B}",
				Label:    "Add {B}",
			},
			{
				Cost: ManaAbilityCost{
					Tap:            true,
					SacrificeOther: sacrificeSpec("a Swamp", HasSubtype("Swamp")),
				},
				Produced: "{B}{B}{B}{B}",
				Label:    "Sacrifice a Swamp: Add {B}{B}{B}{B}",
				Answers:  game.AnswerValue,
			},
		},
	})

	// Balduvian Trading Post — "{T}: Add {C}{R}. {1}, {T}: This land
	// deals 1 damage to target attacking creature."
	Register(Spec{
		OracleID:     "7647940e-c99c-401c-ad1d-9ec730f66b6f",
		Name:         "Balduvian Trading Post",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{
			EntersOnlyIfYouSacrifice("Balduvian Trading Post", 1, "an untapped Mountain", untappedLandWithSubtype("mountain")),
		},
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{C}{R}",
			Label:    "Add {C}{R}",
		}},
		Activated: []ActivatedAbility{{
			Label:   "{1}, {T}: This land deals 1 damage to target attacking creature.",
			Cost:    Plus(ManaCost("{1}"), TapCost()),
			Targets: TargetCreature("target attacking creature", AttackingCreature()),
			Purpose: ForTargets(DamageToTarget(0, 1)),
			Effect:  sourceDealsDamageToEachLegalTarget(1),
		}},
	})

	// Soldevi Excavations — "{T}: Add {C}{U}. {1}, {T}: Scry 1."
	Register(Spec{
		OracleID:     "5baa7abe-5bdf-40ce-9a83-a93b7cae71a3",
		Name:         "Soldevi Excavations",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{
			EntersOnlyIfYouSacrifice("Soldevi Excavations", 1, "an untapped Island", untappedLandWithSubtype("island")),
		},
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{C}{U}",
			Label:    "Add {C}{U}",
		}},
		Activated: []ActivatedAbility{{
			Label:   "{1}, {T}: Scry 1.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    Plus(ManaCost("{1}"), TapCost()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return Scry{Player: item.Controller, N: 1}.Apply(NewContext(g, item))
			},
		}},
	})

	// Lotus Vale — "{T}: Add three mana of any one color." One pick
	// minting three tokens of the picked colour (#742).
	Register(Spec{
		OracleID:     "01fc5bb3-ebd7-4ab4-8aef-2ece1e1d9b7c",
		Name:         "Lotus Vale",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{
			EntersOnlyIfYouSacrifice("Lotus Vale", 2, "two untapped lands", untappedLand),
		},
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: OneColorOfAmount(3),
			Label:    "Add three mana of any one color",
		}},
	})

	// Scorched Ruins — "{T}: Add {C}{C}{C}{C}."
	Register(Spec{
		OracleID:     "6ee68855-c8c5-422b-88da-163c09a96416",
		Name:         "Scorched Ruins",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{
			EntersOnlyIfYouSacrifice("Scorched Ruins", 2, "two untapped lands", untappedLand),
		},
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{C}{C}{C}{C}",
			Label:    "Add {C}{C}{C}{C}",
		}},
	})
}
