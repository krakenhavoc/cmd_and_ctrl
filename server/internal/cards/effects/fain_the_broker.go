package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Fain, the Broker — Legendary Creature — Human Warlock {2}{B}, 3/3
// (EDHREC rank 3301):
//
//	"{T}, Sacrifice a creature: Put two +1/+1 counters on target
//	 creature.
//	 {T}, Remove a counter from a creature you control: Create a
//	 Treasure token.
//	 {T}, Sacrifice an artifact: Create a 2/1 white and black Inkling
//	 creature token with flying.
//	 {3}{B}: Untap Fain."
//
// The Strixhaven broker: creatures into counters, artifacts into
// Inklings, and mana into more activations. Three of the four
// abilities are CR 602 activations with the cost vocabulary the
// engine has:
//
//   - The counters: tap plus SacrificeACreature — Fain is a creature
//     and may sacrifice himself, as printed — with the target picked
//     at announce and two +1/+1 counters put on it at resolution
//     (b31CountersOnChosenCreature). The sacrifice is paid at
//     announce, so its dies triggers resolve above the ability.
//   - The Inkling: tap plus b10SacrificeAnArtifact, and the token is
//     a 2/1 white-and-black flier (b31WhiteBlackInklingFlyingToken).
//   - The untap: {3}{B} and Fain untaps if he is still there
//     (b31UntapSelf) — which is what lets him activate twice a turn.
//
// Every {T} on a creature source waits out summoning sickness (CR
// 302.1); the untap does not tap and does not.
//
// DECLARED SIMPLIFICATION, weaker than printed: the Treasure ability
// is not implemented. "Remove a counter from a creature you control"
// is a cost component the engine cannot express — AbilityCost
// carries tap, sacrifice, mana, life, loyalty and crew, and nothing
// that removes a counter (the Walking Ballista gap; Iron Spider and
// Dragon's Hoard carry the same note) — and deferring the removal to
// resolution would let a proliferate in response bank a counter the
// printed card had already spent, the #259 direction. Fain is still
// the sacrifice outlet with the untap, which is what he is played
// for; the ability is whole the day a counter-removal cost lands.
func init() {
	Register(Spec{
		OracleID:     "31b990e3-9bad-4b0a-a9d5-f5b9ed2ad0b0",
		Name:         "Fain, the Broker",
		Completeness: CompletenessCaveats,
		Caveats:      []string{"Removing a counter from a creature to make a Treasure isn't implemented — the other three abilities work as printed."},
		Activated: []ActivatedAbility{
			{
				Label:   "{T}, Sacrifice a creature: Put two +1/+1 counters on target creature.",
				Cost:    Plus(TapCost(), SacrificeACreature()),
				Targets: TargetCreature("target creature"),
				Effect:  b31CountersOnChosenCreature,
			},
			{
				Label: "{T}, Sacrifice an artifact: Create a 2/1 white and black Inkling creature token with flying.",
				Cost:  Plus(TapCost(), b10SacrificeAnArtifact()),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return CreateToken{Controller: item.Controller, Template: b31WhiteBlackInklingFlyingToken(), N: 1}.Apply(NewContext(g, item))
				},
			},
			{
				Label:  "{3}{B}: Untap Fain.",
				Cost:   ManaCost("{3}{B}"),
				Effect: b31UntapSelf,
			},
		},
	})
}
