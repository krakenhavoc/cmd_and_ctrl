package model

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// improvise_test.go covers ADR 0033 §8's improviser with a fake model
// and no network. Every canned reply in here is one the endpoint
// really can produce — a good bundle, a bundle reaching for a verb it
// may not have, a reply that is not JSON, and a call that never
// comes back — because the whole safety argument is about what
// happens when a model is wrong, and a test suite that only feeds it
// good answers proves nothing about that.
//
// The live counterpart is improvise_live_test.go, which needs a key
// and an env var CI never sets.

const (
	grimID = "aaaaaaaa-0000-4000-8000-000000000001"
	hexID  = "aaaaaaaa-0000-4000-8000-000000000002"
	bearID = "aaaaaaaa-0000-4000-8000-000000000003"
)

// improvDeck is testDeck plus two cards the catalog cannot run. They
// are what gives the improviser oracle text to work from: a seat's
// own DeckProfile is the ONLY source, because a single-faced CardView
// carries no oracle text on the wire.
func improvDeck() DeckProfile {
	d := testDeck()
	d.Cards = append(d.Cards,
		DeckCard{
			Name: "Grim Tutor", Cost: "{1}{B}{B}", Type: "Sorcery",
			Oracle: "Search your library for a card, put it into your hand, then shuffle. You lose 3 life.",
		},
		DeckCard{
			Name: "Hex Drain", Cost: "{2}{B}", Type: "Instant",
			Oracle: "Each opponent loses 2 life and you gain 2 life.",
		},
	)
	return d
}

// stackCard is one of this seat's own spells sitting on the stack,
// flagged as something the engine will not run.
func stackCard(id, name, typeLine string) protocol.CardView {
	return protocol.CardView{
		InstanceID: id, Name: name, TypeLine: typeLine,
		Owner: meSeat.String(), Controller: meSeat.String(),
		Unimplemented: true,
	}
}

// castingView is the window right after this seat cast `name`: the
// spell is on the stack and the seat has priority back.
func castingView(id, name, typeLine string) protocol.GameView {
	v := view()
	c := stackCard(id, name, typeLine)
	v.Stack = protocol.ZoneView{Kind: "stack", Count: 1, Cards: []protocol.CardView{c}}
	v.StackItems = []protocol.StackItemView{{
		ID: "item-" + id, Kind: "spell", Controller: meSeat.String(),
		Owner: meSeat.String(), SourceCardID: id, Label: name,
	}}
	return v
}

// resolvedInGraveyard is the next window: the spell resolved into
// silence and is in its owner's graveyard.
func resolvedInGraveyard(id, name, typeLine string) protocol.GameView {
	v := view()
	c := stackCard(id, name, typeLine)
	v.Seats[0].Graveyard = protocol.ZoneView{
		Kind: "graveyard", Owner: meSeat.String(), Count: 1,
		Cards: []protocol.CardView{c},
	}
	// Something for a bundle to point at.
	v.Battlefield = protocol.ZoneView{Kind: "battlefield", Count: 1, Cards: []protocol.CardView{{
		InstanceID: bearID, Name: "Bear", TypeLine: "Creature — Bear",
		Owner: oppSeat.String(), Controller: oppSeat.String(), Power: 2, Toughness: 2,
	}}}
	return v
}

func improvInput(v protocol.GameView) aiseat.Input {
	return aiseat.Input{Seat: meSeat, View: v, Moves: []legal.Move{pass()}}
}

// improvPolicy is a funnel with a deck profile that has oracle text
// and improvisation on.
func improvPolicy(t *testing.T, client Client, tweak func(*Config)) *Policy {
	t.Helper()
	return testPolicy(t, client, &stubB{index: 0, reason: "stub"}, func(c *Config) {
		c.Deck = improvDeck()
		if tweak != nil {
			tweak(c)
		}
	})
}

