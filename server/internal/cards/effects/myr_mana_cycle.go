package effects

// myr_mana_cycle.go — the Mirrodin Myr mana-creature cycle:
//
//	"{T}: Add {X}."
//
// Artifact Creature — Myr, 1/1, {2}. Fyndhorn Elves' shape on an
// artifact body, which is the whole reason a deck plays one: it is
// an artifact for Krark-Clan Ironworks and a Myr for Myr Retriever.
// Leaden Myr is in the roadmap's batch 16 (#309); the other four
// (Gold, Silver, Iron, Copper) belong in this table when their
// batches reach them — a card that belongs to an existing cycle goes
// in the cycle's table, never in a new file.
//
// A single fixed colour, so activating it adds the mana immediately
// with no colour prompt, under the same CR 302.6 summoning-sickness
// gate the engine applies to every tap ability on a creature.
//
// No simplification.
func init() {
	for _, myr := range []struct{ oracleID, name, color string }{
		{"f62cabf0-df0d-4c4f-a93a-9340967d1775", "Leaden Myr", "B"},
	} {
		Register(Spec{
			OracleID:     myr.oracleID,
			Name:         myr.name,
			Completeness: CompletenessFull,
			ManaAbilities: []ManaAbility{{
				Cost:     ManaAbilityCost{Tap: true},
				Produced: "{" + myr.color + "}",
				Label:    "Add {" + myr.color + "}",
			}},
		})
	}
}
