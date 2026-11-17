package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_dmu_b11_test.go — Dominaria United batch 11 (S82, #5487).

const (
	dmuBriarHydraOracle    = "caaedd67-e074-47c4-a956-da88a01f8996"
	dmuNaelOracle          = "6044696a-0f64-474a-840e-8ed4dbd30221"
	dmuSphinxOracle        = "7ee394a2-84de-42e8-834e-6c1030be0aef"
	dmuRadhaOracle         = "fbf0bde7-cb13-402b-b3ab-81c787ce4827"
	dmuZarOracle           = "1402c792-6e2f-4a85-8a0c-8db3c8ce5f1f"
	dmuThreatsOracle       = "24108160-e72b-4451-b1b1-b2b3ac3ede85"
	dmuSilverbackOracle    = "b1a0ae56-77aa-43ad-9aa8-dc1ab779e9d9"
	dmuRulikOracle         = "741dc9c3-1998-4f4c-9c07-6f34fea2bae4"
	dmuTatyovaOracle       = "7d5e734f-3efc-4124-b1d8-4fe3e9d3db28"
	dmuBeastcallerOracle   = "d777e428-cc50-40b1-a241-41576d2575d1"
	dmuB11CreatureTypeLine = "Creature — Bear"
)

func TestDMUB11CardsAreRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		dmuBriarHydraOracle:  "Briar Hydra",
		dmuNaelOracle:        "Nael, Avizoa Aeronaut",
		dmuSphinxOracle:      "Sphinx of Clear Skies",
		dmuRadhaOracle:       "Radha, Coalition Warlord",
		dmuZarOracle:         "Zar Ojanen, Scion of Efrava",
		dmuThreatsOracle:     "Threats Undetected",
		dmuSilverbackOracle:  "Silverback Elder",
		dmuRulikOracle:       "Rulik Mons, Warren Chief",
		dmuTatyovaOracle:     "Tatyova, Steward of Tides",
		dmuBeastcallerOracle: "Quirion Beastcaller",
	} {
		spec, ok := Lookup(oracle)
		if !ok || spec.Name != name || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: registered=%v name=%q completeness=%v caveats=%v", name, ok, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

// dmuB11Counters reads the +1/+1 counters on a battlefield card.
func dmuB11Counters(t *testing.T, g *game.Game, id uuid.UUID) int {
	t.Helper()
	var n int
	g.ReadSnapshot(func() {
		if c, ok := g.LookupCardForEffect(id); ok {
			n = c.Counters[game.CounterPlusOne]
		}
	})
	return n
}

// dmuB11PickIfAsked answers a trigger-target prompt when one is open.
func dmuB11PickIfAsked(t *testing.T, g *game.Game, chooser, card uuid.UUID) {
	t.Helper()
	for i := 0; i < 8 && latestPickTarget(g, chooser) == nil && len(g.StackMeta) == 0; i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatal(err)
		}
	}
	if latestPickTarget(g, chooser) != nil {
		pickCard(t, g, chooser, card)
	}
}