// improvise runs one window with a real deadline on it, the way the
// runner does.
func improvise(t *testing.T, p *Policy, in aiseat.Input, budget time.Duration) (aiseat.Improvisation, bool) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), budget)
	defer cancel()
	return p.Improvise(ctx, in)
}

// bundleReply is a well-formed answer: a drain the card really does.
func bundleReply() *FakeClient {
	return &FakeClient{
		Label: "improv-bundle",
		Reply: func(int, Request) (Response, error) {
			return Response{
				Text: fmt.Sprintf(`{"effect": "lose 3 life (the library search cannot be improvised)",
				  "why": "no catalog spec",
				  "steps": [{"verb": "change_life", "player": %q, "params": {"delta": -3}}]}`, meSeat),
				Usage: Usage{InputTokens: 900, OutputTokens: 60},
			}, nil
		},
	}
}

// --- the window ------------------------------------------------------

// The shape of the whole feature in one test: a spell crosses the
// stack, resolves into silence, and the seat offers a bundle for it —
// and NOT before it resolves, because the engine has to charge the
// mana first.
func TestUncataloguedSpellIsImprovisedAfterItResolves(t *testing.T) {
	fake := bundleReply()
	p := improvPolicy(t, fake, nil)

	if _, ok := improvise(t, p, improvInput(castingView(grimID, "Grim Tutor", "Sorcery")), 2*time.Second); ok {
		t.Fatal("improvised a spell that is still on the stack — the engine has not charged for it yet")
	}
	if fake.Calls() != 0 {
		t.Errorf("the model was called %d times while the spell was still on the stack", fake.Calls())
	}

	im, ok := improvise(t, p, improvInput(resolvedInGraveyard(grimID, "Grim Tutor", "Sorcery")), 2*time.Second)
	if !ok {
		t.Fatal("the spell resolved into silence and nothing was improvised")
	}
	if fake.Calls() != 1 {
		t.Errorf("model calls = %d, want exactly 1", fake.Calls())
	}
	if im.Card != "Grim Tutor" {
		t.Errorf("Card = %q; the announcement must name the card the TRACKER caught, never one out of the reply", im.Card)
	}
	if err := im.Validate(); err != nil {
		t.Fatalf("the bundle does not pass the rail it has to cross: %v", err)
	}
	if len(im.Steps) != 1 || im.Steps[0].Type != aiseat.VerbChangeLife || im.Steps[0].Player != meSeat {
		t.Errorf("steps = %+v", im.Steps)
	}
	if !strings.Contains(im.Announcement(), "lose 3 life") {
		t.Errorf("the announcement does not carry the effect: %q", im.Announcement())
	}
	if st := p.Stats(); st.ByImprov[ImprovOffered] != 1 || st.ImprovCalls != 1 || st.ImprovUsage.InputTokens != 900 {
		t.Errorf("stats = %+v", st)
	}
}

// One improvisation per card instance, for the life of the seat. It
// is half of what makes the spend bounded (the per-game cap is the
// other half), and it is what stops a card sitting in a graveyard
// being re-improvised on every window for the rest of the game.
func TestOneImprovisationPerCardInstance(t *testing.T) {
	fake := bundleReply()
	p := improvPolicy(t, fake, nil)

	improvise(t, p, improvInput(castingView(grimID, "Grim Tutor", "Sorcery")), 2*time.Second)
	resolved := improvInput(resolvedInGraveyard(grimID, "Grim Tutor", "Sorcery"))
	if _, ok := improvise(t, p, resolved, 2*time.Second); !ok {
		t.Fatal("the first offer did not happen")
	}
	for i := 0; i < 5; i++ {
		if _, ok := improvise(t, p, resolved, 2*time.Second); ok {
			t.Fatalf("offered the same card instance again on window %d", i+2)
		}
	}
	if fake.Calls() != 1 {
		t.Errorf("model calls = %d, want 1 — the card was improvised more than once", fake.Calls())
	}
}

