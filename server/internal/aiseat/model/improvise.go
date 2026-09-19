package model

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// improvise.go is ADR 0033 §8's improviser, for the `assisted` and
// `strong` tiers. §8 and S31 sub-PR 8 built the whole rail — the
// atomic bundle, the validated announcement, the replay tag, the free
// undo — and then nothing implemented `aiseat.Improviser`, so the
// section described a feature no game ever reached. This is the
// missing half (#686; the design is the 2026-09-19 amendment to §8).
//
// # The bot pays for the card
//
// This is the decision everything else hangs off, and it is easy to
// get wrong in the obvious direction. The four sandbox verbs cannot
// tap a land. An improviser that took a card out of hand ITSELF and
// applied its text would therefore be casting it for free — the one
// thing at this table worse than a bot quietly skipping a card it
// cannot run.
//
// So improvisation is not something the seat does INSTEAD of casting.
// The ordinary funnel casts the card through the ordinary move list,
// the engine charges the mana, the spell resolves into silence
// because the catalog has no spec for it, and only then does this
// file supply the text the engine did not carry out. That is exactly
// the sequence a human on this server performs by hand, and §8
// defines improvisation as doing what a human does.
//
// # The window, and its bound
//
// One improvisation per card instance, at the moment that instance
// leaves the stack, for a spell this seat cast whose CardView carries
// the `unimplemented` bit. Nothing else: an unimplemented permanent's
// ongoing abilities have no wire signal for "this should have fired"
// and would recur without bound, an opponent's card is never this
// seat's to correct, and a seat that owes a pending choice (#1000's
// OwedInStep, #1016's GuardsStackItem, every other PendingChoices
// entry) is answering a prompt that halts the table and must not be
// wandering off to write a bundle.
//
// # What the model is shown, and what it cannot be
//
// The card's oracle text — from the seat's own DeckProfile, which is
// configuration and not something read off the board — plus the
// seat's own filtered GameView, annotated with the instance IDs and
// player IDs a bundle is allowed to name. The IDs are the only thing
// here that a Layer C prompt does not already carry, and they are not
// new information: every one of them is already on the wire to this
// seat. Like the rest of this package, nothing here can reach an
// opponent's hand, because nothing here can reach a *game.Game.

// improvPrimer is the static, cache-marked half of an improvisation
// call: who the model is, the four verbs with their exact wire
// params, and the shape of the answer.
//
// It is a SEPARATE prefix from the decision primer rather than a
// section bolted onto it. Prompt caching is a prefix match, so
// sharing would buy nothing unless the two tasks shared their whole
// head — and they do not: one asks for an integer out of a closed
// list, the other asks for a bundle of verbs. Mixing the two
// instruction sets to save a cache entry on a call that happens a
// handful of times a game is the wrong trade.
const improvPrimer = `You are playing one seat in a four-player game of Magic: the Gathering (Commander) on a private hobby server.

A spell YOU cast has just resolved and the server's rules engine did not carry out its text — this server implements only a few hundred cards, and this one is not among them. You are being asked to apply that card's effect BY HAND, with the same small set of sandbox verbs a human at this table uses for the same job.

The card's cost is already paid and the card has already moved to where it belongs. Do not pay a cost, do not refund one, and do not move the card again unless its own text says to.

THE ONLY VERBS YOU MAY USE
- move_card — move one card between zones.
  params: {"instance_id": "<uuid>", "src": {"kind": "<zone>", "owner": "<player uuid>"}, "dst": {"kind": "<zone>", "owner": "<player uuid>"}}
  zones: hand, battlefield, graveyard, exile, command, library. The battlefield is shared — omit "owner" for it.
- change_life — change one player's life total. Put the player's id in the step's "player" field, not in params.
  params: {"delta": -3}
- add_counter — put counters on a permanent, or take them off with a negative delta.
  params: {"instance_id": "<uuid>", "name": "+1/+1", "delta": 1}
- mark_damage — mark damage on a creature.
  params: {"instance_id": "<uuid>", "delta": 3}

Any other verb makes the whole bundle invalid and NOTHING happens.

RULES
- Use only the instance_id and player values listed in the board below. An id you invent throws the whole bundle away.
- At most 12 steps. A card that needs more than twelve is a card to decline.
- Do exactly what the card's text says and nothing more. No extra value, nothing the card does not reach.
- You cannot see the contents of any library, so you cannot name a card in one. A "draw a card", "search your library" or "look at the top" clause CANNOT be improvised. Leave it out, and say in "effect" which part you left out.
- If the spell did not really resolve — it was countered, its only target is gone — or if you cannot do the card faithfully with these verbs, answer with an empty "steps" list. That is a normal, cheap, correct answer and it is better than a wrong bundle.
- Everything you do is announced in chat, named as an improvisation, and ANY player can undo the whole bundle for free. Write "effect" as the plain-words sentence the table will be shown.

HOW TO ANSWER
Reply with a single JSON object and nothing else:
  {"effect": "<at most twenty words: what you did>", "why": "<at most twelve words>", "steps": [{"verb": "change_life", "player": "<player uuid>", "params": {"delta": -3}}]}
No preamble, no code fence, no text outside the JSON, and no internal or system XML tags.`

