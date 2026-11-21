package testsru

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const (
	gapVanguardOracle    = "3da6d24b-ecbd-4a41-a37e-84a19b243414"
	gapLinebreakerOracle = "db432387-58d3-4473-9442-b5b586dbea38"
	gapChrysalisOracle   = "d27bebc9-348d-4b8e-816a-a4da66fb5618"
	gapFrogmyrOracle     = "ccc7de36-aafc-4d87-80f7-fc2effc976eb"
	gapKudoOracle        = "070d22cf-2eed-4d04-b932-4b8339ea9b7a"
	gapTrainerOracle     = "9e0461e7-ccc4-413b-be80-83311be46861"
	gapDragonOracle      = "be007706-3da7-4dd3-84db-dbb5d8831bba"
	gapNaduOracle        = "0ff11a61-6fe3-4c70-aba8-417023a5e54e"
	gapAshlingOracle     = "6fe9fd59-9cac-4572-8f94-bc19bea44f94"
)

func gapB52Counters(g *game.Game, id uuid.UUID) int {
	c, _ := effectstest.BattlefieldCard(g, id)
	return c.Counters[game.CounterPlusOne]
}

func gapB52Stock(g *game.Game, p *game.Player, n int) {
	g.WithWriteLock(func() {
		for i := 0; i < n; i++ {
			p.Library.PushTop(game.NewCard("Filler", p.ID))
		}
	})
}

func gapB52Colored(g *game.Game, id uuid.UUID, colors ...string) {
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == id {
				g.Battlefield.Cards[i].Colors = colors
			}
		}
	})
}

func TestGapB52VanguardPutsCountersOnColorlessCreaturesWhenCastAndWhenItAttacks(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	colorless := effectstest.PushVanillaCreature(g, me.ID, "Colorless", 2, 2)
	green := effectstest.PushVanillaCreature(g, me.ID, "Green", 2, 2)
	gapB52Colored(g, green, "G")
	theirs := effectstest.PushVanillaCreature(g, opp.ID, "Theirs", 2, 2)
	id := effectstest.CastCatalogSpell(t, g, "Titans' Vanguard", "Creature — Eldrazi", gapVanguardOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if gapB52Counters(g, colorless) != 1 || gapB52Counters(g, green) != 0 || gapB52Counters(g, theirs) != 0 {
		t.Fatalf("after the cast: colorless %d, green %d, theirs %d, want 1, 0, 0",
			gapB52Counters(g, colorless), gapB52Counters(g, green), gapB52Counters(g, theirs))
	}
	if got := gapB52Counters(g, id); got != 0 {
		t.Errorf("the spell put a counter on itself before it was a creature: %d", got)
	}
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == id {
				g.Battlefield.Cards[i].SummonedThisTurn = false
			}
		}
	})
	effectstest.DeclareAttack(t, g, opp.ID, id)
	effectstest.PassPriorityAroundTable(t, g)
	if gapB52Counters(g, id) != 1 || gapB52Counters(g, colorless) != 2 {
		t.Errorf("after the attack: Vanguard %d, colorless %d, want 1 and 2", gapB52Counters(g, id), gapB52Counters(g, colorless))
	}
}

func TestGapB52LinebreakerGivesHasteAndPowerPerEldrazi(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushCatalogPermanent(g, me.ID, "Eldrazi Linebreaker", "Creature — Eldrazi", gapLinebreakerOracle, false)
	effectstest.PushDiesCreatureForTest(g, me.ID, "Other Eldrazi", "", "Creature — Eldrazi", 1, 1)
	bear := effectstest.PushVanillaCreature(g, me.ID, "Bear", 2, 2)
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	pick := effectstest.LatestChoiceOfKindFor(g, game.PendingChoicePickTarget, me.ID)
	if pick == nil {
		t.Fatal("no target prompt for the begin-combat trigger")
	}
	if err := g.ResolvePickTarget(pick.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: bear}); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasEffectiveKeyword(t, g, bear, "haste") {
		t.Error("the target did not gain haste")
	}
	if p := effectstest.EffectivePower(t, g, bear); p != 4 {
		t.Errorf("power = %d, want 4 (2 + two Eldrazi)", p)
	}
}