// An opponent's uncatalogued spell is never this seat's to apply.
// The nil-caller grant in §8 exists so a bot's OWN card can reach
// another board, not so a bot can run somebody else's card.
func TestAnOpponentsUncataloguedSpellIsNeverImprovised(t *testing.T) {
	fake := bundleReply()
	p := improvPolicy(t, fake, nil)

	onStack := castingView(grimID, "Grim Tutor", "Sorcery")
	onStack.Stack.Cards[0].Controller = oppSeat.String()
	onStack.Stack.Cards[0].Owner = oppSeat.String()
	onStack.StackItems[0].Controller = oppSeat.String()
	onStack.StackItems[0].Owner = oppSeat.String()

	gone := resolvedInGraveyard(grimID, "Grim Tutor", "Sorcery")
	gone.Seats[0].Graveyard = protocol.ZoneView{Kind: "graveyard"}
	gone.Seats[1].Graveyard = protocol.ZoneView{Kind: "graveyard", Count: 1,
		Cards: []protocol.CardView{stackCard(grimID, "Grim Tutor", "Sorcery")}}

	improvise(t, p, improvInput(onStack), 2*time.Second)
	if _, ok := improvise(t, p, improvInput(gone), 2*time.Second); ok {
		t.Fatal("improvised an opponent's spell")
	}
	if fake.Calls() != 0 {
		t.Errorf("the model was asked about an opponent's card %d times", fake.Calls())
	}
}

// An unimplemented PERMANENT's triggered ability is not a window.
// There is no wire signal for "this trigger should have fired", the
// condition recurs, and a bound on it would be a rate limit rather
// than a rule. The mechanical expression of that is the source having
// to be a card ON THE STACK: an ability's source sits on the
// battlefield.
func TestAnUnimplementedPermanentsAbilityIsNotAWindow(t *testing.T) {
	fake := bundleReply()
	p := improvPolicy(t, fake, nil)

	v := view()
	v.Battlefield = protocol.ZoneView{Kind: "battlefield", Count: 1, Cards: []protocol.CardView{{
		InstanceID: hexID, Name: "Hex Drain", TypeLine: "Enchantment",
		Owner: meSeat.String(), Controller: meSeat.String(), Unimplemented: true,
	}}}
	// The ability is on the stack; its source is the permanent.
	v.StackItems = []protocol.StackItemView{{
		ID: "trig", Kind: "trigger", Controller: meSeat.String(),
		Owner: meSeat.String(), SourceCardID: hexID, Label: "Hex Drain trigger",
	}}

	improvise(t, p, improvInput(v), 2*time.Second)
	after := view()
	after.Battlefield = v.Battlefield
	if _, ok := improvise(t, p, improvInput(after), 2*time.Second); ok {
		t.Fatal("improvised a permanent's triggered ability")
	}
	if fake.Calls() != 0 {
		t.Errorf("the model was called %d times for a trigger", fake.Calls())
	}
}

// A seat that owes a pending choice is answering a prompt that halts
// the table (#1000's OwedInStep, #1016's GuardsStackItem). It is a
// "not now", not a "never": the queue keeps, and the card is
// improvised on the next clean window.
func TestAPendingChoiceHoldsTheImprovisationRatherThanLosingIt(t *testing.T) {
	fake := bundleReply()
	p := improvPolicy(t, fake, nil)

	improvise(t, p, improvInput(castingView(grimID, "Grim Tutor", "Sorcery")), 2*time.Second)

	owing := resolvedInGraveyard(grimID, "Grim Tutor", "Sorcery")
	owing.PendingChoices = []protocol.PendingChoiceView{{
		ID: "c1", Kind: "pay_or_else", Chooser: meSeat.String(), Count: 1,
		Reason: "pay {2} or sacrifice",
	}}
	if _, ok := improvise(t, p, improvInput(owing), 2*time.Second); ok {
		t.Fatal("improvised while the seat owed a halting prompt")
	}
	if fake.Calls() != 0 {
		t.Errorf("the model was called %d times while a prompt was open", fake.Calls())
	}

	if _, ok := improvise(t, p, improvInput(resolvedInGraveyard(grimID, "Grim Tutor", "Sorcery")), 2*time.Second); !ok {
		t.Fatal("the card was lost rather than held while the prompt was open")
	}
}

