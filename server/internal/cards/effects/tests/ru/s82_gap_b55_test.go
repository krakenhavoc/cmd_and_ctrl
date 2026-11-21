package testsru

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const (
	gapMindlessConscriptionOracle = "bf65e547-9d81-4f4b-8adb-46aee9088906"
	gapMetastaticEvangelOracle    = "0a27bd77-ec13-4e8d-acad-9aa9c66b3b73"
	gapProudPackRhinoOracle       = "292cd535-916e-486e-b0cc-ecdb9477287d"
	gapQuestOracle                = "407d9d09-3db6-49b0-9e80-4e1538156e44"
)

func TestGapMindlessConscriptionAmassesOnEntryAndOnTheThirdDraw(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.CastWithCost(t, g, "Mindless Conscription", "Enchantment", "", gapMindlessConscriptionOracle)
	effectstest.PassPriorityAroundTable(t, g)
	army := effectstest.ArmyOf(t, g, me.ID)
	if army.Counters[game.CounterPlusOne] != 3 || !army.HasSubtype("Zombie") {
		t.Fatalf("after entering: Army counters %d, Zombie %v; want 3 and a Zombie", army.Counters[game.CounterPlusOne], army.HasSubtype("Zombie"))
	}
	draws := func() {
		g.WithWriteLock(func() { _ = g.DrawNForEffect(me.ID, 1) })
		effectstest.PassPriorityAroundTable(t, g)
	}
	// Draws one and two do nothing; the third amasses onto the same Army.
	for g.TurnTallyFor(me.ID).CardsDrawn < 2 {
		draws()
	}
	if got := effectstest.ArmyOf(t, g, me.ID).Counters[game.CounterPlusOne]; got != 3 {
		t.Fatalf("before the third draw: %d counters, want 3", got)
	}
	draws()
	if got := effectstest.ArmyOf(t, g, me.ID).Counters[game.CounterPlusOne]; got != 6 {
		t.Errorf("after the third draw: %d counters, want 6", got)
	}
	draws()
	if got := effectstest.ArmyOf(t, g, me.ID).Counters[game.CounterPlusOne]; got != 6 {
		t.Errorf("after the fourth draw: %d counters, want 6", got)
	}
}

func TestGapMetastaticEvangelProliferatesForAnotherNontokenCreatureOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	evangel := effectstest.B12Push(g, me.ID, "Metastatic Evangel", "Creature — Phyrexian Human Cleric", gapMetastaticEvangelOracle, 3, 1)
	ally := effectstest.B12Creature(g, me.ID, "Ally", "Creature — Bear", 2, 2)
	effectstest.PutCounters(t, g, ally, game.CounterPlusOne, 1)
	effectstest.AdvanceToMain(t, g)

	// A token entering does nothing.
	g.WithWriteLock(func() {
		_ = effects.CreateToken{Template: effects.TokenCard("1/1 white Soldier"), N: 1}.Apply(effects.NewContext(g, &game.StackItem{Controller: me.ID, SourceCardID: evangel}))
	})
	effectstest.SettleAnsweringProliferate(t, g)
	if got := effectstest.CounterCount(g, ally, game.CounterPlusOne); got != 1 {
		t.Fatalf("a token entered: counters %d, want 1", got)
	}
	// Another nontoken creature does.
	effectstest.CastRingCreature(t, g, "Grizzly Bears", "Creature — Bear", "", 2, 2)
	effectstest.SettleAnsweringProliferate(t, g)
	if got := effectstest.CounterCount(g, ally, game.CounterPlusOne); got != 2 {
		t.Errorf("a nontoken creature entered: counters %d, want 2", got)
	}
}

func TestGapMetastaticEvangelDoesNotTriggerForItself(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	ally := effectstest.B12Creature(g, me.ID, "Ally", "Creature — Bear", 2, 2)
	effectstest.PutCounters(t, g, ally, game.CounterPlusOne, 1)
	effectstest.CastRingCreature(t, g, "Metastatic Evangel", "Creature — Phyrexian Human Cleric", gapMetastaticEvangelOracle, 3, 1)
	effectstest.SettleAnsweringProliferate(t, g)
	if got := effectstest.CounterCount(g, ally, game.CounterPlusOne); got != 1 {
		t.Errorf("the Evangel's own entry proliferated: counters %d, want 1", got)
	}
}

