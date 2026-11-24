package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b064_cards_test.go — Kaldheim, batch 064 (damage): each test plays the
// real catalog entry through the engine.

const (
	oldB064Alrund    = "0a69c3dd-52a3-4ef6-b53f-b0780e0dc5f8"
	oldB064FrostBite = "128170ed-c86a-4f02-9244-28197be90c10"
	oldB064Cinder    = "13efba9f-aedd-4ec0-a387-60c09b887341"
	oldB064Brawl     = "25a0fe3b-59c8-496a-a8d6-fca966ca6602"
	oldB064Reidane   = "35ae7b25-6d0f-49f8-a40d-dd2cf88e6f44"
	oldB064Fumarole  = "52befc89-e895-4dd5-95e4-8f6d03e627aa"
	oldB064Ravager   = "980c32ff-c154-4ac9-adc2-ed3a85c8d71c"
	oldB064Bearer    = "b3f7c4ac-b788-4307-9bc3-d99ce310e07c"
	oldB064Flinger   = "f701ada1-e9e1-42ce-9a62-c11bcec03da7"
)

func TestOldB064EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB064Alrund:         "Alrund, God of the Cosmos",
		oldB064Alrund + "#1":  "Hakka, Whispering Raven",
		oldB064FrostBite:      "Frost Bite",
		oldB064Cinder:         "Cinderheart Giant",
		oldB064Brawl:          "Blizzard Brawl",
		oldB064Reidane:        "Reidane, God of the Worthy",
		oldB064Reidane + "#1": "Valkmira, Protector's Shield",
		oldB064Fumarole:       "Tundra Fumarole",
		oldB064Ravager:        "Basalt Ravager",
		oldB064Bearer:         "Calamity Bearer",
		oldB064Flinger:        "Surtland Flinger",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok || spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: %+v registered=%v, want Full with no caveats", name, spec, ok)
		}
	}
}

// oldB064PickPlayer answers a pending target prompt with an opposing player.
func oldB064PickPlayer(t *testing.T, g *game.Game, me, opp *game.Player) {
	t.Helper()
	p := effectstest.LatestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("no target prompt")
	}
	if err := g.ResolvePickTarget(p.ID, p.Chooser, game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID}); err != nil {
		t.Fatalf("ResolvePickTarget: %v", err)
	}
}

func oldB064Snow(g *game.Game, owner uuid.UUID, n int, basic string) {
	for i := 0; i < n; i++ {
		effectstest.BlbB1Push(g, owner, "Snow-Covered "+basic, "Basic Snow Land — "+basic, "", 0, 0)
	}
}

func TestOldB064FrostBiteDealsThreeWithThreeSnowPermanents(t *testing.T) {
	for _, snow := range []int{2, 3} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		ogre := effectstest.B12Creature(g, opp.ID, "Ogre", "Creature — Ogre", 3, 3)
		oldB064Snow(g, me.ID, snow, "Mountain")
		effectstest.ToMainForCost(t, g)
		effectstest.B06AddMana(me, "R")
		id := effectstest.HandCardFull(me, "Frost Bite", "Snow Instant", "{R}", oldB064FrostBite, []string{"R"})
		if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, Targets: effectstest.CardTarget(ogre)}); err != nil {
			t.Fatalf("cast: %v", err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		if died := !g.Battlefield.Contains(ogre); died != (snow >= 3) {
			t.Errorf("%d snow permanents: the 3/3 died = %v, want %v", snow, died, snow >= 3)
		}
	}
}

