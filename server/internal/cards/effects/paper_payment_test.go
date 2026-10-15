package effects

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// paper_payment_test.go — #3855 (ADR 0068 amendment 2026-10-15, "What
// was paid on paper"): a waived payment that something reads asks the
// caster what they paid outside the pool, one test per reader family,
// plus the no-reader path that stays one click. The converge family
// (Painful Truths) and the amount family (Mockingbird, Mana Sculpt) are
// in counter_costs_and_mana_spent_test.go and mana_spent_total_test.go;
// "your casts" (Satoru) in batch30_test.go.

const ppBoromirOracle = "d72b57e5-1c6e-4f02-968f-7e2dd24f0d3c"

// ppHand puts a vanilla card in p's hand.
func ppHand(g *game.Game, p *game.Player, name, typeLine, cost, oracle string) uuid.UUID {
	id := uuid.New()
	g.WithWriteLock(func() {
		p.Hand.PushTop(game.Card{
			InstanceID: id, Name: name, TypeLine: typeLine, ManaCost: cost, OracleID: oracle,
			Power: 2, Toughness: 2, Owner: p.ID, Controller: p.ID,
		})
	})
	return id
}

func ppAsks(t *testing.T, err error) *game.PaperPaymentRequiredError {
	t.Helper()
	var ask *game.PaperPaymentRequiredError
	if !errors.As(err, &ask) {
		t.Fatalf("cast = %v, want a paper payment prompt", err)
	}
	return ask
}

// A cast nothing reads stays one click: no prompt, the record keeps the
// pool's share and the shortfall's amount, and the log line of a
// one-click Cast anyway names the shortfall.
func TestPaperPaymentNoReaderStaysOneClick(t *testing.T) {
	for _, tc := range []struct {
		name   string
		params game.CastSpellParams
		paper  string
	}{
		{"strict off", game.CastSpellParams{}, ""},
		{"cast anyway", game.CastSpellParams{Strict: true, ForceCast: true}, "{1}{G}"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			advanceToMain(t, g)
			id := ppHand(g, me, "Plain Bear", "Creature — Bear", "{1}{G}", "")
			if err := g.CastSpell(me.ID, id, tc.params); err != nil {
				t.Fatalf("a cast nothing reads asked: %v", err)
			}
			paid := g.StackMeta[id].Paid
			if !paid.OnPaper || paid.PaperMana != 2 || paid.Spent().Total() != 2 || paid.NoManaSpent() {
				t.Errorf("record = %+v, want the {1}{G} recorded as paid on paper", paid)
			}
			ev := lastEventOfKind(g, game.EventCast)
			if ev.PaperPaid != tc.paper || ev.PaperDeclared || ev.Unpaid {
				t.Errorf("EventCast = {%q %v %v}, want PaperPaid %q", ev.PaperPaid, ev.PaperDeclared, ev.Unpaid, tc.paper)
			}
		})
	}
}

