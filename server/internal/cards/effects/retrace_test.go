package effects

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// retrace_test.go — #2528, CR 702.81. Printed retrace (Flame Jab and its
// nine siblings) and Six's grant, against the REAL catalog.

// retraceCards is the roster with the number a card file is most likely
// to get wrong: the printed mana cost the retrace cast pays. Every cost
// here was read off the Scryfall dump (2026-10-05) rather than typed from
// memory.
var retraceCards = []struct {
	name   string
	oracle string
	cost   string
}{
	{"Flame Jab", "05eccdee-84f8-42d5-b79c-36d081656915", "{R}"},
	{"Raven's Crime", "a21c85f3-482b-47e5-9321-0ca21e110bd8", "{B}"},
	{"Syphon Life", "c9367dd3-1a50-4eae-993e-c055dc6c4dc5", "{1}{B}{B}"},
	{"Oona's Grace", "380e9992-b5d2-4fbe-a8c3-c37220846e0a", "{2}{U}"},
	{"Monstrify", "0f58f791-469a-4a22-996e-4906c0914858", "{3}{G}"},
	{"Savage Conception", "494050f4-0a55-415d-9ae9-feb17e61c4e1", "{3}{G}{G}"},
	{"Waves of Aggression", "10991b1a-7dd3-4fbf-a4c8-200dc62fc605", "{3}{R/W}{R/W}"},
	{"Embrace the Unknown", "9f36dd20-350c-4046-b57f-e6b5cc9aa999", "{2}{R}"},
	{"Spitting Image", "a30dee74-86e3-4888-980e-b22437fbbb66", "{4}{G/U}{G/U}"},
	{"Throes of Chaos", "e3444fcf-70ed-4d6e-aea9-030af15cad56", "{3}{R}"},
	{"Decaying Time Loop", "13e94530-defb-4c9b-9ec7-bb7789ec2630", "{3}{R}"},
}

// The declaration test: a card file that priced the cast but forgot to
// open the zone, or opened it and priced it as a replacement of the mana
// cost, is caught here. Register panics on the reverse mistake.
func TestPrintedRetraceCardsDeclareBothHalves(t *testing.T) {
	for _, c := range retraceCards {
		t.Run(c.name, func(t *testing.T) {
			if !game.CardCastableFromZone(c.oracle, game.ZoneGraveyard) {
				t.Fatalf("%s does not open the graveyard", c.name)
			}
			offers := game.AlternativeCostsOfferedFromZone(c.oracle, game.ZoneGraveyard)
			if len(offers) != 1 || offers[0].Key != "retrace" {
				t.Fatalf("%s graveyard offers = %+v, want one retrace", c.name, offers)
			}
			o := offers[0]
			if o.ManaCost != c.cost {
				t.Errorf("%s retrace pays %q, want its printed %q — retrace never replaces the mana", c.name, o.ManaCost, c.cost)
			}
			if o.DiscardFromHand == nil || o.CardPaymentCount() != 1 {
				t.Fatalf("%s retrace names no land discard — it is a graveyard cast for the printed price", c.name)
			}
			if o.ExileOnLeavingStack {
				t.Errorf("%s exiles itself on leaving the stack — that is flashback, and retrace repeats", c.name)
			}
			if got := game.AlternativeCostsOfferedFromZone(c.oracle, game.ZoneHand); len(got) != 0 {
				t.Errorf("%s offers %+v from hand", c.name, got)
			}
		})
	}
}

// retraceTable is a main phase on seat 0's turn with two Forests on the
// battlefield and a hand holding one land and one spell.
func retraceTable(t *testing.T) (g *game.Game, me *game.Player, land, spare uuid.UUID) {
	t.Helper()
	g = newCatalogGame(t)
	advanceToMain(t, g)
	me = g.Seats[0]
	me.Hand.Cards = nil
	land = uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: land, Name: "Hand Forest", TypeLine: "Basic Land — Forest",
		Owner: me.ID, Controller: me.ID, KnownBy: map[uuid.UUID]bool{me.ID: true}})
	spare = uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: spare, Name: "Spare Spell", TypeLine: "Sorcery", ManaCost: "{2}",
		Owner: me.ID, Controller: me.ID, KnownBy: map[uuid.UUID]bool{me.ID: true}})
	return
}