// A card the seat's own deck profile has no oracle text for is a card
// with nothing to improvise FROM. That is a thin prompt, not a
// crash — a server with no Scryfall dump builds name-only profiles
// (deckprofile.Build) and must still seat a bot.
func TestNoOracleTextMeansNoImprovisation(t *testing.T) {
	fake := bundleReply()
	p := improvPolicy(t, fake, func(c *Config) { c.Deck = testDeck() }) // no Grim Tutor

	improvise(t, p, improvInput(castingView(grimID, "Grim Tutor", "Sorcery")), 2*time.Second)
	if _, ok := improvise(t, p, improvInput(resolvedInGraveyard(grimID, "Grim Tutor", "Sorcery")), 2*time.Second); ok {
		t.Fatal("improvised a card it has no oracle text for")
	}
	if fake.Calls() != 0 {
		t.Errorf("the model was called %d times with no oracle text to send it", fake.Calls())
	}
	if st := p.Stats(); st.ByImprov[ImprovNoOracle] != 1 {
		t.Errorf("stats = %+v", st.ByImprov)
	}
}

// --- the prompt -------------------------------------------------------

// The prompt has to carry the card's oracle text and the exact IDs a
// bundle is allowed to name — and nothing that is not already on the
// wire to this seat.
func TestImprovisationPromptCarriesOracleTextAndTheIDsItMayName(t *testing.T) {
	fake := bundleReply()
	p := improvPolicy(t, fake, nil)
	improvise(t, p, improvInput(castingView(grimID, "Grim Tutor", "Sorcery")), 2*time.Second)
	improvise(t, p, improvInput(resolvedInGraveyard(grimID, "Grim Tutor", "Sorcery")), 2*time.Second)

	reqs := fake.Requests()
	if len(reqs) != 1 {
		t.Fatalf("calls = %d, want 1", len(reqs))
	}
	req := reqs[0]
	if len(req.System) != 1 || !req.System[0].Cache {
		t.Fatalf("system blocks = %+v, want one cached primer", req.System)
	}
	for _, want := range []string{"move_card", "change_life", "add_counter", "mark_damage", "empty \"steps\" list"} {
		if !strings.Contains(req.System[0].Text, want) {
			t.Errorf("the primer never mentions %q", want)
		}
	}
	for _, want := range []string{
		"Grim Tutor",
		"Search your library for a card", // the oracle text
		grimID,                           // the card itself
		bearID,                           // something a bundle could point at
		meSeat.String(), oppSeat.String(),
		"in your graveyard",
	} {
		if !strings.Contains(req.User, want) {
			t.Errorf("the prompt is missing %q\n---\n%s", want, req.User)
		}
	}
	// The improvisation primer is a DIFFERENT prefix from the
	// decision primer. Sharing would buy no cache — a prefix match is
	// all-or-nothing — and would tell the model to answer with one
	// number.
	if strings.Contains(req.System[0].Text, "ONE NUMBER") {
		t.Error("the improvisation primer is carrying the decision primer's instructions")
	}
}

// The improvisation call asks the FRONTIER slot. Writing a bundle
// from oracle text is the hardest thing the seat is ever asked to do
// and the rarest; the cheap model is the wrong economy here.
func TestImprovisationAsksTheFrontierModel(t *testing.T) {
	fake := bundleReply()
	p := improvPolicy(t, fake, func(c *Config) {
		c.Routine.ID = "cheap-model"
		c.Frontier.ID = "frontier-model"
		c.Improv.ID = ""
	})
	improvise(t, p, improvInput(castingView(grimID, "Grim Tutor", "Sorcery")), 2*time.Second)
	improvise(t, p, improvInput(resolvedInGraveyard(grimID, "Grim Tutor", "Sorcery")), 2*time.Second)

	reqs := fake.Requests()
	if len(reqs) != 1 {
		t.Fatalf("calls = %d, want 1", len(reqs))
	}
	if reqs[0].Model != "frontier-model" {
		t.Errorf("improvisation asked %q, want the frontier model", reqs[0].Model)
	}
	if reqs[0].MaxTokens <= 256 {
		t.Errorf("MaxTokens = %d; a bundle needs more room than an index", reqs[0].MaxTokens)
	}
}

