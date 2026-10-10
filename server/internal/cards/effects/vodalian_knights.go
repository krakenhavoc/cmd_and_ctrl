package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Vodalian Knights — Creature — Merfolk Knight {1}{U}{U}, 2/2:
//
//	"First strike
//	 This creature can't attack unless defending player controls an
//	 Island.
//	 When you control no Islands, sacrifice this creature.
//	 {U}: This creature gains flying until end of turn."
//
// ADR 0107 (#1858, #1879). The attack restriction is §2's (CR 508.1c),
// with the defending player worked out per target (CR 508.5, 508.5a), so
// in Commander it may attack only the opponents who control an Island.
// The sacrifice is §1's CR 603.8 state trigger over its own controller's
// Islands: it triggers as soon as the last one is gone and not again while
// it waits or is on the stack. Both halves read one PermanentQuery.
//
// First strike is the printed keyword; the {U} row grants flying until end
// of turn.
//
// No simplification.
func init() {
	q := QuerySubtype("Island")
	Register(Spec{
		OracleID:        "f6daa28f-e5ce-440c-8dc2-b36f59ae0d4f",
		Name:            "Vodalian Knights",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"first strike"},
		Static: []game.StaticAbility{
			CantAttackUnlessDefendingPlayerControls(q),
		},
		Triggered: []game.TriggeredAbility{
			WhenYouControlNo(q, "Vodalian Knights — sacrifice it", SacrificeThisIfStillOnBattlefield),
		},
		Activated: []ActivatedAbility{{
			Label:   "{U}: This creature gains flying until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerCombatGrant},
			Cost:    ManaCost("{U}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return GrantKeywordUntilEOT{Target: item.SourceCardID, Keywords: []string{"flying"},
					Label: "Vodalian Knights — flying"}.Apply(NewContext(g, item))
			},
		}},
	})
}
