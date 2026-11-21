package effects

import (
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_gap_b73_cards_test.go — S82 GAP batch 73 (#5948), the Ravnica
// Remastered creatures that were missing. Each test plays the catalog entry
// through the engine and asserts what its printed text says.

const (
	gapFrilledMysticOracle    = "724bc69c-cd09-4ff8-91e7-70500b9474e1"
	gapSharktocrabOracle      = "fb8e43f2-41d5-4ffb-a61e-940477749e00"
	gapSimicGuildmageOracle   = "8b70d296-2756-4bda-b51e-f93be395fce7"
	gapConclaveCavalierOracle = "653de668-beb0-4be7-bb05-30de5f52a3f1"
	gapPhytohydraOracle       = "4f5a7b65-b7f4-4bad-acfd-35edfb8f86a7"
	gapSelesnyaEvangelOracle  = "4a2fea39-99d6-4766-9f2c-0a63925349cb"
	gapTolsimirWolfbloodOracl = "30ee2790-1f5d-4073-aae7-5e7ae30421d2"
	gapArclightPhoenixOracle  = "a80ea391-54d1-4b61-95a3-f8deb4cbe842"
)

func TestGapB73CardsAreFullWithNoCaveats(t *testing.T) {
	for oracle, name := range map[string]string{
		gapFrilledMysticOracle: "Frilled Mystic", gapSharktocrabOracle: "Sharktocrab",
		gapSimicGuildmageOracle: "Simic Guildmage", gapConclaveCavalierOracle: "Conclave Cavalier",
		gapPhytohydraOracle: "Phytohydra", gapSelesnyaEvangelOracle: "Selesnya Evangel",
		gapTolsimirWolfbloodOracl: "Tolsimir Wolfblood", gapArclightPhoenixOracle: "Arclight Phoenix",
	} {
		spec, ok := Lookup(oracle)
		if !ok || spec.Name != name || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: %+v (registered %v)", name, spec, ok)
		}
	}
}

// gapOnBattlefieldNamed lists the ids of `controller`'s battlefield cards
// with this name.
func gapOnBattlefieldNamed(g *game.Game, controller uuid.UUID, name string) []uuid.UUID {
	var out []uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.Controller == controller && c.Name == name {
			out = append(out, c.InstanceID)
		}
	}
	return out
}

// --- Frilled Mystic ---------------------------------------------------------

func gapMysticCounters(t *testing.T, answer bool) (*game.Game, uuid.UUID, uuid.UUID) {
	t.Helper()
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	spell := castCatalogSpell(t, g, "A Bear", "Creature — Bear", "", nil)
	mystic := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: mystic, Name: "Frilled Mystic", TypeLine: "Creature — Elf Lizard Wizard",
		OracleID: gapFrilledMysticOracle, Power: 3, Toughness: 2, Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, mystic, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("cast Frilled Mystic with the Bear on the stack: %v", err)
	}
	passPriorityAroundTable(t, g)
	answerTriggerTargets(t, g, me.ID, spell)
	answerLatestTriggerPrompt(t, g, me.ID, answer)
	passPriorityAroundTable(t, g)
	return g, spell, mystic
}

func TestGapB73FrilledMysticHasFlashAndMayCounterATargetSpell(t *testing.T) {
	spec, _ := Lookup(gapFrilledMysticOracle)
	if !slices.Contains(spec.PrintedKeywords, "flash") {
		t.Error("no flash")
	}
	g, spell, mystic := gapMysticCounters(t, true)
	if g.Battlefield.Contains(spell) {
		t.Error("the countered spell resolved")
	}
	if !g.Battlefield.Contains(mystic) {
		t.Error("the Mystic is not on the battlefield")
	}
}

func TestGapB73FrilledMysticMayDeclineToCounter(t *testing.T) {
	g, spell, _ := gapMysticCounters(t, false)
	if !g.Battlefield.Contains(spell) {
		t.Error("the spell was countered although the controller said no")
	}
}

// --- Sharktocrab --------------------------------------------------------------

