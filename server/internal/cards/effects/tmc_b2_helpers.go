package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// tmc_b2_helpers.go — shared bodies for Teenage Mutant Ninja Turtles
// Commander, batch 2 (#5011).

// APlayerIsAtOrBelowHalfTheirStartingLife passes when some player still in
// the game has a life total less than or equal to half their starting
// life total: "This spell costs {2} less to cast if a player's life total
// is less than or equal to half their starting life total" (Game Over,
// CR 601.2f). Half is not rounded: 2·life ≤ starting life, so a 41-life
// start would want 20 or less.
func APlayerIsAtOrBelowHalfTheirStartingLife() CostPredicate {
	return func(q game.CostQuery) bool {
		if q.Game == nil {
			return false
		}
		start := startingLifeOf(q.Game)
		for _, p := range q.Game.Seats {
			if p != nil && !p.Eliminated && 2*p.Life <= start {
				return true
			}
		}
		return false
	}
}

// tmcB2LeftCreature is the creature an EventLTB names, for "whenever …
// creature … leaves the battlefield" (Rat King, Pale Piper): any exit,
// not only dying, with the creature type read as it last existed on the
// battlefield (CR 603.10a, leftAsType).
func tmcB2LeftCreature(ev game.Event, g *game.Game) (game.Card, bool) {
	if ev.Kind != game.EventLTB {
		return game.Card{}, false
	}
	c, ok := g.LookupCardForEffect(ev.CardID)
	if !ok || !leftAsType(ev, c, "creature") {
		return game.Card{}, false
	}
	return c, true
}

// tmcB2CountersAmongPermanentsYouControl sums every counter of every
// kind on the permanents `controller` controls — "the number of counters
// among permanents you control" (Dimension X Pizzasaur).
func tmcB2CountersAmongPermanentsYouControl(g *game.Game, controller uuid.UUID) int {
	n := 0
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller != controller {
			continue
		}
		for _, v := range c.Counters {
			if v > 0 {
				n += v
			}
		}
	}
	return n
}

// tmcB2AnotherCreatureDied is the delayed condition "whenever another
// creature dies" for a delayed trigger whose "this creature" is its
// source object (Electric Seaweed): a creature died, and it was not that
// object. The dead card's epoch is one past the one it had on the
// battlefield (a zone move starts a new object, CR 400.7), so the source
// card dying again after it came back as a new object is "another".
func tmcB2AnotherCreatureDied(ev game.Event, dt *game.DelayedTrigger, g *game.Game, _ game.EffectParams) bool {
	dead, ok := diedCreature(ev, g)
	if !ok {
		return false
	}
	self := dt.SourceObject
	if self.ID == uuid.Nil {
		self.ID = dt.SourceCardID
	}
	return dead.InstanceID != self.ID || dead.ObjectEpoch-1 != self.Epoch
}

var (
	// "Whenever another creature dies" (Electric Seaweed, #5011).
	tmcB2AnotherCreatureDiedCondition = game.DelayedCondition("dies/another-creature-than-the-source", tmcB2AnotherCreatureDied)

	// "This creature deals 1 damage to each non-Wall creature" (Electric
	// Seaweed, #5011): from the delayed trigger's source, as it last
	// existed if it has left (CR 608.2h).
	tmcB2OneDamageToEachNonWallBody = game.SimpleDelayedBody("electric-seaweed/one-damage-to-each-non-wall-creature",
		func(g *game.Game, item *game.StackItem) error {
			return damageEachMatching(NewContext(g, item), And(Creature(), Not(OfCreatureType("Wall"))), 1)
		})
)

// goadAllCreaturesYourOpponentsControl is "goad all creatures your
// opponents control" (Fast Forward) and "goad all creatures you don't
// control" (Disrupt Decorum): every creature on the battlefield whose
// controller is not the resolving item's, as the effect resolves.
func goadAllCreaturesYourOpponentsControl(_ *game.StackItem, ctx *Context) error {
	return GoadAllMatching(ctx, func(_ *game.Game, c game.Card) bool {
		return c.Controller != ctx.Controller()
	})
}

// OpponentsYouAttackedThisTurn counts the opponents still in the game
// that `you` attacked this turn: a creature of yours was declared
// attacking that player (an attack on a planeswalker or battle is not an
// attack on its player). Only the active player declares attackers
// (CR 508.1), so off your own turn the count is zero.
func OpponentsYouAttackedThisTurn(g *game.Game, you uuid.UUID) int {
	active := g.ActivePlayer()
	if active == nil || active.ID != you {
		return 0
	}
	seen := map[uuid.UUID]bool{}
	for _, a := range g.TurnTally.Attacks {
		if a.Defender == you || seen[a.Defender] {
			continue
		}
		if p := g.PlayerByIDForEffect(a.Defender); p == nil || p.Eliminated {
			continue
		}
		seen[a.Defender] = true
	}
	return len(seen)
}

// mutagenIfDealtDamage is "you create a Mutagen token for each creature
// dealt damage this way" (Shellshock) as a damage recipient's rider: a
// recipient that was actually dealt more than 0 damage (CR 120.8: X of 0,
// or damage prevented, is none) earns its controller-of-the-spell one
// Mutagen token. A creature the damage is lethal to still counts: it was
// dealt the damage before state-based actions see it. An on-disk key.
var mutagenIfDealtDamage = recipientTail("recipient/create-a-mutagen-if-dealt-damage", func(ctx *Context, _ uuid.UUID, dealt int, _ game.TailOutcome) error {
	if dealt <= 0 {
		return nil
	}
	return CreateToken{Template: MutagenToken(), N: 1}.Apply(ctx)
})
