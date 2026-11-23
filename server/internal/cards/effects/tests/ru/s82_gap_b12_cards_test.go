package testsru

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
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
	gapB12Researcher = effectstest.IwCard{Oracle: gapB12ResearcherOracle, Front: "Aberrant Researcher", FrontType: "Creature — Human Insect", Cost: "{3}{U}", FrontText: "Flying\nAt the beginning of your upkeep, mill a card. If an instant or sorcery card was milled this way, transform this creature.", Fp: "3", Ft: "2",
		Back: "Perfected Form", BackText: "Flying", Bp: "5", Bt: "4", Colors: []string{"U"}, Keywords: []string{"Flying"}}
	gapB12Docent = effectstest.IwCard{Oracle: gapB12DocentOracle, Front: "Docent of Perfection", FrontType: "Creature — Insect Horror", Cost: "{3}{U}{U}", FrontText: "Flying", Fp: "5", Ft: "4",
		Back: "Final Iteration", BackText: "Flying", Bp: "6", Bt: "5", Colors: []string{"U"}, Keywords: []string{"Flying"}}
	gapB12Gossip = effectstest.IwCard{Oracle: gapB12GossipOracle, Front: "Town Gossipmonger", FrontType: "Creature — Human", Cost: "{W}", FrontText: "{T}, Tap an untapped creature you control: Transform this creature.", Fp: "1", Ft: "1",
		Back: "Incited Rabble", BackText: "This creature attacks each combat if able.\n{2}: This creature gets +1/+0 until end of turn.", Bp: "2", Bt: "3", Colors: []string{"W"}, Keywords: nil}
	gapB12Angler = effectstest.IwCard{Oracle: gapB12AnglerOracle, Front: "Grizzled Angler", FrontType: "Creature — Human", Cost: "{2}{U}", FrontText: "{T}: Mill two cards.", Fp: "2", Ft: "3",
		Back: "Grisly Anglerfish", BackText: "{6}: Creatures your opponents control attack this turn if able.", Bp: "4", Bt: "5", Colors: []string{"U"}, Keywords: nil}
	gapB12Thing = effectstest.IwCard{Oracle: gapB12ThingOracle, Front: "Thing in the Ice", FrontType: "Creature — Horror", Cost: "{1}{U}", FrontText: "Defender", Fp: "0", Ft: "4",
		Back: "Awoken Horror", BackText: "When this creature transforms into Awoken Horror, return all non-Horror creatures to their owners' hands.", Bp: "7", Bt: "8", Colors: []string{"U"}, Keywords: []string{"Defender"}}
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
		spec, ok := effects.Lookup(id)
		if !ok || spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s registration: %+v", name, spec)
		}
	}
}

// gapB12PutBackType is iwPut with a back face whose type line is the card's
// own (iwCard.row() gives every back face "Creature — Werewolf").
func gapB12PutBackType(t *testing.T, g *game.Game, w effectstest.IwCard, p *game.Player, backType string) uuid.UUID {
	t.Helper()
	row := w.Row()
	row.TypeLine = w.FrontType + " // " + backType
	row.CardFaces[1].TypeLine = backType
	id := effectstest.ImportToHand(row, p)
	c, ok := effectstest.CardInZone(p.Hand, id)
	if !ok {
		t.Fatalf("%s did not reach the hand", w.Front)
	}
	p.Hand.Remove(id)
	effectstest.PushBattlefieldCardWithTimestamp(g, c)
	return id
}

func gapB12Target(id uuid.UUID) []game.TargetRef {
	return []game.TargetRef{{Kind: game.TargetCard, ID: id}}
}

// --- Makeshift Mauler ---

func TestGapB12MaulerNeedsACreatureCardToExile(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	effectstest.B06AddMana(me, "U", "C", "C", "C")
	if _, err := effectstest.CastFromHandErr(g, me, "Makeshift Mauler", "Creature — Zombie Horror", "{3}{U}", gapB12MaulerOracle,
		game.CastSpellParams{Strict: true}); err == nil {
		t.Fatal("cast with no creature card in the graveyard")
	}
	spell := effectstest.GraveyardCardOf(g, me, "A Spell", "Sorcery", "{1}", "")
	creature := effectstest.GraveyardCardOf(g, me, "A Bear", "Creature — Bear", "{1}", "")
	effectstest.B06AddMana(me, "U", "C", "C", "C")
	if _, err := effectstest.CastFromHandErr(g, me, "Makeshift Mauler", "Creature — Zombie Horror", "{3}{U}", gapB12MaulerOracle,
		game.CastSpellParams{Strict: true, ExileIDs: []uuid.UUID{spell}}); err == nil {
		t.Fatal("a sorcery card paid the cost")
	}
	effectstest.B06AddMana(me, "U", "C", "C", "C")
	if _, err := effectstest.CastFromHandErr(g, me, "Makeshift Mauler", "Creature — Zombie Horror", "{3}{U}", gapB12MaulerOracle,
		game.CastSpellParams{Strict: true, ExileIDs: []uuid.UUID{creature}}); err != nil {
		t.Fatalf("cast exiling a creature card: %v", err)
	}
	if !g.Exile.Contains(creature) {
		t.Error("the creature card was not exiled")
	}
}

