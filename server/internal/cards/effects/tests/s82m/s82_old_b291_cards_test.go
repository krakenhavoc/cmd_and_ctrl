package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b291_cards_test.go — Shadows over Innistrad, batch 291 (#5543).

const (
	soiCreepingDreadOracle = "3bb16720-18dc-44b7-adb0-c9b7adda9ebe"
	soiCallBloodlineOracle = "8e90300b-c0da-4484-b065-e2fafa4e1b15"
	soiDrogskolOracle      = "a8bac5a8-5a60-4f7a-a4d6-cb922da92472"
	soiIndignationOracle   = "c724dd2e-c62d-4a20-a1b9-54264d0b811a"
	soiEvangelOracle       = "e2fc5069-374b-4feb-acd9-976ddf90e504"
	soiSlayersPlateOracle  = "05004b4a-2e2c-4a88-a994-8576684fb64a"
	soiSkeletonKeyOracle   = "328834f6-1f49-417a-a2d6-d430b8dca177"
	soiCenserOracle        = "330f2e60-810d-4fe0-9a35-b3f2b29e3e2a"
	soiMurderersAxeOracle  = "f0ea953a-10c3-4ec7-8f03-f7aed75dbb94"
)

func TestSoiB291EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		soiCreepingDreadOracle:  "Creeping Dread",
		soiCallBloodlineOracle:  "Call the Bloodline",
		soiDrogskolOracle:       "Drogskol Cavalry",
		soiIndignationOracle:    "Liliana's Indignation",
		soiEvangelOracle:        "Pious Evangel",
		soiEvangelOracle + "#1": "Wayward Disciple",
		soiSlayersPlateOracle:   "Slayer's Plate",
		soiSkeletonKeyOracle:    "Skeleton Key",
		soiCenserOracle:         "True-Faith Censer",
		soiMurderersAxeOracle:   "Murderer's Axe",
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

func soiB291Hands(g *game.Game) {
	for _, p := range g.Seats {
		p.Hand.Cards = p.Hand.Cards[:0]
	}
}

func TestSoiB291CreepingDreadDrainsOpponentsSharingATypeWithYourDiscard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	effectstest.B12Push(g, me.ID, "Creeping Dread", "Enchantment", soiCreepingDreadOracle, 0, 0)
	soiB291Hands(g)
	mine := effectstest.HandCardForTest(me, "My Bear", "Creature — Bear", "")
	c1 := effectstest.HandCardForTest(g.Seats[1], "Their Bear", "Creature — Bear", "")
	c2 := effectstest.HandCardForTest(g.Seats[2], "Forest", "Basic Land — Forest", "")
	c3 := effectstest.HandCardForTest(g.Seats[3], "Rock", "Artifact Creature — Golem", "")
	before := [4]int{me.Life, g.Seats[1].Life, g.Seats[2].Life, g.Seats[3].Life}
	effectstest.AdvanceToUpkeepOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerDiscard(t, g, me.ID, mine)
	effectstest.AnswerDiscard(t, g, g.Seats[1].ID, c1)
	effectstest.AnswerDiscard(t, g, g.Seats[2].ID, c2)
	effectstest.AnswerDiscard(t, g, g.Seats[3].ID, c3)
	want := [4]int{before[0], before[1] - 3, before[2], before[3] - 3}
	got := [4]int{me.Life, g.Seats[1].Life, g.Seats[2].Life, g.Seats[3].Life}
	if got != want {
		t.Errorf("life %v, want %v (creature sharers lose 3, the land discarder does not)", got, want)
	}
}

