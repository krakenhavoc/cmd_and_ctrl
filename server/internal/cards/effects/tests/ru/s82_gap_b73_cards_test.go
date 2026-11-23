package testsru

import (
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
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
		spec, ok := effects.Lookup(oracle)
		if !ok || spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
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
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	spell := effectstest.CastCatalogSpell(t, g, "A Bear", "Creature — Bear", "", nil)
	mystic := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: mystic, Name: "Frilled Mystic", TypeLine: "Creature — Elf Lizard Wizard",
		OracleID: gapFrilledMysticOracle, Power: 3, Toughness: 2, Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, mystic, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("cast Frilled Mystic with the Bear on the stack: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerTriggerTargets(t, g, me.ID, spell)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, answer)
	effectstest.PassPriorityAroundTable(t, g)
	return g, spell, mystic
}

func TestGapB73FrilledMysticHasFlashAndMayCounterATargetSpell(t *testing.T) {
	spec, _ := effects.Lookup(gapFrilledMysticOracle)
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
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	crab := effectstest.B12Push(g, me.ID, "Sharktocrab", "Creature — Shark Octopus Crab", gapSharktocrabOracle, 4, 4)
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)

	effectstest.AddPool(t, g, me.ID, "{C}{C}{G}{U}")
	if err := g.ActivateCatalogAbility(me.ID, crab, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("adapt: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerTriggerTargets(t, g, me.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountersOn(g, crab, game.CounterPlusOne); n != 1 {
		t.Fatalf("%d counters after adapt 1, want 1", n)
	}
	if !effectstest.TappedNow(g, bear) {
		t.Fatal("the target was not tapped")
	}
	// A second adapt does nothing: it already has a counter (CR 701.46a).
	effectstest.AddPool(t, g, me.ID, "{C}{C}{G}{U}")
	if err := g.ActivateCatalogAbility(me.ID, crab, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("second adapt: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountersOn(g, crab, game.CounterPlusOne); n != 1 {
		t.Errorf("%d counters after a second adapt, want 1", n)
	}
	if effectstest.LatestPickTarget(g, me.ID) != nil {
		t.Error("a second adapt that put no counters triggered the Sharktocrab")
	}
	for i := 0; i < len(g.Seats); i++ {
		effectstest.EndTurn(t, g)
	}
	effectstest.AssertTurnOf(t, g, 0, false, "my next turn")
	if !effectstest.TappedNow(g, bear) {
		t.Error("its controller's next untap step untapped it")
	}
	for i := 0; i < len(g.Seats); i++ {
		effectstest.EndTurn(t, g)
	}
	if effectstest.TappedNow(g, bear) {
		t.Error("it stayed tapped through a second untap step")
	}
}

func TestGapB73SharktocrabTriggersOnCountersFromAnySource(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	crab := effectstest.B12Push(g, me.ID, "Sharktocrab", "Creature — Shark Octopus Crab", gapSharktocrabOracle, 4, 4)
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	mine := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(mine, game.CounterPlusOne, 1) })
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestPickTarget(g, me.ID) != nil || effectstest.TappedNow(g, bear) {
		t.Fatal("a counter on another creature triggered the Sharktocrab")
	}
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(crab, game.CounterPlusOne, 2) })
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerTriggerTargets(t, g, me.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.TappedNow(g, bear) {
		t.Error("counters put on the Sharktocrab by another effect did not tap the target")
	}
	// Its own creatures are not legal targets.
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(crab, game.CounterPlusOne, 1) })
	effectstest.PassPriorityAroundTable(t, g)
	if p := effectstest.LatestPickTarget(g, me.ID); p != nil && slices.Contains(p.PickTargetCards, mine) {
		t.Error("a creature the controller controls was offered as the target")
	}
}

// --- Simic Guildmage ----------------------------------------------------------

