package testsru

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_bro_b05_cards_test.go — S82 The Brothers' War, batch 5 (#5482):
// Aeronaut Cavalry; Air Marshal; Conscripted Infantry; Emergency Weld; Mass
// Production; Mishra's Onslaught; Veteran's Powerblade; Yotian Tactician;
// Zephyr Sentinel; Fortified Beachhead; Rescue Retriever; Siege Veteran.

const (
	broB05CavalryOracle    = "153292f9-e0d1-43a7-a083-275c49d0554a"
	broB05MarshalOracle    = "59b2edf0-00e4-4cd4-88ac-6dd74e9b1b14"
	broB05InfantryOracle   = "cf0f3aee-78b6-4a98-8e09-8861b73c2e74"
	broB05WeldOracle       = "86b0e978-5b4e-4bb7-bfda-be899c632d8e"
	broB05MassOracle       = "20150b58-a442-472c-8191-b44581322eeb"
	broB05OnslaughtOracle  = "3684edb5-bbf1-4ed7-b5b7-f24b452dbd64"
	broB05PowerbladeOracle = "da0b4e01-b312-4f37-a034-824ce58db02e"
	broB05TacticianOracle  = "0d9d7f04-6319-4596-8b74-de5a794936e5"
	broB05SentinelOracle   = "dcddd143-a7d9-4a38-a2da-dd6bd2e48205"
	broB05BeachheadOracle  = "387fe395-e4a0-4fb2-8d6c-88a1a21d2ed8"
	broB05RetrieverOracle  = "86f9897f-84aa-4c8d-8335-cfdb2f565311"
	broB05VeteranOracle    = "4d4fc571-a859-4460-a3c7-3f76a671ab1b"
)

func TestBroB05CardsAreRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		broB05CavalryOracle:    "Aeronaut Cavalry",
		broB05MarshalOracle:    "Air Marshal",
		broB05InfantryOracle:   "Conscripted Infantry",
		broB05WeldOracle:       "Emergency Weld",
		broB05MassOracle:       "Mass Production",
		broB05OnslaughtOracle:  "Mishra's Onslaught",
		broB05PowerbladeOracle: "Veteran's Powerblade",
		broB05TacticianOracle:  "Yotian Tactician",
		broB05SentinelOracle:   "Zephyr Sentinel",
		broB05BeachheadOracle:  "Fortified Beachhead",
		broB05RetrieverOracle:  "Rescue Retriever",
		broB05VeteranOracle:    "Siege Veteran",
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

// broB05Soldiers counts the Soldier artifact tokens a player controls.
func broB05Soldiers(g *game.Game, controller uuid.UUID) int {
	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.Controller == controller && effects.IsToken(c) && c.Name == "Soldier" && c.IsArtifact() && c.IsCreature() &&
			c.Power == 1 && c.Toughness == 1 && len(c.Colors) == 0 {
			n++
		}
	}
	return n
}

func broB05Cast(t *testing.T, g *game.Game, p *game.Player, name, typeLine, oracle string, params game.CastSpellParams) uuid.UUID {
	t.Helper()
	effectstest.AdvanceToMain(t, g)
	return effectstest.CastFromHandForTest(t, g, p, name, typeLine, "{0}", oracle, params)
}

