package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b369_cards_test.go — Journey into Nyx, batch 369 (#5558, S82 OLD).
// Every test plays the real catalog entry through the engine.

const (
	oldB369StrengthOracle    = "23643af9-00b7-45e1-8471-7031d08bbe30"
	oldB369MaggotOracle      = "419b5b4f-cac7-44bf-b1e2-655b5a5d37b1"
	oldB369HallOracle        = "716632dd-2947-4641-a1db-50e0a2b0b4e8"
	oldB369ColossusOracle    = "ad5e2208-df68-4119-b225-a7c133b70359"
	oldB369WardenOracle      = "b03397e5-56af-47c0-ab9a-746a0c99c568"
	oldB369ThraxesOracle     = "52289774-d292-4028-91fb-8826fa004dc8"
	oldB369SpiteOracle       = "aab5f76b-1eb0-4fbc-ab01-593161083c16"
	oldB369FlamespeakerOracl = "b290a59a-5752-46bb-90be-9792b57bd467"
	oldB369StarfallOracle    = "b2c840a7-24b4-4ae3-a13f-ed524e18d488"
	oldB369KnowledgeOracle   = "ed68b5ca-c52a-40e0-89ac-6276a3c4fdeb"
)

func TestOldB369CardsAreRegisteredAndWhole(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB369StrengthOracle:    "Strength from the Fallen",
		oldB369MaggotOracle:      "Brain Maggot",
		oldB369HallOracle:        "Hall of Triumph",
		oldB369ColossusOracle:    "Quarry Colossus",
		oldB369WardenOracle:      "Nessian Game Warden",
		oldB369ThraxesOracle:     "Spawn of Thraxes",
		oldB369SpiteOracle:       "Spite of Mogis",
		oldB369FlamespeakerOracl: "Prophetic Flamespeaker",
		oldB369StarfallOracle:    "Starfall",
		oldB369KnowledgeOracle:   "Knowledge and Power",
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

func oldB369Seats(g *game.Game) (me, opp *game.Player) {
	return g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
}

func oldB369Card(id uuid.UUID) []game.TargetRef {
	return []game.TargetRef{{Kind: game.TargetCard, ID: id}}
}

func TestOldB369BrainMaggotTakesANonlandCardUntilItLeaves(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := oldB369Seats(g)
	for opp.Hand.Size() > 0 {
		_, _ = opp.Hand.PopTop()
	}
	land := effectstest.EoHand(opp, "Forest", "Basic Land — Forest")
	creature := effectstest.EoHand(opp, "Bear", "Creature — Bear")
	effectstest.CastCatalogSpell(t, g, "Brain Maggot", "Enchantment Creature — Insect", oldB369MaggotOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if p := effectstest.LatestPickTarget(g, me.ID); p != nil {
		effectstest.PickPlayer(t, g, me.ID, opp.ID)
	}
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.OpenVariant(t, g)
	if err := g.ResolvePendingChoice(pick.ID, me.ID, []uuid.UUID{land}); err == nil {
		t.Fatal("a land card was accepted")
	}
	if err := g.ResolvePendingChoice(pick.ID, me.ID, []uuid.UUID{creature}); err != nil {
		t.Fatalf("resolve the pick: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.ZoneOfCard(g, creature) != game.ZoneExile {
		t.Fatalf("the creature card is in %v, want exile", effectstest.ZoneOfCard(g, creature))
	}
	var maggot uuid.UUID
	for _, card := range g.BattlefieldCardsForEffect() {
		if card.Name == "Brain Maggot" {
			maggot = card.InstanceID
		}
	}
	effectstest.B27Kill(g, maggot)
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	if !opp.Hand.Contains(creature) {
		t.Errorf("the card did not return to its owner's hand (it is in %v)", effectstest.ZoneOfCard(g, creature))
	}
}

func TestOldB369HallOfTriumphPumpsTheChosenColor(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := oldB369Seats(g)
	red := effectstest.B31Push(g, me.ID, "Raging Goblin", "Creature — Goblin", "", "{R}", 1, 1, "R")
	green := effectstest.B31Push(g, me.ID, "Grizzly Bears", "Creature — Bear", "", "{1}{G}", 2, 2, "G")
	theirs := effectstest.B31Push(g, opp.ID, "Their Goblin", "Creature — Goblin", "", "{R}", 1, 1, "R")
	effectstest.PushChosenColorPermanent(t, g, me.ID, "Hall of Triumph", "Legendary Artifact", oldB369HallOracle, "R")
	for id, want := range map[uuid.UUID][2]int{red: {2, 2}, green: {2, 2}, theirs: {1, 1}} {
		if p, tough := effectstest.PtOf(t, g, id); p != want[0] || tough != want[1] {
			t.Errorf("%s is %d/%d, want %d/%d", id, p, tough, want[0], want[1])
		}
	}
}

func TestOldB369QuarryColossusTucksBeneathTheTopPlainsCount(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := oldB369Seats(g)
	for i := 0; i < 5; i++ {
		effectstest.TopOfLibrary(opp, "Filler", "Sorcery", "{1}")
	}
	effectstest.PushLandFor(g, me.ID, "Plains", "Basic Land — Plains")
	effectstest.PushLandFor(g, me.ID, "Plains", "Basic Land — Plains")
	effectstest.PushLandFor(g, me.ID, "Island", "Basic Land — Island")
	victim := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Quarry Colossus", "Creature — Giant", oldB369ColossusOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerPickTargetIfOpen(t, g, victim)
	effectstest.PassPriorityAroundTable(t, g)
	if _, ok := effectstest.BattlefieldCard(g, victim); ok {
		t.Fatal("the creature is still on the battlefield")
	}
	pos := -1
	g.ReadSnapshot(func() {
		n := len(opp.Library.Cards)
		for i, c := range opp.Library.Cards {
			if c.InstanceID == victim {
				pos = n - 1 - i
			}
		}
	})
	if pos != 2 {
		t.Errorf("the creature is %d cards from the top, want 2 (beneath the top two, one per Plains)", pos)
	}
}

func TestOldB369NessianGameWardenLooksAtAsManyAsForests(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := oldB369Seats(g)
	effectstest.PushLandFor(g, me.ID, "Forest", "Basic Land — Forest")
	effectstest.PushLandFor(g, me.ID, "Forest", "Basic Land — Forest")
	deep := effectstest.TopOfLibrary(me, "Deep Bear", "Creature — Bear", "{1}{G}")
	spell := effectstest.TopOfLibrary(me, "Opt", "Instant", "{U}")
	bear := effectstest.TopOfLibrary(me, "Top Bear", "Creature — Bear", "{1}{G}")
	effectstest.CastCatalogSpell(t, g, "Nessian Game Warden", "Creature — Beast", oldB369WardenOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.LatestChooseCardsFor(g, me.ID)
	if pick == nil || pick.ChooseMax != 1 {
		t.Fatalf("no take-one prompt: %+v", pick)
	}
	seen := map[uuid.UUID]bool{}
	for _, id := range pick.ChooseCards {
		seen[id] = true
	}
	if !seen[bear] || seen[spell] || seen[deep] {
		t.Errorf("offered %v: want only the creature card among the top two", pick.ChooseCards)
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{bear}); err != nil {
		t.Fatalf("take the creature: %v", err)
	}
	if !me.Hand.Contains(bear) {
		t.Fatal("the creature card is not in hand")
	}
}

func TestOldB369SpawnOfThraxesDealsDamageEqualToMountains(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := oldB369Seats(g)
	for i := 0; i < 3; i++ {
		effectstest.PushLandFor(g, me.ID, "Mountain", "Basic Land — Mountain")
	}
	effectstest.PushLandFor(g, me.ID, "Forest", "Basic Land — Forest")
	life := opp.Life
	effectstest.CastCatalogSpell(t, g, "Spawn of Thraxes", "Creature — Dragon", oldB369ThraxesOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerPickTargetPlayerIfOpen(t, g, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-3 {
		t.Errorf("life %d → %d, want 3 damage (three Mountains)", life, opp.Life)
	}
}

func TestOldB369SpiteOfMogisCountsInstantsAndSorceriesAndScries(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := oldB369Seats(g)
	g.WithWriteLock(func() { me.Graveyard.Cards = nil })
	for _, tl := range []string{"Instant", "Sorcery", "Creature — Bear", "Instant"} {
		me.Graveyard.PushTop(game.Card{InstanceID: uuid.New(), Name: "G", TypeLine: tl, Owner: me.ID, Controller: me.ID})
	}
	victim := effectstest.B12Creature(g, opp.ID, "Their Giant", "Creature — Giant", 2, 6)
	effectstest.CastCatalogSpell(t, g, "Spite of Mogis", "Sorcery", oldB369SpiteOracle, oldB369Card(victim))
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.DamageOn(t, g, victim); got != 3 {
		t.Errorf("damage = %d, want 3 (two instants and a sorcery in the graveyard)", got)
	}
	if effectstest.ScryChoiceFor(g, me.ID) == nil {
		t.Error("no scry prompt")
	}
}

func TestOldB369PropheticFlamespeakerImpulsesOnEachHit(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := oldB369Seats(g)
	effectstest.AdvanceToMain(t, g)
	id := effectstest.B12Push(g, me.ID, "Prophetic Flamespeaker", "Creature — Human Shaman", oldB369FlamespeakerOracl, 1, 3)
	if !effectstest.HasKeywordOnBattlefield(t, g, id, "double strike") || !effectstest.HasKeywordOnBattlefield(t, g, id, "trample") {
		t.Fatal("missing double strike or trample")
	}
	effectstest.TopOfLibrary(me, "One", "Sorcery", "{1}")
	effectstest.TopOfLibrary(me, "Two", "Sorcery", "{1}")
	effectstest.TopOfLibrary(me, "Three", "Sorcery", "{1}")
	library := len(me.Library.Cards)
	effectstest.DeclareAttack(t, g, opp.ID, id)
	effectstest.AdvanceTo(t, g, game.StepCombatDamage)
	effectstest.PassPriorityAroundTable(t, g)
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	if n := library - len(me.Library.Cards); n != 2 {
		t.Errorf("%d cards exiled, want 2 (double strike deals damage twice)", n)
	}
}

func TestOldB369StarfallHitsAnEnchantmentCreaturesController(t *testing.T) {
	for _, c := range []struct {
		typeLine string
		life     int
	}{{"Enchantment Creature — Nymph", 3}, {"Creature — Bear", 0}} {
		g := effectstest.NewCatalogGame(t)
		_, opp := oldB369Seats(g)
		victim := effectstest.B12Creature(g, opp.ID, "Victim", c.typeLine, 2, 5)
		life := opp.Life
		effectstest.CastCatalogSpell(t, g, "Starfall", "Instant", oldB369StarfallOracle, oldB369Card(victim))
		effectstest.PassPriorityAroundTable(t, g)
		if got := effectstest.DamageOn(t, g, victim); got != 3 {
			t.Errorf("%s: damage = %d, want 3", c.typeLine, got)
		}
		if opp.Life != life-c.life {
			t.Errorf("%s: life %d → %d, want -%d", c.typeLine, life, opp.Life, c.life)
		}
	}
}

func TestOldB369KnowledgeAndPowerDamagesWhenYouScryAndPay(t *testing.T) {
	for _, pay := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		me, opp := oldB369Seats(g)
		effectstest.PushCatalogPermanent(g, me.ID, "Knowledge and Power", "Enchantment", oldB369KnowledgeOracle, false)
		victim := effectstest.B12Creature(g, opp.ID, "Their Giant", "Creature — Giant", 2, 6)
		effectstest.SeedLibrary(me, "A", "B")
		g.WithWriteLock(func() { g.EmitEvent(game.Event{Kind: game.EventScry, Actor: me.ID}) })
		effectstest.PassPriorityAroundTable(t, g)
		effectstest.AnswerPickTargetIfOpen(t, g, victim)
		effectstest.PassPriorityAroundTable(t, g)
		effectstest.B09AddColorless(me, 2)
		effectstest.AnswerPayUnless(t, g, me.ID, pay)
		effectstest.PassPriorityAroundTable(t, g)
		want := 0
		if pay {
			want = 2
		}
		if got := effectstest.DamageOn(t, g, victim); got != want {
			t.Errorf("pay=%v: damage = %d, want %d", pay, got, want)
		}
	}
}

func TestOldB369StrengthFromTheFallenCountsCreatureCardsOnConstellation(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := oldB369Seats(g)
	g.WithWriteLock(func() { me.Graveyard.Cards = nil })
	for _, tl := range []string{"Creature — Bear", "Creature — Elf", "Instant"} {
		me.Graveyard.PushTop(game.Card{InstanceID: uuid.New(), Name: "G", TypeLine: tl, Owner: me.ID, Controller: me.ID})
	}
	bear := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	effectstest.PushCatalogPermanent(g, me.ID, "Strength from the Fallen", "Enchantment", oldB369StrengthOracle, false)
	shrine := effectstest.PushCatalogPermanent(g, me.ID, "Other Shrine", "Enchantment", "", false)
	g.WithWriteLock(func() { g.EmitEvent(game.Event{Kind: game.EventETB, CardID: shrine, Actor: me.ID}) })
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerPickTargetIfOpen(t, g, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.PtOf(t, g, bear); p != 4 || tough != 4 {
		t.Errorf("the Bear is %d/%d, want 4/4 (two creature cards in the graveyard)", p, tough)
	}
}