func TestOldB064BlizzardBrawlPumpsAndProtectsWithThreeSnowPermanents(t *testing.T) {
	for _, snow := range []int{0, 3} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		mine := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
		theirs := effectstest.B12Creature(g, opp.ID, "Their Ogre", "Creature — Ogre", 3, 3)
		oldB064Snow(g, me.ID, snow, "Forest")
		effectstest.ToMainForCost(t, g)
		effectstest.B06AddMana(me, "G")
		id := effectstest.HandCardFull(me, "Blizzard Brawl", "Snow Sorcery", "{G}", oldB064Brawl, []string{"G"})
		targets := []game.TargetRef{{Kind: game.TargetCard, ID: mine}, {Kind: game.TargetCard, ID: theirs}}
		if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, Targets: targets}); err != nil {
			t.Fatalf("cast: %v", err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		if snow == 3 {
			if !g.Battlefield.Contains(mine) {
				t.Error("with three snow permanents my indestructible bear died in the fight")
			}
			if g.Battlefield.Contains(theirs) {
				t.Error("their 3/3 survived the 3 damage of a pumped 3/2")
			}
		} else if g.Battlefield.Contains(mine) {
			t.Error("with no snow permanents my 2/2 survived 3 damage")
		}
	}
}

func TestOldB064TundraFumaroleAddsColourlessForSnowManaAndKeepsIt(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ogre := effectstest.B12Creature(g, opp.ID, "Ogre", "Creature — Ogre", 4, 4)
	oldB064Snow(g, me.ID, 2, "Mountain")
	effectstest.BlbB1Push(g, me.ID, "Mountain", "Basic Land — Mountain", "", 0, 0)
	effectstest.ToMainForCost(t, g)
	id := effectstest.HandCardFull(me, "Tundra Fumarole", "Snow Sorcery", "{1}{R}{R}", oldB064Fumarole, []string{"R"})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, AutoTap: true, Targets: effectstest.CardTarget(ogre)}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(ogre) {
		t.Error("the 4/4 survived 4 damage")
	}
	if got := me.ManaPool.Total(); got != 2 {
		t.Fatalf("%d mana in the pool, want {C}{C} for the two snow mana spent", got)
	}
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	if got := me.ManaPool.Total(); got != 2 {
		t.Errorf("%d mana after the step ended, want it kept until end of turn", got)
	}
}

func TestOldB064CinderheartGiantHitsARandomOpposingCreatureWhenItDies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	giant := effectstest.B12Push(g, me.ID, "Cinderheart Giant", "Creature — Giant Berserker", oldB064Cinder, 7, 6)
	mine := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 3, 3)
	effectstest.KillCreature(t, g, me.ID, giant)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(theirs) {
		t.Error("their only creature survived 7 damage")
	}
	if !g.Battlefield.Contains(mine) {
		t.Error("my own creature was hit")
	}
}

func TestOldB064BasaltRavagerDealsTheGreatestSharedTypeCount(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Creature(g, me.ID, "Elf A", "Creature — Elf Druid", 1, 1)
	effectstest.B12Creature(g, me.ID, "Elf B", "Creature — Elf Warrior", 1, 1)
	effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 1, 1)
	life := opp.Life
	effectstest.CastRingCreature(t, g, "Basalt Ravager", "Creature — Giant Wizard", oldB064Ravager, 4, 2)
	effectstest.PassPriorityAroundTable(t, g)
	oldB064PickPlayer(t, g, me, opp)
	effectstest.PassPriorityAroundTable(t, g)
	// Two Elves are the largest group; the Giant Wizard is alone.
	if opp.Life != life-2 {
		t.Errorf("life %d -> %d, want 2 damage", life, opp.Life)
	}
}

func TestOldB064CalamityBearerDoublesOnlyGiantDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Calamity Bearer", "Creature — Giant Berserker", oldB064Bearer, 3, 4)
	giant := effectstest.B12Creature(g, me.ID, "Giant", "Creature — Giant", 3, 3)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 3, 3)
	life := opp.Life
	g.WithWriteLock(func() { _ = g.DealDamageToPlayerForEffect(giant, opp.ID, 3) })
	if opp.Life != life-6 {
		t.Errorf("a Giant's 3 damage took the life from %d to %d, want 6 lost", life, opp.Life)
	}
	life = opp.Life
	g.WithWriteLock(func() { _ = g.DealDamageToPlayerForEffect(bear, opp.ID, 3) })
	if opp.Life != life-3 {
		t.Errorf("a Bear's 3 damage took the life from %d to %d, want 3 lost", life, opp.Life)
	}
	theirs := effectstest.B12Creature(g, opp.ID, "Their Giant", "Creature — Giant", 3, 3)
	life = me.Life
	g.WithWriteLock(func() { _ = g.DealDamageToPlayerForEffect(theirs, me.ID, 3) })
	if me.Life != life-3 {
		t.Errorf("an opposing Giant's damage was doubled: %d -> %d", life, me.Life)
	}
}