func TestBroB05AeronautCavalryCountersAnotherSoldier(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	soldier := effectstest.B12Creature(g, me.ID, "My Soldier", "Creature — Human Soldier", 2, 2)
	bear := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Soldier", "Creature — Human Soldier", 2, 2)
	cav := effectstest.CastFromHandForTest(t, g, me, "Aeronaut Cavalry", "Creature — Human Soldier", "{0}", broB05CavalryOracle, game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if p := effectstest.LatestPickTarget(g, me.ID); p != nil {
		for _, bad := range []uuid.UUID{cav, bear, theirs} {
			if err := g.ResolvePickTarget(p.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: bad}); err == nil {
				t.Errorf("an illegal target %s was accepted", bad)
			}
		}
		effectstest.PickCard(t, g, me.ID, soldier)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CounterCount(g, soldier, game.CounterPlusOne); got != 1 {
		t.Errorf("the other Soldier has %d counters, want 1", got)
	}
	if effectstest.CounterCount(g, cav, game.CounterPlusOne) != 0 || effectstest.CounterCount(g, bear, game.CounterPlusOne) != 0 {
		t.Error("a counter landed on the Cavalry or a non-Soldier")
	}
	if !effectstest.HasEffectiveKeyword(t, g, cav, "flying") {
		t.Error("the Cavalry lacks flying")
	}
}

func TestBroB05AirMarshalGivesASoldierFlying(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	marshal := effectstest.MirB1Push(g, me.ID, "Air Marshal", "Creature — Human Soldier", broB05MarshalOracle, 2, 1)
	soldier := effectstest.B12Creature(g, me.ID, "Soldier", "Creature — Human Soldier", 2, 2)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	if err := effectstest.MirB1Activate(g, me.ID, marshal, 0, effectstest.CardTarget(soldier)...); err == nil {
		t.Fatal("activated without paying {3}")
	}
	effectstest.DcMana(t, g, me.ID, "{C}{C}{C}")
	if err := effectstest.MirB1Activate(g, me.ID, marshal, 0, effectstest.CardTarget(bear)...); err == nil {
		t.Fatal("a non-Soldier was accepted as the target")
	}
	if err := effectstest.MirB1Activate(g, me.ID, marshal, 0, effectstest.CardTarget(soldier)...); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasKeywordOnBattlefield(t, g, soldier, "flying") {
		t.Error("the Soldier did not gain flying")
	}
	if effectstest.HasKeywordOnBattlefield(t, g, bear, "flying") {
		t.Error("the Bear gained flying")
	}
}

func TestBroB05ConscriptedInfantryLeavesASoldierWhenItDies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	inf := effectstest.B12Push(g, me.ID, "Conscripted Infantry", "Creature — Human Soldier", broB05InfantryOracle, 3, 1)
	effectstest.KillCreature(t, g, me.ID, inf)
	effectstest.PassPriorityAroundTable(t, g)
	if got := broB05Soldiers(g, me.ID); got != 1 {
		t.Errorf("%d Soldier artifact tokens, want 1", got)
	}
}

