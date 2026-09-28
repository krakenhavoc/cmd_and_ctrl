package effects

// Wall of Roots — Creature — Plant Wall {1}{G}, 0/5 (#1664):
//
//	"Defender
//	 Put a -0/-1 counter on this creature: Add {G}. Activate only once
//	 each turn."
//
// The first printed MANA ability whose cost puts a counter on its
// source — ManaAbilityCost.AddCounter was declared for both ability
// kinds (#789) and only Devoted Druid's non-mana untapper used it.
// It waited on #1664: a -0/-1 counter used to be stored and change
// nothing, so the Wall made {G} every turn for free forever. Now each
// counter takes a point of toughness (game.PTCounterDelta, CR 122.1a)
// and the fifth activation kills it, as in paper.
//
// No {T} in the cost, so it works while the Wall is tapped and on the
// turn it arrives (CR 302.6 is about the {T} symbol). The once-a-turn
// limit is OncePerTurnActivation keyed on this ability's label, the
// same per-object activation tally a CR 602 ability reads, which the
// mana-ability path writes at announce. The auto-tapper never plans it
// (manaCounterCostPlannable refuses an added-counter cost: that is the
// player's toughness to spend, not the planner's), so it is activated
// by hand. CR 614.16: a counter placed as a cost is not placed by an
// effect, so Doubling Season does not make it cost two.
func init() {
	Register(Spec{
		OracleID:        "3a21a6ae-b2f2-4f0c-acfd-5f3e8d63fd2f",
		Name:            "Wall of Roots",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"defender"},
		ManaAbilities: []ManaAbility{{
			Cost:      ManaAbilityCost{AddCounter: AddCounterToThis(wallOfRootsCounter, 1).AddCounter},
			Produced:  "{G}",
			Label:     wallOfRootsLabel,
			Condition: OncePerTurnActivation(wallOfRootsLabel),
		}},
	})
}

const (
	wallOfRootsCounter = "-0/-1"
	wallOfRootsLabel   = "Put a -0/-1 counter on this creature: Add {G}"
)
