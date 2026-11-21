package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_gap_b12_cards_test.go — Innistrad Remastered creatures, S82 GAP
// batch 12 (#5948), played through the real catalog entries.

const (
	gapB12MaulerOracle     = "03025ec0-75fe-439d-b0c1-9bdb0f0ff336"
	gapB12MistRavenOracle  = "12654457-3aae-4046-a976-8e41c3f78927"
	gapB12WandererOracle   = "1803850e-9499-4bf7-a4d9-6bf655cbe87d"
	gapB12GeistOracle      = "2450d35b-7d87-4885-b396-f8e1d3ac0b22"
	gapB12ManglerOracle    = "4c502f1b-7361-48c0-8def-0d3b3e9ff628"
	gapB12ResearcherOracle = "722ee05f-6815-41b7-9813-5ec3902c1e9f"
	gapB12DocentOracle     = "85e797de-da27-4cd5-8fa4-4aef948988b2"
	gapB12GossipOracle     = "8a8e225b-14e4-464f-bebe-89b4c9be7fe1"
	gapB12AnglerOracle     = "d37effd3-d194-4609-9994-6d70038a68f9"
	gapB12NavigatorOracle  = "ebef6d84-b724-4a84-bcde-8eb6dc06717d"
	gapB12ThingOracle      = "f2c10591-4c0c-4286-af91-e8d8d10a4c9d"
)

var (
	gapB12Researcher = iwCard{gapB12ResearcherOracle, "Aberrant Researcher", "Creature — Human Insect", "{3}{U}", "Flying\nAt the beginning of your upkeep, mill a card. If an instant or sorcery card was milled this way, transform this creature.", "3", "2",
		"Perfected Form", "Flying", "5", "4", []string{"U"}, []string{"Flying"}}
	gapB12Docent = iwCard{gapB12DocentOracle, "Docent of Perfection", "Creature — Insect Horror", "{3}{U}{U}", "Flying", "5", "4",
		"Final Iteration", "Flying", "6", "5", []string{"U"}, []string{"Flying"}}
	gapB12Gossip = iwCard{gapB12GossipOracle, "Town Gossipmonger", "Creature — Human", "{W}", "{T}, Tap an untapped creature you control: Transform this creature.", "1", "1",
		"Incited Rabble", "This creature attacks each combat if able.\n{2}: This creature gets +1/+0 until end of turn.", "2", "3", []string{"W"}, nil}
	gapB12Angler = iwCard{gapB12AnglerOracle, "Grizzled Angler", "Creature — Human", "{2}{U}", "{T}: Mill two cards.", "2", "3",
		"Grisly Anglerfish", "{6}: Creatures your opponents control attack this turn if able.", "4", "5", []string{"U"}, nil}
	gapB12Thing = iwCard{gapB12ThingOracle, "Thing in the Ice", "Creature — Horror", "{1}{U}", "Defender", "0", "4",
		"Awoken Horror", "When this creature transforms into Awoken Horror, return all non-Horror creatures to their owners' hands.", "7", "8", []string{"U"}, []string{"Defender"}}
)

func TestGapB12EveryCardIsRegisteredFull(t *testing.T) {
	for id, name := range map[string]string{
		gapB12MaulerOracle: "Makeshift Mauler", gapB12MistRavenOracle: "Mist Raven",
		gapB12WandererOracle: "Mausoleum Wanderer", gapB12GeistOracle: "Battleground Geist",
		gapB12ManglerOracle: "Stitched Mangler", gapB12ResearcherOracle: "Aberrant Researcher",
		gapB12ResearcherOracle + "#1": "Perfected Form", gapB12DocentOracle: "Docent of Perfection",
		gapB12DocentOracle + "#1": "Final Iteration", gapB12GossipOracle: "Town Gossipmonger",
		gapB12GossipOracle + "#1": "Incited Rabble", gapB12AnglerOracle: "Grizzled Angler",
		gapB12AnglerOracle + "#1": "Grisly Anglerfish", gapB12NavigatorOracle: "Deadeye Navigator",
		gapB12ThingOracle: "Thing in the Ice", gapB12ThingOracle + "#1": "Awoken Horror",
	} {
		spec, ok := Lookup(id)
		if !ok || spec.Name != name || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s registration: %+v", name, spec)
		}
	}
}

