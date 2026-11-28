package testss82m

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b838_cards_test.go — Ice Age's cumulative-upkeep cards, played from the
// real catalog entries (S82 OLD batch 838, #5636).

const (
	b838MusicianOracle = "0d6e9468-adb2-4feb-87cf-9113db8717b1"
	b838TwistOracle    = "2c6f9fdb-ee27-4cd8-9d4d-1ea8b2830f73"
	b838MaggotsOracle  = "2d6a387f-7ad4-419e-951a-9dcd4c9ac823"
	b838PollenOracle   = "2f18967d-c2cd-449d-a288-df6343672743"
	b838SnapOracle     = "3dcbcf5e-bdbe-46b3-881e-5ac4030c6e38"
	b838StormOracle    = "567d3de7-8d56-4b6d-a59a-f8674172f595"
	b838HallsOracle    = "57d7deaf-ee08-49ab-b239-1c77a1f65225"
	b838BlizzardOracle = "698feaaa-2fff-4809-81d1-4caad44c6c0a"
	b838ChasmOracle    = "73e7a2ad-d11c-4867-b97d-f971809da778"
	b838PigeonsOracle  = "2d30a634-43c6-4e78-ba10-7ef7ee17ac40"
)

func TestS82OldB838CardsAreRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		b838MusicianOracle: "Musician",
		b838TwistOracle:    "Reality Twist",
		b838MaggotsOracle:  "Flow of Maggots",
		b838PollenOracle:   "Fyndhorn Pollen",
		b838SnapOracle:     "Cold Snap",
		b838StormOracle:    "Energy Storm",
		b838HallsOracle:    "Halls of Mist",
		b838BlizzardOracle: "Blizzard",
		b838ChasmOracle:    "Glacial Chasm",
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

func TestFyndhornPollenShrinksEveryCreatureAndPumpsTheShrinkForTheTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	mine := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 3, 3)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	pollen := effectstest.PushCatalogPermanent(g, me.ID, "Fyndhorn Pollen", "Enchantment", b838PollenOracle, false)
	if p, tough := effectstest.PtOf(t, g, mine); p != 2 || tough != 3 {
		t.Errorf("my bear is %d/%d, want 2/3", p, tough)
	}
	if p, _ := effectstest.PtOf(t, g, theirs); p != 1 {
		t.Errorf("their bear has power %d, want 1", p)
	}
	effectstest.ApaMana(me, "G", "G")
	effectstest.B16Activate(t, g, me.ID, pollen, 0, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if p, _ := effectstest.PtOf(t, g, mine); p != 1 {
		t.Errorf("after the pump, my bear has power %d, want 1", p)
	}
	if p, _ := effectstest.PtOf(t, g, theirs); p != 0 {
		t.Errorf("after the pump, their bear has power %d, want 0", p)
	}
	effectstest.AdvancePastCleanupForTest(t, g)
	if p, _ := effectstest.PtOf(t, g, mine); p != 2 {
		t.Errorf("next turn my bear has power %d, want 2 (static only)", p)
	}
}

