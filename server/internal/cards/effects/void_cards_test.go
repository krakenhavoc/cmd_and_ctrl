package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// void_cards_test.go — #2128: void (CR 207.2c) and the Edge of Eternities
// cards that read it. The seam itself (the two tally cells) is pinned in
// game/turn_tally_void_test.go; these pin each card, and the one thing
// every end-step card shares: an intervening if.

const (
	voidforgedTitanOracle      = "f7afe534-fbb0-4f8b-9e98-aa49d4005465"
	kavaronSkywardenOracle     = "c1f1ba17-b5e8-4bb3-a767-befd52e2db72"
	insatiableSkittermawOracle = "69a68dec-4b01-4528-81fc-a3c5f5097f72"
	interceptorMechanOracle    = "ded2b1b7-8c12-4f6f-a3be-a8f7b62b0819"
	elegyAcolyteOracle         = "f7d52eac-edc9-464b-a163-07c136aa6e4c"
	tragicTrajectoryOracle     = "e5e19326-3e90-439c-a3b2-749d43a1af0c"
	plasmaBoltOracle           = "ac5258c1-5b46-4892-a445-bfd56875f039"
	decodeTransmissionsOracle  = "42510758-6167-420b-bb70-65f914b7a259"
	hymnOfTheFallerOracle      = "ab1379b7-0016-4d2a-a83b-952666e43b89"
	temporalInterventionOracle = "5e8a2416-e311-41ac-a3cd-be1c2441d542"
	hylderbladeOracle          = "64ca59b8-6e90-46c8-8047-6172b9c84226"
	choraleOfTheVoidOracle     = "cd50bff0-c461-4023-b62a-6746eaaeab0e"
	alpharaelOracle            = "ed43f1f9-ef96-47bc-8a54-e59154319f9d"
	rovingActuatorOracle       = "78ac1e6c-2704-4844-9e87-642324c23997"
)

// voidFor turns void on the cheapest way: an artifact nobody controls
// anything for is sacrificed. Any seat's permanent will do, which is the
// point of a table-wide fact.
func voidFor(t *testing.T, g *game.Game, owner uuid.UUID) {
	t.Helper()
	rock := pushCatalogPermanent(g, owner, "Relic", "Artifact", "", false)
	g.WithWriteLock(func() {
		if err := g.SacrificePermanentForEffect(rock); err != nil {
			t.Fatalf("setup sacrifice: %v", err)
		}
	})
}

// --- the end-step payoffs --------------------------------------------

func TestVoidforgedTitanDrawsAndLosesLifeOnlyWithVoid(t *testing.T) {
	// No void: nothing happens at the end step.
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushCatalogPermanent(g, me.ID, "Voidforged Titan", "Artifact Creature — Robot Warrior", voidforgedTitanOracle, false)
	hand, life := me.Hand.Size(), me.Life
	advanceTo(t, g, game.StepEnd)
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand || me.Life != life {
		t.Fatalf("no void: hand %d → %d, life %d → %d", hand, me.Hand.Size(), life, me.Life)
	}

	// A land leaving is not void.
	g2 := newCatalogGame(t)
	me2 := g2.Seats[0]
	pushCatalogPermanent(g2, me2.ID, "Voidforged Titan", "Artifact Creature — Robot Warrior", voidforgedTitanOracle, false)
	forest := pushCatalogPermanent(g2, me2.ID, "Forest", "Basic Land — Forest", "", false)
	g2.WithWriteLock(func() { _ = g2.SacrificePermanentForEffect(forest) })
	hand2 := me2.Hand.Size()
	advanceTo(t, g2, game.StepEnd)
	passPriorityAroundTable(t, g2)
	if me2.Hand.Size() != hand2 {
		t.Fatalf("a land leaving drew a card")
	}

	// Void, off an OPPONENT's permanent: one card and one life.
	g3 := newCatalogGame(t)
	me3, opp3 := g3.Seats[0], g3.Seats[1]
	pushCatalogPermanent(g3, me3.ID, "Voidforged Titan", "Artifact Creature — Robot Warrior", voidforgedTitanOracle, false)
	voidFor(t, g3, opp3.ID)
	hand3, life3 := me3.Hand.Size(), me3.Life
	advanceTo(t, g3, game.StepEnd)
	passPriorityAroundTable(t, g3)
	if me3.Hand.Size() != hand3+1 || me3.Life != life3-1 {
		t.Errorf("void: hand %d → %d (want +1), life %d → %d (want -1)", hand3, me3.Hand.Size(), life3, me3.Life)
	}
}

