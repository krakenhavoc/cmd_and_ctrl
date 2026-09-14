package effects

// Repurposing Bay — Artifact {2}{U} (EDHREC rank 3795):
//
//	"{2}, {T}, Sacrifice another artifact: Search your library for an
//	 artifact card with mana value equal to 1 plus the sacrificed
//	 artifact's mana value, put that card onto the battlefield, then
//	 shuffle. Activate only as a sorcery."
//
// Birthing Pod for artifacts, on an artifact — Oswald Fiddlebender's
// ability one mana over, with "another" in the cost. The cost is
// mana, tap and "sacrifice another artifact", paid at announce; the
// search reads the sacrificed artifact's mana value back off the
// event log (b17PermanentSacrificedToPay), accepts only artifact
// cards at exactly that value plus one, and puts the pick onto the
// battlefield untapped. Sorcery speed, as printed.
//
// "Another" is enforced by NAME rather than by instance, because the
// sacrifice clause is declared at init() before any Bay exists. In a
// singleton format that is the same artifact; a token copy of the
// Bay would also be excluded — weaker than printed, never stronger,
// and not worth a caveat: nobody pods away a Repurposing Bay with a
// Repurposing Bay.
//
// Rides the catalog-wide deterministic search pick when only one
// card qualifies — the S22 chooser asks only when there is a choice.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "1787ac2f-762d-4f3a-b7e5-12db6d3d470d",
		Name:         "Repurposing Bay",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:        "{2}, {T}, Sacrifice another artifact: Search your library for an artifact card with mana value equal to 1 plus the sacrificed artifact's mana value, put that card onto the battlefield, then shuffle. Activate only as a sorcery.",
			Cost:         Plus(ManaCost("{2}"), TapCost(), b36SacrificeAnotherArtifact("Repurposing Bay")),
			SorcerySpeed: true,
			Effect:       b36RepurposingBaySearch,
		}},
	})
}
