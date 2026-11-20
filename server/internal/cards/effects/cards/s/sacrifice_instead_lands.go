package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
// The sacrifice is an effect's (SacrificeAllKeyedForEffect): every
// EventSacrifice payoff sees it, and "if you do" means every named
// permanent really left the battlefield. The clause reads effective
// characteristics, so a land an Urborg made a Swamp is a Swamp.
//
// No simplification on any of the seven.

func init() {
	// Heart of Yavimaya — "{T}: Add {G}. {T}: Target creature gets
	// +1/+1 until end of turn."
	effects.Register(effects.Spec{
		OracleID:     "6c9a854c-0509-4ed4-9d94-c45b823b65e5",
		Name:         "Heart of Yavimaya",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{
			effects.EntersOnlyIfYouSacrifice("Heart of Yavimaya", 1, "a Forest", effects.IsLandWithSubtype("forest")),
		},
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{G}",
			Label:    "Add {G}",
		}},
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}: Target creature gets +1/+1 until end of turn.",
			Purpose: effects.ForTargets(effects.BenefitUntilEOT(0, 1, 1)),
			Cost:    effects.TapCost(),
			Targets: effects.TargetCreature("target creature"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				for _, t := range ctx.LegalTargets() {
					if err := (effects.BoostUntilEOT{Target: t.ID, Power: 1, Toughness: 1,
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
	effects.Register(effects.Spec{
		OracleID:     "8b370db5-dfb9-4ea0-9017-bae3e767b041",
		Name:         "Kjeldoran Outpost",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{
			effects.EntersOnlyIfYouSacrifice("Kjeldoran Outpost", 1, "a Plains", effects.IsLandWithSubtype("plains")),
		},
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{W}",
			Label:    "Add {W}",
		}},
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{W}, {T}: Create a 1/1 white Soldier creature token.",
			Purpose: game.Purpose{Answers: game.AnswerMakesBlocker},
			Cost:    effects.Plus(effects.ManaCost("{1}{W}"), effects.TapCost()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.CreateToken{Template: effects.TokenCard("1/1 white Soldier"), N: 1}.Apply(effects.NewContext(g, item))
			},
		}},
	})

	// Lake of the Dead — "{T}: Add {B}. {T}, Sacrifice a Swamp: Add
	// {B}{B}{B}{B}."
	effects.Register(effects.Spec{
		OracleID:     "bdf476e5-1d57-4b17-b45b-d52fd75aadeb",
		Name:         "Lake of the Dead",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{
			effects.EntersOnlyIfYouSacrifice("Lake of the Dead", 1, "a Swamp", effects.IsLandWithSubtype("swamp")),
		},
		ManaAbilities: []effects.ManaAbility{
			{
				Cost:     effects.ManaAbilityCost{Tap: true},
				Produced: "{B}",
				Label:    "Add {B}",
			},
			{
				Cost: effects.ManaAbilityCost{
					Tap:            true,
					SacrificeOther: effects.SacrificeSpec("a Swamp", effects.HasSubtype("Swamp")),
				},
				Produced: "{B}{B}{B}{B}",
				Label:    "Sacrifice a Swamp: Add {B}{B}{B}{B}",
				Answers:  game.AnswerValue,
			},
		},
	})

	// Balduvian Trading Post — "{T}: Add {C}{R}. {1}, {T}: This land
	// deals 1 damage to target attacking creature."
	effects.Register(effects.Spec{
		OracleID:     "7647940e-c99c-401c-ad1d-9ec730f66b6f",
		Name:         "Balduvian Trading Post",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{
			effects.EntersOnlyIfYouSacrifice("Balduvian Trading Post", 1, "an untapped Mountain", effects.UntappedLandWithSubtype("mountain")),
		},
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{C}{R}",
			Label:    "Add {C}{R}",
		}},
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}, {T}: This land deals 1 damage to target attacking creature.",
			Cost:    effects.Plus(effects.ManaCost("{1}"), effects.TapCost()),
			Targets: effects.TargetCreature("target attacking creature", effects.AttackingCreature()),
			Purpose: effects.ForTargets(effects.DamageToTarget(0, 1)),
			Effect:  effects.SourceDealsDamageToEachLegalTarget(1),
		}},
	})

	// Soldevi Excavations — "{T}: Add {C}{U}. {1}, {T}: Scry 1."
	effects.Register(effects.Spec{
		OracleID:     "5baa7abe-5bdf-40ce-9a83-a93b7cae71a3",
		Name:         "Soldevi Excavations",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{
			effects.EntersOnlyIfYouSacrifice("Soldevi Excavations", 1, "an untapped Island", effects.UntappedLandWithSubtype("island")),
		},
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{C}{U}",
			Label:    "Add {C}{U}",
		}},
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}, {T}: Scry 1.",
			Purpose: game.Purpose{Answers: game.AnswerValue, Scry: 1},
			Cost:    effects.Plus(effects.ManaCost("{1}"), effects.TapCost()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.Scry{Player: item.Controller, N: 1}.Apply(effects.NewContext(g, item))
			},
		}},
	})

	// Lotus Vale — "{T}: Add three mana of any one color." One pick
	// minting three tokens of the picked colour (#742).
	effects.Register(effects.Spec{
		OracleID:     "01fc5bb3-ebd7-4ab4-8aef-2ece1e1d9b7c",
		Name:         "Lotus Vale",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{
			effects.EntersOnlyIfYouSacrifice("Lotus Vale", 2, "two untapped lands", effects.UntappedLand),
		},
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: effects.OneColorOfAmount(3),
			Label:    "Add three mana of any one color",
		}},
	})

	// Scorched Ruins — "{T}: Add {C}{C}{C}{C}."
	effects.Register(effects.Spec{
		OracleID:     "6ee68855-c8c5-422b-88da-163c09a96416",
		Name:         "Scorched Ruins",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{
			effects.EntersOnlyIfYouSacrifice("Scorched Ruins", 2, "two untapped lands", effects.UntappedLand),
		},
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{C}{C}{C}{C}",
			Label:    "Add {C}{C}{C}{C}",
		}},
	})
}
