package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// The Great Henge — Legendary Artifact {7}{G}{G}:
//
//	"This spell costs {X} less to cast, where X is the greatest power
//	 among creatures you control.
//	 {T}: Add {G}{G}. You gain 2 life.
//	 Whenever a nontoken creature you control enters, put a +1/+1
//	 counter on it and draw a card."
//
// Three clauses, three slots, and the last of them to arrive was the
// first: #746 made `SelfCostModifiers` take an amount computed per
// cast, which is what Ghalta already uses. The reduction spends
// against GENERIC mana only and stops there, so with a big enough
// creature the Henge costs exactly {G}{G} and never less — the
// printed ruling, and the reason the card is playable at all.
//
// "GREATEST power", not total: one 9/9 pays for it and nine 1/1s do
// not. Power is read post-layer and after counters, so a creature
// pumped in response is counted (the cost is locked in at CR 601.2f,
// with the spell already on the stack, so a pump AFTER announce is
// too late).
//
// "You gain 2 life" is a mana-ability RIDER, not a cost and not a
// second ability: it happens after the mana lands in the pool, in the
// same atomic resolution (CR 605.3b), and it happens whether or not
// the controller needed it. The painlands' damage rider is the same
// slot with the sign flipped.
//
// The trigger reads the ENTERING permanent, so it captures ev.CardID
// in Build rather than using one of the shorthand constructors: the
// counter goes on the creature that entered, not on the Henge. It
// checks the creature is still on the battlefield when the trigger
// resolves — a creature killed in response gets no counter, and the
// draw still happens, which is what "put a +1/+1 counter on it and
// draw a card" does when the first half cannot (CR 608.2c).
//
// "Nontoken" is printed and load-bearing: a Scute Swarm turn does not
// draw a library.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "78427103-9543-41fb-b6d4-72963fe87275",
		Name:         "The Great Henge",
		Completeness: CompletenessFull,
		SelfCostModifiers: []game.CostModifier{
			CostsLessEach(greatestPowerAmongCreaturesYouControl(),
				"This spell costs {X} less to cast, where X is the greatest power among creatures you control."),
		},
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{G}{G}",
			Label:    "Add {G}{G}. You gain 2 life",
			Rider: func(g *game.Game, controller, source uuid.UUID) error {
				return g.ChangePlayerLifeForEffect(source, controller, 2)
			},
		}},
		Triggered: []game.TriggeredAbility{{
			Watches:   []game.EventKind{game.EventETB},
			AppliesTo: nontokenCreatureEnteredUnderYourControl,
			Key:       "The Great Henge — put a +1/+1 counter on it and draw a card",
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				entered := item.Trigger.Event.CardID
				if onBattlefield(g, entered) {
					if err := (AddCounter{Target: entered, Kind: "+1/+1", N: 1}.Apply(ctx)); err != nil {
						return err
					}
				}
				return DrawCards{Player: item.Controller, N: 1}.Apply(ctx)
			},
		}},
	})
}

// greatestPowerAmongCreaturesYouControl is the Henge's X: the biggest
// single power among the caster's creatures, the Henge itself never
// being one (it is on the stack). A board with no creatures reduces
// nothing, and a negative power is floored at zero rather than
// charging extra.
func greatestPowerAmongCreaturesYouControl() func(q game.CostQuery) int {
	return func(q game.CostQuery) int {
		if q.Game == nil || q.Game.Battlefield == nil {
			return 0
		}
		best := 0
		for _, c := range q.Game.Battlefield.Cards {
			if c.Controller != q.Controller || !c.IsCreature() {
				continue
			}
			if p := c.CurrentPower(); p > best {
				best = p
			}
		}
		return best
	}
}

// nontokenCreatureEnteredUnderYourControl is "whenever a nontoken
// creature you control enters" — the Henge's clause and Ohran
// Frostfang's. The Henge itself is not a creature, so no "another" is
// needed.
func nontokenCreatureEnteredUnderYourControl(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	c, ok := enteredUnderYourControl(ev, source, g, false)
	return ok && c.IsCreature() && !c.IsToken()
}
