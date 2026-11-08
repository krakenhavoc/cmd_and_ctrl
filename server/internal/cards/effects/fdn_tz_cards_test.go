package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// fdn_tz_cards_test.go — Foundations, T–Z (#4465): the cards from that
// stretch of the alphabet that the S–Z batch left and that existing seams
// express. Each test plays the real catalog entry through the engine and
// asserts what its text says.

const (
	fdnTZTerrorOfMountVelusOracle  = "8161248c-16a3-47db-a3ae-cdfd878d49e0"
	fdnTZTreetopSnarespinnerOracle = "2f57a220-9f64-4143-979a-d6aaf63d49b2"
	fdnTZTwinbladeBlessingOracle   = "b3e1674f-dba1-4570-9990-e5dc643e237d"
	fdnTZUnflinchingCourageOracle  = "0e969a27-1609-4ab4-b0db-46b1af8066a9"
	fdnTZUntamedHungerOracle       = "550abd9b-9aae-4812-9bc0-7f131d6dcde8"
	fdnTZVampireInterloperOracle   = "b753d7dd-4c6b-4472-a009-671572f927eb"
	fdnTZVampireNeonateOracle      = "c0045925-25bf-403c-b721-5f05ba30985b"
	fdnTZVampireSoulcallerOracle   = "bbea8ade-a641-4132-b306-f39fe2f14311"
	fdnTZViashinoPyromancerOracle  = "5e78e6e0-60f0-4c3c-b8dd-bd673f5152b8"
	fdnTZVoraciousGreatsharkOracle = "5d96840e-f9b3-4c86-8f59-f12e5736318f"
	fdnTZWaryThespianOracle        = "448cc829-1596-4aec-b180-c0c1e0566f6b"
	fdnTZWildheartInvokerOracle    = "ff8fb796-af5d-4d64-ab3c-7f2469dfa6bf"
)

func TestFdnTZEveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		fdnTZTerrorOfMountVelusOracle:  "Terror of Mount Velus",
		fdnTZTreetopSnarespinnerOracle: "Treetop Snarespinner",
		fdnTZTwinbladeBlessingOracle:   "Twinblade Blessing",
		fdnTZUnflinchingCourageOracle:  "Unflinching Courage",
		fdnTZUntamedHungerOracle:       "Untamed Hunger",
		fdnTZVampireInterloperOracle:   "Vampire Interloper",
		fdnTZVampireNeonateOracle:      "Vampire Neonate",
		fdnTZVampireSoulcallerOracle:   "Vampire Soulcaller",
		fdnTZViashinoPyromancerOracle:  "Viashino Pyromancer",
		fdnTZVoraciousGreatsharkOracle: "Voracious Greatshark",
		fdnTZWaryThespianOracle:        "Wary Thespian",
		fdnTZWildheartInvokerOracle:    "Wildheart Invoker",
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

func TestFdnTZKeywordBodiesCarryTheirPrintedKeywords(t *testing.T) {
	for _, c := range []struct {
		name, typeLine, oracle string
		power, toughness       int
		keywords               []string
	}{
		{"Terror of Mount Velus", "Creature — Dragon", fdnTZTerrorOfMountVelusOracle, 5, 5, []string{"flying", "double strike"}},
		{"Treetop Snarespinner", "Creature — Spider", fdnTZTreetopSnarespinnerOracle, 1, 4, []string{"reach", "deathtouch"}},
		{"Vampire Interloper", "Creature — Vampire Scout", fdnTZVampireInterloperOracle, 2, 1, []string{"flying"}},
		{"Vampire Soulcaller", "Creature — Vampire Warlock", fdnTZVampireSoulcallerOracle, 3, 2, []string{"flying"}},
	} {
		g := newCatalogGame(t)
		id := b12Push(g, g.Seats[0].ID, c.name, c.typeLine, c.oracle, c.power, c.toughness)
		assertKeywords(t, g, id, c.keywords...)
	}
}

// --- The Auras ----------------------------------------------------------------