func TestGapProudPackRhinoShieldCounterMode(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.CastRingCreature(t, g, "Proud Pack-Rhino", "Creature — Rhino", gapProudPackRhinoOracle, 3, 3)
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceModePick, me.ID)
	if pick == nil {
		t.Fatalf("no mode prompt: %+v", g.PendingChoices)
	}
	if err := g.ResolveModePick(pick.ID, me.ID, []int{0}); err != nil {
		t.Fatalf("ResolveModePick: %v", err)
	}
	effectstest.PickCard(t, g, me.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CounterCount(g, bear, game.CounterShield); got != 1 {
		t.Errorf("shield counters %d, want 1", got)
	}
}

func TestGapProudPackRhinoProliferateMode(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.PutCounters(t, g, bear, game.CounterPlusOne, 1)
	effectstest.CastRingCreature(t, g, "Proud Pack-Rhino", "Creature — Rhino", gapProudPackRhinoOracle, 3, 3)
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceModePick, me.ID)
	if pick == nil {
		t.Fatalf("no mode prompt: %+v", g.PendingChoices)
	}
	if err := g.ResolveModePick(pick.ID, me.ID, []int{1}); err != nil {
		t.Fatalf("ResolveModePick: %v", err)
	}
	effectstest.SettleAnsweringProliferate(t, g)
	if got := effectstest.CounterCount(g, bear, game.CounterPlusOne); got != 2 {
		t.Errorf("counters %d after proliferate, want 2", got)
	}
}

func TestGapQuestForTheNecropolisGathersCountersAndRaisesTheDead(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	quest := effectstest.PushPermanentForTest(g, me.ID, "Quest for the Necropolis", gapQuestOracle, "Enchantment")
	effectstest.AdvanceToMain(t, g)
	effectstest.B31PlayLand(t, g, "Forest", "Basic Land — Forest", "")
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CounterCount(g, quest, "quest"); got != 1 {
		t.Fatalf("landfall: quest counters %d, want 1", got)
	}
	effectstest.PutCounters(t, g, quest, "quest", 2)
	dead := effectstest.PushGraveyardCardTyped(opp, "Dead Ogre", "Creature — Ogre")
	// {5}{B} less three is {2}{B}: two mana is short.
	effectstest.FloatMana(t, g, me, "{B}{B}")
	if err := g.ActivateCatalogAbility(me.ID, quest, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(dead), Strict: true}); err == nil {
		t.Fatal("activated for {B}{B}, short of the reduced {2}{B}")
	}
	effectstest.FloatMana(t, g, me, "{B}{B}{B}")
	if err := g.ActivateCatalogAbility(me.ID, quest, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(dead), Strict: true}); err != nil {
		t.Fatalf("activate for the reduced cost: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	c, ok := effectstest.BattlefieldCard(g, dead)
	if !ok || c.Controller != me.ID {
		t.Errorf("the creature card did not enter under my control (on battlefield %v)", ok)
	}
	if g.Battlefield.Contains(quest) {
		t.Error("the enchantment was not sacrificed")
	}
}

func TestGapQuestForTheNecropolisIsSorcerySpeed(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	quest := effectstest.PushPermanentForTest(g, me.ID, "Quest for the Necropolis", gapQuestOracle, "Enchantment")
	dead := effectstest.PushGraveyardCardTyped(opp, "Dead Ogre", "Creature — Ogre")
	effectstest.AdvanceToMain(t, g)
	effectstest.FloatMana(t, g, me, "{B}{B}{B}{B}{B}{B}")
	effectstest.CastWithCost(t, g, "Distraction", "Instant", "{C}", "")
	if err := g.ActivateCatalogAbility(me.ID, quest, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(dead), Strict: true}); err == nil {
		t.Error("activated with something on the stack")
	}
}

const gapPathOfAnnihilationOracle = "eca6ffeb-463c-40b0-90b4-7e09e5579c10"

