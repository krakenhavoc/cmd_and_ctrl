package testsl

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func TestLciQRUZEveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		effectstest.LciQRUZWhirlpoolOracle:   "Quicksand Whirlpool",
		effectstest.LciQRUZSpiketailOracle:   "Rampaging Spiketail",
		effectstest.LciQRUZRayOfRuinOracle:   "Ray of Ruin",
		effectstest.LciQRUZPrairieOracle:     "Restless Prairie",
		effectstest.LciQRUZRidgelineOracle:   "Restless Ridgeline",
		effectstest.LciQRUZVentsOracle:       "Restless Vents",
		effectstest.LciQRUZHeraldGuideOracle: "River Herald Guide",
		effectstest.LciQRUZHeraldScoutOracle: "River Herald Scout",
		effectstest.LciQRUZLurkerBatOracle:   "Ruin-Lurker Bat",
		effectstest.LciQRUZRockslideOracle:   "Rumbling Rockslide",
		effectstest.LciQRUZBoulderOracle:     "Runaway Boulder",
		effectstest.LciQRUZUnluckyDropOracle: "Unlucky Drop",
		effectstest.LciQRUZInquisitorOracle:  "Vito's Inquisitor",
		effectstest.LciQRUZWanderglyphOracle: "Volatile Wanderglyph",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

// --- fixtures ------------------------------------------------------------------

// lciQRUZLandsControlledBy counts the lands a player controls.
func lciQRUZLandsControlledBy(g *game.Game, player uuid.UUID) int {
	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.Controller == player && c.IsLand() {
			n++
		}
	}
	return n
}

// --- Quicksand Whirlpool -----------------------------------------------------------

func TestLciQRUZQuicksandWhirlpoolCostsThreeLessAgainstATappedCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	untapped := effectstest.B12Creature(g, opp.ID, "Ox", "Creature — Ox", 2, 2)
	tapped := effectstest.B12Creature(g, opp.ID, "Elk", "Creature — Elk", 2, 2)
	g.WithWriteLock(func() { effectstest.B16Tap(g, tapped) })
	effectstest.AdvanceToMain(t, g)
	card := game.Card{Name: "Quicksand Whirlpool", TypeLine: "Instant", ManaCost: "{5}{W}", OracleID: effectstest.LciQRUZWhirlpoolOracle}
	if err := effectstest.SocB19CastWithPool(g, me, card, "W", 3, game.CastSpellParams{Targets: effectstest.CardRefs(untapped)}); err == nil {
		t.Fatal("cast with three mana against an untapped creature")
	}
	card.InstanceID = uuid.New()
	if err := effectstest.SocB19CastWithPool(g, me, card, "W", 3, game.CastSpellParams{Targets: effectstest.CardRefs(tapped)}); err != nil {
		t.Fatalf("cast for {2}{W} against a tapped creature: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(tapped) || !g.Exile.Contains(tapped) {
		t.Error("the tapped creature was not exiled")
	}
	if !g.Battlefield.Contains(untapped) {
		t.Error("the untapped creature left the battlefield")
	}
}

func TestLciQRUZQuicksandWhirlpoolAtFullPriceExilesAnUntappedCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ox := effectstest.B12Creature(g, opp.ID, "Ox", "Creature — Ox", 2, 2)
	effectstest.AdvanceToMain(t, g)
	card := game.Card{Name: "Quicksand Whirlpool", TypeLine: "Instant", ManaCost: "{5}{W}", OracleID: effectstest.LciQRUZWhirlpoolOracle}
	if err := effectstest.SocB19CastWithPool(g, me, card, "W", 6, game.CastSpellParams{Targets: effectstest.CardRefs(ox)}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Exile.Contains(ox) {
		t.Error("the creature was not exiled")
	}
}

// --- Rampaging Spiketail -----------------------------------------------------------