// dmuB11Attack attacks the first opponent with one creature and resolves
// the attack triggers.
func dmuB11Attack(t *testing.T, g *game.Game, me, opp *game.Player, creature uuid.UUID) {
	t.Helper()
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(creature, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	lockInAttacks(t, g)
}

func TestDMUB11BriarHydraPutsDomainCountersOnATarget(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	hydra := b12Push(g, me.ID, "Briar Hydra", "Creature — Plant Hydra", dmuBriarHydraOracle, 6, 6)
	bear := b12Creature(g, me.ID, "Bear", dmuB11CreatureTypeLine, 2, 2)
	domainLands(g, me, opp, "Forest", "Island") // Forest, Island, Plains
	advanceToMain(t, g)

	dealCombatDamageToPlayer(g, hydra, opp.ID, 6)
	dmuB11PickIfAsked(t, g, me.ID, bear)
	passPriorityAroundTable(t, g)
	if n := dmuB11Counters(t, g, bear); n != 3 {
		t.Errorf("counters on the chosen creature: %d, want 3", n)
	}
	if n := dmuB11Counters(t, g, hydra); n != 0 {
		t.Errorf("counters on the Hydra: %d, want 0", n)
	}
}

func TestDMUB11BriarHydraIgnoresNoncombatDamage(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	hydra := b12Push(g, me.ID, "Briar Hydra", "Creature — Plant Hydra", dmuBriarHydraOracle, 6, 6)
	domainLands(g, me, opp, "Forest")
	advanceToMain(t, g)
	dealNoncombatDamage(g, me.ID, hydra, opp.ID)
	passPriorityAroundTable(t, g)
	if n := dmuB11Counters(t, g, hydra); n != 0 {
		t.Errorf("noncombat damage gave %d counters", n)
	}
}

func TestDMUB11RadhaPumpsAnotherCreatureWhenTapped(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	radha := b12Push(g, me.ID, "Radha, Coalition Warlord", "Legendary Creature — Elf Warrior", dmuRadhaOracle, 3, 3)
	bear := b12Creature(g, me.ID, "Bear", dmuB11CreatureTypeLine, 2, 2)
	domainLands(g, me, opp, "Forest", "Mountain") // three types
	dmuB11Attack(t, g, me, opp, radha)
	dmuB11PickIfAsked(t, g, me.ID, bear)
	passPriorityAroundTable(t, g)
	if c := gaLayered(t, g, bear).Effective(); c.Power != 5 || c.Toughness != 5 {
		t.Errorf("Bear %d/%d, want 5/5", c.Power, c.Toughness)
	}
	if c := gaLayered(t, g, radha).Effective(); c.Power != 3 {
		t.Errorf("Radha pumped herself: power %d", c.Power)
	}
}

func TestDMUB11RadhaHasNoTargetWithoutAnotherCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	radha := b12Push(g, me.ID, "Radha, Coalition Warlord", "Legendary Creature — Elf Warrior", dmuRadhaOracle, 3, 3)
	domainLands(g, me, opp, "Forest")
	dmuB11Attack(t, g, me, opp, radha)
	passPriorityAroundTable(t, g)
	if latestPickTarget(g, me.ID) != nil {
		t.Error("a target prompt opened with no other creature")
	}
	if c := gaLayered(t, g, radha).Effective(); c.Power != 3 {
		t.Errorf("Radha %d power, want 3", c.Power)
	}
}

func TestDMUB11ZarOjanenCountersEachSmallCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	zar := b12Push(g, me.ID, "Zar Ojanen, Scion of Efrava", "Legendary Creature — Cat Warrior", dmuZarOracle, 4, 4)
	small := b12Creature(g, me.ID, "Small", dmuB11CreatureTypeLine, 1, 1)
	mid := b12Creature(g, me.ID, "Mid", dmuB11CreatureTypeLine, 2, 3)
	big := b12Creature(g, me.ID, "Big", dmuB11CreatureTypeLine, 4, 4)
	theirs := b12Creature(g, opp.ID, "Theirs", dmuB11CreatureTypeLine, 1, 1)
	domainLands(g, me, opp, "Forest", "Island", "Swamp") // four types
	dmuB11Attack(t, g, me, opp, zar)
	passPriorityAroundTable(t, g)
	for id, want := range map[uuid.UUID]int{small: 1, mid: 1, big: 0, zar: 0, theirs: 0} {
		if n := dmuB11Counters(t, g, id); n != want {
			t.Errorf("counters on %s: %d, want %d", id, n, want)
		}
	}
}

func TestDMUB11QuirionBeastcallerGrowsOnCreatureSpellsAndSpreadsOnDeath(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bc := b12Push(g, me.ID, "Quirion Beastcaller", "Creature — Dryad Warrior", dmuBeastcallerOracle, 2, 2)
	other := b12Creature(g, me.ID, "Bear", dmuB11CreatureTypeLine, 2, 2)
	advanceToMain(t, g)
	avB1ResolveCreature(t, g, me, "Cast Bear", dmuB11CreatureTypeLine, "")
	passPriorityAroundTable(t, g)
	avB1ResolveCreature(t, g, me, "Cast Bear Two", dmuB11CreatureTypeLine, "")
	passPriorityAroundTable(t, g)
	if n := dmuB11Counters(t, g, bc); n != 2 {
		t.Fatalf("counters after two creature spells: %d, want 2", n)
	}
	// A noncreature spell adds nothing.
	castCatalogSpell(t, g, "Plain Sorcery", "Sorcery", "", nil)
	passPriorityAroundTable(t, g)
	if n := dmuB11Counters(t, g, bc); n != 2 {
		t.Fatalf("a sorcery added a counter: %d", n)
	}

	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(bc, game.DestroyOptions{}); err != nil {
			t.Fatal(err)
		}
	})
	g.RunStateChecksForTest()
	if c := latestPickTarget(g, me.ID); c == nil {
		t.Fatalf("no target prompt for the death trigger: %+v", g.PendingChoices)
	}
	pickCard(t, g, me.ID, other)
	passPriorityAroundTable(t, g)
	if n := dmuB11Counters(t, g, other); n != 2 {
		t.Errorf("counters distributed to the other creature: %d, want 2", n)
	}
}