func TestSoiB291CreepingDreadDrainsNobodyWhenYouHaveNothingToDiscard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	effectstest.B12Push(g, me.ID, "Creeping Dread", "Enchantment", soiCreepingDreadOracle, 0, 0)
	soiB291Hands(g)
	c1 := effectstest.HandCardForTest(g.Seats[1], "Their Bear", "Creature — Bear", "")
	life := g.Seats[1].Life
	effectstest.AdvanceToUpkeepOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerDiscard(t, g, g.Seats[1].ID, c1)
	if g.Seats[1].Life != life {
		t.Errorf("life %d → %d: nobody shares a type with a discard that never happened", life, g.Seats[1].Life)
	}
}

func TestSoiB291CallTheBloodlineMakesOneLifelinkKnightEachTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	cb := effectstest.B12Push(g, me.ID, "Call the Bloodline", "Enchantment", soiCallBloodlineOracle, 0, 0)
	d1 := effectstest.HandCardForTest(me, "Pitch One", "Sorcery", "")
	d2 := effectstest.HandCardForTest(me, "Pitch Two", "Sorcery", "")
	effectstest.FillPool(me, 2)
	if err := g.ActivateCatalogAbility(me.ID, cb, 0, game.ActivateAbilityParams{DiscardIDs: []uuid.UUID{d1}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(d1) {
		t.Error("the discard was not paid")
	}
	if n := effectstest.TokensNamedControlledBy(g, "Vampire Knight", me.ID); n != 1 {
		t.Fatalf("%d Vampire Knight tokens, want 1", n)
	}
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Vampire Knight" && !effectstest.HasKeywordOnBattlefield(t, g, c.InstanceID, "lifelink") {
			t.Error("the token lacks lifelink")
		}
	}
	if err := g.ActivateCatalogAbility(me.ID, cb, 0, game.ActivateAbilityParams{DiscardIDs: []uuid.UUID{d2}}); err == nil {
		t.Error("a second activation in the same turn was allowed")
	}
}

func TestSoiB291DrogskolCavalryGainsLifeForAnotherSpiritOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	cav := effectstest.B12Push(g, me.ID, "Drogskol Cavalry", "Creature — Spirit Knight", soiDrogskolOracle, 4, 4)
	life := me.Life
	effectstest.CastCreatureCard(t, g, "Gruff Bear", "Creature — Bear", "", 2, 2)
	if me.Life != life {
		t.Errorf("a Bear gained life: %d → %d", life, me.Life)
	}
	effectstest.CastCreatureCard(t, g, "Lost Spirit", "Creature — Spirit", "", 1, 1)
	if me.Life != life+2 {
		t.Errorf("life %d → %d, want +2 for another Spirit", life, me.Life)
	}
	effectstest.FillPoolColored(me, "W", 1)
	effectstest.FillPool(me, 3)
	if err := g.ActivateCatalogAbility(me.ID, cav, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.TokensNamedControlledBy(g, "Spirit", me.ID); n != 1 {
		t.Errorf("%d Spirit tokens, want 1", n)
	}
	if me.Life != life+4 {
		t.Errorf("life %d → %d, want +4 once the token entered", life, me.Life)
	}
}

func TestSoiB291DrogskolCavalryDoesNotTriggerForItself(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	life := me.Life
	effectstest.CastCreatureCard(t, g, "Drogskol Cavalry", "Creature — Spirit Knight", soiDrogskolOracle, 4, 4)
	if me.Life != life {
		t.Errorf("the Cavalry gained life for its own entry: %d → %d", life, me.Life)
	}
}

func TestSoiB291LilianasIndignationDrainsPerMilledCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	me.Library.Cards = me.Library.Cards[:0]
	for _, c := range []struct{ name, tl string }{
		{"Deep Bear", "Creature — Bear"}, {"Mill Land", "Land"}, {"Mill Elf", "Creature — Elf"}, {"Top Wolf", "Creature — Wolf"},
	} {
		me.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: c.name, TypeLine: c.tl, Owner: me.ID, Controller: me.ID})
	}
	effectstest.FillPoolColored(me, "B", 1)
	effectstest.FillPool(me, 3)
	life := opp.Life
	effectstest.B19CastXModal(t, g, "Liliana's Indignation", "Sorcery", soiIndignationOracle, "{X}{B}", 3, nil, effectstest.Pr6Player(opp.ID))
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-4 {
		t.Errorf("life %d → %d, want -4: the top three are a Wolf, an Elf and a Land, so two creature cards", life, opp.Life)
	}
}

