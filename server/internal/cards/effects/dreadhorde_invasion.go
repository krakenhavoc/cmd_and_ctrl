package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Dreadhorde Invasion — Enchantment {1}{B} (EDHREC rank 1476):
//
//	"At the beginning of your upkeep, you lose 1 life and amass
//	 Zombies 1.
//	 Whenever a Zombie token you control with power 6 or greater
//	 attacks, it gains lifelink until end of turn."
//
// The card that makes amass's FIND-OR-CREATE half visible: the first
// upkeep has no Army and mints one, and every upkeep after that finds
// the same Army and adds a counter to it. A two-mana enchantment that
// built a new 0/0 every turn would be a different (and much worse)
// card, which is exactly the shape #1236 said no existing primitive
// could express.
//
// The second ability reads "a Zombie token", and after an amass the
// Army token IS a Zombie — the token is created as one (CR 701.47a),
// so the lifelink clause finds it with no help from this file. A
// 6-power Army is six upkeeps in, or fewer with a doubler.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "01deabbb-af6a-4998-99a7-35b7cfa9ef77",
		Name:         "Dreadhorde Invasion",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			AtYourUpkeep("Dreadhorde Invasion — lose 1 life, amass Zombies 1",
				func(g *game.Game, item *game.StackItem) error {
					if err := g.ChangePlayerLifeForEffect(item.SourceCardID, item.Controller, -1); err != nil {
						return err
					}
					return Amass{Subtype: "Zombie", N: 1}.Apply(NewContext(g, item))
				}),
			On(game.EventAttack, dreadhordeBigZombieTokenAttacked,
				"Dreadhorde Invasion — lifelink until end of turn",
				dreadhordeGrantLifelink),
		},
	})
}

// dreadhordeBigZombieTokenAttacked is "whenever a Zombie token you
// control with power 6 or greater attacks".
//
// CurrentPower, not Effective().Power: +1/+1 counters are added on
// top of the layered value (Card.PowerForComparison), and the counters
// the amass put on the Army are the only way this clause is ever
// reached in a deck built around the card.
func dreadhordeBigZombieTokenAttacked(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.Kind != game.EventAttack {
		return false
	}
	c, ok := g.LookupCardForEffect(ev.CardID)
	if !ok {
		return false
	}
	return c.Controller == source.Controller && c.IsToken() &&
		c.HasSubtype("Zombie") && c.CurrentPower() >= 6
}

// dreadhordeGrantLifelink is "it gains lifelink until end of turn" —
// the attacker the trigger fired on, captured at Build time by the
// event and re-read here.
func dreadhordeGrantLifelink(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	attacker := ctx.Trigger().Event.CardID
	if attacker == uuid.Nil {
		return nil
	}
	return GrantKeywordUntilEOT{
		Target:   attacker,
		Keywords: []string{"lifelink"},
		Label:    "Dreadhorde Invasion — lifelink until end of turn",
	}.Apply(ctx)
}