// --- the four canned failures -----------------------------------------

// A verb outside the closed list is NOT filtered out here. It travels
// to aiseat.Improvisation.Validate — the rail §8 built, which every
// improviser must cross and which announces the refusal — so the
// allow-list is enforced in exactly one place.
func TestAForbiddenVerbTravelsToTheRailAndIsRefusedThere(t *testing.T) {
	fake := &FakeClient{Reply: func(int, Request) (Response, error) {
		return Response{Text: fmt.Sprintf(`{"effect": "scoop them out", "why": "why not",
		  "steps": [
		    {"verb": "change_life", "player": %q, "params": {"delta": -2}},
		    {"verb": "concede", "player": %q, "params": {}}
		  ]}`, oppSeat, oppSeat)}, nil
	}}
	p := improvPolicy(t, fake, nil)
	improvise(t, p, improvInput(castingView(hexID, "Hex Drain", "Instant")), 2*time.Second)
	im, ok := improvise(t, p, improvInput(resolvedInGraveyard(hexID, "Hex Drain", "Instant")), 2*time.Second)
	if !ok {
		t.Fatal("the bundle was dropped in the model package; the rail must be the thing that refuses it")
	}
	if err := im.Validate(); !errors.Is(err, aiseat.ErrImprovVerb) {
		t.Fatalf("Validate = %v, want ErrImprovVerb", err)
	}
	// And the table gets told, naming the card.
	line := im.RefusalAnnouncement()
	if !strings.Contains(line, "Hex Drain") || !strings.Contains(line, "nothing was changed") {
		t.Errorf("the refusal line is not usable disclosure: %q", line)
	}
}

// A reply that is not a bundle is dropped. Nothing was attempted
// against the board, so nothing is announced either.
func TestMalformedImprovisationReplyIsDropped(t *testing.T) {
	for name, text := range map[string]string{
		"prose":       "I would destroy the Bear, I think.",
		"truncated":   `{"effect": "drain", "steps": [{"verb": "change_life",`,
		"empty":       "",
		"wrong shape": `{"effect": "drain", "steps": "all of them"}`,
	} {
		t.Run(name, func(t *testing.T) {
			p := improvPolicy(t, AlwaysText(text), nil)
			improvise(t, p, improvInput(castingView(hexID, "Hex Drain", "Instant")), 2*time.Second)
			if _, ok := improvise(t, p, improvInput(resolvedInGraveyard(hexID, "Hex Drain", "Instant")), 2*time.Second); ok {
				t.Fatal("a reply that is not a bundle produced one")
			}
			if st := p.Stats(); st.ByImprov[ImprovMalformed] != 1 {
				t.Errorf("stats = %+v", st.ByImprov)
			}
		})
	}
}

// A call that never comes back costs the table its budget and nothing
// else. ADR 0033 §10: the table never waits on a bot, and the way to
// honour that with a network call inside the deadline is to have
// nothing riding on the answer.
func TestAnImprovisationThatTimesOutIsDropped(t *testing.T) {
	p := improvPolicy(t, NeverAnswers(), func(c *Config) {
		c.MinBudget = 10 * time.Millisecond
		c.MaxCall = 80 * time.Millisecond
	})
	improvise(t, p, improvInput(castingView(hexID, "Hex Drain", "Instant")), 2*time.Second)

	started := time.Now()
	_, ok := improvise(t, p, improvInput(resolvedInGraveyard(hexID, "Hex Drain", "Instant")), 2*time.Second)
	took := time.Since(started)
	if ok {
		t.Fatal("a call that never answered produced a bundle")
	}
	if took > time.Second {
		t.Errorf("the improvisation held the seat for %v; MaxCall is 80ms", took)
	}
	if st := p.Stats(); st.ByImprov[ImprovError] != 1 {
		t.Errorf("stats = %+v", st.ByImprov)
	}
}