func TestLciQRUZRampagingSpiketailPumpsAndShieldsACreatureYouControl(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	spiketail := effectstest.CastHoldCreature(t, g, "Rampaging Spiketail", "Creature — Dinosaur", effectstest.LciQRUZSpiketailOracle, 5, 6)
	pick := effectstest.LatestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("no enters-trigger target prompt")
	}
	if effectstest.HasID(pick.PickTargetCards, theirs) || !effectstest.HasID(pick.PickTargetCards, bear) || !effectstest.HasID(pick.PickTargetCards, spiketail) {
		t.Errorf("offered %v; want the creatures you control only", pick.PickTargetCards)
	}
	effectstest.PickCard(t, g, me.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.EffectivePower(t, g, bear), effectstest.EffectiveToughness(t, g, bear); p != 4 || tough != 2 {
		t.Errorf("the targeted bear is %d/%d, want 4/2", p, tough)
	}
	if !effectstest.HasEffectiveKeyword(t, g, bear, "indestructible") {
		t.Error("the targeted bear did not gain indestructible")
	}
	if effectstest.HasEffectiveKeyword(t, g, spiketail, "indestructible") {
		t.Error("the Spiketail itself gained indestructible")
	}
	effectstest.EndTurn(t, g)
	if p := effectstest.EffectivePower(t, g, bear); p != 2 || effectstest.HasEffectiveKeyword(t, g, bear, "indestructible") {
		t.Errorf("the effect outlasted the turn: power %d", p)
	}
}

func TestLciQRUZRampagingSpiketailSwampcyclesForASwamp(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	ids := effectstest.SeedSearchLibrary(me,
		effectstest.SearchTestLand("Swamp", "Basic Land — Swamp"),
		effectstest.SearchTestLand("Forest", "Basic Land — Forest"),
	)
	card := effectstest.PushCatalogHandCard(me, "Rampaging Spiketail", "Creature — Dinosaur", effectstest.LciQRUZSpiketailOracle)
	effectstest.FillPool(me, 2)
	if err := g.ActivateCatalogAbility(me.ID, card, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("swampcycle: %v", err)
	}
	if !me.Graveyard.Contains(card) {
		t.Error("the Spiketail was not discarded to swampcycle")
	}
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.SearchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no search prompt")
	}
	if len(c.SearchCards) != 1 || c.SearchCards[0] != ids[0] {
		t.Fatalf("search offers %v, want only the Swamp", c.SearchCards)
	}
	if err := g.ResolveSearchLibrary(c.ID, me.ID, []uuid.UUID{ids[0]}); err != nil {
		t.Fatalf("ResolveSearchLibrary: %v", err)
	}
	if !me.Hand.Contains(ids[0]) {
		t.Error("the Swamp did not go to the hand")
	}
}

// --- Ray of Ruin -------------------------------------------------------------------

func TestLciQRUZRayOfRuinTargetsACreatureVehicleOrNonbasicLand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	basic := effectstest.B12Permanent(g, opp.ID, "Their Swamp", "Basic Land — Swamp")
	if err := effectstest.CastCatalogSpellErr(t, g, "Ray of Ruin", "Sorcery", effectstest.LciQRUZRayOfRuinOracle, effectstest.CardRefs(basic)); err == nil {
		t.Error("a basic land was accepted as the target")
	}
	for _, tc := range []struct {
		name, typeLine string
		power          int
	}{
		{"a creature", "Creature — Bear", 2},
		{"a noncreature Vehicle", "Artifact — Vehicle", 0},
		{"a nonbasic land", "Land", 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			var target uuid.UUID
			if tc.power > 0 {
				target = effectstest.B12Creature(g, opp.ID, "Their Thing", tc.typeLine, tc.power, tc.power)
			} else {
				target = effectstest.B12Permanent(g, opp.ID, "Their Thing", tc.typeLine)
			}
			effectstest.SeedLibrary(me, "Top")
			effectstest.CastCatalogSpell(t, g, "Ray of Ruin", "Sorcery", effectstest.LciQRUZRayOfRuinOracle, effectstest.CardRefs(target))
			effectstest.PassPriorityAroundTable(t, g)
			if g.Battlefield.Contains(target) || !g.Exile.Contains(target) {
				t.Error("the target was not exiled")
			}
			sc := effectstest.ScryChoiceFor(g, me.ID)
			if sc == nil {
				t.Fatal("no scry 1 after the exile")
			}
			if len(sc.ScryCards) != 1 {
				t.Errorf("scry looks at %d cards, want 1", len(sc.ScryCards))
			}
		})
	}
}