func TestOldB064SurtlandFlingerSacrificesForDamageDoubledForAGiant(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine string
		power, damage  int
	}{{"Giant", "Creature — Giant", 3, 6}, {"Bear", "Creature — Bear", 3, 3}} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			flinger := effectstest.PushCatalogPermanent(g, me.ID, "Surtland Flinger", "Creature — Giant Berserker", oldB064Flinger, false)
			fodder := effectstest.B12Creature(g, me.ID, tc.name, tc.typeLine, tc.power, 3)
			life := opp.Life
			effectstest.DeclareAttack(t, g, opp.ID, flinger)
			effectstest.PassPriorityAroundTable(t, g)
			pick := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceOwnPermanents, me.ID)
			if pick == nil {
				t.Fatal("no sacrifice prompt")
			}
			if err := g.ResolveOwnPermanents(pick.ID, me.ID, []uuid.UUID{fodder}); err != nil {
				t.Fatalf("sacrifice: %v", err)
			}
			effectstest.PassPriorityAroundTable(t, g)
			oldB064PickPlayer(t, g, me, opp)
			effectstest.PassPriorityAroundTable(t, g)
			if g.Battlefield.Contains(fodder) {
				t.Error("the creature was not sacrificed")
			}
			if opp.Life != life-tc.damage {
				t.Errorf("life %d -> %d, want %d damage", life, opp.Life, tc.damage)
			}
		})
	}
}

func TestOldB064ReidaneTapsOpposingSnowLandsAndTaxesBigNoncreatureSpells(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Reidane, God of the Worthy", "Legendary Creature — God", oldB064Reidane, 2, 3)
	effectstest.AdvanceToUpkeepOf(t, g, 1)
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	snow := uuid.New()
	opp.Hand.PushTop(game.Card{InstanceID: snow, Name: "Snow-Covered Island", TypeLine: "Basic Snow Land — Island", Owner: opp.ID, Controller: opp.ID})
	if err := g.CastSpell(opp.ID, snow, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("play the snow land: %v", err)
	}
	if c, ok := effectstest.BattlefieldCard(g, snow); !ok || !c.Tapped {
		t.Error("an opposing snow land did not enter tapped")
	}

	// A noncreature spell of mana value 4 costs {2} more; a cheaper one does not.
	effectstest.B06AddMana(opp, "U", "U", "U", "U")
	big := effectstest.HandCardFull(opp, "Big Sorcery", "Sorcery", "{3}{U}", "", []string{"U"})
	if err := g.CastSpell(opp.ID, big, game.CastSpellParams{Strict: true}); err == nil {
		t.Error("a mana value 4 sorcery was cast for its printed cost under Reidane")
	}
	effectstest.B06AddMana(opp, "U", "U")
	if err := g.CastSpell(opp.ID, big, game.CastSpellParams{Strict: true}); err != nil {
		t.Errorf("a mana value 4 sorcery with the {2} tax paid: %v", err)
	}
}

