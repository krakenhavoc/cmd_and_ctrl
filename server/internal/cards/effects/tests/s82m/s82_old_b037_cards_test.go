package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b037_cards_test.go — S82 older sets, Modern Horizons 2 batch 37
// (#5501): the Arcbound cards, Break Ties, Wren's Run Hydra, Bannerhide
// Krushok and the three Equipment.

const (
	oldB37Javelineer = "48e3f6dd-68ab-420a-ad4d-fe7ee39486d3"
	oldB37Zabaz      = "549dc9f3-1fda-4ad3-87bc-8a990800380b"
	oldB37Shikari    = "acb8d6a0-4bbd-4a42-bcac-3656367cc051"
	oldB37Slasher    = "bf86376d-474d-4658-8cf3-d0e2a01cdfff"
	oldB37Tracker    = "ff86f385-dc3e-4bfe-83e7-9229afab4aa9"
	oldB37BreakTies  = "1e44bb4e-aeea-4f43-9cbd-4c24a9b6b022"
	oldB37Hydra      = "a2196bc3-2fe6-49d4-8c79-d6beb3d6c5d8"
	oldB37Krushok    = "e3191093-1aaa-4002-a603-7c8d132b6a31"
	oldB37Spear      = "2202faf1-5638-46bf-8bf4-91870b669717"
	oldB37Spike      = "28b22709-3e69-4eb7-8a3d-b75b6e65f30e"
	oldB37Gloves     = "41074ef0-8686-41b5-9b99-7103192da915"

	oldB37ArtCreature = "Artifact Creature — Construct"
)

func TestOldB37CardsAreFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB37Javelineer: "Arcbound Javelineer", oldB37Zabaz: "Zabaz, the Glimmerwasp",
		oldB37Shikari: "Arcbound Shikari", oldB37Slasher: "Arcbound Slasher",
		oldB37Tracker: "Arcbound Tracker", oldB37BreakTies: "Break Ties",
		oldB37Hydra: "Wren's Run Hydra", oldB37Krushok: "Bannerhide Krushok",
		oldB37Spear: "Lightning Spear", oldB37Spike: "Barbed Spike", oldB37Gloves: "Vectis Gloves",
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

func TestOldB37JavelineerRemovesXCountersForXDamageToAnAttacker(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	jav := effectstest.MrdEnter(t, g, me, "Arcbound Javelineer", "Artifact Creature — Soldier", oldB37Javelineer)
	effectstest.AddCounters(t, g, jav, game.CounterPlusOne, 2)
	if got := effectstest.CountersOn(g, jav, game.CounterPlusOne); got != 3 {
		t.Fatalf("%d counters, want modular 1 plus two", got)
	}
	idle := effectstest.B12Creature(g, opp.ID, "Idle Ogre", "Creature — Ogre", 4, 5)
	attacker := effectstest.B12Creature(g, me.ID, "Attacker", "Creature — Bear", 2, 2)
	effectstest.DeclareAttack(t, g, opp.ID, attacker)
	err := g.ActivateCatalogAbility(me.ID, jav, 0, game.ActivateAbilityParams{
		Strict: true, Targets: effectstest.CardTargets(idle),
		CounterSourceIDs: []uuid.UUID{jav}, CounterCounts: []int{1},
	})
	if err == nil {
		t.Fatal("targeted a creature that is neither attacking nor blocking")
	}
	effectstest.B16Activate(t, g, me.ID, jav, 0, game.ActivateAbilityParams{
		Strict: true, Targets: effectstest.CardTargets(attacker),
		CounterSourceIDs: []uuid.UUID{jav}, CounterCounts: []int{2},
	})
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(attacker) {
		t.Error("a 2/2 survived 2 damage")
	}
	if got := effectstest.CountersOn(g, jav, game.CounterPlusOne); got != 1 {
		t.Errorf("%d counters left, want 1", got)
	}
	if !effectstest.IsTapped(g, jav) {
		t.Error("the Javelineer did not tap as a cost")
	}
}