func TestGapB52ChrysalisMakesTwoSpawnAndGrowsWhenYouSacrificeAnEldrazi(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	id := effectstest.CastCatalogSpell(t, g, "Writhing Chrysalis", "Creature — Eldrazi Drone", gapChrysalisOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.RandomTokenCount(g, "Eldrazi Spawn"); got != 2 {
		t.Fatalf("Spawn tokens = %d, want 2", got)
	}
	var spawn uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Eldrazi Spawn" {
			spawn = c.InstanceID
			break
		}
	}
	g.WithWriteLock(func() { _ = g.SacrificePermanentForEffect(spawn) })
	effectstest.PassPriorityAroundTable(t, g)
	if got := gapB52Counters(g, id); got != 1 {
		t.Errorf("counters = %d, want 1", got)
	}
	nonEldrazi := effectstest.PushVanillaCreature(g, g.Seats[0].ID, "Bear", 2, 2)
	g.WithWriteLock(func() { _ = g.SacrificePermanentForEffect(nonEldrazi) })
	effectstest.PassPriorityAroundTable(t, g)
	if got := gapB52Counters(g, id); got != 1 {
		t.Errorf("sacrificing a non-Eldrazi changed the counters to %d", got)
	}
}

func TestGapB52FrogmyrPrototypedWithAffinityCostsLess(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	for i := 0; i < 2; i++ {
		effectstest.PushCatalogPermanent(g, me.ID, "Trinket", "Artifact", "", false)
	}
	c := effectstest.PrototypeCard(t, me.ID, gapFrogmyrOracle, "Frogmyr Enforcer", "Artifact Creature — Frog Myr", "{7}", 4, 4,
		"Prototype {3}{R} — 2/2 (You may cast this spell with different mana cost, color, and size. It keeps its abilities and types.)\nAffinity for artifacts")
	me.Hand.PushTop(c)
	effectstest.AdvanceToMain(t, g)
	effectstest.FloatMana(t, g, me, "{R}")
	effectstest.FloatMana(t, g, me, "{C}")
	if err := g.CastSpell(me.ID, c.InstanceID, game.CastSpellParams{Face: game.PrototypeFace, Strict: true}); err != nil {
		t.Fatalf("prototyped with two artifacts for {R}{C}: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p := effectstest.EffectivePower(t, g, c.InstanceID); p != 2 {
		t.Errorf("prototyped power = %d, want 2", p)
	}
}

func TestGapB52FrogmyrUnprototypedCostsSevenLessPerArtifact(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	for i := 0; i < 5; i++ {
		effectstest.PushCatalogPermanent(g, me.ID, "Trinket", "Artifact", "", false)
	}
	c := effectstest.PrototypeCard(t, me.ID, gapFrogmyrOracle, "Frogmyr Enforcer", "Artifact Creature — Frog Myr", "{7}", 4, 4,
		"Prototype {3}{R} — 2/2 (You may cast this spell with different mana cost, color, and size. It keeps its abilities and types.)\nAffinity for artifacts")
	me.Hand.PushTop(c)
	effectstest.AdvanceToMain(t, g)
	effectstest.FloatMana(t, g, me, "{C}")
	if err := g.CastSpell(me.ID, c.InstanceID, game.CastSpellParams{Strict: true}); err == nil {
		t.Fatal("cast for {C} against {7} less five artifacts")
	}
	effectstest.FloatMana(t, g, me, "{C}")
	if err := g.CastSpell(me.ID, c.InstanceID, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("{C}{C} should pay {7} with affinity for five artifacts: %v", err)
	}
}

func TestGapB52KudoMakesOtherCreaturesTwoTwoBears(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	kudo := effectstest.PushCatalogPermanent(g, me.ID, "Kudo, King Among Bears", "Legendary Creature — Bear", gapKudoOracle, false)
	mine := effectstest.PushDiesCreatureForTest(g, me.ID, "Big Dragon", "", "Creature — Dragon", 5, 5)
	theirs := effectstest.PushDiesCreatureForTest(g, opp.ID, "Their Giant", "", "Creature — Giant", 7, 7)
	g.WithWriteLock(func() { g.BumpLayerVersionForTest() })
	for _, id := range []uuid.UUID{mine, theirs} {
		if p, tough := effectstest.EffectivePower(t, g, id), effectstest.EffectiveToughness(t, g, id); p != 2 || tough != 2 {
			t.Errorf("%v is %d/%d, want 2/2", id, p, tough)
		}
		c := effectstest.BattlefieldCardCopy(t, g, id)
		if !c.HasSubtype("Bear") {
			t.Errorf("%v is not a Bear", id)
		}
	}
	if !effectstest.BattlefieldCardCopy(t, g, mine).HasSubtype("Dragon") {
		t.Error("it stopped being a Dragon: Bear is in addition to its other types")
	}
	if p := effectstest.EffectivePower(t, g, kudo); p != 1 {
		// pushCatalogPermanent seeds 1/1; Kudo itself is not "other".
		t.Errorf("Kudo's own power = %d, want its own 1", p)
	}
}

func TestGapB52TrainerReducesAuraAndEquipmentSpellsByItsPower(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	trainer := effectstest.PushCatalogPermanent(g, me.ID, "Golden-Tail Trainer", "Creature — Fox Samurai", gapTrainerOracle, false)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == trainer {
				g.Battlefield.Cards[i].Power, g.Battlefield.Cards[i].Toughness = 2, 3
			}
		}
	})
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Test Blade", TypeLine: "Artifact — Equipment", ManaCost: "{3}",
		Owner: me.ID, Controller: me.ID})
	effectstest.AdvanceToMain(t, g)
	effectstest.FloatMana(t, g, me, "{C}")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("an Equipment for {3} with a power-2 Trainer should cost {1}: %v", err)
	}
	other := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: other, Name: "Test Rock", TypeLine: "Artifact", ManaCost: "{3}",
		Owner: me.ID, Controller: me.ID})
	effectstest.FloatMana(t, g, me, "{C}")
	if err := g.CastSpell(me.ID, other, game.CastSpellParams{Strict: true}); err == nil {
		t.Error("a plain artifact got the discount")
	}
}

