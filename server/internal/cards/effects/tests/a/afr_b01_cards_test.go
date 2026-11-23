package testsa

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// afr_b01_cards_test.go — Adventures in the Forgotten Realms, batch 1
// (#5499). Each test plays the real catalog entry.

type afrB01Card struct {
	name, typeLine, oracle, cost string
	power, toughness             int
	keywords                     []string
}

var afrB01Cards = []afrB01Card{
	{"Adult Gold Dragon", "Creature — Dragon", "a361fdd1-7f97-4aed-8e24-ab2a03387f73", "{3}{R}{W}", 4, 3, []string{"flying", "lifelink", "haste"}},
	{"Hulking Bugbear", "Creature — Goblin", "c336eb5b-e185-4d4b-b106-3a641ba9aad1", "{1}{R}{R}", 3, 3, []string{"haste"}},
	{"Steadfast Paladin", "Creature — Dwarf Knight", "5e95618c-4f58-4ae9-9e5e-deffd5e210a4", "{1}{W}", 2, 2, []string{"lifelink"}},
	{"Underdark Basilisk", "Creature — Basilisk", "3b61b6da-44ed-4766-8134-fca37afd3f2a", "{1}{G}", 1, 2, []string{"deathtouch"}},
	{"Soulknife Spy", "Creature — Elf Rogue", "813c0654-7e3b-435b-b7a5-148f0c36a4be", "{2}{U}", 3, 2, nil},
	{"Owlbear", "Creature — Bird Bear", "0fab815f-bccf-40ef-95e9-ba64b137fb8c", "{3}{G}{G}", 4, 4, []string{"trample"}},
	{"Hill Giant Herdgorger", "Creature — Giant", "05c5752a-6ccb-45ff-bb85-5f89efd5944b", "{4}{G}{G}", 7, 6, nil},
	{"Priest of Ancient Lore", "Creature — Dwarf Cleric", "a89b0fd5-84b2-487c-8b03-e78d97276fd3", "{2}{W}", 2, 1, nil},
	{"Silver Raven", "Artifact Creature — Bird", "4da64806-8a1d-4926-a168-695df43309cf", "{U}", 1, 1, []string{"flying"}},
	{"Red Dragon", "Creature — Dragon", "ff6e4346-463c-445d-8f72-11cb35dd99ee", "{4}{R}{R}", 4, 4, []string{"flying"}},
	{"Black Dragon", "Creature — Dragon", "ad804712-18ba-49d7-9af2-4361edaa6cf8", "{5}{B}{B}", 4, 4, []string{"flying"}},
	{"White Dragon", "Creature — Dragon", "bb03e682-4525-40d2-8c49-6328f84bd36a", "{4}{W}{W}", 4, 4, []string{"flying"}},
	{"Arborea Pegasus", "Creature — Pegasus", "9a17452d-3adb-42ec-99f8-0083740cefa0", "{3}{W}", 2, 3, []string{"flying"}},
}

func afrB01Cast(t *testing.T, g *game.Game, c afrB01Card) uuid.UUID {
	t.Helper()
	me := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	me.Hand.PushTop(game.Card{
		InstanceID: id, Name: c.name, TypeLine: c.typeLine, OracleID: c.oracle, ManaCost: c.cost,
		Power: c.power, Toughness: c.toughness, Owner: me.ID, Controller: me.ID,
	})
	effectstest.AdvanceToMain(t, g)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true}); err == nil {
		t.Fatalf("%s was cast without its mana", c.name)
	}
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast %s: %v", c.name, err)
	}
	return id
}

func afrB01Named(name string) afrB01Card {
	for _, c := range afrB01Cards {
		if c.name == name {
			return c
		}
	}
	panic(name)
}

