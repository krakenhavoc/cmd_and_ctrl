package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b805_test.go — Mirage cards with counters, S82 OLD batch 805, played
// through the real catalog entries.

const (
	oldB805Asmira    = "a0397f06-d557-4c9f-af89-238ab38ada3e"
	oldB805Purraj    = "ae86965c-3330-48bf-abf7-3142c61099ea"
	oldB805Growth    = "b3235d21-ce52-4a99-a2fe-99b0fe7cc8bc"
	oldB805MireShade = "b77a7f4a-1b3a-440e-977c-1ba39e450581"
	oldB805Vortex    = "bcb2f489-9d58-424d-96c1-a124a8d25a14"
	oldB805Mob       = "c56f0f4c-c85a-4931-b22a-42dc5d1dc7dd"
	oldB805Spirit    = "ce785306-eaf5-42b3-9840-884cd690ae0a"
	oldB805Wall      = "e6f5f80c-2620-4558-8a5f-8fc36cea4793"
)

func TestOldB805CardsAreFullWithNoCaveats(t *testing.T) {
	for id, name := range map[string]string{
		oldB805Asmira: "Asmira, Holy Avenger", oldB805Purraj: "Purraj of Urborg",
		oldB805Growth: "Malignant Growth", oldB805MireShade: "Mire Shade",
		oldB805Vortex: "Energy Vortex", oldB805Mob: "Zombie Mob",
		oldB805Spirit: "Discordant Spirit", oldB805Wall: "Wall of Resistance",
	} {
		spec, ok := effects.Lookup(id)
		if !ok || spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: registered %v as %q (%v, caveats %v)", name, ok, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

// oldB805Owned seeds a creature `owner` owns and `controller` controls.
func oldB805Owned(g *game.Game, owner, controller uuid.UUID) uuid.UUID {
	return effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Borrowed Bear", TypeLine: "Creature — Bear",
		Power: 2, Toughness: 2, Owner: owner, Controller: controller,
	})
}

func TestOldB805AsmiraGrowsPerCreatureOfYoursPutIntoYourGraveyardByOwner(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	asmira := effectstest.B12Push(g, me.ID, "Asmira, Holy Avenger", "Legendary Creature — Human Cleric", oldB805Asmira, 2, 3)
	if !effectstest.HasEffectiveKeyword(t, g, asmira, "flying") {
		t.Error("Asmira does not fly")
	}
	for i := 0; i < 2; i++ {
		id := effectstest.B12Push(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
		effectstest.KillCreature(t, g, me.ID, id)
	}
	// A creature the opponent owns but I control goes to THEIR graveyard; one
	// I own but the opponent controls goes to mine.
	effectstest.KillCreature(t, g, opp.ID, oldB805Owned(g, opp.ID, me.ID))
	effectstest.KillCreature(t, g, me.ID, oldB805Owned(g, me.ID, opp.ID))
	effectstest.KillCreature(t, g, opp.ID, effectstest.B12Push(g, opp.ID, "Opposing Bear", "Creature — Bear", "", 2, 2))
	effectstest.AdvanceToEndStep(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountersOn(g, asmira, game.CounterPlusOne); n != 3 {
		t.Errorf("+1/+1 counters = %d, want 3 (two of mine and the one I own under their control)", n)
	}
}

func TestOldB805AsmiraTriggersOnAnOpponentsEndStepToo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	asmira := effectstest.B12Push(g, me.ID, "Asmira, Holy Avenger", "Legendary Creature — Human Cleric", oldB805Asmira, 2, 3)
	effectstest.AdvanceToEndStepOf(t, g, 1)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.KillCreature(t, g, me.ID, effectstest.B12Push(g, me.ID, "Bear", "Creature — Bear", "", 2, 2))
	effectstest.AdvanceToEndStepOf(t, g, 2)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountersOn(g, asmira, game.CounterPlusOne); n != 0 {
		t.Errorf("counters = %d before any creature of mine has died this turn, want 0", n)
	}
	if !g.Battlefield.Contains(asmira) {
		t.Fatal("Asmira is gone")
	}
}

func TestOldB805PurrajHasFirstStrikeOnlyWhileAttacking(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	purraj := effectstest.B12Push(g, me.ID, "Purraj of Urborg", "Legendary Creature — Cat Warrior", oldB805Purraj, 2, 3)
	if effectstest.HasEffectiveKeyword(t, g, purraj, "first strike") {
		t.Fatal("Purraj has first strike while not attacking")
	}
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if _, err := g.DeclareAttackersWith([]game.AttackDeclaration{{Attacker: purraj, Target: g.Seats[1].ID}}, game.DeclareAttackersParams{}); err != nil {
		t.Fatalf("declare attackers: %v", err)
	}
	if !effectstest.HasEffectiveKeyword(t, g, purraj, "first strike") {
		t.Error("Purraj does not have first strike while attacking")
	}
}

func TestOldB805PurrajGrowsOnlyOnABlackSpellAndOnlyWhenBPaid(t *testing.T) {
	for _, tc := range []struct {
		name  string
		color string
		cost  string
		pay   bool
		want  int
	}{
		{"black spell, paid", "B", "{B}", true, 1},
		{"black spell, declined", "B", "{B}", false, 0},
		{"red spell", "R", "{R}", true, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			effectstest.AdvanceToMain(t, g)
			purraj := effectstest.B12Push(g, opp.ID, "Purraj of Urborg", "Legendary Creature — Cat Warrior", oldB805Purraj, 2, 3)
			spell := effectstest.DstHand(me, "Spell", "Sorcery", tc.cost, tc.color)
			if err := g.CastSpell(me.ID, spell, game.CastSpellParams{}); err != nil {
				t.Fatalf("cast: %v", err)
			}
			if tc.color == "R" {
				if effectstest.TriggerOnStackFrom(g, "Purraj of Urborg") {
					t.Fatal("a red spell triggered Purraj")
				}
				return
			}
			effectstest.PassPriorityAroundTable(t, g)
			effectstest.B06AddMana(opp, "B")
			effectstest.AnswerPayUnless(t, g, opp.ID, tc.pay)
			effectstest.PassPriorityAroundTable(t, g)
			if n := effectstest.CountersOn(g, purraj, game.CounterPlusOne); n != tc.want {
				t.Errorf("counters = %d, want %d", n, tc.want)
			}
		})
	}
}