// --- Mist Raven ---

func TestGapB12MistRavenBouncesATargetCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Mist Raven", "Creature — Bird", gapB12MistRavenOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PickTriggerTarget(t, g, me.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if !opp.Hand.Contains(bear) {
		t.Error("the Bear was not returned to its owner's hand")
	}
	effectstest.AssertKeywords(t, g, effectstest.BattlefieldIDNamed(g, me.ID, "Mist Raven"), "flying")
}

// --- Battleground Geist ---

func TestGapB12GeistPumpsOnlyYourOtherSpirits(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	geist := effectstest.B12Push(g, me.ID, "Battleground Geist", "Creature — Spirit", gapB12GeistOracle, 3, 3)
	mine := effectstest.B12Creature(g, me.ID, "My Spirit", "Creature — Spirit", 1, 1)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Spirit", "Creature — Spirit", 1, 1)
	effectstest.AssertKeywords(t, g, geist, "flying")
	for _, tc := range []struct {
		name string
		id   uuid.UUID
		want int
	}{{"the Geist", geist, 3}, {"your Spirit", mine, 2}, {"your Bear", bear, 2}, {"their Spirit", theirs, 1}} {
		if got := effectstest.LayeredCard(t, g, tc.id).CurrentPower(); got != tc.want {
			t.Errorf("%s has power %d, want %d", tc.name, got, tc.want)
		}
	}
}

// --- Stitched Mangler ---

func TestGapB12ManglerEntersTappedAndFreezesACreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me, opp := g.Seats[seat], g.Seats[(seat+1)%len(g.Seats)]
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Stitched Mangler", "Creature — Zombie Horror", gapB12ManglerOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PickTriggerTarget(t, g, me.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if _, c := effectstest.CardWhere(g, me, effectstest.BattlefieldIDNamed(g, me.ID, "Stitched Mangler")); !c.Tapped {
		t.Error("the Mangler did not enter tapped")
	}
	if _, c := effectstest.CardWhere(g, opp, bear); !c.Tapped {
		t.Fatal("the Bear was not tapped")
	}
	effectstest.AdvanceToUpkeepOf(t, g, (seat+1)%len(g.Seats))
	if _, c := effectstest.CardWhere(g, opp, bear); !c.Tapped {
		t.Error("the Bear untapped during its controller's next untap step")
	}
}

// --- Mausoleum Wanderer ---

func TestGapB12WandererGrowsWhenAnotherSpiritEnters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	wanderer := effectstest.B12Push(g, me.ID, "Mausoleum Wanderer", "Creature — Spirit", gapB12WandererOracle, 1, 1)
	effectstest.AssertKeywords(t, g, wanderer, "flying")
	effectstest.CastCatalogSpell(t, g, "Another Spirit", "Creature — Spirit", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.LayeredCard(t, g, wanderer).CurrentPower(); got != 2 {
		t.Errorf("power %d after another Spirit entered, want 2", got)
	}
	effectstest.CastCatalogSpell(t, g, "A Bear", "Creature — Bear", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.LayeredCard(t, g, wanderer).CurrentPower(); got != 2 {
		t.Errorf("power %d after a non-Spirit entered, want 2", got)
	}
}

func TestGapB12WandererCountersAnInstantUnlessItsControllerPaysItsPower(t *testing.T) {
	for _, pays := range []bool{false, true} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
		wanderer := effectstest.B12Push(g, me.ID, "Mausoleum Wanderer", "Creature — Spirit", gapB12WandererOracle, 1, 1)
		spell := effectstest.Batch01OpponentCasts(t, g, opp, "Big Bolt", effectstest.LightningBoltOracle, "{2}{R}",
			[]game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID}})
		if pays {
			effectstest.B06AddMana(opp, "R")
		}
		if err := g.ActivateCatalogAbility(me.ID, wanderer, 0, game.ActivateAbilityParams{
			Targets: gapB12Target(spell), Strict: true}); err != nil {
			t.Fatalf("activate: %v", err)
		}
		if g.Battlefield.Contains(wanderer) {
			t.Fatal("the Wanderer was not sacrificed as a cost")
		}
		effectstest.PassUntilTaxed(t, g, opp.ID)
		effectstest.AnswerPayUnless(t, g, opp.ID, pays)
		effectstest.PassPriorityAroundTable(t, g)
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
		g, me := effectstest.IwTable(t)
		id := effectstest.IwPut(t, g, gapB12Researcher, me)
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
		effectstest.MonSettle(t, g)
		effectstest.PassPriorityAroundTable(t, g)
		want := 0
		if hit {
			want = 1
		}
		if effectstest.IwFace(g, id) != want {
			t.Errorf("hit=%v: face %d, want %d", hit, effectstest.IwFace(g, id), want)
		}
		if !me.Graveyard.Contains(top.InstanceID) {
			t.Errorf("hit=%v: the top card was not milled", hit)
		}
	}
}

