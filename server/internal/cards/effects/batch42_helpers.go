package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// batch42_helpers.go — the shared bodies behind the card-coverage
// roadmap's batch 42 (#449, `edhrec_rank` 4354–4455). Own file per the
// #231 convention; every package-level name carries the b42 prefix
// because other batches land beside this one.
//
// What is NOT here, because the package already had it: "when this
// permanent enters" is WhenThisEnters, "whenever you cast a <kind>
// spell" is WheneverYouCast, "whenever another creature you control
// dies" is WheneverACreatureYouControlDies with AnotherCreatureDied,
// "landfall" is Landfall, "at the beginning of your end step" is
// AtYourEndStep, the "you may" gate on a trigger is Optional, the
// keyword grant over a predicate is b16GrantKeywords, the tribal one
// is TribalKeywordGrant, "N damage to each opponent" is
// damageToEachOpponent, "each opponent discards a card" is
// eachOpponentDiscardsOne, "draw then discard" is lootOne, "the life
// you've lost this turn" is b18LifeLostThisTurn, and "a creature card
// left your graveyard" is b02CreatureCardLeftYourGraveyard.

// --- cost predicates -----------------------------------------------

// b42CreatureSpellWithFlying is Watcher of the Spheres' "creature
// spells with flying you cast". Read through game.HasKeyword rather
// than off the type line, because the spell is in no zone the layer
// engine reaches and HasKeyword is the accessor that falls back to
// the catalog's printed keywords for exactly that case.
func b42CreatureSpellWithFlying() CostPredicate {
	return func(q game.CostQuery) bool {
		c := q.Card
		return c.IsCreature() && game.HasKeyword(&c, "flying")
	}
}

// --- board reads ---------------------------------------------------

// b42CountCreaturesWithSubtype counts every creature on the
// battlefield with the named subtype, whoever controls it —
// Timberwatch Elf's "the number of Elves on the battlefield" counts
// the whole table's, as printed.
//
// Effective subtypes, so a changeling and anything a layer-4 effect
// has retyped are Elves here too.
func b42CountCreaturesWithSubtype(g *game.Game, subtype string) int {
	n := 0
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.IsCreature() && c.HasSubtype(subtype) {
			n++
		}
	}
	return n
}

// b42GreatestPowerControlledBy is the scalar behind Orcish
// Siegemaster's "X is the greatest power among creatures you
// control". Layered power (CurrentPower), and the Siegemaster counts
// itself — the printed text says "creatures you control", not
// "other".
//
// Zero when the controller has no creatures, which cannot happen
// while the trigger's own source is attacking but is the right answer
// for any other caller.
func b42GreatestPowerControlledBy(g *game.Game, controller uuid.UUID) int {
	best := 0
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller != controller || !c.IsCreature() {
			continue
		}
		if p := c.CurrentPower(); p > best {
			best = p
		}
	}
	return best
}

// b42ControlsPlaneswalkerNamed reports whether `player` controls a
// planeswalker with the named subtype — Liliana's Triumph's "if you
// control a Liliana planeswalker". Subtype, not card name: every
// Liliana card shares the Liliana planeswalker type and a Liliana
// that is not a planeswalker (a creature card named Liliana) does not
// count.
func b42ControlsPlaneswalkerNamed(g *game.Game, player uuid.UUID, subtype string) bool {
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == player && c.IsPlaneswalker() && c.HasSubtype(subtype) {
			return true
		}
	}
	return false
}

// b42CardsInGraveyardOf collects the cards in `player`'s graveyard
// that match — Crypt Incursion's "all creature cards from target
// player's graveyard". Snapshot first, act second: the exile opens a
// CR 614 window per card and the zone is not safe to walk while it
// is being emptied.
func b42CardsInGraveyardOf(g *game.Game, player uuid.UUID, match func(game.Card) bool) []uuid.UUID {
	p := g.PlayerByIDForEffect(player)
	if p == nil {
		return nil
	}
	var ids []uuid.UUID
	for _, c := range p.Graveyard.Cards {
		if match == nil || match(c) {
			ids = append(ids, c.InstanceID)
		}
	}
	return ids
}

// --- static-ability scopes -----------------------------------------

// b42CreatureYouControlWithAPlusOneCounter is Pridemalkin's second
// sentence: "each creature you control with a +1/+1 counter on it".
// Counters are read off the live card rather than the layered
// characteristic because a counter is not a continuous effect — it is
// a physical marker the layer pass reads, not one it writes.
func b42CreatureYouControlWithAPlusOneCounter(target *game.Card, _ *game.Game, source *game.Card) bool {
	return target.IsCreature() &&
		target.Controller == source.Controller &&
		target.Counters[game.CounterPlusOne] > 0
}

// --- tokens --------------------------------------------------------

// b42BlueMerfolkHexproofToken is kept as a function because a card
// passes it as a value; the data lives in tokens_table.go.
func b42BlueMerfolkHexproofToken() game.Card {
	return TokenCard("1/1 blue Merfolk with hexproof")
}

// b42BlackBatFlyingToken is kept as a function because a card passes
// it as a value; the data lives in tokens_table.go.
func b42BlackBatFlyingToken() game.Card { return TokenCard("1/1 black Bat with flying") }