// Improvisation outcomes, as a record and a Stats key. Exactly one is
// counted per window in which this seat had a card it could have
// improvised.
const (
	// ImprovApplied — a bundle was handed to the runner. Whether the
	// rail then accepted it is the RUNNER's counter
	// (aiseat.Stats.Improvisations / ImprovRefused); this side only
	// knows what it offered.
	ImprovOffered = "offered"
	// ImprovDeclined — the model answered with no steps. A
	// first-class answer: the spell did not really resolve, or the
	// card cannot be done faithfully with four verbs.
	ImprovDeclined = "declined"
	// ImprovNoOracle — the seat's deck profile has no oracle text for
	// the card, so there is nothing to improvise FROM. A seat whose
	// deck profile did not load plays with no improvisation at all,
	// which is the old behaviour and a safe one.
	ImprovNoOracle = "no-oracle"
	// ImprovNoBudget — the runner's deadline left too little time to
	// try, or the per-game call cap is spent. No call was made.
	ImprovNoBudget = "no-budget"
	// ImprovCapped — MaxImprovCalls is spent for this seat.
	ImprovCapped = "capped"
	// ImprovError — the call failed: outage, HTTP error, timeout.
	ImprovError = "error"
	// ImprovMalformed — the reply was not a bundle.
	ImprovMalformed = "malformed"
)

// defaultMaxImprovCalls is the hard per-game, per-seat cap on
// improvisation model calls. It counts CALLS rather than applied
// bundles because the cap exists to bound spend, and a call that
// failed cost the same as one that worked.
//
// Eight is chosen against the deck rather than against a price: a
// bot's deck is curated and catalog-only (ADR 0033 §7), so a seat
// that reaches this number in one game is a seat whose deck has
// outrun the catalog — which is worth a look, not worth more calls.
// Together with one-shot-per-instance it is a closed bound on what
// improvisation can cost a game, which is what #735 needs something
// to measure against.
const defaultMaxImprovCalls = 8

// resolvedSpell is one of this seat's own uncatalogued spells, caught
// on its way off the stack.
type resolvedSpell struct {
	// ID is the card's instance ID.
	ID string
	// Name is what the table calls it, and what the announcement
	// must name.
	Name string
	// Where is the zone it landed in, in words for the prompt.
	Where string
	// Found is false when the card is nowhere the seat can see it —
	// shuffled into a library, exiled face down. There is then no
	// honest "it is now in ..." line, and no improvisation.
	Found bool
}