func TestGapB73SharktocrabAdaptsOnceAndFreezesATargetForOneUntapStep(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	crab := b12Push(g, me.ID, "Sharktocrab", "Creature — Shark Octopus Crab", gapSharktocrabOracle, 4, 4)
	bear := b12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	advanceToMain(t, g)

	addPool(t, g, me.ID, "{C}{C}{G}{U}")
	if err := g.ActivateCatalogAbility(me.ID, crab, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("adapt: %v", err)
	}
	passPriorityAroundTable(t, g)
	answerTriggerTargets(t, g, me.ID, bear)
	passPriorityAroundTable(t, g)
	if n := countersOn(g, crab, game.CounterPlusOne); n != 1 {
		t.Fatalf("%d counters after adapt 1, want 1", n)
	}
	if !tappedNow(g, bear) {
		t.Fatal("the target was not tapped")
	}
	// A second adapt does nothing: it already has a counter (CR 701.46a).
	addPool(t, g, me.ID, "{C}{C}{G}{U}")
	if err := g.ActivateCatalogAbility(me.ID, crab, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("second adapt: %v", err)
	}
	passPriorityAroundTable(t, g)
	if n := countersOn(g, crab, game.CounterPlusOne); n != 1 {
		t.Errorf("%d counters after a second adapt, want 1", n)
	}
	if latestPickTarget(g, me.ID) != nil {
		t.Error("a second adapt that put no counters triggered the Sharktocrab")
	}
	for i := 0; i < len(g.Seats); i++ {
		endTurn(t, g)
	}
	assertTurnOf(t, g, 0, false, "my next turn")
	if !tappedNow(g, bear) {
		t.Error("its controller's next untap step untapped it")
	}
	for i := 0; i < len(g.Seats); i++ {
		endTurn(t, g)
	}
	if tappedNow(g, bear) {
		t.Error("it stayed tapped through a second untap step")
	}
}

func TestGapB73SharktocrabTriggersOnCountersFromAnySource(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	crab := b12Push(g, me.ID, "Sharktocrab", "Creature — Shark Octopus Crab", gapSharktocrabOracle, 4, 4)
	bear := b12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	mine := b12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	advanceToMain(t, g)
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(mine, game.CounterPlusOne, 1) })
	passPriorityAroundTable(t, g)
	if latestPickTarget(g, me.ID) != nil || tappedNow(g, bear) {
		t.Fatal("a counter on another creature triggered the Sharktocrab")
	}
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(crab, game.CounterPlusOne, 2) })
	passPriorityAroundTable(t, g)
	answerTriggerTargets(t, g, me.ID, bear)
	passPriorityAroundTable(t, g)
	if !tappedNow(g, bear) {
		t.Error("counters put on the Sharktocrab by another effect did not tap the target")
	}
	// Its own creatures are not legal targets.
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(crab, game.CounterPlusOne, 1) })
	passPriorityAroundTable(t, g)
	if p := latestPickTarget(g, me.ID); p != nil && slices.Contains(p.PickTargetCards, mine) {
		t.Error("a creature the controller controls was offered as the target")
	}
}

// --- Simic Guildmage ----------------------------------------------------------

func TestGapB73SimicGuildmageMovesACounterBetweenCreaturesOfOneController(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	gm := b12Push(g, me.ID, "Simic Guildmage", "Creature — Elf Wizard", gapSimicGuildmageOracle, 2, 2)
	from := b12Creature(g, opp.ID, "From", "Creature — Bear", 2, 2)
	to := b12Creature(g, opp.ID, "To", "Creature — Bear", 2, 2)
	mine := b12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	advanceToMain(t, g)
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(from, game.CounterPlusOne, 1) })

	two := func(a, b uuid.UUID) []game.TargetRef {
		return []game.TargetRef{{Kind: game.TargetCard, ID: a, Slot: 0}, {Kind: game.TargetCard, ID: b, Slot: 1}}
	}
	addPool(t, g, me.ID, "{C}{G}")
	if err := g.ActivateCatalogAbility(me.ID, gm, 0, game.ActivateAbilityParams{Strict: true, Targets: two(from, mine)}); err == nil {
		t.Fatal("the second creature has another controller")
	}
	if err := g.ActivateCatalogAbility(me.ID, gm, 0, game.ActivateAbilityParams{Strict: true, Targets: two(from, from)}); err == nil {
		t.Fatal("the second target must be another creature")
	}
	p7Activate(t, g, me, gm, 0, game.ActivateAbilityParams{Strict: true, Targets: two(from, to)})
	if countersOn(g, from, game.CounterPlusOne) != 0 || countersOn(g, to, game.CounterPlusOne) != 1 {
		t.Errorf("from %d, to %d; want 0 and 1", countersOn(g, from, game.CounterPlusOne), countersOn(g, to, game.CounterPlusOne))
	}
	// With no counter to move, nothing is put on the second creature.
	addPool(t, g, me.ID, "{C}{G}")
	p7Activate(t, g, me, gm, 0, game.ActivateAbilityParams{Strict: true, Targets: two(from, to)})
	if countersOn(g, to, game.CounterPlusOne) != 1 {
		t.Errorf("a counter appeared from nowhere: to has %d", countersOn(g, to, game.CounterPlusOne))
	}
}