func TestOldB37ZabazAddsOneToAModularTriggersCounters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	zabaz := effectstest.MrdEnter(t, g, me, "Zabaz, the Glimmerwasp", "Legendary Artifact Creature — Insect", oldB37Zabaz)
	if got := effectstest.CountersOn(g, zabaz, game.CounterPlusOne); got != 1 {
		t.Fatalf("Zabaz entered with %d counters, want modular 1 only (an entry is not a trigger)", got)
	}
	// A counter placed any other way gets no bonus.
	effectstest.AddCounters(t, g, zabaz, game.CounterPlusOne, 1)
	if got := effectstest.CountersOn(g, zabaz, game.CounterPlusOne); got != 2 {
		t.Fatalf("a plain counter made %d in all, want 2", got)
	}
	tracker := effectstest.MrdEnter(t, g, me, "Arcbound Tracker", oldB37ArtCreature, oldB37Tracker)
	if got := effectstest.CountersOn(g, tracker, game.CounterPlusOne); got != 2 {
		t.Fatalf("the Tracker entered with %d counters, want 2", got)
	}
	effectstest.B25Destroy(g, tracker)
	effectstest.AnswerPickTarget(t, g, zabaz)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	// 2 already there, plus the Tracker's 2, plus Zabaz's one more.
	if got := effectstest.CountersOn(g, zabaz, game.CounterPlusOne); got != 5 {
		t.Errorf("Zabaz has %d counters after the modular trigger, want 5", got)
	}
}

func TestOldB37ModularWithoutZabazIsUnchanged(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	host := effectstest.MrdPermanent(g, me.ID, "Host", oldB37ArtCreature, "{1}", 1, 1)
	tracker := effectstest.MrdEnter(t, g, me, "Arcbound Tracker", oldB37ArtCreature, oldB37Tracker)
	effectstest.B25Destroy(g, tracker)
	effectstest.AnswerPickTarget(t, g, host)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountersOn(g, host, game.CounterPlusOne); got != 2 {
		t.Errorf("the host has %d counters, want 2", got)
	}
}

func TestOldB37ZabazAbilities(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	zabaz := effectstest.MrdEnter(t, g, me, "Zabaz, the Glimmerwasp", "Legendary Artifact Creature — Insect", oldB37Zabaz)
	rock := effectstest.MrdPermanent(g, me.ID, "Rock", "Artifact", "{1}", 0, 0)
	opp := g.Seats[1]
	theirs := effectstest.MrdPermanent(g, opp.ID, "Their Rock", "Artifact", "{1}", 0, 0)
	effectstest.ApaMana(me, "R")
	if err := g.ActivateCatalogAbility(me.ID, zabaz, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardTargets(theirs)}); err == nil {
		t.Error("destroyed an artifact its controller does not control")
	}
	effectstest.ApaActivate(t, g, me, zabaz, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardTargets(rock)})
	if g.Battlefield.Contains(rock) {
		t.Error("the Rock survived {R}: destroy target artifact you control")
	}
	if effectstest.HasKeywordOnBattlefield(t, g, zabaz, "flying") {
		t.Fatal("Zabaz flew before {W}")
	}
	effectstest.ApaMana(me, "W")
	effectstest.ApaActivate(t, g, me, zabaz, 1, game.ActivateAbilityParams{Strict: true})
	if !effectstest.HasKeywordOnBattlefield(t, g, zabaz, "flying") {
		t.Error("Zabaz did not gain flying")
	}
	effectstest.AdvancePastCleanupOf(t, g, 0)
	if effectstest.HasKeywordOnBattlefield(t, g, zabaz, "flying") {
		t.Error("flying outlived the turn")
	}
}

func TestOldB37ShikariPutsACounterOnEachOtherArtifactCreatureYouControl(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	a := effectstest.MrdPermanent(g, me.ID, "Art A", oldB37ArtCreature, "{1}", 1, 1)
	b := effectstest.MrdPermanent(g, me.ID, "Art B", oldB37ArtCreature, "{1}", 1, 1)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.MrdPermanent(g, opp.ID, "Their Art", oldB37ArtCreature, "{1}", 1, 1)
	rock := effectstest.MrdPermanent(g, me.ID, "Rock", "Artifact", "{1}", 0, 0)
	shikari := effectstest.MrdEnter(t, g, me, "Arcbound Shikari", oldB37ArtCreature, oldB37Shikari)
	effectstest.PassPriorityAroundTable(t, g)
	for id, want := range map[uuid.UUID]int{a: 1, b: 1, bear: 0, theirs: 0, rock: 0, shikari: 2} {
		if got := effectstest.CountersOn(g, id, game.CounterPlusOne); got != want {
			t.Errorf("%v has %d counters, want %d", id, got, want)
		}
	}
}