// A spell cast for its warp cost is void with no permanent leaving.
func TestVoidforgedTitanSeesAWarpedSpell(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushCatalogPermanent(g, me.ID, "Voidforged Titan", "Artifact Creature — Robot Warrior", voidforgedTitanOracle, false)
	warpCreature(t, g)
	hand := me.Hand.Size()
	advanceTo(t, g, game.StepEnd)
	eorSettle(t, g) // the warp exile and the Titan trigger are ordered by their controller
	if me.Hand.Size() != hand+1 {
		t.Errorf("warped spell: hand %d → %d, want +1", hand, me.Hand.Size())
	}

	// The same creature cast the ordinary way is not warped.
	g2 := newCatalogGame(t)
	me2 := g2.Seats[0]
	pushCatalogPermanent(g2, me2.ID, "Voidforged Titan", "Artifact Creature — Robot Warrior", voidforgedTitanOracle, false)
	castAndResolveCreature(t, g2, "Weftstalker Ardent", "Creature — Drix Artificer", weftstalkerWarpOracle)
	passPriorityAroundTable(t, g2)
	hand2 := me2.Hand.Size()
	advanceTo(t, g2, game.StepEnd)
	passPriorityAroundTable(t, g2)
	if me2.Hand.Size() != hand2 {
		t.Errorf("an ordinary cast turned void on: hand %d → %d", hand2, me2.Hand.Size())
	}
}

// A restore taken while void is live still pays off at the end step.
func TestVoidforgedTitanPaysOffAfterARestore(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushCatalogPermanent(g, me.ID, "Voidforged Titan", "Artifact Creature — Robot Warrior", voidforgedTitanOracle, false)
	voidFor(t, g, g.Seats[1].ID)
	back := restoreThroughJSON(t, g)
	hand := back.Seats[0].Hand.Size()
	advanceTo(t, back, game.StepEnd)
	passPriorityAroundTable(t, back)
	if back.Seats[0].Hand.Size() != hand+1 {
		t.Errorf("restored game: hand %d → %d, want +1", hand, back.Seats[0].Hand.Size())
	}
}

func TestVoidCounterCreaturesGrowOnlyWithVoid(t *testing.T) {
	for _, c := range []struct{ name, typeLine, oracle string }{
		{"Kavaron Skywarden", "Creature — Kavu Soldier", kavaronSkywardenOracle},
		{"Insatiable Skittermaw", "Creature — Insect Horror", insatiableSkittermawOracle},
		{"Interceptor Mechan", "Artifact Creature — Robot", interceptorMechanOracle},
	} {
		t.Run(c.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[0]
			id := pushCatalogPermanent(g, me.ID, c.name, c.typeLine, c.oracle, false)
			advanceTo(t, g, game.StepEnd)
			passPriorityAroundTable(t, g)
			if n := battlefieldPlusOnes(g, id); n != 0 {
				t.Fatalf("no void: %d counters", n)
			}

			g2 := newCatalogGame(t)
			id2 := pushCatalogPermanent(g2, g2.Seats[0].ID, c.name, c.typeLine, c.oracle, false)
			voidFor(t, g2, g2.Seats[0].ID)
			advanceTo(t, g2, game.StepEnd)
			passPriorityAroundTable(t, g2)
			if n := battlefieldPlusOnes(g2, id2); n != 1 {
				t.Errorf("void: %d counters, want 1", n)
			}
		})
	}
}

// The trigger is the OPPONENT's end step to nobody: only its controller's
// end step counts.
func TestVoidEndStepTriggersOnlyOnYourOwnEndStep(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	id := pushCatalogPermanent(g, opp.ID, "Kavaron Skywarden", "Creature — Kavu Soldier", kavaronSkywardenOracle, false)
	voidFor(t, g, opp.ID)
	advanceTo(t, g, game.StepEnd)
	passPriorityAroundTable(t, g)
	if n := battlefieldPlusOnes(g, id); n != 0 {
		t.Errorf("another player's end step grew it: %d counters", n)
	}
}

