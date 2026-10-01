package effects

// Spellskite — Artifact Creature — Phyrexian Horror {2}, 0/4:
//
//	"{U/P}: Change a target of target spell or ability to this
//	 creature. ({U/P} can be paid with either {U} or 2 life.)"
//
// The card the pinned retarget was built for (#1743, ADR 0019's
// 2026-10-01 amendment). "Change a target … to this creature" is CR
// 115.7b with the destination fixed: ChangeTargets.ToSource, over the
// same gate a free retarget runs, so each of the card's 2020-08-07
// rulings is the engine's behaviour rather than this file's:
//
//   - the ability may target ANY spell or ability — one Spellskite is
//     not a legal target for, or one with no targets at all — and
//     then no targets are changed. So the clause is a bare
//     TargetSpellOrAbility, with no "single target" predicate;
//   - Spellskite must be a legal target for the slot it takes, judged
//     for the redirected item's controller: the slot's clause, its
//     zones, hexproof, shroud, protection;
//   - a change that would make another target illegal ("another
//     target creature", a clause that already names Spellskite) is
//     not made;
//   - with several instances of the word "target", or several targets
//     under one ("any number of target permanents", Deepglow Skate),
//     Spellskite's controller chooses which ONE target changes, as
//     the ability resolves — a prompt over the current targets that
//     could legally become Spellskite, opened only when there are two
//     or more;
//   - if Spellskite leaves the battlefield before the ability
//     resolves, nothing changes.
//
// The new target becomes a target of the redirected spell or ability
// (CR 115.7), so a ward on Spellskite — or anything watching for it
// becoming a target — triggers, attributed to that item's controller.
//
// {U/P} is a Phyrexian symbol on an activation (#917,
// ActivateAbilityParams.PhyrexianLife), so the ability costs {U} or 2
// life, and the activation menu offers both.
//
// DECLARED CAVEAT: two instances of "target" that chose the SAME
// object (CR 601.2c allows it unless the second says "another") are
// one thing to click on the board, and picking it changes the earlier
// slot. Which slot becomes Spellskite matters when the two clauses do
// different things — Soul's Fire aimed at one creature twice — and
// the board cannot point at an instance of a word, only at an object.
// Weaker than printed, never stronger.
func init() {
	Register(Spec{
		OracleID:     "e0aa6ce0-ca31-433b-ac6c-32b8675cdb71",
		Name:         "Spellskite",
		Completeness: CompletenessCaveats,
		Caveats: []string{
			"If a spell or ability targets the same thing more than once, you can't choose which of those targets changes to Spellskite — the first one does.",
		},
		Activated: []ActivatedAbility{{
			Label:   "{U/P}: Change a target of target spell or ability to this creature.",
			Cost:    ManaCost("{U/P}"),
			Targets: TargetSpellOrAbility("target spell or ability"),
			Effect:  changeATargetToThisCreature("Spellskite — choose the target to change to Spellskite", false),
		}},
	})
}