// improvTracker is the seat's memory of its own spells. It is the
// only state this package keeps across decisions, and it keeps the
// smallest thing that answers the question: which of my uncatalogued
// spells were on the stack a moment ago and are not now.
//
// Per seat, so the lock is uncontended in practice; it is here
// because a Policy is documented as safe from one goroutine at a time
// and Stats is read from another.
type improvTracker struct {
	mu sync.Mutex
	// onStack is the ids currently being watched, in the order they
	// were first seen, so two spells resolving in one window come
	// back in a deterministic order.
	onStack []resolvedSpell
	// resolved is the queue of spells that have left the stack and
	// have not yet been offered.
	resolved []resolvedSpell
	// seen is every id this tracker has ever admitted, so a card
	// cannot re-enter the queue — one improvisation per instance, for
	// the life of the seat.
	seen  map[string]bool
	calls int
}

func newImprovTracker() *improvTracker {
	return &improvTracker{seen: map[string]bool{}}
}

// observe updates the tracking from one view. Called on every window,
// including the ones that will not improvise: missing a spell's
// passage across the stack is how the feature silently stops working.
func (t *improvTracker) observe(v *protocol.GameView, me string) {
	stack := make(map[string]*protocol.CardView, len(v.Stack.Cards))
	for i := range v.Stack.Cards {
		stack[v.Stack.Cards[i].InstanceID] = &v.Stack.Cards[i]
	}

	t.mu.Lock()
	defer t.mu.Unlock()

	// Admit this seat's uncatalogued spells. Requiring the source to
	// be a CARD ON THE STACK is what confines this to spells: a
	// triggered or activated ability's source sits on the
	// battlefield, so it is never in this map.
	for i := range v.StackItems {
		it := &v.StackItems[i]
		if it.Controller != me {
			continue
		}
		c := stack[it.SourceCardID]
		if c == nil || !c.Unimplemented || t.seen[c.InstanceID] {
			continue
		}
		t.seen[c.InstanceID] = true
		t.onStack = append(t.onStack, resolvedSpell{ID: c.InstanceID, Name: cardName(c)})
	}

	// Promote the ones that have left.
	keep := t.onStack[:0]
	for _, sp := range t.onStack {
		if stack[sp.ID] != nil {
			keep = append(keep, sp)
			continue
		}
		sp.Where, sp.Found = whereIs(v, sp.ID, me)
		t.resolved = append(t.resolved, sp)
	}
	t.onStack = keep
}

// take pops the next spell to improvise, or reports none. Popping is
// the one-shot rule: a spell offered here is never offered again,
// whatever the caller then does with it.
func (t *improvTracker) take() (resolvedSpell, bool) {
	t.mu.Lock()
	defer t.mu.Unlock()
	for len(t.resolved) > 0 {
		sp := t.resolved[0]
		t.resolved = t.resolved[1:]
		if sp.Found {
			return sp, true
		}
	}
	return resolvedSpell{}, false
}

// spend claims one of the per-game call budget, reporting false when
// the cap is already gone.
func (t *improvTracker) spend(cap int) bool {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.calls >= cap {
		return false
	}
	t.calls++
	return true
}

// whereIs locates a card in the seat's own view and says where in
// words. The seat cannot see a library's contents, so a card shuffled
// away is simply not found — and a spell nobody can point at is not
// one to improvise.
func whereIs(v *protocol.GameView, id, me string) (string, bool) {
	for i := range v.Battlefield.Cards {
		if v.Battlefield.Cards[i].InstanceID == id {
			return "on the battlefield", true
		}
	}
	for i := range v.Exile.Cards {
		if v.Exile.Cards[i].InstanceID == id {
			return "in exile", true
		}
	}
	for i := range v.Seats {
		s := &v.Seats[i]
		who := "their"
		if s.ID == me {
			who = "your"
		}
		for _, z := range []struct {
			name  string
			cards []protocol.CardView
		}{
			{"graveyard", s.Graveyard.Cards},
			{"hand", s.Hand.Cards},
			{"command zone", s.Command.Cards},
		} {
			for j := range z.cards {
				if z.cards[j].InstanceID == id {
					return "in " + who + " " + z.name, true
				}
			}
		}
	}
	return "", false
}