func TestFdnTZAurasPumpAndGrantTheirKeywords(t *testing.T) {
	for _, c := range []struct {
		name, oracle     string
		power, toughness int
		keywords         []string
	}{
		{"Untamed Hunger", fdnTZUntamedHungerOracle, 4, 3, []string{"menace"}},
		{"Unflinching Courage", fdnTZUnflinchingCourageOracle, 4, 4, []string{"trample", "lifelink"}},
		{"Twinblade Blessing", fdnTZTwinbladeBlessingOracle, 2, 2, []string{"double strike"}},
	} {
		g := newCatalogGame(t)
		me := g.Seats[0]
		bear := pushVanillaCreature(g, me.ID, "Bear", 2, 2)
		aura := castCatalogSpell(t, g, c.name, auraTypeLine, c.oracle, cardRefs(bear))
		passPriorityAroundTable(t, g)
		if host := attachmentHostOf(t, g, aura); host.Kind != game.TargetCard || host.ID != bear {
			t.Fatalf("%s: attached to %+v, want the Bear", c.name, host)
		}
		if p, tough := effectivePower(t, g, bear), effectiveToughness(t, g, bear); p != c.power || tough != c.toughness {
			t.Errorf("%s: the Bear is %d/%d, want %d/%d", c.name, p, tough, c.power, c.toughness)
		}
		for _, kw := range c.keywords {
			if !hasEffectiveKeyword(t, g, bear, kw) {
				t.Errorf("%s: the Bear lacks %s", c.name, kw)
			}
		}
	}
}

func TestFdnTZTwinbladeBlessingCanBeCastWithASpellOnTheStack(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := pushVanillaCreature(g, me.ID, "Bear", 2, 2)
	advanceToMain(t, g)
	csPushSpell(g, csSpell{card: csSorcery})
	blessing := game.Card{InstanceID: uuid.New(), Name: "Twinblade Blessing", TypeLine: auraTypeLine,
		OracleID: fdnTZTwinbladeBlessingOracle, Owner: me.ID, Controller: me.ID}
	me.Hand.PushTop(blessing)
	if err := g.CastSpell(me.ID, blessing.InstanceID, game.CastSpellParams{Targets: cardRefs(bear)}); err != nil {
		t.Fatalf("a flash Aura could not be cast with a spell on the stack: %v", err)
	}
}

// --- Vampire Interloper and Vampire Soulcaller ------------------------------

func TestFdnTZVampireInterloperAndSoulcallerCantBlock(t *testing.T) {
	for _, c := range []struct{ name, typeLine, oracle string }{
		{"Vampire Interloper", "Creature — Vampire Scout", fdnTZVampireInterloperOracle},
		{"Vampire Soulcaller", "Creature — Vampire Warlock", fdnTZVampireSoulcallerOracle},
	} {
		g := newCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		attacker := pushVanillaCreature(g, me.ID, "Bear", 2, 2)
		vampire := b12Push(g, opp.ID, c.name, c.typeLine, c.oracle, 2, 1)
		advanceTo(t, g, game.StepDeclareAttackers)
		if err := g.DeclareAttacker(attacker, opp.ID); err != nil {
			t.Fatalf("DeclareAttacker: %v", err)
		}
		advanceTo(t, g, game.StepDeclareBlockers)
		if err := g.DeclareBlocker(vampire, attacker); err == nil {
			t.Errorf("%s blocked", c.name)
		}
	}
}

func TestFdnTZVampireSoulcallerReturnsACreatureCardToHand(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := pushGraveyardPermanent(me, "Dead Bear", "Creature — Bear", "{1}{G}")
	rock := pushGraveyardPermanent(me, "Dead Rock", "Artifact", "{1}")
	castCatalogSpell(t, g, "Vampire Soulcaller", "Creature — Vampire Warlock", fdnTZVampireSoulcallerOracle, nil)
	passPriorityAroundTable(t, g)
	b04WaitForPick(t, g, me.ID)
	pick := latestPickTarget(g, me.ID)
	if err := g.ResolvePickTarget(pick.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: rock}); err == nil {
		t.Fatal("a noncreature card was accepted")
	}
	pickTriggerTarget(t, g, me.ID, bear)
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(bear) || me.Graveyard.Contains(bear) {
		t.Error("the creature card did not return to your hand")
	}
	if !me.Graveyard.Contains(rock) {
		t.Error("the artifact card left your graveyard")
	}
}

// --- Vampire Neonate ----------------------------------------------------------

