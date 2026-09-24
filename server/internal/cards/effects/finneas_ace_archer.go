package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Finneas, Ace Archer — Legendary Creature — Rabbit Archer {G}{W},
// 2/2 (EDHREC rank 4521):
//
//	"Reach, vigilance
//	 Whenever Finneas attacks, put a +1/+1 counter on each other
//	 creature you control that's a token or a Rabbit. Then if
//	 creatures you control have total power 10 or greater, draw a
//	 card."
//
// A two-mana Selesnya commander that grows a token board every
// combat and draws off it. The vigilance is what makes the loop
// work: Finneas attacks, the board gets bigger, and he is still
// untapped to block — so the deck can go wide without ever opening
// itself up.
//
// "Each OTHER creature you control that's a TOKEN OR A RABBIT" —
// Finneas is a Rabbit and is excluded by the "other". The two halves
// of the predicate are an OR, so an opponent's token gets nothing and
// your nontoken Rabbit gets a counter.
//
// The "then if" is the intervening-if's cousin: a second condition
// checked as the trigger resolves, AFTER the counters have landed, so
// those counters count towards the ten. Total power is CurrentPower
// summed over the creatures the controller has at that moment —
// counters, anthems and pumps all count (CR 208.3), and a creature
// shrunk below zero contributes nothing rather than subtracting.
//
// The affected set is snapshotted BEFORE the first counter lands, so
// the walk cannot see its own work; the counters are then applied to
// that fixed list.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "69924636-138e-4baa-a378-0fd7df5b847d",
		Name:            "Finneas, Ace Archer",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"reach", "vigilance"},
		Triggered: []game.TriggeredAbility{
			On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return attackDeclared(ev, source)
			}, "Finneas, Ace Archer — a +1/+1 counter on each other token or Rabbit, then draw on total power 10", func(g *game.Game, item *game.StackItem) error {
				var ids []uuid.UUID
				for _, c := range g.BattlefieldCardsForEffect() {
					if c.InstanceID == item.SourceCardID || c.Controller != item.Controller || !c.IsCreature() {
						continue
					}
					if IsToken(c) || c.HasSubtype("Rabbit") {
						ids = append(ids, c.InstanceID)
					}
				}
				// #1290: total power has to be read AFTER every
				// counter in the batch has LANDED, not on the next
				// line — any one of them can pause on a CR 616
				// ordering prompt (a Doubling Season / Hardened
				// Scales board), and reading before it resumes would
				// undercount.
				return b13PutCounterOnEachThen(NewContext(g, item), ids, func(g *game.Game) error {
					if b43TotalPowerControlled(g, item.Controller) < 10 {
						return nil
					}
					return DrawCards{Player: item.Controller, N: 1}.Apply(NewContext(g, item))
				})
			}),
		},
	})
}
