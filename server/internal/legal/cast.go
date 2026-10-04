package legal

import (
	"fmt"
	"sort"
	"strings"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// castParams is the cast_spell wire payload this package emits. Every
// cast is sent strict + auto_tap: the bot pays for what it casts, and
// affordability is decided here so the engine never has to reject.
type castParams struct {
	InstanceID string `json:"instance_id"`
	FromZone   string `json:"from_zone,omitempty"`
	// AlternativeCost is the key the cast claims (CR 118.9) — the
	// offer a granted permission synthesises for a non-hand cast
	// (ADR 0066), since a zone a permission prices cannot be cast
	// from without paying that price.
	AlternativeCost string `json:"alternative_cost,omitempty"`
	// AltCostIDs are the cards paid to the NON-MANA half of the
	// claimed cost (CR 601.2b) — Force of Will's pitched blue card,
	// Daze's Island, escape's N other cards from the graveyard.
	// Exactly CardPaymentCount entries, or none.
	AltCostIDs []string `json:"alt_cost_ids,omitempty"`
	// PhyrexianLife is how many of the cost's Phyrexian symbols the
	// cast pays with 2 life each instead of mana (CR 107.4c, CR
	// 601.2b, #1677) — CastSpellParams.PhyrexianLife. See
	// castPayment for which counts are offered.
	PhyrexianLife int          `json:"phyrexian_life,omitempty"`
	Targets       []targetWire `json:"targets,omitempty"`
	Modes         []int        `json:"modes,omitempty"`
	XValue        int          `json:"x_value,omitempty"`
	DiscardIDs    []string     `json:"discard_ids,omitempty"`
	SacrificeIDs  []string     `json:"sacrifice_ids,omitempty"`
	// OptionalCosts are the optional additional costs this move pays
	// (CR 601.2b, ADR 0073), as positions in the card's OptionalCosts
	// slice, repeated once per payment for a multikicker.
	OptionalCosts []int `json:"optional_costs,omitempty"`
	// CostBranch is the either/or additional cost's branch this move
	// pays (CR 601.2b, ADR 0100 §6) — CastSpellParams.CostBranch. Nil
	// for every card without one.
	CostBranch *int `json:"cost_branch,omitempty"`
	// TeamworkIDs / BlightIDs pay an announced teamwork or blight cost
	// (#1703) — CastSpellParams.TeamworkIDs / BlightIDs.
	TeamworkIDs []string `json:"teamwork_ids,omitempty"`
	BlightIDs   []string `json:"blight_ids,omitempty"`
	// DelveIDs are the graveyard cards exiled to delve (CR 702.66a,
	// ADR 0100 §6) — CastSpellParams.DelveIDs.
	DelveIDs []string `json:"delve_ids,omitempty"`
	// Distribution is the division of a "divided as you choose"
	// clause (#1563, CR 601.2d), target id → share. The enumerator
	// always announces game.EvenDistribution — the amount split as
	// evenly as possible, the remainder to the earliest targets — so
	// every divided move it offers is one the gate accepts.
	Distribution map[string]int `json:"distribution,omitempty"`
	// GiftOpponent is the opponent a gift is promised to (CR
	// 702.174a, ADR 0089) — set exactly when OptionalCosts names the
	// card's gift cost.
	GiftOpponent string `json:"gift_opponent,omitempty"`
	Strict       bool   `json:"strict,omitempty"`
	AutoTap      bool   `json:"auto_tap,omitempty"`
	// Face is the printed face being cast or played (ADR 0034).
	// Omitted — the front — for every single-faced card.
	Face int `json:"face,omitempty"`
	// Fuse casts both halves of a split card with fuse from hand
	// (CR 702.102a, ADR 0103) — CastSpellParams.Fuse.
	Fuse bool `json:"fuse,omitempty"`
}

// castZone is one pile the walk below looks in. `mine` says the pile
// belongs to the seat being enumerated, which is what decides whether
// the card's OWN text can open it: flashback, escape and Gravecrawler
// all say "your graveyard", so in somebody else's a permission is the
// only key (#1022). A library is the same story with CR 401.5's "your
// library" (#1035), and nothing prints a cast surface on a library at
// all, so there `mine` only ever narrows.
type castZone struct {
	z    *game.Zone
	kind game.ZoneKind
	from string
	mine bool
}

// castZones are the places a cast or a land play can come out of, in
// the order the moves are emitted. ONE walk since #673: the zone a
// card is in decides which surfaces have to be checked, never which
// kinds of cost are available, because "where may I cast this from"
// and "what may I pay for it" are separate questions the engine
// answers in separate functions (cast_zones.go).
//
// Every seat's graveyard and library top is walked rather than only
// this one's, and only when something has granted something
// (`anyGrant`): ADR 0066 makes a permission a statement about an
// OBJECT, and CastSpell finds the card in the pile it is actually in
// (#1022, #1035, castSourceZoneLocked). A four-player table with no
// grants in play walks exactly the piles it always did, and one with
// grants pays three extra piles plus three extra CARDS — the loop
// below trims every library to its top (CR 401.5) before it looks at
// anything.
func (e *enumerator) castZones(anyGrant bool) []castZone {
	g, p := e.g, e.p
	out := []castZone{
		{p.Hand, game.ZoneHand, "hand", true},
		{p.Command, game.ZoneCommand, "command", true},
		{p.Graveyard, game.ZoneGraveyard, "graveyard", true},
		{g.Exile, game.ZoneExile, "exile", true},
		{p.Library, game.ZoneLibrary, "library", true},
	}
	if !anyGrant {
		return out
	}
	for _, other := range g.Seats {
		if other == nil || other.ID == e.seat {
			continue
		}
		out = append(out,
			castZone{other.Graveyard, game.ZoneGraveyard, "graveyard", false},
			castZone{other.Library, game.ZoneLibrary, "library", false})
	}
	return out
}

// castMoves enumerates land plays and spell casts from EVERY zone the
// seat can cast out of — hand, the command zone, a graveyard, exile
// and the top of a library — at every price the engine would accept
// for each (#673).
//
// Two questions, two engine functions, and the whole of this method is
// asking them in a loop:
//
//   - MAY this card be cast from this zone: the card's own
//     declaration (CastableZonesFor — flashback, escape, Gravecrawler),
//     or a granted CastPermission (ADR 0066 — Snapcaster, cascade,
//     impulse exile, warp's recast, a foretold or suspended card).
//   - AT WHAT PRICE: game.CastOffersForLocked, which lists the printed
//     mana cost and every alternative cost claimable from that zone,
//     already filtered to the ones payable right now.
//
// Both are the functions CastSpell validates with, so a bot can never
// be offered a cast the engine will refuse, at a zone it will refuse,
// or at a price it will not charge.
func (e *enumerator) castMoves() {
	g := e.g
	if g.SplitSecondActive {
		return
	}
	// CR 305.1's land window, asked of the engine's one sorcery-timing
	// read rather than a copy of it (#1352: the copy that used to live
	// here was right about abilities on the stack while the engine was
	// not, and a copy that is right today is the one that drifts).
	speed := g.SorcerySpeedOpenLocked(e.seat)
	// #500: the allowance is the player's, not a literal one — a
	// controlled Exploration or a one-turn grant raises it. Same
	// helper the engine's own refusal reads, so the enumerator can
	// never offer a land play CastSpell will reject.
	landOwed := g.LandDropsRemainingLocked(e.seat) > 0
	// The fast negative, and the same one the view takes: in almost
	// every game nothing grants anything. It no longer gates the
	// GRAVEYARD walk, because a card's own flashback or escape opens
	// that zone with nothing granted — but exile and the library are
	// reachable only through a grant (cast_zones.go), so skipping
	// them costs nothing and an enumeration runs on every decision.
	anyGrant := g.AnyCastPermissionsForEffect()

	for _, zone := range e.castZones(anyGrant) {
		if zone.z == nil {
			continue
		}
		cards := zone.z.Cards
		switch zone.kind {
		case game.ZoneExile:
			if !anyGrant {
				continue
			}
		case game.ZoneLibrary:
			// CR 401.5: only the top card is ever open, and the top is
			// the LAST element. Checking one card rather than walking
			// the library also keeps this loop from touching hidden
			// information it has no business reading.
			if !anyGrant || len(cards) == 0 {
				continue
			}
			cards = cards[len(cards)-1:]
		}
		for _, c := range cards {
			e.castMovesFromZone(c, zone.kind, zone.from, speed, landOwed, anyGrant, zone.mine)
		}
	}
}

// castMovesFromZone expands ONE card sitting in ONE zone: its faces,
// its land play, and one call into castMovesForCard per price the
// engine would let this seat announce.
//
// `mine` is false for another seat's graveyard (#1022) or library
// (#1035), where the card's own declaration opens nothing and a
// permission is the only way in.
func (e *enumerator) castMovesFromZone(c game.Card, kind game.ZoneKind, from string, speed, landOwed, anyGrant, mine bool) {
	g := e.g
	// CastPermissionForLocked answers nil unless the window is open
	// for this seat, so there is no second liveness test here — one
	// function reads the duration (#945).
	var perm *game.CastPermission
	if anyGrant {
		perm = g.CastPermissionForLocked(e.seat, c, kind)
	}
	// The per-card fast negative for the three zones that are a cast
	// surface only sometimes. CastOffersForLocked would answer with an
	// empty list anyway — this just declines to build one for every
	// card in a thirty-card graveyard on every bot decision.
	switch kind {
	case game.ZoneGraveyard, game.ZoneExile, game.ZoneLibrary:
		// A card's own declaration opens ITS OWNER's zone and nobody
		// else's: flashback, escape and Gravecrawler all print "your
		// graveyard" (#1022), and nothing prints a library cast at all
		// (#1035). In another seat's pile the permission is the whole
		// answer, which is also what CastSpell's zone lookup enforces
		// — castSourceZoneLocked finds the card there only under one.
		// #1171: game.CardCastableFromAnyFace, which is the same
		// predicate the VIEW asks now. One question per face rather
		// than one for the card — an MDFC's halves are separate
		// catalog entries and only the back may print flashback — and
		// until #1171 the view asked it of face 0 alone, so a card
		// this loop offered a bot reached the frame with no announce
		// stamps at all.
		if perm == nil && !(mine && game.CardCastableFromAnyFace(c, kind)) {
			return
		}
	}
	// ADR 0034: a modal DFC is two playable objects sharing one
	// instance, and since #719 so is an ADVENTURE card — CR 715.3 lets
	// the caster choose the creature or the Adventure — so enumerate
	// each face as its own move and let the bot pick between them.
	// CastableFaces returns [0] for everything else, so this loop runs
	// once for every single-faced card and the enumeration is
	// unchanged for them.
	//
	// A grant that NAMES faces NARROWS the choice to exactly those (a
	// defeated Siege's back, CR 715.4's "cast the creature from
	// exile"), which is the same rule faceForCastLocked applies — so
	// the enumerator cannot offer a half the announce path refuses.
	//
	// #992: through game.CastableFacesUnder, which is the same pair of
	// rules the VIEW now walks to publish per-face announce data. Two
	// copies of "the card's layout, unless a grant names faces" is a
	// picker row the enumerator would not offer, or the other way
	// round.
	for _, face := range game.CastableFacesUnder(c, perm, e.seat) {
		// The face is materialised onto a COPY, exactly as CastSpell
		// does, so all the type, cost and catalog reads below see the
		// chosen half without any of them learning about faces.
		card := c
		card.SetFace(face)
		if card.IsLand() {
			e.landPlayMove(card, kind, from, speed, landOwed, perm)
			continue
		}
		for _, offer := range g.CastOffersForLocked(e.seat, card, kind, perm) {
			// The bot POLICY on top of the rule, and the only thing
			// this package adds to the engine's answer: CR 119.4 lets
			// a player pay life down to exactly zero, the next
			// state-based check then kills them, and a bot offered
			// that line would take it.
			if offer != nil && offer.Life > 0 && e.p.Life <= offer.Life {
				continue
			}
			e.castMovesForCard(card, from, kind, perm, offer)
		}
	}
	// ADR 0103, CR 702.102a: a split card with fuse in hand may also be
	// cast as both halves, for both costs, with no alternative cost.
	if kind == game.ZoneHand && perm == nil && game.HasFuse(c) && !game.FusedHalvesDeclareExtras(c.OracleID) {
		fused := game.FusedSpell(c)
		for _, offer := range g.CastOffersForLocked(e.seat, fused, kind, perm) {
			if offer == nil {
				e.castMovesForCard(fused, from, kind, perm, nil)
			}
		}
	}
}

// landPlayMove emits the one move for playing a land out of `kind`,
// when CR 305 allows it.
func (e *enumerator) landPlayMove(card game.Card, kind game.ZoneKind, from string, speed, landOwed bool, perm *game.CastPermission) {
	// CR 305: main phase, empty stack, your turn, and the per-turn
	// land-play allowance. The engine enforces all four since #500;
	// the check stays here so a bot is never OFFERED a move that would
	// be refused.
	if !speed || !landOwed {
		return
	}
	// CR 903.4 is a permission to CAST a commander, not to play a land
	// out of the command zone, and no other rule opens that door.
	if kind == game.ZoneCommand {
		return
	}
	// CR 305.1: playing a land is not casting, so a cast-only
	// permission strands it.
	if perm != nil && perm.CastOnly {
		return
	}
	// ADR 0109 §4, CR 101.2: "can't" beats "can". The engine's own gate,
	// asked before the drop count in castSpellLocked, so a bot is never
	// offered a land a "players can't play lands" effect refuses.
	if e.g.LandPlayGateLocked(e.seat, card, kind) != nil {
		return
	}
	label := "Play " + card.Name
	if kind != game.ZoneHand {
		label += " from " + from
	}
	e.add(Move{
		Type:   TypeCastSpell,
		Player: e.seat,
		Kind:   KindLand,
		Label:  label,
		Source: card.InstanceID,
		Params: mustJSON(castParams{
			InstanceID: card.InstanceID.String(),
			FromZone:   from,
			Face:       card.ActiveFace,
		}),
	})
}

// castMovesForCard expands one non-land card at ONE announced price
// into concrete casts: every legal (modes × targets × cost payment)
// combination the seat can afford, capped at MaxExpansionPerSource.
//
// `offer` is the CR 118.9 cost this expansion pays — nil for the
// printed mana cost, otherwise one of the entries
// game.CastOffersForLocked listed for this card in this zone. The
// caller loops the offers; each is its own set of moves, because a
// flashed-back Faithless Looting and a hard-cast one are different
// prices with different consequences, exactly as a kicked and an
// unkicked cast are (ADR 0073 §9).
func (e *enumerator) castMovesForCard(card game.Card, from string, kind game.ZoneKind, perm *game.CastPermission, offer *game.AlternativeCost) {
	// CR 708.4, ADR 0082 decision 2: a cast that claims a face-down
	// offer announces a 2/2 CREATURE SPELL with no name and no text,
	// and the engine stamps that onto its own copy of the card before
	// any announce gate reads it. The enumerator's walk is the same
	// walk and has to see the same object, or it offers a morph the
	// modes, targets and optional costs of the card underneath —
	// every one of which CastSpell would then reject.
	//
	// One line, and the timing falls out of it: the stamped copy is a
	// creature that is not an instant and has no flash, so the
	// sorcery-speed gate below demands a sorcery-speed window for a
	// face-down cast without a rule of its own.
	if offer != nil && offer.FaceDown != nil {
		card.SetFaceDown(offer.FaceDown.Kind)
	}
	// ADR 0107 §4, CR 712.11a: the other offer that changes what the
	// spell IS. A disturb cast puts the card on the stack back face up,
	// and CastSpell turns its own copy over before the target, timing
	// and cast gates read it, so this walk reads the back face too —
	// Spectral Binding's "enchant creature", not Binding Geist's empty
	// clause. The move's Face is the back face, which CastSpell accepts
	// for this claim.
	card = offer.CastFaceOf(card)
	// #1665: a HAND permission (miracle) applies to the one offer it
	// opens and to nothing else — the same narrowing CastSpell makes
	// once the claim is known, so the miracle cast is offered at
	// instant speed and the printed cast only where a sorcery could be
	// cast. A no-op for every other zone. See game.CastPermission.ForClaim.
	perm = perm.ForClaim(offer)
	// Timing (CR 307.1 / 702.8), #1195: THE engine's own read, the
	// same one CastSpell calls and the view stamps `castable_here`
	// from — so a bot is never offered a cast the engine will refuse
	// for timing, and never denied one a Vedalken Orrery opens. It
	// folds the card's own timing, the permission's ADR 0066
	// override, the per-player grants and, last, the per-player
	// restrictions (CR 101.2).
	//
	// BELOW THE FACE-DOWN STAMP, and that placement is load-bearing:
	// CR 708.4 makes a face-down cast a 2/2 creature spell with no
	// name and no text whatever the card underneath is, so a morphed
	// INSTANT is sorcery-speed. Asked above the offer loop — where
	// #1195 first put it, before ADR 0082 existed — it would read the
	// face-up card and offer a bot a morph at instant speed that
	// CastSpell refuses. The engine reads the same stamped copy at
	// the same point, which is what keeps the two answers identical.
	//
	// Once per (card, face, offer) rather than once per announced set
	// of optional costs: nothing in ADR 0073's optional half can
	// change a timing answer.
	if !e.g.CastTimingOpenLocked(e.seat, card, kind, perm) {
		return
	}
	// ADR 0073 §9: a card with optional additional costs is several
	// casts, not one — an unkicked Burst Lightning and a kicked one
	// are different moves at different prices with different effects,
	// and a bot that was only ever offered the cheap one would never
	// kick anything. Each announced set walks the whole expansion
	// below on its own.
	//
	// optionalCostSets is [nil] for every card that offers none,
	// which is every card in the catalog before #664 — so this loop
	// runs exactly once for them and the enumeration is unchanged.
	optional := game.OptionalCostsFor(game.CatalogKey(card))
	for _, chosen := range optionalCostSets(optional, e.opts.MaxExpansionPerSource) {
		// ADR 0089: a set that promises a gift is one announcement
		// per opponent who could receive it — "which opponent" is
		// part of paying the cost (CR 702.174a), and a bot offered
		// only the first seat could never choose to feed the player
		// who is furthest behind.
		for _, to := range giftRecipients(e.g, e.seat, optional, chosen) {
			// ADR 0100 §6: an either/or additional cost is one
			// announcement per branch the seat can pay — at most
			// three — each priced with its own cost_branch. [nil] for
			// every card without branches, so the loop runs once.
			for _, branch := range e.castCostBranches(card) {
				e.castMovesPayingOptional(card, from, perm, offer, chosen, to, branch)
			}
		}
	}
}

// giftWire is the wire form of a gift recipient: omitted for
// uuid.Nil, which is every cast that promises nothing.
func giftWire(id uuid.UUID) string {
	if id == uuid.Nil {
		return ""
	}
	return id.String()
}

// giftRecipients is the gift half of one announced optional-cost set:
// [uuid.Nil] when the set promises no gift, one entry per opponent
// who may receive it when it does, and nothing at all when a gift is
// announced and nobody is left to promise it to — that set is not a
// legal announcement (the engine's validateGiftChoiceLocked).
//
// The list is game.GiftOpponentsLocked verbatim, the one the engine
// and the view read, so a bot can never be offered a recipient the
// server refuses.
func giftRecipients(g *game.Game, seat uuid.UUID, optional []game.AdditionalCost, chosen []int) []uuid.UUID {
	gift := false
	for _, i := range chosen {
		if i >= 0 && i < len(optional) && optional[i].ChoosesOpponent {
			gift = true
			break
		}
	}
	if !gift {
		return []uuid.UUID{uuid.Nil}
	}
	return g.GiftOpponentsLocked(seat)
}

// maxEnumeratedCostPayments caps how many ways the enumerator will
// offer to pay the CARD-shaped half of one alternative cost — which
// blue card Force of Will pitches, which five cards an Uro escapes
// with. Not a rule; a policy, documented in docs/bot.md beside
// maxEnumeratedRepeats.
//
// THREE since #1013, and it was ONE before, for ADR 0033 §1's
// corollary rather than squeamishness about combinatorics: a variable
// in a COST must not become an arity of the target/mode cross product.
// Escape-five over a twenty-card graveyard is 15,504 payments, every
// one of them the same spell with the same targets, so a cap of twelve
// spent on payments would never offer the second target of anything.
//
// Two things changed and the corollary still holds.
//
// The payments are no longer INDISTINGUISHABLE. That was a statement
// about the POLICY rather than about Magic — the heuristic priced the
// battlefield and the seats and a card in a graveyard was worth
// nothing to it, so it could only pick between payments by index, and
// an Uro over a graveyard holding a second Uro, a Snapcaster target
// and three lands ate whichever three were oldest. A policy that
// implements aiseat.CostFuelPricer now prices a card as FUEL, and
// cheapestFuelFirst sorts the pool by it, so payment number one is the
// best one the policy can name.
//
// And the extra payments spend NO TARGET BUDGET. They are offered in a
// second pass, out of whatever MaxExpansionPerSource the target walk
// did not use, against the FIRST announcement it made — the same spell
// with the same targets at a different price, which is what they all
// are. An Uro with no targets has eleven unspent and gets its
// alternatives; a removal spell with an escape cost over a wide board
// spends its budget on targets and gets none, which is the corollary
// enforced rather than restated.
const maxEnumeratedCostPayments = 3

// maxEnumeratedRepeats caps how many times the enumerator will offer
// to pay one REPEATABLE optional cost (multikicker) in a single
// announcement. Not a rule — multikicker is unbounded in paper — but
// the expansion here is already modes × targets × cost payments, and
// multiplying it by the seat's available mana is how a bot decision
// stops terminating. Documented as policy in docs/bot.md.
const maxEnumeratedRepeats = 3

// optionalCostSets is the bot's announcement policy for optional
// additional costs (ADR 0073 §9): decline everything, or pay exactly
// ONE of the offered costs — once, or up to maxEnumeratedRepeats
// times for a repeatable one.
//
// NOT enumerated, deliberately and for the reason escape's card
// component is not: paying TWO different optional costs at once
// (Thornscape Battlemage's "Kicker {R} and/or {W}") is a power-set
// search whose every member also needs its own affordability probe,
// and no card in the catalog offers two. A bot simply does not take
// that line yet; it is never offered one it cannot pay for.
func optionalCostSets(costs []game.AdditionalCost, budget int) [][]int {
	if len(costs) == 0 {
		return [][]int{nil}
	}
	out := [][]int{nil}
	for i := range costs {
		reps := min(costs[i].MaxPayments(), maxEnumeratedRepeats)
		for k := 1; k <= reps; k++ {
			if budget > 0 && len(out) >= budget {
				return out
			}
			set := make([]int, k)
			for j := range set {
				set[j] = i
			}
			out = append(out, set)
		}
	}
	return out
}

// optionalCostLabel spells the announced optional costs into the
// move's label — " (Kicker {4})", " (Multikicker {G} ×3)" — so the
// kicked and unkicked casts of one card are distinguishable in the
// move log and in a bot-eval trace. Empty for a move that pays none.
func optionalCostLabel(optional []game.AdditionalCost, chosen []int) string {
	if len(chosen) == 0 {
		return ""
	}
	counts := make(map[int]int, len(chosen))
	for _, i := range chosen {
		counts[i]++
	}
	out := ""
	for i := range optional {
		n, ok := counts[i]
		if !ok {
			continue
		}
		if out != "" {
			out += ", "
		}
		out += optional[i].Label
		if n > 1 {
			out += fmt.Sprintf(" ×%d", n)
		}
	}
	if out == "" {
		return ""
	}
	return " (" + out + ")"
}

// costPaymentDemands sums what one cast owes in CARDS across its
// CR 601.2f payment plan (ADR 0073 §4): the mandatory additional cost
// plus each announced optional one. Reports the total discard count
// and the single sacrifice clause to expand.
//
// `ok` is false when the plan carries TWO sacrifice clauses, which
// this package does not enumerate — see the call site.
func costPaymentDemands(mandatory *game.AdditionalCost, optional []game.AdditionalCost, chosen []int) (discards int, sacrifice *game.TargetSpec, ok bool) {
	add := func(c *game.AdditionalCost) bool {
		if c == nil || c.Empty() {
			return true
		}
		discards += c.DiscardCards
		if c.Sacrifice == nil {
			return true
		}
		if sacrifice != nil {
			return false
		}
		sacrifice = c.Sacrifice
		return true
	}
	if !add(mandatory) {
		return 0, nil, false
	}
	for _, i := range chosen {
		if i < 0 || i >= len(optional) {
			continue
		}
		if !add(&optional[i]) {
			return 0, nil, false
		}
	}
	return discards, sacrifice, true
}

// castMovesPayingOptional is castMovesForCard for ONE announced set
// of optional additional costs — the unkicked cast, or the kicked
// one. `chosen` is nil for every card that offers none.
//
// `giftTo` is the opponent a gift is promised to when `chosen` names a
// gift cost, and uuid.Nil otherwise (ADR 0089).
//
// `branch` is the either/or branch this expansion pays (ADR 0100 §6),
// nil for a card without one.
func (e *enumerator) castMovesPayingOptional(card game.Card, from string, perm *game.CastPermission, offer *game.AlternativeCost, chosen []int, giftTo uuid.UUID, branch *int) {
	g, p := e.g, e.p
	// #662: the spell IS its own source (CR 702.16b), so every legal
	// set below is computed against the card's colour and type. An
	// enumerator that passed only the seat would offer the bot a
	// pro-red creature for its red spell and the server would refuse
	// the move — the #347 / #544 failure mode.
	castSrc := game.SourceObject(e.seat, &card)

	// Timing is NOT re-asked here. It was a copy of the engine's
	// three lines until #1195 — `!card.IsInstant() &&
	// !HasKeyword(&card, "flash")` plus the permission's override —
	// and it is now one call to game.CastTimingOpenLocked, made once
	// per (card, face) by castMovesFromZone above. Nothing about the
	// offer or the announced optional costs can change the answer.

	// The clause this cast announces under, which is the offer's
	// business as much as the card's: overload DELETES the target
	// clause ("change 'target' to 'each'") and cleave SWAPS it for a
	// wider one, and the engine reads the same function at announce
	// (CR 601.2c) and again at resolution. An enumerator that offered
	// an overloaded Cyclonic Rift with a target would have every one
	// of those moves refused with ErrInvalidParam.
	cardSpec := game.TargetSpecUnderAlternativeCost(game.TargetSpecFor(game.CatalogKey(card)), offer)
	// ADR 0089 §3: and a paid optional cost may swap it again — a
	// promised gift's "instead … target …" (CR 702.174m). The same
	// function, in the same order, the engine applies at announce.
	cardSpec = game.TargetSpecUnderOptionalCosts(cardSpec, game.OptionalCostsFor(game.CatalogKey(card)), chosen)

	// A card the catalog marks as targeted the S13.1 way (free-form
	// target_mode, no structured spec) cannot be enumerated: the
	// engine demands a target but nothing says which are legal. Under
	// an offer that clears the clause there is no target to demand,
	// so the cast is enumerable after all.
	if game.TargetModeFor(game.CatalogKey(card)) != "" && cardSpec == nil && !offer.Clears() {
		return
	}

	// Cost. A cost the parser can't read is not enumerable: since
	// #289 the engine rejects such a cast with ErrUnparseableCost
	// (split and adventure cards import a joined "{1}{R} // {1}{U}"),
	// so offering the move would hand the client an action that is
	// guaranteed to fail. Enumerating it free — what this used to
	// do, mirroring the engine's old silent downgrade — is worse:
	// it advertises a free spell that isn't one.
	//
	// #696: THE engine's pricer, called with the announcement this
	// move will actually send. It settles the CR 118.9 swap (the
	// permission's price, when it has one, IS the cost this cast pays
	// — ADR 0066), the commander tax, the mana half of the announced
	// optional costs (ADR 0073 §3) and the "spend mana as though any
	// colour" fold, in that order, so a bot is never offered a cast at
	// a price the engine will not charge. A copy of the first three
	// used to live here.
	//
	// #673: `offer` is the entry game.CastOffersForLocked listed for
	// this card in this zone — the card's own flashback or overload as
	// readily as the one a permission synthesises — and the pricer
	// takes it off the announcement below, so both kinds reach the
	// same walk.
	fromZone := game.ZoneHand
	switch from {
	case "command":
		fromZone = game.ZoneCommand
	case "graveyard":
		fromZone = game.ZoneGraveyard
	case "exile":
		fromZone = game.ZoneExile
	case "library":
		fromZone = game.ZoneLibrary
	}
	// The announcement, built once and reused for every repricing
	// below — the target-set loop reprices against it and changes
	// nothing but Targets.
	announce := game.CastSpellParams{
		FromZone:        from,
		AlternativeCost: offerKey(offer),
		OptionalCosts:   chosen,
		GiftOpponent:    giftTo,
		// ADR 0100 §2: the branch's mana joins the price here, in the
		// one pricer, so Lightning Axe's "pay {5}" move is priced at
		// {5}{R} exactly as CastSpell charges it.
		CostBranch: branch,
		// ADR 0034: `card` has already had SetFace applied by the
		// caller, so this is the face the move announces.
		Face: card.ActiveFace,
		// ADR 0103: and a fused copy announces the fused cast.
		Fuse: card.Fused,
	}
	price, err := e.g.PriceCastForEffect(e.seat, card, announce)
	if err != nil {
		return
	}
	// CastPrice.Base, not .Total: the convoke / waterbend subtraction
	// is a PAYMENT (CR 601.2h) and this package enumerates no tap
	// payments, so the cost it searches an X against must still carry
	// its {X} slot. Reading Total would settle X into generic and
	// Chord of Calling would only ever be offered at X=0.
	cost := price.Base
	// CR 118.6: a spell with no mana cost (Ancestral Vision, Living
	// End) can't be cast by paying it, and ParseCost reads that empty
	// string as a free {0}. Every move this function builds pays the
	// printed cost, so none of them is legal; the engine refuses the
	// cast with ErrNoManaCost.
	if game.HasNoManaCost(card) && offer == nil && (perm == nil || perm.Cost == "") {
		return
	}
	optional := game.OptionalCostsFor(game.CatalogKey(card))
	// S28: the board's cost modifiers (CR 601.2f). Same reasoning as
	// the parse gate above — a move enumerated at the printed price
	// while a Sphere of Resistance sits on the table is a move the
	// engine will reject for insufficient mana, and a bot that keeps
	// picking rejected moves stalls. A modifier the engine refuses to
	// price (ErrCostModifier) makes the cast unenumerable for the
	// same reason an unparseable cost does.
	//
	// Applied below rather than read off price.Total because this
	// package prices the SAME cast once per candidate target set (a
	// per-target surcharge, ADR 0048 addendum §14), and they all apply
	// to the one pre-modifier total above.
	//
	// Priced at X=0 even though affordableX is about to search for a
	// bigger X. The only modifier kind X can change the answer for is
	// a CostFloor (Trinisphere), and X=0 is the branch where the
	// floor applies — so the search starts from the most expensive
	// reading and can only narrow the X it offers. Conservative in
	// the direction that never advertises an unaffordable move.
	// #760, ADR 0073 §7: the one announce-time cast gate, the twin of
	// the split-second check at the top of castMoves and beside it
	// for the same reason — a bot offered a move the engine will
	// refuse keeps picking it and stalls. THE SAME FUNCTION CastSpell
	// calls, so the two cannot disagree about what is banned.
	//
	// The announced optional costs ride along because the gate takes
	// the announcement: CR 601.3a lets a choice made while proposing
	// the spell lift a ban, and this is the choice that has been made
	// by now.
	if err := g.CastGateLocked(e.seat, card, fromZone, announce); err != nil {
		return
	}
	// ADR 0048 addendum §14: when something on the board or the card
	// itself prices by target (Fireball's surcharge, Price of Fame's
	// discount), one price up front is not the price — and a
	// nil-targets price cannot even be used as a gate, because a
	// target-reading REDUCTION makes the real cast cheaper than it.
	// So the up-front price and its early return run only when
	// nothing reads targets, which is every board without such a
	// card; otherwise each (modes, targets) set below is priced on its
	// own and carries its own X.
	spend := game.ManaSpendForCast(card)
	perTarget := e.g.CastPriceReadsTargetsForEffect(card)
	// Additional costs (CR 601.2f). Read before the X search because
	// one of them can PRICE X: Toxic Deluge's "pay X life" is the
	// whole of its X, and the mana cost it prints has no {X} slot at
	// all (#957). The payments themselves are expanded below.
	//
	// ADR 0100 §2: for an either/or cost, the branch this expansion
	// announces — the one ChosenAdditionalCost hands CastSpell's plan.
	addCost, err := game.ChosenAdditionalCost(game.AdditionalCostFor(game.CatalogKey(card)), branch)
	if err != nil {
		return
	}
	// #810: the one X rule. A spell whose whole effect is X (Fireball,
	// Stroke of Genius) is not offered at X=0, where it would resolve
	// for nothing; a spell with a fixed rider still is. #957 is the
	// same rule reaching the non-mana half of the price: the floor is
	// the same one, and xCeilingFromCost is what a "pay X life" cost
	// can pay for (CR 119.4, less one so the seat survives its own
	// sweep). Both live in x.go.
	xFloor := enumeratedXFloor(game.CatalogKey(card), 0)
	xLifeCeiling := xCeilingFromCost(addCost, p.Life)
	// #1677: Phyrexian symbols paid with life (CR 107.4c). The offer's
	// own life (Force of Will's "pay 1 life") is held back so the two
	// together never claim more than the seat has (CR 119.4). A "pay X
	// life" additional cost turns the life path off: X and the symbol
	// count would both be priced out of one life total, no printed
	// card prints both, and the mana path is still offered.
	// ADR 0100 §2: a branch's fixed "pay 3 life" is held back the same
	// way.
	lifeReserved := offerLife(offer) + branchLife(addCost)
	lifeAllowed := addCost == nil || !addCost.PayLifeX

	// Modes → each choice of modes yields its own clause list, and
	// each clause its own picks (#764). Options with no legal target
	// are dropped before any combination is built, so the budget is
	// never spent on selections the engine would refuse (ADR 0065
	// §6).
	//
	// Computed BEFORE the price search below, which used to run once
	// for the whole card: ADR 0065's 2026-09-23 amendment lets a mode
	// carry its own cost (CR 702.172a, Spree), so — unlike every modal
	// card before it — the price this cast owes can depend on WHICH
	// modes are chosen, and the search has to run once per mode
	// selection rather than once for the card.
	modeSpec := game.ModeSpecFor(game.CatalogKey(card))
	modeSets := [][]int{nil}
	if modeSpec != nil {
		// #1655: under THIS announcement's optional costs — a kicked
		// Inscription of Ruin may take every bullet, an unkicked one
		// exactly one, and each is its own move.
		modeSets = e.legalModeSets(castSrc, modeSpec, game.ModeCountQuery{
			Chooser:       e.seat,
			OracleID:      game.CatalogKey(card),
			OptionalCosts: chosen,
		}, game.ModeAbility{})
		if len(modeSets) == 0 {
			return
		}
	}

	// The additional cost's PAYMENTS (CR 601.2f), read above. Discards
	// choose from the rest of the hand; a sacrifice chooses from the
	// seat's own permanents matching the clause. A pay-X-life needs no
	// payment set of its own — the announced X is the payment, and it
	// was priced with the rest of X above.
	//
	// ADR 0073: the demands are summed across the mandatory cost AND
	// whichever optional ones this move announces, so a kicked
	// Gatekeeper of Malakir expands its kicker's sacrifice exactly as
	// a mandatory one would. The wire lists are flat and the engine
	// walks them in the same order.
	discards, sacrifice, ok := costPaymentDemands(addCost, optional, chosen)
	if !ok {
		// Two card-shaped sacrifice clauses on one cast (a mandatory
		// one AND a kicker's). Not enumerated: the two pools have to
		// be searched together and written into one flat list, and no
		// card in the catalog asks for it. The bot does not take the
		// line; it is never offered one it cannot pay.
		return
	}
	// CR 601.2b's card component of the ALTERNATIVE cost, which is a
	// different list on the wire from the additional cost's (they are
	// paid at different steps and one of them vanishes when the offer
	// is declined — see CastSpellParams.AltCostIDs).
	//
	// The candidates come from the engine's own acceptance predicate,
	// so a payment this builds is a payment
	// validateAlternativeCostPaymentLocked accepts; and the whole
	// search collapses to ONE payment by policy — see
	// maxEnumeratedCostPayments.
	altCostSets := [][]uuid.UUID{nil}
	if want := offer.CardPaymentCount(); want > 0 {
		// #1013: the candidates come back in ZONE order, which is an
		// arbitrary answer to "which three cards does this Uro eat".
		// cheapestFuelFirst asks the seat's own policy what each one is
		// worth to KEEP and sorts the cheapest to the front, so the
		// first combination below is the best payment the policy can
		// name and the next few are that payment with its last card
		// swapped for the next-cheapest.
		pool := e.cheapestFuelFirst(g.AltCostCandidatesLocked(e.seat, card.InstanceID, offer))
		altCostSets = combinations(pool, want, want, maxEnumeratedCostPayments)
		if len(altCostSets) == 0 {
			// Unreachable through CastOffersForLocked, which already
			// dropped an offer with too few candidates. Kept because
			// this function is the one that writes the payment: a
			// future caller that skips the offer filter must not be
			// able to emit a cast with an unpayable cost.
			return
		}
	}
	discardSets := [][]uuid.UUID{nil}
	sacrificeSets := [][]uuid.UUID{nil}
	// ADR 0100 §6: a VARIABLE sacrifice clause on a cast — "sacrifice X
	// lands", "you may sacrifice any number of creatures". Its count is
	// part of the announcement, so each count is its own move, priced on
	// its own: Torgaar's "{2} less for each creature sacrificed this way"
	// reads the count at CR 601.2f (CostQuery.Sacrificing), and for the X
	// form the count IS the X (CR 107.3i). Register holds such a plan to
	// this one sacrifice clause, so the payment is the whole list.
	varSac := sacrifice != nil && game.SacrificeCostVariable(sacrifice)
	sacFromX := varSac && game.SacrificeCountFromX(sacrifice)
	var sacOrdered []uuid.UUID
	if discards > 0 {
		var pool []uuid.UUID
		for _, h := range p.Hand.Cards {
			if h.InstanceID != card.InstanceID {
				pool = append(pool, h.InstanceID)
			}
		}
		discardSets = combinations(pool, discards, discards, e.opts.MaxExpansionPerSource)
		if len(discardSets) == 0 {
			return
		}
	}
	if sacrifice != nil {
		// Cost, not target — see SpecCandidatesForEffect.
		lt := g.SpecCandidatesForEffect(e.seat, sacrifice)
		var pool []uuid.UUID
		for _, id := range lt.Cards {
			if c := findBattlefield(g, id); c != nil && c.Controller == e.seat {
				pool = append(pool, id)
			}
		}
		if varSac {
			sacOrdered = g.SacrificePaymentOrderForEffect(e.cheapestFuelFirst(pool), uuid.Nil)
			sacrificeSets = castVariableSacrificePayments(sacOrdered, sacrifice, xFloor)
		} else {
			// #747: N from the clause, one payment per cast for N ≥ 2,
			// nothing offered when the caster controls fewer than N.
			sacrificeSets = e.sacrificePayments(pool, sacrifice, uuid.Nil)
		}
		if len(sacrificeSets) == 0 {
			return
		}
	}

	// #1703: teamwork's taps and blight's creature. ONE payment each,
	// not every subset — crewPayment's discipline, and for crew's
	// reason: the number is a floor, and the policy has nothing to
	// choose between two sets that clear it. A set of optional costs
	// the seat cannot pay is not announced at all.
	// ADR 0100 §2: a branch's blight is paid through the same walk.
	teamIDs, blightIDs, ok := e.teamworkBlightPayment(addCost, optional, chosen)
	if !ok {
		return
	}

	// ADR 0100 §6: delve's graveyard. The candidates are the engine's
	// own walk, in the seat's fuel order when a policy supplies one, so
	// the cards a payment eats first are the ones the seat misses
	// least. An empty graveyard is no delve at all.
	var delvePool []uuid.UUID
	if g.DelveForLocked(e.seat, card) {
		delvePool = e.cheapestFuelFirst(g.DelveOptionsForEffect(e.seat, card.InstanceID))
	}

	budget := e.opts.MaxExpansionPerSource
	emit := e.castMoveEmitter(g, card, from, offer, optional, chosen, giftTo, teamIDs, blightIDs, branch, addCost)
	// #1013: the first announcement the expansion makes, kept so the
	// ALTERNATIVE cost payments can be offered against it below.
	var first *announcedCast
	for _, modes := range modeSets {
		// Spree (CR 702.172a): this selection's own mana joins the
		// base cost at the same point printedCostLocked adds it
		// (ADR 0073 §3's precedence), so the price this candidate is
		// judged against is exactly what CastSpell will charge for it
		// (#544). A no-op — modeCost equals cost — for a ModeSpec with
		// no Cost on any option, which is every modal card before S45.
		modeCost, err := game.AddModeCostMana(cost, modeSpec, modes)
		if err != nil {
			continue
		}
		x := 0
		// #1677: how many Phyrexian symbols this announcement pays with
		// life — solved with X, because striking a symbol changes what
		// X the pool can afford (the same pair affordablePayment solves
		// for an ability).
		phyLife := 0
		// #1242: the priced cost the X was solved against, kept for the
		// per-payment affordability check in the expansion below. Solved
		// per MODE SELECTION now rather than once for the card, because
		// modeCost — and so this price — can differ between selections
		// (Spree, S45). Since #1677 it is the cost LEFT FOR MANA once
		// the life has struck its symbols; printedAll is the cost
		// before the strike, which the all-life payment below reads.
		var pricedAll, printedAll game.ParsedCost
		// ADR 0100 §6: the delve payment the X was solved with — the
		// fewest cards that make the cast affordable — and the full
		// budget, which is offered once, against the first announcement.
		var delveIDs []uuid.UUID
		var delveFull *delvePayment
		// ADR 0100 §6: a variable sacrifice is priced per payment below,
		// because its count changes the price — so there is no one
		// up-front price to gate on. A Torgaar the seat cannot afford at
		// full price may well be affordable with three creatures
		// sacrificed.
		if !perTarget && !varSac {
			priced, err := e.g.ApplyCostModifiersForEffect(modeCost, game.CostQuery{
				Card:       card,
				Controller: e.seat,
				FromZone:   fromZone,
			})
			if err != nil {
				continue
			}
			pay, ok := e.solveCast(priced, spend, xFloor, xLifeCeiling, lifeReserved, lifeAllowed, delvePool)
			if !ok {
				continue
			}
			x, phyLife, pricedAll, printedAll = pay.x, pay.phyrexianLife, pay.cost, priced
			delveIDs, delveFull = pay.delve, pay.delveFull
		}
		// The budget is spent MODES-outermost: every mode selection
		// gets at least one target set before any gets a second, so a
		// bot is never offered only the first bullet of a charm
		// (ADR 0065 §6).
		steps := game.AnnouncedClauses(cardSpec, modeSpec, modes)
		// #1657: a divided amount read off the board or off the offer
		// this expansion claims (Avacyn's Judgment's madness X), sized
		// exactly as CastSpell's gate will size it.
		e.g.BindDivideAmountsForEffect(steps, game.DivideAmountArgs{
			Controller: e.seat, Source: card.InstanceID, AltCost: offerKey(offer),
		})
		// #619, CR 601.2c. Crackle with Power's target count IS X
		// ("deals five times X damage to each of up to X targets"),
		// and the announce path refuses any cast where the two
		// disagree. So for such a step the enumerator does not pick an
		// X and a target list independently — it picks the targets,
		// and the X it announces is how many it picked.
		xSteps := stepsCountedByX(steps)
		if len(xSteps) > 0 {
			if modeCost.XSlots == 0 && !sacFromX {
				// The step's X is announced by a cost this package
				// cannot price — Waterbender's Restoration's
				// waterbend {X}, paid by tapping artifacts and
				// creatures. The only announcement the enumerator
				// could make for it is X=0, which buys no targets and
				// a spell that does nothing, and any larger one would
				// be a move the engine refuses for an unpaid cost. So
				// the cast is not enumerable, the same answer an
				// unparseable cost gets.
				continue
			}
			// The arities to generate are bounded by the largest X
			// the seat could announce. Under a per-target price that
			// is not known until each set is priced, so open the step
			// to MaxX and let the affordability check below drop what
			// cannot be paid; the budget caps the expansion either
			// way.
			bound := x
			if perTarget {
				bound = e.opts.MaxX
			}
			// ADR 0100 §6: "Sacrifice X creatures. Destroy X target
			// creatures" — X is the sacrifice count, so the arities are
			// bounded by the permanents the seat can sacrifice.
			if sacFromX {
				bound = min(len(sacOrdered), e.opts.MaxX)
			}
			if bound < 1 {
				continue
			}
			steps = openXCountedSteps(steps, bound)
		}
		// #1559: "with mana value X or less" is judged under the X
		// the cast will announce. Bound to the largest X the seat can
		// pay, so the sets built below are ones the engine accepts at
		// that X; a set priced lower per target is re-checked in the
		// loop against its own X.
		xBound := game.StepsBoundByX(steps)
		if xBound {
			game.BindStepsXForEffect(steps, x)
		}
		targetSets := e.legalStepSets(castSrc, steps, budget)
		if len(targetSets) == 0 {
			continue
		}
		for _, targets := range targetSets {
			setX, setLife := x, phyLife
			setCost, setPrinted := pricedAll, printedAll
			setDelve, setDelveFull := delveIDs, delveFull
			if perTarget {
				// §14: priced with this set's targets. An unaffordable
				// set is skipped before any budget is spent on it, so
				// a Fireball the seat can pay for at one target is not
				// crowded out by the three-target sets it cannot.
				priced, err := e.g.ApplyCostModifiersForEffect(modeCost, game.CostQuery{
					Card:       card,
					Controller: e.seat,
					FromZone:   fromZone,
					Targets:    targets,
				})
				if err != nil {
					continue
				}
				pay, ok := e.solveCast(priced, spend, xFloor, xLifeCeiling, lifeReserved, lifeAllowed, delvePool)
				if !ok {
					continue
				}
				setX, setLife, setCost, setPrinted = pay.x, pay.phyrexianLife, pay.cost, priced
				setDelve, setDelveFull = pay.delve, pay.delveFull
			}
			// The payments this target set is offered with. For
			// "sacrifice X … destroy X target …" the targets fix X, and
			// so the ONE payment: the first X of the payment order (ADR
			// 0100 §6 — for the X form the X search and the payment are
			// a single dimension).
			sacSets := sacrificeSets
			if len(xSteps) > 0 {
				// setX is the largest announcement this seat can pay
				// for; a set that filled more X-counted slots than
				// that is a cast it cannot make. The cost is monotonic
				// in X, so the comparison is the whole affordability
				// check. For a sacrifice-X card the bound is the
				// permanents it can sacrifice instead.
				k, ok := announcedXCount(steps, xSteps, targets)
				if !ok || k < 1 {
					continue
				}
				if sacFromX {
					if k > len(sacOrdered) {
						continue
					}
					sacSets = [][]uuid.UUID{sacOrdered[:k]}
				} else if k > setX {
					continue
				}
				setX = k
			}
			if xBound && setX != x && !e.g.TargetsWithinXForEffect(steps, targets, setX) {
				continue
			}
			// #1563: a divided clause needs a share of at least 1 per
			// target out of the amount this X gives, so a set with more
			// targets than that is a cast the engine refuses (#544).
			dist, ok := game.EvenDistribution(steps, targets, setX)
			if !ok {
				continue
			}
			for _, discards := range discardSets {
				for _, sacs := range sacSets {
					if budget <= 0 {
						return
					}
					payX, payLife, payCost, payPrinted := setX, setLife, setCost, setPrinted
					payDelve, payDelveFull := setDelve, setDelveFull
					if varSac {
						// ADR 0100 §6: priced with THIS payment's count,
						// through the pricer CastSpell charges with — the
						// discount (CostQuery.Sacrificing) and, for the X
						// form, the X the count announces.
						if sacFromX {
							payX = len(sacs)
						}
						var tgts []game.TargetRef
						if perTarget {
							tgts = targets
						}
						priced, err := e.g.ApplyCostModifiersForEffect(modeCost, game.CostQuery{
							Card:        card,
							Controller:  e.seat,
							FromZone:    fromZone,
							XValue:      payX,
							Targets:     tgts,
							Sacrificing: len(sacs),
						})
						if err != nil {
							continue
						}
						pay, ok := e.solveCast(priced, spend, xFloor, xLifeCeiling, lifeReserved, lifeAllowed, delvePool)
						if !ok {
							continue
						}
						if !sacFromX {
							if len(xSteps) == 0 {
								payX = pay.x
							} else if pay.x < payX {
								continue
							}
						}
						payLife, payCost, payPrinted = pay.phyrexianLife, pay.cost, priced
						payDelve, payDelveFull = pay.delve, pay.delveFull
					}
					// #1242: CastSpell's auto-tap will not spend a card
					// or permanent this cast names to its additional cost
					// (game.CastAutoTapExclusions), so the affordability
					// solved above — with nothing named — has to hold with
					// these named. Village Rites naming the Eldrazi Spawn
					// that was also its {B} is a cast the engine refuses;
					// offering it is #544.
					//
					// #1703: and the same for the creatures tapped to
					// teamwork — a Llanowar Elves in the team cannot
					// also make the {G}.
					//
					// #1727: and for the alternative cost's own card
					// component — an Eldrazi Spawn named to Dread
					// Return's "sacrifice three creatures" cannot also
					// make the {1} a Thalia adds to the flashback.
					if len(discards)+len(sacs)+len(teamIDs)+len(blightIDs)+len(altCostSets[0]) > 0 &&
						!e.canPayExcluding(payCost, payX, spend, game.CastAutoTapExclusions(game.CastSpellParams{
							DiscardIDs:   discards,
							SacrificeIDs: sacs,
							TeamworkIDs:  teamIDs,
							BlightIDs:    blightIDs,
							AltCostIDs:   altCostSets[0],
						})) {
						continue
					}
					budget--
					if first == nil {
						first = &announcedCast{
							modes:     modes,
							targets:   targets,
							x:         payX,
							life:      payLife,
							cost:      payCost,
							printed:   payPrinted,
							dist:      dist,
							discards:  discards,
							sacs:      sacs,
							team:      teamIDs,
							blight:    blightIDs,
							alt:       altCostSets[0],
							delve:     payDelve,
							delveFull: payDelveFull,
						}
					}
					emit(altCostSets[0], modes, targets, payX, payLife, dist, discards, sacs, payDelve)
				}
			}
		}
	}
	// #1013: the ALTERNATIVE payments, out of whatever expansion budget
	// the target walk did not use — never out of ITS budget. ADR 0033
	// §1's corollary is that a variable in a COST must not become an
	// arity of the target cross product, and a payment that displaced
	// a target would be exactly that. An Uro with no targets has
	// eleven unspent and gets its alternatives; a removal spell with an
	// escape cost over a wide board spends its budget on targets and
	// gets none.
	//
	// They repeat the FIRST announcement, because that is what they
	// are: the same spell with the same targets at a different price.
	if first == nil {
		return
	}
	// #1677: the ALL-LIFE payment of the first announcement, out of
	// the same leftover budget and for the same reason — it is the
	// same spell with the same targets at a different price, so it
	// must not displace a target set. See castPayment for why this is
	// the one extra count offered.
	if budget > 0 && lifeAllowed {
		if n, ok := e.allLifePayment(first, spend, lifeReserved); ok {
			budget--
			emit(altCostSets[0], first.modes, first.targets, first.x, n, first.dist, first.discards, first.sacs, first.delve)
		}
	}
	// ADR 0100 owner decision 3: the FULL delve payment of the first
	// announcement, out of the leftover budget for the #1013 reason — it
	// is the same spell with the same targets, paid with more of the
	// graveyard and less mana, and must not displace a target set.
	// Murktide Regent and Soulflayer want it; the policy prices the fuel
	// either way.
	if budget > 0 && first.delveFull != nil && e.delveFullAffordable(first, spend) {
		budget--
		emit(altCostSets[0], first.modes, first.targets, first.delveFull.x, first.life, first.dist, first.discards, first.sacs, first.delveFull.ids)
	}
	for _, altPaid := range altCostSets[1:] {
		if budget <= 0 {
			return
		}
		// #1727: the same affordability, with THIS payment's cards
		// kept away from the auto-tapper (game.CastAutoTapExclusions) —
		// a different three creatures may include the Spawn the first
		// payment left free to make the mana.
		if !e.canPayExcluding(first.cost, first.x, spend, first.autoTapExclusions(altPaid)) {
			continue
		}
		budget--
		emit(altPaid, first.modes, first.targets, first.x, first.life, first.dist, first.discards, first.sacs, first.delve)
	}
}

// announcedCast is one announcement the expansion has already emitted:
// what an alternative-cost payment is offered AGAINST (#1013).
type announcedCast struct {
	modes   []int
	targets []game.TargetRef
	x       int
	// life is how many Phyrexian symbols the announcement pays with
	// life (#1677); printed is its cost before that strike, and cost
	// the mana it pays after it.
	life     int
	cost     game.ParsedCost
	printed  game.ParsedCost
	dist     map[uuid.UUID]int
	discards []uuid.UUID
	sacs     []uuid.UUID
	// team and blight are the #1703 payments, excluded from the
	// auto-tap plan the all-life payment is re-checked against.
	team   []uuid.UUID
	blight []uuid.UUID
	// alt is the alternative cost's card payment the announcement was
	// emitted with (#1727) — excluded from the auto-tap plan like the
	// rest.
	alt []uuid.UUID
	// delve is the announcement's delve payment (ADR 0100 §6), and
	// delveFull the full-budget one offered beside it — nil when the
	// two are the same set.
	delve     []uuid.UUID
	delveFull *delvePayment
}

// castEmitter writes one concrete cast move.
//
// Lifted out of the expansion loop by #1013 so the alternative-payment
// pass reuses it verbatim rather than growing a second copy of the
// label and the params — the #815 / #866 lesson at the move layer: two
// writers of one move shape drift, and the one that drifts is the one
// nobody reads.
type castEmitter func(altPaid []uuid.UUID, modes []int, targets []game.TargetRef, setX, phyLife int, dist map[uuid.UUID]int, discards, sacs, delve []uuid.UUID)

// castMoveEmitter builds that writer for one (card, zone, offer,
// optional-cost) announcement. Everything it closes over is fixed for
// the whole expansion; everything that varies is an argument.
func (e *enumerator) castMoveEmitter(
	g *game.Game,
	card game.Card,
	from string,
	offer *game.AlternativeCost,
	optional []game.AdditionalCost,
	chosen []int,
	giftTo uuid.UUID,
	teamIDs, blightIDs []uuid.UUID,
	branch *int,
	paying *game.AdditionalCost,
) castEmitter {
	// #1918: fixed for the whole expansion — it reads the offer and the
	// board, never the targets (a cast it applies to has none).
	idle := e.idleCastHint(card, offer, chosen)
	return func(altPaid []uuid.UUID, modes []int, targets []game.TargetRef, setX, phyLife int, dist map[uuid.UUID]int, discards, sacs, delve []uuid.UUID) {
		label := "Cast " + card.Name
		switch from {
		case "command":
			label += " from the command zone"
		case "graveyard", "exile", "library":
			label += " from " + from
		}
		if setX > 0 {
			label += fmt.Sprintf(" for X=%d", setX)
		}
		// #673: the price. A flashed-back Faithless Looting, an escaped
		// Uro and a hard-cast one are otherwise the same line in the
		// move log, and a bot eval that cannot tell them apart cannot
		// explain why the bot escaped.
		label += altCostLabel(g, offer, altPaid)
		// ADR 0073: the kicked and unkicked casts are otherwise the
		// same line in the move log, and a bot eval that cannot tell
		// them apart cannot explain why the bot kicked.
		label += optionalCostLabel(optional, chosen)
		// ADR 0100: the two branches of an either/or cost are otherwise
		// the same line in the move log.
		if branch != nil && paying != nil {
			label += " (" + paying.Label + ")"
		}
		if giftTo != uuid.Nil {
			label += " → " + playerName(g, giftTo)
		}
		if phyLife > 0 {
			label += fmt.Sprintf(" paying %d life for Phyrexian mana", phyLife*game.PhyrexianLifePerSymbol)
		}
		// ADR 0100: the delved and the undelved casts are otherwise the
		// same line in the move log.
		if len(delve) > 0 {
			label += fmt.Sprintf(" delving %d", len(delve))
		}
		// ADR 0100 §6: the counts of a variable sacrifice are otherwise
		// the same line in the move log.
		if paying != nil && game.SacrificeCostVariable(paying.Sacrifice) {
			label += sacrificeLabel(g, sacs)
		}
		label += targetLabel(g, targets)
		e.add(Move{
			Type:   TypeCastSpell,
			Player: e.seat,
			Kind:   KindCast,
			Label:  label,
			Source: card.InstanceID,
			// CR 119.4: the life half of the offer is a price Params
			// cannot name, so a policy reading only the payload would
			// price Force of Will's pitch as free. See MoveCost.
			// #1677: and the Phyrexian symbols this move pays with
			// life, which Params names only as a count.
			// ADR 0100: and a branch's fixed "pay 3 life".
			Cost: withPhyrexianLife(moveCost(offerLife(offer)+branchLife(paying), 0), phyLife),
			// A modal spell may have a counter mode and a burn mode
			// in the same expansion (Cryptic Command); the flag is
			// per ANNOUNCEMENT, not per card, so only the modes that
			// actually chose a stack target come back flagged.
			TargetsStack: targetsStackObject(g, targets),
			IdleHint:     idle,
			Params: mustJSON(castParams{
				InstanceID:      card.InstanceID.String(),
				FromZone:        from,
				AlternativeCost: offerKey(offer),
				AltCostIDs:      idStrings(altPaid),
				PhyrexianLife:   phyLife,
				Targets:         wireTargets(targets),
				Modes:           modes,
				XValue:          setX,
				DiscardIDs:      idStrings(discards),
				SacrificeIDs:    idStrings(sacs),
				OptionalCosts:   chosen,
				CostBranch:      branch,
				TeamworkIDs:     idStrings(teamIDs),
				BlightIDs:       idStrings(blightIDs),
				DelveIDs:        idStrings(delve),
				Distribution:    distributionWire(dist),
				GiftOpponent:    giftWire(giftTo),
				Strict:          true,
				AutoTap:         true,
				// ADR 0034: `card` has already had SetFace applied by
				// the caller, so ActiveFace IS the face this move casts.
				Face: card.ActiveFace,
				Fuse: card.Fused,
			}),
		})
	}
}

// idleCastHint is Move.IdleHint for a cast under `offer` (#1918): the
// player-facing reason an overloaded spell would do nothing right now,
// or "" when it would do something, or when the cast is not one this
// asks about.
//
// Only an offer that DELETES the target clause is asked about — the
// one shape where the engine knows exactly what the spell would touch
// without being told: CR 702.96a turns "target X" into "each X", so the
// set is the clause's own filter. And it is read without the targeting
// gate (SpecCandidatesForEffect, not LegalTargetsForEffect): "each" is
// not targeting (CR 702.96b), so a hexproof or shrouded permanent, or
// one with protection, is still something an overloaded Cyclonic Rift
// returns. The clause is the one the cast would have announced without
// the offer, after any optional cost that swaps it (ADR 0089 §3), as
// castMovesPayingOptional reads it.
//
// Caller holds g.mu, as for the rest of the enumeration.
func (e *enumerator) idleCastHint(card game.Card, offer *game.AlternativeCost, chosen []int) string {
	if !offer.Clears() {
		return ""
	}
	key := game.CatalogKey(card)
	spec := game.TargetSpecUnderOptionalCosts(game.TargetSpecFor(key), game.OptionalCostsFor(key), chosen)
	if spec == nil {
		return ""
	}
	lt := e.g.SpecCandidatesForEffect(e.seat, spec)
	if len(lt.Players)+len(lt.Cards) > 0 {
		return ""
	}
	how := "Cast this way"
	if offer.Key == "overload" {
		how = "Overloaded"
	}
	what := "there's nothing for it to affect."
	if noun, ok := strings.CutPrefix(spec.Label, "target "); ok && noun != "" {
		what = "there's no " + noun + "."
	}
	return how + ", this does nothing right now: " + what
}

// distributionWire is a division on the wire: target id string →
// share. Nil for the announcement that divides nothing.
func distributionWire(dist map[uuid.UUID]int) map[string]int {
	if len(dist) == 0 {
		return nil
	}
	out := make(map[string]int, len(dist))
	for id, v := range dist {
		out[id.String()] = v
	}
	return out
}

// cheapestFuelFirst orders the candidates for a cost's CARD-shaped half
// so the payment the enumerator builds first is the one the seat would
// miss least (#1013).
//
// The pool arrives in ZONE order — the oldest cards in the graveyard,
// the left of the hand — which is an arbitrary answer to "which three
// cards does this Uro eat", and arbitrary is what the cap of one made
// permanent: an escape over a graveyard holding a second Uro, a
// Snapcaster target and three lands ate whichever three were oldest.
//
// A HOOK rather than a scorer here, for the reason Options.OrderTargets
// is one (ADR 0033 §1): what a card in a graveyard is worth to a seat
// is a POLICY question, and `legal` may not import `aiseat`. A seat
// with no opinion gets zone order, which is byte-identical to what it
// got before this existed.
//
// The sort is STABLE and ASCENDING in "worth keeping", so equal prices
// keep zone order and two enumerations of one board produce the same
// move list. The returned slice is fresh: the pool comes from the
// engine and must not be reordered under it.
func (e *enumerator) cheapestFuelFirst(pool []uuid.UUID) []uuid.UUID {
	if e.opts.OrderCostFuel == nil || len(pool) < 2 {
		return pool
	}
	out := append([]uuid.UUID(nil), pool...)
	price := make(map[uuid.UUID]float64, len(out))
	for _, id := range out {
		price[id] = e.opts.OrderCostFuel(TargetCandidate{ID: id})
	}
	sort.SliceStable(out, func(i, j int) bool { return price[out[i]] < price[out[j]] })
	return out
}

// affordableXFrom reports whether the seat can pay cost right now —
// from the floating pool, by the auto-tapper's plan, or by the two
// together — and, for an {X} cost, the largest X it can pay, at or
// above `floor`, up to MaxX. Mirrors the engine's auto_tap + strict
// path: the pool is consulted first, then a plan is sought for what
// the pool is missing (ADR 0118 §1, the pool top-up).
//
// The floor is the announcement's lower bound, and it has two sources,
// both settled by enumeratedXFloor (x.go) before the call: a printed
// "X can't be 0" (Helm of Obedience), and #810's rule that a move
// whose whole effect is X is not worth offering at X=0. A seat that
// cannot pay for the floor has no move at all rather than a free one.
//
// The scan still starts at the floor and still breaks on the first
// unaffordable value, because the cost is monotonic in X: every
// extra point of X buys the same XSlots generic symbols. Nothing
// here enumerates a RANGE — exactly one X comes back, so X never
// enters an expansion cross product (see activatedMoves for why
// that matters).
func (e *enumerator) affordableXFrom(cost game.ParsedCost, spend game.ManaSpendContext, floor int) (int, bool) {
	return e.affordableXExcluding(cost, spend, floor, nil)
}

// affordableXExcluding is affordableXFrom with sources the payment
// may not use — the one caller is an activated ability whose cost
// includes {T}, which cannot tap its own source for mana.
func (e *enumerator) affordableXExcluding(
	cost game.ParsedCost,
	spend game.ManaSpendContext,
	floor int,
	excluded map[uuid.UUID]bool,
) (int, bool) {
	if cost.XSlots == 0 {
		// No {X}: the floor is meaningless and X is always zero.
		return 0, e.canPayExcluding(cost, 0, spend, excluded)
	}
	if floor < 0 {
		floor = 0
	}
	best, ok := -1, false
	for x := floor; x <= e.opts.MaxX; x++ {
		if e.canPayExcluding(cost, x, spend, excluded) {
			best, ok = x, true
			continue
		}
		break
	}
	return best, ok
}

// `spend` is the #352 spend context — what the mana would be paid
// for — so restricted mana in the pool counts toward a cast it may
// legally fund and toward no other.
func (e *enumerator) canPay(cost game.ParsedCost, x int, spend game.ManaSpendContext) bool {
	return e.canPayExcluding(cost, x, spend, nil)
}

// canPayExcluding is canPay with sources the auto-tapper may not
// reach for. It has to mirror exactly what the engine excludes, or
// the enumerator's answer and the engine's answer disagree — which
// is the #544 failure mode, one cost component over.
func (e *enumerator) canPayExcluding(
	cost game.ParsedCost,
	x int,
	spend game.ManaSpendContext,
	excluded map[uuid.UUID]bool,
) bool {
	// #1600: the cost as this seat may pay it — widened under Chromatic
	// Orrery's "you may spend mana as though it were mana of any color"
	// — through the function every engine payment reads it with. Here,
	// in the one affordability probe the cast, activation, special
	// action, delve, waterbend and pay-unless moves all ask, so none of
	// them can offer a payment the engine would refuse or hide one it
	// would accept.
	cost = e.g.CostAsPaidByForEffect(e.seat, spend, cost, x)
	if e.p.ManaPool.CanPayFor(cost, x, spend) {
		return true
	}
	// ADR 0118 §1: the pool and a plan TOGETHER, through the function
	// every engine auto-tap payer reads — {G} floating plus one Forest
	// pays {1}{G}, where asking the pool alone and the lands alone
	// called it unpayable twice.
	return e.g.AutoTapTopUpForEffectExcluding(e.seat, cost, x, spend, excluded)
}

// castPaymentSolve is one priced way to pay a cast's mana cost: the X
// it announces, how many Phyrexian symbols it pays with life, and the
// cost left for mana once those symbols are struck.
type castPaymentSolve struct {
	x             int
	phyrexianLife int
	cost          game.ParsedCost
}

// castPayment solves a cast's mana cost for the pair CR 601.2b makes
// the caster announce: the X, and how many of the cost's Phyrexian
// symbols are paid with 2 life each (CR 107.4c, #1677). The spell
// twin of affordablePayment, which does the same for an activated
// ability (#917); before #1677 a cast was priced on the mana path only,
// so a bot holding Dismember and one Swamp was never offered it, at
// any life total.
//
// WHICH COUNTS, and why only these. A cost printing n Phyrexian
// symbols has n+1 legal announcements, and every one of them is the
// same spell with the same targets at a different price, so offering
// all of them would make the symbol count an arity of the target
// cross product — the thing ADR 0033 §1's corollary forbids a cost
// variable to become. Two are enough to carry every decision a policy
// has to make:
//
//   - The PRIMARY payment, solved here: all mana when the seat can pay
//     it, otherwise the FEWEST symbols by life that make the cast
//     affordable. Mana first because life is a real cost and the
//     enumerator never spends it to save mana the board could produce;
//     fewest because every further symbol is 2 life that buys nothing
//     the smaller count did not already buy.
//   - The ALL-LIFE payment (allLifePayment), offered once per card out
//     of whatever budget the target walk left: the free cast that
//     saves the mana for something else — Gitaxian Probe with an
//     Island untapped, Dismember off an empty board before a second
//     spell.
//
// Every intermediate count (more than the fewest, fewer than all) is
// the same trade as one of those two, paid partly, and is not offered.
//
// Bounded by the same two things the engine validates against, so an
// offered payment is one CastSpell accepts: the symbols the cost
// actually prints, and game.CanPayLifeLocked — CR 119.4's "down to 0"
// and CR 119.8's locked total. `reserved` is life the announcement
// already owes elsewhere (the offer's own life component), so the two
// together can never claim more than the seat has. The strike itself
// is game.PhyrexianLifePlan, the function the engine reduces the cost
// with, read against the same pool.
func (e *enumerator) castPayment(
	priced game.ParsedCost,
	spend game.ManaSpendContext,
	floor, lifeCeiling, reserved int,
	lifeAllowed bool,
) (castPaymentSolve, bool) {
	if x, ok := e.announcedX(priced, spend, floor, lifeCeiling); ok {
		return castPaymentSolve{x: x, cost: priced}, true
	}
	if !lifeAllowed {
		return castPaymentSolve{}, false
	}
	for n := 1; n <= priced.PhyrexianSymbols(); n++ {
		reduced, life := game.PhyrexianLifePlan(priced, e.p.ManaPool, spend, n)
		if !e.g.CanPayLifeLocked(e.p, reserved+life) {
			break
		}
		if x, ok := e.announcedX(reduced, spend, floor, lifeCeiling); ok {
			return castPaymentSolve{x: x, phyrexianLife: n, cost: reduced}, true
		}
	}
	return castPaymentSolve{}, false
}

// allLifePayment is the second count castPayment's comment names:
// every Phyrexian symbol of an announcement already emitted paid with
// life, at that announcement's X and with its additional-cost
// payments. Reports false when the first announcement already paid
// every symbol with life, when the seat cannot pay that much life, or
// when the mana left over is somehow unaffordable — striking more
// symbols only removes requirements, so the last is a belt, but the
// enumerator offers nothing it has not priced.
func (e *enumerator) allLifePayment(first *announcedCast, spend game.ManaSpendContext, reserved int) (int, bool) {
	n := first.printed.PhyrexianSymbols()
	if n == 0 || n <= first.life {
		return 0, false
	}
	reduced, life := game.PhyrexianLifePlan(first.printed, e.p.ManaPool, spend, n)
	if !e.g.CanPayLifeLocked(e.p, reserved+life) {
		return 0, false
	}
	if !e.canPayExcluding(reduced, first.x, spend, first.autoTapExclusions(first.alt)) {
		return 0, false
	}
	return n, true
}

// autoTapExclusions is what CastSpell's auto-tap will not spend on this
// announcement's mana (game.CastAutoTapExclusions): its additional-cost
// payments, its teamwork and blight creatures, and `alt`, the
// alternative cost's card payment it is being offered with (#1727).
func (a *announcedCast) autoTapExclusions(alt []uuid.UUID) map[uuid.UUID]bool {
	return game.CastAutoTapExclusions(game.CastSpellParams{
		DiscardIDs:   a.discards,
		SacrificeIDs: a.sacs,
		TeamworkIDs:  a.team,
		BlightIDs:    a.blight,
		AltCostIDs:   alt,
	})
}

// legalModeSets lists every distinct mode selection of size lo..hi,
// the bounds the announce gate will hold this announcement to — the
// printed Min / Max, or the raised ones while a conditional mode count
// holds (#1590, #1655, game.ModeBoundsForEffect). A seat without a
// commander is never offered Jeska's Will's "both", and a kicked
// Depth Defiler-shaped count is never offered one bullet.
//
// `ab` names the ability whose "that hasn't been chosen" memory
// applies (ADR 0097); a spell passes the zero value. A used mode is
// never offered, because the filter below is the one the activation
// gate and the trigger's mode_pick read (#544).
func (e *enumerator) legalModeSets(src game.TargetSource, ms *game.ModeSpec, q game.ModeCountQuery, ab game.ModeAbility) [][]int {
	// ADR 0065 §6, "prefer the modes that have legal targets": an
	// option whose clause cannot be filled is dropped before any
	// combination is built, so the budget never goes on a selection
	// the engine would refuse at announce.
	options := e.g.ChoosableModeOptionsForEffect(src, ms, ab)
	if !game.EnoughChoosableModes(len(options), ms) {
		return nil
	}
	lo, hi := e.g.ModeBoundsForEffect(ms, q)
	if !ms.Repeatable && lo > len(options) {
		// A forced count the board cannot fill: the gate refuses every
		// selection, so there is no move (#544).
		return nil
	}
	if hi <= 0 || (!ms.Repeatable && hi > len(options)) {
		hi = len(options)
	}
	budget := e.opts.MaxExpansionPerSource
	var out [][]int
	add := func(sel []int) bool {
		out = append(out, append([]int(nil), sel...))
		return len(out) < budget
	}
	if lo == 0 {
		if !add(nil) {
			return out
		}
	}
	if lo < 1 {
		lo = 1
	}
	// CR 700.2d: the all-one-option selections first, so a
	// repeatable spec whose only legal option is one mode is not
	// crowded out by mixed multisets.
	if ms.Repeatable {
		for _, opt := range options {
			for n := lo; n <= hi; n++ {
				sel := make([]int, n)
				for i := range sel {
					sel[i] = opt
				}
				if !add(sel) {
					return out
				}
			}
		}
	}
	var rec func(start int, cur []int) bool
	rec = func(start int, cur []int) bool {
		if len(cur) >= lo && len(cur) <= hi {
			if !(ms.Repeatable && len(cur) == 1) && !add(cur) {
				return false
			}
		}
		if len(cur) == hi {
			return true
		}
		for i := start; i < len(options); i++ {
			if !rec(i+1, append(cur, options[i])) {
				return false
			}
		}
		return true
	}
	rec(0, nil)
	return out
}

// legalStepSets is the cartesian product of each clause's legal
// picks, in step order, capped at `budget` (#764, ADR 0065 §6). An
// announcement with no steps yields the single empty set, which is
// how an untargeted cast stays one move.
func (e *enumerator) legalStepSets(src game.TargetSource, steps []game.AnnouncedClause, budget int) [][]game.TargetRef {
	out := [][]game.TargetRef{nil}
	for i := range steps {
		clause := steps[i].Clause
		picks := e.legalTargetSets(src, &clause, budget)
		if len(picks) == 0 {
			return nil
		}
		next := make([][]game.TargetRef, 0, budget)
		for _, prefix := range out {
			for _, pick := range picks {
				if len(next) >= budget {
					break
				}
				combined := append([]game.TargetRef(nil), prefix...)
				skip := false
				for _, p := range pick {
					p.Mode, p.Slot = steps[i].Mode, steps[i].Slot
					// CR 601.2c: a Distinct clause may not repeat an
					// object an earlier clause took.
					if clause.Distinct {
						for _, seen := range prefix {
							if seen.ID == p.ID {
								skip = true
							}
						}
					}
					combined = append(combined, p)
				}
				if skip {
					continue
				}
				next = append(next, combined)
			}
		}
		if len(next) == 0 {
			return nil
		}
		out = next
	}
	return out
}

// stepsCountedByX lists the indexes of the announcement's steps whose
// target count is the announced X rather than a printed constant
// (Crackle with Power, Doppelgang, Heliod's Intervention's first
// mode). Empty — the overwhelming majority — means nothing here ties
// X to anything.
func stepsCountedByX(steps []game.AnnouncedClause) []int {
	var out []int
	for i := range steps {
		if steps[i].Clause.CountFromX {
			out = append(out, i)
		}
	}
	return out
}

// openXCountedSteps returns a copy of `steps` with every X-counted
// clause opened to 1..bound, so legalStepSets expands one target set
// per arity and the caller can read the arity back off the set.
//
// The engine's own resolveStepCountsFromX pins the same clauses to a
// single announced X; this is that operation with the X still unknown,
// which is the whole difference between validating an announcement and
// building one.
//
// The floor is 1, not 0. X=0 on such a step is a spell cast with no
// targets, which for every card in the catalog that has one means a
// spell that does nothing — #810's rule, and the reason this needs no
// zero case of its own. A CountFromX card with a fixed rider would
// want one; none exists, and x.go is where that would be decided.
func openXCountedSteps(steps []game.AnnouncedClause, bound int) []game.AnnouncedClause {
	out := append([]game.AnnouncedClause(nil), steps...)
	for _, i := range stepsCountedByX(out) {
		out[i].Clause.Min, out[i].Clause.Max = 1, bound
		out[i].Clause.CountFromX = false
	}
	return out
}

// announcedXCount reads the X a target set is announcing: how many
// refs answer the X-counted steps. False when two such steps disagree,
// which is a set no announcement could cover — one X, one count
// (CR 601.2c).
//
// Matched on (Mode, Slot), the coordinates legalStepSets stamps onto
// every ref, which is exactly what the engine's stepTargetCount reads.
func announcedXCount(steps []game.AnnouncedClause, xSteps []int, targets []game.TargetRef) (int, bool) {
	k := -1
	for _, i := range xSteps {
		n := 0
		for _, t := range targets {
			if t.Kind == game.TargetSelf || t.Kind == game.TargetNone {
				continue
			}
			if t.Mode == steps[i].Mode && t.Slot == steps[i].Slot {
				n++
			}
		}
		if k >= 0 && n != k {
			return 0, false
		}
		k = n
	}
	return k, k >= 0
}

// legalTargetSets expands a target clause into concrete target
// lists: the empty list when Min is 0, then every k-subset of the
// legal candidates for k in max(Min,1)..Max, players before cards,
// stopping at budget entries.
//
// `src` is the spell or ability doing the targeting, not just the
// seat: CR 702.16b tests protection against the SOURCE, so an
// enumerator that passed only a seat would offer the bot a pro-red
// creature as a target for its red spell and the server would then
// refuse the move — the #347 / #544 failure mode.
func (e *enumerator) legalTargetSets(src game.TargetSource, spec *game.TargetSpec, budget int) [][]game.TargetRef {
	lt := e.g.LegalTargetsForEffect(src, spec)
	cands := make([]game.TargetRef, 0, len(lt.Players)+len(lt.Cards))
	for _, id := range lt.Players {
		cands = append(cands, game.TargetRef{Kind: game.TargetPlayer, ID: id})
	}
	for _, id := range lt.Cards {
		cands = append(cands, game.TargetRef{Kind: game.TargetCard, ID: id})
	}
	e.orderCandidates(cands)
	// #1559: a clause with a set rule ("each with a different mana
	// value") never yields a set two of whose picks share a key — the
	// engine refuses that set at announce, and offering it is #544.
	// #1807: nor one whose picks hold two keys under a sameness rule
	// ("from a single graveyard"), so every set stays in one group.
	keys := setRuleKeys{
		different: e.g.TargetDifferenceKeysForEffect(spec, lt.Cards),
		same:      e.g.TargetSamenessKeysForEffect(spec, lt.Cards),
	}
	lo, hi := spec.Min, spec.Max
	if hi <= 0 || hi > len(cands) {
		hi = len(cands)
	}
	var out [][]game.TargetRef
	if lo == 0 {
		out = append(out, nil)
	}
	if lo < 1 {
		lo = 1
	}
	if lo > len(cands) {
		return out
	}
	for k := lo; k <= hi && len(out) < budget; k++ {
		var rec func(start int, cur []game.TargetRef)
		rec = func(start int, cur []game.TargetRef) {
			if len(out) >= budget {
				return
			}
			if len(cur) == k {
				out = append(out, append([]game.TargetRef(nil), cur...))
				return
			}
			for i := start; i < len(cands); i++ {
				if keys.breaks(cur, cands[i]) {
					continue
				}
				rec(i+1, append(cur, cands[i]))
			}
		}
		rec(0, nil)
	}
	return out
}

// setRuleKeys is a clause's set-rule keys, per legal card: `different`
// for a rule no two picks may share a key under (#1559), `same` for
// one every pick must share a key under (#1807). Either is nil when
// the clause has no such rule.
type setRuleKeys struct {
	different map[uuid.UUID]string
	same      map[uuid.UUID]string
}

// breaks reports whether adding cand to cur would give a set the
// engine's announce gate refuses.
func (k setRuleKeys) breaks(cur []game.TargetRef, cand game.TargetRef) bool {
	return (k.different != nil && sharesSetKey(k.different, cur, cand)) ||
		(k.same != nil && leavesSetKey(k.same, cur, cand))
}

// sharesSetKey reports whether cand's set-rule key is already held by
// a pick in cur (#1559). A pick with no key shares nothing.
func sharesSetKey(keys map[uuid.UUID]string, cur []game.TargetRef, cand game.TargetRef) bool {
	k, ok := keys[cand.ID]
	if !ok {
		return false
	}
	for _, p := range cur {
		if pk, ok := keys[p.ID]; ok && pk == k {
			return true
		}
	}
	return false
}

// leavesSetKey reports whether cand's sameness key differs from one a
// pick in cur holds (#1807): a card from a second graveyard. A pick
// with no key fits any group.
func leavesSetKey(keys map[uuid.UUID]string, cur []game.TargetRef, cand game.TargetRef) bool {
	k, ok := keys[cand.ID]
	if !ok {
		return false
	}
	for _, p := range cur {
		if pk, ok := keys[p.ID]; ok && pk != k {
			return true
		}
	}
	return false
}

// orderCandidates sorts a clause's candidates so the ones the seat's
// policy cares most about are the ones that survive
// MaxExpansionPerSource (#687, ADR 0033 §1).
//
// The cap is spent in candidate order, so before this the answer to
// "which twelve of an opponent's twenty permanents may the bot point
// a removal spell at" was whatever order LegalTargetsForEffect
// happened to walk the battlefield in — and the table leader's
// Blightsteel Colossus could simply be absent from the move list, at
// which point no policy could pick it.
//
// A no-op without Options.OrderTargets, which is every caller but a
// bot seat. Sorted STABLY, so the engine's order is the tiebreak and
// two enumerations of one board agree.
func (e *enumerator) orderCandidates(cands []game.TargetRef) {
	if e.opts.OrderTargets == nil || len(cands) < 2 {
		return
	}
	// Priced once per candidate rather than inside the comparator: a
	// policy's score is a board read, and sort.SliceStable would ask
	// for it O(n log n) times.
	score := make(map[uuid.UUID]float64, len(cands))
	for _, c := range cands {
		score[c.ID] = e.opts.OrderTargets(TargetCandidate{
			ID:     c.ID,
			Player: c.Kind == game.TargetPlayer,
		})
	}
	sort.SliceStable(cands, func(i, j int) bool {
		return score[cands[i].ID] > score[cands[j].ID]
	})
}

// combinations returns every k-subset of pool for k in min..max, in
// pool order, up to limit entries. Returns nil when the pool is too
// small for min.
func combinations(pool []uuid.UUID, lo, hi, limit int) [][]uuid.UUID {
	if lo > len(pool) {
		return nil
	}
	if hi > len(pool) {
		hi = len(pool)
	}
	var out [][]uuid.UUID
	for k := lo; k <= hi && len(out) < limit; k++ {
		var rec func(start int, cur []uuid.UUID)
		rec = func(start int, cur []uuid.UUID) {
			if len(out) >= limit {
				return
			}
			if len(cur) == k {
				out = append(out, append([]uuid.UUID(nil), cur...))
				return
			}
			for i := start; i < len(pool); i++ {
				rec(i+1, append(cur, pool[i]))
			}
		}
		rec(0, nil)
	}
	return out
}

// offerKey is the wire key of a resolved alternative cost, or "" when
// the cast pays the printed price. Nil-safe so the emit site needs no
// guard.
func offerKey(offer *game.AlternativeCost) string {
	if offer == nil {
		return ""
	}
	return offer.Key
}

// offerLife is the life an offer charges (CR 119.4), or 0. Nil-safe.
func offerLife(offer *game.AlternativeCost) int {
	if offer == nil {
		return 0
	}
	return offer.Life
}

// altCostLabel spells the claimed alternative cost into the move's
// label — " (Flashback {2}{R})", " (Escape—{1}{G}{U}, Exile five
// other cards from your graveyard, exiling Mountain, Opt, …)".
//
// Empty for a cast that pays the printed price, which is every cast
// in almost every game. The cards paid are NAMED rather than counted:
// #673 asks for it explicitly, and an escape line whose log entry did
// not say what it ate is unreviewable.
func altCostLabel(g *game.Game, offer *game.AlternativeCost, paid []uuid.UUID) string {
	if offer == nil {
		return ""
	}
	label := offer.Label
	if label == "" {
		label = offer.Key
	}
	if len(paid) > 0 {
		names := make([]string, 0, len(paid))
		for _, id := range paid {
			names = append(names, cardName(g, id))
		}
		// The verb is the component's, not a generic "paying": a Daze
		// that logged "exiling Island" would be describing a different
		// card.
		verb := ", exiling "
		if offer.ReturnToHand != nil {
			verb = ", returning "
		}
		label += verb + strings.Join(names, ", ")
	}
	return " (" + label + ")"
}