func TestGapB73SimicGuildmageMovesAnAuraToAnotherPermanentOfTheHostsController(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	gm := b12Push(g, me.ID, "Simic Guildmage", "Creature — Elf Wizard", gapSimicGuildmageOracle, 2, 2)
	from := apaPush(g, opp.ID, opp.ID, apaCreature("Bear", "oracle-bear", 2, 2))
	to := apaPush(g, opp.ID, opp.ID, apaCreature("Ox", "oracle-ox", 2, 4))
	mine := apaPush(g, me.ID, me.ID, apaCreature("Mine", "oracle-mine", 2, 2))
	aura := kitsunePacifism(t, g, me, from)
	advanceToMain(t, g)

	apaMana(me, "C", "U")
	if err := g.ActivateCatalogAbility(me.ID, gm, 1, game.ActivateAbilityParams{Strict: true, Targets: cardRefs(from)}); err == nil {
		t.Fatal("a creature is not an Aura")
	}
	if err := g.ActivateCatalogAbility(me.ID, gm, 1, game.ActivateAbilityParams{Strict: true, Targets: cardRefs(aura)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	choice := chooseCardsChoiceFor(g, me.ID)
	if choice == nil {
		t.Fatal("no choice was offered")
	}
	if slices.Contains(choice.ChooseCards, mine) || slices.Contains(choice.ChooseCards, from) || !slices.Contains(choice.ChooseCards, to) {
		t.Errorf("offered %v; want the host's controller's other permanents only", choice.ChooseCards)
	}
	answerChooseCards(t, g, me.ID, to)
	passPriorityAroundTable(t, g)
	if !attachedToNow(t, g, aura, to) {
		t.Error("the Aura did not move")
	}
}

// --- Selesnya Evangel -----------------------------------------------------------

func TestGapB73SelesnyaEvangelTapsAnotherCreatureToMakeASaproling(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	ev := b12Push(g, me.ID, "Selesnya Evangel", "Creature — Elf Shaman", gapSelesnyaEvangelOracle, 1, 2)
	other := b12Creature(g, me.ID, "Other", "Creature — Bear", 2, 2)
	advanceToMain(t, g)

	addPool(t, g, me.ID, "{C}")
	if err := g.ActivateCatalogAbility(me.ID, ev, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("activated without tapping a creature")
	}
	if err := g.ActivateCatalogAbility(me.ID, ev, 0, game.ActivateAbilityParams{Strict: true, TapIDs: []uuid.UUID{ev}}); err == nil {
		t.Fatal("the Evangel tapped itself for both costs")
	}
	p7Activate(t, g, me, ev, 0, game.ActivateAbilityParams{Strict: true, TapIDs: []uuid.UUID{other}})
	if n, _ := countTokensNamed(g, me.ID, "Saproling"); n != 1 {
		t.Errorf("%d Saprolings, want 1", n)
	}
	if !tappedNow(g, ev) || !tappedNow(g, other) {
		t.Error("the Evangel and the other creature are both tapped to pay")
	}
	addPool(t, g, me.ID, "{C}")
	if err := g.ActivateCatalogAbility(me.ID, ev, 0, game.ActivateAbilityParams{Strict: true, TapIDs: []uuid.UUID{other}}); err == nil {
		t.Error("activated with a tapped Evangel")
	}
}

// --- Conclave Cavalier -------------------------------------------------------------

func TestGapB73ConclaveCavalierLeavesTwoVigilantElfKnightsWhenItDies(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	cav := b12Push(g, me.ID, "Conclave Cavalier", "Creature — Centaur Knight", gapConclaveCavalierOracle, 4, 4)
	advanceToMain(t, g)
	if !hasEffectiveKeyword(t, g, cav, "vigilance") {
		t.Error("no vigilance")
	}
	killCreature(t, g, me.ID, cav)
	passPriorityAroundTable(t, g)
	knights := gapOnBattlefieldNamed(g, me.ID, "Elf Knight")
	if len(knights) != 2 {
		t.Fatalf("%d Elf Knights, want 2", len(knights))
	}
	for _, id := range knights {
		if !hasEffectiveKeyword(t, g, id, "vigilance") {
			t.Error("an Elf Knight has no vigilance")
		}
		if p, tough := effectivePTOf(t, g, id); p != 2 || tough != 2 {
			t.Errorf("Elf Knight is %d/%d, want 2/2", p, tough)
		}
	}
}

func TestGapB73ConclaveCavalierMakesNoTokensWhenExiled(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	cav := b12Push(g, me.ID, "Conclave Cavalier", "Creature — Centaur Knight", gapConclaveCavalierOracle, 4, 4)
	advanceToMain(t, g)
	exileCreature(t, g, cav)
	passPriorityAroundTable(t, g)
	if n := len(gapOnBattlefieldNamed(g, me.ID, "Elf Knight")); n != 0 {
		t.Errorf("%d Elf Knights after an exile, want 0", n)
	}
}

// --- Phytohydra ------------------------------------------------------------------

func TestGapB73PhytohydraTurnsDamageIntoCounters(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	hydra := b12Push(g, me.ID, "Phytohydra", "Creature — Plant Hydra", gapPhytohydraOracle, 1, 1)
	src := b12Creature(g, g.Seats[1].ID, "Pinger", "Creature — Bear", 1, 1)
	advanceToMain(t, g)
	g.WithWriteLock(func() { _ = g.DealDamageToCreatureForEffect(src, hydra, 3) })
	checkState(t, g)
	if n := countersOn(g, hydra, game.CounterPlusOne); n != 3 {
		t.Fatalf("%d counters after 3 damage, want 3", n)
	}
	c, ok := g.LookupCardForEffect(hydra)
	if !ok || c.DamageMarked != 0 {
		t.Errorf("the Hydra has %d damage marked (found %v), want none", c.DamageMarked, ok)
	}
	if p, tough := c.CurrentPower(), c.CurrentToughness(); p != 4 || tough != 4 {
		t.Errorf("Phytohydra is %d/%d, want 4/4", p, tough)
	}
	g.WithWriteLock(func() { _ = g.DealDamageToCreatureForEffect(src, hydra, 100) })
	checkState(t, g)
	if !g.Battlefield.Contains(hydra) || countersOn(g, hydra, game.CounterPlusOne) != 103 {
		t.Errorf("100 more damage should be 100 more counters (counters %d)", countersOn(g, hydra, game.CounterPlusOne))
	}
}

func TestGapB73PhytohydraDoesNotChangeDamageToOtherCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b12Push(g, me.ID, "Phytohydra", "Creature — Plant Hydra", gapPhytohydraOracle, 1, 1)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	src := b12Creature(g, g.Seats[1].ID, "Pinger", "Creature — Bear", 1, 1)
	advanceToMain(t, g)
	g.WithWriteLock(func() { _ = g.DealDamageToCreatureForEffect(src, bear, 2) })
	checkState(t, g)
	if g.Battlefield.Contains(bear) {
		t.Error("lethal damage to another creature was replaced")
	}
}

// --- Tolsimir Wolfblood -------------------------------------------------------------

func TestGapB73TolsimirWolfbloodPumpsOtherGreenAndWhiteCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	tol := b12Push(g, me.ID, "Tolsimir Wolfblood", "Legendary Creature — Elf Warrior", gapTolsimirWolfbloodOracl, 3, 4)
	green := apaPush(g, me.ID, me.ID, game.Card{Name: "Green", TypeLine: "Creature — Bear", Power: 2, Toughness: 2, Colors: []string{"G"}})
	white := apaPush(g, me.ID, me.ID, game.Card{Name: "White", TypeLine: "Creature — Bear", Power: 2, Toughness: 2, Colors: []string{"W"}})
	both := apaPush(g, me.ID, me.ID, game.Card{Name: "Both", TypeLine: "Creature — Bear", Power: 2, Toughness: 2, Colors: []string{"G", "W"}})
	red := apaPush(g, me.ID, me.ID, game.Card{Name: "Red", TypeLine: "Creature — Bear", Power: 2, Toughness: 2, Colors: []string{"R"}})
	theirs := apaPush(g, opp.ID, opp.ID, game.Card{Name: "Theirs", TypeLine: "Creature — Bear", Power: 2, Toughness: 2, Colors: []string{"G"}})
	advanceToMain(t, g)
	for id, want := range map[uuid.UUID]int{tol: 3, green: 3, white: 3, both: 4, red: 2, theirs: 2} {
		if p, _ := effectivePTOf(t, g, id); p != want {
			t.Errorf("%s has power %d, want %d", findBattlefieldCardForTest(g, id).Name, p, want)
		}
	}
}

