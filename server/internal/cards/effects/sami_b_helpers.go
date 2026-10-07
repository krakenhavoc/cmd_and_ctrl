package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Sami Whammy deck, slice B (#2190) — the cantrip and sac-for-mana
// artifacts share a handful of pieces. They live here, once, so the
// clone gate (coverage/clones.go) stays at its baseline.

// samiAnyColorMana is "Add one mana of any color" with the given cost,
// which is every rock of this family: the pipe slot queues the colour
// pick a Birds of Paradise activation does.
func samiAnyColorMana(cost ManaAbilityCost, label string) ManaAbility {
	return ManaAbility{Cost: cost, Produced: "{W|U|B|R|G}", Label: label}
}

// samiDrawOnETB is "When this artifact enters, draw a card." A trigger,
// so it uses the stack and an opponent gets a window (CR 603.3).
func samiDrawOnETB(name string) game.TriggeredAbility {
	return WhenThisEnters(name+" — draw a card", Do(DrawCards{N: 1}))
}

// samiFoodLifeAbility is the printed Food ability on a nontoken
// artifact: "{2}, {T}, Sacrifice this artifact: You gain 3 life."
func samiFoodLifeAbility() ActivatedAbility {
	return ActivatedAbility{
		Label:  "{2}, {T}, Sacrifice this artifact: You gain 3 life",
		Cost:   Plus(ManaCost("{2}"), TapCost(), SacrificeThis()),
		Effect: b36GainLife(3),
	}
}

// samiDrawRider is a mana ability's "Draw a card." clause (Chromatic
// Sphere). A rider is not a cost: it runs inside the same atomic
// activation, after the mana is added.
func samiDrawRider(g *game.Game, controller, _ uuid.UUID) error {
	return g.DrawNForEffect(controller, 1)
}

// samiAnyColorManaThenDraw is samiAnyColorMana with the draw rider.
func samiAnyColorManaThenDraw(cost ManaAbilityCost, label string) ManaAbility {
	m := samiAnyColorMana(cost, label)
	m.Rider = samiDrawRider
	return m
}