func TestDMUB11TatyovaGrantsFlyingToLandCreaturesOnly(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Push(g, me.ID, "Tatyova, Steward of Tides", "Legendary Creature — Merfolk Druid", dmuTatyovaOracle, 3, 3)
	animated := b12Push(g, me.ID, "Animated Forest", "Land Creature — Forest", "", 2, 2)
	plain := pushLandFor(g, me.ID, "Plains", "Basic Land — Plains")
	bear := b12Creature(g, me.ID, "Bear", dmuB11CreatureTypeLine, 2, 2)
	theirs := b12Push(g, opp.ID, "Their Land Creature", "Land Creature — Forest", "", 2, 2)
	g.WithWriteLock(func() { g.BumpLayerVersionForTest() })
	if !avB22Has(t, g, animated, "flying") {
		t.Error("a land creature you control lacks flying")
	}
	for name, id := range map[string]uuid.UUID{"a plain land": plain, "a nonland creature": bear, "an opponent's land creature": theirs} {
		if avB22Has(t, g, id, "flying") {
			t.Errorf("%s has flying", name)
		}
	}
}

func TestDMUB11TatyovaAnimatesALandOnlyWithSevenLands(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b12Push(g, me.ID, "Tatyova, Steward of Tides", "Legendary Creature — Merfolk Druid", dmuTatyovaOracle, 3, 3)
	var lands []uuid.UUID
	for i := 0; i < 5; i++ {
		lands = append(lands, pushLandFor(g, me.ID, "Forest", "Basic Land — Forest"))
	}
	advanceToMain(t, g)

	// The sixth land is not enough: the intervening "if" fails (CR 603.4).
	enter := func() {
		g.WithWriteLock(func() {
			id := uuid.New()
			g.Battlefield.PushTop(game.Card{InstanceID: id, Name: "Forest", TypeLine: "Basic Land — Forest", Owner: me.ID, Controller: me.ID})
			g.EmitEvent(game.Event{Kind: game.EventETB, CardID: id, Actor: me.ID})
		})
	}
	enter()
	passPriorityAroundTable(t, g)
	if latestPickTarget(g, me.ID) != nil || len(g.StackMeta) != 0 {
		t.Fatal("the trigger went on the stack with six lands")
	}

	// A seventh land through an effect makes seven.
	enter()
	dmuB11PickIfAsked(t, g, me.ID, lands[0])
	passPriorityAroundTable(t, g)
	c := gaLayered(t, g, lands[0])
	if !c.IsCreature() || !c.IsLand() || c.CurrentPower() != 3 || c.CurrentToughness() != 3 || !game.HasKeyword(&c, "haste") || !c.HasSubtype("Elemental") {
		t.Errorf("animated land: %+v", c)
	}
	if !game.HasKeyword(&c, "flying") {
		t.Error("the animated land creature lacks Tatyova's flying")
	}
}

