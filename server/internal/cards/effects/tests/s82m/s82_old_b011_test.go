package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const (
	oldB11Prossh   = "868882d2-ed4e-4171-a17c-478a341080fb"
	oldB11Shaddam  = "9726cf0e-5fd2-4699-8192-2fafa2b36963"
	oldB11Phelddag = "97f3afdb-dcee-44da-b210-c8d20030baff"
	oldB11Sonic    = "b07a7b00-9394-48a9-a597-17f391495909"
	oldB11Pharika  = "b7d05d00-1e69-4d11-bff0-c1ca2026aad6"
	oldB11Revel    = "cc3fcd26-c088-4335-9a35-0e11019e3d5b"
	oldB11Hansk    = "d818f863-17ee-4387-94bf-d7e6e8ab590c"
)

func TestOldB11CardsAreFullWithNoCaveats(t *testing.T) {
	want := map[string]string{
		oldB11Prossh: "Prossh, Skyraider of Kher", oldB11Shaddam: "Padishah Emperor Shaddam IV",
		oldB11Phelddag: "Questing Phelddagrif", oldB11Sonic: "Sonic the Hedgehog",
		oldB11Pharika: "Pharika, God of Affliction", oldB11Revel: "Revel in Riches", oldB11Hansk: "Hansk, Slayer Zealot",
	}
	for oracle, name := range want {
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

func TestOldB11ProsshMakesAKoboldPerManaSpent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.ToMain(t, g)
	id := uuid.New()
	me.Hand.PushTop(game.Card{
		InstanceID: id, Name: "Prossh, Skyraider of Kher", TypeLine: "Legendary Creature — Dragon", OracleID: oldB11Prossh,
		ManaCost: "{3}{B}{R}{G}", Colors: []string{"B", "R", "G"}, Power: 5, Toughness: 5, Owner: me.ID, Controller: me.ID,
	})
	effectstest.FloatForTest(g, me, "BRGBRG")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Kobolds of Kher Keep" && c.Controller == me.ID {
			n++
			if c.Power != 0 || c.Toughness != 1 {
				t.Errorf("Kobold %d/%d, want 0/1", c.Power, c.Toughness)
			}
		}
	}
	if n != 6 {
		t.Errorf("Kobolds: %d, want 6 (the mana spent)", n)
	}
}

func TestOldB11ProsshSacrificesAnotherCreatureForPlusOnePower(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.ToMain(t, g)
	prossh := effectstest.B12Push(g, me.ID, "Prossh, Skyraider of Kher", "Legendary Creature — Dragon", oldB11Prossh, 5, 5)
	fodder := effectstest.B12Creature(g, me.ID, "Kobold", "Creature — Kobold", 0, 1)
	if err := g.ActivateCatalogAbility(me.ID, prossh, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{fodder}}); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(fodder) {
		t.Error("the creature was not sacrificed")
	}
	if c := effectstest.B12Card(t, g, prossh); c.Effective().Power != 6 {
		t.Errorf("Prossh's power %d, want 6", c.Effective().Power)
	}
	if err := g.ActivateCatalogAbility(me.ID, prossh, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{prossh}}); err == nil {
		t.Error("Prossh sacrificed himself to his own ability")
	}
}

func TestOldB11ShaddamBecomesMonarchAndPaysLifeForTreasure(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.ToMain(t, g)
	if _, err := effectstest.Msc17Cast(t, g, me, game.Card{
		Name: "Padishah Emperor Shaddam IV", TypeLine: "Legendary Creature — Human Noble", OracleID: oldB11Shaddam, Power: 2, Toughness: 4,
	}, game.CastSpellParams{}); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Monarch != me.ID {
		t.Fatalf("monarch %v, want the controller", g.Monarch)
	}
	// A nontoken creature dies: pay 1 life, get a Treasure.
	victim := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	life := me.Life
	g.WithWriteLock(func() { _ = g.SacrificePermanentForEffect(victim) })
	effectstest.PassPriorityAroundTable(t, g)
	prompt := effectstest.LatestTriggerPrompt(g, me.ID)
	if prompt == nil {
		t.Fatalf("no optional prompt: %+v", g.PendingChoices)
	}
	if err := g.ResolveTriggerPrompt(prompt.ID, me.ID, true); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life-1 || effectstest.CountOnBattlefield(g, "Treasure", me.ID) != 1 {
		t.Errorf("life %d -> %d, Treasures %d; want 1 life paid and one Treasure", life, me.Life, effectstest.CountOnBattlefield(g, "Treasure", me.ID))
	}
	// A token dying makes nothing.
	tok := effectstest.PushToken(g, opp.ID, effects.TokenCard("1/1 white Soldier"))
	g.WithWriteLock(func() { _ = g.SacrificePermanentForEffect(tok) })
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestTriggerPrompt(g, me.ID) != nil {
		t.Error("a token's death offered the Treasure")
	}
}

