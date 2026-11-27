package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b672_cards_test.go — Prophecy: Removal, batch 672 (#5617). Each test
// plays the real catalog entry and asserts what its text says.

const (
	b672MineBearer = "19fba3c8-3363-4f2f-a2fb-01e59a77447c"
	b672Fracture   = "3495d83a-b103-42be-8708-9ce971b352bd"
	b672Devastate  = "90b3e2df-182d-4dad-9934-597a048fefc6"
	b672Verse      = "c70ae479-0bd1-481c-885f-16bfa096cf58"
	b672Mageta     = "cd170b1c-879b-414e-9f0e-5e7fb099e7bb"
	b672Arsonist   = "d32f973e-38db-4f74-aec6-258d25c1fb7d"
)

func TestB672CardsAreRegisteredAndWhole(t *testing.T) {
	for oracle, name := range map[string]string{
		b672MineBearer: "Mine Bearer", b672Fracture: "Aura Fracture", b672Devastate: "Devastate",
		b672Verse: "Calming Verse", b672Mageta: "Mageta the Lion", b672Arsonist: "Keldon Arsonist",
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

func b672Table(t *testing.T) (*game.Game, *game.Player, *game.Player) {
	t.Helper()
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.SeatsForCombat(g)
	effectstest.AdvanceToMain(t, g)
	return g, me, opp
}

func TestB672MineBearerSacrificesToDestroyAnAttacker(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bearer := effectstest.B12Push(g, opp.ID, "Mine Bearer", "Creature — Human Soldier", b672MineBearer, 1, 1)
	attacker := effectstest.ReqCreature(g, me.ID, "Ogre", "", 4, 4)
	idle := effectstest.ReqCreature(g, me.ID, "Idle", "", 2, 2)
	effectstest.DeclareAttack(t, g, opp.ID, attacker)
	g.Turn.PriorityHolder = 1
	if err := g.ActivateCatalogAbility(opp.ID, bearer, 0, game.ActivateAbilityParams{Strict: true,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: idle}}}); err == nil {
		t.Fatal("a creature that is not attacking was a legal target")
	}
	if err := g.ActivateCatalogAbility(opp.ID, bearer, 0, game.ActivateAbilityParams{Strict: true,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: attacker}}}); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	g.RunStateChecksForTest()
	if effectstest.BattlefieldHas(g, attacker) {
		t.Error("the attacker survived")
	}
	if effectstest.BattlefieldHas(g, bearer) {
		t.Error("the Bearer was not sacrificed")
	}
	if !effectstest.BattlefieldHas(g, idle) {
		t.Error("the idle creature died")
	}
}

func TestB672AuraFractureSacrificesALandToDestroyAnEnchantment(t *testing.T) {
	g, me, opp := b672Table(t)
	fracture := effectstest.PushCatalogPermanent(g, me.ID, "Aura Fracture", "Enchantment", b672Fracture, false)
	land := effectstest.GaPut(g, me.ID, "Plains", "Basic Land — Plains", "")
	spare := effectstest.GaPut(g, me.ID, "Plains", "Basic Land — Plains", "")
	ench := effectstest.B12Permanent(g, opp.ID, "Their Enchantment", "Enchantment")
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	if err := g.ActivateCatalogAbility(me.ID, fracture, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{land},
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}}}); err == nil {
		t.Fatal("a creature was a legal target")
	}
	if err := g.ActivateCatalogAbility(me.ID, fracture, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{bear},
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: ench}}}); err == nil {
		t.Fatal("a nonland permanent paid the land sacrifice")
	}
	if err := g.ActivateCatalogAbility(me.ID, fracture, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{land},
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: ench}}}); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.BattlefieldHas(g, ench) || effectstest.BattlefieldHas(g, land) {
		t.Error("want the enchantment destroyed and the land sacrificed")
	}
	if !effectstest.BattlefieldHas(g, spare) || !effectstest.BattlefieldHas(g, fracture) {
		t.Error("the other land or the Fracture left")
	}
}

func TestB672KeldonArsonistSacrificesTwoLandsToDestroyALand(t *testing.T) {
	g, me, opp := b672Table(t)
	arsonist := effectstest.B12Push(g, me.ID, "Keldon Arsonist", "Creature — Human Soldier", b672Arsonist, 1, 1)
	l1 := effectstest.GaPut(g, me.ID, "Mountain", "Basic Land — Mountain", "")
	l2 := effectstest.GaPut(g, me.ID, "Mountain", "Basic Land — Mountain", "")
	l3 := effectstest.GaPut(g, me.ID, "Mountain", "Basic Land — Mountain", "")
	theirs := effectstest.GaPut(g, opp.ID, "Island", "Basic Land — Island", "")
	target := []game.TargetRef{{Kind: game.TargetCard, ID: theirs}}
	effectstest.FloatForTest(g, me, "R")
	if err := g.ActivateCatalogAbility(me.ID, arsonist, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{l1}, Targets: target}); err == nil {
		t.Fatal("one land paid for a two-land sacrifice")
	}
	if err := g.ActivateCatalogAbility(me.ID, arsonist, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{l1, l2}, Targets: target}); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.BattlefieldHas(g, theirs) || effectstest.BattlefieldHas(g, l1) || effectstest.BattlefieldHas(g, l2) {
		t.Error("want their land destroyed and both of mine sacrificed")
	}
	if !effectstest.BattlefieldHas(g, l3) {
		t.Error("the third land left")
	}
}