// gapB12PutBackType is iwPut with a back face whose type line is the card's
// own (iwCard.row() gives every back face "Creature — Werewolf").
func gapB12PutBackType(t *testing.T, g *game.Game, w iwCard, p *game.Player, backType string) uuid.UUID {
	t.Helper()
	row := w.row()
	row.TypeLine = w.frontType + " // " + backType
	row.CardFaces[1].TypeLine = backType
	id := importToHand(row, p)
	c, ok := cardInZone(p.Hand, id)
	if !ok {
		t.Fatalf("%s did not reach the hand", w.front)
	}
	p.Hand.Remove(id)
	pushBattlefieldCardWithTimestamp(g, c)
	return id
}

func gapB12Target(id uuid.UUID) []game.TargetRef {
	return []game.TargetRef{{Kind: game.TargetCard, ID: id}}
}

// --- Makeshift Mauler ---

func TestGapB12MaulerNeedsACreatureCardToExile(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	b06AddMana(me, "U", "C", "C", "C")
	if _, err := castFromHandErr(g, me, "Makeshift Mauler", "Creature — Zombie Horror", "{3}{U}", gapB12MaulerOracle,
		game.CastSpellParams{Strict: true}); err == nil {
		t.Fatal("cast with no creature card in the graveyard")
	}
	spell := graveyardCardOf(g, me, "A Spell", "Sorcery", "{1}", "")
	creature := graveyardCardOf(g, me, "A Bear", "Creature — Bear", "{1}", "")
	b06AddMana(me, "U", "C", "C", "C")
	if _, err := castFromHandErr(g, me, "Makeshift Mauler", "Creature — Zombie Horror", "{3}{U}", gapB12MaulerOracle,
		game.CastSpellParams{Strict: true, ExileIDs: []uuid.UUID{spell}}); err == nil {
		t.Fatal("a sorcery card paid the cost")
	}
	b06AddMana(me, "U", "C", "C", "C")
	if _, err := castFromHandErr(g, me, "Makeshift Mauler", "Creature — Zombie Horror", "{3}{U}", gapB12MaulerOracle,
		game.CastSpellParams{Strict: true, ExileIDs: []uuid.UUID{creature}}); err != nil {
		t.Fatalf("cast exiling a creature card: %v", err)
	}
	if !g.Exile.Contains(creature) {
		t.Error("the creature card was not exiled")
	}
}

// --- Mist Raven ---

func TestGapB12MistRavenBouncesATargetCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	bear := b12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	castCatalogSpell(t, g, "Mist Raven", "Creature — Bird", gapB12MistRavenOracle, nil)
	passPriorityAroundTable(t, g)
	pickTriggerTarget(t, g, me.ID, bear)
	passPriorityAroundTable(t, g)
	if !opp.Hand.Contains(bear) {
		t.Error("the Bear was not returned to its owner's hand")
	}
	assertKeywords(t, g, battlefieldIDNamed(g, me.ID, "Mist Raven"), "flying")
}

// --- Battleground Geist ---

func TestGapB12GeistPumpsOnlyYourOtherSpirits(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	geist := b12Push(g, me.ID, "Battleground Geist", "Creature — Spirit", gapB12GeistOracle, 3, 3)
	mine := b12Creature(g, me.ID, "My Spirit", "Creature — Spirit", 1, 1)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Spirit", "Creature — Spirit", 1, 1)
	assertKeywords(t, g, geist, "flying")
	for _, tc := range []struct {
		name string
		id   uuid.UUID
		want int
	}{{"the Geist", geist, 3}, {"your Spirit", mine, 2}, {"your Bear", bear, 2}, {"their Spirit", theirs, 1}} {
		if got := layeredCard(t, g, tc.id).CurrentPower(); got != tc.want {
			t.Errorf("%s has power %d, want %d", tc.name, got, tc.want)
		}
	}
}

// --- Stitched Mangler ---

func TestGapB12ManglerEntersTappedAndFreezesACreature(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me, opp := g.Seats[seat], g.Seats[(seat+1)%len(g.Seats)]
	bear := b12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	castCatalogSpell(t, g, "Stitched Mangler", "Creature — Zombie Horror", gapB12ManglerOracle, nil)
	passPriorityAroundTable(t, g)
	pickTriggerTarget(t, g, me.ID, bear)
	passPriorityAroundTable(t, g)
	if _, c := cardWhere(g, me, battlefieldIDNamed(g, me.ID, "Stitched Mangler")); !c.Tapped {
		t.Error("the Mangler did not enter tapped")
	}
	if _, c := cardWhere(g, opp, bear); !c.Tapped {
		t.Fatal("the Bear was not tapped")
	}
	advanceToUpkeepOf(t, g, (seat+1)%len(g.Seats))
	if _, c := cardWhere(g, opp, bear); !c.Tapped {
		t.Error("the Bear untapped during its controller's next untap step")
	}
}