func TestGapPathOfAnnihilationMakesSpawnAndGivesEldraziAnyColorMana(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.CastWithCost(t, g, "Path of Annihilation", "Enchantment", "", gapPathOfAnnihilationOracle)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.TokensNamedControlledBy(g, "Eldrazi Spawn", me.ID); n != 2 {
		t.Fatalf("entered: %d Spawn, want 2", n)
	}
	eldrazi := effectstest.GaPut(g, me.ID, "Eldrazi Test", "Creature — Eldrazi", "")
	other := effectstest.GaPut(g, me.ID, "Bear", "Creature — Bear", "")
	theirs := effectstest.GaPut(g, opp.ID, "Their Eldrazi", "Creature — Eldrazi", "")
	if n := effectstest.GrantedManaRowsOf(g, eldrazi); n != 1 {
		t.Fatalf("my Eldrazi has %d granted mana rows, want 1", n)
	}
	if n := effectstest.GrantedManaRowsOf(g, other); n != 0 {
		t.Errorf("a non-Eldrazi has %d granted rows, want 0", n)
	}
	if n := effectstest.GrantedManaRowsOf(g, theirs); n != 0 {
		t.Errorf("an opponent's Eldrazi has %d granted rows, want 0", n)
	}
	if err := effectstest.TapGrantedMana(t, g, me.ID, eldrazi, "R"); err != nil {
		t.Fatalf("tap the Eldrazi: %v", err)
	}
	if got := effectstest.Batch01PoolColors(me); len(got) != 1 || got[0] != "R" {
		t.Errorf("pool %v, want [R]", got)
	}
}