func TestB672DevastateDestroysALandThenPingsEverything(t *testing.T) {
	g, me, opp := b672Table(t)
	land := effectstest.GaPut(g, opp.ID, "Island", "Basic Land — Island", "")
	small := effectstest.B12Creature(g, opp.ID, "Small", "Creature — Bear", 2, 1)
	big := effectstest.B12Creature(g, me.ID, "Big", "Creature — Bear", 2, 3)
	myLife, theirLife := me.Life, opp.Life
	effectstest.CastCatalogSpell(t, g, "Devastate", "Sorcery", b672Devastate, []game.TargetRef{{Kind: game.TargetCard, ID: land}})
	effectstest.PassPriorityAroundTable(t, g)
	g.RunStateChecksForTest()
	if effectstest.BattlefieldHas(g, land) {
		t.Error("the land survived")
	}
	if me.Life != myLife-1 || opp.Life != theirLife-1 {
		t.Errorf("life %d/%d, want each down 1", me.Life, opp.Life)
	}
	if effectstest.BattlefieldHas(g, small) || !effectstest.BattlefieldHas(g, big) {
		t.Error("want the 2/1 dead and the 2/3 alive")
	}
}

func TestB672DevastateCantTargetANonland(t *testing.T) {
	g, _, opp := b672Table(t)
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	if err := effectstest.CastCatalogSpellErrStrict(t, g, "Devastate", "Sorcery", b672Devastate, []game.TargetRef{{Kind: game.TargetCard, ID: bear}}); err == nil {
		t.Error("Devastate targeted a creature")
	}
}

func TestB672CalmingVerseDestroysTheirsThenYoursOnlyWithAnUntappedLand(t *testing.T) {
	for _, untapped := range []bool{true, false} {
		g, me, opp := b672Table(t)
		land := effectstest.GaPut(g, me.ID, "Forest", "Basic Land — Forest", "")
		if !untapped {
			b672Tap(g, land)
		}
		mine := effectstest.B12Permanent(g, me.ID, "My Enchantment", "Enchantment")
		theirs := effectstest.B12Permanent(g, opp.ID, "Their Enchantment", "Enchantment")
		effectstest.CastCatalogSpell(t, g, "Calming Verse", "Sorcery", b672Verse, nil)
		effectstest.PassPriorityAroundTable(t, g)
		if effectstest.BattlefieldHas(g, theirs) {
			t.Errorf("untapped=%v: their enchantment survived", untapped)
		}
		if got := !effectstest.BattlefieldHas(g, mine); got != untapped {
			t.Errorf("untapped=%v: my enchantment destroyed=%v", untapped, got)
		}
	}
}

func b672Tap(g *game.Game, id uuid.UUID) {
	g.WithWriteLock(func() { effectstest.FindBattlefieldCardForTest(g, id).Tapped = true })
}

func TestB672MagetaDestroysEveryOtherCreatureEvenARegeneratingOne(t *testing.T) {
	g, me, opp := b672Table(t)
	mageta := effectstest.B12Push(g, me.ID, "Mageta the Lion", "Legendary Creature — Human Spellshaper", b672Mageta, 3, 3)
	mine := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	shielded := effectstest.B12Creature(g, opp.ID, "Shielded", "Creature — Bear", 2, 2)
	effectstest.P1Regenerate(t, g, shielded)
	effectstest.GiveHandCards(me, 3)
	hand := len(me.Hand.Cards)
	cards := []uuid.UUID{me.Hand.Cards[0].InstanceID, me.Hand.Cards[1].InstanceID}
	effectstest.FloatForTest(g, me, "WWWW")
	if err := g.ActivateCatalogAbility(me.ID, mageta, 0, game.ActivateAbilityParams{Strict: true, DiscardIDs: cards[:1]}); err == nil {
		t.Fatal("one discard paid for a two-discard cost")
	}
	if err := g.ActivateCatalogAbility(me.ID, mageta, 0, game.ActivateAbilityParams{Strict: true, DiscardIDs: cards}); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	g.RunStateChecksForTest()
	if !effectstest.BattlefieldHas(g, mageta) {
		t.Error("Mageta was destroyed")
	}
	for name, id := range map[string]uuid.UUID{"mine": mine, "theirs": theirs, "shielded": shielded} {
		if effectstest.BattlefieldHas(g, id) {
			t.Errorf("%s survived", name)
		}
	}
	if len(me.Hand.Cards) != hand-2 || !effectstest.PermanentTapped(g, mageta) {
		t.Errorf("hand %d (want %d), tapped %v", len(me.Hand.Cards), hand-2, effectstest.PermanentTapped(g, mageta))
	}
}