func TestFdnTZVampireNeonateDrainsEachOpponentForOne(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	neonate := b12Push(g, me.ID, "Vampire Neonate", "Creature — Vampire", fdnTZVampireNeonateOracle, 0, 3)
	advanceToMain(t, g)
	before := b27Lives(g)
	b06AddMana(me, "C")
	if err := g.ActivateCatalogAbility(me.ID, neonate, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("activated with one mana; it costs {2}")
	}
	b06AddMana(me, "C")
	b16Activate(t, g, me.ID, neonate, 0, game.ActivateAbilityParams{Strict: true})
	passPriorityAroundTable(t, g)
	for i, l := range b27Lives(g) {
		want := before[i] - 1
		if g.Seats[i].ID == me.ID {
			want = before[i] + 1
		}
		if l != want {
			t.Errorf("seat %d life %d, want %d", i, l, want)
		}
	}
	if !b16Tapped(t, g, neonate) {
		t.Error("the Neonate did not tap")
	}
}

// --- Viashino Pyromancer --------------------------------------------------------

func TestFdnTZViashinoPyromancerHitsTargetPlayerForTwo(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := pushVanillaCreature(g, opp.ID, "Their Bear", 2, 2)
	life := opp.Life
	castCatalogSpell(t, g, "Viashino Pyromancer", "Creature — Lizard Wizard", fdnTZViashinoPyromancerOracle, nil)
	passPriorityAroundTable(t, g)
	b04WaitForPick(t, g, me.ID)
	pick := latestPickTarget(g, me.ID)
	if err := g.ResolvePickTarget(pick.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: bear}); err == nil {
		t.Fatal("a creature was accepted as the target")
	}
	if err := g.ResolvePickTarget(pick.ID, me.ID, game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID}); err != nil {
		t.Fatalf("ResolvePickTarget: %v", err)
	}
	passPriorityAroundTable(t, g)
	if opp.Life != life-2 {
		t.Errorf("the target player is at %d, want %d", opp.Life, life-2)
	}
}

// --- Treetop Snarespinner -------------------------------------------------------

func TestFdnTZTreetopSnarespinnerCountersYourCreatureAtSorcerySpeed(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	spider := b12Push(g, me.ID, "Treetop Snarespinner", "Creature — Spider", fdnTZTreetopSnarespinnerOracle, 1, 4)
	mine := pushVanillaCreature(g, me.ID, "My Bear", 2, 2)
	theirs := pushVanillaCreature(g, opp.ID, "Their Bear", 2, 2)
	advanceToMain(t, g)
	if err := g.ActivateCatalogAbility(me.ID, spider, 0, game.ActivateAbilityParams{Targets: cardRefs(theirs)}); err == nil {
		t.Fatal("an opponent's creature was accepted as the target")
	}
	csPushSpell(g, csSpell{card: csSorcery})
	if err := g.ActivateCatalogAbility(me.ID, spider, 0, game.ActivateAbilityParams{Targets: cardRefs(mine)}); err == nil {
		t.Fatal("activated with a spell on the stack; it is sorcery speed")
	}
	passPriorityAroundTable(t, g)
	b16Activate(t, g, me.ID, spider, 0, game.ActivateAbilityParams{Targets: cardRefs(mine)})
	passPriorityAroundTable(t, g)
	if n := countersOn(g, mine, game.CounterPlusOne); n != 1 {
		t.Errorf("your creature has %d +1/+1 counters, want 1", n)
	}
}

// --- Voracious Greatshark -------------------------------------------------------

func TestFdnTZVoraciousGreatsharkCountersACreatureSpell(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	bear := csPushSpell(g, csSpell{card: csBear, caster: 1, controller: 1})
	shark := game.Card{InstanceID: uuid.New(), Name: "Voracious Greatshark", TypeLine: "Creature — Shark",
		OracleID: fdnTZVoraciousGreatsharkOracle, Power: 5, Toughness: 4, Owner: me.ID, Controller: me.ID}
	me.Hand.PushTop(shark)
	if err := g.CastSpell(me.ID, shark.InstanceID, game.CastSpellParams{}); err != nil {
		t.Fatalf("the Greatshark has flash, but could not be cast with a spell on the stack: %v", err)
	}
	passPriorityAroundTable(t, g)
	if pick := latestPickTarget(g, me.ID); pick != nil {
		pickTriggerTarget(t, g, me.ID, bear)
	}
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(shark.InstanceID) {
		t.Error("the Greatshark is not on the battlefield")
	}
	if g.Battlefield.Contains(bear) || !opp.Graveyard.Contains(bear) {
		t.Error("the creature spell was not countered")
	}
}

