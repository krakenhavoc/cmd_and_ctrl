package effects

// Krark-Clan Ironworks — Artifact {4} (EDHREC rank 1329):
//
//	"Sacrifice an artifact: Add {C}{C}."
//
// Ashnod's Altar for artifacts, and the engine behind every "KCI"
// combo deck: a mana ability (no stack, CR 605.3a, activatable any
// number of times while something else resolves) that turns any
// artifact — a Treasure, a Clue, a Myr Retriever, the Ironworks
// itself — into two colourless mana. The sacrifice-another clause is
// the same one the Altar uses, so it opens the same picker and fires
// the same sacrifice event, which is what Crime Novelist (this
// batch) and every other "whenever you sacrifice an artifact" payoff
// watch. The Ironworks is an artifact, so it is a legal thing to
// sacrifice to its own ability, as printed.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "68e1f7e0-a9b3-437f-8086-0c0cb85f2880",
		Name:         "Krark-Clan Ironworks",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost: ManaAbilityCost{
				SacrificeOther: b10SacrificeAnArtifact().SacrificeOther,
			},
			Produced: "{C}{C}",
			Label:    "Sacrifice an artifact: Add {C}{C}",
		}},
	})
}