func TestInterceptorMechanReturnsAnArtifactOrCreatureCard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := pushGraveyardPermanent(me, "Grizzly Bears", "Creature — Bear", "{1}{G}")
	relic := pushGraveyardPermanent(me, "Sol Ring", "Artifact", "{1}")
	bolt := pushGraveyardPermanent(me, "Lightning Bolt", "Instant", "{R}")
	theirs := pushGraveyardPermanent(g.Seats[1], "Serra Angel", "Creature — Angel", "{3}{W}{W}")

	castAndResolveCreature(t, g, "Interceptor Mechan", "Artifact Creature — Robot", interceptorMechanOracle)
	prompt := latestPickTarget(g, me.ID)
	if prompt == nil {
		t.Fatal("no target prompt")
	}
	if !hasID(prompt.PickTargetCards, bear) || !hasID(prompt.PickTargetCards, relic) {
		t.Errorf("legal set %v should hold the creature and the artifact", prompt.PickTargetCards)
	}
	if hasID(prompt.PickTargetCards, bolt) || hasID(prompt.PickTargetCards, theirs) {
		t.Errorf("legal set %v holds an instant or an opponent's card", prompt.PickTargetCards)
	}
	pickCard(t, g, me.ID, relic)
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(relic) {
		t.Error("the artifact card did not return to hand")
	}
}

func TestElegyAcolyteMakesARobotWithVoidAndDrawsOnceForABatch(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushCatalogPermanent(g, me.ID, "Elegy Acolyte", "Creature — Human Cleric", elegyAcolyteOracle, false)
	advanceTo(t, g, game.StepEnd)
	passPriorityAroundTable(t, g)
	if n := onBattlefieldNamed(g, "Robot"); n != 0 {
		t.Fatalf("no void made %d Robots", n)
	}

	g2 := newCatalogGame(t)
	me2 := g2.Seats[0]
	pushCatalogPermanent(g2, me2.ID, "Elegy Acolyte", "Creature — Human Cleric", elegyAcolyteOracle, false)
	warpCreature(t, g2)
	advanceTo(t, g2, game.StepEnd)
	eorSettle(t, g2)
	if n := onBattlefieldNamed(g2, "Robot"); n != 1 {
		t.Errorf("void (a warped spell): %d Robots, want 1", n)
	}
	var robot game.Card
	for _, c := range g2.Battlefield.Cards {
		if c.Name == "Robot" {
			robot = c
		}
	}
	if robot.Power != 2 || robot.Toughness != 2 || !robot.IsArtifact() {
		t.Errorf("token = %s %d/%d, want a 2/2 artifact Robot", robot.TypeLine, robot.Power, robot.Toughness)
	}

	// Two creatures connecting at once draw ONE card and cost ONE life.
	g3 := newCatalogGame(t)
	me3, opp3 := g3.Seats[0], g3.Seats[1]
	pushCatalogPermanent(g3, me3.ID, "Elegy Acolyte", "Creature — Human Cleric", elegyAcolyteOracle, false)
	a := pushVanillaCreature(g3, me3.ID, "Bear A", 2, 2)
	b := pushVanillaCreature(g3, me3.ID, "Bear B", 2, 2)
	hand, life := me3.Hand.Size(), me3.Life
	attackWith(t, g3, opp3.ID, a, b)
	passPriorityAroundTable(t, g3)
	if me3.Hand.Size() != hand+1 || me3.Life != life-1 {
		t.Errorf("batch: hand %d → %d (want +1), life %d → %d (want -1)", hand, me3.Hand.Size(), life, me3.Life)
	}
}

// --- the spells --------------------------------------------------------

