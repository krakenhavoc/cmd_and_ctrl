package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// activation_conditions.go — builders for an ability's "Activate only
// if …" and "Activate only during …" instructions (CR 602.1b, #743,
// ADR 0020's activation-condition addendum).
//
// Each builder returns the closure ActivatedAbility.Condition and
// ManaAbility.Condition both take, so a helper written for one kind of
// ability works for the other. ControlsAtLeast, the first of them,
// predates this file and lives in mana_derivation.go.
//
// The contract every closure here keeps:
//
//   - READ-ONLY. It runs under g.mu — for write inside the activation,
//     for read in the view and the legal enumerator — so it reads
//     BattlefieldCardsForEffect, PlayerByIDForEffect, g.Turn and
//     g.Seats, and never a public locking accessor.
//   - PUBLIC INFORMATION ONLY. Every viewer receives the
//     condition_unmet flag the view computes from it.
//   - "You" is `controller`, the activating player; "this" is `source`.
//
// A builder is added with its first card, not before. Append-only: a
// card PR adds a builder, it does not change what an existing one
// means.

// ActivationCondition is the shape of ActivatedAbility.Condition and
// ManaAbility.Condition.
type ActivationCondition = func(g *game.Game, controller, source uuid.UUID) bool

// IsYourTurn reports whether `controller` is the active player. Reads
// g.Turn and g.Seats directly, because g.ActivePlayer takes the read
// lock and a condition already runs under g.mu.
func IsYourTurn(g *game.Game, controller uuid.UUID) bool {
	if g.Turn.ActiveSeat < 0 || g.Turn.ActiveSeat >= len(g.Seats) {
		return false
	}
	active := g.Seats[g.Turn.ActiveSeat]
	return active != nil && active.ID == controller
}

// DuringYourTurn — "Activate only during your turn" (Sanctum of
// Eternity). Any step of your turn, with or without an item on the
// stack; not a sorcery-speed window.
func DuringYourTurn() ActivationCondition {
	return func(g *game.Game, controller, _ uuid.UUID) bool {
		return IsYourTurn(g, controller)
	}
}

// DuringStep — "Activate only during the end of combat step"
// (Desert). Any player's such step, since the printed clause names a
// step and not a turn; pair it with DuringYourTurn when a card wants
// both.
//
// Deliberately NOT SorcerySpeed: a sorcery-speed gate means "your
// main phase with an empty stack", which would forbid exactly the
// window this clause OPENS. The two are different instructions and
// #743 exists because the engine used to conflate them.
func DuringStep(step game.Step) ActivationCondition {
	return func(g *game.Game, _, _ uuid.UUID) bool {
		return g.Turn.Step == step
	}
}

// countControlled counts the battlefield permanents `player` controls
// that match.
func countControlled(g *game.Game, player uuid.UUID, match func(game.Card) bool) int {
	n := 0
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == player && match(c) {
			n++
		}
	}
	return n
}

// eachOpponent calls fn for every seated, non-eliminated player other
// than `controller`, and reports whether fn returned true for any.
func eachOpponent(g *game.Game, controller uuid.UUID, fn func(opponent uuid.UUID) bool) bool {
	for _, p := range g.Seats {
		if p == nil || p.ID == controller || p.Eliminated {
			continue
		}
		if fn(p.ID) {
			return true
		}
	}
	return false
}

// OpponentControlsAtLeast — "Activate only if an opponent controls
// four or more lands" (Tectonic Edge). True when ONE opponent controls
// at least n on their own: in a four-player game the opponents' lands
// are not added together.
func OpponentControlsAtLeast(n int, match func(game.Card) bool) ActivationCondition {
	return func(g *game.Game, controller, _ uuid.UUID) bool {
		return eachOpponent(g, controller, func(opp uuid.UUID) bool {
			return countControlled(g, opp, match) >= n
		})
	}
}

// OpponentControlsMore — "Activate only if an opponent controls more
// lands than you" (Weathered Wayfarer). The same per-opponent
// comparison Keeper of the Accord's intervening-if makes, against the
// activator's own count.
func OpponentControlsMore(match func(game.Card) bool) ActivationCondition {
	return func(g *game.Game, controller, _ uuid.UUID) bool {
		return eachOpponent(g, controller, func(opp uuid.UUID) bool {
			return b11OpponentControlsMoreThanYou(g, controller, opp, match)
		})
	}
}

// GraveyardAtLeast — threshold, "Activate only if there are seven or
// more cards in your graveyard" (Barbarian Ring, Cephalid Coliseum).
// A nil match counts every card.
func GraveyardAtLeast(n int, match func(game.Card) bool) ActivationCondition {
	return func(g *game.Game, controller, _ uuid.UUID) bool {
		p := g.PlayerByIDForEffect(controller)
		if p == nil || p.Graveyard == nil {
			return false
		}
		count := 0
		for _, c := range p.Graveyard.Cards {
			if match == nil || match(c) {
				count++
			}
		}
		return count >= n
	}
}

// NoCardsInHand — "Activate only if you have no cards in hand" (Sea
// Gate Wreckage). A hand's size is public; its contents are not read.
func NoCardsInHand() ActivationCondition {
	return func(g *game.Game, controller, _ uuid.UUID) bool {
		p := g.PlayerByIDForEffect(controller)
		return p != nil && (p.Hand == nil || len(p.Hand.Cards) == 0)
	}
}