func TestSoiB291PiousEvangelGainsLifeWhenItOrAnotherCreatureEnters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	life := me.Life
	effectstest.CastCreatureCard(t, g, "Pious Evangel", "Creature — Human Cleric", soiEvangelOracle, 2, 2)
	if me.Life != life+1 {
		t.Fatalf("life %d → %d, want +1 for its own entry", life, me.Life)
	}
	effectstest.CastCreatureCard(t, g, "Gruff Bear", "Creature — Bear", "", 2, 2)
	if me.Life != life+2 {
		t.Errorf("life %d → %d, want +2 after another creature entered", life, me.Life)
	}
}

func soiB291Evangel(owner uuid.UUID, back bool) game.Card {
	c := game.Card{
		InstanceID: uuid.New(), OracleID: soiEvangelOracle, Layout: game.LayoutTransform,
		Owner: owner, Controller: owner,
		Faces: []game.Face{
			{Name: "Pious Evangel", TypeLine: "Creature — Human Cleric", ManaCost: "{2}{W}", Colors: []string{"W"}, Power: 2, Toughness: 2},
			{Name: "Wayward Disciple", TypeLine: "Creature — Human Cleric", Colors: []string{"W"}, Power: 2, Toughness: 4},
		},
	}
	if back {
		c.SetFace(1)
	} else {
		c.SetFace(0)
	}
	return c
}

func TestSoiB291PiousEvangelTransformsBySacrificingAnotherPermanent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	ev := effectstest.PushBattlefieldCardWithTimestamp(g, soiB291Evangel(me.ID, false))
	fodder := effectstest.B12Permanent(g, me.ID, "Trinket", "Artifact")
	effectstest.FillPool(me, 2)
	// Sacrificing itself is not a legal payment.
	if err := g.ActivateCatalogAbility(me.ID, ev, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{ev}}); err == nil {
		t.Fatal("the Evangel sacrificed itself to its own ability")
	}
	if err := g.ActivateCatalogAbility(me.ID, ev, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{fodder}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(fodder) {
		t.Error("the other permanent was not sacrificed")
	}
	if c := effectstest.RvBattlefieldCard(g, ev); c == nil || c.Name != "Wayward Disciple" {
		t.Fatalf("the Evangel did not transform: %+v", c)
	}
}

func TestSoiB291WaywardDiscipleDrainsWhenACreatureYouControlDies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	effectstest.PushBattlefieldCardWithTimestamp(g, soiB291Evangel(me.ID, true))
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	myLife, oppLife := me.Life, opp.Life
	effectstest.KillCreature(t, g, opp.ID, theirs)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != myLife || opp.Life != oppLife {
		t.Fatalf("an opponent's creature died and life moved: %d, %d", me.Life, opp.Life)
	}
	effectstest.KillCreature(t, g, me.ID, bear)
	effectstest.AnswerPickTargetPlayer(t, g, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != oppLife-1 || me.Life != myLife+1 {
		t.Errorf("life me %d → %d, opponent %d → %d, want +1 and -1", myLife, me.Life, oppLife, opp.Life)
	}
}

