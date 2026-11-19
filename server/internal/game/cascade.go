package game

import (
	"github.com/google/uuid"
)

// cascade.go — S28: CR 702.85.
//
//	"Cascade (When you cast this spell, exile cards from the top of
//	 your library until you exile a nonland card that costs less. You
//	 may cast it without paying its mana cost. Put the exiled cards on
//	 the bottom of your library in a random order.)"
//
// Three engine facts the keyword needs and the engine did not have:
//
//  1. **A triggered ability of a SPELL.** Cascade triggers when the
//     spell is cast, which is to say while it is on the stack and
//     nowhere near the battlefield the trigger harvester scans. That
//     is why TriggeredAbility grew a FromStack flag and the harvester
//     grew a narrow cast-only stack scan (triggers.go) rather than a
//     blanket one: a permanent's ETB-watching trigger must not start
//     firing while its own spell is still on the stack.
//
//  2. **Reproducible randomness.** "In a random order" has to come
//     out the same after an undo or a restore, so the order draws from
//     the game's keyed RNG (rng.go, ADR 0054) on the owner's
//     "random_order" stream. Reaching for math/rand directly would
//     make it neither rewindable nor persistable.
//
//  3. **A yes/no during a resolution.** "You MAY cast it" is a
//     decision taken while the trigger is resolving, which is the
//     PendingChoice machinery's job (PendingChoiceMayCast).
//
// THE CAST happens during the trigger's resolution (CR 608.2g, #4243).
// CR 702.85a: "You may cast it without paying its mana cost", and CR
// 608.2g makes that a cast made while the cascade trigger is still
// resolving: no player receives priority after it, the spell becomes the
// topmost object on the stack above the trigger's controller's own
// spell, and the trigger carries on resolving. So the hit is not offered
// with a yes/no and cast later: a resolution_cast prompt
// (resolution_copy_cast.go, ADR 0066's 2026-10-22 and 2026-10-23
// amendments) holds the resolution open while the caster casts the hit,
// with its targets, modes and X chosen then, or declines. The prompt is
// plain data (the hit, and the cards cascade passed over in
// PendingChoice.CastPile), so a table paused on it is a restore point.
//
// Under the prompt the hit still carries the pass-closed permission
// ADR 0099 gave it, so the price, the timing, the mana value cap and
// the way an uncast hit goes to the bottom are all one definition. A
// cast consumes the permission; a decline lapses it, and the hit goes to
// the bottom of the library in the same random order as the cards
// cascade passed over (CR 702.85a: "put all cards exiled this way that
// weren't cast on the bottom of your library in a random order").
//
// Outside an open resolution (nothing calls cascade from one, but the
// engine does not assume it) there is nothing to hold open: the
// permission alone is the window, closing on the caster's next priority
// pass as it did before, and the passed-over cards are bottomed at once.
//
// ADR 0099 gave the permission the printed card's shape in three ways:
//
//   - TimingFlash (CR 608.2g). A cascaded sorcery is castable in combat
//     or on another player's turn. Before, it obeyed its own timing and
//     was simply lost there — weaker than printed.
//   - MaxSpellManaValue, one less than the cascading spell's. "If the
//     resulting spell's mana value is less than this spell's mana
//     value" (CR 702.85a) refuses a modal DFC's expensive back face or
//     an adventure half that costs more, which the grant used to allow.
//   - LapseOnPass. The window closes on the caster's NEXT priority pass
//     and the uncast hit goes to the bottom of the library then. It used
//     to stay castable until the end step and be bottomed by a delayed
//     trigger, which let a player hold a free spell for the rest of the
//     turn — stronger than printed.

// CascadeForEffect performs one cascade for `controller` off a spell
// with mana value `lessThan`, per CR 702.85a. `source` is the
// cascading spell, for attribution on the prompt and the event log.
//
// Returns nil in every "nothing happened" case — an empty library, a
// library with no qualifying card — because cascade does as much as
// it can (CR 608.2c) and a deck with nothing cheap in it is a legal
// board state, not an error.
//
// Caller must hold g.mu (this runs from a resolving trigger's
// Effect).
func (g *Game) CascadeForEffect(controller, source uuid.UUID, lessThan int) error {
	pile, hit, err := g.exileUntilLocked(controller, func(c Card) bool { return cascadeHit(c, lessThan) })
	if err != nil {
		return err
	}
	// "As you cascade, you may put a land card from among the exiled cards
	// onto the battlefield tapped" (Averna, the Chaos Bloom): after the last
	// card is exiled and before the hit is offered. A land is never the hit,
	// so the lands are in the pile.
	if g.playerInGameLocked(controller) && g.asYouCascadePutsALandLocked(controller) {
		var lands, rest []uuid.UUID
		for _, id := range pile {
			if c, ok := g.cardInZoneLocked(g.Exile, id); ok && c.IsLand() {
				lands = append(lands, id)
			} else {
				rest = append(rest, id)
			}
		}
		if len(lands) > 0 {
			carry := append([]uuid.UUID{hit}, rest...)
			if g.queueRevealPickKeyedLocked(RevealPickPrompt{
				Chooser:  controller,
				Owner:    controller,
				Source:   source,
				Question: "Cascade — you may put a land card from among the exiled cards onto the battlefield tapped",
				Cards:    lands,
				Min:      0,
				Max:      1,
				ThenKey:  cascadeLandThen,
				Carry:    carry,
				Nums:     []int{lessThan},
			}) > 0 {
				return nil
			}
		}
	}
	return g.cascadeFinishLocked(controller, source, lessThan, pile, hit)
}