// An outage is the same path, and is what the S31 outage drill is
// about: a seat whose endpoint has gone away plays on, it just never
// improvises.
func TestAnOutageDropsTheImprovisation(t *testing.T) {
	p := improvPolicy(t, AlwaysFails(nil), nil)
	improvise(t, p, improvInput(castingView(hexID, "Hex Drain", "Instant")), 2*time.Second)
	if _, ok := improvise(t, p, improvInput(resolvedInGraveyard(hexID, "Hex Drain", "Instant")), 2*time.Second); ok {
		t.Fatal("an unreachable endpoint produced a bundle")
	}
	if st := p.Stats(); st.ByImprov[ImprovError] != 1 {
		t.Errorf("stats = %+v", st.ByImprov)
	}
}

// An empty step list is the model declining, and it is a first-class
// answer rather than a failure: it is how "this spell was countered"
// and "I cannot do this faithfully with four verbs" are said.
func TestAnEmptyStepListIsADecline(t *testing.T) {
	fake := AlwaysText(`{"effect": "nothing — it was countered", "why": "countered", "steps": []}`)
	p := improvPolicy(t, fake, nil)
	improvise(t, p, improvInput(castingView(hexID, "Hex Drain", "Instant")), 2*time.Second)
	if _, ok := improvise(t, p, improvInput(resolvedInGraveyard(hexID, "Hex Drain", "Instant")), 2*time.Second); ok {
		t.Fatal("an empty step list produced a bundle")
	}
	if st := p.Stats(); st.ByImprov[ImprovDeclined] != 1 {
		t.Errorf("stats = %+v", st.ByImprov)
	}
}

// --- the budget -------------------------------------------------------

// The per-game cap is what makes improvisation spend bounded rather
// than merely rare, and it counts CALLS: a call that failed cost the
// same as one that worked.
func TestImprovisationIsCappedPerGame(t *testing.T) {
	fake := bundleReply()
	p := improvPolicy(t, fake, func(c *Config) { c.MaxImprovCalls = 1 })

	for _, card := range []struct{ id, name, typeLine string }{
		{grimID, "Grim Tutor", "Sorcery"},
		{hexID, "Hex Drain", "Instant"},
	} {
		improvise(t, p, improvInput(castingView(card.id, card.name, card.typeLine)), 2*time.Second)
		improvise(t, p, improvInput(resolvedInGraveyard(card.id, card.name, card.typeLine)), 2*time.Second)
	}
	if fake.Calls() != 1 {
		t.Errorf("model calls = %d, want 1 — the per-game cap did not hold", fake.Calls())
	}
	if st := p.Stats(); st.ByImprov[ImprovCapped] != 1 {
		t.Errorf("stats = %+v", st.ByImprov)
	}
}

// A deadline too short to attempt anything makes no call at all,
// exactly as Layer C's MinBudget does.
func TestNoBudgetMakesNoImprovisationCall(t *testing.T) {
	fake := bundleReply()
	p := improvPolicy(t, fake, func(c *Config) { c.MinBudget = 500 * time.Millisecond })
	improvise(t, p, improvInput(castingView(hexID, "Hex Drain", "Instant")), 2*time.Second)
	if _, ok := improvise(t, p, improvInput(resolvedInGraveyard(hexID, "Hex Drain", "Instant")), 50*time.Millisecond); ok {
		t.Fatal("improvised with no time to do it in")
	}
	if fake.Calls() != 0 {
		t.Errorf("the model was called %d times with no budget", fake.Calls())
	}
	if st := p.Stats(); st.ByImprov[ImprovNoBudget] != 1 {
		t.Errorf("stats = %+v", st.ByImprov)
	}
}