func addForests(g *game.Game, me *game.Player, n int) {
	for i := 0; i < n; i++ {
		pushCatalogPermanent(g, me.ID, "Forest", "Basic Land — Forest", "", false)
	}
}

func inGraveyard(p *game.Player, id uuid.UUID) bool { return p.Graveyard.Contains(id) }

func graveyardCardOf(g *game.Game, p *game.Player, name, typeLine, cost, oracle string) uuid.UUID {
	id := uuid.New()
	p.Graveyard.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, ManaCost: cost,
		OracleID: oracle, Owner: p.ID, Controller: p.ID,
		KnownBy: map[uuid.UUID]bool{p.ID: true}})
	return id
}

func payMana(t *testing.T, g *game.Game, p *game.Player, mana string) {
	t.Helper()
	if err := g.AddManaForEffect(p.ID, uuid.Nil, mana); err != nil {
		t.Fatal(err)
	}
}

func retraceCast(g *game.Game, p *game.Player, card uuid.UUID, discard ...uuid.UUID) error {
	return g.CastSpell(p.ID, card, game.CastSpellParams{
		FromZone: "graveyard", AlternativeCost: "retrace", AltCostIDs: discard, Strict: true,
	})
}

type retraceMove struct {
	AlternativeCost string   `json:"alternative_cost"`
	FromZone        string   `json:"from_zone"`
	AltCostIDs      []string `json:"alt_cost_ids"`
}

// retraceMoves is every enumerated cast of `card` out of the graveyard.
func retraceMoves(g *game.Game, seat, card uuid.UUID) []retraceMove {
	var out []retraceMove
	for _, m := range legal.EnumerateFor(g, seat) {
		if m.Type != legal.TypeCastSpell || m.Source != card {
			continue
		}
		var p retraceMove
		if err := json.Unmarshal(m.Params, &p); err != nil {
			continue
		}
		if p.FromZone == "graveyard" {
			out = append(out, p)
		}
	}
	return out
}

func graveyardViewOf(t *testing.T, g *game.Game, viewer, card uuid.UUID) protocol.CardView {
	t.Helper()
	v := protocol.ViewOfGameFor(g, viewer.String())
	for _, s := range v.Seats {
		for _, c := range s.Graveyard.Cards {
			if c.InstanceID == card.String() {
				return c
			}
		}
	}
	t.Fatalf("card %s is not in any graveyard view", card)
	return protocol.CardView{}
}

// ---------------------------------------------------------------- printed

