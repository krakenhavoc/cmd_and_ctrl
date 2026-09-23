package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Boom Scholar — Creature — Goblin Advisor {1}{R}{G}, 3/3:
//
//	"Exhaust abilities of other permanents you control cost {2} less
//	 to activate.
//	 Exhaust — {4}{R}{G}: Creatures and Vehicles you control gain
//	 trample until end of turn. Put two +1/+1 counters on this
//	 creature. (Activate each exhaust ability only once.)"
//
// The cheapest of the three seams #1184 opened, and the one that
// needed the least new machinery: the CR 601.2f cost pass already
// existed, with its ordering, its generic floor and its
// negative-amount refusal. What it could not do was look at the
// ABILITY — `CostQuery` carried the SOURCE permanent, so "of other
// permanents you control" was already expressible, and nothing at all
// said which of that permanent's abilities was being priced or
// whether it printed the keyword.
//
// Two fields closed it. `CostQuery.Ability` is the ability, and
// `CostModifier.Activations` is the partition: a modifier prices
// casts or activations and never both. The partition is not
// bookkeeping — Sphere of Resistance's AppliesTo is nil, meaning
// "every spell", and without it every {T} ability in the game would
// have started costing {1} more the day the field appeared.
//
// "OTHER permanents you control" is the half that matters on the
// board: the Goblin prints an exhaust ability of its own, and
// discounting that too would be a cheaper card than the one in the
// pack. The clause also reaches exhaust MANA abilities (CR 605.1a),
// which the engine's activation pricing does not run over yet — see
// the caveat.
//
// The trample grant pins its affected set at resolution (CR 611.2c),
// so a creature that arrives after the ability resolves does not
// gain it, and a Vehicle counts whether or not it is currently
// animated — "creatures AND Vehicles" is the printed answer to
// exactly that question.
//
// BOTH DECLARED CAVEATS ARE CLEARED. #1207 fixed their root causes —
// ActivatedAbilityView / ManaAbilityView now stamp ChargedManaCost
// alongside the printed cost, and Game.ManaAbilityManaCostForEffect
// runs the same CR 601.2f pass over a CR 605 mana ability's own cost
// — without this file changing at all: the CostModifier below already
// read q.Ability.Exhaust and "another permanent you control", and
// once CostQuery.Ability.Mana stopped being permanently false, both
// clauses simply started reaching a mana ability too. See
// boom_scholar_test.go for the closing proof.
func init() {
	Register(Spec{
		OracleID:     "48296cc0-0141-47c6-9ec8-ada6171fee6f",
		Name:         "Boom Scholar",
		Completeness: CompletenessFull,
		CostModifiers: []game.CostModifier{
			ActivationCostsLess(2,
				"Exhaust abilities of other permanents you control cost {2} less to activate.",
				AnExhaustAbilityCost(), OfAnotherPermanentYouControl()),
		},
		Activated: []ActivatedAbility{{
			Label: "Exhaust — {4}{R}{G}: Creatures and Vehicles you control gain trample until end of turn. " +
				"Put two +1/+1 counters on this creature.",
			Exhaust: true,
			Cost:    ManaCost("{4}{R}{G}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				if err := (GrantKeywordUntilEOT{
					Match:    And(YouControl(), Or(Creature(), Subtype("Vehicle"))),
					Keywords: []string{"trample"},
					Label:    "Boom Scholar — trample until end of turn",
				}).Apply(ctx); err != nil {
					return err
				}
				return AddCounter{
					Target: item.SourceCardID,
					Kind:   game.CounterPlusOne,
					N:      2,
				}.Apply(ctx)
			},
		}},
	})
}