func TestFdnTZVoraciousGreatsharkCantTargetAnInstantSpell(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	instant := csPushSpell(g, csSpell{card: csBlueInstant, caster: 1, controller: 1})
	shark := game.Card{InstanceID: uuid.New(), Name: "Voracious Greatshark", TypeLine: "Creature — Shark",
		OracleID: fdnTZVoraciousGreatsharkOracle, Power: 5, Toughness: 4, Owner: me.ID, Controller: me.ID}
	me.Hand.PushTop(shark)
	if err := g.CastSpell(me.ID, shark.InstanceID, game.CastSpellParams{}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	passPriorityAroundTable(t, g)
	if pick := latestPickTarget(g, me.ID); pick != nil {
		if err := g.ResolvePickTarget(pick.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: instant}); err == nil {
			t.Fatal("an instant spell was accepted as the target")
		}
	}
}

// --- Wary Thespian --------------------------------------------------------------

func TestFdnTZWaryThespianSurveilsAsItEntersAndAsItDies(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	seedLibrary(me, "A", "B", "C")
	castCatalogSpell(t, g, "Wary Thespian", "Creature — Cat Druid", fdnTZWaryThespianOracle, nil)
	passPriorityAroundTable(t, g)
	eoeQZKeepOnTop(t, g, me.ID, eoeQZSurveil(t, g, me.ID, 1))
	thespian := avB1Named(g, me.ID, "Wary Thespian")
	if thespian == uuid.Nil {
		t.Fatal("Wary Thespian is not on the battlefield")
	}
	b15Destroy(t, g, thespian)
	eoeQZSurveil(t, g, me.ID, 1)
}

// --- Wildheart Invoker ----------------------------------------------------------

func TestFdnTZWildheartInvokerPumpsATargetAndGrantsTrample(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	invoker := b12Push(g, me.ID, "Wildheart Invoker", "Creature — Elf Shaman", fdnTZWildheartInvokerOracle, 4, 3)
	bear := pushVanillaCreature(g, me.ID, "Bear", 2, 2)
	advanceToMain(t, g)
	for range 7 {
		b06AddMana(me, "C")
	}
	if err := g.ActivateCatalogAbility(me.ID, invoker, 0, game.ActivateAbilityParams{Strict: true, Targets: cardRefs(bear)}); err == nil {
		t.Fatal("activated with seven mana; it costs {8}")
	}
	b06AddMana(me, "C")
	b16Activate(t, g, me.ID, invoker, 0, game.ActivateAbilityParams{Strict: true, Targets: cardRefs(bear)})
	passPriorityAroundTable(t, g)
	if p, tough := effectivePower(t, g, bear), effectiveToughness(t, g, bear); p != 7 || tough != 7 {
		t.Errorf("the Bear is %d/%d, want 7/7", p, tough)
	}
	if !hasEffectiveKeyword(t, g, bear, "trample") {
		t.Error("the Bear did not gain trample")
	}
	if b16Tapped(t, g, invoker) {
		t.Error("the Invoker tapped; its cost has no {T}")
	}
}

// --- Terror of Mount Velus ------------------------------------------------------

func TestFdnTZTerrorOfMountVelusGivesYourCreaturesDoubleStrike(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := pushVanillaCreature(g, me.ID, "My Bear", 2, 2)
	theirs := pushVanillaCreature(g, opp.ID, "Their Bear", 2, 2)
	castHoldCreature(t, g, "Terror of Mount Velus", "Creature — Dragon", fdnTZTerrorOfMountVelusOracle, 5, 5)
	if !hasEffectiveKeyword(t, g, mine, "double strike") {
		t.Error("your creature did not gain double strike")
	}
	if hasEffectiveKeyword(t, g, theirs, "double strike") {
		t.Error("an opponent's creature gained double strike")
	}
	late := pushVanillaCreature(g, me.ID, "Late Bear", 2, 2)
	if hasEffectiveKeyword(t, g, late, "double strike") {
		t.Error("a creature that arrived after the trigger resolved gained double strike")
	}
	advanceToUpkeepOf(t, g, 1)
	if hasEffectiveKeyword(t, g, mine, "double strike") {
		t.Error("double strike outlasted the turn")
	}
}