func TestOldB37SlasherEntersWithModularFourAndRiot(t *testing.T) {
	for _, counter := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		id := effectstest.CastCatalogSpell(t, g, "Arcbound Slasher", oldB37ArtCreature, oldB37Slasher, nil)
		effectstest.PassPriorityAroundTable(t, g)
		effectstest.AnswerRiotFor(t, g, counter)
		effectstest.PassPriorityAroundTable(t, g)
		want := 4
		if counter {
			want = 5
		}
		if got := effectstest.CountersOn(g, id, game.CounterPlusOne); got != want {
			t.Errorf("riot counter=%v: %d counters, want %d", counter, got, want)
		}
		if got := effectstest.HasKeywordOnBattlefield(t, g, id, "haste"); got == counter {
			t.Errorf("riot counter=%v: haste=%v", counter, got)
		}
	}
}

func TestOldB37TrackerGrowsOnEverySpellButTheFirst(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	tracker := effectstest.MrdEnter(t, g, me, "Arcbound Tracker", oldB37ArtCreature, oldB37Tracker)
	if got := effectstest.CountersOn(g, tracker, game.CounterPlusOne); got != 2 {
		t.Fatalf("entered with %d counters, want 2", got)
	}
	for i, want := range []int{2, 3, 4} {
		effectstest.CastCatalogSpell(t, g, "Filler", "Sorcery", "", nil)
		effectstest.PassPriorityAroundTable(t, g)
		if got := effectstest.CountersOn(g, tracker, game.CounterPlusOne); got != want {
			t.Errorf("after spell %d the Tracker has %d counters, want %d", i+1, got, want)
		}
	}
}

func TestOldB37BreakTiesModes(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	art := effectstest.MrdPermanent(g, opp.ID, "Their Rock", "Artifact", "{1}", 0, 0)
	ench := effectstest.B12Permanent(g, opp.ID, "Their Aura", "Enchantment")
	card := game.NewCard("Buried Bear", opp.ID)
	card.TypeLine = "Creature — Bear"
	opp.Graveyard.PushTop(card)

	effectstest.P1g3CastModal(t, g, "Break Ties", "Instant", oldB37BreakTies, []int{0}, game.TargetRef{Kind: game.TargetCard, ID: art})
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(art) || !g.Battlefield.Contains(ench) {
		t.Error("mode 1 did not destroy exactly the artifact")
	}
	effectstest.P1g3CastModal(t, g, "Break Ties", "Instant", oldB37BreakTies, []int{1}, game.TargetRef{Kind: game.TargetCard, ID: ench})
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(ench) {
		t.Error("mode 2 did not destroy the enchantment")
	}
	effectstest.P1g3CastModal(t, g, "Break Ties", "Instant", oldB37BreakTies, []int{2}, game.TargetRef{Kind: game.TargetCard, ID: card.InstanceID})
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Graveyard.Contains(card.InstanceID) || !g.Exile.Contains(card.InstanceID) {
		t.Error("mode 3 did not exile the graveyard card")
	}
}

func TestOldB37BreakTiesRefusesAnIllegalTarget(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	bear := effectstest.B12Creature(g, g.Seats[1].ID, "Bear", "Creature — Bear", 2, 2)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	id := effectstest.B13HandCard(me, "Break Ties", "Instant", oldB37BreakTies)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, Modes: []int{0}, Targets: effectstest.CardTargets(bear)}); err == nil {
		t.Error("destroy target artifact accepted a creature")
	}
}

func TestOldB37ReinforceFromTheHand(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, oracle string
		mana                   []string
		n                      int
	}{
		{"Break Ties", "Instant", oldB37BreakTies, []string{"W"}, 1},
		{"Bannerhide Krushok", "Creature — Beast", oldB37Krushok, []string{"G", "C"}, 2},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			effectstest.AdvanceToMain(t, g)
			bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
			card := effectstest.B13HandCard(me, tc.name, tc.typeLine, tc.oracle)
			effectstest.ApaMana(me, tc.mana...)
			idx := 0
			if tc.name == "Break Ties" {
				idx = 0
			}
			effectstest.ApaActivate(t, g, me, card, idx, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardTargets(bear)})
			if got := effectstest.CountersOn(g, bear, game.CounterPlusOne); got != tc.n {
				t.Errorf("%d +1/+1 counters, want %d", got, tc.n)
			}
			if effectstest.B12ZoneOf(g, card) != game.ZoneGraveyard {
				t.Error("the card was not discarded to pay for reinforce")
			}
		})
	}
}