func TestTragicTrajectoryIsTenWithVoid(t *testing.T) {
	for _, tc := range []struct {
		name string
		void bool
		dies bool
	}{{"no void is -2/-2", false, false}, {"void is -10/-10", true, true}} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			victim := b12Creature(g, g.Seats[1].ID, "Big Bear", "Creature — Bear", 4, 4)
			if tc.void {
				voidFor(t, g, g.Seats[2].ID)
			}
			castCatalogSpell(t, g, "Tragic Trajectory", "Sorcery", tragicTrajectoryOracle,
				[]game.TargetRef{{Kind: game.TargetCard, ID: victim}})
			passPriorityAroundTable(t, g)
			if died := !g.Battlefield.Contains(victim); died != tc.dies {
				t.Errorf("destroyed = %v, want %v", died, tc.dies)
			}
		})
	}
}

// Void turned on by the warp cost of an earlier spell, with nothing
// leaving the battlefield.
func TestTragicTrajectoryCountsAWarpedSpell(t *testing.T) {
	g := newCatalogGame(t)
	victim := b12Creature(g, g.Seats[1].ID, "Big Bear", "Creature — Bear", 4, 4)
	warpCreature(t, g)
	castCatalogSpell(t, g, "Tragic Trajectory", "Sorcery", tragicTrajectoryOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: victim}})
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(victim) {
		t.Error("a warped spell did not turn void on")
	}
}

func TestPlasmaBoltDealsThreeWithVoid(t *testing.T) {
	for _, tc := range []struct {
		void bool
		want int
	}{{false, 2}, {true, 3}} {
		g := newCatalogGame(t)
		opp := g.Seats[1]
		if tc.void {
			voidFor(t, g, g.Seats[2].ID)
		}
		life := opp.Life
		castCatalogSpell(t, g, "Plasma Bolt", "Sorcery", plasmaBoltOracle,
			[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
		passPriorityAroundTable(t, g)
		if got := life - opp.Life; got != tc.want {
			t.Errorf("void=%v: %d damage, want %d", tc.void, got, tc.want)
		}
	}
}

func TestDecodeTransmissionsDrainsInsteadWithVoid(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	castCatalogSpell(t, g, "Decode Transmissions", "Sorcery", decodeTransmissionsOracle, nil)
	hand, life := me.Hand.Size(), me.Life
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+2 || me.Life != life-2 {
		t.Errorf("no void: hand %d → %d (want +2), life %d → %d (want -2)", hand, me.Hand.Size(), life, me.Life)
	}
	for _, p := range g.Seats[1:] {
		if p.Life != 40 {
			t.Errorf("no void: %s lost life", p.Name)
		}
	}

	g2 := newCatalogGame(t)
	me2 := g2.Seats[0]
	voidFor(t, g2, g2.Seats[1].ID)
	castCatalogSpell(t, g2, "Decode Transmissions", "Sorcery", decodeTransmissionsOracle, nil)
	hand2, life2 := me2.Hand.Size(), me2.Life
	passPriorityAroundTable(t, g2)
	if me2.Hand.Size() != hand2+2 || me2.Life != life2 {
		t.Errorf("void: hand %d → %d (want +2), life %d → %d (want unchanged)", hand2, me2.Hand.Size(), life2, me2.Life)
	}
	for _, p := range g2.Seats[1:] {
		if p.Life != 38 {
			t.Errorf("void: %s life = %d, want 38", p.Name, p.Life)
		}
	}
}

func TestHymnOfTheFallerSurveilsDrawsAndDrawsAgainWithVoid(t *testing.T) {
	for _, void := range []bool{false, true} {
		g := newCatalogGame(t)
		me := g.Seats[0]
		seedLibrary(me, "Draw C", "Draw B", "Draw A", "Bin Me")
		if void {
			voidFor(t, g, g.Seats[1].ID)
		}
		castCatalogSpell(t, g, "Hymn of the Faller", "Sorcery", hymnOfTheFallerOracle, nil)
		hand, life := me.Hand.Size(), me.Life
		passPriorityAroundTable(t, g)

		c := surveilChoiceFor(g, me.ID)
		if c == nil {
			t.Fatalf("void=%v: no surveil prompt", void)
		}
		if me.Hand.Size() != hand || me.Life != life {
			t.Fatalf("void=%v: the draw happened before the surveil was answered", void)
		}
		// Bin the top card: the card drawn is the one beneath it.
		if err := g.ResolveSurveil(c.ID, me.ID, c.ScryCards, nil); err != nil {
			t.Fatalf("ResolveSurveil: %v", err)
		}
		passPriorityAroundTable(t, g)
		want := 1
		if void {
			want = 2
		}
		if got := me.Hand.Size() - hand; got != want {
			t.Errorf("void=%v: drew %d, want %d", void, got, want)
		}
		if me.Life != life-1 {
			t.Errorf("void=%v: life %d → %d, want -1", void, life, me.Life)
		}
		if me.Hand.Cards[me.Hand.Size()-1].Name == "Bin Me" {
			t.Errorf("void=%v: drew the surveilled card", void)
		}
	}
}

func TestTemporalInterventionCostsTwoLessWithVoid(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	card := game.Card{InstanceID: uuid.New(), Name: "Temporal Intervention", TypeLine: "Sorcery",
		ManaCost: "{2}{B}", OracleID: temporalInterventionOracle, Owner: me.ID, Controller: me.ID}
	me.Hand.PushTop(card)
	price := func() int {
		t.Helper()
		p, err := g.PriceCast(me.ID, card, game.CastSpellParams{})
		if err != nil {
			t.Fatalf("PriceCast: %v", err)
		}
		return p.Total.ManaValue()
	}
	if got := price(); got != 3 {
		t.Fatalf("no void: mana value %d, want 3", got)
	}
	voidFor(t, g, g.Seats[1].ID)
	if got := price(); got != 1 {
		t.Fatalf("void: mana value %d, want 1", got)
	}

	// And the charge agrees: {B} floating pays for it.
	for g.Turn.Step != game.StepPrecombatMain && g.Turn.Step != game.StepPostcombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	me.ManaPool.AddMana(game.ManaToken{Color: "B"})
	if err := g.CastSpell(me.ID, card.InstanceID, game.CastSpellParams{
		Strict:  true,
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: g.Seats[1].ID}},
	}); err != nil {
		t.Fatalf("CastSpell for {B} with void: %v", err)
	}
	if n := len(me.ManaPool); n != 0 {
		t.Errorf("%d mana left floating", n)
	}
}