// --- Restless Prairie --------------------------------------------------------------

func TestLciQRUZRestlessPrairieEntersTappedAndBecomesALlama(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	id := effectstest.B12PlayFromHand(t, g, "Restless Prairie", "Land", effectstest.LciQRUZPrairieOracle, game.CastSpellParams{})
	if !effectstest.IsTapped(g, id) {
		t.Error("Restless Prairie entered untapped")
	}
	g2 := effectstest.NewCatalogGame(t)
	prairie := effectstest.LciQRUZRestless(t, g2, "Restless Prairie", effectstest.LciQRUZPrairieOracle, 2, "G", "W")
	effectstest.LciQRUZAssertAnimated(t, g2, prairie, "Llama", []string{"G", "W"}, 3, 3)
}

func TestLciQRUZRestlessPrairieAttackPumpsOtherCreaturesYouControl(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	bear := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	prairie := effectstest.LciQRUZRestless(t, g, "Restless Prairie", effectstest.LciQRUZPrairieOracle, 2, "G", "W")
	effectstest.DeclareAttack(t, g, opp.ID, prairie)
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.EffectivePower(t, g, bear), effectstest.EffectiveToughness(t, g, bear); p != 3 || tough != 3 {
		t.Errorf("your other creature is %d/%d, want 3/3", p, tough)
	}
	if p, tough := effectstest.EffectivePower(t, g, prairie), effectstest.EffectiveToughness(t, g, prairie); p != 3 || tough != 3 {
		t.Errorf("the Prairie is %d/%d, want 3/3 (it is not an other creature)", p, tough)
	}
	if p := effectstest.EffectivePower(t, g, theirs); p != 2 {
		t.Errorf("an opponent's creature has power %d, want 2", p)
	}
}

// --- Restless Ridgeline ------------------------------------------------------------

func TestLciQRUZRestlessRidgelineBecomesADinosaur(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	id := effectstest.B12PlayFromHand(t, g, "Restless Ridgeline", "Land", effectstest.LciQRUZRidgelineOracle, game.CastSpellParams{})
	if !effectstest.IsTapped(g, id) {
		t.Error("Restless Ridgeline entered untapped")
	}
	g2 := effectstest.NewCatalogGame(t)
	ridge := effectstest.LciQRUZRestless(t, g2, "Restless Ridgeline", effectstest.LciQRUZRidgelineOracle, 2, "R", "G")
	effectstest.LciQRUZAssertAnimated(t, g2, ridge, "Dinosaur", []string{"R", "G"}, 3, 4)
}

func TestLciQRUZRestlessRidgelinePumpsAndUntapsAnotherAttacker(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	bear := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	idle := effectstest.B12Creature(g, me.ID, "My Idle Bear", "Creature — Bear", 2, 2)
	ridge := effectstest.LciQRUZRestless(t, g, "Restless Ridgeline", effectstest.LciQRUZRidgelineOracle, 2, "R", "G")
	effectstest.DeclareAttack(t, g, opp.ID, ridge, bear)
	pick := effectstest.LatestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("no attack-trigger target prompt")
	}
	if effectstest.HasID(pick.PickTargetCards, ridge) || effectstest.HasID(pick.PickTargetCards, idle) || !effectstest.HasID(pick.PickTargetCards, bear) {
		t.Errorf("offered %v; want another attacking creature only", pick.PickTargetCards)
	}
	if !effectstest.TappedNow(g, bear) {
		t.Fatal("the attacking bear is not tapped")
	}
	effectstest.PickCard(t, g, me.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.EffectivePower(t, g, bear), effectstest.EffectiveToughness(t, g, bear); p != 4 || tough != 2 {
		t.Errorf("the target is %d/%d, want 4/2", p, tough)
	}
	if effectstest.TappedNow(g, bear) {
		t.Error("the target was not untapped")
	}
	if !effectstest.TappedNow(g, ridge) {
		t.Error("the Ridgeline itself untapped")
	}
}

