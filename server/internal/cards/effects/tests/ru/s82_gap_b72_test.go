package testsru

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_gap_b72_test.go — Ravnica Remastered creatures, batch 72 (#5948). Each
// test plays the real catalog entry through the engine.

const (
	gapB72LoamingOracle     = "cf7352bf-3971-4bb7-88b9-65ab9a219289"
	gapB72BorborygmosOracle = "b7505737-1fc8-4ee8-8b1d-5dde73b6a76d"
	gapB72GuildmageOracle   = "75645b1f-0e12-4a05-b47f-0d2e2616c3ca"
	gapB72MaulerOracle      = "db0b6e35-d7f2-435a-87d7-9e8c36426f4d"
	gapB72SunderOracle      = "38b23fe1-0e76-494a-ad9b-bc22c7d21a5e"
)

func TestGapB72EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		gapB72LoamingOracle:     "Loaming Shaman",
		gapB72BorborygmosOracle: "Borborygmos Enraged",
		gapB72GuildmageOracle:   "Gruul Guildmage",
		gapB72MaulerOracle:      "Scab-Clan Mauler",
		gapB72SunderOracle:      "Sunder Shaman",
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

func TestGapB72LoamingShamanShufflesChosenCardsFromAPlayersGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a, b, kept := uuid.New(), uuid.New(), uuid.New()
	for _, id := range []uuid.UUID{a, b, kept} {
		opp.Graveyard.PushTop(game.Card{InstanceID: id, Name: "Grave " + id.String()[:4], TypeLine: "Sorcery", Owner: opp.ID, Controller: opp.ID})
	}
	libBefore := len(opp.Library.Cards)
	effectstest.AdvanceToMain(t, g)
	effectstest.EnterFromHand(t, g, me.ID, "Loaming Shaman", "Creature — Centaur Shaman", gapB72LoamingOracle)
	effectstest.PickTargetRefs(t, g, me.ID, nil, game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID})
	effectstest.PickTargetRefs(t, g, me.ID, nil,
		game.TargetRef{Kind: game.TargetCard, ID: a, Slot: 1},
		game.TargetRef{Kind: game.TargetCard, ID: b, Slot: 1})
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Graveyard.Contains(a) || opp.Graveyard.Contains(b) {
		t.Error("a chosen card is still in the graveyard")
	}
	if !opp.Graveyard.Contains(kept) {
		t.Error("an unchosen card left the graveyard")
	}
	if got := len(opp.Library.Cards); got != libBefore+2 {
		t.Errorf("library has %d cards, want %d", got, libBefore+2)
	}
}

func TestGapB72ScabClanMaulerNeedsAnOpponentDealtDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	mauler := effectstest.CastHoldCreature(t, g, "Scab-Clan Mauler", "Creature — Human Berserker", gapB72MaulerOracle, 1, 1)
	if p, tg := effectstest.PtOf(t, g, mauler); p != 1 || tg != 1 {
		t.Errorf("with no opponent damaged the Mauler is %d/%d, want 1/1", p, tg)
	}
	if !effectstest.HasEffectiveKeyword(t, g, mauler, "trample") {
		t.Error("no trample")
	}
	g2 := effectstest.NewCatalogGame(t)
	_, opp := effectstest.GptB1Seats(g2)
	effectstest.GptB1DamagePlayer(g2, opp.ID, 1)
	m2 := effectstest.CastHoldCreature(t, g2, "Scab-Clan Mauler", "Creature — Human Berserker", gapB72MaulerOracle, 1, 1)
	if n := effectstest.CounterCount(g2, m2, game.CounterPlusOne); n != 2 {
		t.Errorf("%d +1/+1 counters, want 2", n)
	}
	if p, tg := effectstest.PtOf(t, g2, m2); p != 3 || tg != 3 {
		t.Errorf("after bloodthirst the Mauler is %d/%d, want 3/3", p, tg)
	}
}