func TestGapB52TrainerPumpsOtherModifiedCreaturesWhenItAttacks(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	trainer := effectstest.PushCatalogPermanent(g, me.ID, "Golden-Tail Trainer", "Creature — Fox Samurai", gapTrainerOracle, false)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == trainer {
				g.Battlefield.Cards[i].Power, g.Battlefield.Cards[i].Toughness = 2, 3
			}
		}
	})
	modified := effectstest.PushVanillaCreature(g, me.ID, "Modified", 2, 2)
	plain := effectstest.PushVanillaCreature(g, me.ID, "Plain", 2, 2)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == modified {
				g.Battlefield.Cards[i].Counters = map[string]int{game.CounterPlusOne: 1}
			}
		}
	})
	effectstest.DeclareAttack(t, g, opp.ID, trainer)
	effectstest.PassPriorityAroundTable(t, g)
	if p := effectstest.EffectivePower(t, g, modified); p != 4 {
		t.Errorf("modified creature power = %d, want 4 (2 + X)", p)
	}
	if p := effectstest.EffectivePower(t, g, plain); p != 2 {
		t.Errorf("unmodified creature power = %d, want 2", p)
	}
	if p := effectstest.EffectivePower(t, g, trainer); p != 2 {
		t.Errorf("the Trainer pumped itself: power %d", p)
	}
}

func TestGapB52DragonGainsPlusThreeAndFirebreathingAtThreshold(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	dragon := effectstest.PushCatalogPermanent(g, me.ID, "Fledgling Dragon", "Creature — Dragon", gapDragonOracle, false)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == dragon {
				g.Battlefield.Cards[i].Power, g.Battlefield.Cards[i].Toughness = 2, 2
			}
		}
	})
	for i := 0; i < 6; i++ {
		effectstest.PushGraveyardCardTyped(me, "Spent", "Sorcery")
	}
	g.WithWriteLock(func() { g.BumpLayerVersionForTest() })
	if p := effectstest.EffectivePower(t, g, dragon); p != 2 {
		t.Fatalf("below threshold power = %d, want 2", p)
	}
	effectstest.FloatMana(t, g, me, "{R}")
	if err := g.ActivateCatalogAbility(me.ID, dragon, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("the firebreathing ability was available below threshold")
	}
	effectstest.PushGraveyardCardTyped(me, "Spent", "Sorcery")
	g.WithWriteLock(func() { g.BumpLayerVersionForTest() })
	if p, tough := effectstest.EffectivePower(t, g, dragon), effectstest.EffectiveToughness(t, g, dragon); p != 5 || tough != 5 {
		t.Fatalf("at threshold %d/%d, want 5/5", p, tough)
	}
	if err := g.ActivateCatalogAbility(me.ID, dragon, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("firebreathing at threshold: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p := effectstest.EffectivePower(t, g, dragon); p != 6 {
		t.Errorf("power after {R} = %d, want 6", p)
	}
}