// Flame Jab, the reference case: cast from the graveyard for {R} plus a
// land from hand, which is DISCARDED as a cost, and the spell is back in
// the graveyard after it resolves, ready to go again.
func TestFlameJabRetracesForItsPrintedCostPlusAnDiscardedLand(t *testing.T) {
	const oracle = "05eccdee-84f8-42d5-b79c-36d081656915"
	g, me, land, spare := retraceTable(t)
	opp := g.Seats[1]
	jab := graveyardCardOf(g, me, "Flame Jab", "Sorcery", "{R}", oracle)

	// A graveyard cast that claims nothing is refused: the zone is priced.
	payMana(t, g, me, "{R}")
	if err := g.CastSpell(me.ID, jab, game.CastSpellParams{FromZone: "graveyard", Strict: true}); err == nil {
		t.Fatal("a graveyard cast that claimed no retrace was accepted — a free Flame Jab")
	}
	// Not the printed cost alone: a spell in hand does not pay a land.
	if err := retraceCast(g, me, jab, spare); err == nil {
		t.Fatal("a nonland card paid retrace's discard")
	}
	if !me.Hand.Contains(spare) {
		t.Fatal("a rejected cast discarded a card")
	}
	// And no discard at all.
	if err := retraceCast(g, me, jab); err == nil {
		t.Fatal("retrace cast without discarding a land")
	}
	lifeBefore := opp.Life
	if err := g.CastSpell(me.ID, jab, game.CastSpellParams{
		FromZone: "graveyard", AlternativeCost: "retrace", AltCostIDs: []uuid.UUID{land}, Strict: true,
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}},
	}); err != nil {
		t.Fatalf("retrace cast: %v", err)
	}
	// Paid with the spell already on the stack (CR 601.2a before 601.2h):
	// the land is in the graveyard while the spell is still waiting.
	if g.Stack.Size() != 1 {
		t.Fatalf("stack = %d items, want the retraced spell", g.Stack.Size())
	}
	if me.Hand.Contains(land) || !inGraveyard(me, land) {
		t.Fatal("the land was not discarded to pay for the cast")
	}
	if len(me.ManaPool) != 0 {
		t.Errorf("{R} was not paid; pool = %v", me.ManaPool)
	}
	passPriorityAroundTable(t, g)
	if opp.Life != lifeBefore-1 {
		t.Errorf("Flame Jab dealt %d, want 1", lifeBefore-opp.Life)
	}
	if !inGraveyard(me, jab) {
		t.Fatal("a resolved retrace spell must be in the graveyard to retrace again")
	}
	// Again: a second land, a second {R}.
	land2 := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: land2, Name: "Second Forest", TypeLine: "Basic Land — Forest",
		Owner: me.ID, Controller: me.ID, KnownBy: map[uuid.UUID]bool{me.ID: true}})
	advanceToMain(t, g)
	g.Turn.Step = game.StepPrecombatMain
	payMana(t, g, me, "{R}")
	if err := g.CastSpell(me.ID, jab, game.CastSpellParams{
		FromZone: "graveyard", AlternativeCost: "retrace", AltCostIDs: []uuid.UUID{land2}, Strict: true,
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}},
	}); err != nil {
		t.Fatalf("second retrace: %v", err)
	}
}

// Claiming retrace on the hand cast is refused: the discard belongs to the
// graveyard cast, and the hand cast pays only the printed cost.
func TestRetraceIsNotClaimableFromHand(t *testing.T) {
	const oracle = "05eccdee-84f8-42d5-b79c-36d081656915"
	g, me, land, _ := retraceTable(t)
	jab := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: jab, Name: "Flame Jab", TypeLine: "Sorcery", ManaCost: "{R}",
		OracleID: oracle, Owner: me.ID, Controller: me.ID, KnownBy: map[uuid.UUID]bool{me.ID: true}})
	payMana(t, g, me, "{R}")
	if err := g.CastSpell(me.ID, jab, game.CastSpellParams{
		AlternativeCost: "retrace", AltCostIDs: []uuid.UUID{land}, Strict: true,
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: g.Seats[1].ID}},
	}); err == nil {
		t.Fatal("retrace was claimable from hand")
	}
	if !me.Hand.Contains(land) {
		t.Fatal("a refused cast discarded the land")
	}
	if err := g.CastSpell(me.ID, jab, game.CastSpellParams{
		Strict: true, Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: g.Seats[1].ID}},
	}); err != nil {
		t.Fatalf("the plain hand cast: %v", err)
	}
	if !me.Hand.Contains(land) {
		t.Error("the hand cast discarded a land")
	}
}

