package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Animar, Soul of Elements — Legendary Creature — Elemental
// {G}{U}{R}, 1/1:
//
//	"Protection from white and from black
//	 Whenever you cast a creature spell, put a +1/+1 counter on
//	 Animar, Soul of Elements.
//	 Creature spells you cast cost {1} less to cast for each +1/+1
//	 counter on Animar, Soul of Elements."
//
// The archetypal snowball commander, and the catalog's second
// scaling cost modifier (Damping Sphere is the first): the discount
// is a function of the counters, so the amount hook reads the
// modifier's own source.
//
// The ordering the card lives or dies by comes out right for free.
// The cost is determined at CR 601.2f, while the cast is being
// announced; the "whenever you cast" trigger goes on the stack at
// CR 603.3, once the spell is already there. So a creature cast into
// an Animar with two counters is discounted by {2}, and the third
// counter arrives afterwards for the NEXT one. A version that
// counted the counter first would discount every spell by one too
// much.
//
// And the discount spends generic mana only, which is why Animar
// decks play colourless creatures: no number of counters ever pays
// for a {G}{U} spell's coloured half.
//
// SANDBOX SIMPLIFICATION — "protection from white and from black" is
// NOT implemented. ADR 0038 explains why the keyword has no home:
// CR 702.16e tests the quality against the SOURCE of the spell or
// ability, and the engine's targeting choke point never sees the
// source object. Strictly weaker than printed — Animar is a legal
// target for a white removal spell here.
func init() {
	Register(Spec{
		OracleID:     "725880b2-1675-414f-b61b-cf6533797dbf",
		Name:         "Animar, Soul of Elements",
		Completeness: CompletenessCaveats,
		Caveats:      []string{"Protection from white and from black isn't granted, so Animar can be targeted and damaged by white and black sources."},
		CostModifiers: []game.CostModifier{
			CostsLessEach(CountersOnSource(game.CounterPlusOne),
				"Creature spells you cast cost {1} less to cast for each +1/+1 counter on Animar, Soul of Elements.",
				YourSpell(), CreatureSpell()),
		},
		Triggered: []game.TriggeredAbility{
			On(game.EventCast, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return creatureSpellCastByYou(ev, source, g)
			}, "Animar, Soul of Elements — put a +1/+1 counter on it", func(g *game.Game, item *game.StackItem) error {
				return AddCounter{
					Target: item.SourceCardID,
					Kind:   game.CounterPlusOne,
					N:      1,
				}.Apply(NewContext(g, item))
			}),
		},
	})
}