func TestDMUB11RulikMonsPutsALandOrMakesAGoblin(t *testing.T) {
	// A land on top, taken: no Goblin.
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	rulik := b12Push(g, me.ID, "Rulik Mons, Warren Chief", "Legendary Creature — Goblin", dmuRulikOracle, 3, 3)
	lib := dftMPLibraryTop(me, c2b15Forest(), c2b15Bear())
	dmuB11Attack(t, g, me, opp, rulik)
	passPriorityAroundTable(t, g)
	answerChooseCards(t, g, me.ID, lib[0])
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(lib[0]) || !c2b16IsTapped(g, lib[0]) {
		t.Error("the land is not on the battlefield tapped")
	}
	if n := dmuB11Goblins(g, me.ID); n != 0 {
		t.Errorf("a Goblin was made after a land was put: %d", n)
	}

	// A land on top, declined: a Goblin, and the land stays on top.
	g = newCatalogGame(t)
	me, opp = g.Seats[0], g.Seats[1]
	rulik = b12Push(g, me.ID, "Rulik Mons, Warren Chief", "Legendary Creature — Goblin", dmuRulikOracle, 3, 3)
	lib = dftMPLibraryTop(me, c2b15Forest(), c2b15Bear())
	dmuB11Attack(t, g, me, opp, rulik)
	passPriorityAroundTable(t, g)
	answerChooseCards(t, g, me.ID)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(lib[0]) {
		t.Error("a declined land entered")
	}
	if n := dmuB11Goblins(g, me.ID); n != 1 {
		t.Errorf("Goblins after declining: %d, want 1", n)
	}
	if top := me.Library.Cards[len(me.Library.Cards)-1].InstanceID; top != lib[0] {
		t.Error("the declined land is not still on top")
	}

	// A nonland on top: a Goblin.
	g = newCatalogGame(t)
	me, opp = g.Seats[0], g.Seats[1]
	rulik = b12Push(g, me.ID, "Rulik Mons, Warren Chief", "Legendary Creature — Goblin", dmuRulikOracle, 3, 3)
	dftMPLibraryTop(me, c2b15Bear(), c2b15Forest())
	dmuB11Attack(t, g, me, opp, rulik)
	passPriorityAroundTable(t, g)
	if n := dmuB11Goblins(g, me.ID); n != 1 {
		t.Errorf("Goblins after a nonland: %d, want 1", n)
	}
}

func dmuB11Goblins(g *game.Game, owner uuid.UUID) int {
	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.Controller == owner && c.IsToken() && c.Name == "Goblin" && c.Power == 1 && c.Toughness == 1 {
			n++
		}
	}
	return n
}

func TestDMUB11SilverbackElderModes(t *testing.T) {
	cast := func(t *testing.T, g *game.Game, me *game.Player) *game.PendingChoice {
		t.Helper()
		advanceToMain(t, g)
		avB1Cast(t, g, me, game.Card{Name: "Cast Bear", TypeLine: dmuB11CreatureTypeLine, Power: 2, Toughness: 2}, game.CastSpellParams{})
		c := modePickChoiceFor(g, me.ID)
		if c == nil {
			t.Fatalf("no mode prompt: %+v", g.PendingChoices)
		}
		return c
	}
	t.Run("destroy", func(t *testing.T) {
		g := newCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		b12Push(g, me.ID, "Silverback Elder", "Creature — Ape Shaman", dmuSilverbackOracle, 5, 7)
		rock := b12Permanent(g, opp.ID, "Rock", "Artifact")
		aura := b12Permanent(g, opp.ID, "Aura", "Enchantment")
		land := b12Permanent(g, opp.ID, "Island", "Basic Land — Island")
		answerMode(t, g, cast(t, g, me), me.ID, 0)
		b04WaitForPick(t, g, me.ID)
		if err := g.ResolvePickTarget(latestPickTarget(g, me.ID).ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: land}); err == nil {
			t.Fatal("a land is not an artifact or enchantment")
		}
		pickCard(t, g, me.ID, rock)
		passPriorityAroundTable(t, g)
		if g.Battlefield.Contains(rock) {
			t.Error("the artifact survived")
		}
		if !g.Battlefield.Contains(aura) {
			t.Error("the other target was destroyed")
		}
	})
	t.Run("gain life", func(t *testing.T) {
		g := newCatalogGame(t)
		me := g.Seats[0]
		b12Push(g, me.ID, "Silverback Elder", "Creature — Ape Shaman", dmuSilverbackOracle, 5, 7)
		life := me.Life
		answerMode(t, g, cast(t, g, me), me.ID, 2)
		passPriorityAroundTable(t, g)
		if me.Life != life+4 {
			t.Errorf("life %d, want %d", me.Life, life+4)
		}
	})
	t.Run("dig", func(t *testing.T) {
		g := newCatalogGame(t)
		me := g.Seats[0]
		b12Push(g, me.ID, "Silverback Elder", "Creature — Ape Shaman", dmuSilverbackOracle, 5, 7)
		lib := dftMPLibraryTop(me, c2b15Bear(), c2b15Forest(), c2b15Spell(), c2b15Bear(), c2b15Bear(), c2b15Forest())
		answerMode(t, g, cast(t, g, me), me.ID, 1)
		passPriorityAroundTable(t, g)
		ch := chooseCardsChoiceFor(g, me.ID)
		if ch == nil || ch.ChooseMin != 0 || ch.ChooseMax != 1 {
			t.Fatalf("want an optional pick of one land: %+v", ch)
		}
		answerChooseCards(t, g, me.ID, lib[1])
		if !g.Battlefield.Contains(lib[1]) || !c2b16IsTapped(g, lib[1]) {
			t.Error("the land is not on the battlefield tapped")
		}
		if top := me.Library.Cards[len(me.Library.Cards)-1].InstanceID; top != lib[5] {
			t.Error("the sixth card is not on top")
		}
		for _, i := range []int{0, 2, 3, 4} {
			if !me.Library.Contains(lib[i]) {
				t.Errorf("card %d is not back in the library", i)
			}
		}
	})
}