// CatalogAsYouCascadePutsALand is the catalog hook for a permanent's "As you
// cascade, you may put a land card from among the exiled cards onto the
// battlefield tapped" (Averna, the Chaos Bloom), wired at init through
// CardDef.AsYouCascadePutALand. Nil, or false, means the permanent grants
// nothing.
var CatalogAsYouCascadePutsALand func(key string) bool

// asYouCascadePutsALandLocked reports whether a permanent `controller`
// controls says it, read live off the battlefield and stored nowhere
// (keyed by CatalogAbilityKey, so a permanent that lost its abilities
// grants nothing). Caller must hold g.mu.
func (g *Game) asYouCascadePutsALandLocked(controller uuid.UUID) bool {
	if CatalogAsYouCascadePutsALand == nil || g.Battlefield == nil {
		return false
	}
	for i := range g.Battlefield.Cards {
		c := &g.Battlefield.Cards[i]
		if c.Controller != controller {
			continue
		}
		if key := catalogAbilityKeyOf(c); key != "" && CatalogAsYouCascadePutsALand(key) {
			return true
		}
	}
	return false
}

// cascadeLandThen is the answer to the land prompt: the chosen land (at most
// one) enters tapped, and the cascade carries on once its entry has settled
// (a land may ask its own question first). The prompt's Carry is the hit
// (uuid.Nil when the library ran out) then the non-land cards passed over;
// the lands not chosen rejoin the pile. An on-disk key: never renamed,
// never reused.
var cascadeLandThen = RegisterChooseCardsThen("choose-cards/cascade-put-a-land", func(g *Game, r CardsChosen) error {
	if len(r.Carry) == 0 || len(r.Nums) == 0 {
		return nil
	}
	hit := r.Carry[0]
	pile := append([]uuid.UUID(nil), r.Carry[1:]...)
	var picked uuid.UUID
	if r.Answered && len(r.Picked) > 0 {
		picked = r.Picked[0]
	}
	for _, id := range r.Candidates {
		if id != picked {
			pile = append(pile, id)
		}
	}
	finish := KeyedTail{
		Ref:    cascadeFinishTail,
		Player: r.Chooser,
		Source: r.Source,
		Carry:  append([]uuid.UUID{hit}, pile...),
		N:      r.Nums[0],
	}
	if picked == uuid.Nil {
		return finish.Run(g, nil)
	}
	return g.PutOntoBattlefieldTogetherKeyedForEffect(
		[]BatchEntry{{CardID: picked, From: ZoneExile}},
		ZoneEntryOptions{Controller: r.Chooser, Tapped: true}, finish)
})

// cascadeFinishTail is the rest of the cascade after the land step.
var cascadeFinishTail = RegisterTail("then/cascade-finish", func(g *Game, o TailOutcome) error {
	t := o.Tail
	if len(t.Carry) == 0 {
		return nil
	}
	return g.cascadeFinishLocked(t.Player, t.Source, t.N, t.Carry[1:], t.Carry[0])
})

// cascadeFinishLocked is the hit and the bottoming: everything CR 702.85a
// says after the exiling. `hit` is uuid.Nil when no card qualified.
func (g *Game) cascadeFinishLocked(controller, source uuid.UUID, lessThan int, pile []uuid.UUID, hit uuid.UUID) error {
	if hit == uuid.Nil {
		// Ran the library out without finding anything. Everything
		// exiled goes back to the bottom; the deck is reordered but
		// not lost.
		return g.PutOnBottomInRandomOrderForEffect(controller, ZoneExile, pile)
	}
	// The pile and the hit are IDs saved before the prompt. Every
	// bottom below names ZoneExile, so a card that left exile while
	// the question was open stays where it went rather than being
	// pulled back.
	if !g.playerInGameLocked(controller) {
		// A caster who has left cannot cast it: the hit joins the pile.
		return g.PutOnBottomInRandomOrderForEffect(controller, ZoneExile, append(pile, hit))
	}
	g.grantFreeCastLocked(controller, hit, lessThan)
	if _, _, ok := g.ResolutionCastCandidateLocked(controller, hit); !ok {
		return g.PutOnBottomInRandomOrderForEffect(controller, ZoneExile, append(pile, hit))
	}
	// "You MAY cast it." Declining is a real choice — a cascade into
	// a card you do not want cast is a decision, not a formality.
	if g.queueResolutionCastOfferLocked(resolutionCastOffer{
		Chooser:      controller,
		Source:       source,
		Reason:       "Cascade",
		Candidates:   []uuid.UUID{hit},
		Keyword:      MayCastKeywordCascade,
		AcceptLabel:  "Cast it free",
		DeclineLabel: "Put it on the bottom",
		Pile:         pile,
	}) != uuid.Nil {
		return nil
	}
	// Outside a resolution: the permission is the window, and the
	// cards passed over go to the bottom now.
	return g.PutOnBottomInRandomOrderForEffect(controller, ZoneExile, pile)
}