func TestTemporalInterventionCostsTwoLessAfterAWarp(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	card := game.Card{InstanceID: uuid.New(), Name: "Temporal Intervention", TypeLine: "Sorcery",
		ManaCost: "{2}{B}", OracleID: temporalInterventionOracle, Owner: me.ID, Controller: me.ID}
	me.Hand.PushTop(card)
	warpCreature(t, g)
	p, err := g.PriceCast(me.ID, card, game.CastSpellParams{})
	if err != nil {
		t.Fatalf("PriceCast: %v", err)
	}
	if got := p.Total.ManaValue(); got != 1 {
		t.Errorf("after a warped spell: mana value %d, want 1", got)
	}
}

func TestTemporalInterventionTakesANonlandCard(t *testing.T) {
	g := newCatalogGame(t)
	caster, victim := g.Seats[0], g.Seats[1]
	ids := revealHand(victim, rhForest(), rhBolt())
	castCatalogSpell(t, g, "Temporal Intervention", "Sorcery", temporalInterventionOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: victim.ID}})
	passPriorityAroundTable(t, g)
	pick := openPick(t, g)
	if !sameIDs(pick.DiscardOptions, ids[1:]) {
		t.Errorf("options = %v, want only the Bolt", pick.DiscardOptions)
	}
	if err := g.ResolvePendingChoice(pick.ID, caster.ID, []uuid.UUID{ids[1]}); err != nil {
		t.Fatalf("taking the Bolt: %v", err)
	}
	g.SettleResolution()
	if !victim.Graveyard.Contains(ids[1]) || !victim.Hand.Contains(ids[0]) {
		t.Error("the Bolt should be discarded and the Forest kept")
	}
}

// --- the permanents ----------------------------------------------------