// --- Restless Vents ----------------------------------------------------------------

func TestLciQRUZRestlessVentsBecomesAnInsectWithMenace(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	id := effectstest.B12PlayFromHand(t, g, "Restless Vents", "Land", effectstest.LciQRUZVentsOracle, game.CastSpellParams{})
	if !effectstest.IsTapped(g, id) {
		t.Error("Restless Vents entered untapped")
	}
	g2 := effectstest.NewCatalogGame(t)
	vents := effectstest.LciQRUZRestless(t, g2, "Restless Vents", effectstest.LciQRUZVentsOracle, 1, "B", "R")
	effectstest.LciQRUZAssertAnimated(t, g2, vents, "Insect", []string{"B", "R"}, 2, 3)
	if !effectstest.HasEffectiveKeyword(t, g2, vents, "menace") {
		t.Error("the animated Vents has no menace")
	}
	effectstest.EndTurn(t, g2)
	if c := effectstest.S58p6View(t, g2, vents); c.IsCreature() || effectstest.HasEffectiveKeyword(t, g2, vents, "menace") {
		t.Error("the animation outlasted the turn")
	}
}

func TestLciQRUZRestlessVentsAttackMayRummage(t *testing.T) {
	for _, accept := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
		effectstest.SeedLibrary(me, "Drawn")
		discard := effectstest.PushHandCard(g, me)
		vents := effectstest.LciQRUZRestless(t, g, "Restless Vents", effectstest.LciQRUZVentsOracle, 1, "B", "R")
		hand, library := me.Hand.Size(), me.Library.Size()
		effectstest.DeclareAttack(t, g, opp.ID, vents)
		effectstest.PassPriorityAroundTable(t, g)
		if accept {
			effectstest.AnswerDiscard(t, g, me.ID, discard)
		} else {
			effectstest.AnswerDiscard(t, g, me.ID)
		}
		effectstest.PassPriorityAroundTable(t, g)
		wantDrawn := 0
		if accept {
			wantDrawn = 1
		}
		if got := library - me.Library.Size(); got != wantDrawn {
			t.Errorf("accept=%v: drew %d, want %d", accept, got, wantDrawn)
		}
		if me.Hand.Size() != hand {
			t.Errorf("accept=%v: hand %d -> %d, want unchanged", accept, hand, me.Hand.Size())
		}
		if accept && !me.Graveyard.Contains(discard) {
			t.Error("the chosen card was not discarded")
		}
	}
}

// --- River Herald Guide and Scout ----------------------------------------------------

func TestLciQRUZRiverHeraldGuideHasVigilanceAndExploresALandToHand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	top := effectstest.PutOnTopForTest(me, "Forest", "Basic Land — Forest")
	guide := effectstest.Msc17Play(t, g, me, game.Card{Name: "River Herald Guide", TypeLine: "Creature — Merfolk Scout", OracleID: effectstest.LciQRUZHeraldGuideOracle, Power: 3, Toughness: 1})
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(top) {
		t.Error("the revealed land did not go to the hand")
	}
	if got := effectstest.CountersOn(g, guide, game.CounterPlusOne); got != 0 {
		t.Errorf("%d +1/+1 counters after revealing a land, want 0", got)
	}
	if !effectstest.HasEffectiveKeyword(t, g, guide, "vigilance") {
		t.Error("River Herald Guide has no vigilance")
	}
}