func TestOldB805MireShadeSacrificesASwampAtSorcerySpeedOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	shade := effectstest.B12Push(g, me.ID, "Mire Shade", "Creature — Shade", oldB805MireShade, 1, 1)
	swamp := effectstest.B12Permanent(g, me.ID, "Swamp", "Basic Land — Swamp")
	forest := effectstest.B12Permanent(g, me.ID, "Forest", "Basic Land — Forest")
	effectstest.B06AddMana(me, "B")
	effectstest.B16Activate(t, g, me.ID, shade, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{swamp}})
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountersOn(g, shade, game.CounterPlusOne); n != 1 {
		t.Errorf("counters = %d, want 1", n)
	}
	if g.Battlefield.Contains(swamp) {
		t.Error("the Swamp was not sacrificed")
	}
	// No Swamp left, so it can't be activated again; a Forest does not pay it.
	effectstest.B06AddMana(me, "B")
	if err := g.ActivateCatalogAbility(me.ID, shade, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{forest}}); err == nil {
		t.Error("a Forest paid for the Swamp sacrifice")
	}
}

func TestOldB805MireShadeCannotActivateOutsideSorcerySpeed(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	shade := effectstest.B12Push(g, me.ID, "Mire Shade", "Creature — Shade", oldB805MireShade, 1, 1)
	swamp := effectstest.B12Permanent(g, me.ID, "Swamp", "Basic Land — Swamp")
	effectstest.B06AddMana(me, "B")
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.ActivateCatalogAbility(me.ID, shade, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{swamp}}); err == nil {
		t.Error("activated in combat")
	}
}

func TestOldB805WallOfResistanceGrowsOnlyAfterTakingDamage(t *testing.T) {
	for _, damaged := range []bool{false, true} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		effectstest.AdvanceToMain(t, g)
		wall := effectstest.B12Push(g, me.ID, "Wall of Resistance", "Creature — Wall", oldB805Wall, 0, 3)
		if !effectstest.HasEffectiveKeyword(t, g, wall, "defender") || !effectstest.HasEffectiveKeyword(t, g, wall, "flying") {
			t.Fatal("the Wall lacks defender or flying")
		}
		if damaged {
			pinger := effectstest.B12Push(g, g.Seats[1].ID, "Pinger", "Creature — Test", "", 1, 1)
			g.WithWriteLock(func() { _ = g.DealDamageToCreatureForEffect(pinger, wall, 1) })
		}
		effectstest.AdvanceToEndStep(t, g)
		effectstest.PassPriorityAroundTable(t, g)
		want := 0
		if damaged {
			want = 1
		}
		if n := effectstest.CountersOn(g, wall, "+0/+1"); n != want {
			t.Errorf("damaged=%v: +0/+1 counters = %d, want %d", damaged, n, want)
		}
	}
}