func TestHylderbladeAttachesAtTheEndStepOnlyWithVoid(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	blade := pushCatalogPermanent(g, me.ID, "Hylderblade", "Artifact — Equipment", hylderbladeOracle, false)
	bear := b12Creature(g, me.ID, "Grizzly Bears", "Creature — Bear", 2, 2)
	advanceTo(t, g, game.StepEnd)
	passPriorityAroundTable(t, g)
	if c, _ := cardOnBattlefield(g, blade); c.AttachedTo.ID == bear {
		t.Fatal("attached with no void")
	}

	g2 := newCatalogGame(t)
	me2 := g2.Seats[0]
	blade2 := pushCatalogPermanent(g2, me2.ID, "Hylderblade", "Artifact — Equipment", hylderbladeOracle, false)
	bear2 := b12Creature(g2, me2.ID, "Grizzly Bears", "Creature — Bear", 2, 2)
	voidFor(t, g2, g2.Seats[1].ID)
	advanceTo(t, g2, game.StepEnd)
	pickCard(t, g2, me2.ID, bear2)
	passPriorityAroundTable(t, g2)
	c, ok := cardOnBattlefield(g2, blade2)
	if !ok || c.AttachedTo.ID != bear2 {
		t.Errorf("void: equipped to %v, want the Bear", c.AttachedTo.ID)
	}
	if power := currentPowerFor(t, g2, bear2); power != 5 {
		t.Errorf("equipped Bear power = %d, want 5", power)
	}
}

func TestChoraleOfTheVoidSacrificesItselfUnlessVoid(t *testing.T) {
	setup := func() (*game.Game, uuid.UUID) {
		g := newCatalogGame(t)
		me := g.Seats[0]
		host := b12Creature(g, me.ID, "Grizzly Bears", "Creature — Bear", 2, 2)
		aura := pushCatalogPermanent(g, me.ID, "Chorale of the Void", "Enchantment — Aura", choraleOfTheVoidOracle, false)
		g.WithWriteLock(func() {
			if err := g.AttachForEffect(aura, game.TargetRef{Kind: game.TargetCard, ID: host}); err != nil {
				t.Fatalf("attach: %v", err)
			}
		})
		return g, aura
	}

	g, aura := setup()
	advanceTo(t, g, game.StepEnd)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(aura) {
		t.Error("no void: the Aura was not sacrificed")
	}

	g2, aura2 := setup()
	voidFor(t, g2, g2.Seats[1].ID)
	advanceTo(t, g2, game.StepEnd)
	passPriorityAroundTable(t, g2)
	if !g2.Battlefield.Contains(aura2) {
		t.Error("void: the Aura was sacrificed anyway")
	}

	// It is NOT an intervening if: a warp cast made in response, while the
	// trigger is on the stack, still saves it. Here the only void comes
	// after the trigger went on the stack.
	g3, aura3 := setup()
	advanceTo(t, g3, game.StepEnd)
	if !triggerOnStackFrom(g3, "Chorale of the Void") {
		t.Fatal("the end-step trigger did not go on the stack with no void (it is not an intervening if)")
	}
	voidFor(t, g3, g3.Seats[1].ID)
	passPriorityAroundTable(t, g3)
	if !g3.Battlefield.Contains(aura3) {
		t.Error("void arriving while the trigger was on the stack did not save the Aura")
	}
}