func TestSoiB291SlayersPlatePumpsAndLeavesASpiritWhenAHumanDies(t *testing.T) {
	for _, tc := range []struct {
		typeLine string
		spirits  int
	}{{"Creature — Human Soldier", 1}, {"Creature — Bear", 0}} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		effectstest.AdvanceToMain(t, g)
		host := effectstest.B12Creature(g, me.ID, "Host", tc.typeLine, 2, 2)
		effectstest.SomB1Equip(t, g, me.ID, "Slayer's Plate", soiSlayersPlateOracle, host)
		if p, tough := effectstest.EffectivePTOf(t, g, host); p != 6 || tough != 4 {
			t.Errorf("%s: %d/%d, want 6/4", tc.typeLine, p, tough)
		}
		effectstest.KillCreature(t, g, me.ID, host)
		effectstest.PassPriorityAroundTable(t, g)
		if n := effectstest.TokensNamedControlledBy(g, "Spirit", me.ID); n != tc.spirits {
			t.Errorf("%s: %d Spirit tokens, want %d", tc.typeLine, n, tc.spirits)
		}
	}
}

func TestSoiB291SkeletonKeyGivesSkulkAndLoots(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	host := effectstest.B12Creature(g, me.ID, "Host", "Creature — Rogue", 2, 2)
	effectstest.SomB1Equip(t, g, me.ID, "Skeleton Key", soiSkeletonKeyOracle, host)
	if !effectstest.HasKeywordOnBattlefield(t, g, host, "skulk") {
		t.Error("the equipped creature lacks skulk")
	}
	soiB291Hands(g)
	keep := effectstest.HandCardForTest(me, "Kept Card", "Sorcery", "")
	effectstest.DealCombatDamageToPlayer(g, host, opp.ID, 2)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerMayChoice(t, g, me.ID, true)
	if me.Hand.Size() != 2 {
		t.Fatalf("hand %d, want the kept card plus the drawn one", me.Hand.Size())
	}
	effectstest.AnswerDiscard(t, g, me.ID, keep)
	if me.Hand.Size() != 1 || !me.Graveyard.Contains(keep) {
		t.Errorf("hand %d after the discard, kept card in graveyard %v", me.Hand.Size(), me.Graveyard.Contains(keep))
	}
}

func TestSoiB291TrueFaithCenserPumpsHumansMore(t *testing.T) {
	for _, tc := range []struct {
		typeLine string
		power    int
	}{{"Creature — Human Soldier", 4}, {"Creature — Bear", 3}} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		effectstest.AdvanceToMain(t, g)
		host := effectstest.B12Creature(g, me.ID, "Host", tc.typeLine, 2, 2)
		effectstest.SomB1Equip(t, g, me.ID, "True-Faith Censer", soiCenserOracle, host)
		if p, tough := effectstest.EffectivePTOf(t, g, host); p != tc.power || tough != 3 {
			t.Errorf("%s: %d/%d, want %d/3", tc.typeLine, p, tough, tc.power)
		}
		if !effectstest.HasKeywordOnBattlefield(t, g, host, "vigilance") {
			t.Errorf("%s: no vigilance", tc.typeLine)
		}
	}
}

func TestSoiB291MurderersAxeEquipsByDiscarding(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	host := effectstest.B12Creature(g, me.ID, "Host", "Creature — Bear", 2, 2)
	axe := effectstest.B12Push(g, me.ID, "Murderer's Axe", "Artifact — Equipment", soiMurderersAxeOracle, 0, 0)
	pitch := effectstest.HandCardForTest(me, "Pitch", "Sorcery", "")
	target := []game.TargetRef{{Kind: game.TargetCard, ID: host}}
	if err := g.ActivateCatalogAbility(me.ID, axe, 0, game.ActivateAbilityParams{Targets: target}); err == nil {
		t.Fatal("equipped without discarding a card")
	}
	if err := g.ActivateCatalogAbility(me.ID, axe, 0, game.ActivateAbilityParams{Targets: target, DiscardIDs: []uuid.UUID{pitch}}); err != nil {
		t.Fatalf("equip: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(pitch) {
		t.Error("the discard was not paid")
	}
	if p, tough := effectstest.EffectivePTOf(t, g, host); p != 4 || tough != 4 {
		t.Errorf("%d/%d, want 4/4", p, tough)
	}
}