func TestGapB73SimicGuildmageMovesACounterBetweenCreaturesOfOneController(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	gm := effectstest.B12Push(g, me.ID, "Simic Guildmage", "Creature — Elf Wizard", gapSimicGuildmageOracle, 2, 2)
	from := effectstest.B12Creature(g, opp.ID, "From", "Creature — Bear", 2, 2)
	to := effectstest.B12Creature(g, opp.ID, "To", "Creature — Bear", 2, 2)
	mine := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(from, game.CounterPlusOne, 1) })

	two := func(a, b uuid.UUID) []game.TargetRef {
		return []game.TargetRef{{Kind: game.TargetCard, ID: a, Slot: 0}, {Kind: game.TargetCard, ID: b, Slot: 1}}
	}
	effectstest.AddPool(t, g, me.ID, "{C}{G}")
	if err := g.ActivateCatalogAbility(me.ID, gm, 0, game.ActivateAbilityParams{Strict: true, Targets: two(from, mine)}); err == nil {
		t.Fatal("the second creature has another controller")
	}
	if err := g.ActivateCatalogAbility(me.ID, gm, 0, game.ActivateAbilityParams{Strict: true, Targets: two(from, from)}); err == nil {
		t.Fatal("the second target must be another creature")
	}
	effectstest.P7Activate(t, g, me, gm, 0, game.ActivateAbilityParams{Strict: true, Targets: two(from, to)})
	if effectstest.CountersOn(g, from, game.CounterPlusOne) != 0 || effectstest.CountersOn(g, to, game.CounterPlusOne) != 1 {
		t.Errorf("from %d, to %d; want 0 and 1", effectstest.CountersOn(g, from, game.CounterPlusOne), effectstest.CountersOn(g, to, game.CounterPlusOne))
	}
	// With no counter to move, nothing is put on the second creature.
	effectstest.AddPool(t, g, me.ID, "{C}{G}")
	effectstest.P7Activate(t, g, me, gm, 0, game.ActivateAbilityParams{Strict: true, Targets: two(from, to)})
	if effectstest.CountersOn(g, to, game.CounterPlusOne) != 1 {
		t.Errorf("a counter appeared from nowhere: to has %d", effectstest.CountersOn(g, to, game.CounterPlusOne))
	}
}

func TestGapB73SimicGuildmageMovesAnAuraToAnotherPermanentOfTheHostsController(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	gm := effectstest.B12Push(g, me.ID, "Simic Guildmage", "Creature — Elf Wizard", gapSimicGuildmageOracle, 2, 2)
	from := effectstest.ApaPush(g, opp.ID, opp.ID, effectstest.ApaCreature("Bear", "oracle-bear", 2, 2))
	to := effectstest.ApaPush(g, opp.ID, opp.ID, effectstest.ApaCreature("Ox", "oracle-ox", 2, 4))
	mine := effectstest.ApaPush(g, me.ID, me.ID, effectstest.ApaCreature("Mine", "oracle-mine", 2, 2))
	aura := r35KitsunePacifism(t, g, me, from)
	effectstest.AdvanceToMain(t, g)

	effectstest.ApaMana(me, "C", "U")
	if err := g.ActivateCatalogAbility(me.ID, gm, 1, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardRefs(from)}); err == nil {
		t.Fatal("a creature is not an Aura")
	}
	if err := g.ActivateCatalogAbility(me.ID, gm, 1, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardRefs(aura)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	choice := effectstest.ChooseCardsChoiceFor(g, me.ID)
	if choice == nil {
		t.Fatal("no choice was offered")
	}
	if slices.Contains(choice.ChooseCards, mine) || slices.Contains(choice.ChooseCards, from) || !slices.Contains(choice.ChooseCards, to) {
		t.Errorf("offered %v; want the host's controller's other permanents only", choice.ChooseCards)
	}
	effectstest.AnswerChooseCards(t, g, me.ID, to)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.AttachedToNow(t, g, aura, to) {
		t.Error("the Aura did not move")
	}
}

// --- Selesnya Evangel -----------------------------------------------------------