func TestGapB73TolsimirWolfbloodTapsForVoja(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	tol := b12Push(g, me.ID, "Tolsimir Wolfblood", "Legendary Creature — Elf Warrior", gapTolsimirWolfbloodOracl, 3, 4)
	advanceToMain(t, g)
	p7Activate(t, g, me, tol, 0, game.ActivateAbilityParams{Strict: true})
	vojas := gapOnBattlefieldNamed(g, me.ID, "Voja")
	if len(vojas) != 1 {
		t.Fatalf("%d Voja tokens, want 1", len(vojas))
	}
	c, _ := g.LookupCardForEffect(vojas[0])
	if !c.IsLegendary() || !IsToken(c) || !c.HasSubtype("Wolf") {
		t.Errorf("Voja is %q, want a legendary Wolf token", c.TypeLine)
	}
	// Green and white, so a Tolsimir anthem makes it 4/4.
	if p, tough := effectivePTOf(t, g, vojas[0]); p != 4 || tough != 4 {
		t.Errorf("Voja is %d/%d with two anthems, want 4/4", p, tough)
	}
	if err := g.ActivateCatalogAbility(me.ID, tol, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Error("activated while tapped")
	}
}

// --- Arclight Phoenix -----------------------------------------------------------------

func gapPhoenixInGraveyard(g *game.Game, me *game.Player) uuid.UUID {
	id := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: id, Name: "Arclight Phoenix", TypeLine: "Creature — Phoenix",
		OracleID: gapArclightPhoenixOracle, Power: 3, Toughness: 2, Owner: me.ID, Controller: me.ID})
	return id
}