func TestBroB05EmergencyWeldReturnsACardAndMakesASoldier(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	art := effectstest.GraveyardCardOf(g, me, "Grave Gizmo", "Artifact", "{1}", "")
	land := effectstest.GraveyardCardOf(g, me, "Grave Land", "Land", "", "")
	hand := func() uuid.UUID {
		id := uuid.New()
		me.Hand.PushTop(game.Card{InstanceID: id, Name: "Emergency Weld", TypeLine: "Sorcery", ManaCost: "{1}{B}",
			OracleID: broB05WeldOracle, Owner: me.ID, Controller: me.ID})
		return id
	}
	effectstest.PayMana(t, g, me, "{C}{B}")
	if err := g.CastSpell(me.ID, hand(), game.CastSpellParams{Strict: true, Targets: effectstest.CardTarget(land)}); err == nil {
		t.Fatal("a land card was accepted as the target")
	}
	if err := g.CastSpell(me.ID, hand(), game.CastSpellParams{Strict: true, Targets: effectstest.CardTarget(art)}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(art) {
		t.Error("the artifact card did not return to hand")
	}
	if got := broB05Soldiers(g, me.ID); got != 1 {
		t.Errorf("%d Soldier tokens, want 1", got)
	}
}

func TestBroB05MassProductionMakesFourSoldiers(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	broB05Cast(t, g, me, "Mass Production", "Sorcery", broB05MassOracle, game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if got := broB05Soldiers(g, me.ID); got != 4 {
		t.Errorf("%d Soldier tokens, want 4", got)
	}
}

func TestBroB05MishrasOnslaughtModes(t *testing.T) {
	t.Run("two Soldiers", func(t *testing.T) {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		broB05Cast(t, g, me, "Mishra's Onslaught", "Instant", broB05OnslaughtOracle, game.CastSpellParams{Modes: []int{0}})
		effectstest.PassPriorityAroundTable(t, g)
		if got := broB05Soldiers(g, me.ID); got != 2 {
			t.Errorf("%d Soldier tokens, want 2", got)
		}
	})
	t.Run("pump", func(t *testing.T) {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		effectstest.AdvanceToMain(t, g)
		mine := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
		theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
		effectstest.CastFromHandForTest(t, g, me, "Mishra's Onslaught", "Instant", "{0}", broB05OnslaughtOracle, game.CastSpellParams{Modes: []int{1}})
		effectstest.PassPriorityAroundTable(t, g)
		if p := effectstest.EffectivePower(t, g, mine); p != 4 {
			t.Errorf("my creature has power %d, want 4", p)
		}
		if p := effectstest.EffectivePower(t, g, theirs); p != 2 {
			t.Errorf("their creature has power %d, want 2", p)
		}
		if broB05Soldiers(g, me.ID) != 0 {
			t.Error("the pump mode made tokens")
		}
	})
}

func TestBroB05VeteransPowerbladeEquipsSoldiersCheaply(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	soldier := effectstest.HobB4Creature(g, me.ID, "Guard", "Creature — Human Soldier", 2, 2)
	bear := effectstest.SeedBear(g, me.ID)
	blade := effectstest.SeedEquipment(g, me.ID, "Veteran's Powerblade", broB05PowerbladeOracle)
	if err := effectstest.HobB4Activate(t, g, me.ID, blade, 0, "", effectstest.CardRefs(soldier)); err == nil {
		t.Fatal("Equip Soldier {W} was activated without paying")
	}
	if err := effectstest.HobB4Activate(t, g, me.ID, blade, 0, "{W}", effectstest.CardRefs(bear)); err == nil {
		t.Fatal("Equip Soldier {W} accepted a non-Soldier")
	}
	if err := effectstest.HobB4Activate(t, g, me.ID, blade, 0, "{W}", effectstest.CardRefs(soldier)); err != nil {
		t.Fatalf("Equip Soldier {W}: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if host := effectstest.AttachmentHostOf(t, g, blade); host.ID != soldier {
		t.Fatalf("the Powerblade is on %v, want the Soldier", host)
	}
	if p, tg := effectstest.EffectivePower(t, g, soldier), effectstest.EffectiveToughness(t, g, soldier); p != 4 || tg != 2 {
		t.Errorf("the equipped Soldier is %d/%d, want 4/2", p, tg)
	}
	if err := effectstest.HobB4Activate(t, g, me.ID, blade, 1, "{C}{C}", effectstest.CardRefs(bear)); err != nil {
		t.Fatalf("Equip {2}: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p := effectstest.EffectivePower(t, g, bear); p != 4 {
		t.Errorf("the equipped Bear has power %d, want 4", p)
	}
}

func TestBroB05YotianTacticianPumpsOtherSoldiersYouControl(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	tac := effectstest.B12Push(g, me.ID, "Yotian Tactician", "Creature — Human Soldier", broB05TacticianOracle, 3, 4)
	soldier := effectstest.B12Creature(g, me.ID, "Soldier", "Creature — Human Soldier", 2, 2)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Human Soldier", 2, 2)
	for id, want := range map[uuid.UUID]int{tac: 3, soldier: 3, bear: 2, theirs: 2} {
		if p := effectstest.EffectivePower(t, g, id); p != want {
			t.Errorf("%s has power %d, want %d", effectstest.B12Card(t, g, id).Name, p, want)
		}
	}
}

func TestBroB05ZephyrSentinelCountersOnlyWhenItBouncedASoldier(t *testing.T) {
	for _, tc := range []struct {
		typeLine string
		want     int
	}{{"Creature — Human Soldier", 1}, {"Creature — Bear", 0}} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		effectstest.AdvanceToMain(t, g)
		other := effectstest.B12Creature(g, me.ID, "Other", tc.typeLine, 2, 2)
		sen := effectstest.CastFromHandForTest(t, g, me, "Zephyr Sentinel", "Creature — Human Soldier", "{0}", broB05SentinelOracle, game.CastSpellParams{})
		effectstest.PassPriorityAroundTable(t, g)
		if effectstest.LatestPickTarget(g, me.ID) != nil {
			effectstest.PickCard(t, g, me.ID, other)
		}
		effectstest.PassPriorityAroundTable(t, g)
		if !me.Hand.Contains(other) {
			t.Errorf("%s: the creature was not returned to hand", tc.typeLine)
		}
		if got := effectstest.CounterCount(g, sen, game.CounterPlusOne); got != tc.want {
			t.Errorf("%s: the Sentinel has %d counters, want %d", tc.typeLine, got, tc.want)
		}
	}
}

func TestBroB05ZephyrSentinelWithNoTargetDoesNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	sen := effectstest.CastFromHandForTest(t, g, me, "Zephyr Sentinel", "Creature — Human Soldier", "{0}", broB05SentinelOracle, game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.CounterCount(g, sen, game.CounterPlusOne) != 0 {
		t.Error("the Sentinel got a counter with nothing bounced")
	}
	for _, kw := range []string{"flying", "flash"} {
		if !effectstest.HasEffectiveKeyword(t, g, sen, kw) {
			t.Errorf("the Sentinel lacks %s", kw)
		}
	}
}

func TestBroB05SiegeVeteranCountersAtCombatAndMakesSoldiersWhenSoldiersDie(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	vet := effectstest.B12Push(g, me.ID, "Siege Veteran", "Creature — Human Soldier", broB05VeteranOracle, 2, 2)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	soldier := effectstest.B12Creature(g, me.ID, "Soldier", "Creature — Human Soldier", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Human Soldier", 2, 2)
	token := effectstest.B12Creature(g, me.ID, "Soldier", "Token Creature — Soldier", 1, 1)
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	if p := effectstest.LatestPickTarget(g, me.ID); p != nil {
		if err := g.ResolvePickTarget(p.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: theirs}); err == nil {
			t.Error("an opponent's creature was accepted")
		}
		effectstest.PickCard(t, g, me.ID, bear)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CounterCount(g, bear, game.CounterPlusOne); got != 1 {
		t.Errorf("the Bear has %d counters, want 1", got)
	}
	before := broB05Soldiers(g, me.ID)
	effectstest.KillCreature(t, g, opp.ID, theirs)
	effectstest.KillCreature(t, g, me.ID, bear)
	effectstest.KillCreature(t, g, me.ID, token)
	effectstest.PassPriorityAroundTable(t, g)
	if got := broB05Soldiers(g, me.ID); got != before {
		t.Errorf("an opponent's Soldier, a non-Soldier or a token made a Soldier (%d, was %d)", got, before)
	}
	effectstest.KillCreature(t, g, me.ID, soldier)
	effectstest.PassPriorityAroundTable(t, g)
	if got := broB05Soldiers(g, me.ID); got != before+1 {
		t.Errorf("%d Soldier tokens after a nontoken Soldier died, want %d", got, before+1)
	}
	effectstest.KillCreature(t, g, me.ID, vet)
	effectstest.PassPriorityAroundTable(t, g)
	if got := broB05Soldiers(g, me.ID); got != before+1 {
		t.Errorf("the Veteran's own death made a Soldier (%d)", got)
	}
}

func TestBroB05RescueRetrieverCountersOtherSoldiersAndPreventsTheirDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	soldier := effectstest.B12Creature(g, me.ID, "Soldier", "Creature — Human Soldier", 2, 5)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 5)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Human Soldier", 2, 2)
	ret := effectstest.CastFromHandForTest(t, g, me, "Rescue Retriever", "Creature — Dog Soldier", "{0}", broB05RetrieverOracle, game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	for id, want := range map[uuid.UUID]int{soldier: 1, bear: 0, theirs: 0, ret: 0} {
		if got := effectstest.CounterCount(g, id, game.CounterPlusOne); got != want {
			t.Errorf("%s has %d counters, want %d", effectstest.B12Card(t, g, id).Name, got, want)
		}
	}
	// Prevention: every damage, combat or not, to an attacking Soldier
	// other than the Retriever; nothing to the Retriever, a non-attacker
	// or a non-Soldier.
	blocker := effectstest.B12Creature(g, opp.ID, "Blocker", "Creature — Bear", 2, 2)
	set := func(id uuid.UUID, attacking bool) {
		g.WithWriteLock(func() {
			for i := range g.Battlefield.Cards {
				if g.Battlefield.Cards[i].InstanceID == id {
					if attacking {
						g.Battlefield.Cards[i].AttackingTarget = opp.ID
					} else {
						g.Battlefield.Cards[i].AttackingTarget = uuid.Nil
					}
				}
			}
		})
	}
	damage := func(id uuid.UUID) int {
		c, _ := effectstest.BattlefieldCard(g, id)
		return c.DamageMarked
	}
	set(soldier, true)
	set(bear, true)
	set(ret, true)
	g.WithWriteLock(func() {
		_ = g.DealDamageToCreatureForEffect(blocker, soldier, 2)
		_ = g.DealDamageToCreatureForEffect(blocker, bear, 2)
		_ = g.DealDamageToCreatureForEffect(blocker, ret, 2)
	})
	if got := damage(soldier); got != 0 {
		t.Errorf("an attacking Soldier took %d noncombat damage, want 0", got)
	}
	if got := damage(bear); got != 2 {
		t.Errorf("an attacking non-Soldier took %d, want 2", got)
	}
	if got := damage(ret); got != 2 {
		t.Errorf("the Retriever itself took %d, want 2", got)
	}
	set(soldier, false)
	g.WithWriteLock(func() { _ = g.DealDamageToCreatureForEffect(blocker, soldier, 1) })
	if got := damage(soldier); got != 1 {
		t.Errorf("a non-attacking Soldier took %d, want 1", got)
	}
	set(soldier, true)
	if err := g.MarkCombatDamage(blocker, soldier, 1); err != nil {
		t.Fatal(err)
	}
	if got := damage(soldier); got != 1 {
		t.Errorf("an attacking Soldier took combat damage (now %d, want 1)", got)
	}
}

func TestBroB05FortifiedBeachheadEntersAndPumps(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	land := effectstest.B12Push(g, me.ID, "Fortified Beachhead", "Land", broB05BeachheadOracle, 0, 0)
	soldier := effectstest.B12Creature(g, me.ID, "Soldier", "Creature — Human Soldier", 2, 2)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.DcMana(t, g, me.ID, "{C}{C}{C}{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, land, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p := effectstest.EffectivePower(t, g, soldier); p != 3 {
		t.Errorf("the Soldier has power %d, want 3", p)
	}
	if p := effectstest.EffectivePower(t, g, bear); p != 2 {
		t.Errorf("the Bear has power %d, want 2", p)
	}
	if !effectstest.FindCardAnywhere(t, g, land).Tapped {
		t.Error("the land did not tap for its cost")
	}
}

func broB05PlayBeachhead(t *testing.T, g *game.Game, me *game.Player) uuid.UUID {
	t.Helper()
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Fortified Beachhead", TypeLine: "Land", OracleID: broB05BeachheadOracle, Owner: me.ID, Controller: me.ID})
	effectstest.AdvanceToMain(t, g)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("play the land: %v", err)
	}
	return id
}

func TestBroB05FortifiedBeachheadEntersTappedUnlessASoldierIsRevealedOrControlled(t *testing.T) {
	// Declined reveal, no Soldier: tapped.
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	soldierCard := effectstest.HandCardFull(me, "Hand Soldier", "Creature — Human Soldier", "{1}{W}", "", nil)
	effectstest.HandCardFull(me, "Hand Bear", "Creature — Bear", "{1}{G}", "", nil)
	land := broB05PlayBeachhead(t, g, me)
	effectstest.AnswerEntryReveal(t, g, me.ID)
	if !effectstest.B13Tapped(t, g, land) {
		t.Error("a declined reveal with no Soldier should enter tapped")
	}
	if !me.Hand.Contains(soldierCard) {
		t.Error("the Soldier card left the hand")
	}

	// Reveal a Soldier card: untapped; only the Soldier is offered.
	g = effectstest.NewCatalogGame(t)
	me = g.Seats[g.Turn.ActiveSeat]
	soldierCard = effectstest.HandCardFull(me, "Hand Soldier", "Creature — Human Soldier", "{1}{W}", "", nil)
	bear := effectstest.HandCardFull(me, "Hand Bear", "Creature — Bear", "{1}{G}", "", nil)
	land = broB05PlayBeachhead(t, g, me)
	c := effectstest.EntryRevealChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no reveal prompt")
	}
	for _, id := range c.DiscardOptions {
		if id == bear {
			t.Error("a non-Soldier card was offered")
		}
	}
	effectstest.AnswerEntryReveal(t, g, me.ID, soldierCard)
	if effectstest.B13Tapped(t, g, land) {
		t.Error("a revealed Soldier card should let it enter untapped")
	}
	if err := g.ActivateManaAbility(me.ID, land, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("tap for mana: %v", err)
	}

	// Controlling a Soldier: no question, untapped.
	g = effectstest.NewCatalogGame(t)
	me = g.Seats[g.Turn.ActiveSeat]
	effectstest.B12Creature(g, me.ID, "Board Soldier", "Creature — Human Soldier", 2, 2)
	effectstest.HandCardFull(me, "Hand Soldier", "Creature — Human Soldier", "{1}{W}", "", nil)
	land = broB05PlayBeachhead(t, g, me)
	if effectstest.EntryRevealChoiceFor(g, me.ID) != nil {
		t.Error("asked to reveal while controlling a Soldier")
	}
	if effectstest.B13Tapped(t, g, land) {
		t.Error("controlling a Soldier should let it enter untapped")
	}
}