func TestAfrB01EveryCardIsRegisteredFullWithItsKeywords(t *testing.T) {
	for _, c := range afrB01Cards {
		spec, ok := effects.Lookup(c.oracle)
		if !ok {
			t.Errorf("%s is not registered", c.name)
			continue
		}
		if spec.Name != c.name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", c.name, spec.Name, spec.Completeness, spec.Caveats)
		}
		if len(spec.PrintedKeywords) != len(c.keywords) {
			t.Errorf("%s: keywords %v, want %v", c.name, spec.PrintedKeywords, c.keywords)
			continue
		}
		for i, k := range c.keywords {
			if spec.PrintedKeywords[i] != k {
				t.Errorf("%s: keywords %v, want %v", c.name, spec.PrintedKeywords, c.keywords)
			}
		}
	}
}

func TestAfrB01KeywordCreaturesHaveTheirKeywordsOnTheBattlefield(t *testing.T) {
	for _, c := range afrB01Cards {
		for _, kw := range c.keywords {
			t.Run(c.name+"/"+kw, func(t *testing.T) {
				g := effectstest.NewCatalogGame(t)
				id := effectstest.PushCatalogPermanent(g, g.Seats[0].ID, c.name, c.typeLine, c.oracle, false)
				if !effectstest.HasKeywordOnBattlefield(t, g, id, kw) {
					t.Errorf("%s lacks %s", c.name, kw)
				}
			})
		}
	}
}

func TestAfrB01EntersTriggersOnYourself(t *testing.T) {
	for _, tc := range []struct {
		name  string
		check func(t *testing.T, g *game.Game, handBefore, lifeBefore int)
	}{
		{"Owlbear", func(t *testing.T, g *game.Game, hand, _ int) {
			if got := g.Seats[0].Hand.Size(); got != hand+1 {
				t.Errorf("hand %d, want %d (cast one, drew one)", got, hand+1)
			}
		}},
		{"Priest of Ancient Lore", func(t *testing.T, g *game.Game, hand, life int) {
			if got := g.Seats[0].Hand.Size(); got != hand+1 {
				t.Errorf("hand %d, want %d", got, hand+1)
			}
			if got := g.Seats[0].Life; got != life+1 {
				t.Errorf("life %d, want %d", got, life+1)
			}
		}},
		{"Hill Giant Herdgorger", func(t *testing.T, g *game.Game, _, life int) {
			if got := g.Seats[0].Life; got != life+3 {
				t.Errorf("life %d, want %d", got, life+3)
			}
		}},
		{"Red Dragon", func(t *testing.T, g *game.Game, _, life int) {
			if g.Seats[0].Life != life {
				t.Error("the caster lost life")
			}
			for _, s := range g.Seats[1:] {
				if s.Life != life-4 {
					t.Errorf("%s at %d, want %d", s.ID, s.Life, life-4)
				}
			}
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			life := g.Seats[0].Life
			hand := g.Seats[0].Hand.Size()
			id := afrB01Cast(t, g, afrB01Named(tc.name))
			effectstest.PassPriorityAroundTable(t, g)
			if !g.Battlefield.Contains(id) {
				t.Fatal("not on the battlefield")
			}
			tc.check(t, g, hand, life)
		})
	}
}

func TestAfrB01SilverRavenScriesOne(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	afrB01Cast(t, g, afrB01Named("Silver Raven"))
	effectstest.PassPriorityAroundTable(t, g)
	scry := effectstest.LatestChoiceOfKind(g, game.PendingChoiceScry)
	if scry == nil || len(scry.ScryCards) != 1 {
		t.Fatalf("scry prompt: %+v", scry)
	}
	if err := g.ResolveScry(scry.ID, me.ID, nil, scry.ScryCards); err != nil {
		t.Fatalf("ResolveScry: %v", err)
	}
}