func TestColdSnapCumulativeUpkeepAgesPaysAndSacrifices(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[1]
	snap := effectstest.PushCatalogPermanent(g, me.ID, "Cold Snap", "Enchantment", b838SnapOracle, false)
	effectstest.AdvanceToUpkeepOf(t, g, 1)
	effectstest.AnswerAnyTriggerOrderPrompt(t, g, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
	prompt := effectstest.RemoraUpkeepPrompt(t, g, me.ID)
	if got := effectstest.AgeCountersOn(t, g, snap); got != 1 {
		t.Fatalf("age counters = %d, want 1", got)
	}
	if prompt.PayCost != "{2}" {
		t.Fatalf("cost = %q, want {2}", prompt.PayCost)
	}
	effectstest.AnswerPayUnless(t, g, me.ID, false)
	if _, ok := effectstest.BattlefieldCard(g, snap); ok {
		t.Error("an unpaid cumulative upkeep did not sacrifice Cold Snap")
	}
}

func TestColdSnapDamagesEachPlayerForTheirSnowLands(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Cold Snap", "Enchantment", b838SnapOracle, false)
	effectstest.PushLandFor(g, opp.ID, "Snow-Covered Island", "Basic Snow Land — Island")
	effectstest.PushLandFor(g, opp.ID, "Snow-Covered Swamp", "Basic Snow Land — Swamp")
	effectstest.PushLandFor(g, opp.ID, "Mountain", "Basic Land — Mountain")
	life := opp.Life
	effectstest.AdvanceToUpkeepOf(t, g, 1)
	effectstest.PassPriorityAroundTable(t, g)
	if got := life - opp.Life; got != 2 {
		t.Errorf("the opponent lost %d life, want 2 (two snow lands)", got)
	}
	// A player with no snow land is not damaged.
	myLife := me.Life
	effectstest.LeaveUpkeepThenAdvanceTo(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != myLife {
		t.Errorf("a player with no snow land lost %d life", myLife-me.Life)
	}
}

func TestFlowOfMaggotsCanBeBlockedOnlyByWalls(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	maggots := effectstest.ApaPush(g, me.ID, me.ID, game.Card{Name: "Flow of Maggots", TypeLine: "Creature — Insect", OracleID: b838MaggotsOracle, Power: 2, Toughness: 2})
	bear := effectstest.ApaPush(g, opp.ID, opp.ID, game.Card{Name: "Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2})
	wall := effectstest.ApaPush(g, opp.ID, opp.ID, game.Card{Name: "Wall", TypeLine: "Creature — Wall", Power: 0, Toughness: 4})
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(maggots, opp.ID); err != nil {
		t.Fatal(err)
	}
	effectstest.LockInAttacks(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(bear, maggots); err == nil {
		t.Error("a non-Wall creature blocked Flow of Maggots")
	}
	if err := g.DeclareBlocker(wall, maggots); err != nil {
		t.Errorf("a Wall could not block Flow of Maggots: %v", err)
	}
}

func b838CastBlizzard(t *testing.T, g *game.Game, p *game.Player) error {
	t.Helper()
	effectstest.AdvanceToMain(t, g)
	id := uuid.New()
	p.Hand.PushTop(game.Card{InstanceID: id, Name: "Blizzard", TypeLine: "Enchantment", ManaCost: "{G}{G}", OracleID: b838BlizzardOracle, Owner: p.ID, Controller: p.ID})
	effectstest.ApaMana(p, "G", "G")
	return g.CastSpell(p.ID, id, game.CastSpellParams{Strict: true})
}

func TestBlizzardIsCastOnlyWithASnowLand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushLandFor(g, me.ID, "Forest", "Basic Land — Forest")
	if err := b838CastBlizzard(t, g, me); err == nil {
		t.Fatal("cast without a snow land: accepted")
	}
	effectstest.PushLandFor(g, me.ID, "Snow-Covered Forest", "Basic Snow Land — Forest")
	if err := b838CastBlizzard(t, g, me); err != nil {
		t.Fatalf("cast with a snow land: %v", err)
	}
}

func TestBlizzardAndEnergyStormKeepFlyersTapped(t *testing.T) {
	for _, tc := range []struct {
		name, oracle string
	}{{"Blizzard", b838BlizzardOracle}, {"Energy Storm", b838StormOracle}} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			effectstest.PushCatalogPermanent(g, me.ID, tc.name, "Enchantment", tc.oracle, false)
			flyer := effectstest.PushTappedForTest(g, opp.ID, "Carrier Pigeons", b838PigeonsOracle, "Creature — Bird")
			walker := effectstest.PushTappedForTest(g, opp.ID, "Bear", "", "Creature — Bear")
			effectstest.AdvanceToUpkeepOf(t, g, 1)
			if !effectstest.TappedForTest(t, g, flyer) {
				t.Error("a creature with flying untapped")
			}
			if effectstest.TappedForTest(t, g, walker) {
				t.Error("a creature without flying stayed tapped")
			}
		})
	}
}

func TestEnergyStormPreventsDamageFromInstantsAndSorceriesOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	storm := effectstest.PushCatalogPermanent(g, me.ID, "Energy Storm", "Enchantment", b838StormOracle, false)
	effectstest.ApaMana(me, "R")
	life := opp.Life
	bolt := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: bolt, Name: "Lightning Bolt", TypeLine: "Instant", ManaCost: "{R}", OracleID: effectstest.LightningBoltOracle, Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, bolt, game.CastSpellParams{Targets: []game.TargetRef{effectstest.PlayerTarget(opp.ID)}}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life {
		t.Errorf("Lightning Bolt dealt %d, want 0", life-opp.Life)
	}
	// Damage from a permanent is not prevented.
	g.WithWriteLock(func() { _ = g.DealDamageToPlayerForEffect(storm, opp.ID, 2) })
	if got := life - opp.Life; got != 2 {
		t.Errorf("damage from a permanent: lost %d, want 2", got)
	}
}