func TestDMUB11NaelLooksAtDomainCardsAndDrawsWithFiveTypes(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	nael := b12Push(g, me.ID, "Nael, Avizoa Aeronaut", "Legendary Creature — Elf Scout", dmuNaelOracle, 2, 4)
	lib := dftMPLibraryTop(me, c2b15Bear(), c2b15Forest(), c2b15Spell(), c2b15Bear(), c2b15Bear())
	domainLands(g, me, opp, "Forest", "Island") // three types: look at three
	advanceToMain(t, g)
	hand := me.Hand.Size()
	dealCombatDamageToPlayer(g, nael, opp.ID, 2)
	passPriorityAroundTable(t, g)
	pick := latestRevealPickFor(g, me.ID)
	if pick == nil {
		t.Fatalf("no look prompt: %+v", g.PendingChoices)
	}
	if len(pick.ChooseCards) != 3 {
		t.Fatalf("looked at %d cards, want 3", len(pick.ChooseCards))
	}
	if err := g.ResolveRevealPick(pick.ID, me.ID, []uuid.UUID{lib[1]}); err != nil {
		t.Fatalf("ResolveRevealPick: %v", err)
	}
	passPriorityAroundTable(t, g)
	if top := me.Library.Cards[len(me.Library.Cards)-1].InstanceID; top != lib[1] {
		t.Error("the kept card is not on top")
	}
	if me.Hand.Size() != hand {
		t.Error("drew with only three basic land types")
	}
	_ = lib
}

func TestDMUB11NaelDrawsAfterArrangingWithFiveTypes(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	nael := b12Push(g, me.ID, "Nael, Avizoa Aeronaut", "Legendary Creature — Elf Scout", dmuNaelOracle, 2, 4)
	lib := dftMPLibraryTop(me, c2b15Bear(), c2b15Forest(), c2b15Spell(), c2b15Bear(), c2b15Bear(), c2b15Bear())
	domainLands(g, me, opp, "Island", "Swamp", "Mountain", "Forest") // all five
	advanceToMain(t, g)
	hand := me.Hand.Size()
	dealCombatDamageToPlayer(g, nael, opp.ID, 2)
	passPriorityAroundTable(t, g)
	pick := latestRevealPickFor(g, me.ID)
	if pick == nil || len(pick.ChooseCards) != 5 {
		t.Fatalf("want a look at five cards: %+v", pick)
	}
	if err := g.ResolveRevealPick(pick.ID, me.ID, []uuid.UUID{lib[2]}); err != nil {
		t.Fatalf("ResolveRevealPick: %v", err)
	}
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 || !me.Hand.Contains(lib[2]) {
		t.Errorf("hand %d, want %d holding the kept card", me.Hand.Size(), hand+1)
	}
}