func TestOldB11ShaddamDestroysTheGreatestPowerOpposingCreatureAtTheEndStep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Padishah Emperor Shaddam IV", "Legendary Creature — Human Noble", oldB11Shaddam, 2, 4)
	mine := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Ogre", 9, 9)
	small := effectstest.B12Creature(g, opp.ID, "Small", "Creature — Bear", 2, 2)
	big := effectstest.B12Creature(g, opp.ID, "Big", "Creature — Ogre", 5, 5)
	effectstest.AdvanceToEndStep(t, g)
	if p := effectstest.LatestPickTarget(g, me.ID); p != nil {
		if err := g.ResolvePickTargets(p.ID, me.ID, []game.TargetRef{{Kind: game.TargetCard, ID: small}}); err == nil {
			t.Fatal("a creature without the greatest power was a legal target")
		}
		if err := g.ResolvePickTargets(p.ID, me.ID, []game.TargetRef{{Kind: game.TargetCard, ID: big}}); err != nil {
			t.Fatal(err)
		}
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(big) || !g.Battlefield.Contains(small) || !g.Battlefield.Contains(mine) {
		t.Errorf("big %v small %v mine %v on the battlefield", g.Battlefield.Contains(big), g.Battlefield.Contains(small), g.Battlefield.Contains(mine))
	}
}

func TestOldB11QuestingPhelddagrifGivesTheOpponentASomethingForEachColour(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.ToMain(t, g)
	pd := effectstest.B12Push(g, me.ID, "Questing Phelddagrif", "Creature — Phelddagrif", oldB11Phelddag, 4, 4)
	target := []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}
	activate := func(row int, mana string) {
		t.Helper()
		effectstest.FloatForTest(g, me, mana)
		if err := g.ActivateCatalogAbility(me.ID, pd, row, game.ActivateAbilityParams{Strict: true, Targets: target}); err != nil {
			t.Fatal(err)
		}
		effectstest.PassPriorityAroundTable(t, g)
	}
	activate(0, "G")
	if c := effectstest.B12Card(t, g, pd); c.Effective().Power != 5 {
		t.Errorf("power %d, want 5", c.Effective().Power)
	}
	if got := effectstest.CountOnBattlefield(g, "Hippo", opp.ID); got != 1 {
		t.Errorf("the opponent's Hippos: %d, want 1", got)
	}
	life := opp.Life
	activate(1, "W")
	if opp.Life != life+2 {
		t.Errorf("opponent life %d, want %d", opp.Life, life+2)
	}
	if !effectstest.HasKeywordOf(g, pd, "protection from black") || !effectstest.HasKeywordOf(g, pd, "protection from red") {
		t.Error("no protection from black and from red")
	}
	hand := opp.Hand.Size()
	activate(2, "U")
	if !effectstest.HasKeywordOf(g, pd, "flying") {
		t.Error("no flying")
	}
	effectstest.AnswerMayChoice(t, g, opp.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Hand.Size() != hand+1 {
		t.Errorf("opponent hand %d, want %d", opp.Hand.Size(), hand+1)
	}
}

func TestOldB11SonicPumpsFlashAndHasteCreaturesAndMakesTappedTreasure(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	sonic := effectstest.B12Push(g, me.ID, "Sonic the Hedgehog", "Legendary Creature — Hedgehog Warrior", oldB11Sonic, 2, 4)
	hasty := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Hasty", TypeLine: "Creature — Goblin", Power: 1, Toughness: 3,
		Keywords: []string{"haste"}, Owner: me.ID, Controller: me.ID,
	})
	plain := effectstest.B12Creature(g, me.ID, "Plain", "Creature — Bear", 2, 2)
	effectstest.AttackWithTrigger(t, g, sonic, opp.ID, uuid.Nil)
	for id, want := range map[uuid.UUID]int{sonic: 1, hasty: 1, plain: 0} {
		if got := effectstest.CounterCount(g, id, game.CounterPlusOne); got != want {
			t.Errorf("%s: %d counters, want %d", effectstest.B12Card(t, g, id).Name, got, want)
		}
	}
}

