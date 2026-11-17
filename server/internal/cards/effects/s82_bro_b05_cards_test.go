package effects

import (
	"testing"

	"github.com/google/uuid"

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
		spec, ok := Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

// broB05Soldiers counts the Soldier artifact tokens a player controls.
func broB05Soldiers(g *game.Game, controller uuid.UUID) int {
	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.Controller == controller && IsToken(c) && c.Name == "Soldier" && c.IsArtifact() && c.IsCreature() &&
			c.Power == 1 && c.Toughness == 1 && len(c.Colors) == 0 {
			n++
		}
	}
	return n
}

func broB05Cast(t *testing.T, g *game.Game, p *game.Player, name, typeLine, oracle string, params game.CastSpellParams) uuid.UUID {
	t.Helper()
	advanceToMain(t, g)
	return castFromHandForTest(t, g, p, name, typeLine, "{0}", oracle, params)
}

func TestBroB05AeronautCavalryCountersAnotherSoldier(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	soldier := b12Creature(g, me.ID, "My Soldier", "Creature — Human Soldier", 2, 2)
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Soldier", "Creature — Human Soldier", 2, 2)
	cav := castFromHandForTest(t, g, me, "Aeronaut Cavalry", "Creature — Human Soldier", "{0}", broB05CavalryOracle, game.CastSpellParams{})
	passPriorityAroundTable(t, g)
	if p := latestPickTarget(g, me.ID); p != nil {
		for _, bad := range []uuid.UUID{cav, bear, theirs} {
			if err := g.ResolvePickTarget(p.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: bad}); err == nil {
				t.Errorf("an illegal target %s was accepted", bad)
			}
		}
		pickCard(t, g, me.ID, soldier)
	}
	passPriorityAroundTable(t, g)
	if got := counterCount(g, soldier, game.CounterPlusOne); got != 1 {
		t.Errorf("the other Soldier has %d counters, want 1", got)
	}
	if counterCount(g, cav, game.CounterPlusOne) != 0 || counterCount(g, bear, game.CounterPlusOne) != 0 {
		t.Error("a counter landed on the Cavalry or a non-Soldier")
	}
	if !hasEffectiveKeyword(t, g, cav, "flying") {
		t.Error("the Cavalry lacks flying")
	}
}

func TestBroB05AirMarshalGivesASoldierFlying(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	marshal := mirB1Push(g, me.ID, "Air Marshal", "Creature — Human Soldier", broB05MarshalOracle, 2, 1)
	soldier := b12Creature(g, me.ID, "Soldier", "Creature — Human Soldier", 2, 2)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	if err := mirB1Activate(g, me.ID, marshal, 0, cardTarget(soldier)...); err == nil {
		t.Fatal("activated without paying {3}")
	}
	dcMana(t, g, me.ID, "{C}{C}{C}")
	if err := mirB1Activate(g, me.ID, marshal, 0, cardTarget(bear)...); err == nil {
		t.Fatal("a non-Soldier was accepted as the target")
	}
	if err := mirB1Activate(g, me.ID, marshal, 0, cardTarget(soldier)...); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !hasKeywordOnBattlefield(t, g, soldier, "flying") {
		t.Error("the Soldier did not gain flying")
	}
	if hasKeywordOnBattlefield(t, g, bear, "flying") {
		t.Error("the Bear gained flying")
	}
}

func TestBroB05ConscriptedInfantryLeavesASoldierWhenItDies(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	inf := b12Push(g, me.ID, "Conscripted Infantry", "Creature — Human Soldier", broB05InfantryOracle, 3, 1)
	killCreature(t, g, me.ID, inf)
	passPriorityAroundTable(t, g)
	if got := broB05Soldiers(g, me.ID); got != 1 {
		t.Errorf("%d Soldier artifact tokens, want 1", got)
	}
}