// --- Mausoleum Wanderer ---

func TestGapB12WandererGrowsWhenAnotherSpiritEnters(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	wanderer := b12Push(g, me.ID, "Mausoleum Wanderer", "Creature — Spirit", gapB12WandererOracle, 1, 1)
	assertKeywords(t, g, wanderer, "flying")
	castCatalogSpell(t, g, "Another Spirit", "Creature — Spirit", "", nil)
	passPriorityAroundTable(t, g)
	if got := layeredCard(t, g, wanderer).CurrentPower(); got != 2 {
		t.Errorf("power %d after another Spirit entered, want 2", got)
	}
	castCatalogSpell(t, g, "A Bear", "Creature — Bear", "", nil)
	passPriorityAroundTable(t, g)
	if got := layeredCard(t, g, wanderer).CurrentPower(); got != 2 {
		t.Errorf("power %d after a non-Spirit entered, want 2", got)
	}
}

func TestGapB12WandererCountersAnInstantUnlessItsControllerPaysItsPower(t *testing.T) {
	for _, pays := range []bool{false, true} {
		g := newCatalogGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
		wanderer := b12Push(g, me.ID, "Mausoleum Wanderer", "Creature — Spirit", gapB12WandererOracle, 1, 1)
		spell := batch01OpponentCasts(t, g, opp, "Big Bolt", lightningBoltOracle, "{2}{R}",
			[]game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID}})
		if pays {
			b06AddMana(opp, "R")
		}
		if err := g.ActivateCatalogAbility(me.ID, wanderer, 0, game.ActivateAbilityParams{
			Targets: gapB12Target(spell), Strict: true}); err != nil {
			t.Fatalf("activate: %v", err)
		}
		if g.Battlefield.Contains(wanderer) {
			t.Fatal("the Wanderer was not sacrificed as a cost")
		}
		passUntilTaxed(t, g, opp.ID)
		answerPayUnless(t, g, opp.ID, pays)
		passPriorityAroundTable(t, g)
		if pays && opp.Graveyard.Contains(spell) && me.Life == 40 {
			t.Error("the tax was paid but the spell was countered anyway")
		}
		if !pays && !opp.Graveyard.Contains(spell) {
			t.Error("the spell was not countered although its controller declined the {1}")
		}
		if !pays && me.Life != 40 {
			t.Errorf("a countered Bolt still hit for %d", 40-me.Life)
		}
	}
}

// --- Aberrant Researcher ---

func TestGapB12ResearcherTransformsOnlyWhenAnInstantOrSorceryIsMilled(t *testing.T) {
	for _, hit := range []bool{true, false} {
		g, me := iwTable(t)
		id := iwPut(t, g, gapB12Researcher, me)
		top := game.Card{InstanceID: uuid.New(), Name: "Bear", TypeLine: "Creature — Bear", Owner: me.ID, Controller: me.ID}
		if hit {
			top = game.Card{InstanceID: uuid.New(), Name: "Spell", TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID}
		}
		me.Library.PushTop(top)
		// The next upkeep of its controller: walk around the table.
		for i := 0; i < 400; i++ {
			if _, err := g.AdvanceStep(); err != nil {
				t.Fatalf("AdvanceStep: %v", err)
			}
			if g.Turn.Step == game.StepUpkeep && g.Seats[g.Turn.ActiveSeat].ID == me.ID {
				break
			}
		}
		monSettle(t, g)
		passPriorityAroundTable(t, g)
		want := 0
		if hit {
			want = 1
		}
		if iwFace(g, id) != want {
			t.Errorf("hit=%v: face %d, want %d", hit, iwFace(g, id), want)
		}
		if !me.Graveyard.Contains(top.InstanceID) {
			t.Errorf("hit=%v: the top card was not milled", hit)
		}
	}
}

// --- Docent of Perfection // Final Iteration ---

func gapB12CastInstant(t *testing.T, g *game.Game) {
	t.Helper()
	castCatalogSpell(t, g, "Test Instant", "Instant", "", nil)
	passPriorityAroundTable(t, g)
}