func TestGapB72ScabClanMaulerIgnoresDamageToItsController(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := effectstest.GptB1Seats(g)
	effectstest.GptB1DamagePlayer(g, me.ID, 3)
	m := effectstest.CastHoldCreature(t, g, "Scab-Clan Mauler", "Creature — Human Berserker", gapB72MaulerOracle, 1, 1)
	if n := effectstest.CounterCount(g, m, game.CounterPlusOne); n != 0 {
		t.Errorf("damage to its own controller put %d counters on the Mauler", n)
	}
}

func TestGapB72SunderShamanRefusesASecondBlocker(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	shaman := effectstest.B12Push(g, me.ID, "Sunder Shaman", "Creature — Giant Shaman", gapB72SunderOracle, 5, 5)
	a := effectstest.B12Creature(g, opp.ID, "Bear A", "Creature — Bear", 2, 2)
	b := effectstest.B12Creature(g, opp.ID, "Bear B", "Creature — Bear", 2, 2)
	effectstest.BrAttack(t, g, shaman)
	effectstest.BrRefusal(t, g.DeclareBlockers([]game.BlockDeclaration{
		{Blocker: a, Attacker: shaman}, {Blocker: b, Attacker: shaman},
	}), game.BlockReasonTooManyBlockers)
	if err := g.DeclareBlockers([]game.BlockDeclaration{{Blocker: a, Attacker: shaman}}); err != nil {
		t.Errorf("a single blocker was refused: %v", err)
	}
}

func TestGapB72SunderShamanDestroysAnArtifactOrEnchantmentOfTheDamagedPlayer(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	shaman := effectstest.B12Push(g, me.ID, "Sunder Shaman", "Creature — Giant Shaman", gapB72SunderOracle, 5, 5)
	theirs := effectstest.B12Permanent(g, opp.ID, "Their Rock", "Artifact")
	theirsAura := effectstest.B12Permanent(g, opp.ID, "Their Ward", "Enchantment")
	third := effectstest.B12Permanent(g, g.Seats[2].ID, "Third's Rock", "Artifact")
	mine := effectstest.B12Permanent(g, me.ID, "My Rock", "Artifact")
	creature := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	effectstest.AttackWith(t, g, opp.ID, shaman)
	effectstest.AdvanceTo(t, g, game.StepCombatDamage)
	effectstest.B04WaitForPick(t, g, me.ID)
	p := effectstest.LatestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("no target prompt after combat damage")
	}
	for _, id := range []uuid.UUID{theirs, theirsAura} {
		if !effectstest.HasID(p.PickTargetCards, id) {
			t.Errorf("the damaged player's permanent %s was not offered", id)
		}
	}
	for _, id := range []uuid.UUID{third, mine, creature} {
		if effectstest.HasID(p.PickTargetCards, id) {
			t.Errorf("a permanent that is not the damaged player's artifact or enchantment was offered: %s", id)
		}
	}
	effectstest.PickCard(t, g, me.ID, theirsAura)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.FindBattlefieldCardForTest(g, theirsAura) != nil {
		t.Error("the enchantment survived")
	}
	if effectstest.FindBattlefieldCardForTest(g, theirs) == nil {
		t.Error("the other artifact was destroyed too")
	}
}

func TestGapB72GruulGuildmagePumpsATargetCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	mage := effectstest.B21Push(g, me.ID, "Gruul Guildmage", "Creature — Human Shaman", gapB72GuildmageOracle, 2, 2, "R", "G")
	effectstest.AdvanceToMain(t, g)
	effectstest.FillPoolColored(me, "G", 4)
	if err := g.ActivateCatalogAbility(me.ID, mage, 1, game.ActivateAbilityParams{Strict: true, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: mage}}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p, tg := effectstest.EffectivePower(t, g, mage), effectstest.EffectiveToughness(t, g, mage); p != 4 || tg != 4 {
		t.Errorf("the Guildmage is %d/%d, want 4/4", p, tg)
	}
}