// cascadeHit reports whether an exiled card stops the cascade: a
// nonland card whose mana value is strictly less than the cascading
// spell's (CR 702.85a).
//
// A card whose printed cost the engine cannot READ is deliberately
// not a hit. Split and adventure cards import a joined cost
// ("{1}{R} // {1}{U}") that ParseCost rejects and Card.ManaValue
// reports as zero — which would make every one of them a hit for any
// cascade, and hand the player a free cast of a card whose real mana
// value might be higher than the cascading spell's. Skipping them
// keeps cascade weaker than printed (it may pass over a card it
// should have stopped on) rather than stronger, which is the side to
// err on. The loop still terminates: worst case it exiles the
// library, which is a legal cascade outcome.
func cascadeHit(c Card, lessThan int) bool {
	if c.IsLand() {
		return false
	}
	if _, err := ParseCost(c.ManaCost); err != nil {
		return false
	}
	return c.ManaValue() < lessThan
}

// grantFreeCastLocked stamps the cascade hit's free cast.
//
// The price is spelled "{0}" rather than left empty because
// CastPermission.Cost treats the empty string as "no override, pay
// the printed cost" — which for a cascade hit would be the exact
// opposite of what the keyword grants. "{0}" parses to the zero cost,
// so the cast is free, and it locks X at 0 (CR 107.3b).
//
// CastOnly, because cascade says "you may CAST it". TimingFlash,
// because the cast happens during the trigger's resolution
// (CR 608.2g). MaxSpellManaValue is lessThan-1: the resulting spell's
// mana value must be LESS than the cascading spell's (CR 702.85a).
// LapseOnPass closes the window on the caster's next pass and bottoms
// the card then (ADR 0099 §4).
//
// Caller must hold g.mu.
func (g *Game) grantFreeCastLocked(controller, cardID uuid.UUID, lessThan int) {
	c, ok := g.cardInZoneLocked(g.Exile, cardID)
	if !ok {
		return
	}
	limit := lessThan - 1
	g.GrantCastPermissionToCardsForEffect(CastPermission{
		Player:            controller,
		Zone:              ZoneExile,
		Duration:          g.UntilEndOfTurnDuration(),
		Cost:              "{0}",
		Timing:            TimingFlash,
		CastOnly:          true,
		MaxSpellManaValue: &limit,
		LapseOnPass:       LapseToLibraryBottom,
		Label:             "Cascade — cast it without paying its mana cost",
	}, []Card{c})
}

// cascadeBottomBody is the end-step delayed trigger cascade scheduled
// before ADR 0099: at the beginning of the next end step, a hit still
// in exile under its free-cast grant went to the bottom of its owner's
// library. Nothing schedules it any more — the grant's window closes
// on the caster's next pass instead (LapseOnPass) — but the key stays
// registered, because a restore point written by an older binary may
// hold one, and an effect key is never deleted (ADR 0041 phase 3).
//
// The body is registered in init rather than by a var initialiser: it
// reaches the exit primitives, which reach the scheduler, and an
// initialiser would be an initialisation cycle. The card rides on the
// item's Targets and the controller is the item's, so it reads no
// params.
func init() { SimpleDelayedBody("cascade/bottom-if-not-cast", cascadeBottom) }

func cascadeBottom(g *Game, item *StackItem) error {
	controller := item.Controller
	for _, t := range item.Targets {
		if t.Kind != TargetCard || t.ID == uuid.Nil {
			continue
		}
		cardID := t.ID
		c, ok := g.cardInZoneLocked(g.Exile, cardID)
		if !ok {
			continue
		}
		perm := g.CastPermissionForLocked(controller, c, ZoneExile)
		if perm == nil || perm.Cost != "{0}" {
			continue
		}
		// The shared random bottom (random_bottom.go) with a pile
		// of one: it routes through the exit path, so the card
		// goes to its OWNER's library and a commander's owner is
		// asked about the command zone.
		if err := g.PutOnBottomInRandomOrderForEffect(controller, ZoneExile, []uuid.UUID{cardID}); err != nil {
			return err
		}
	}
	return nil
}

// cardInZoneLocked returns a copy of a card in the given zone.
// Caller must hold g.mu.
func (g *Game) cardInZoneLocked(z *Zone, id uuid.UUID) (Card, bool) {
	if z == nil {
		return Card{}, false
	}
	for i := range z.Cards {
		if z.Cards[i].InstanceID == id {
			return z.Cards[i], true
		}
	}
	return Card{}, false
}