func TestGapB12DocentMakesWizardsAndTransformsAtThree(t *testing.T) {
	g, me := iwTable(t)
	docent := iwPut(t, g, gapB12Docent, me)
	for i := 1; i <= 3; i++ {
		gapB12CastInstant(t, g)
		if got := countBattlefieldNamed(g, me.ID, "Human Wizard"); got != i {
			t.Fatalf("after %d instants: %d Wizard tokens, want %d", i, got, i)
		}
		wantFace := 0
		if i == 3 {
			wantFace = 1
		}
		if iwFace(g, docent) != wantFace {
			t.Fatalf("after %d instants: face %d, want %d", i, iwFace(g, docent), wantFace)
		}
	}
	wizard := battlefieldIDNamed(g, me.ID, "Human Wizard")
	if got := layeredCard(t, g, wizard).CurrentPower(); got != 3 {
		t.Errorf("a Wizard token has power %d under Final Iteration, want 3", got)
	}
	if !hasEffectiveKeyword(t, g, wizard, "flying") {
		t.Error("a Wizard token does not fly under Final Iteration")
	}
	gapB12CastInstant(t, g)
	if got := countBattlefieldNamed(g, me.ID, "Human Wizard"); got != 4 {
		t.Errorf("Final Iteration made %d Wizards in all, want 4", got)
	}
	if iwFace(g, docent) != 1 {
		t.Error("Final Iteration turned back over")
	}
}

func TestGapB12DocentIgnoresACreatureSpell(t *testing.T) {
	g, me := iwTable(t)
	iwPut(t, g, gapB12Docent, me)
	castCatalogSpell(t, g, "A Bear", "Creature — Bear", "", nil)
	passPriorityAroundTable(t, g)
	if got := countBattlefieldNamed(g, me.ID, "Human Wizard"); got != 0 {
		t.Errorf("a creature spell made %d Wizards", got)
	}
}

// --- Town Gossipmonger // Incited Rabble ---

func TestGapB12GossipmongerTapsAnotherCreatureToTransform(t *testing.T) {
	g, me := iwTable(t)
	gossip := iwPut(t, g, gapB12Gossip, me)
	other := b12Creature(g, me.ID, "Other", "Creature — Bear", 1, 1)
	if err := g.ActivateCatalogAbility(me.ID, gossip, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("activated without naming an untapped creature to tap")
	}
	if err := g.ActivateCatalogAbility(me.ID, gossip, 0, game.ActivateAbilityParams{TapIDs: []uuid.UUID{other}, Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if iwFace(g, gossip) != 1 {
		t.Fatalf("Town Gossipmonger is on face %d, want Incited Rabble", iwFace(g, gossip))
	}
	if _, c := cardWhere(g, me, other); !c.Tapped {
		t.Error("the other creature was not tapped")
	}
}

func TestGapB12IncitedRabbleMustAttackAndPumps(t *testing.T) {
	g, me := iwTable(t)
	rabble := iwPutBack(t, g, gapB12Gossip, me)
	if err := hobB4Activate(t, g, me.ID, rabble, 0, broB04Pool("{2}"), nil); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := layeredCard(t, g, rabble).CurrentPower(); got != 3 {
		t.Errorf("power %d after the pump, want 3", got)
	}
	advanceTo(t, g, game.StepDeclareAttackers)
	g.WithWriteLock(func() {
		if g.AttackRequirementsUnmetForEffect() == nil {
			t.Error("Incited Rabble owes an attack but nothing says so")
		}
	})
}

// --- Grizzled Angler // Grisly Anglerfish ---

func TestGapB12AnglerMillsTwoAndTransformsForAColorlessCreatureCard(t *testing.T) {
	for _, hit := range []bool{true, false} {
		g, me := iwTable(t)
		angler := iwPut(t, g, gapB12Angler, me)
		for i := 0; i < 3; i++ {
			me.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: "Filler", TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID})
		}
		if hit {
			me.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: "Eldrazi Drone", TypeLine: "Creature — Eldrazi Drone", ManaCost: "{2}",
				Owner: me.ID, Controller: me.ID})
		} else {
			me.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: "Green Bear", TypeLine: "Creature — Bear", ManaCost: "{1}{G}", Colors: []string{"G"},
				Owner: me.ID, Controller: me.ID})
		}
		before := me.Graveyard.Size()
		if err := g.ActivateCatalogAbility(me.ID, angler, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
			t.Fatalf("activate: %v", err)
		}
		passPriorityAroundTable(t, g)
		if got := me.Graveyard.Size() - before; got != 2 {
			t.Errorf("hit=%v: milled %d, want 2", hit, got)
		}
		want := 0
		if hit {
			want = 1
		}
		if iwFace(g, angler) != want {
			t.Errorf("hit=%v: face %d, want %d", hit, iwFace(g, angler), want)
		}
	}
}

