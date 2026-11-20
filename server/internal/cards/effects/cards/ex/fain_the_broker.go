package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
// The Strixhaven broker: creatures into counters, counters into
// Treasure, artifacts into Inklings, and mana into more activations.
// All four abilities are CR 602 activations:
//
//   - The counters: tap plus SacrificeACreature — Fain is a creature
//     and may sacrifice himself, as printed — with the target picked
//     at announce and two +1/+1 counters put on it at resolution
//     (b31CountersOnChosenCreature). The sacrifice is paid at
//     announce, so its dies triggers resolve above the ability.
//   - The Treasure: tap plus "Remove a counter from a creature you
//     control" — RemoveCountersFrom with no kind (#625). "A counter"
//     is any kind, so the activator names the creature AND the kind
//     at announce (a +1/+1 counter, a -1/-1 counter, a stun counter),
//     and the counter comes off then, so a proliferate in response
//     cannot bank it. The creature is chosen, not targeted: a hexproof
//     creature of yours still pays, and Fain may pay from himself.
//     Spending the last +1/+1 counter off a printed 0/0 (Hangarback
//     Walker, an X hydra) kills it, as printed: the state check reads
//     Card.LostLastCounter, not the placeholder convention.
//   - The Inkling: tap plus b10SacrificeAnArtifact, and the token is
//     a 2/1 white-and-black flier (b31WhiteBlackInklingFlyingToken).
//   - The untap: {3}{B} and Fain untaps if he is still there
//     (b31UntapSelf) — which is what lets him activate twice a turn.
//
// Every {T} on a creature source waits out summoning sickness (CR
// 302.6); the untap does not tap and does not.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "31b990e3-9bad-4b0a-a9d5-f5b9ed2ad0b0",
		Name:         "Fain, the Broker",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{T}, Sacrifice a creature: Put two +1/+1 counters on target creature.",
				Cost:    effects.Plus(effects.TapCost(), effects.SacrificeACreature()),
				Targets: effects.TargetCreature("target creature"),
				Effect:  effects.B31CountersOnChosenCreature,
			},
			{
				Label:   "{T}, Remove a counter from a creature you control: Create a Treasure token.",
				Purpose: game.Purpose{Answers: game.AnswerValue},
				Cost:    effects.Plus(effects.TapCost(), effects.RemoveCountersFrom("", 1, "a creature you control", effects.Creature())),
				Effect:  effects.Do(effects.CreateToken{Template: effects.TreasureToken(), N: 1}),
			},
			{
				Label:   "{T}, Sacrifice an artifact: Create a 2/1 white and black Inkling creature token with flying.",
				Purpose: game.Purpose{Answers: game.AnswerMakesBlocker},
				Cost:    effects.Plus(effects.TapCost(), effects.B10SacrificeAnArtifact()),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return effects.CreateToken{Controller: item.Controller, Template: effects.TokenCard("2/1 white and black Inkling with flying"), N: 1}.Apply(effects.NewContext(g, item))
				},
			},
			{
				Label:   "{3}{B}: Untap Fain.",
				Purpose: game.Purpose{Answers: game.AnswerValue},
				Cost:    effects.ManaCost("{3}{B}"),
				Effect:  effects.B31UntapSelf,
			},
		},
	})
}
