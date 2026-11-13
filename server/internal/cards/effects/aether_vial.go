package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Aether Vial — Artifact {1} (Darksteel, #5605):
//
//	"At the beginning of your upkeep, you may put a charge counter on
//	 this artifact.
//	 {T}: You may put a creature card with mana value equal to the number
//	 of charge counters on this artifact from your hand onto the
//	 battlefield."
//
// The upkeep counter is a "you may" (CR 603.5). The tap ability reads the
// Vial's charge counters as it resolves, from its last-known information
// if it has left the battlefield in response (CR 608.2h), and offers the
// creature cards in hand whose mana value is exactly that number. The put
// is Copper Gnomes' PutFromHandOntoBattlefield: no cast, and declining is
// a real answer. A card with {X} in its cost counts X as 0 in the hand
// (CR 202.3e).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "fc148e1e-dff0-448e-9f16-625341754356",
		Name:         "Aether Vial",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Optional(AtYourUpkeep("Aether Vial — put a charge counter on it",
				Do(CounterOnThis{Kind: game.CounterCharge, N: 1})),
				"Aether Vial — put a charge counter on it?"),
		},
		Activated: []ActivatedAbility{{
			Label:   "{T}: You may put a creature card with mana value equal to the number of charge counters on this artifact from your hand onto the battlefield.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    TapCost(),
			Effect:  aetherVialPut,
		}},
	})
}

// aetherVialPut is the tap ability's body: a creature card from the hand
// whose mana value equals the Vial's charge counters.
func aetherVialPut(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	vial, ok := ctx.SourcePermanent()
	if !ok {
		return nil
	}
	return PutFromHandOntoBattlefield{
		Match:    And(Creature(), b03ManaValueIs(vial.Counters[game.CounterCharge])),
		Optional: true,
		Label:    "Aether Vial — you may put a creature card with mana value equal to its charge counters from your hand onto the battlefield",
	}.Apply(ctx)
}