func TestOldB805ZombieMobEntersWithACounterPerCreatureCardThenExilesThem(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.GraveyardCreature(me, "Dead One", "Bear", 2, 2)
	effectstest.GraveyardCreature(me, "Dead Two", "Elf", 1, 1)
	effectstest.GraveyardCreature(me, "Dead Three", "Cat", 3, 3)
	effectstest.GraveyardSpell(me, "Dead Spell", "{1}{R}")
	effectstest.GraveyardCreature(g.Seats[1], "Their Dead", "Bear", 2, 2)
	mob := effectstest.CastCreatureCard(t, g, "Zombie Mob", "Creature — Zombie", oldB805Mob, 2, 0)
	if n := effectstest.CountersOn(g, mob, game.CounterPlusOne); n != 3 {
		t.Errorf("+1/+1 counters = %d, want 3 (the creature cards in my graveyard)", n)
	}
	if got := effectstest.EffectivePowerWithCounters(t, g, mob); got != 5 {
		t.Errorf("power = %d, want 5", got)
	}
	for _, c := range me.Graveyard.Cards {
		if c.IsCreature() {
			t.Errorf("creature card %q is still in my graveyard", c.Name)
		}
	}
	if effectstest.GraveyardSize(g, me.ID) != 1 {
		t.Errorf("graveyard holds %d cards, want only the non-creature one", effectstest.GraveyardSize(g, me.ID))
	}
	if effectstest.GraveyardSize(g, g.Seats[1].ID) != 1 {
		t.Error("the opponent's graveyard was touched")
	}
}

func TestOldB805DiscordantSpiritGrowsOnAnOpponentsEndStepByTheDamageYouTook(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMainOf(t, g, 1)
	spirit := effectstest.B12Push(g, me.ID, "Discordant Spirit", "Creature — Spirit", oldB805Spirit, 2, 2)
	src := effectstest.B12Push(g, g.Seats[1].ID, "Pinger", "Creature — Test", "", 1, 1)
	g.WithWriteLock(func() { _ = g.DealDamageToPlayerForEffect(src, me.ID, 3) })
	g.WithWriteLock(func() { _ = g.DealDamageToPlayerForEffect(src, me.ID, 2) })
	effectstest.AdvanceToEndStepOf(t, g, 1)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountersOn(g, spirit, game.CounterPlusOne); n != 5 {
		t.Errorf("+1/+1 counters = %d, want 5", n)
	}
}

func TestOldB805DiscordantSpiritGetsNothingWithoutDamageAndLosesCountersOnYourEndStep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMainOf(t, g, 1)
	spirit := effectstest.B12Push(g, me.ID, "Discordant Spirit", "Creature — Spirit", oldB805Spirit, 2, 2)
	effectstest.AdvanceToEndStepOf(t, g, 1)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountersOn(g, spirit, game.CounterPlusOne); n != 0 {
		t.Errorf("counters = %d with no damage taken, want 0", n)
	}
	// On my own turn the damage does not count, and the end step clears the counters.
	effectstest.AddCountersForTest(g, spirit, game.CounterPlusOne, 4)
	effectstest.AdvanceToMainOf(t, g, 0)
	src := effectstest.B12Push(g, g.Seats[1].ID, "Pinger", "Creature — Test", "", 1, 1)
	g.WithWriteLock(func() { _ = g.DealDamageToPlayerForEffect(src, me.ID, 3) })
	effectstest.AdvanceToEndStep(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountersOn(g, spirit, game.CounterPlusOne); n != 0 {
		t.Errorf("counters = %d after my end step, want 0", n)
	}
}