// Compile-time proof that the funnel really is an Improviser, which
// is what gives `assisted` and `strong` the behaviour: tiers.New
// returns this type for both, so nothing else has to be wired.
var _ aiseat.Improviser = (*Policy)(nil)

// Improvise implements aiseat.Improviser. ADR 0033 §8.
//
// Returning false is the overwhelmingly common answer and is the
// answer on every failure: an outage, a timeout, a reply that is not
// a bundle and a model that declines all come back the same way, and
// all of them cost the table nothing but the latency already spent.
// The runner then asks for an ordinary move, exactly as if this were
// not implemented.
//
// What it does NOT do is police the verb list. A bundle naming
// something outside the four is handed up and refused by
// aiseat.Improvisation.Validate — the rail §8 built and the tests
// pin — which announces the refusal to the table. One enforcement
// point on the path every improviser must cross is worth more than a
// second copy of the allow-list in the one policy that has it today.
func (p *Policy) Improvise(ctx context.Context, in aiseat.Input) (aiseat.Improvisation, bool) {
	// Tracking runs on every window, including the ones that cannot
	// improvise: it is a diff across views, and a skipped view is a
	// spell that crossed the stack unseen.
	me := in.Seat.String()
	p.improv.observe(&in.View, me)

	if !p.cfg.Improvise || p.cfg.Client == nil {
		return aiseat.Improvisation{}, false
	}
	// A seat that owes a choice is answering a prompt that halts the
	// table (#1000, #1016). The queue keeps; this is a "not now", not
	// a "never".
	if owesChoice(&in.View, me) {
		return aiseat.Improvisation{}, false
	}
	sp, ok := p.improv.take()
	if !ok {
		return aiseat.Improvisation{}, false
	}

	card, ok := p.deckCard(sp.Name)
	if !ok || strings.TrimSpace(card.Oracle) == "" {
		// Nothing to improvise FROM. Announcing this would be telling
		// the table about the server's own missing decklist, which
		// is not news about the game.
		p.rec.improv(ImprovNoOracle, 0, Usage{})
		p.cfg.Log.Info("bot cannot improvise a card it has no oracle text for",
			"tier", p.cfg.Tier, "card", sp.Name)
		return aiseat.Improvisation{}, false
	}

	budget := p.budget(ctx)
	if budget < p.cfg.MinBudget {
		p.rec.improv(ImprovNoBudget, 0, Usage{})
		return aiseat.Improvisation{}, false
	}
	if !p.improv.spend(p.cfg.MaxImprovCalls) {
		p.rec.improv(ImprovCapped, 0, Usage{})
		p.cfg.Log.Warn("bot has spent its improvisation budget for this game; the card was left alone",
			"tier", p.cfg.Tier, "card", sp.Name, "cap", p.cfg.MaxImprovCalls)
		return aiseat.Improvisation{}, false
	}

	profile := p.cfg.Improv
	req := Request{
		Model:     profile.ID,
		System:    []Block{{Text: improvPrimer, Cache: true}},
		User:      p.buildImprovPrompt(in, sp, card),
		MaxTokens: profile.MaxTokens,
		Effort:    profile.Effort,
		Thinking:  profile.Thinking,
	}

	callCtx, cancel := context.WithTimeout(ctx, budget)
	defer cancel()
	started := time.Now()
	resp, err := p.cfg.Client.Complete(callCtx, req)
	latency := time.Since(started)
	if err != nil {
		p.rec.improv(ImprovError, latency, resp.Usage)
		p.cfg.Log.Warn("bot improvisation call failed; the card was left alone",
			"tier", p.cfg.Tier, "model", profile.ID, "card", sp.Name, "took", latency, "err", err)
		return aiseat.Improvisation{}, false
	}

	im, perr := parseImprovisation(resp.Text, sp.Name, profile.ID)
	switch {
	case perr != nil:
		p.rec.improv(ImprovMalformed, latency, resp.Usage)
		p.cfg.Log.Warn("bot improvisation reply was not a bundle; the card was left alone",
			"tier", p.cfg.Tier, "model", profile.ID, "card", sp.Name, "reply", Truncate(resp.Text, 200))
		return aiseat.Improvisation{}, false
	case len(im.Steps) == 0:
		// The model's way of saying "this spell did not really
		// resolve" or "I cannot do this faithfully". Silent by
		// design: nothing was attempted against the board, so there
		// is nothing about the game to disclose.
		p.rec.improv(ImprovDeclined, latency, resp.Usage)
		p.cfg.Log.Info("bot declined to improvise a card it could not run",
			"tier", p.cfg.Tier, "card", sp.Name, "why", Truncate(im.Effect, 120))
		return aiseat.Improvisation{}, false
	}

	p.rec.improv(ImprovOffered, latency, resp.Usage)
	p.cfg.Log.Info("bot is improvising an uncatalogued card",
		"tier", p.cfg.Tier, "model", profile.ID, "card", im.Card,
		"effect", im.Effect, "steps", len(im.Steps), "took", latency)
	return im, true
}