func TestGapB73ArclightPhoenixReturnsAfterThreeInstantsAndSorceries(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	phoenix := gapPhoenixInGraveyard(g, me)
	for i, tl := range []string{"Instant", "Sorcery", "Instant"} {
		castCatalogSpell(t, g, "Spell "+string(rune('A'+i)), tl, "", nil)
		passPriorityAroundTable(t, g)
	}
	advanceTo(t, g, game.StepBeginCombat)
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(phoenix) {
		t.Fatal("the Phoenix did not return")
	}
}

func TestGapB73ArclightPhoenixStaysDownWithTwoOrWrongSpells(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	phoenix := gapPhoenixInGraveyard(g, me)
	for i, tl := range []string{"Instant", "Sorcery", "Creature — Bear"} {
		castCatalogSpell(t, g, "Spell "+string(rune('A'+i)), tl, "", nil)
		passPriorityAroundTable(t, g)
	}
	advanceTo(t, g, game.StepBeginCombat)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(phoenix) {
		t.Fatal("the Phoenix returned after two instant and sorcery spells and a creature spell")
	}
}

func TestGapB73ArclightPhoenixHasFlyingAndHaste(t *testing.T) {
	spec, _ := Lookup(gapArclightPhoenixOracle)
	if !slices.Contains(spec.PrintedKeywords, "flying") || !slices.Contains(spec.PrintedKeywords, "haste") {
		t.Errorf("keywords %v", spec.PrintedKeywords)
	}
}
