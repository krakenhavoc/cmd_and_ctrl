package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Volshe Tideturner — Creature — Merfolk Wizard {1}{U}, 1/3 (Dominaria
// United, #5487):
//
//	"{T}: Add {U}. Spend this mana only to cast an instant or sorcery
//	 spell or a kicked spell."
//
// Hydro-Channeler's cast-only mana ability (CR 605.1a) whose restriction is
// an either/or: an instant or sorcery spell, or a spell cast with its
// kicker paid (ManaRestrictKickedSpell, CR 702.33a). The kicker is chosen
// as the spell is announced (CR 601.2b), so the mana can pay for the whole
// of a kicked spell, the kicker's share included. Summoning sickness gates
// the tap (CR 302.6).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "eab16fd1-735c-411f-a23a-0b93131c2eb2",
		Name:         "Volshe Tideturner",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{U}",
			Restrictions: []string{game.ManaRestrictAnyOf(
				[]string{game.ManaRestrictCast, game.ManaRestrictAnyType("Instant", "Sorcery")},
				[]string{game.ManaRestrictCast, game.ManaRestrictKickedSpell},
			)},
			Label: "Add {U}. Spend this mana only to cast an instant or sorcery spell or a kicked spell",
		}},
	})
}