// --- turning it off ---------------------------------------------------

// Two ways a seat never improvises, and both must leave it a complete
// seat: no transport (the outage-drill deployment) and the operator's
// switch (CMDCTRL_BOT_IMPROVISE=0).
func TestImprovisationCanBeOff(t *testing.T) {
	cases := map[string]func(*Config){
		"no client":  func(c *Config) { c.Client = nil },
		"switch off": func(c *Config) { c.Improvise = false },
	}
	for name, tweak := range cases {
		t.Run(name, func(t *testing.T) {
			fake := bundleReply()
			p := improvPolicy(t, fake, tweak)
			improvise(t, p, improvInput(castingView(hexID, "Hex Drain", "Instant")), 2*time.Second)
			if _, ok := improvise(t, p, improvInput(resolvedInGraveyard(hexID, "Hex Drain", "Instant")), 2*time.Second); ok {
				t.Fatal("improvised with improvisation off")
			}
			if fake.Calls() != 0 {
				t.Errorf("the model was called %d times", fake.Calls())
			}
			// And the seat still decides.
			if _, err := p.Decide(context.Background(), castWindow()); err != nil {
				t.Errorf("a seat that does not improvise stopped being a policy: %v", err)
			}
		})
	}
}

// --- parsing ----------------------------------------------------------

func TestParseImprovisation(t *testing.T) {
	t.Run("a fence and a sentence survive", func(t *testing.T) {
		raw := "Sure, here you go:\n```json\n{\"effect\":\"drain 2\",\"why\":\"text\",\"steps\":[{\"verb\":\"change_life\",\"player\":\"" +
			oppSeat.String() + "\",\"params\":{\"delta\":-2}}]}\n```\nHope that helps."
		im, err := parseImprovisation(raw, "Hex Drain", "m")
		if err != nil {
			t.Fatalf("parse: %v", err)
		}
		if len(im.Steps) != 1 || im.Steps[0].Player != oppSeat {
			t.Fatalf("steps = %+v", im.Steps)
		}
		if err := im.Validate(); err != nil {
			t.Errorf("Validate: %v", err)
		}
	})

	t.Run("\"type\" is accepted beside \"verb\"", func(t *testing.T) {
		raw := `{"effect":"drain","steps":[{"type":"change_life","player":"` + oppSeat.String() + `","params":{"delta":-2}}]}`
		im, err := parseImprovisation(raw, "Hex Drain", "m")
		if err != nil {
			t.Fatalf("parse: %v", err)
		}
		if im.Steps[0].Type != aiseat.VerbChangeLife {
			t.Errorf("Type = %q", im.Steps[0].Type)
		}
	})

	t.Run("a player id that will not parse is refused, not dropped", func(t *testing.T) {
		raw := `{"effect":"drain","steps":[{"verb":"change_life","player":"the opponent","params":{"delta":-2}}]}`
		im, err := parseImprovisation(raw, "Hex Drain", "m")
		if err != nil {
			t.Fatalf("parse: %v", err)
		}
		if len(im.Steps) != 1 {
			t.Fatalf("the step was dropped, leaving a bundle that does half the card: %+v", im.Steps)
		}
		if err := im.Validate(); !errors.Is(err, aiseat.ErrImprovPlayer) {
			t.Errorf("Validate = %v, want ErrImprovPlayer", err)
		}
	})

	t.Run("the card name comes from the tracker, not the reply", func(t *testing.T) {
		raw := `{"card":"Black Lotus","effect":"drain","steps":[{"verb":"change_life","player":"` +
			oppSeat.String() + `","params":{"delta":-2}}]}`
		im, err := parseImprovisation(raw, "Hex Drain", "m")
		if err != nil {
			t.Fatalf("parse: %v", err)
		}
		if im.Card != "Hex Drain" {
			t.Errorf("Card = %q — a model naming a card it did not touch would make the disclosure a lie", im.Card)
		}
	})
}