// --- Docent of Perfection // Final Iteration ---

func gapB12CastInstant(t *testing.T, g *game.Game) {
	t.Helper()
	effectstest.CastCatalogSpell(t, g, "Test Instant", "Instant", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
}

func TestGapB12DocentMakesWizardsAndTransformsAtThree(t *testing.T) {
	g, me := effectstest.IwTable(t)
	docent := effectstest.IwPut(t, g, gapB12Docent, me)
	for i := 1; i <= 3; i++ {
		gapB12CastInstant(t, g)
		if got := effectstest.CountBattlefieldNamed(g, me.ID, "Human Wizard"); got != i {
			t.Fatalf("after %d instants: %d Wizard tokens, want %d", i, got, i)
		}
		wantFace := 0
		if i == 3 {
			wantFace = 1
		}
		if effectstest.IwFace(g, docent) != wantFace {
			t.Fatalf("after %d instants: face %d, want %d", i, effectstest.IwFace(g, docent), wantFace)
		}
	}
	wizard := effectstest.BattlefieldIDNamed(g, me.ID, "Human Wizard")
	if got := effectstest.LayeredCard(t, g, wizard).CurrentPower(); got != 3 {
		t.Errorf("a Wizard token has power %d under Final Iteration, want 3", got)
	}
	if !effectstest.HasEffectiveKeyword(t, g, wizard, "flying") {
		t.Error("a Wizard token does not fly under Final Iteration")
	}
	gapB12CastInstant(t, g)
	if got := effectstest.CountBattlefieldNamed(g, me.ID, "Human Wizard"); got != 4 {
		t.Errorf("Final Iteration made %d Wizards in all, want 4", got)
	}
	if effectstest.IwFace(g, docent) != 1 {
		t.Error("Final Iteration turned back over")
	}
}

func TestGapB12DocentIgnoresACreatureSpell(t *testing.T) {
	g, me := effectstest.IwTable(t)
	effectstest.IwPut(t, g, gapB12Docent, me)
	effectstest.CastCatalogSpell(t, g, "A Bear", "Creature — Bear", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountBattlefieldNamed(g, me.ID, "Human Wizard"); got != 0 {
		t.Errorf("a creature spell made %d Wizards", got)
	}
}

// --- Town Gossipmonger // Incited Rabble ---

func TestGapB12GossipmongerTapsAnotherCreatureToTransform(t *testing.T) {
	g, me := effectstest.IwTable(t)
	gossip := effectstest.IwPut(t, g, gapB12Gossip, me)
	other := effectstest.B12Creature(g, me.ID, "Other", "Creature — Bear", 1, 1)
	if err := g.ActivateCatalogAbility(me.ID, gossip, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("activated without naming an untapped creature to tap")
	}
	if err := g.ActivateCatalogAbility(me.ID, gossip, 0, game.ActivateAbilityParams{TapIDs: []uuid.UUID{other}, Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.IwFace(g, gossip) != 1 {
		t.Fatalf("Town Gossipmonger is on face %d, want Incited Rabble", effectstest.IwFace(g, gossip))
	}
	if _, c := effectstest.CardWhere(g, me, other); !c.Tapped {
		t.Error("the other creature was not tapped")
	}
}

func TestGapB12IncitedRabbleMustAttackAndPumps(t *testing.T) {
	g, me := effectstest.IwTable(t)
	rabble := effectstest.IwPutBack(t, g, gapB12Gossip, me)
	if err := effectstest.HobB4Activate(t, g, me.ID, rabble, 0, effectstest.BroB04Pool("{2}"), nil); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.LayeredCard(t, g, rabble).CurrentPower(); got != 3 {
		t.Errorf("power %d after the pump, want 3", got)
	}
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	g.WithWriteLock(func() {
		if g.AttackRequirementsUnmetForEffect() == nil {
			t.Error("Incited Rabble owes an attack but nothing says so")
		}
	})
}

// --- Grizzled Angler // Grisly Anglerfish ---

func TestGapB12AnglerMillsTwoAndTransformsForAColorlessCreatureCard(t *testing.T) {
	for _, hit := range []bool{true, false} {
		g, me := effectstest.IwTable(t)
		angler := effectstest.IwPut(t, g, gapB12Angler, me)
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
		effectstest.PassPriorityAroundTable(t, g)
		if got := me.Graveyard.Size() - before; got != 2 {
			t.Errorf("hit=%v: milled %d, want 2", hit, got)
		}
		want := 0
		if hit {
			want = 1
		}
		if effectstest.IwFace(g, angler) != want {
			t.Errorf("hit=%v: face %d, want %d", hit, effectstest.IwFace(g, angler), want)
		}
	}
}

func TestGapB12AnglerfishForcesOpposingCreaturesToAttack(t *testing.T) {
	g, me := effectstest.IwTable(t)
	oppSeat := (g.Turn.ActiveSeat + 1) % len(g.Seats)
	opp := g.Seats[oppSeat]
	fish := effectstest.IwPutBack(t, g, gapB12Angler, me)
	effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMainOf(t, g, oppSeat)
	if err := effectstest.HobB4Activate(t, g, me.ID, fish, 0, effectstest.BroB04Pool("{6}"), nil); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	g.WithWriteLock(func() {
		if g.AttackRequirementsUnmetForEffect() == nil {
			t.Error("the opponent's creature is not required to attack")
		}
	})
}

// --- Deadeye Navigator ---

func TestGapB12NavigatorGrantsABlinkToBothPairedCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	id := effectstest.SbondHand(me, "Deadeye Navigator", "Creature — Spirit", gapB12NavigatorOracle, 5, 5, game.KeywordSoulbond)
	effectstest.SbondCast(t, g, me, id)
	effectstest.NghcPassOnce(t, g)
	effectstest.AnswerOwnPermanents(t, g, me.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.SbondPartner(g, id) != bear {
		t.Fatal("the Navigator is not paired with the Bear")
	}
	// The Bear now has the granted ability: blink it.
	if err := effectstest.HobB4Activate(t, g, me.ID, bear, 0, effectstest.BroB04Pool("{1}{U}"), nil); err != nil {
		t.Fatalf("activate the granted blink on the Bear: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(bear) {
		// The returning creature is a new object with a new ID on some paths.
		if effectstest.BattlefieldIDNamed(g, me.ID, "Bear") == uuid.Nil {
			t.Error("the Bear did not return to the battlefield")
		}
	}
	if effectstest.SbondPartner(g, id) != uuid.Nil {
		t.Error("the pair did not end when the Bear was blinked")
	}
}

// --- Thing in the Ice // Awoken Horror ---

func TestGapB12ThingInTheIceEntersWithFourIceCounters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	thing := effectstest.CastHoldCreature(t, g, "Thing in the Ice", "Creature — Horror", gapB12ThingOracle, 0, 4)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.LayeredCard(t, g, thing).Counters["ice"]; got != 4 {
		t.Errorf("%d ice counters, want 4", got)
	}
	effectstest.AssertKeywords(t, g, thing, "defender")
}

func TestGapB12ThingInTheIceTransformsOnTheLastCounterAndBouncesNonHorrors(t *testing.T) {
	g, me := effectstest.IwTable(t)
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	thing := gapB12PutBackType(t, g, gapB12Thing, me, "Creature — Kraken Horror")
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == thing {
				g.Battlefield.Cards[i].Counters = map[string]int{"ice": 2}
			}
		}
	})
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	horror := effectstest.B12Creature(g, opp.ID, "Other Horror", "Creature — Horror", 1, 1)
	effectstest.CastCatalogSpell(t, g, "Creature Spell", "Creature — Bear", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	gapB12CastInstant(t, g)
	if got := effectstest.LayeredCard(t, g, thing).Counters["ice"]; got != 1 {
		t.Fatalf("%d ice counters after one instant (a creature spell does not count), want 1", got)
	}
	if effectstest.IwFace(g, thing) != 0 {
		t.Fatal("transformed with an ice counter left")
	}
	gapB12CastInstant(t, g)
	if effectstest.IwFace(g, thing) != 1 {
		t.Fatalf("Thing in the Ice is on face %d with no ice counters, want Awoken Horror", effectstest.IwFace(g, thing))
	}
	if !me.Hand.Contains(bear) || !opp.Hand.Contains(theirs) {
		t.Error("non-Horror creatures were not returned to their owners' hands")
	}
	if !g.Battlefield.Contains(horror) || !g.Battlefield.Contains(thing) {
		t.Error("a Horror was bounced")
	}
}