func TestOldB37HydraEntersWithXCountersAndReinforcesForX(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	id := effectstest.CastCatalogSpell(t, g, "Wren's Run Hydra", "Creature — Hydra", oldB37Hydra, nil)
	_ = id
	g2 := effectstest.NewCatalogGame(t)
	me2 := g2.Seats[g2.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g2)
	bear := effectstest.B12Creature(g2, me2.ID, "Bear", "Creature — Bear", 2, 2)
	card := effectstest.B13HandCard(me2, "Wren's Run Hydra", "Creature — Hydra", oldB37Hydra)
	effectstest.ApaMana(me2, "G", "G", "C", "C", "C")
	effectstest.ApaActivate(t, g2, me2, card, 0, game.ActivateAbilityParams{Strict: true, XValue: 3, Targets: effectstest.CardTargets(bear)})
	if got := effectstest.CountersOn(g2, bear, game.CounterPlusOne); got != 3 {
		t.Errorf("reinforce X=3 put %d counters, want 3", got)
	}
	_ = me
}

func TestOldB37HydraCastForXEntersWithXCounters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.Pr8CastX(t, g, "Wren's Run Hydra", "Creature — Hydra", oldB37Hydra, 4, nil)
	var found bool
	for _, c := range g.Battlefield.Cards {
		if c.OracleID == oldB37Hydra {
			found = true
			if got := c.Counters[game.CounterPlusOne]; got != 4 {
				t.Errorf("entered with %d counters, want 4", got)
			}
		}
	}
	if !found {
		t.Fatal("the Hydra did not enter")
	}
}

func TestOldB37KrushokScavenge(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	card := game.NewCard("Bannerhide Krushok", me.ID)
	card.TypeLine, card.OracleID, card.Power, card.Toughness = "Creature — Beast", oldB37Krushok, 4, 4
	me.Graveyard.PushTop(card)
	effectstest.ApaMana(me, "G", "G", "C", "C", "C", "C", "C")
	effectstest.ApaActivate(t, g, me, card.InstanceID, 1, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardTargets(bear)})
	if got := effectstest.CountersOn(g, bear, game.CounterPlusOne); got != 4 {
		t.Errorf("scavenge put %d counters, want 4 (its power)", got)
	}
	if me.Graveyard.Contains(card.InstanceID) {
		t.Error("the Krushok stayed in the graveyard")
	}
}

func TestOldB37LightningSpearPumpsAndSacrificesForThree(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	bear := effectstest.SeedBear(g, me.ID)
	spear := effectstest.SeedEquipment(g, me.ID, "Lightning Spear", oldB37Spear)
	effectstest.EquipTo(t, g, me.ID, spear, bear)
	if p := effectstest.EffectivePower(t, g, bear); p != 3 {
		t.Errorf("equipped power %d, want 3", p)
	}
	if !effectstest.HasEffectiveKeyword(t, g, bear, "trample") {
		t.Error("the equipped creature lacks trample")
	}
	life := opp.Life
	effectstest.ApaMana(me, "R", "C", "C")
	effectstest.ApaActivate(t, g, me, spear, 0, game.ActivateAbilityParams{Strict: true, Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}})
	if got := life - opp.Life; got != 3 {
		t.Errorf("the Spear dealt %d, want 3", got)
	}
	if g.Battlefield.Contains(spear) {
		t.Error("the Spear was not sacrificed")
	}
}

func TestOldB37BarbedSpikeMakesAThopterAndAttachesToIt(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	spike := effectstest.CastCatalogSpell(t, g, "Barbed Spike", "Artifact — Equipment", oldB37Spike, nil)
	effectstest.PassPriorityAroundTable(t, g)
	var thopter uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Thopter" && c.Controller == me.ID {
			thopter = c.InstanceID
		}
	}
	if thopter == uuid.Nil {
		t.Fatal("no Thopter token")
	}
	if !effectstest.AttachedTo(t, g, spike, thopter) {
		t.Error("the Spike is not attached to the Thopter")
	}
	if p := effectstest.EffectivePower(t, g, thopter); p != 2 {
		t.Errorf("the Thopter has power %d, want 2", p)
	}
	if !effectstest.HasEffectiveKeyword(t, g, thopter, "flying") {
		t.Error("the Thopter lacks flying")
	}
}

func TestOldB37VectisGlovesGiveArtifactLandwalk(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	bear := effectstest.SeedBear(g, me.ID)
	gloves := effectstest.SeedEquipment(g, me.ID, "Vectis Gloves", oldB37Gloves)
	effectstest.EquipTo(t, g, me.ID, gloves, bear)
	if p := effectstest.EffectivePower(t, g, bear); p != 4 {
		t.Errorf("equipped power %d, want 4", p)
	}
	if !effectstest.HasEffectiveKeyword(t, g, bear, "artifact landwalk") {
		t.Error("the equipped creature lacks artifact landwalk")
	}
}