func TestLciQRUZRiverHeraldScoutExploresANonlandForACounter(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	top := effectstest.PutOnTopForTest(me, "Some Sorcery", "Sorcery")
	scout := effectstest.Msc17Play(t, g, me, game.Card{Name: "River Herald Scout", TypeLine: "Creature — Merfolk Scout", OracleID: effectstest.LciQRUZHeraldScoutOracle, Power: 1, Toughness: 2})
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountersOn(g, scout, game.CounterPlusOne); got != 1 {
		t.Errorf("%d +1/+1 counters, want 1", got)
	}
	effectstest.AnswerExploreQuestion(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(top) || me.Library.Contains(top) {
		t.Error("the revealed card did not go to the graveyard when the player chose to")
	}
}

// --- Ruin-Lurker Bat ---------------------------------------------------------------

func TestLciQRUZRuinLurkerBatScriesAtYourEndStepIfYouDescended(t *testing.T) {
	for _, tc := range []struct {
		name     string
		typeLine string
		scry     bool
	}{
		{"a permanent card milled", "Artifact", true},
		{"only an instant card milled", "Instant", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[0]
			bat := effectstest.B12Push(g, me.ID, "Ruin-Lurker Bat", "Creature — Bat", effectstest.LciQRUZLurkerBatOracle, 1, 1)
			if !effectstest.HasEffectiveKeyword(t, g, bat, "flying") || !effectstest.HasEffectiveKeyword(t, g, bat, "lifelink") {
				t.Error("Ruin-Lurker Bat is missing flying or lifelink")
			}
			effectstest.LciDGMillOne(t, g, me, tc.typeLine)
			effectstest.SeedLibrary(me, "Top")
			effectstest.AdvanceToEndStepOf(t, g, 0)
			if !tc.scry {
				if effectstest.TriggersOnStackFrom(g, bat) != 0 {
					t.Fatal("the end-step trigger went on the stack without a descent")
				}
				return
			}
			effectstest.PassPriorityAroundTable(t, g)
			sc := effectstest.ScryChoiceFor(g, me.ID)
			if sc == nil {
				t.Fatal("no scry 1 after a descent")
			}
			if len(sc.ScryCards) != 1 {
				t.Errorf("scry looks at %d cards, want 1", len(sc.ScryCards))
			}
		})
	}
}

// --- Rumbling Rockslide ------------------------------------------------------------

func TestLciQRUZRumblingRockslideDealsDamageEqualToYourLands(t *testing.T) {
	for _, tc := range []struct {
		name  string
		extra int // toughness above the land count
		dies  bool
	}{
		{"toughness equal to your lands dies", 0, true},
		{"one more toughness survives", 1, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			for i := 0; i < 3; i++ {
				effectstest.B12Permanent(g, me.ID, "Mountain", "Basic Land — Mountain")
			}
			// An opponent's lands are not yours: these would kill the
			// survivor if they were counted.
			for i := 0; i < 4; i++ {
				effectstest.B12Permanent(g, opp.ID, "Their Island", "Basic Land — Island")
			}
			lands := lciQRUZLandsControlledBy(g, me.ID)
			target := effectstest.B12Creature(g, opp.ID, "Their Giant", "Creature — Giant", 1, lands+tc.extra)
			effectstest.CastCatalogSpell(t, g, "Rumbling Rockslide", "Sorcery", effectstest.LciQRUZRockslideOracle, effectstest.CardRefs(target))
			effectstest.PassPriorityAroundTable(t, g)
			if died := !g.Battlefield.Contains(target); died != tc.dies {
				t.Errorf("with %d lands against toughness %d: died=%v, want %v", lands, lands+tc.extra, died, tc.dies)
			}
		})
	}
}

// --- Runaway Boulder ---------------------------------------------------------------