func TestOldB11SonicMakesATappedTreasureWhenAFlashOrHasteCreatureIsDealtDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.ToMain(t, g)
	effectstest.B12Push(g, me.ID, "Sonic the Hedgehog", "Legendary Creature — Hedgehog Warrior", oldB11Sonic, 2, 4)
	plain := effectstest.B12Creature(g, me.ID, "Plain", "Creature — Bear", 2, 5)
	effectstest.B12Bolt(t, g, game.TargetRef{Kind: game.TargetCard, ID: plain})
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountOnBattlefield(g, "Treasure", me.ID); got != 0 {
		t.Fatalf("a creature without flash or haste made %d Treasures", got)
	}
	hasty := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Hasty", TypeLine: "Creature — Goblin", Power: 1, Toughness: 5,
		Keywords: []string{"haste"}, Owner: me.ID, Controller: me.ID,
	})
	effectstest.B12Bolt(t, g, game.TargetRef{Kind: game.TargetCard, ID: hasty})
	effectstest.PassPriorityAroundTable(t, g)
	n, tapped := 0, 0
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Treasure" && c.Controller == me.ID {
			n++
			if c.Tapped {
				tapped++
			}
		}
	}
	if n != 1 || tapped != 1 {
		t.Errorf("Treasures %d (tapped %d), want one tapped", n, tapped)
	}
}

func TestOldB11PharikaIsNotACreatureBelowSevenDevotionAndExilesForASnake(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.ToMain(t, g)
	pharika := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Pharika, God of Affliction", TypeLine: "Legendary Enchantment Creature — God",
		OracleID: oldB11Pharika, ManaCost: "{1}{B}{G}", Colors: []string{"B", "G"}, Power: 5, Toughness: 5, Owner: me.ID, Controller: me.ID,
	})
	if effectstest.B39IsCreature(t, g, pharika) {
		t.Error("Pharika is a creature at devotion 2")
	}
	dead := uuid.New()
	opp.Graveyard.PushTop(game.Card{InstanceID: dead, Name: "Dead Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2, Owner: opp.ID, Controller: opp.ID})
	effectstest.FloatForTest(g, me, "BG")
	if err := g.ActivateCatalogAbility(me.ID, pharika, 0, game.ActivateAbilityParams{
		Strict: true, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: dead}},
	}); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.B12ZoneOf(g, dead); got != game.ZoneExile {
		t.Errorf("the card is in %s, want exile", got)
	}
	if got := effectstest.CountOnBattlefield(g, "Snake", opp.ID); got != 1 {
		t.Errorf("the owner's Snakes: %d, want 1", got)
	}
	if got := effectstest.CountOnBattlefield(g, "Snake", me.ID); got != 0 {
		t.Errorf("the activator got %d Snakes", got)
	}
}

func TestOldB11RevelInRichesMakesTreasureForOpponentDeathsAndWinsAtTen(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.ToMain(t, g)
	effectstest.B12Push(g, me.ID, "Revel in Riches", "Enchantment", oldB11Revel, 0, 0)
	foe := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	mine := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() { _ = g.SacrificePermanentForEffect(mine) })
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountOnBattlefield(g, "Treasure", me.ID); got != 0 {
		t.Fatalf("your own creature dying made %d Treasures", got)
	}
	g.WithWriteLock(func() { _ = g.SacrificePermanentForEffect(foe) })
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountOnBattlefield(g, "Treasure", me.ID); got != 1 {
		t.Fatalf("an opponent's creature dying made %d Treasures, want 1", got)
	}
	effectstest.B12ToMyNextUpkeep(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if g.State == game.StateEnded {
		t.Fatal("won with one Treasure")
	}
	effectstest.PushTokens(g, me.ID, effects.TreasureToken(), 9)
	effectstest.B12ToMyNextUpkeep(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if g.State != game.StateEnded || g.Outcome == nil || g.Outcome.Winner != me.ID {
		t.Errorf("with ten Treasures the controller did not win (state %v)", g.State)
	}
}

func TestOldB11HanskGivesTheOpponentWalkersAndDrawsWhenTheirZombieDies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	hansk := effectstest.B12Push(g, me.ID, "Hansk, Slayer Zealot", "Legendary Creature — Human Archer", oldB11Hansk, 4, 4)
	effectstest.B12ToMyNextUpkeep(t, g)
	if p := effectstest.LatestPickTarget(g, me.ID); p != nil {
		if err := g.ResolvePickTargets(p.ID, me.ID, []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}); err != nil {
			t.Fatal(err)
		}
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountOnBattlefield(g, "Walker", opp.ID); got != 3 {
		t.Fatalf("the opponent's Walkers: %d, want 3", got)
	}
	var walker uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Walker" && c.Controller == opp.ID {
			walker = c.InstanceID
			break
		}
	}
	hand := me.Hand.Size()
	if err := g.ActivateCatalogAbility(me.ID, hansk, 0, game.ActivateAbilityParams{
		Strict: true, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: walker}},
	}); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(walker) {
		t.Error("2 damage did not kill the 2/2 Walker")
	}
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d -> %d, want one card drawn for the Zombie that died", hand, me.Hand.Size())
	}
}