func TestGapB52NaduRevealsOnTargetingAtMostTwiceEachTurn(t *testing.T) {
	const boltOracle = "4457ed35-7c10-48c8-9776-456485fdf070"
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushCatalogPermanent(g, me.ID, "Nadu, Winged Wisdom", "Legendary Creature — Bird Wizard", gapNaduOracle, false)
	bear := effectstest.PushDiesCreatureForTest(g, me.ID, "Big Bear", "", "Creature — Bear", 9, 9)
	land := game.NewCard("Top Forest", me.ID)
	land.TypeLine = "Basic Land — Forest"
	spell := game.NewCard("Top Spell", me.ID)
	spell.TypeLine = "Sorcery"
	filler := game.NewCard("Filler Three", me.ID)
	filler.TypeLine = "Sorcery"
	g.WithWriteLock(func() {
		me.Library.PushTop(filler)
		me.Library.PushTop(spell)
		me.Library.PushTop(land)
		g.BumpLayerVersionForTest()
	})
	target := []game.TargetRef{{Kind: game.TargetCard, ID: bear}}
	for i := 1; i <= 3; i++ {
		effectstest.CastCatalogSpell(t, g, "Lightning Bolt", "Instant", boltOracle, target)
		effectstest.PassPriorityAroundTable(t, g)
		switch i {
		case 1:
			if !g.Battlefield.Contains(land.InstanceID) {
				t.Fatal("a land on top did not enter the battlefield")
			}
		case 2:
			if !me.Hand.Contains(spell.InstanceID) {
				t.Fatal("a nonland on top did not go to the hand")
			}
		case 3:
			if me.Hand.Contains(filler.InstanceID) {
				t.Error("the ability triggered a third time in one turn")
			}
		}
	}
}

func TestGapB52AshlingLoots_DealsDamageSecondTimeAndAddsManaThirdTime(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	gapB52Stock(g, me, 10)
	effectstest.PushCatalogPermanent(g, me.ID, "Ashling, Flame Dancer", "Legendary Creature — Elemental Shaman", gapAshlingOracle, false)
	victim := effectstest.PushDiesCreatureForTest(g, opp.ID, "Victim", "", "Creature — Bear", 2, 2)
	mine := effectstest.PushDiesCreatureForTest(g, me.ID, "Mine", "", "Creature — Bear", 2, 2)
	oppLife := opp.Life
	for i := 1; i <= 3; i++ {
		pushHandForAshling(me)
		effectstest.CastCatalogSpell(t, g, "Test Instant", "Instant", "", nil)
		effectstest.PassPriorityAroundTable(t, g)
		if c := effectstest.ChooseCardsChoiceFor(g, me.ID); c != nil {
			if err := g.ResolveChooseCards(c.ID, me.ID, c.ChooseCards[:1]); err != nil {
				t.Fatal(err)
			}
		}
		effectstest.PassPriorityAroundTable(t, g)
		switch i {
		case 1:
			if opp.Life != oppLife || !g.Battlefield.Contains(victim) {
				t.Fatalf("first resolution dealt damage")
			}
		case 2:
			if opp.Life != oppLife-2 {
				t.Errorf("second resolution: opponent life %d, want %d", opp.Life, oppLife-2)
			}
			if g.Battlefield.Contains(victim) {
				t.Error("second resolution left the opponent's creature alive")
			}
			if !g.Battlefield.Contains(mine) {
				t.Error("second resolution hurt my own creature")
			}
		case 3:
			if got := effectstest.HazelPoolTotal(me); got < 4 {
				t.Errorf("third resolution: pool has %d mana, want {R}{R}{R}{R}", got)
			}
		}
	}
}

func pushHandForAshling(p *game.Player) {
	c := game.NewCard("Discard Me", p.ID)
	c.TypeLine = "Sorcery"
	p.Hand.PushTop(c)
}