// A hand with no land has no retrace: the enumerator offers nothing, the
// view greys the card, and the engine refuses.
func TestRetraceWithoutALandInHandIsNotOffered(t *testing.T) {
	const oracle = "05eccdee-84f8-42d5-b79c-36d081656915"
	g, me, land, spare := retraceTable(t)
	jab := graveyardCardOf(g, me, "Flame Jab", "Sorcery", "{R}", oracle)
	pushCatalogPermanent(g, me.ID, "Mountain", "Basic Land — Mountain", "", false)
	got := retraceMoves(g, me.ID, jab)
	if len(got) == 0 {
		t.Fatal("with a land in hand the enumerator offers no retrace")
	}
	for _, m := range got {
		// Every offered payment is the one retrace there is: claim the
		// key and discard the one land in hand.
		if m.AlternativeCost != "retrace" || len(m.AltCostIDs) != 1 || m.AltCostIDs[0] != land.String() {
			t.Fatalf("the enumerator offers %+v, want retrace discarding the hand land", m)
		}
	}
	if v := graveyardViewOf(t, g, me.ID, jab); !v.CastableHere {
		t.Fatal("the view does not mark a payable retrace castable_here")
	} else if len(v.AlternativeCosts) != 1 || v.AlternativeCosts[0].Key != "retrace" ||
		v.AlternativeCosts[0].PayOptions == nil || len(v.AlternativeCosts[0].PayOptions.Cards) != 1 ||
		v.AlternativeCosts[0].PayOptions.Cards[0] != land.String() {
		t.Fatalf("the retrace offer on the wire = %+v, want the one land as its pay option", v.AlternativeCosts)
	}
	// Take the land away.
	me.Hand.Cards = nil
	me.Hand.PushTop(game.Card{InstanceID: spare, Name: "Spare Spell", TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID})
	if got := retraceMoves(g, me.ID, jab); len(got) != 0 {
		t.Errorf("no land in hand, yet the enumerator offers %+v", got)
	}
	if v := graveyardViewOf(t, g, me.ID, jab); v.CastableHere {
		t.Error("no land in hand, yet the view marks the card castable_here")
	}
	if err := retraceCast(g, me, jab, spare); err == nil {
		t.Error("the engine accepted a retrace with no land to discard")
	}
}

// Raven's Crime pins the discard prompt the card brings, through a retrace.
func TestRavensCrimeRetracesAndTheTargetDiscards(t *testing.T) {
	const oracle = "a21c85f3-482b-47e5-9321-0ca21e110bd8"
	g, me, land, _ := retraceTable(t)
	opp := g.Seats[1]
	opp.Hand.Cards = nil
	victim := uuid.New()
	opp.Hand.PushTop(game.Card{InstanceID: victim, Name: "Victim", TypeLine: "Instant", Owner: opp.ID, Controller: opp.ID})
	crime := graveyardCardOf(g, me, "Raven's Crime", "Sorcery", "{B}", oracle)
	payMana(t, g, me, "{B}")
	if err := g.CastSpell(me.ID, crime, game.CastSpellParams{
		FromZone: "graveyard", AlternativeCost: "retrace", AltCostIDs: []uuid.UUID{land}, Strict: true,
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}},
	}); err != nil {
		t.Fatalf("retrace cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !inGraveyard(me, land) {
		t.Error("the retrace land was not discarded")
	}
	if got := opp.Hand.Size(); got > 1 {
		t.Errorf("opponent hand = %d", got)
	}
}

// ---------------------------------------------------------------- Six

func sixTable(t *testing.T) (g *game.Game, me *game.Player, six, land uuid.UUID) {
	t.Helper()
	g, me, land, _ = retraceTable(t)
	addForests(g, me, 4)
	six = pushCatalogPermanent(g, me.ID, "Six", "Legendary Creature — Treefolk", sixOracle, false)
	return
}