// SourceHasCountersAtLeast — "Activate only if this enchantment has
// four or more quest counters on it" (Luminarch Ascension).
func SourceHasCountersAtLeast(kind string, n int) ActivationCondition {
	return func(g *game.Game, _, source uuid.UUID) bool {
		c, ok := g.LookupCardForEffect(source)
		return ok && c.Counters[kind] >= n
	}
}

// LifeAtLeastAboveStarting — "Activate only if you have at least 7
// life more than your starting life total" (Speaker of the Heavens).
// The starting total is the format's, game.StartingLife.
func LifeAtLeastAboveStarting(n int) ActivationCondition {
	return func(g *game.Game, controller, _ uuid.UUID) bool {
		p := g.PlayerByIDForEffect(controller)
		return p != nil && p.Life >= game.StartingLife+n
	}
}

// MatchCreatureWithPowerAtLeast matches a creature whose current
// power is at least n — Bonders' Enclave's "a creature with power 4
// or greater". Post-layer, counters included.
func MatchCreatureWithPowerAtLeast(n int) func(game.Card) bool {
	return func(c game.Card) bool {
		return c.IsCreature() && c.CurrentPower() >= n
	}
}

// MatchColor matches a permanent of the given colour letter —
// Leechridden Swamp's "two or more black permanents". Reads
// Card.HasColor, the same colour test the target predicates use.
func MatchColor(color string) func(game.Card) bool {
	return func(c game.Card) bool { return c.HasColor(color) }
}

// MatchLegendaryCreature matches a legendary creature — Rivendell's
// and Minas Tirith's "if you control a legendary creature". The same
// test b08ControlsLegendaryCreature makes for their entry.
func MatchLegendaryCreature(c game.Card) bool {
	return c.IsCreature() && c.IsLegendary()
}

// AllConditions is the AND of several activation conditions, for a
// card that prints two of them in one sentence — Vivi Ornitier's
// "Activate only during your turn and only once each turn".
//
// Every condition must hold. An empty list is true, which is the
// identity an unconditional ability already has. The conditions are
// asked in the order given and the walk stops at the first false, so
// a cheap gate (whose turn is it) can be written before an expensive
// one (a walk of this turn's events).
func AllConditions(conds ...ActivationCondition) ActivationCondition {
	return func(g *game.Game, controller, source uuid.UUID) bool {
		for _, cond := range conds {
			if cond != nil && !cond(g, controller, source) {
				return false
			}
		}
		return true
	}
}

// SourceIsAttacking — "Activate only if this creature is attacking"
// (Glint-Horn Buccaneer). Reads Card.AttackingTarget directly: a
// non-nil target means DeclareAttacker has marked the permanent an
// attacker and it hasn't left combat since (ClearCombat and a zone
// exit both clear the field).
func SourceIsAttacking() ActivationCondition {
	return func(g *game.Game, _, source uuid.UUID) bool {
		c, ok := g.LookupCardForEffect(source)
		return ok && c.AttackingTarget != uuid.Nil
	}
}

// DuringYourUpkeep — "Activate only during your upkeep" (Magus of the
// Mirror). Both halves are load-bearing: the step is the upkeep AND
// the activator is the active player, so an opponent's upkeep is not
// a window. Not a sorcery-speed gate — the upkeep is not a main
// phase, and SorcerySpeed would forbid exactly the step the card
// opens.
func DuringYourUpkeep() ActivationCondition {
	return func(g *game.Game, controller, _ uuid.UUID) bool {
		return g.Turn.Step == game.StepUpkeep && IsYourTurn(g, controller)
	}
}

// OncePerTurnActivation — "Activate only once each turn" (CR 602.1b)
// on an ordinary activated ability with no cost component of its own
// to carry the limit (a loyalty ability gets its once-per-turn from
// LoyaltyCost; this is for everything else — Beledros Witherbloom's
// "Pay 10 life: Untap all lands you control").
//
// Reads Game.ActivatedThisTurn(source, label): the per-OBJECT (CR
// 400.7) count of how many times THIS ability has been ACTIVATED this
// turn. Since the Condition runs at announce — before any cost is
// paid, and the enumerator and the view consult the same closure — a
// second attempt this turn is never offered rather than paid for and
// then doing nothing.
//
// #1213 moved it off ResolvedThisTurn. The rule counts announcements
// (CR 602.1b, "Activate only once each turn"), and the two tallies
// disagree in both directions on a printed line:
//
//   - two activations held on the stack at once both RESOLVE later, so
//     a resolution count let the second one be announced — stronger
//     than printed, the #259 direction;
//   - an activation countered or fizzled never resolves, so a
//     resolution count handed the player a second use of an ability
//     they had already spent.
//
// The activation tally (activation_tally.go, #1181/#1183) is written
// at the announce on both activation paths, which is exactly where the
// rule is read.
//
// label must equal the ability's own Label exactly, the same contract
// ActivatedThisTurn keeps for every other reader of the per-object
// tally.
func OncePerTurnActivation(label string) ActivationCondition {
	return func(g *game.Game, _, source uuid.UUID) bool {
		return g.ActivatedThisTurn(source, label) == 0
	}
}
