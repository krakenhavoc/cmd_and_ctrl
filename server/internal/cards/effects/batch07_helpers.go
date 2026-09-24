package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// batch07_helpers.go — the shared bodies behind the card-coverage
// roadmap's batch 07 (#300, `edhrec_rank` 798–901). Own file per the
// #231 convention; every package-level name carries the b07 prefix
// because batches 05 and 06 are being written in parallel.
//
// What is NOT here, because main already had it by the time this
// batch was finished: "an opponent loses life" is b04OpponentLostLife,
// "nonbasic" is b03Nonbasic, "another" by name is b03NotNamed, and
// the 1/1 flying Thopter is tokens.go's ThopterToken.

// b07IsModified is CR 700.9's "modified": a creature with a counter
// on it, equipped, or enchanted by an Aura its controller controls —
// Kodama of the West Tree's word. Counters are read off the card;
// the two attachment halves walk the battlefield for anything whose
// AttachedTo points at the creature (#379's relation), so an
// opponent's Aura does NOT modify (as printed — "Auras you control")
// but an opponent's Equipment somehow attached to your creature
// does. Reads effective subtypes, so it is safe inside a layer
// recompute, where the static that grants trample calls it.
//
// Caller must hold g.mu.
func b07IsModified(g *game.Game, c game.Card) bool {
	for _, n := range c.Counters {
		if n > 0 {
			return true
		}
	}
	if g.Battlefield == nil {
		return false
	}
	for i := range g.Battlefield.Cards {
		a := &g.Battlefield.Cards[i]
		if !a.IsAttachedTo(c.InstanceID) {
			continue
		}
		if a.HasSubtype("Equipment") {
			return true
		}
		if a.IsAura() && a.Controller == c.Controller {
			return true
		}
	}
	return false
}

// drainEachOpponent is "each opponent loses 1 life and you gain 1
// life" — Ayara's ETB, Nadier's Nightblade.
func drainEachOpponent(g *game.Game, item *game.StackItem) error {
	if err := eachOpponentLosesLife(g, item, 1); err != nil {
		return err
	}
	return GainLife{Player: item.Controller, Amount: 1}.Apply(NewContext(g, item))
}

// b07TripleDamageFromYourSources is Fiery Emancipation and City on
// Fire: "if a source you control would deal damage to a permanent or
// player, it deals triple that damage instead." Angrath's Marauders'
// doubler with a 3, and — unlike Solphim — combat damage included.
func b07TripleDamageFromYourSources() game.ReplacementEffect {
	return game.ReplacementEffect{
		Watches: []game.EventKind{game.EventDealDamage},
		AppliesTo: func(ev *game.ReplacementEvent, g *game.Game, src *game.Card) bool {
			if ev.Kind != game.RepEventDamage || ev.DamageAmount <= 0 {
				return false
			}
			return damageSourceControlledBy(ev, g, src.Controller)
		},
		Replace: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) error {
			ev.DamageAmount *= 3
			return nil
		},
		Controller: func(_ *game.ReplacementEvent, _ *game.Game, src *game.Card) uuid.UUID {
			return src.Controller
		},
	}
}

// b07SearchBasicOntoBattlefield is "search your library for a basic
// land card, put it onto the battlefield (tapped), then shuffle" for
// `player`, from the resolving item — Kodama's trigger, Ghost
// Quarter's and Field of Ruin's consolation searches. `player` need
// not be the item's controller: Ghost Quarter's search belongs to
// the destroyed land's controller, and the prompt is addressed to
// whoever `player` is.
func b07SearchBasicOntoBattlefield(g *game.Game, item *game.StackItem, player uuid.UUID, tapped, optional bool, reason string) error {
	return SearchLibrary{
		Player:        player,
		Predicate:     IsBasicLand,
		Dest:          game.ZoneBattlefield,
		Limit:         1,
		Shuffle:       true,
		TappedOnEntry: tapped,
		Optional:      optional,
		Reason:        reason,
	}.Apply(NewContext(g, item))
}

// b07TemptWithDiscoveryOffer is the Commander 2013 "tempting offer",
// run from opponent `i` onward, one opponent at a time in seat
// order:
//
//	"Each opponent may search their library for a land card and put
//	 it onto the battlefield. For each opponent who searches a
//	 library this way, search your library for a land card and put
//	 it onto the battlefield."
//
// Strictly sequential — each step's Then launches the next — for the
// reason Cultivate's comment gives: two open prompts over the SAME
// library would offer a card its owner is in the middle of taking,
// and every opponent who accepts hands the caster another search of
// the caster's own library. Serialising is also what the printed
// text does: the offer is made to each opponent in turn order.
//
// "Who searches" is read as "who took a land": an opponent's decline
// and an opponent's search-and-take-nothing both come back as an
// empty find, and the engine cannot tell them apart. Weaker than
// printed (an opponent who searched for nothing would owe the caster
// a land), never stronger, and declared on the card.
//
// A package-level func, not a closure over the Context: the chain
// lives on pending choices that outlive the resolution that started
// it, and it captures only IDs.
func b07TemptWithDiscoveryOffer(g *game.Game, controller, source uuid.UUID, opponents []uuid.UUID, i int) error {
	if i >= len(opponents) {
		return nil
	}
	next := func(g *game.Game) error {
		return b07TemptWithDiscoveryOffer(g, controller, source, opponents, i+1)
	}
	return g.SearchLibraryThenForEffect(game.SearchLibrarySpec{
		Player:   opponents[i],
		Source:   source,
		Pred:     b03IsLandCard,
		Dest:     game.ZoneBattlefield,
		Limit:    1,
		Shuffle:  true,
		Optional: true,
		Reason:   "Tempt with Discovery — you may search for a land (the caster gets one too if you do)",
		Then: func(g *game.Game, found []uuid.UUID) error {
			if len(found) == 0 {
				return next(g)
			}
			return b07TemptWithDiscoveryCasterSearch(g, controller, source, next)
		},
	})
}

// b07TemptWithDiscoveryCasterSearch is the caster's own "search your
// library for a land card and put it onto the battlefield" — the
// opening one and every one an opponent's acceptance earns. Not
// optional and untapped, as printed. `then` continues the offer once
// this search has settled.
func b07TemptWithDiscoveryCasterSearch(g *game.Game, controller, source uuid.UUID, then func(*game.Game) error) error {
	return g.SearchLibraryThenForEffect(game.SearchLibrarySpec{
		Player:  controller,
		Source:  source,
		Pred:    b03IsLandCard,
		Dest:    game.ZoneBattlefield,
		Limit:   1,
		Shuffle: true,
		Reason:  "Tempt with Discovery — a land onto the battlefield",
		Then: func(g *game.Game, _ []uuid.UUID) error {
			return then(g)
		},
	})
}
