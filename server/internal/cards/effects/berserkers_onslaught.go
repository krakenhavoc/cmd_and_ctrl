package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Berserkers' Onslaught — Enchantment {3}{R}{R} (EDHREC rank 2382):
//
//	"Attacking creatures you control have double strike."
//
// The aggro deck's five-mana anthem. Printed as a static, written as
// a trigger, and the reason is an engine gap worth naming: the layer
// engine recomputes only when its version counter moves, and
// declaring an attacker moves nothing — no EventTapCard (the tap is
// written directly), no zone move, no counter — so a layer 6 static
// gated on Card.AttackingTarget stays cached from before the attack
// and the creature reaches combat damage without double strike. A
// probe test showed exactly that (2 damage from a 2/2 attacker, not
// 4). So the card watches EventAttack instead — one trigger per
// attacking creature the controller controls, the Adeline shape —
// and grants double strike until end of turn through the turn-scoped
// static registry, which does invalidate the cache.
//
// Declared simplification: the grant lasts the turn rather than the
// attack, and lands only on creatures DECLARED as attackers. A
// creature put onto the battlefield attacking is not covered
// (weaker than printed); a creature removed from combat keeps double
// strike until end of turn, which changes nothing outside combat
// (an attacker cannot block the same turn, and a fight is not
// combat damage). Never stronger. The seam is a layer-cache bump on
// attack declaration, after which this becomes a one-line static.
func init() {
	Register(Spec{
		OracleID:     "85d2948c-1a79-418d-80c4-bc1012a4d313",
		Name:         "Berserkers' Onslaught",
		Completeness: CompletenessCaveats,
		Caveats:      []string{"Double strike is granted as each of your creatures is declared as an attacker and lasts until end of turn, so a creature that enters the battlefield already attacking doesn't get it."},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventAttack},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return attackDeclaredByYou(ev, source.Controller)
			},
			Key: "Berserkers' Onslaught — the attacking creature has double strike",
			Effect: func(g *game.Game, item *game.StackItem) error {
				if item.Trigger == nil {
					return nil
				}
				return GrantKeywordUntilEOT{
					Target:   item.Trigger.Event.CardID,
					Keywords: []string{"double strike"},
					Label:    "Berserkers' Onslaught — double strike",
				}.Apply(NewContext(g, item))
			},
		}},
	})
}