func TestLciQRUZRunawayBoulderDealsSixToACreatureAnOpponentControls(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := effectstest.B12Creature(g, me.ID, "My Giant", "Creature — Giant", 6, 6)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Giant", "Creature — Giant", 6, 6)
	effectstest.CastHoldCreature(t, g, "Runaway Boulder", "Artifact", effectstest.LciQRUZBoulderOracle, 0, 0)
	pick := effectstest.LatestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("no enters-trigger target prompt")
	}
	if effectstest.HasID(pick.PickTargetCards, mine) || !effectstest.HasID(pick.PickTargetCards, theirs) {
		t.Errorf("offered %v; want only a creature an opponent controls", pick.PickTargetCards)
	}
	effectstest.PickCard(t, g, me.ID, theirs)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(theirs) {
		t.Error("the opponent's 6/6 survived 6 damage")
	}
	if !g.Battlefield.Contains(mine) {
		t.Error("your own creature died")
	}
}

func TestLciQRUZRunawayBoulderHasFlash(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	// Cartographer's Companion, an artifact without flash, is refused here.
	slow := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: slow, Name: "Cartographer's Companion", TypeLine: "Artifact Creature — Gnome",
		OracleID: "460c0f0b-bbf4-4307-9b44-350175d83bfe", Power: 2, Toughness: 1, Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, slow, game.CastSpellParams{}); err == nil {
		t.Fatal("an artifact without flash was cast at the beginning of combat")
	}
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Runaway Boulder", TypeLine: "Artifact", OracleID: effectstest.LciQRUZBoulderOracle,
		Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast at the beginning of combat: %v", err)
	}
}

func TestLciQRUZRunawayBoulderCycles(t *testing.T) {
	effectstest.DftSTCycles(t, "Runaway Boulder", "Artifact", effectstest.LciQRUZBoulderOracle, 0)
}

// --- Unlucky Drop ------------------------------------------------------------------

func TestLciQRUZUnluckyDropLetsTheOwnerChooseTopOrBottom(t *testing.T) {
	for _, tc := range []struct {
		name   string
		bottom bool
	}{{"top", false}, {"bottom", true}} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			rock := effectstest.PushArtifactForTest(g, opp.ID, "Their Mind Stone")
			effectstest.CastCatalogSpell(t, g, "Unlucky Drop", "Instant", effectstest.LciQRUZUnluckyDropOracle, effectstest.CardRefs(rock))
			effectstest.PassPriorityAroundTable(t, g)
			if effectstest.LatestChoiceOfKindFor(g, game.PendingChoicePutInLibrary, me.ID) != nil {
				t.Fatal("the caster was asked; the owner chooses")
			}
			c := effectstest.LatestChoiceOfKindFor(g, game.PendingChoicePutInLibrary, opp.ID)
			if c == nil {
				t.Fatalf("the owner was not asked: %+v", g.PendingChoices)
			}
			var top, bottom []uuid.UUID
			if tc.bottom {
				bottom = []uuid.UUID{rock}
			} else {
				top = []uuid.UUID{rock}
			}
			if err := g.ResolvePutInLibrary(c.ID, opp.ID, bottom, top); err != nil {
				t.Fatalf("ResolvePutInLibrary: %v", err)
			}
			if g.Battlefield.Contains(rock) {
				t.Fatal("the artifact is still on the battlefield")
			}
			cards := opp.Library.Cards
			// The bottom of a library is Cards[0], the top the last entry.
			if tc.bottom && cards[0].InstanceID != rock {
				t.Errorf("the bottom card is %q, want the artifact", cards[0].Name)
			}
			if !tc.bottom && cards[len(cards)-1].InstanceID != rock {
				t.Errorf("the top card is %q, want the artifact", cards[len(cards)-1].Name)
			}
		})
	}
}

func TestLciQRUZUnluckyDropCantTargetALand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	land := effectstest.B12Permanent(g, opp.ID, "Their Island", "Basic Land — Island")
	if err := effectstest.CastCatalogSpellErr(t, g, "Unlucky Drop", "Instant", effectstest.LciQRUZUnluckyDropOracle, effectstest.CardRefs(land)); err == nil {
		t.Error("a land was accepted as the target")
	}
}