// Vexing Bauble watches every cast. A waived cast under it asks; one
// declared cast without paying its mana cost is countered as printed,
// one declared paid is not, and one declared free that the pool still
// paid part of spent mana (CR 601.2h), so it is not countered either.
// The declaration is public: the cast's log fields carry it.
func TestPaperPaymentVexingBaubleCountersADeclaredFreeCast(t *testing.T) {
	for _, tc := range []struct {
		name      string
		pool      string
		decl      game.PaperPayment
		countered bool
	}{
		{"declared free", "", game.PaperPayment{Free: true}, true},
		{"declared paid", "", game.PaperPayment{Mana: []game.PaperMana{{Color: "G"}, {Color: "R"}}}, false},
		{"declared free, the pool paid part", "G", game.PaperPayment{Free: true}, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
			// The Bauble is the OPPONENT's: it reads every player's casts.
			b12Push(g, opp.ID, "Vexing Bauble", "Artifact", vexingBaubleOracle, 0, 0)
			advanceToMain(t, g)
			floatForTest(g, me, tc.pool)
			id := ppHand(g, me, "Bear", "Creature — Bear", "{1}{G}", "")
			ask := ppAsks(t, g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, ForceCast: true}))
			if len(ask.Slots) != 2-len(tc.pool) {
				t.Fatalf("prompt = %+v, want %d slot(s)", ask, 2-len(tc.pool))
			}
			decl := tc.decl
			if decl.Mana != nil && tc.pool != "" {
				t.Fatal("bad case")
			}
			if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, ForceCast: true, PaperPayment: &decl}); err != nil {
				t.Fatalf("cast: %v", err)
			}
			ev := lastEventOfKind(g, game.EventCast)
			if ev.Unpaid != decl.Free || (ev.PaperPaid != "") == decl.Free {
				t.Errorf("EventCast = {Unpaid:%v PaperPaid:%q}, want the declaration said", ev.Unpaid, ev.PaperPaid)
			}
			passPriorityAroundTable(t, g)
			z := b12ZoneOf(g, id)
			if tc.countered && z != game.ZoneGraveyard {
				t.Errorf("the Bear is in %s, want countered", z)
			}
			if !tc.countered && z != game.ZoneBattlefield {
				t.Errorf("the Bear is in %s, want it resolved", z)
			}
		})
	}
}

// A cast the engine already knows is free — a {0} cost, in any payment
// mode, Cast anyway included — waives nothing, so it asks nothing, and
// it is "no mana spent": the Bauble counters it.
func TestPaperPaymentEngineKnownFreeCastAsksNothing(t *testing.T) {
	for _, tc := range []struct {
		name   string
		params game.CastSpellParams
	}{
		{"strict", game.CastSpellParams{Strict: true}},
		{"strict off", game.CastSpellParams{}},
		{"cast anyway", game.CastSpellParams{Strict: true, ForceCast: true}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			b12Push(g, me.ID, "Vexing Bauble", "Artifact", vexingBaubleOracle, 0, 0)
			advanceToMain(t, g)
			id := ppHand(g, me, "Free Bear", "Creature — Bear", "{0}", "")
			if err := g.CastSpell(me.ID, id, tc.params); err != nil {
				t.Fatalf("a free cast asked: %v", err)
			}
			if paid := g.StackMeta[id].Paid; !paid.NoManaSpent() {
				t.Fatalf("a {0} cast is not a known nothing: %+v", paid)
			}
			passPriorityAroundTable(t, g)
			if z := b12ZoneOf(g, id); z != game.ZoneGraveyard {
				t.Errorf("the free spell is in %s, want countered", z)
			}
		})
	}
}

// Boromir reads only his controller's OPPONENTS' casts: his own
// controller's waived cast stays one click, an opponent's asks.
func TestPaperPaymentBoromirAsksOnlyOpponents(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	b12Push(g, me.ID, "Boromir, Warden of the Tower", "Legendary Creature — Human Soldier", ppBoromirOracle, 3, 3)
	advanceToMain(t, g)
	mine := ppHand(g, me, "My Bear", "Creature — Bear", "{1}{G}", "")
	if err := g.CastSpell(me.ID, mine, game.CastSpellParams{}); err != nil {
		t.Fatalf("Boromir's controller was asked: %v", err)
	}
	passPriorityAroundTable(t, g)
	theirs := ppHand(g, opp, "Their Bolt", "Instant", "{R}", "")
	ppAsks(t, g.CastSpell(opp.ID, theirs, game.CastSpellParams{Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID}}}))
}

