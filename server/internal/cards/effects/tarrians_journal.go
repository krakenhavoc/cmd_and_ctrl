package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

const tarriansJournalOracleID = "a75b02ba-b0c8-47e3-a05c-e9ba221a7578"

// Tarrian's Journal — Legendary Artifact — Book {1}{B}, the front of a
// transforming card whose back is The Tomb of Aclazotz (The Lost Caverns of
// Ixalan, #4470):
//
//	"{T}, Sacrifice another artifact or creature: Draw a card. Activate
//	 only as a sorcery.
//	 {2}, {T}, Discard your hand: Transform Tarrian's Journal."
//
// The sacrifice and the discard of the whole hand are costs, paid as the
// abilities are activated (CR 602.2b), so an empty hand still pays the
// second.
//
// The Tomb of Aclazotz:
//
//	"{T}: Add {B}.
//	 {T}: You may cast a creature spell from your graveyard this turn. If
//	 you do, it enters with a finality counter on it and is a Vampire in
//	 addition to its other types."
//
// A standing permission over the controller's graveyard for this turn
// (Yawgmoth's Will's shape), narrowed to creature cards, whose riders belong
// to the permission: a creature cast through it enters with a finality counter
// (CR 122.1h exiles it instead of letting it die) and is a Vampire in
// addition to its other types for as long as it is that permanent
// (CastPermission.GainsSubtypes). A creature cast from the graveyard any
// other way is not cast "this way" and gets neither.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     tarriansJournalOracleID,
		Name:         "Tarrian's Journal",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{
			{
				Label:        "{T}, Sacrifice another artifact or creature: Draw a card. Activate only as a sorcery.",
				Cost:         Plus(TapCost(), SacrificeAnotherN(1, "another artifact or creature", Or(Artifact(), Creature()))),
				SorcerySpeed: true,
				Purpose:      game.Purpose{Draws: 1},
				Effect:       Do(DrawCards{N: 1}),
			},
			{
				Label:   "{2}, {T}, Discard your hand: Transform Tarrian's Journal.",
				Cost:    Plus(ManaCost("{2}"), TapCost(), DiscardYourHand()),
				Purpose: game.Purpose{Answers: game.AnswerValue},
				Effect:  Do(TransformThis{}),
			},
		},
	})
	Register(Spec{
		OracleID:     tarriansJournalOracleID + "#1",
		Name:         "The Tomb of Aclazotz",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{B}",
			Label:    "{T}: Add {B}",
		}},
		Activated: []ActivatedAbility{{
			Label:   "{T}: You may cast a creature spell from your graveyard this turn. If you do, it enters with a finality counter on it and is a Vampire in addition to its other types.",
			Cost:    TapCost(),
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Effect: func(g *game.Game, item *game.StackItem) error {
				g.GrantCastPermissionForEffect(game.CastPermission{
					Player:                item.Controller,
					Zone:                  game.ZoneGraveyard,
					Scope:                 game.ScopeStanding,
					CastOnly:              true,
					Filter:                game.PermissionFilter{CreatureOnly: true},
					Source:                item.SourceCardID,
					SourceName:            "The Tomb of Aclazotz",
					EntersWithCounterKind: game.CounterFinality,
					EntersWithCounters:    1,
					GainsSubtypes:         []string{"Vampire"},
					Label:                 "Cast a creature spell from your graveyard (The Tomb of Aclazotz)",
				})
				return nil
			},
		}},
	})
}