// --- Vito's Inquisitor -------------------------------------------------------------

func TestLciQRUZVitosInquisitorSacrificesAnotherForACounterAndMenace(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	inq := effectstest.B12Push(g, me.ID, "Vito's Inquisitor", "Creature — Vampire Knight", effectstest.LciQRUZInquisitorOracle, 3, 3)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	rock := effectstest.B12Permanent(g, me.ID, "Mana Rock", "Artifact")
	shrine := effectstest.B12Permanent(g, me.ID, "Shrine", "Enchantment")
	effectstest.AdvanceToMain(t, g)
	effectstest.FillPoolColored(me, "B", 1)
	if err := g.ActivateCatalogAbility(me.ID, inq, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{inq}}); err == nil {
		t.Fatal("the Inquisitor sacrificed itself")
	}
	if err := g.ActivateCatalogAbility(me.ID, inq, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{shrine}}); err == nil {
		t.Fatal("an enchantment is neither a creature nor an artifact")
	}
	effectstest.B16Activate(t, g, me.ID, inq, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{bear}})
	if g.Battlefield.Contains(bear) {
		t.Error("the sacrificed creature is still on the battlefield")
	}
	if got := effectstest.CountersOn(g, inq, game.CounterPlusOne); got != 1 {
		t.Errorf("%d +1/+1 counters, want 1", got)
	}
	if !effectstest.HasEffectiveKeyword(t, g, inq, "menace") {
		t.Error("the Inquisitor did not gain menace")
	}
	effectstest.FillPoolColored(me, "B", 1)
	effectstest.B16Activate(t, g, me.ID, inq, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{rock}})
	if got := effectstest.CountersOn(g, inq, game.CounterPlusOne); got != 2 {
		t.Errorf("%d +1/+1 counters after sacrificing an artifact too, want 2", got)
	}
	effectstest.EndTurn(t, g)
	if effectstest.HasEffectiveKeyword(t, g, inq, "menace") {
		t.Error("menace outlasted the turn")
	}
	if got := effectstest.CountersOn(g, inq, game.CounterPlusOne); got != 2 {
		t.Errorf("%d +1/+1 counters after the turn, want 2 (counters stay)", got)
	}
}

// --- Volatile Wanderglyph ----------------------------------------------------------

func TestLciQRUZVolatileWanderglyphMayRummageWhenItBecomesTapped(t *testing.T) {
	for _, accept := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		effectstest.SeedLibrary(me, "Drawn")
		discard := effectstest.PushHandCard(g, me)
		glyph := effectstest.B12Push(g, me.ID, "Volatile Wanderglyph", "Artifact Creature — Golem", effectstest.LciQRUZWanderglyphOracle, 2, 2)
		effectstest.AdvanceToMain(t, g)
		if effectstest.DiscardChoiceFor(g, me.ID) != nil {
			t.Fatal("a rummage is offered before the Wanderglyph has tapped")
		}
		hand, library := me.Hand.Size(), me.Library.Size()
		effectstest.DeclareAttack(t, g, opp.ID, glyph)
		effectstest.PassPriorityAroundTable(t, g)
		if accept {
			effectstest.AnswerDiscard(t, g, me.ID, discard)
		} else {
			effectstest.AnswerDiscard(t, g, me.ID)
		}
		effectstest.PassPriorityAroundTable(t, g)
		wantDrawn := 0
		if accept {
			wantDrawn = 1
		}
		if got := library - me.Library.Size(); got != wantDrawn {
			t.Errorf("accept=%v: drew %d, want %d", accept, got, wantDrawn)
		}
		if me.Hand.Size() != hand {
			t.Errorf("accept=%v: hand %d -> %d, want unchanged", accept, hand, me.Hand.Size())
		}
	}
}