// The reference case: a creature card in your graveyard has retrace while
// Six is on the battlefield on your turn. It is cast for its PRINTED cost
// plus a discarded land and resolves onto the battlefield.
func TestSixGrantsRetraceToANonlandPermanentCard(t *testing.T) {
	g, me, _, land := sixTable(t)
	bear := graveyardCardOf(g, me, "Gravebound Bear", "Creature — Bear", "{1}{G}", "")

	// Not castable for free, and not castable without claiming retrace.
	payMana(t, g, me, "{G}{G}")
	if err := g.CastSpell(me.ID, bear, game.CastSpellParams{FromZone: "graveyard", Strict: true}); err == nil {
		t.Fatal("a graveyard cast under Six that claimed no retrace was accepted")
	}
	if err := retraceCast(g, me, bear); err == nil {
		t.Fatal("retrace cast without discarding a land")
	}
	if err := retraceCast(g, me, bear, land); err != nil {
		t.Fatalf("retrace under Six: %v", err)
	}
	if !inGraveyard(me, land) {
		t.Fatal("the discarded land is not in the graveyard")
	}
	if len(me.ManaPool) != 0 {
		t.Errorf("the printed {1}{G} was not paid; pool = %v", me.ManaPool)
	}
	passPriorityAroundTable(t, g)
	got, ok := battlefieldCard(g, bear)
	if !ok {
		t.Fatal("the retraced creature did not resolve onto the battlefield")
	}
	if got.Controller != me.ID {
		t.Error("the retraced creature is not under the caster's control")
	}
}

// "Nonland permanent cards": an instant, a sorcery and a land are not
// covered. A land card cannot be played from the graveyard under Six
// either: retrace opens a CAST, and the land is what pays for it.
func TestSixDoesNotGrantRetraceToInstantsSorceriesOrLands(t *testing.T) {
	g, me, _, land := sixTable(t)
	g.Turn.Step = game.StepPrecombatMain
	instant := graveyardCardOf(g, me, "Dead Bolt", "Instant", "{R}", "")
	sorcery := graveyardCardOf(g, me, "Dead Jab", "Sorcery", "{R}", "")
	grave := graveyardCardOf(g, me, "Dead Forest", "Basic Land — Forest", "", "")
	for _, id := range []uuid.UUID{instant, sorcery, grave} {
		payMana(t, g, me, "{R}{R}{G}")
		if err := retraceCast(g, me, id, land); err == nil {
			t.Errorf("retrace opened %s, which is not a nonland permanent card", id)
		}
		if len(retraceMoves(g, me.ID, id)) != 0 {
			t.Errorf("the enumerator offers %s", id)
		}
	}
	for _, typ := range []string{"Artifact", "Enchantment", "Legendary Planeswalker — Jace", "Artifact Creature — Golem"} {
		id := graveyardCardOf(g, me, "Dead "+typ, typ, "{1}", "")
		f := game.PermissionFilter{NonLandPermanentOnly: true}
		c, _ := g.LookupCardForEffect(id)
		if !f.Matches(c) {
			t.Errorf("%s is a nonland permanent card and must match", typ)
		}
	}
}

// "During your turn": the grant is Six's controller's alone and only on
// their turn. A flash creature isolates the grant's timing from the
// card's own: castable at instant speed on your own turn, refused on an
// opponent's.
func TestSixRetraceIsLiveOnlyDuringYourTurn(t *testing.T) {
	g, me, _, land := sixTable(t)
	flash := game.Card{InstanceID: uuid.New(), Name: "Ambush Bear", TypeLine: "Creature — Bear", ManaCost: "{1}{G}",
		Keywords: []string{"flash"}, Owner: me.ID, Controller: me.ID, KnownBy: map[uuid.UUID]bool{me.ID: true}}
	me.Graveyard.PushTop(flash)
	// An opponent's turn, with priority in my hands.
	g.Turn.ActiveSeat = 1
	g.Turn.PriorityHolder = 0
	g.Turn.Step = game.StepUpkeep
	payMana(t, g, me, "{G}{G}")
	if err := retraceCast(g, me, flash.InstanceID, land); err == nil {
		t.Fatal("retrace was castable on an opponent's turn")
	}
	if len(retraceMoves(g, me.ID, flash.InstanceID)) != 0 {
		t.Error("the enumerator offers a retrace on an opponent's turn")
	}
	if v := graveyardViewOf(t, g, me.ID, flash.InstanceID); v.CastableHere {
		t.Error("the view marks a retrace castable_here on an opponent's turn")
	}
	// My own turn, an instant-speed window.
	g.Turn.ActiveSeat = 0
	g.Turn.PriorityHolder = 0
	g.Turn.Step = game.StepUpkeep
	if len(retraceMoves(g, me.ID, flash.InstanceID)) == 0 {
		t.Error("the enumerator does not offer the flash creature at instant speed on my own turn")
	}
	if err := retraceCast(g, me, flash.InstanceID, land); err != nil {
		t.Fatalf("retrace of a flash creature on my own turn: %v", err)
	}
}

