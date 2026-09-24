package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Pyromancer's Goggles — Legendary Artifact {5} (EDHREC rank 2825):
//
//	"{T}: Add {R}. When that mana is spent to cast a red instant or
//	 sorcery spell, copy that spell and you may choose new targets for
//	 the copy."
//
// The card #1547's spend riders were opened for beside Cavern of Souls.
// The {R} is unrestricted — it pays for anything, and the auto-tapper
// plans the Goggles like any red source — and carries a trigger rider
// whose filter is the printed clause: a cast (not an activation), of a
// RED object that is an instant OR a sorcery. When the token pays for
// such a spell, the trigger goes on the stack above it, so the copy
// resolves first.
//
// "Copy that spell" names the spell without targeting it (CR 115.10),
// so it is read off the trigger's payload and copied from last-known
// information if it was countered in response — CR 608.2h, the storm
// and Doublecast reading (copyTheSpellYouJustCast). The copy is not
// cast (CR 707.10), so it spent no mana: it does not trigger the
// Goggles again, and it is not uncounterable because the original was.
//
// "When THAT mana is spent" is one activation's mana: a Mana Reflection
// that doubles the {R} is still one copy (ManaSpendRider.Production).
//
// One declared simplification, weaker than printed: with strict mana
// off the pool is never spent (ADR 0068 §3), so no token pays and the
// spell is never copied.
func init() {
	Register(Spec{
		OracleID:     "f76bcbfe-483f-4e63-8425-76feca1abf3e",
		Name:         "Pyromancer's Goggles",
		Completeness: CompletenessCaveats,
		Caveats:      []string{"With strict mana off, the game doesn't see which mana you spent, so the spell is never copied."},
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{R}",
			Label:    "Add {R}",
			SpendRiders: []game.ManaSpendRider{WhenManaSpent("Pyromancer's Goggles", game.ManaSpendTrigger{
				Label:  "Pyromancer's Goggles — copy that spell",
				Effect: copyTheSpellYouJustCast,
			}, ManaRestrictCast, ManaRestrictColor("R"), ManaRestrictAnyType("Instant", "Sorcery"))},
		}},
	})
}