func TestAfrB01SoulknifeSpyDrawsOnCombatDamageToAPlayerOnly(t *testing.T) {
	c := afrB01Named("Soulknife Spy")
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	id := effectstest.PushCatalogPermanent(g, me.ID, c.name, c.typeLine, c.oracle, false)
	hand := me.Hand.Size()
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{Kind: game.EventDealDamage, Actor: me.ID, Source: id, Target: g.Seats[1].ID, Amount: 3})
	})
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand {
		t.Error("drew on noncombat damage")
	}
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{Kind: game.EventDealDamage, Actor: me.ID, Source: id, Target: g.Seats[1].ID, Amount: 3, Combat: true})
	})
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d, want %d", me.Hand.Size(), hand+1)
	}
}

func TestAfrB01BlackDragonGivesAnOpponentsCreatureMinusThreeMinusThree(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := effectstest.InvB1Push(g, me.ID, "Mine", "Creature — Bear", "", 4, 4)
	theirs := effectstest.InvB1Push(g, opp.ID, "Theirs", "Creature — Ogre", "", 5, 5)
	effectstest.PlsB1Cast(t, g, "Black Dragon", "Creature — Dragon", afrB01Named("Black Dragon").oracle, "{5}{B}{B}", 4, 4)
	effectstest.PlsB1Settle(t, g, func(c *game.PendingChoice) game.TargetRef {
		if effectstest.HasID(c.PickTargetCards, mine) {
			t.Error("a creature the caster controls was offered")
		}
		return game.TargetRef{Kind: game.TargetCard, ID: theirs}
	}, nil)
	if p, tg := effectstest.EffectivePower(t, g, theirs), effectstest.EffectiveToughness(t, g, theirs); p != 2 || tg != 2 {
		t.Errorf("target is %d/%d, want 2/2", p, tg)
	}
	if p := effectstest.EffectivePower(t, g, mine); p != 4 {
		t.Errorf("own creature power %d, want 4", p)
	}
}

func TestAfrB01WhiteDragonTapsAndFreezesAnOpponentsCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := effectstest.InvB1Push(g, me.ID, "Mine", "Creature — Bear", "", 2, 2)
	theirs := effectstest.InvB1Push(g, opp.ID, "Theirs", "Creature — Ogre", "", 5, 5)
	effectstest.PlsB1Cast(t, g, "White Dragon", "Creature — Dragon", afrB01Named("White Dragon").oracle, "{4}{W}{W}", 4, 4)
	effectstest.PlsB1Settle(t, g, func(c *game.PendingChoice) game.TargetRef {
		if effectstest.HasID(c.PickTargetCards, mine) {
			t.Error("a creature the caster controls was offered")
		}
		return game.TargetRef{Kind: game.TargetCard, ID: theirs}
	}, nil)
	c, _ := effectstest.BattlefieldCard(g, theirs)
	if !c.Tapped || len(c.NextUntapSkips) != 1 {
		t.Errorf("target tapped %v, markers %d", c.Tapped, len(c.NextUntapSkips))
	}
	if m, _ := effectstest.BattlefieldCard(g, mine); m.Tapped || len(m.NextUntapSkips) != 0 {
		t.Error("the caster's creature was affected")
	}
}

func TestAfrB01ArboreaPegasusPumpsAndGrantsFlyingUntilEndOfTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	theirs := effectstest.InvB1Push(g, opp.ID, "Theirs", "Creature — Ogre", "", 2, 2)
	effectstest.PlsB1Cast(t, g, "Arborea Pegasus", "Creature — Pegasus", afrB01Named("Arborea Pegasus").oracle, "{3}{W}", 2, 3)
	effectstest.PlsB1Settle(t, g, func(*game.PendingChoice) game.TargetRef {
		return game.TargetRef{Kind: game.TargetCard, ID: theirs}
	}, nil)
	if p, tg := effectstest.EffectivePower(t, g, theirs), effectstest.EffectiveToughness(t, g, theirs); p != 3 || tg != 3 {
		t.Errorf("target is %d/%d, want 3/3", p, tg)
	}
	if !effectstest.HasKeywordOnBattlefield(t, g, theirs, "flying") {
		t.Error("the target did not gain flying")
	}
}