func TestOldB805MalignantGrowthTaxesPutsAGrowthCounterAndSacrificesIfUnpaid(t *testing.T) {
	for _, pay := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[1]
		growth := effectstest.B12Push(g, me.ID, "Malignant Growth", "Enchantment", oldB805Growth, 0, 0)
		effectstest.AdvanceToUpkeepOf(t, g, 1)
		effectstest.AnswerAnyTriggerOrderPrompt(t, g, me.ID)
		effectstest.PassPriorityAroundTable(t, g)
		prompt := effectstest.RemoraUpkeepPrompt(t, g, me.ID)
		if prompt.PayCost != "{1}" {
			t.Fatalf("cost = %q, want {1}", prompt.PayCost)
		}
		if pay {
			effectstest.B06AddMana(me, "G")
		}
		if err := g.ResolvePayUnless(prompt.ID, me.ID, pay); err != nil {
			t.Fatalf("answer: %v", err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		if got := g.Battlefield.Contains(growth); got != pay {
			t.Fatalf("pay=%v: on the battlefield = %v", pay, got)
		}
		if pay {
			if n := effectstest.CountersOn(g, growth, "growth"); n != 1 {
				t.Errorf("growth counters = %d, want 1", n)
			}
			if n := effectstest.AgeCountersOn(t, g, growth); n != 1 {
				t.Errorf("age counters = %d, want 1", n)
			}
		}
	}
}

func TestOldB805MalignantGrowthDrawsAnOpponentExtraCardsAndHurtsThem(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, other := g.Seats[0], g.Seats[1], g.Seats[2]
	growth := effectstest.B12Push(g, me.ID, "Malignant Growth", "Enchantment", oldB805Growth, 0, 0)
	effectstest.AddCountersForTest(g, growth, "growth", 2)
	effectstest.Msc18Library(opp, 12)
	effectstest.Msc18Library(other, 12)
	effectstest.AdvanceToDrawStepOf(t, g, 1)
	hand, life := opp.Hand.Size(), opp.Life
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Hand.Size() != hand+2 {
		t.Errorf("hand %d, want %d (two additional cards)", opp.Hand.Size(), hand+2)
	}
	if opp.Life != life-2 {
		t.Errorf("life %d, want %d", opp.Life, life-2)
	}
	// Each opponent's draw step triggers it, not only the first one's.
	effectstest.AdvanceToDrawStepOf(t, g, 2)
	hand, life = other.Hand.Size(), other.Life
	effectstest.PassPriorityAroundTable(t, g)
	if other.Hand.Size() != hand+2 || other.Life != life-2 {
		t.Errorf("second opponent: hand %d (want %d), life %d (want %d)", other.Hand.Size(), hand+2, other.Life, life-2)
	}
}

func oldB805ChooseOpponent(t *testing.T, g *game.Game, chosen uuid.UUID) uuid.UUID {
	t.Helper()
	me := g.Seats[0]
	id := effectstest.CastCatalogSpell(t, g, "Energy Vortex", "Enchantment", oldB805Vortex, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerChosenPlayer(t, g, me.ID, chosen)
	return id
}

func TestOldB805EnergyVortexDamagesTheChosenPlayerUnlessTheyPayPerCounter(t *testing.T) {
	for _, tc := range []struct {
		name     string
		counters int
		pay      bool
		lost     int
	}{
		{"declined", 2, false, 3},
		{"paid", 2, true, 0},
		{"no counters", 0, false, 3},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			chosen, other := g.Seats[1], g.Seats[2]
			vortex := oldB805ChooseOpponent(t, g, chosen.ID)
			if tc.counters > 0 {
				effectstest.AddCountersForTest(g, vortex, "vortex", tc.counters)
			}
			effectstest.AdvanceToUpkeepOf(t, g, 1)
			start := chosen.Life
			effectstest.PassPriorityAroundTable(t, g)
			if tc.counters > 0 {
				prompt := effectstest.RemoraUpkeepPrompt(t, g, chosen.ID)
				if prompt.PayCost != "{1}{1}" && prompt.PayCost != "{2}" {
					t.Fatalf("cost = %q, want {2}", prompt.PayCost)
				}
				if tc.pay {
					effectstest.B06AddMana(chosen, "C", "C")
				}
				if err := g.ResolvePayUnless(prompt.ID, chosen.ID, tc.pay); err != nil {
					t.Fatalf("answer: %v", err)
				}
				effectstest.PassPriorityAroundTable(t, g)
			}
			if chosen.Life != start-tc.lost {
				t.Errorf("life %d -> %d, want -%d", start, chosen.Life, tc.lost)
			}
			// Another opponent's upkeep is untouched.
			effectstest.AdvanceToUpkeepOf(t, g, 2)
			startOther := other.Life
			effectstest.PassPriorityAroundTable(t, g)
			if other.Life != startOther {
				t.Errorf("an unchosen opponent lost %d", startOther-other.Life)
			}
		})
	}
}

func TestOldB805EnergyVortexCountersAreAddedOnlyInYourUpkeepAndWipedNextOne(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	vortex := effectstest.B12Push(g, me.ID, "Energy Vortex", "Enchantment", oldB805Vortex, 0, 0)
	effectstest.AdvanceToMain(t, g)
	effectstest.B06AddMana(me, "C", "C", "C")
	if err := g.ActivateCatalogAbility(me.ID, vortex, 0, game.ActivateAbilityParams{XValue: 3}); err == nil {
		t.Fatal("activated outside my upkeep")
	}
	effectstest.AdvanceToUpkeepOf(t, g, 1)
	effectstest.AdvanceToUpkeepOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g) // the removal trigger resolves first
	effectstest.B06AddMana(me, "C", "C", "C")
	if err := g.ActivateCatalogAbility(me.ID, vortex, 0, game.ActivateAbilityParams{XValue: 3}); err != nil {
		t.Fatalf("activate in my upkeep: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountersOn(g, vortex, "vortex"); n != 3 {
		t.Fatalf("vortex counters = %d, want 3", n)
	}
	effectstest.AdvanceToUpkeepOf(t, g, 1)
	effectstest.AdvanceToUpkeepOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountersOn(g, vortex, "vortex"); n != 0 {
		t.Errorf("vortex counters = %d after my next upkeep, want 0", n)
	}
}
