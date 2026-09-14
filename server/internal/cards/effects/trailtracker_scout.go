package effects

// Trailtracker Scout — Creature — Raccoon Scout {1}{G}, 1/3 (EDHREC
// rank 3396):
//
//	"{T}: Add one mana of any color.
//	 Whenever you expend 8, return up to one target permanent card
//	 from your graveyard to your hand. (You expend 8 as you spend
//	 your eighth total mana to cast spells during a turn.)"
//
// A two-mana rainbow dork with a late-game regrowth attached. The
// mana ability is the Birds of Paradise shape — any colour, the
// printed width, so the pipe is not narrowed to the commander's
// identity — and summoning sickness applies, as the engine enforces
// for a creature's tap ability.
//
// Sandbox simplification, declared (the Magda posture: one whole
// ability omitted): the expend trigger is not implemented. "Expend
// 8" is a running count of mana spent on spells within a turn, and
// the engine keeps no such tally — the mana-spent event only fires
// under strict-mode casts, so a permissive-mode cast paid on paper
// would never count and the trigger could not be trusted to fire at
// all. The Scout is a mana dork, which is what the card is played
// for; weaker than printed, never stronger.
func init() {
	Register(Spec{
		OracleID:     "d38646dc-d1ea-473c-aaf1-df8ba327459f",
		Name:         "Trailtracker Scout",
		Completeness: CompletenessCaveats,
		Caveats:      []string{"The expend trigger isn't implemented — spending your eighth mana in a turn doesn't return a permanent card from your graveyard."},
		ManaAbilities: []ManaAbility{{
			Cost:                    ManaAbilityCost{Tap: true},
			Produced:                "{W|U|B|R|G}",
			Label:                   "Add one mana of any color",
			IgnoreCommanderIdentity: true,
		}},
	})
}
