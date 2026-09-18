package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Quicksilver Amulet — Artifact {4} (EDHREC rank 4047):
//
//	"{4}, {T}: You may put a creature card from your hand onto the
//	 battlefield."
//
// Four mana a turn for any creature, at instant speed, forever. The
// big-creature decks play it because the price stops mattering once
// the creature costs eight — and because the creature is PUT onto the
// battlefield, not cast, which is the whole card.
//
// Everything that follows from "put" rather than "cast":
//
//   - Counterspells do not answer it. There is no spell.
//   - Cast triggers do not fire — an Eldrazi's "when you cast this
//     spell" clause gets nothing, which is exactly why Desolation Twin
//     is a worse Amulet target than its mana value suggests. ETB
//     triggers fire normally.
//   - Additional costs, cost increases and cost reductions are all
//     irrelevant; the Amulet's {4} is the whole price.
//   - It happens at instant speed, so the creature can arrive as a
//     surprise blocker on an opponent's turn.
//
// "You MAY" is a real decline: the activation is worth making with
// nothing worth putting down (to leave mana open, or because the
// ability was activated in response to something), and the prompt
// offers that. A player with no creature card in hand is not prompted
// at all.
//
// The {T} means summoning sickness never applies — the Amulet is not
// a creature — but it does mean one creature per turn cycle unless
// something untaps it.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "e1096a98-a631-43f5-97d8-325001d608b3",
		Name:         "Quicksilver Amulet",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label: "{4}, {T}: You may put a creature card from your hand onto the battlefield.",
			Cost:  Plus(ManaCost("{4}"), TapCost()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return PutFromHandOntoBattlefield{
					Match:    Creature(),
					Optional: true,
					Label:    "Quicksilver Amulet — you may put a creature card from your hand onto the battlefield",
				}.Apply(NewContext(g, item))
			},
		}},
	})
}