func TestBroB05EmergencyWeldReturnsACardAndMakesASoldier(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	art := graveyardCardOf(g, me, "Grave Gizmo", "Artifact", "{1}", "")
	land := graveyardCardOf(g, me, "Grave Land", "Land", "", "")
	hand := func() uuid.UUID {
		id := uuid.New()
		me.Hand.PushTop(game.Card{InstanceID: id, Name: "Emergency Weld", TypeLine: "Sorcery", ManaCost: "{1}{B}",
			OracleID: broB05WeldOracle, Owner: me.ID, Controller: me.ID})
		return id
	}
	payMana(t, g, me, "{C}{B}")
	if err := g.CastSpell(me.ID, hand(), game.CastSpellParams{Strict: true, Targets: cardTarget(land)}); err == nil {
		t.Fatal("a land card was accepted as the target")
	}
	if err := g.CastSpell(me.ID, hand(), game.CastSpellParams{Strict: true, Targets: cardTarget(art)}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(art) {
		t.Error("the artifact card did not return to hand")
	}
	if got := broB05Soldiers(g, me.ID); got != 1 {
		t.Errorf("%d Soldier tokens, want 1", got)
	}
}

func TestBroB05MassProductionMakesFourSoldiers(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	broB05Cast(t, g, me, "Mass Production", "Sorcery", broB05MassOracle, game.CastSpellParams{})
	passPriorityAroundTable(t, g)
	if got := broB05Soldiers(g, me.ID); got != 4 {
		t.Errorf("%d Soldier tokens, want 4", got)
	}
}

func TestBroB05MishrasOnslaughtModes(t *testing.T) {
	t.Run("two Soldiers", func(t *testing.T) {
		g := newCatalogGame(t)
		me := g.Seats[0]
		broB05Cast(t, g, me, "Mishra's Onslaught", "Instant", broB05OnslaughtOracle, game.CastSpellParams{Modes: []int{0}})
		passPriorityAroundTable(t, g)
		if got := broB05Soldiers(g, me.ID); got != 2 {
			t.Errorf("%d Soldier tokens, want 2", got)
		}
	})
	t.Run("pump", func(t *testing.T) {
		g := newCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		advanceToMain(t, g)
		mine := b12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
		theirs := b12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
		castFromHandForTest(t, g, me, "Mishra's Onslaught", "Instant", "{0}", broB05OnslaughtOracle, game.CastSpellParams{Modes: []int{1}})
		passPriorityAroundTable(t, g)
		if p := effectivePower(t, g, mine); p != 4 {
			t.Errorf("my creature has power %d, want 4", p)
		}
		if p := effectivePower(t, g, theirs); p != 2 {
			t.Errorf("their creature has power %d, want 2", p)
		}
		if broB05Soldiers(g, me.ID) != 0 {
			t.Error("the pump mode made tokens")
		}
	})
}

func TestBroB05VeteransPowerbladeEquipsSoldiersCheaply(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	soldier := hobB4Creature(g, me.ID, "Guard", "Creature — Human Soldier", 2, 2)
	bear := seedBear(g, me.ID)
	blade := seedEquipment(g, me.ID, "Veteran's Powerblade", broB05PowerbladeOracle)
	if err := hobB4Activate(t, g, me.ID, blade, 0, "", cardRefs(soldier)); err == nil {
		t.Fatal("Equip Soldier {W} was activated without paying")
	}
	if err := hobB4Activate(t, g, me.ID, blade, 0, "{W}", cardRefs(bear)); err == nil {
		t.Fatal("Equip Soldier {W} accepted a non-Soldier")
	}
	if err := hobB4Activate(t, g, me.ID, blade, 0, "{W}", cardRefs(soldier)); err != nil {
		t.Fatalf("Equip Soldier {W}: %v", err)
	}
	passPriorityAroundTable(t, g)
	if host := attachmentHostOf(t, g, blade); host.ID != soldier {
		t.Fatalf("the Powerblade is on %v, want the Soldier", host)
	}
	if p, tg := effectivePower(t, g, soldier), effectiveToughness(t, g, soldier); p != 4 || tg != 2 {
		t.Errorf("the equipped Soldier is %d/%d, want 4/2", p, tg)
	}
	if err := hobB4Activate(t, g, me.ID, blade, 1, "{C}{C}", cardRefs(bear)); err != nil {
		t.Fatalf("Equip {2}: %v", err)
	}
	passPriorityAroundTable(t, g)
	if p := effectivePower(t, g, bear); p != 4 {
		t.Errorf("the equipped Bear has power %d, want 4", p)
	}
}

func TestBroB05YotianTacticianPumpsOtherSoldiersYouControl(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	tac := b12Push(g, me.ID, "Yotian Tactician", "Creature — Human Soldier", broB05TacticianOracle, 3, 4)
	soldier := b12Creature(g, me.ID, "Soldier", "Creature — Human Soldier", 2, 2)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Theirs", "Creature — Human Soldier", 2, 2)
	for id, want := range map[uuid.UUID]int{tac: 3, soldier: 3, bear: 2, theirs: 2} {
		if p := effectivePower(t, g, id); p != want {
			t.Errorf("%s has power %d, want %d", b12Card(t, g, id).Name, p, want)
		}
	}
}

func TestBroB05ZephyrSentinelCountersOnlyWhenItBouncedASoldier(t *testing.T) {
	for _, tc := range []struct {
		typeLine string
		want     int
	}{{"Creature — Human Soldier", 1}, {"Creature — Bear", 0}} {
		g := newCatalogGame(t)
		me := g.Seats[0]
		advanceToMain(t, g)
		other := b12Creature(g, me.ID, "Other", tc.typeLine, 2, 2)
		sen := castFromHandForTest(t, g, me, "Zephyr Sentinel", "Creature — Human Soldier", "{0}", broB05SentinelOracle, game.CastSpellParams{})
		passPriorityAroundTable(t, g)
		if latestPickTarget(g, me.ID) != nil {
			pickCard(t, g, me.ID, other)
		}
		passPriorityAroundTable(t, g)
		if !me.Hand.Contains(other) {
			t.Errorf("%s: the creature was not returned to hand", tc.typeLine)
		}
		if got := counterCount(g, sen, game.CounterPlusOne); got != tc.want {
			t.Errorf("%s: the Sentinel has %d counters, want %d", tc.typeLine, got, tc.want)
		}
	}
}

func TestBroB05ZephyrSentinelWithNoTargetDoesNothing(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	sen := castFromHandForTest(t, g, me, "Zephyr Sentinel", "Creature — Human Soldier", "{0}", broB05SentinelOracle, game.CastSpellParams{})
	passPriorityAroundTable(t, g)
	passPriorityAroundTable(t, g)
	if counterCount(g, sen, game.CounterPlusOne) != 0 {
		t.Error("the Sentinel got a counter with nothing bounced")
	}
	for _, kw := range []string{"flying", "flash"} {
		if !hasEffectiveKeyword(t, g, sen, kw) {
			t.Errorf("the Sentinel lacks %s", kw)
		}
	}
}

func TestBroB05SiegeVeteranCountersAtCombatAndMakesSoldiersWhenSoldiersDie(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	vet := b12Push(g, me.ID, "Siege Veteran", "Creature — Human Soldier", broB05VeteranOracle, 2, 2)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	soldier := b12Creature(g, me.ID, "Soldier", "Creature — Human Soldier", 2, 2)
	theirs := b12Creature(g, opp.ID, "Theirs", "Creature — Human Soldier", 2, 2)
	token := b12Creature(g, me.ID, "Soldier", "Token Creature — Soldier", 1, 1)
	advanceTo(t, g, game.StepBeginCombat)
	if p := latestPickTarget(g, me.ID); p != nil {
		if err := g.ResolvePickTarget(p.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: theirs}); err == nil {
			t.Error("an opponent's creature was accepted")
		}
		pickCard(t, g, me.ID, bear)
	}
	passPriorityAroundTable(t, g)
	if got := counterCount(g, bear, game.CounterPlusOne); got != 1 {
		t.Errorf("the Bear has %d counters, want 1", got)
	}
	before := broB05Soldiers(g, me.ID)
	killCreature(t, g, opp.ID, theirs)
	killCreature(t, g, me.ID, bear)
	killCreature(t, g, me.ID, token)
	passPriorityAroundTable(t, g)
	if got := broB05Soldiers(g, me.ID); got != before {
		t.Errorf("an opponent's Soldier, a non-Soldier or a token made a Soldier (%d, was %d)", got, before)
	}
	killCreature(t, g, me.ID, soldier)
	passPriorityAroundTable(t, g)
	if got := broB05Soldiers(g, me.ID); got != before+1 {
		t.Errorf("%d Soldier tokens after a nontoken Soldier died, want %d", got, before+1)
	}
	killCreature(t, g, me.ID, vet)
	passPriorityAroundTable(t, g)
	if got := broB05Soldiers(g, me.ID); got != before+1 {
		t.Errorf("the Veteran's own death made a Soldier (%d)", got)
	}
}