func TestGapB72GruulGuildmageSacrificesALandToHitAPlayerOrPlaneswalker(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mage := effectstest.B21Push(g, me.ID, "Gruul Guildmage", "Creature — Human Shaman", gapB72GuildmageOracle, 2, 2, "R", "G")
	land := effectstest.B12Permanent(g, me.ID, "Forest", "Basic Land — Forest")
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	walker := effectstest.ApaPush(g, opp.ID, opp.ID, game.Card{Name: "Their Walker", TypeLine: "Planeswalker — Test", Counters: map[string]int{"loyalty": 5}})
	effectstest.AdvanceToMain(t, g)
	effectstest.FillPoolColored(me, "R", 8)
	hit := func(sac uuid.UUID, target game.TargetRef) error {
		return g.ActivateCatalogAbility(me.ID, mage, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{sac}, Targets: []game.TargetRef{target}})
	}
	if err := hit(bear, game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID}); err == nil {
		t.Error("a creature paid the land sacrifice")
	}
	if err := hit(land, game.TargetRef{Kind: game.TargetCard, ID: bear}); err == nil {
		t.Error("a creature was a legal target")
	}
	life := opp.Life
	if err := hit(land, game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID}); err != nil {
		t.Fatalf("activate at a player: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-2 {
		t.Errorf("life %d, want %d", opp.Life, life-2)
	}
	if effectstest.FindBattlefieldCardForTest(g, land) != nil {
		t.Error("the land was not sacrificed")
	}
	land2 := effectstest.B12Permanent(g, me.ID, "Forest", "Basic Land — Forest")
	if err := hit(land2, game.TargetRef{Kind: game.TargetCard, ID: walker}); err != nil {
		t.Fatalf("activate at a planeswalker: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CounterCount(g, walker, "loyalty"); n != 3 {
		t.Errorf("the planeswalker has %d loyalty, want 3", n)
	}
}

func TestGapB72BorborygmosDiscardsALandForThreeDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	boro := effectstest.B27Push(g, me.ID, "Borborygmos Enraged", "Legendary Creature — Cyclops", gapB72BorborygmosOracle, "{4}{R}{R}{G}{G}", 7, 6, "R", "G")
	land := effectstest.HandCardForTest(me, "Mountain", "Basic Land — Mountain", "")
	spell := effectstest.HandCardForTest(me, "Spell", "Sorcery", "")
	life := opp.Life
	effectstest.AdvanceToMain(t, g)
	for _, bad := range []uuid.UUID{spell} {
		if err := g.ActivateCatalogAbility(me.ID, boro, 0, game.ActivateAbilityParams{Strict: true, DiscardIDs: []uuid.UUID{bad}, Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}}); err == nil {
			t.Error("a nonland card paid the discard")
		}
	}
	if err := g.ActivateCatalogAbility(me.ID, boro, 0, game.ActivateAbilityParams{Strict: true, DiscardIDs: []uuid.UUID{land}, Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-3 {
		t.Errorf("life %d, want %d", opp.Life, life-3)
	}
	if !me.Graveyard.Contains(land) {
		t.Error("the land is not in the graveyard")
	}
}

func TestGapB72BorborygmosRevealsThreeAndTakesTheLands(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	boro := effectstest.B27Push(g, me.ID, "Borborygmos Enraged", "Legendary Creature — Cyclops", gapB72BorborygmosOracle, "{4}{R}{R}{G}{G}", 7, 6, "R", "G")
	top := effectstest.HornTopFive(me, []struct{ Name, TypeLine string }{
		{"Forest", "Basic Land — Forest"},
		{"Spell", "Sorcery"},
		{"Mountain", "Basic Land — Mountain"},
		{"Deep", "Basic Land — Forest"},
	})
	effectstest.AttackWith(t, g, opp.ID, boro)
	effectstest.AdvanceTo(t, g, game.StepCombatDamage)
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{top[0], top[2]} {
		if !me.Hand.Contains(id) {
			t.Errorf("land %s is not in hand", id)
		}
	}
	if !me.Graveyard.Contains(top[1]) {
		t.Error("the nonland card should be in the graveyard")
	}
	if z := g.FindCardZoneForEffect(top[3]); z == nil || z.Kind != game.ZoneLibrary {
		t.Errorf("the fourth card is in %v, want the library", z)
	}
}