func TestGapB73SelesnyaEvangelTapsAnotherCreatureToMakeASaproling(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	ev := effectstest.B12Push(g, me.ID, "Selesnya Evangel", "Creature — Elf Shaman", gapSelesnyaEvangelOracle, 1, 2)
	other := effectstest.B12Creature(g, me.ID, "Other", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)

	effectstest.AddPool(t, g, me.ID, "{C}")
	if err := g.ActivateCatalogAbility(me.ID, ev, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("activated without tapping a creature")
	}
	if err := g.ActivateCatalogAbility(me.ID, ev, 0, game.ActivateAbilityParams{Strict: true, TapIDs: []uuid.UUID{ev}}); err == nil {
		t.Fatal("the Evangel tapped itself for both costs")
	}
	effectstest.P7Activate(t, g, me, ev, 0, game.ActivateAbilityParams{Strict: true, TapIDs: []uuid.UUID{other}})
	if n, _ := effectstest.CountTokensNamed(g, me.ID, "Saproling"); n != 1 {
		t.Errorf("%d Saprolings, want 1", n)
	}
	if !effectstest.TappedNow(g, ev) || !effectstest.TappedNow(g, other) {
		t.Error("the Evangel and the other creature are both tapped to pay")
	}
	effectstest.AddPool(t, g, me.ID, "{C}")
	if err := g.ActivateCatalogAbility(me.ID, ev, 0, game.ActivateAbilityParams{Strict: true, TapIDs: []uuid.UUID{other}}); err == nil {
		t.Error("activated with a tapped Evangel")
	}
}

// --- Conclave Cavalier -------------------------------------------------------------

func TestGapB73ConclaveCavalierLeavesTwoVigilantElfKnightsWhenItDies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	cav := effectstest.B12Push(g, me.ID, "Conclave Cavalier", "Creature — Centaur Knight", gapConclaveCavalierOracle, 4, 4)
	effectstest.AdvanceToMain(t, g)
	if !effectstest.HasEffectiveKeyword(t, g, cav, "vigilance") {
		t.Error("no vigilance")
	}
	effectstest.KillCreature(t, g, me.ID, cav)
	effectstest.PassPriorityAroundTable(t, g)
	knights := gapOnBattlefieldNamed(g, me.ID, "Elf Knight")
	if len(knights) != 2 {
		t.Fatalf("%d Elf Knights, want 2", len(knights))
	}
	for _, id := range knights {
		if !effectstest.HasEffectiveKeyword(t, g, id, "vigilance") {
			t.Error("an Elf Knight has no vigilance")
		}
		if p, tough := effectstest.EffectivePTOf(t, g, id); p != 2 || tough != 2 {
			t.Errorf("Elf Knight is %d/%d, want 2/2", p, tough)
		}
	}
}

func TestGapB73ConclaveCavalierMakesNoTokensWhenExiled(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	cav := effectstest.B12Push(g, me.ID, "Conclave Cavalier", "Creature — Centaur Knight", gapConclaveCavalierOracle, 4, 4)
	effectstest.AdvanceToMain(t, g)
	effectstest.ExileCreature(t, g, cav)
	effectstest.PassPriorityAroundTable(t, g)
	if n := len(gapOnBattlefieldNamed(g, me.ID, "Elf Knight")); n != 0 {
		t.Errorf("%d Elf Knights after an exile, want 0", n)
	}
}

// --- Phytohydra ------------------------------------------------------------------

func TestGapB73PhytohydraTurnsDamageIntoCounters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	hydra := effectstest.B12Push(g, me.ID, "Phytohydra", "Creature — Plant Hydra", gapPhytohydraOracle, 1, 1)
	src := effectstest.B12Creature(g, g.Seats[1].ID, "Pinger", "Creature — Bear", 1, 1)
	effectstest.AdvanceToMain(t, g)
	g.WithWriteLock(func() { _ = g.DealDamageToCreatureForEffect(src, hydra, 3) })
	effectstest.CheckState(t, g)
	if n := effectstest.CountersOn(g, hydra, game.CounterPlusOne); n != 3 {
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
	effectstest.CheckState(t, g)
	if !g.Battlefield.Contains(hydra) || effectstest.CountersOn(g, hydra, game.CounterPlusOne) != 103 {
		t.Errorf("100 more damage should be 100 more counters (counters %d)", effectstest.CountersOn(g, hydra, game.CounterPlusOne))
	}
}

func TestGapB73PhytohydraDoesNotChangeDamageToOtherCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B12Push(g, me.ID, "Phytohydra", "Creature — Plant Hydra", gapPhytohydraOracle, 1, 1)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	src := effectstest.B12Creature(g, g.Seats[1].ID, "Pinger", "Creature — Bear", 1, 1)
	effectstest.AdvanceToMain(t, g)
	g.WithWriteLock(func() { _ = g.DealDamageToCreatureForEffect(src, bear, 2) })
	effectstest.CheckState(t, g)
	if g.Battlefield.Contains(bear) {
		t.Error("lethal damage to another creature was replaced")
	}
}