func TestChoraleOfTheVoidRaisesFromTheDefendingPlayersGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, third := g.Seats[0], g.Seats[1], g.Seats[2]
	host := pushCatalogPermanent(g, me.ID, "Grizzly Bears", "Creature — Bear", "", false)
	aura := pushCatalogPermanent(g, me.ID, "Chorale of the Void", "Enchantment — Aura", choraleOfTheVoidOracle, false)
	g.WithWriteLock(func() {
		if err := g.AttachForEffect(aura, game.TargetRef{Kind: game.TargetCard, ID: host}); err != nil {
			t.Fatalf("attach: %v", err)
		}
	})
	angel := pushGraveyardPermanent(opp, "Serra Angel", "Creature — Angel", "{3}{W}{W}")
	relic := pushGraveyardPermanent(opp, "Sol Ring", "Artifact", "{1}")
	elsewhere := pushGraveyardPermanent(third, "Hill Giant", "Creature — Giant", "{3}{R}")
	mine := pushGraveyardPermanent(me, "Runeclaw Bear", "Creature — Bear", "{1}{G}")

	declareAttack(t, g, opp.ID, host)
	prompt := latestPickTarget(g, me.ID)
	if prompt == nil {
		t.Fatal("no target prompt for the attack trigger")
	}
	if !hasID(prompt.PickTargetCards, angel) {
		t.Errorf("legal set %v misses the defending player's creature", prompt.PickTargetCards)
	}
	for _, bad := range []uuid.UUID{relic, elsewhere, mine} {
		if hasID(prompt.PickTargetCards, bad) {
			t.Errorf("legal set %v holds a card outside the defending player's creature cards", prompt.PickTargetCards)
		}
	}
	pickCard(t, g, me.ID, angel)
	passPriorityAroundTable(t, g)
	answerOptionPick(t, g, me.ID, 0)

	var raised game.Card
	var found bool
	g.ReadSnapshot(func() {
		if g.Battlefield.Contains(angel) {
			raised, found = g.LookupCardForEffect(angel)
		}
	})
	if !found {
		t.Fatal("the Angel did not enter the battlefield")
	}
	if raised.Controller != me.ID || !raised.Tapped || raised.AttackingTarget == uuid.Nil {
		t.Errorf("controller=%v tapped=%v attacking=%v, want mine, tapped and attacking", raised.Controller, raised.Tapped, raised.AttackingTarget)
	}
}

func TestAlpharaelHalvesTheDefendersLifeOnlyWithVoid(t *testing.T) {
	for _, tc := range []struct {
		void bool
		life int
		want int
	}{{false, 40, 40}, {true, 40, 20}, {true, 21, 10}} {
		g := newCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		boss := pushCatalogPermanent(g, me.ID, "Alpharael, Stonechosen", "Legendary Creature — Human Cleric", alpharaelOracle, false)
		opp.Life = tc.life
		if tc.void {
			voidFor(t, g, g.Seats[2].ID)
		}
		declareAttack(t, g, opp.ID, boss)
		passPriorityAroundTable(t, g)
		// Alpharael itself connects for 3 afterwards only in the damage step;
		// the trigger resolves in the declare-attackers step.
		if opp.Life != tc.want {
			t.Errorf("void=%v life %d: defender at %d, want %d", tc.void, tc.life, opp.Life, tc.want)
		}
		_ = boss
	}
}

// Only the defending player loses the life, not every opponent.
func TestAlpharaelOnlyHurtsTheDefendingPlayer(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	boss := pushCatalogPermanent(g, me.ID, "Alpharael, Stonechosen", "Legendary Creature — Human Cleric", alpharaelOracle, false)
	voidFor(t, g, g.Seats[2].ID)
	declareAttack(t, g, opp.ID, boss)
	passPriorityAroundTable(t, g)
	if g.Seats[2].Life != 40 || g.Seats[3].Life != 40 || me.Life != 40 {
		t.Errorf("lives: me %d, third %d, fourth %d; want untouched", me.Life, g.Seats[2].Life, g.Seats[3].Life)
	}
}

// Ward—Discard a card at random: the payer decides only whether to pay.
func TestAlpharaelWardDiscardsAtRandom(t *testing.T) {
	g, me, opp, warded := wardTable(t, "Alpharael, Stonechosen", "Legendary Creature — Human Cleric", alpharaelOracle)
	fillHand(opp, "Filler A", "Filler B", "Filler C")
	boltWardedCreature(t, g, opp, warded)

	prompt := passUntilConfirmFor(t, g, opp.ID)
	if prompt == nil {
		t.Fatal("no ward prompt for the spell's controller")
	}
	if confirmChoiceFor(g, me.ID) != nil {
		t.Error("the ward permanent's controller must not be asked to pay")
	}
	hand, yard := opp.Hand.Size(), opp.Graveyard.Size()
	if err := g.ResolveConfirm(prompt.ID, opp.ID, true); err != nil {
		t.Fatalf("ResolveConfirm: %v", err)
	}
	if n := discardOwed(g, opp.ID); n != 0 {
		t.Fatalf("the payer was asked to choose %d cards, but the discard is at random", n)
	}
	if opp.Hand.Size() != hand-1 || opp.Graveyard.Size() != yard+1 {
		t.Errorf("hand %d → %d, graveyard %d → %d: want exactly one card discarded", hand, opp.Hand.Size(), yard, opp.Graveyard.Size())
	}
	passPriorityAroundTable(t, g)
	if !boltResolved(g, warded) {
		t.Error("the ward was paid — the Lightning Bolt should have resolved")
	}
}