// owesChoice reports whether the seat has a pending choice addressed
// to it.
func owesChoice(v *protocol.GameView, me string) bool {
	for i := range v.PendingChoices {
		if v.PendingChoices[i].Chooser == me {
			return true
		}
	}
	return false
}

// deckCard finds a card in this seat's own deck profile by name.
func (p *Policy) deckCard(name string) (DeckCard, bool) {
	c, ok := p.deckIndex[strings.ToLower(strings.TrimSpace(name))]
	return c, ok
}

// --- the prompt -----------------------------------------------------

// buildImprovPrompt renders the per-call half: the card, where it
// went, and the board with the IDs a bundle may name.
func (p *Policy) buildImprovPrompt(in aiseat.Input, sp resolvedSpell, card DeckCard) string {
	v := &in.View
	me := in.Seat.String()
	var b strings.Builder

	b.WriteString("THE CARD THE ENGINE DID NOT RUN\n")
	b.WriteString(card.Name)
	if card.Cost != "" {
		b.WriteString(" " + card.Cost)
	}
	if card.Type != "" {
		b.WriteString(" — " + card.Type)
	}
	b.WriteByte('\n')
	fmt.Fprintf(&b, "Oracle text: %s\n", OneLine(card.Oracle))
	fmt.Fprintf(&b, "You cast it. It has just left the stack and is now %s (instance_id=%s).\n", sp.Where, sp.ID)
	b.WriteString("Nothing happened when it resolved.\n")

	fmt.Fprintf(&b, "\nTURN %d — %s\n", v.Turn.Number, stepName(v.Turn.Step))

	b.WriteString("\nPLAYERS\n")
	for i := range v.Seats {
		s := &v.Seats[i]
		if s.Eliminated {
			fmt.Fprintf(&b, "  %s — player=%s — ELIMINATED\n", seatLabel(s, me), s.ID)
			continue
		}
		fmt.Fprintf(&b, "  %s — player=%s — %d life, %d cards in hand, %d in library\n",
			seatLabel(s, me), s.ID, s.Life, s.Hand.Count, s.Library.Count)
	}

	b.WriteString("\nBATTLEFIELD\n")
	wrote := false
	for i := range v.Battlefield.Cards {
		c := &v.Battlefield.Cards[i]
		if i >= p.cfg.MaxZoneCards {
			b.WriteString("  … (further permanents not listed)\n")
			break
		}
		fmt.Fprintf(&b, "  instance_id=%s — %s — controlled by %s\n",
			c.InstanceID, permanentLabel(c), seatLabelByID(v, c.Controller, me))
		wrote = true
	}
	if !wrote {
		b.WriteString("  (empty)\n")
	}

	for i := range v.Seats {
		s := &v.Seats[i]
		mine := s.ID == me
		p.writeIDZone(&b, fmt.Sprintf("%s — GRAVEYARD", strings.ToUpper(seatLabel(s, me))), s.Graveyard.Cards)
		if mine {
			p.writeIDZone(&b, "YOUR HAND", s.Hand.Cards)
			p.writeIDZone(&b, "YOUR COMMAND ZONE", s.Command.Cards)
		}
	}
	p.writeIDZone(&b, "EXILE", v.Exile.Cards)

	if len(v.StackItems) > 0 {
		b.WriteString("\nSTILL ON THE STACK (bottom first)\n")
		for i := range v.StackItems {
			it := &v.StackItems[i]
			fmt.Fprintf(&b, "  %s — cast by %s\n", stackLabel(it, v), seatLabelByID(v, it.Controller, me))
		}
	}

	b.WriteString("\nWrite the bundle that applies the card's text, or an empty steps list.")
	return b.String()
}