func TestGapB12AnglerfishForcesOpposingCreaturesToAttack(t *testing.T) {
	g, me := iwTable(t)
	oppSeat := (g.Turn.ActiveSeat + 1) % len(g.Seats)
	opp := g.Seats[oppSeat]
	fish := iwPutBack(t, g, gapB12Angler, me)
	b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	advanceToMainOf(t, g, oppSeat)
	if err := hobB4Activate(t, g, me.ID, fish, 0, broB04Pool("{6}"), nil); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	advanceTo(t, g, game.StepDeclareAttackers)
	g.WithWriteLock(func() {
		if g.AttackRequirementsUnmetForEffect() == nil {
			t.Error("the opponent's creature is not required to attack")
		}
	})
}

// --- Deadeye Navigator ---

func TestGapB12NavigatorGrantsABlinkToBothPairedCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceTo(t, g, game.StepPrecombatMain)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	id := sbondHand(me, "Deadeye Navigator", "Creature — Spirit", gapB12NavigatorOracle, 5, 5, game.KeywordSoulbond)
	sbondCast(t, g, me, id)
	nghcPassOnce(t, g)
	answerOwnPermanents(t, g, me.ID, bear)
	passPriorityAroundTable(t, g)
	if sbondPartner(g, id) != bear {
		t.Fatal("the Navigator is not paired with the Bear")
	}
	// The Bear now has the granted ability: blink it.
	if err := hobB4Activate(t, g, me.ID, bear, 0, broB04Pool("{1}{U}"), nil); err != nil {
		t.Fatalf("activate the granted blink on the Bear: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(bear) {
		// The returning creature is a new object with a new ID on some paths.
		if battlefieldIDNamed(g, me.ID, "Bear") == uuid.Nil {
			t.Error("the Bear did not return to the battlefield")
		}
	}
	if sbondPartner(g, id) != uuid.Nil {
		t.Error("the pair did not end when the Bear was blinked")
	}
}

// --- Thing in the Ice // Awoken Horror ---

func TestGapB12ThingInTheIceEntersWithFourIceCounters(t *testing.T) {
	g := newCatalogGame(t)
	thing := castHoldCreature(t, g, "Thing in the Ice", "Creature — Horror", gapB12ThingOracle, 0, 4)
	passPriorityAroundTable(t, g)
	if got := layeredCard(t, g, thing).Counters["ice"]; got != 4 {
		t.Errorf("%d ice counters, want 4", got)
	}
	assertKeywords(t, g, thing, "defender")
}

func TestGapB12ThingInTheIceTransformsOnTheLastCounterAndBouncesNonHorrors(t *testing.T) {
	g, me := iwTable(t)
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	thing := gapB12PutBackType(t, g, gapB12Thing, me, "Creature — Kraken Horror")
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == thing {
				g.Battlefield.Cards[i].Counters = map[string]int{"ice": 2}
			}
		}
	})
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	horror := b12Creature(g, opp.ID, "Other Horror", "Creature — Horror", 1, 1)
	castCatalogSpell(t, g, "Creature Spell", "Creature — Bear", "", nil)
	passPriorityAroundTable(t, g)
	gapB12CastInstant(t, g)
	if got := layeredCard(t, g, thing).Counters["ice"]; got != 1 {
		t.Fatalf("%d ice counters after one instant (a creature spell does not count), want 1", got)
	}
	if iwFace(g, thing) != 0 {
		t.Fatal("transformed with an ice counter left")
	}
	gapB12CastInstant(t, g)
	if iwFace(g, thing) != 1 {
		t.Fatalf("Thing in the Ice is on face %d with no ice counters, want Awoken Horror", iwFace(g, thing))
	}
	if !me.Hand.Contains(bear) || !opp.Hand.Contains(theirs) {
		t.Error("non-Horror creatures were not returned to their owners' hands")
	}
	if !g.Battlefield.Contains(horror) || !g.Battlefield.Contains(thing) {
		t.Error("a Horror was bounced")
	}
}
