package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Staff of Compleation — Artifact {3}:
//
//	"{T}, Pay 1 life: Destroy target permanent you own.
//	 {T}, Pay 2 life: Add one mana of any color.
//	 {T}, Pay 3 life: Proliferate.
//	 {T}, Pay 4 life: Draw a card.
//	 {5}: Untap this artifact."
//
// Five abilities, all sharing the one tap symbol, plus the {5} that
// buys a second activation the same turn — the same shape as Fain,
// the Broker's "{3}{B}: Untap Fain" (b31UntapSelf, reused verbatim
// below rather than duplicated).
//
//   - The destroy targets a permanent YOU OWN, not one you control —
//     the printed "own" is the whole reason to run this over ordinary
//     removal (it can take back a stolen permanent, or sacrifice one
//     of your own to fizzle a targeted removal spell aimed at it).
//     YouOwn() is the predicate that reads Card.Owner rather than
//     Card.Controller.
//   - "Add one mana of any color" is a true mana ability (CR 605.3):
//     it does not use the stack, so it goes in ManaAbilities with the
//     life as a COST component (ManaAbilityCost.Life), the same shape
//     Mana Confluence uses — the activation is rejected outright, and
//     the artifact stays tapped, if the life can't be paid (CR 119.4).
//   - The other three are ordinary CR 602 activated abilities: they
//     use the stack, so a Stifle can answer the destroy or the draw
//     in response.
//
// Proliferate asks the controller what to proliferate (#2525,
// proliferate.go): the "choose any number of permanents and/or
// players" is a real prompt with the beneficial pick suggested. It
// used to be picked FOR the controller, which is why this card
// carried a caveat until that issue closed.
func init() {
	Register(Spec{
		OracleID:     "11c7662f-e688-40a6-98fd-6ae89d231b44",
		Name:         "Staff of Compleation",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true, Life: 2},
			Produced: "{W|U|B|R|G}",
			Label:    "Pay 2 life: Add one mana of any color",
		}},
		Activated: []ActivatedAbility{
			{
				Label:   "{T}, Pay 1 life: Destroy target permanent you own.",
				Cost:    Plus(TapCost(), PayLife(1)),
				Targets: TargetPermanent("target permanent you own", YouOwn()),
				Effect:  destroyFirstLegalCardTarget,
			},
			{
				Label: "{T}, Pay 3 life: Proliferate.",
				Cost:  Plus(TapCost(), PayLife(3)),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return Proliferate{}.Apply(NewContext(g, item))
				},
			},
			{
				Label:  "{T}, Pay 4 life: Draw a card.",
				Cost:   Plus(TapCost(), PayLife(4)),
				Effect: Do(DrawCards{N: 1}),
			},
			{
				Label:  "{5}: Untap this artifact.",
				Cost:   ManaCost("{5}"),
				Effect: b31UntapSelf,
			},
		},
	})
}