func TestBroB05RescueRetrieverCountersOtherSoldiersAndPreventsTheirDamage(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	soldier := b12Creature(g, me.ID, "Soldier", "Creature — Human Soldier", 2, 5)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 5)
	theirs := b12Creature(g, opp.ID, "Theirs", "Creature — Human Soldier", 2, 2)
	ret := castFromHandForTest(t, g, me, "Rescue Retriever", "Creature — Dog Soldier", "{0}", broB05RetrieverOracle, game.CastSpellParams{})
	passPriorityAroundTable(t, g)
	for id, want := range map[uuid.UUID]int{soldier: 1, bear: 0, theirs: 0, ret: 0} {
		if got := counterCount(g, id, game.CounterPlusOne); got != want {
			t.Errorf("%s has %d counters, want %d", b12Card(t, g, id).Name, got, want)
		}
	}
	// Prevention: every damage, combat or not, to an attacking Soldier
	// other than the Retriever; nothing to the Retriever, a non-attacker
	// or a non-Soldier.
	blocker := b12Creature(g, opp.ID, "Blocker", "Creature — Bear", 2, 2)
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
		c, _ := battlefieldCard(g, id)
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
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	land := b12Push(g, me.ID, "Fortified Beachhead", "Land", broB05BeachheadOracle, 0, 0)
	soldier := b12Creature(g, me.ID, "Soldier", "Creature — Human Soldier", 2, 2)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	dcMana(t, g, me.ID, "{C}{C}{C}{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, land, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if p := effectivePower(t, g, soldier); p != 3 {
		t.Errorf("the Soldier has power %d, want 3", p)
	}
	if p := effectivePower(t, g, bear); p != 2 {
		t.Errorf("the Bear has power %d, want 2", p)
	}
	if !findCardAnywhere(t, g, land).Tapped {
		t.Error("the land did not tap for its cost")
	}
}