// Nobody else's graveyard and nobody else's turn: an opponent controls no
// Six, so their own graveyard has no retrace, and mine is not theirs.
func TestSixRetraceIsItsControllersAlone(t *testing.T) {
	g, me, _, _ := sixTable(t)
	opp := g.Seats[1]
	theirs := graveyardCardOf(g, opp, "Their Bear", "Creature — Bear", "{1}{G}", "")
	mine := graveyardCardOf(g, me, "My Bear", "Creature — Bear", "{1}{G}", "")
	oppLand := uuid.New()
	opp.Hand.PushTop(game.Card{InstanceID: oppLand, Name: "Their Forest", TypeLine: "Basic Land — Forest",
		Owner: opp.ID, Controller: opp.ID})
	g.Turn.ActiveSeat = 1
	g.Turn.PriorityHolder = 1
	g.Turn.Step = game.StepPrecombatMain
	payMana(t, g, opp, "{G}{G}")
	if err := retraceCast(g, opp, theirs, oppLand); err == nil {
		t.Error("an opponent retraced a card from their own graveyard under MY Six")
	}
	if err := retraceCast(g, opp, mine, oppLand); err == nil {
		t.Error("an opponent retraced a card out of MY graveyard")
	}
}

// "For as long as Six is on the battlefield": the permission is derived,
// so it is gone the moment Six is.
func TestSixRetraceEndsWhenSixLeaves(t *testing.T) {
	g, me, six, land := sixTable(t)
	bear := graveyardCardOf(g, me, "Gravebound Bear", "Creature — Bear", "{1}{G}", "")
	if len(retraceMoves(g, me.ID, bear)) == 0 {
		t.Fatal("control: the enumerator does not offer the retrace under Six")
	}
	if _, err := g.Battlefield.Remove(six); err != nil {
		t.Fatal(err)
	}
	payMana(t, g, me, "{G}{G}")
	if err := retraceCast(g, me, bear, land); err == nil {
		t.Error("retrace survived Six leaving the battlefield")
	}
	if len(retraceMoves(g, me.ID, bear)) != 0 {
		t.Error("the enumerator still offers retrace without Six")
	}
}

// A card that prints its own graveyard cast keeps its own price under
// Six: Flame Jab is a sorcery, so Six covers nothing here, but a
// retrace-printed creature would not be charged twice. The check that
// matters is that Six adds no second discard to a card that already has
// the keyword: an artifact with printed retrace is cast with ONE land.
func TestSixAndAPrintedRetraceDoNotStack(t *testing.T) {
	g, me, _, land := sixTable(t)
	extra := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: extra, Name: "Extra Land", TypeLine: "Basic Land — Forest",
		Owner: me.ID, Controller: me.ID, KnownBy: map[uuid.UUID]bool{me.ID: true}})
	const oracle = "13e94530-defb-4c9b-9ec7-bb7789ec2630" // Decaying Time Loop: printed retrace
	loop := graveyardCardOf(g, me, "Decaying Time Loop", "Instant", "{3}{R}", oracle)
	payMana(t, g, me, "{R}{R}{R}{R}")
	if err := retraceCast(g, me, loop, land, extra); err == nil {
		t.Fatal("two lands were accepted for a single retrace")
	}
	if err := retraceCast(g, me, loop, land); err != nil {
		t.Fatalf("printed retrace under a Six: %v", err)
	}
}