func TestDMUB11SphinxOfClearSkiesOpponentSeparatesAndYouTakeAPile(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	sphinx := b12Push(g, me.ID, "Sphinx of Clear Skies", "Creature — Sphinx", dmuSphinxOracle, 5, 5)
	lib := dftMPLibraryTop(me, c2b15Bear(), c2b15Forest(), c2b15Spell(), c2b15Bear())
	domainLands(g, me, opp, "Island", "Forest") // three types
	advanceToMain(t, g)
	hand := me.Hand.Size()
	dealCombatDamageToPlayer(g, sphinx, opp.ID, 5)
	passPriorityAroundTable(t, g)
	if latestOptionPickFor(g, me.ID) != nil {
		answerChoosePlayer(t, g, me.ID, opp)
	}
	split := latestRevealPickFor(g, opp.ID)
	if split == nil {
		t.Fatalf("the opponent was not asked to separate: %+v", g.PendingChoices)
	}
	if len(split.ChooseCards) != 3 {
		t.Fatalf("separating %d cards, want 3", len(split.ChooseCards))
	}
	if err := g.ResolveRevealPick(split.ID, opp.ID, []uuid.UUID{lib[0]}); err != nil { // pile one: lib[0]; pile two: lib[1], lib[2]
		t.Fatalf("ResolveRevealPick: %v", err)
	}
	pile := latestOptionPickFor(g, me.ID)
	if pile == nil {
		t.Fatalf("no pile pick: %+v", g.PendingChoices)
	}
	if err := g.ResolveOptionPick(pile.ID, me.ID, 1); err != nil {
		t.Fatalf("ResolveOptionPick: %v", err)
	}
	passPriorityAroundTable(t, g)
	toHand, toYard := 0, 0
	for _, id := range lib[:3] {
		switch {
		case me.Hand.Contains(id):
			toHand++
		case me.Graveyard.Contains(id):
			toYard++
		}
	}
	if toHand+toYard != 3 || toHand == 0 && toYard == 0 {
		t.Errorf("hand %d, graveyard %d", toHand, toYard)
	}
	if me.Hand.Size() != hand+toHand {
		t.Errorf("hand size %d", me.Hand.Size())
	}
	if !me.Library.Contains(lib[3]) {
		t.Error("the fourth card left the library")
	}
}

func TestDMUB11ThreatsUndetectedDifferentPowersOpponentChoosesTwo(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	mk := func(name string, power int) uuid.UUID {
		return pushLibraryCardForTest(me, game.Card{Name: name, TypeLine: "Creature — Bear", Power: power, Toughness: 2})
	}
	a, b, c, d, dup := mk("Zero", 0), mk("One", 1), mk("Two", 2), mk("Three", 3), mk("Another Two", 2)
	castCatalogSpell(t, g, "Threats Undetected", "Sorcery", dmuThreatsOracle, nil)
	passPriorityAroundTable(t, g)
	prompt := searchPromptFor(g, me.ID)
	if prompt == nil {
		t.Fatalf("no search prompt: %+v", g.PendingChoices)
	}
	if err := g.ResolveSearchLibrary(prompt.ID, me.ID, []uuid.UUID{a, c, dup}); err == nil {
		t.Fatal("two cards with power two were found")
	}
	if err := g.ResolveSearchLibrary(prompt.ID, me.ID, []uuid.UUID{a, b, c, d}); err != nil {
		t.Fatalf("ResolveSearchLibrary: %v", err)
	}
	if latestOptionPickFor(g, me.ID) != nil {
		answerChoosePlayer(t, g, me.ID, opp)
	}
	pick := latestRevealPickFor(g, opp.ID)
	if pick == nil {
		t.Fatalf("the opponent was not asked to choose: %+v", g.PendingChoices)
	}
	if err := g.ResolveRevealPick(pick.ID, opp.ID, []uuid.UUID{a, d}); err != nil {
		t.Fatalf("ResolveRevealPick: %v", err)
	}
	passPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{b, c} {
		if !me.Hand.Contains(id) {
			t.Errorf("an unchosen card %s is not in hand", id)
		}
	}
	for _, id := range []uuid.UUID{a, d, dup} {
		if !me.Library.Contains(id) {
			t.Errorf("card %s is not in the library", id)
		}
	}
}