// "No mana was spent" on the spell itself: Freestrider Commando asks,
// and enters with its two counters when declared cast without paying
// its mana cost, and with none when declared paid.
func TestPaperPaymentFreestriderCommandoReadsTheDeclaration(t *testing.T) {
	for _, tc := range []struct {
		name string
		decl game.PaperPayment
		want int
	}{
		{"declared free", game.PaperPayment{Free: true}, 2},
		{"declared paid", game.PaperPayment{Mana: []game.PaperMana{{Color: "G"}, {Color: "G"}, {Color: "W"}}}, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			advanceToMain(t, g)
			id := ppHand(g, me, "Freestrider Commando", "Creature — Centaur Mercenary", "{2}{G}", gmrFreestriderOracle)
			ask := ppAsks(t, g.CastSpell(me.ID, id, game.CastSpellParams{}))
			if len(ask.Slots) != 3 || ask.Slots[0].Symbol != "{G}" {
				t.Fatalf("prompt = %+v, want {G} {1} {1}", ask)
			}
			decl := tc.decl
			if err := g.CastSpell(me.ID, id, game.CastSpellParams{PaperPayment: &decl}); err != nil {
				t.Fatalf("cast: %v", err)
			}
			passPriorityAroundTable(t, g)
			if got := counterCount(g, id, game.CounterPlusOne); got != tc.want {
				t.Errorf("+1/+1 counters = %d, want %d", got, tc.want)
			}
		})
	}
}

// Coin of Mastery reads "mana from an artifact source", which a paper
// source can be: its prompt asks, per mana, whether an artifact made it,
// and the creature enters with a counter for each one declared so.
func TestPaperPaymentCoinOfMasteryAsksWhichManaCameFromArtifacts(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	seedPermanentWithOracle(g, me.ID, "Coin of Mastery", "Artifact", gmrCoinOfMasteryOracle)
	id := ppHand(g, me, "Bear", "Creature — Bear", "{2}{G}", "")
	ask := ppAsks(t, g.CastSpell(me.ID, id, game.CastSpellParams{}))
	if !ask.ArtifactSources {
		t.Fatal("Coin of Mastery's prompt does not ask about artifact sources")
	}
	decl := &game.PaperPayment{Mana: []game.PaperMana{{Color: "G"}, {Color: "C", Artifact: true}, {Color: "C", Artifact: true}}}
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{PaperPayment: decl}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := counterCount(g, id, game.CounterPlusOne); got != 2 {
		t.Errorf("+1/+1 counters = %d, want 2 — the two declared from artifacts", got)
	}
}

// An activated ability that reads its own mana asks too: a permissive
// Protective Sphere activation with nothing in the pool, declared paid
// with {R}, offers only the red source.
func TestPaperPaymentProtectiveSphereActivationAsks(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	sphere := pushCatalogPermanent(g, me.ID, "Protective Sphere", "Enchantment", pr6ProtectiveSphere, false)
	red := pr7Creature(g, opp.ID, "Red", 3, "R")
	green := pr7Creature(g, opp.ID, "Green", 3, "G")
	life := me.Life
	err := g.ActivateCatalogAbility(me.ID, sphere, 0, game.ActivateAbilityParams{})
	if ask := ppAsks(t, err); len(ask.Slots) != 1 || ask.Slots[0].Symbol != "{1}" {
		t.Fatalf("prompt = %+v, want the {1}", ask)
	}
	if me.Life != life || len(g.Stack.Cards)+len(g.StackMeta) != 0 {
		t.Fatal("the refused activation paid something or used the stack")
	}
	if err := g.ActivateCatalogAbility(me.ID, sphere, 0, game.ActivateAbilityParams{
		PaperPayment: &game.PaperPayment{Mana: []game.PaperMana{{Color: "R"}}},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	ev := lastEventOfKind(g, game.EventActivateAbility)
	if ev.PaperPaid != "{R}" || !ev.PaperDeclared {
		t.Errorf("EventActivateAbility = {%q %v}, want the {R} declared", ev.PaperPaid, ev.PaperDeclared)
	}
	passPriorityAroundTable(t, g)
	c := pr7SourcePrompt(t, g)
	if pr7Offers(c, green) || !pr7Offers(c, red) {
		t.Fatalf("offered %v: want the red source only", c.ChooseCards)
	}
}

// lastEventOfKind is the most recent event of one kind.
func lastEventOfKind(g *game.Game, kind game.EventKind) game.Event {
	for i := len(g.Events) - 1; i >= 0; i-- {
		if g.Events[i].Kind == kind {
			return g.Events[i]
		}
	}
	return game.Event{}
}