func broB05PlayBeachhead(t *testing.T, g *game.Game, me *game.Player) uuid.UUID {
	t.Helper()
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Fortified Beachhead", TypeLine: "Land", OracleID: broB05BeachheadOracle, Owner: me.ID, Controller: me.ID})
	advanceToMain(t, g)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("play the land: %v", err)
	}
	return id
}

func TestBroB05FortifiedBeachheadEntersTappedUnlessASoldierIsRevealedOrControlled(t *testing.T) {
	// Declined reveal, no Soldier: tapped.
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	soldierCard := handCardFull(me, "Hand Soldier", "Creature — Human Soldier", "{1}{W}", "", nil)
	handCardFull(me, "Hand Bear", "Creature — Bear", "{1}{G}", "", nil)
	land := broB05PlayBeachhead(t, g, me)
	answerEntryReveal(t, g, me.ID)
	if !b13Tapped(t, g, land) {
		t.Error("a declined reveal with no Soldier should enter tapped")
	}
	if !me.Hand.Contains(soldierCard) {
		t.Error("the Soldier card left the hand")
	}

	// Reveal a Soldier card: untapped; only the Soldier is offered.
	g = newCatalogGame(t)
	me = g.Seats[g.Turn.ActiveSeat]
	soldierCard = handCardFull(me, "Hand Soldier", "Creature — Human Soldier", "{1}{W}", "", nil)
	bear := handCardFull(me, "Hand Bear", "Creature — Bear", "{1}{G}", "", nil)
	land = broB05PlayBeachhead(t, g, me)
	c := entryRevealChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no reveal prompt")
	}
	for _, id := range c.DiscardOptions {
		if id == bear {
			t.Error("a non-Soldier card was offered")
		}
	}
	answerEntryReveal(t, g, me.ID, soldierCard)
	if b13Tapped(t, g, land) {
		t.Error("a revealed Soldier card should let it enter untapped")
	}
	if err := g.ActivateManaAbility(me.ID, land, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("tap for mana: %v", err)
	}

	// Controlling a Soldier: no question, untapped.
	g = newCatalogGame(t)
	me = g.Seats[g.Turn.ActiveSeat]
	b12Creature(g, me.ID, "Board Soldier", "Creature — Human Soldier", 2, 2)
	handCardFull(me, "Hand Soldier", "Creature — Human Soldier", "{1}{W}", "", nil)
	land = broB05PlayBeachhead(t, g, me)
	if entryRevealChoiceFor(g, me.ID) != nil {
		t.Error("asked to reveal while controlling a Soldier")
	}
	if b13Tapped(t, g, land) {
		t.Error("controlling a Soldier should let it enter untapped")
	}
}
