package effects

// Temple of the False God — Land:
//
//	"{T}: Add {C}{C}. Activate only if you control five or more
//	 lands."
//
// Rank 81, and the cleanest example of the activation-gate sub-gap:
// one ability, one condition, nothing else. Before #352 the engine had
// no way to express "Activate only if …" on a mana ability, so this
// land either did not exist in the catalog or — much worse — would
// have shipped as an unconditional "{T}: Add {C}{C}", i.e. a strictly
// better Ancient Tomb with no drawback. That is the #259 direction and
// the reason the gate had to land before the card.
//
// CR 602.5a: an activation restriction is checked before any cost is
// paid, so a Temple activated on four lands fails with the land
// untapped and nothing spent. ErrConditionNotMet, not a half-paid
// activation.
//
// The land counts ITSELF among the five — "you control five or more
// lands", and it is one of them. Four other lands plus the Temple is
// the threshold, which is the usual point of confusion about this card
// and is what ControlsAtLeast(5, MatchLand) counts.
//
// The auto-tapper honours the gate: a Temple below the threshold is
// not a source at all, so the planner never proposes tapping it and
// the executor never has to bail out of a plan mid-way. It also
// re-checks at execution time, in case the board moved between
// planning and tapping.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "cfdd5dc6-593e-495a-8cfe-3a56b3c4c7df",
		Name:         "Temple of the False God",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:      ManaAbilityCost{Tap: true},
			Produced:  "{C}{C}",
			Label:     "Add {C}{C} (only with five or more lands)",
			Condition: ControlsAtLeast(5, MatchLand),
		}},
	})
}