// writeIDZone renders one zone as instance_id → card, or nothing at
// all when the zone is empty. A heading with "(empty)" under it on
// four zones per seat is most of the prompt for no information.
func (p *Policy) writeIDZone(b *strings.Builder, heading string, cards []protocol.CardView) {
	if len(cards) == 0 {
		return
	}
	fmt.Fprintf(b, "\n%s\n", heading)
	for i := range cards {
		if i >= p.cfg.MaxZoneCards {
			b.WriteString("  … (rest not listed)\n")
			return
		}
		c := &cards[i]
		fmt.Fprintf(b, "  instance_id=%s — %s", c.InstanceID, cardName(c))
		if c.TypeLine != "" {
			fmt.Fprintf(b, " — %s", c.TypeLine)
		}
		b.WriteByte('\n')
	}
}

// --- parsing the reply ----------------------------------------------

// improvReply is the wire shape asked for in the primer.
type improvReply struct {
	Effect string            `json:"effect"`
	Why    string            `json:"why"`
	Steps  []improvReplyStep `json:"steps"`
}

// improvReplyStep is one step. `verb` is what the primer asks for and
// `type` is accepted beside it, because that is the field name the
// rest of the wire uses and a model that has seen the dispatcher's
// vocabulary reaches for it — the two are the same thing and refusing
// one would be pedantry paid for in dropped bundles.
type improvReplyStep struct {
	Verb   string          `json:"verb"`
	Type   string          `json:"type"`
	Player string          `json:"player"`
	Params json.RawMessage `json:"params"`
}

// parseImprovisation turns a reply into a bundle. Tolerant of
// packaging (a fence, a sentence either side) for the same reason
// parseAnswer is, and deliberately NOT tolerant of content: a verb
// outside the four, a params object the dispatcher will not take and
// an id that is not a card all travel intact to the rail, which
// refuses the bundle and tells the table it did.
//
// card is the name this seat's tracker caught leaving the stack, and
// it is what the announcement names — never a name out of the reply.
// A model that answered about a different card would otherwise
// announce a card it did not touch, which is the disclosure failing
// in the most confusing possible way.
func parseImprovisation(text, card, model string) (aiseat.Improvisation, error) {
	obj := firstJSONObject(strings.TrimSpace(text))
	if obj == "" {
		return aiseat.Improvisation{}, fmt.Errorf("no JSON object in reply %q", Truncate(text, 120))
	}
	var r improvReply
	if err := json.Unmarshal([]byte(obj), &r); err != nil {
		return aiseat.Improvisation{}, fmt.Errorf("reply is not a bundle: %w", err)
	}
	im := aiseat.Improvisation{
		Card:   card,
		Effect: OneLine(strings.TrimSpace(r.Effect)),
		Reason: modelReason(model, r.Why),
	}
	for _, s := range r.Steps {
		verb := strings.TrimSpace(s.Verb)
		if verb == "" {
			verb = strings.TrimSpace(s.Type)
		}
		step := aiseat.ImprovStep{Type: verb, Params: s.Params}
		// A player id that will not parse stays uuid.Nil rather than
		// dropping the step: Validate refuses a nil-player
		// change_life, and being refused loudly beats a bundle
		// quietly missing a leg.
		if id, err := uuid.Parse(strings.TrimSpace(s.Player)); err == nil {
			step.Player = id
		}
		im.Steps = append(im.Steps, step)
	}
	return im, nil
}