func TestHallsOfMistStopsCreaturesThatAttackedLastTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, opp.ID, "Halls of Mist", "Land", b838HallsOracle, false)
	veteran := effectstest.PushTestPermanent(g, me.ID, game.Card{Name: "Veteran", TypeLine: "Creature — Bear", Power: 2, Toughness: 2})
	fresh := effectstest.PushTestPermanent(g, me.ID, game.Card{Name: "Fresh", TypeLine: "Creature — Bear", Power: 2, Toughness: 2})
	effectstest.AdvanceToDeclareAttackersOf(t, g, 0)
	if err := g.DeclareAttacker(veteran, opp.ID); err != nil {
		t.Fatalf("the first attack: %v", err)
	}
	b838GoAroundPaying(t, g)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := effectstest.DeclareResult(g, veteran, opp.ID); !errors.Is(err, game.ErrCantAttack) {
		t.Errorf("a creature that attacked last turn: err = %v, want ErrCantAttack", err)
	}
	if err := effectstest.DeclareResult(g, fresh, opp.ID); err != nil {
		t.Errorf("a creature that did not attack last turn could not attack: %v", err)
	}
}

func TestGlacialChasmEntersWithItsThreeEffects(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	forest := effectstest.PushLandFor(g, me.ID, "Forest", "Basic Land — Forest")
	bear := effectstest.ApaPush(g, me.ID, me.ID, game.Card{Name: "Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2})
	chasm := effectstest.PlayLandFromHand(t, g, "Glacial Chasm", b838ChasmOracle)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerChooseCards(t, g, me.ID, forest)
	effectstest.PassPriorityAroundTable(t, g)
	if _, ok := effectstest.BattlefieldCard(g, forest); ok {
		t.Error("the chosen land was not sacrificed")
	}
	if _, ok := effectstest.BattlefieldCard(g, chasm); !ok {
		t.Fatal("Glacial Chasm is not on the battlefield")
	}
	// Prevent all damage that would be dealt to you.
	life := me.Life
	g.WithWriteLock(func() { _ = g.DealDamageToPlayerForEffect(uuid.New(), me.ID, 3) })
	if me.Life != life {
		t.Errorf("lost %d life, want 0", life-me.Life)
	}
	// An opponent is not protected.
	oppLife := opp.Life
	g.WithWriteLock(func() { _ = g.DealDamageToPlayerForEffect(uuid.New(), opp.ID, 3) })
	if opp.Life != oppLife-3 {
		t.Errorf("the opponent lost %d life, want 3", oppLife-opp.Life)
	}
	// Creatures you control can't attack.
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(bear, opp.ID); err == nil {
		t.Error("a creature its controller controls attacked under Glacial Chasm")
	}
}

func TestGlacialChasmUpkeepCostsTwoLifePerAgeCounter(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[1]
	chasm := effectstest.PushCatalogPermanent(g, me.ID, "Glacial Chasm", "Land", b838ChasmOracle, false)
	effectstest.AdvanceToUpkeepOf(t, g, 1)
	effectstest.PassPriorityAroundTable(t, g)
	prompt := effectstest.RemoraUpkeepPrompt(t, g, me.ID)
	if a := prompt.PayAction(); a == nil || a.Kind != game.PayActionLife || a.Count != 2 {
		t.Fatalf("payment = %+v, want 2 life", a)
	}
	life := me.Life
	effectstest.AnswerPayUnless(t, g, me.ID, true)
	if _, ok := effectstest.BattlefieldCard(g, chasm); !ok {
		t.Error("a paid upkeep sacrificed Glacial Chasm")
	}
	if me.Life != life-2 {
		t.Errorf("life %d -> %d, want 2 paid", life, me.Life)
	}
}

func TestRealityTwistSwapsTheLandTypesThatMakeMana(t *testing.T) {
	for _, tc := range []struct {
		land, typeLine, want string
	}{
		{"Plains", "Basic Land — Plains", "R"},
		{"Swamp", "Basic Land — Swamp", "G"},
		{"Mountain", "Basic Land — Mountain", "W"},
		{"Forest", "Basic Land — Forest", "B"},
		{"Island", "Basic Land — Island", "U"},
	} {
		t.Run(tc.land, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			effectstest.SeedEnchantment(g, b838TwistOracle, "Reality Twist", "Enchantment", me.ID)
			land := effectstest.PushLandFor(g, opp.ID, tc.land, tc.typeLine)
			effectstest.TapForMana(t, g, opp.ID, land)
			if got := effectstest.SortedColors(opp); len(got) != 1 || got[0] != tc.want {
				t.Errorf("%s made %v, want [%s]", tc.land, got, tc.want)
			}
		})
	}
}

func TestMusicianPutsMusicCountersAndGivesTheUpkeepOnce(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	musician := effectstest.PushCatalogPermanent(g, me.ID, "Musician", "Creature — Human Wizard", b838MusicianOracle, false)
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.B16Activate(t, g, me.ID, musician, 0, game.ActivateAbilityParams{Targets: effectstest.CardTarget(bear)})
	effectstest.PassPriorityAroundTable(t, g)
	c, _ := effectstest.BattlefieldCard(g, bear)
	if c.Counters["music"] != 1 {
		t.Fatalf("music counters = %d, want 1", c.Counters["music"])
	}
	n := 0
	for _, tr := range game.TriggersForCard(c) {
		if tr.Key == "At the beginning of your upkeep, destroy this creature unless you pay {1} for each music counter on it" {
			n++
		}
	}
	if n != 1 {
		t.Fatalf("the creature has the upkeep ability %d times, want 1", n)
	}
	// A second counter does not give a second copy of the ability.
	effectstest.B16Tap(g, musician)
	g.WithWriteLock(func() {
		cc := effectstest.BattlefieldCardPtr(t, g, musician)
		cc.Tapped = false
	})
	effectstest.B16Activate(t, g, me.ID, musician, 0, game.ActivateAbilityParams{Targets: effectstest.CardTarget(bear)})
	effectstest.PassPriorityAroundTable(t, g)
	c, _ = effectstest.BattlefieldCard(g, bear)
	if c.Counters["music"] != 2 {
		t.Fatalf("music counters = %d, want 2", c.Counters["music"])
	}
	n = 0
	for _, tr := range game.TriggersForCard(c) {
		if tr.Key == "At the beginning of your upkeep, destroy this creature unless you pay {1} for each music counter on it" {
			n++
		}
	}
	if n != 1 {
		t.Fatalf("after a second counter the creature has the upkeep ability %d times, want 1", n)
	}
	// At its controller's upkeep it costs {2}; declining destroys it.
	effectstest.AdvanceToUpkeepOf(t, g, 1)
	effectstest.PassPriorityAroundTable(t, g)
	prompt := effectstest.RemoraUpkeepPrompt(t, g, opp.ID)
	if prompt.PayCost != "{2}" {
		t.Fatalf("cost = %q, want {2}", prompt.PayCost)
	}
	effectstest.AnswerPayUnless(t, g, opp.ID, false)
	if _, ok := effectstest.BattlefieldCard(g, bear); ok {
		t.Error("an unpaid music upkeep did not destroy the creature")
	}
}

// b838GoAroundPaying plays on until it is seat 0's precombat main again,
// paying every cumulative upkeep that asks on the way.
func b838GoAroundPaying(t *testing.T, g *game.Game) {
	t.Helper()
	left := false
	for i := 0; i < 600; i++ {
		if g.Turn.ActiveSeat != 0 {
			left = true
		}
		if left && g.Turn.ActiveSeat == 0 && g.Turn.Step == game.StepPrecombatMain {
			return
		}
		var open *game.PendingChoice
		for _, c := range g.PendingChoices {
			if c != nil && c.Kind == game.PendingChoicePayUnless {
				open = c
				break
			}
		}
		if open != nil {
			p := g.PlayerByIDForEffect(open.Chooser)
			for n := 0; n < 6; n++ {
				p.ManaPool.AddMana(game.ManaToken{Color: "C"})
			}
			if err := g.ResolvePayUnless(open.ID, open.Chooser, true); err != nil {
				t.Fatalf("pay: %v", err)
			}
			continue
		}
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
		effectstest.PassPriorityAroundTable(t, g)
	}
	t.Fatal("never came around to seat 0's main phase")
}
