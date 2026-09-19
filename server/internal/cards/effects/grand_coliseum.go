package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Grand Coliseum — Land (EDHREC rank 4228):
//
//	"This land enters tapped.
//	 {T}: Add {C}.
//	 {T}: Add one mana of any color. This land deals 1 damage to you."
//
// Tarnished Citadel's better-behaved cousin and the five-colour deck's
// budget fixer: a tapped entry buys you the choice of paying no life
// at all when colourless will do.
//
// The two abilities are declared separately because they really are
// two abilities, and the difference is visible in play: tapping for
// {C} costs nothing, tapping for a colour costs a point. A single
// ability with the damage folded in would tax the colourless mode
// too, which is stronger for the opponent and weaker for the
// controller — still the wrong direction, because it is not the
// printed card.
//
// The pain is a RIDER, not a trigger, and that is the difference
// between this card and City of Brass. Grand Coliseum says "{T}: Add
// one mana of any color. This land deals 1 damage to you" — the
// damage is part of that ability's effect, so it happens only when
// that ability is activated. City of Brass says "Whenever this land
// becomes tapped", which is a real CR 603 trigger that fires on an
// opponent's Icy Manipulator too. Writing either one as the other
// would ship the wrong card.
//
// "Any color" offers all five, with the controller's commander
// identity listed first (owner decision 2026-09-17) — the printed
// text names no identity, so NarrowToCommanderIdentity stays off.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "1cea9b82-d2e9-4758-8ec8-729fcf4bb7d7",
		Name:         "Grand Coliseum",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{SelfEntersTapped()},
		ManaAbilities: []ManaAbility{
			{
				Cost:     ManaAbilityCost{Tap: true},
				Produced: "{C}",
				Label:    "Add {C}",
			},
			{
				Cost:     ManaAbilityCost{Tap: true},
				Produced: "{W|U|B|R|G}",
				Label:    "Add one mana of any color. This land deals 1 damage to you",
				Rider:    PainRider(1),
			},
		},
	})
}
