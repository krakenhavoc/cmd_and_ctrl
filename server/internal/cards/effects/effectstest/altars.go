package effectstest

// altars_test.go — catalog-level coverage for the S21 sacrifice-ANOTHER
// mana cost. internal/game tests the engine against hand-built
// ManaAbilityShapes; these cover the two seams above it that a
// hand-built shape can't reach:
//
//   - wire.go actually carries Spec.ManaAbilities[i].Cost.SacrificeOther
//     into the ManaAbilityShape the engine reads, and
//   - protocol stamps SacrificeOptions, so the client's picker has
//     something to show.
//
// A wiring bug in either seam leaves the engine tests green and the
// card dead in a real game.

const (
	AshnodsAltarOracle   = "4d18bcba-a346-445e-a182-6cc30b7e066d"
	PhyrexianAltarOracle = "8d02b297-97c4-4379-9862-0a462400f66f"
)