func TestOldB064ValkmiraPreventsOneAndTaxesTargeting(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Valkmira, Protector's Shield", "Legendary Artifact", oldB064Reidane+"#1", false)
	mine := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Ogre", "Creature — Ogre", 3, 3)
	life := me.Life
	g.WithWriteLock(func() { _ = g.DealDamageToPlayerForEffect(theirs, me.ID, 3) })
	if me.Life != life-2 {
		t.Errorf("3 damage from an opposing source took me %d -> %d, want 2 lost", life, me.Life)
	}
	life = opp.Life
	g.WithWriteLock(func() { _ = g.DealDamageToPlayerForEffect(mine, opp.ID, 3) })
	if opp.Life != life-3 {
		t.Errorf("my own source's damage was reduced: %d -> %d", life, opp.Life)
	}

	// An opposing spell that targets my creature is countered unless they pay {1}.
	effectstest.AdvanceToUpkeepOf(t, g, 1)
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	effectstest.B06AddMana(opp, "R")
	bite := effectstest.HandCardFull(opp, "Frost Bite", "Snow Instant", "{R}", oldB064FrostBite, []string{"R"})
	if err := g.CastSpell(opp.ID, bite, game.CastSpellParams{Strict: true, Targets: effectstest.CardTarget(mine)}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasPayUnlessFor(g, opp.ID) {
		t.Fatal("the targeting player was not offered the {1}")
	}
	effectstest.AnswerPayUnless(t, g, opp.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(mine) {
		t.Error("the spell resolved although its controller did not pay {1}")
	}
}

func TestOldB064AlrundGrowsWithHandAndForetoldCardsAndHakkaBounces(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	me.Hand.Cards = nil
	alrund := effectstest.PushCatalogPermanent(g, me.ID, "Alrund, God of the Cosmos", "Legendary Creature — God", oldB064Alrund, false)
	for i := 0; i < 3; i++ {
		effectstest.HandCardForTest(me, "Card", "Instant", "")
	}
	if p, tough := effectstest.PtOf(t, g, alrund); p != 4 || tough != 4 {
		t.Errorf("Alrund with three cards in hand is %d/%d, want 4/4", p, tough)
	}
	g.WithWriteLock(func() {
		g.Exile.PushTop(game.Card{InstanceID: uuid.New(), Name: "Foretold", TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID,
			FaceDown: true, FaceDownKind: game.FaceDownForetold})
		g.BumpLayerVersionForTest()
	})
	if p, _ := effectstest.PtOf(t, g, alrund); p != 5 {
		t.Errorf("Alrund with a foretold card is %d power, want 5", p)
	}
}

func TestOldB064AlrundEndStepTakesTheChosenTypeAndBottomsTheRest(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushCatalogPermanent(g, me.ID, "Alrund, God of the Cosmos", "Legendary Creature — God", oldB064Alrund, false)
	// The first card seeded is the top of the library: the two revealed are the Bear
	// and the Sorcery, and the Wolf stays unrevealed.
	effectstest.SeedSearchLibrary(me,
		game.Card{Name: "Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2},
		game.Card{Name: "Sorcery", TypeLine: "Sorcery"},
		game.Card{Name: "Wolf", TypeLine: "Creature — Wolf", Power: 2, Toughness: 2},
	)
	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.LatestOptionPickFor(g, me.ID)
	if pick == nil {
		t.Fatal("no card type prompt")
	}
	idx := -1
	for i, o := range pick.PickOptions {
		if o.Label == "Creature" {
			idx = i
		}
	}
	if idx < 0 {
		t.Fatalf("no Creature option in %+v", pick.PickOptions)
	}
	if err := g.ResolveOptionPick(pick.ID, me.ID, idx); err != nil {
		t.Fatalf("ResolveOptionPick: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	inHand := map[string]bool{}
	for _, c := range me.Hand.Cards {
		inHand[c.Name] = true
	}
	if !inHand["Bear"] || inHand["Sorcery"] || inHand["Wolf"] {
		t.Errorf("hand has Bear=%v Sorcery=%v Wolf=%v, want only the revealed Bear", inHand["Bear"], inHand["Sorcery"], inHand["Wolf"])
	}
	if got := me.Library.Cards[0].Name; got != "Sorcery" {
		t.Errorf("the bottom of the library is %q, want the revealed Sorcery", got)
	}
}

func TestOldB064HakkaReturnsItselfThenScries(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	hakka := effectstest.PushCatalogPermanent(g, me.ID, "Hakka, Whispering Raven", "Legendary Creature — Bird", oldB064Alrund+"#1", false)
	effectstest.DeclareAttack(t, g, opp.ID, hakka)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AdvanceTo(t, g, game.StepCombatDamage)
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(hakka) {
		t.Error("Hakka did not return to its owner's hand")
	}
	if effectstest.ScryChoiceFor(g, me.ID) == nil {
		t.Error("no scry 2 after the bounce")
	}
}