// An empty hand cannot pay: the spell is countered with no prompt.
func TestAlpharaelWardCountersWithAnEmptyHand(t *testing.T) {
	g, _, opp, warded := wardTable(t, "Alpharael, Stonechosen", "Legendary Creature — Human Cleric", alpharaelOracle)
	opp.Hand.Cards = nil
	bolt := boltWardedCreature(t, g, opp, warded)
	passPriorityAroundTable(t, g)
	if boltResolved(g, warded) {
		t.Error("an empty hand paid a discard ward")
	}
	if !opp.Graveyard.Contains(bolt) {
		t.Error("the countered spell should be in its owner's graveyard")
	}
}

func TestRovingActuatorCastsACheapSpellFromTheGraveyardWithVoid(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bolt := gyCard(me, "Lightning Bolt", "Instant", "{R}", lightningBoltOracle)
	gyCard(me, "Divination", "Sorcery", "{2}{U}", "")
	creature := gyCard(me, "Grizzly Bears", "Creature — Bear", "{1}{G}", "")
	voidFor(t, g, g.Seats[2].ID)

	castAndResolveCreature(t, g, "Roving Actuator", "Artifact Creature — Robot", rovingActuatorOracle)
	prompt := latestPickTarget(g, me.ID)
	if prompt == nil {
		t.Fatal("void: no target prompt")
	}
	if !hasID(prompt.PickTargetCards, bolt) || hasID(prompt.PickTargetCards, creature) || len(prompt.PickTargetCards) != 1 {
		t.Errorf("legal set %v should hold only the Bolt (mana value 2 or less, instant or sorcery)", prompt.PickTargetCards)
	}
	pickCard(t, g, me.ID, bolt)
	passPriorityAroundTable(t, g)
	if !inExile(g, bolt) {
		t.Error("the Bolt should have been exiled")
	}
	cast := resolutionCastFor(g, me.ID)
	if cast == nil {
		t.Fatal("no offer to cast the copy")
	}
	cp := copyOf(t, g, "Lightning Bolt")
	life := opp.Life
	if err := g.CastForResolutionChoice(me.ID, cast.ID, cp.InstanceID, game.CastSpellParams{
		FromZone: "exile", Targets: []game.TargetRef{playerTarget(opp.ID)},
	}); err != nil {
		t.Fatalf("casting the copy for free: %v", err)
	}
	passPriorityAroundTable(t, g)
	if opp.Life != life-3 {
		t.Errorf("opponent life %d → %d, want -3 from the free Bolt copy", life, opp.Life)
	}
	if !inExile(g, bolt) {
		t.Error("the original stays in exile")
	}

	// Declining the copy leaves nothing behind.
	g2 := newCatalogGame(t)
	me2 := g2.Seats[0]
	bolt2 := gyCard(me2, "Lightning Bolt", "Instant", "{R}", lightningBoltOracle)
	voidFor(t, g2, g2.Seats[2].ID)
	castAndResolveCreature(t, g2, "Roving Actuator", "Artifact Creature — Robot", rovingActuatorOracle)
	pickCard(t, g2, me2.ID, bolt2)
	passPriorityAroundTable(t, g2)
	declineResolutionCasts(t, g2, me2.ID)
	if n := len(cardCopies(g2)); n != 0 {
		t.Errorf("%d copies left after declining", n)
	}

	// No void: the ability does not trigger at all.
	g3 := newCatalogGame(t)
	me3 := g3.Seats[0]
	gyCard(me3, "Lightning Bolt", "Instant", "{R}", lightningBoltOracle)
	castAndResolveCreature(t, g3, "Roving Actuator", "Artifact Creature — Robot", rovingActuatorOracle)
	if latestPickTarget(g3, me3.ID) != nil {
		t.Error("no void: the enters trigger asked for a target")
	}
}
