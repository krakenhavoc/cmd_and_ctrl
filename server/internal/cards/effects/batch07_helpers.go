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
// "nonbasic" is b03Nonbasic, "another" is Another, and
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
		RowLabel: "If a source you control would deal damage to a permanent or player, it deals triple that damage instead.",
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
// "Who searches" is its own question: each opponent is first asked
// whether to search at all (a confirm prompt), and only a "yes" owes the
// caster a land. The search that follows is not optional — it names a
// quality, so the searcher may still fail to find (CR 701.23b) and the
// caster gets their land all the same, as printed.
//
// A package-level func, not a closure over the Context: the chain
// lives on pending choices that outlive the resolution that started
// it, and it captures only IDs.
func b07TemptWithDiscoveryOffer(ctx *Context, opponents []uuid.UUID) error {
	if len(opponents) == 0 {
		return nil
	}
	// Keyed (#4298): the opponents still to ask, after this one, ride the
	// prompt as its carried IDs, and the caster is the asking item's
	// controller.
	return MayChoice{
		Player:   opponents[0],
		Question: "Tempt with Discovery — search your library for a land (the caster gets one too if you do)?",
		Then:     b07TemptWithDiscoveryAnswer,
		Carry:    opponents[1:],
	}.Apply(ctx)
}

// b07TemptWithDiscoveryAnswer is one opponent's answer to the tempting
// offer: a yes searches for a land and earns the caster one, either way the
// next opponent is asked. Registered in init, because it asks the next
// opponent through b07TemptWithDiscoveryOffer, which names it.
var b07TemptWithDiscoveryAnswer game.ConfirmThen

func init() {
	b07TemptWithDiscoveryAnswer = MayAnswer("confirm/tempt-with-discovery-search", func(ctx *Context, r game.ConfirmAnswer) error {
		rest := r.Carry
		if !r.Accepted {
			return b07TemptWithDiscoveryOffer(ctx, rest)
		}
		// #4253 slice 4: the caster's search and the rest of the offer are
		// registered tails (the opponents still to ask ride Carry, the item
		// rides Asked), so the opponent's fetched land can ask without
		// holding a closure.
		caster := ctx.ItemTail(b07TemptCasterSearch)
		caster.Source, caster.Carry = r.Source, rest
		return ctx.Game.SearchLibraryThenForEffect(game.SearchLibrarySpec{
			Player:   r.Chooser,
			Source:   r.Source,
			Pred:     b03IsLandCard,
			Dest:     game.ZoneBattlefield,
			Limit:    1,
			Shuffle:  true,
			Reason:   "Tempt with Discovery — a land onto the battlefield",
			ThenTail: caster,
		})
	})
}

// The two tails of the chain, which run each other, so they are registered
// in init:
//
//   - b07TemptCasterSearch is the caster's search after an opponent searched;
//   - b07TemptNextOffer is the offer to the next opponent (Tail.Carry), once
//     the caster's search has settled.
//
// On-disk keys.
var b07TemptCasterSearch, b07TemptNextOffer game.TailRef

func init() {
	b07TemptCasterSearch = CardTail("tempt-with-discovery/caster-search", func(ctx *Context, o game.TailOutcome) error {
		return b07TemptWithDiscoveryCasterSearch(ctx, o.Tail.Carry)
	})
	b07TemptNextOffer = CardTail("tempt-with-discovery/next-offer", func(ctx *Context, o game.TailOutcome) error {
		return b07TemptWithDiscoveryOffer(ctx, o.Tail.Carry)
	})
}

// b07TemptWithDiscoveryCasterSearch is the caster's own "search your
// library for a land card and put it onto the battlefield" — the
// opening one and every one an opponent's acceptance earns. Not
// optional and untapped, as printed. The offer continues at the opponents
// still to ask (`rest`) once this search has settled.
func b07TemptWithDiscoveryCasterSearch(ctx *Context, rest []uuid.UUID) error {
	next := ctx.ItemTail(b07TemptNextOffer)
	next.Carry = rest
	return ctx.Game.SearchLibraryThenForEffect(game.SearchLibrarySpec{
		Player:   ctx.Controller(),
		Source:   ctx.Source(),
		Pred:     b03IsLandCard,
		Dest:     game.ZoneBattlefield,
		Limit:    1,
		Shuffle:  true,
		Reason:   "Tempt with Discovery — a land onto the battlefield",
		ThenTail: next,
	})
}