func TestGapPathOfAnnihilationGainsFourForCreatureSpellsOfManaValueSevenOrMore(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushPermanentForTest(g, me.ID, "Path of Annihilation", gapPathOfAnnihilationOracle, "Enchantment")
	effectstest.ToMainForCost(t, g)
	cast := func(name, typeLine, cost string) {
		id := effectstest.PushHandCardWithManaCost(me, name, typeLine, cost)
		for i := range me.Hand.Cards {
			if me.Hand.Cards[i].InstanceID == id {
				me.Hand.Cards[i].Power, me.Hand.Cards[i].Toughness = 1, 1
			}
		}
		if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
			t.Fatalf("cast %s: %v", name, err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		effectstest.PassPriorityAroundTable(t, g)
	}
	life := me.Life
	cast("Six Drop", "Creature — Giant", "{6}")
	if me.Life != life {
		t.Fatalf("a mana value 6 creature spell gained life: %d -> %d", life, me.Life)
	}
	cast("Big Sorcery", "Sorcery", "{7}")
	if me.Life != life {
		t.Fatalf("a mana value 7 sorcery gained life: %d -> %d", life, me.Life)
	}
	cast("Seven Drop", "Creature — Giant", "{7}")
	if me.Life != life+4 {
		t.Errorf("a mana value 7 creature spell: life %d -> %d, want +4", life, me.Life)
	}
}

const (
	gapNyxbornUnicornOracle = "b9304d88-a1de-4300-9425-50350969bd70"
	gapIndebtedSpiritOracle = "c4c0a622-7325-4024-8b2e-6335f9a729e9"
)

func TestGapNyxbornUnicornBestowedGivesPlusTwoAndMentor(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := effectstest.PushVanillaCreature(g, me.ID, "Bear", 2, 2)
	weak := effectstest.PushVanillaCreature(g, me.ID, "Weakling", 1, 1)
	uni := effectstest.BestowCard(t, g, "Nyxborn Unicorn", "Enchantment Creature — Unicorn", gapNyxbornUnicornOracle, 2, 2)
	effectstest.CastBestowed(t, g, uni, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if p := effectstest.EffectivePower(t, g, bear); p != 4 {
		t.Fatalf("host power %d, want 4", p)
	}
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	for _, id := range []uuid.UUID{bear, weak} {
		if err := g.DeclareAttacker(id, opp.ID); err != nil {
			t.Fatalf("DeclareAttacker: %v", err)
		}
	}
	effectstest.LockInAttacks(t, g)
	effectstest.AnswerTriggerTargets(t, g, me.ID, weak)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CounterCount(g, weak, game.CounterPlusOne); got != 1 {
		t.Errorf("the lesser-power attacker has %d counters, want 1 (mentor from the enchanted creature)", got)
	}
}

func TestGapNyxbornUnicornIsAMentoringCreatureWhenCastNormally(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	uni := effectstest.B12Push(g, me.ID, "Nyxborn Unicorn", "Enchantment Creature — Unicorn", gapNyxbornUnicornOracle, 2, 2)
	weak := effectstest.PushVanillaCreature(g, me.ID, "Weakling", 1, 1)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	for _, id := range []uuid.UUID{uni, weak} {
		if err := g.DeclareAttacker(id, opp.ID); err != nil {
			t.Fatalf("DeclareAttacker: %v", err)
		}
	}
	effectstest.LockInAttacks(t, g)
	effectstest.AnswerTriggerTargets(t, g, me.ID, weak)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CounterCount(g, weak, game.CounterPlusOne); got != 1 {
		t.Errorf("mentor put %d counters on the weaker attacker, want 1", got)
	}
}

func TestGapIndebtedSpiritAfterlifeOnTheAuraAndOnTheEnchantedCreature(t *testing.T) {
	// The enchanted creature dies: it has afterlife 1.
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.PushVanillaCreature(g, me.ID, "Bear", 2, 2)
	sp := effectstest.BestowCard(t, g, "Indebted Spirit", "Enchantment Creature — Spirit", gapIndebtedSpiritOracle, 1, 1)
	effectstest.CastBestowed(t, g, sp, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if p := effectstest.EffectivePower(t, g, bear); p != 3 {
		t.Fatalf("host power %d, want 3", p)
	}
	effectstest.B27Kill(g, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.TokensNamedControlledBy(g, "Spirit", me.ID); n != 1 {
		t.Errorf("the enchanted creature died: %d Spirit tokens, want 1", n)
	}

	// The Aura itself goes to the graveyard while attached: its own afterlife.
	g2 := effectstest.NewCatalogGame(t)
	me2 := g2.Seats[0]
	bear2 := effectstest.PushVanillaCreature(g2, me2.ID, "Bear", 2, 2)
	sp2 := effectstest.BestowCard(t, g2, "Indebted Spirit", "Enchantment Creature — Spirit", gapIndebtedSpiritOracle, 1, 1)
	effectstest.CastBestowed(t, g2, sp2, bear2)
	effectstest.PassPriorityAroundTable(t, g2)
	effectstest.B27Kill(g2, sp2)
	effectstest.PassPriorityAroundTable(t, g2)
	if n := effectstest.TokensNamedControlledBy(g2, "Spirit", me2.ID); n != 1 {
		t.Errorf("the Aura was destroyed: %d Spirit tokens, want 1", n)
	}
	if !g2.Battlefield.Contains(bear2) {
		t.Error("the host left with the Aura")
	}
}

func TestGapIndebtedSpiritIsAnAfterlifeCreatureWhenCastNormally(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	sp := effectstest.B12Push(g, me.ID, "Indebted Spirit", "Enchantment Creature — Spirit", gapIndebtedSpiritOracle, 1, 1)
	effectstest.B27Kill(g, sp)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.TokensNamedControlledBy(g, "Spirit", me.ID); n != 1 {
		t.Errorf("%d Spirit tokens, want 1", n)
	}
}

const gapRosecotKnightOracle = "b81a9cf5-4059-45fb-a4e1-f2fe6aa3c841"

func TestGapRosecotKnightTakesAnArtifactOrEnchantmentOrGrows(t *testing.T) {
	t.Run("artifact taken", func(t *testing.T) {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		effectstest.SeedLibrary(me, "A", "B", "C", "D", "E", "F", "G")
		rock := effectstest.TopLibraryCard(me, "Rock", "Artifact")
		knight := effectstest.CastRingCreature(t, g, "Rosecot Knight", "Creature — Human Knight", gapRosecotKnightOracle, 3, 4)
		effectstest.PassPriorityAroundTable(t, g)
		effectstest.AnswerChooseCards(t, g, me.ID, rock)
		effectstest.PassPriorityAroundTable(t, g)
		if !me.Hand.Contains(rock) {
			t.Error("the artifact is not in hand")
		}
		if got := effectstest.CounterCount(g, knight, game.CounterPlusOne); got > 0 {
			t.Errorf("a card went to hand but the Knight has %d counters", got)
		}
	})
	t.Run("nothing to take", func(t *testing.T) {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		effectstest.SeedLibrary(me, "A", "B", "C", "D", "E", "F", "G")
		effectstest.TopLibraryCard(me, "Bear", "Creature — Bear")
		knight := effectstest.CastRingCreature(t, g, "Rosecot Knight", "Creature — Human Knight", gapRosecotKnightOracle, 3, 4)
		effectstest.PassPriorityAroundTable(t, g)
		effectstest.PassPriorityAroundTable(t, g)
		if got := effectstest.CounterCount(g, knight, game.CounterPlusOne); got != 1 {
			t.Errorf("no card taken: %d counters, want 1", got)
		}
	})
	t.Run("declined", func(t *testing.T) {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		effectstest.SeedLibrary(me, "A", "B", "C", "D", "E", "F", "G")
		effectstest.TopLibraryCard(me, "Ench", "Enchantment")
		knight := effectstest.CastRingCreature(t, g, "Rosecot Knight", "Creature — Human Knight", gapRosecotKnightOracle, 3, 4)
		effectstest.PassPriorityAroundTable(t, g)
		effectstest.AnswerChooseCards(t, g, me.ID)
		effectstest.PassPriorityAroundTable(t, g)
		if got := effectstest.CounterCount(g, knight, game.CounterPlusOne); got != 1 {
			t.Errorf("declined: %d counters, want 1", got)
		}
	})
	t.Run("the seventh card is not seen", func(t *testing.T) {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		effectstest.SeedLibrary(me, "A", "B")
		rock := effectstest.TopLibraryCard(me, "Rock", "Artifact")
		for i := 0; i < 6; i++ {
			effectstest.TopLibraryCard(me, "Filler", "Creature — Bear")
		}
		knight := effectstest.CastRingCreature(t, g, "Rosecot Knight", "Creature — Human Knight", gapRosecotKnightOracle, 3, 4)
		effectstest.PassPriorityAroundTable(t, g)
		effectstest.PassPriorityAroundTable(t, g)
		if me.Hand.Contains(rock) {
			t.Error("took the seventh card from the top")
		}
		if got := effectstest.CounterCount(g, knight, game.CounterPlusOne); got != 1 {
			t.Errorf("nothing among the top six: %d counters, want 1", got)
		}
	})
}

const gapVoltstormAngelOracle = "836609f7-48b8-47b0-9af6-94150b048251"

// gapPayThenPickMode plays beginning of combat for the Voltstorm Angel:
// pays the energy and answers the reflexive trigger's mode pick.
func gapPayThenPickMode(t *testing.T, g *game.Game, me *game.Player, mode int) {
	t.Helper()
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	if n := effectstest.SettleEnergy(t, g, me.ID, effectstest.ErPlay{Pay: true}); n != 1 {
		t.Fatalf("answered %d energy prompts, want 1", n)
	}
	pick := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceModePick, me.ID)
	if pick == nil {
		t.Fatalf("no mode prompt after paying: %+v", g.PendingChoices)
	}
	if err := g.ResolveModePick(pick.ID, me.ID, []int{mode}); err != nil {
		t.Fatalf("ResolveModePick: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
}

func TestGapVoltstormAngelGainsVigilanceAndLifelinkWhenPaid(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	angel := effectstest.B12Push(g, me.ID, "Voltstorm Angel", "Creature — Angel", gapVoltstormAngelOracle, 4, 4)
	effectstest.SetEnergy(t, g, me, 3)
	gapPayThenPickMode(t, g, me, 0)
	effectstest.AssertKeywords(t, g, angel, "flying", "vigilance", "lifelink")
	if effectstest.EnergyOf(me) != 1 {
		t.Errorf("energy %d, want 1", effectstest.EnergyOf(me))
	}
}

func TestGapVoltstormAngelPumpsTheOtherCreaturesYouControl(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	angel := effectstest.B12Push(g, me.ID, "Voltstorm Angel", "Creature — Angel", gapVoltstormAngelOracle, 4, 4)
	bear := effectstest.PushVanillaCreature(g, me.ID, "Bear", 2, 2)
	theirs := effectstest.PushVanillaCreature(g, opp.ID, "Their Bear", 2, 2)
	effectstest.SetEnergy(t, g, me, 2)
	gapPayThenPickMode(t, g, me, 1)
	if p := effectstest.EffectivePower(t, g, bear); p != 3 {
		t.Errorf("my Bear's power %d, want 3", p)
	}
	if p := effectstest.EffectivePower(t, g, angel); p != 4 {
		t.Errorf("the Angel pumped itself: power %d", p)
	}
	if p := effectstest.EffectivePower(t, g, theirs); p != 2 {
		t.Errorf("an opponent's creature got the pump: power %d", p)
	}
	if effectstest.HasEffectiveKeyword(t, g, angel, "lifelink") {
		t.Error("the other mode also gave lifelink")
	}
}

func TestGapVoltstormAngelDecliningOrBrokeAsksNoModeAndEntersWithThreeEnergy(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	angel := effectstest.B12Push(g, me.ID, "Voltstorm Angel", "Creature — Angel", gapVoltstormAngelOracle, 4, 4)
	effectstest.SetEnergy(t, g, me, 2)
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	effectstest.SettleEnergy(t, g, me.ID, effectstest.ErPlay{Pay: false})
	if effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceModePick, me.ID) != nil {
		t.Error("a mode was asked though the energy was not paid")
	}
	if effectstest.EnergyOf(me) != 2 || effectstest.HasEffectiveKeyword(t, g, angel, "lifelink") {
		t.Errorf("declined: energy %d", effectstest.EnergyOf(me))
	}

	g2 := effectstest.NewCatalogGame(t)
	me2 := g2.Seats[0]
	effectstest.SetEnergy(t, g2, me2, 0)
	effectstest.CastRingCreature(t, g2, "Voltstorm Angel", "Creature — Angel", gapVoltstormAngelOracle, 4, 4)
	effectstest.PassPriorityAroundTable(t, g2)
	if effectstest.EnergyOf(me2) != 3 {
		t.Errorf("entering gave %d energy, want 3", effectstest.EnergyOf(me2))
	}
}

const gapBirthingRitualOracle = "37a5ab64-12e4-4aad-9a09-46a26fe9953b"

// gapRitualLibrary stacks a library with a known top seven (top first:
// a mana value 3 creature, a mana value 4 creature, a land, then four
// fillers) over a deep bottom, and returns the ids of the first two.
func gapRitualLibrary(me *game.Player) (three, four uuid.UUID, topSeven []uuid.UUID) {
	effectstest.SeedLibrary(me, "B1", "B2", "B3")
	var ids []uuid.UUID
	for i := 0; i < 4; i++ {
		ids = append(ids, effectstest.SeedLibraryTop(me, "Filler", "Instant", "{U}"))
	}
	land := effectstest.SeedLibraryTop(me, "Forest", "Basic Land — Forest", "")
	four = effectstest.SeedLibraryTop(me, "Four Drop", "Creature — Ogre", "{3}{G}")
	three = effectstest.SeedLibraryTop(me, "Three Drop", "Creature — Ogre", "{2}{G}")
	topSeven = append([]uuid.UUID{three, four, land}, ids...)
	return
}

func TestGapBirthingRitualSacrificeThenPutAnXOrLessCreatureAndBottomTheRest(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushPermanentForTest(g, me.ID, "Birthing Ritual", gapBirthingRitualOracle, "Enchantment")
	fodder := effectstest.B27Push(g, me.ID, "Two Drop", "Creature — Goblin", "", "{1}{R}", 2, 2, "R")
	three, four, seven := gapRitualLibrary(me)
	effectstest.AdvanceToEndStepOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerOwnPermanents(t, g, me.ID, fodder)
	// Only the mana value 3 creature (1 + 2) is a legal pick; the 4 is not.
	pick := effectstest.LatestChooseCardsFor(g, me.ID)
	if pick == nil {
		t.Fatalf("no card pick: %+v", g.PendingChoices)
	}
	for _, c := range pick.ChooseCards {
		if c == four {
			t.Fatal("the mana value 4 creature was offered for X = 3")
		}
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{three}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(fodder) || !g.Battlefield.Contains(three) {
		t.Errorf("fodder on battlefield %v, the 3-drop on battlefield %v", g.Battlefield.Contains(fodder), g.Battlefield.Contains(three))
	}
	// The rest (six of the seven) are the bottom of the library.
	if n := len(me.Library.Cards); n < 9 {
		t.Fatalf("library %d", n)
	}
	for _, id := range seven {
		if id == three {
			continue
		}
		if z := effectstest.ZoneOfCard(g, id); z != game.ZoneLibrary {
			t.Errorf("a rest card is in %v, want the library", z)
		}
	}
	top := me.Library.Cards[len(me.Library.Cards)-1].InstanceID
	for _, id := range seven {
		if id == top {
			t.Errorf("a looked-at card is still on top of the library")
		}
	}
}

func TestGapBirthingRitualDecliningSacrificesNothingAndBottomsAllSeven(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushPermanentForTest(g, me.ID, "Birthing Ritual", gapBirthingRitualOracle, "Enchantment")
	fodder := effectstest.B27Push(g, me.ID, "Two Drop", "Creature — Goblin", "", "{1}{R}", 2, 2, "R")
	three, _, seven := gapRitualLibrary(me)
	effectstest.AdvanceToEndStepOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerOwnPermanents(t, g, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(fodder) || g.Battlefield.Contains(three) {
		t.Error("declining sacrificed or put something")
	}
	top := me.Library.Cards[len(me.Library.Cards)-1].InstanceID
	for _, id := range seven {
		if id == top {
			t.Error("a looked-at card is still on top of the library")
		}
	}
}

func TestGapBirthingRitualSacrificedTokenMakesXOne(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushPermanentForTest(g, me.ID, "Birthing Ritual", gapBirthingRitualOracle, "Enchantment")
	token := effectstest.B12Creature(g, me.ID, "Goblin", "Token Creature — Goblin", 1, 1)
	effectstest.SeedLibrary(me, "B1", "B2")
	one := effectstest.SeedLibraryTop(me, "One Drop", "Creature — Elf", "{G}")
	two := effectstest.SeedLibraryTop(me, "Two Drop", "Creature — Elf", "{1}{G}")
	effectstest.AdvanceToEndStepOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerOwnPermanents(t, g, me.ID, token)
	pick := effectstest.LatestChooseCardsFor(g, me.ID)
	if pick == nil {
		t.Fatalf("no card pick: %+v", g.PendingChoices)
	}
	for _, c := range pick.ChooseCards {
		if c == two {
			t.Fatal("a mana value 2 creature was offered for X = 1")
		}
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{one}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(one) {
		t.Error("the 1-drop did not enter")
	}
}

func TestGapBirthingRitualNeedsACreatureAtTheEndStep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushPermanentForTest(g, me.ID, "Birthing Ritual", gapBirthingRitualOracle, "Enchantment")
	_, _, seven := gapRitualLibrary(me)
	effectstest.AdvanceToEndStepOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	if len(g.PendingChoices) != 0 {
		t.Errorf("the Ritual asked something with no creature: %+v", g.PendingChoices)
	}
	if top := me.Library.Cards[len(me.Library.Cards)-1].InstanceID; top != seven[0] {
		t.Error("the library was disturbed")
	}
}

const gapHungerTideOracle = "356edd68-7dd5-4ab6-a738-6cb0eb2d4e56"

func TestGapHungerTideRisesMakesAnInsectOnEachOfTheFirstThreeChapters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.CastCatalogSpell(t, g, "The Hunger Tide Rises", "Enchantment — Saga", gapHungerTideOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	for want := 1; want <= 3; want++ {
		if want > 1 {
			effectstest.AdvanceToPrecombatMainOf(t, g, 0)
			effectstest.PassPriorityAroundTable(t, g)
		}
		if n := effectstest.TokensNamedControlledBy(g, "Insect", me.ID); n != want {
			t.Fatalf("after chapter %d: %d Insects, want %d", want, n, want)
		}
	}
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Insect" && (c.Power != 1 || c.Toughness != 1 || len(c.Colors) != 2) {
			t.Errorf("Insect %d/%d %v, want 1/1 black and green", c.Power, c.Toughness, c.Colors)
		}
	}
}

// gapHungerTideChapterIV puts the Saga one chapter short of IV and advances.
func gapHungerTideChapterIV(t *testing.T, g *game.Game) uuid.UUID {
	t.Helper()
	saga := effectstest.CastCatalogSpell(t, g, "The Hunger Tide Rises", "Enchantment — Saga", gapHungerTideOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.SetCountersOn(g, saga, map[string]int{"lore": 3})
	effectstest.AdvanceToPrecombatMainOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	return saga
}

func TestGapHungerTideRisesChapterIVSearchesTheLibraryForXOrLess(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	a := effectstest.B27Push(g, me.ID, "Fodder A", "Creature — Goblin", "", "{1}{R}", 2, 2, "R")
	b := effectstest.B27Push(g, me.ID, "Fodder B", "Creature — Goblin", "", "{1}{R}", 2, 2, "R")
	keep := effectstest.B27Push(g, me.ID, "Keeper", "Creature — Goblin", "", "{1}{R}", 2, 2, "R")
	effectstest.SeedSearchLibrary(me, game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"})
	within := effectstest.PushLibraryCardForTest(me, game.Card{Name: "Two Drop", TypeLine: "Creature — Ogre", ManaCost: "{1}{G}"})
	tooBig := effectstest.PushLibraryCardForTest(me, game.Card{Name: "Three Drop", TypeLine: "Creature — Ogre", ManaCost: "{2}{G}"})
	effectstest.LgKnowLibrary(me)
	gapHungerTideChapterIV(t, g)
	effectstest.AnswerOwnPermanents(t, g, me.ID, a, b)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerSearchZones(t, g, me.ID, game.SearchZoneLibrary)
	c := effectstest.SearchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no search prompt")
	}
	for _, id := range c.SearchCards {
		if id == tooBig {
			t.Fatal("a mana value 3 creature was offered for two sacrificed creatures")
		}
	}
	effectstest.AnswerSearchByID(t, g, me.ID, within)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(a) || g.Battlefield.Contains(b) || !g.Battlefield.Contains(keep) {
		t.Error("the sacrifice did not take exactly the two chosen creatures")
	}
	if !g.Battlefield.Contains(within) || g.Battlefield.Contains(tooBig) {
		t.Errorf("the 2-drop on battlefield %v, the 3-drop on battlefield %v", g.Battlefield.Contains(within), g.Battlefield.Contains(tooBig))
	}
	if effectstest.LibraryKnownTo(me, me.ID) != 0 {
		t.Error("the library was searched but not shuffled")
	}
}

func TestGapHungerTideRisesChapterIVMayTakeFromTheGraveyardAndThenDoesNotShuffle(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	a := effectstest.B27Push(g, me.ID, "Fodder A", "Creature — Goblin", "", "{1}{R}", 2, 2, "R")
	effectstest.SeedSearchLibrary(me, game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"}, game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"}, game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"})
	effectstest.LgKnowLibrary(me)
	ogre := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: ogre, Name: "One Drop Ogre", TypeLine: "Creature — Ogre", ManaCost: "{G}", Owner: me.ID, Controller: me.ID})
	gapHungerTideChapterIV(t, g)
	effectstest.AnswerOwnPermanents(t, g, me.ID, a)
	effectstest.PassPriorityAroundTable(t, g)
	known := effectstest.LibraryKnownTo(me, me.ID)
	if known == 0 {
		t.Fatal("setup: the library has no known cards left to observe a shuffle with")
	}
	effectstest.AnswerSearchZones(t, g, me.ID, game.SearchZoneGraveyard)
	effectstest.AnswerSearchByID(t, g, me.ID, ogre)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(ogre) || me.Graveyard.Contains(ogre) {
		t.Error("the creature card in the graveyard did not enter")
	}
	if effectstest.LibraryKnownTo(me, me.ID) != known {
		t.Error("the library was shuffled though only the graveyard was searched")
	}
}

func TestGapHungerTideRisesChapterIVSacrificingNothingSearchesForManaValueZero(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B27Push(g, me.ID, "Keeper", "Creature — Goblin", "", "{1}{R}", 2, 2, "R")
	effectstest.SeedSearchLibrary(me, game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"})
	one := effectstest.PushLibraryCardForTest(me, game.Card{Name: "One Drop", TypeLine: "Creature — Elf", ManaCost: "{G}"})
	zero := effectstest.PushLibraryCardForTest(me, game.Card{Name: "Zero Drop", TypeLine: "Creature — Elf", ManaCost: ""})
	gapHungerTideChapterIV(t, g)
	effectstest.AnswerOwnPermanents(t, g, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.TakeSearchPrompts(t, g)
	if g.Battlefield.Contains(one) {
		t.Error("a mana value 1 creature was found with nothing sacrificed")
	}
	_ = zero
}