// --- Tolsimir Wolfblood -------------------------------------------------------------

func TestGapB73TolsimirWolfbloodPumpsOtherGreenAndWhiteCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	tol := effectstest.B12Push(g, me.ID, "Tolsimir Wolfblood", "Legendary Creature — Elf Warrior", gapTolsimirWolfbloodOracl, 3, 4)
	green := effectstest.ApaPush(g, me.ID, me.ID, game.Card{Name: "Green", TypeLine: "Creature — Bear", Power: 2, Toughness: 2, Colors: []string{"G"}})
	white := effectstest.ApaPush(g, me.ID, me.ID, game.Card{Name: "White", TypeLine: "Creature — Bear", Power: 2, Toughness: 2, Colors: []string{"W"}})
	both := effectstest.ApaPush(g, me.ID, me.ID, game.Card{Name: "Both", TypeLine: "Creature — Bear", Power: 2, Toughness: 2, Colors: []string{"G", "W"}})
	red := effectstest.ApaPush(g, me.ID, me.ID, game.Card{Name: "Red", TypeLine: "Creature — Bear", Power: 2, Toughness: 2, Colors: []string{"R"}})
	theirs := effectstest.ApaPush(g, opp.ID, opp.ID, game.Card{Name: "Theirs", TypeLine: "Creature — Bear", Power: 2, Toughness: 2, Colors: []string{"G"}})
	effectstest.AdvanceToMain(t, g)
	for id, want := range map[uuid.UUID]int{tol: 3, green: 3, white: 3, both: 4, red: 2, theirs: 2} {
		if p, _ := effectstest.EffectivePTOf(t, g, id); p != want {
			t.Errorf("%s has power %d, want %d", effectstest.FindBattlefieldCardForTest(g, id).Name, p, want)
		}
	}
}

func TestGapB73TolsimirWolfbloodTapsForVoja(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	tol := effectstest.B12Push(g, me.ID, "Tolsimir Wolfblood", "Legendary Creature — Elf Warrior", gapTolsimirWolfbloodOracl, 3, 4)
	effectstest.AdvanceToMain(t, g)
	effectstest.P7Activate(t, g, me, tol, 0, game.ActivateAbilityParams{Strict: true})
	vojas := gapOnBattlefieldNamed(g, me.ID, "Voja")
	if len(vojas) != 1 {
		t.Fatalf("%d Voja tokens, want 1", len(vojas))
	}
	c, _ := g.LookupCardForEffect(vojas[0])
	if !c.IsLegendary() || !effects.IsToken(c) || !c.HasSubtype("Wolf") {
		t.Errorf("Voja is %q, want a legendary Wolf token", c.TypeLine)
	}
	// Green and white, so a Tolsimir anthem makes it 4/4.
	if p, tough := effectstest.EffectivePTOf(t, g, vojas[0]); p != 4 || tough != 4 {
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
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	phoenix := gapPhoenixInGraveyard(g, me)
	for i, tl := range []string{"Instant", "Sorcery", "Instant"} {
		effectstest.CastCatalogSpell(t, g, "Spell "+string(rune('A'+i)), tl, "", nil)
		effectstest.PassPriorityAroundTable(t, g)
	}
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(phoenix) {
		t.Fatal("the Phoenix did not return")
	}
}

func TestGapB73ArclightPhoenixStaysDownWithTwoOrWrongSpells(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	phoenix := gapPhoenixInGraveyard(g, me)
	for i, tl := range []string{"Instant", "Sorcery", "Creature — Bear"} {
		effectstest.CastCatalogSpell(t, g, "Spell "+string(rune('A'+i)), tl, "", nil)
		effectstest.PassPriorityAroundTable(t, g)
	}
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(phoenix) {
		t.Fatal("the Phoenix returned after two instant and sorcery spells and a creature spell")
	}
}

func TestGapB73ArclightPhoenixHasFlyingAndHaste(t *testing.T) {
	spec, _ := effects.Lookup(gapArclightPhoenixOracle)
	if !slices.Contains(spec.PrintedKeywords, "flying") || !slices.Contains(spec.PrintedKeywords, "haste") {
		t.Errorf("keywords %v", spec.PrintedKeywords)
	}
}
