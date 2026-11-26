package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b359_cards_test.go — S82 OLD batch 359 (#5556, Magic 2015, part 3):
// Mercurial Pretender, Kurkesh, Onakke Ancient, Living Totem, Undergrowth
// Scavenger, Siege Dragon, Blastfire Bolt, Illusory Angel, Thundering Giant and
// Might Makes Right. Each test plays the real catalog entry through the engine.

const (
	oldB359Pretender = "07d0bd24-7731-4abb-9dc6-912aa17943f2"
	oldB359Kurkesh   = "ea7302b3-c9a2-4bc4-81b8-3a8d1d7c3b8b"
	oldB359Totem     = "0a2e2d97-1348-4fa6-ad94-4da23ad8f43b"
	oldB359Scavenger = "42b17f78-d290-40ed-afac-f9362e526999"
	oldB359Dragon    = "206539a6-0e60-4986-bc57-c2b89e08b9f2"
	oldB359Blastfire = "dcddeac9-8de7-4347-a956-522f4f2fc63b"
	oldB359Angel     = "6ee32b5e-3b29-47ff-87f8-913ec05f6b2f"
	oldB359Giant     = "d23329e2-4134-4e49-b9b6-5febef2bce31"
	oldB359Might     = "ec73863f-417b-45d7-a6b7-97a69ce0196a"
)

func TestOldB359EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB359Pretender: "Mercurial Pretender", oldB359Kurkesh: "Kurkesh, Onakke Ancient",
		oldB359Totem: "Living Totem", oldB359Scavenger: "Undergrowth Scavenger",
		oldB359Dragon: "Siege Dragon", oldB359Blastfire: "Blastfire Bolt",
		oldB359Angel: "Illusory Angel", oldB359Giant: "Thundering Giant",
		oldB359Might: "Might Makes Right",
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

func oldB359Attach(g *game.Game, equipment, host uuid.UUID) {
	g.WithWriteLock(func() {
		_ = g.AttachForEffect(equipment, game.TargetRef{Kind: game.TargetCard, ID: host})
	})
}

// --- Mercurial Pretender -------------------------------------------------------

func TestOldB359PretenderCopiesAYourCreatureAndCanReturnToHand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	mine := effectstest.SeedCopyableCreature(g, me.ID, "Grizzly Bears", "Creature — Bear", 2, 2)
	pretender := effectstest.CastCatalogSpell(t, g, "Mercurial Pretender", "Creature — Shapeshifter", oldB359Pretender, nil)
	effectstest.ResolveWithCopyChoice(t, g, mine)
	if got := effectstest.CopyBattlefieldCard(t, g, pretender); got.Name != "Grizzly Bears" {
		t.Fatalf("the Pretender is %q, want a copy of Grizzly Bears", got.Name)
	}
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == pretender {
				g.Battlefield.Cards[i].SummonedThisTurn = false
			}
		}
	})
	if err := g.ActivateCatalogAbility(me.ID, pretender, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("activated with an empty pool")
	}
	effectstest.B06AddMana(me, "U", "U", "U", "U")
	if err := g.ActivateCatalogAbility(me.ID, pretender, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(pretender) {
		t.Error("the Pretender did not return to its owner's hand")
	}
}

func TestOldB359PretenderCannotCopyAnOpponentsCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.SeedCopyableCreature(g, me.ID, "Grizzly Bears", "Creature — Bear", 2, 2)
	theirs := effectstest.SeedCopyableCreature(g, opp.ID, "Runeclaw Bear", "Creature — Bear", 2, 2)
	pretender := effectstest.CastCatalogSpell(t, g, "Mercurial Pretender", "Creature — Shapeshifter", oldB359Pretender, nil)
	effectstest.ResolveWithCopyChoice(t, g, theirs)
	g.RunStateChecksForTest()
	if g.Battlefield.Contains(pretender) && effectstest.CopyBattlefieldCard(t, g, pretender).Name == "Runeclaw Bear" {
		t.Error("the Pretender copied a creature its controller doesn't control")
	}
}

func TestOldB359PretenderDeclinedIsAZeroZeroThatDies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.SeedCopyableCreature(g, me.ID, "Grizzly Bears", "Creature — Bear", 2, 2)
	pretender := effectstest.CastCatalogSpell(t, g, "Mercurial Pretender", "Creature — Shapeshifter", oldB359Pretender, nil)
	effectstest.ResolveWithCopyChoice(t, g, uuid.Nil)
	g.RunStateChecksForTest()
	if g.Battlefield.Contains(pretender) && effectstest.CopyBattlefieldCard(t, g, pretender).Name == "Grizzly Bears" {
		t.Error("a declined copy still copied a creature")
	}
}

// --- Kurkesh, Onakke Ancient ---------------------------------------------------

func TestOldB359KurkeshPaysRToCopyAnArtifactAbility(t *testing.T) {
	for _, pay := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		effectstest.PushCatalogPermanent(g, me.ID, "Kurkesh, Onakke Ancient", "Legendary Creature — Ogre Spirit", oldB359Kurkesh, false)
		monolith := effectstest.PushCatalogPermanent(g, me.ID, "Basalt Monolith", "Artifact", effectstest.AcBasaltMonolithOracle, false)
		effectstest.FloatForTest(g, me, "CCCCCR")
		if err := g.ActivateCatalogAbility(me.ID, monolith, 0, game.ActivateAbilityParams{}); err != nil {
			t.Fatalf("activate: %v", err)
		}
		if effectstest.AcItemOnStack(g, "Kurkesh, Onakke Ancient") == nil {
			t.Fatal("activating an artifact's ability did not trigger Kurkesh")
		}
		effectstest.PassPriorityAroundTable(t, g)
		prompt := effectstest.LatestChoiceOfKind(g, game.PendingChoicePayUnless)
		if prompt == nil {
			t.Fatal("Kurkesh did not offer the {R}")
		}
		if err := g.ResolvePayUnless(prompt.ID, me.ID, pay); err != nil {
			t.Fatalf("ResolvePayUnless: %v", err)
		}
		want := 0
		if pay {
			want = 1
		}
		if n := effectstest.AcCopiesOnStack(g); n != want {
			t.Errorf("pay=%v: %d copies on the stack, want %d", pay, n, want)
		}
	}
}

func TestOldB359KurkeshIgnoresAbilitiesOfNonArtifacts(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushCatalogPermanent(g, me.ID, "Kurkesh, Onakke Ancient", "Legendary Creature — Ogre Spirit", oldB359Kurkesh, false)
	lizard := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Bellows Lizard", TypeLine: "Creature — Lizard", OracleID: "667dae0b-5006-4e38-be80-eae79c2a2a95",
		Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID,
	})
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{R}{R}{R}"); err != nil {
		t.Fatal(err)
	}
	if err := g.ActivateCatalogAbility(me.ID, lizard, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if effectstest.AcItemOnStack(g, "Kurkesh, Onakke Ancient") != nil {
		t.Error("Kurkesh triggered off a creature's ability")
	}
}

// --- Living Totem --------------------------------------------------------------

func TestOldB359LivingTotemMayPutACounterOnAnotherCreature(t *testing.T) {
	for _, yes := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
		totem := effectstest.CastCatalogSpell(t, g, "Living Totem", "Creature — Plant Elemental", oldB359Totem, nil)
		effectstest.PassPriorityAroundTable(t, g)
		if p := effectstest.LatestPickTarget(g, me.ID); p != nil {
			if effectstest.HasID(p.PickTargetCards, totem) {
				t.Errorf("the Totem is a legal target: %v", p.PickTargetCards)
			}
			effectstest.PickCard(t, g, me.ID, bear)
		}
		effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, yes)
		effectstest.PassPriorityAroundTable(t, g)
		want := 2
		if yes {
			want = 3
		}
		if got := effectstest.EffectivePowerWithCounters(t, g, bear); got != want {
			t.Errorf("yes=%v: the Bear's power = %d, want %d", yes, got, want)
		}
	}
}

// --- Undergrowth Scavenger -----------------------------------------------------

func TestOldB359ScavengerEntersWithACounterPerCreatureCardInAllGraveyards(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogGraveyardCard(me, "Mine", "Creature — Bear", "", 2, 2)
	effectstest.PushCatalogGraveyardCard(opp, "Theirs", "Creature — Bear", "", 2, 2)
	effectstest.PushCatalogGraveyardCard(opp, "Their Spell", "Sorcery", "", 0, 0)
	scavenger := effectstest.CastCatalogSpell(t, g, "Undergrowth Scavenger", "Creature — Fungus Horror", oldB359Scavenger, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.EffectivePowerWithCounters(t, g, scavenger); got != 2 {
		t.Errorf("the Scavenger's power = %d, want 2 (two creature cards in graveyards)", got)
	}
}

// --- Siege Dragon --------------------------------------------------------------

func TestOldB359SiegeDragonDestroysOpposingWallsWhenItEnters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	theirWall := effectstest.B12Creature(g, opp.ID, "Their Wall", "Creature — Wall", 0, 4)
	myWall := effectstest.B12Creature(g, me.ID, "My Wall", "Creature — Wall", 0, 4)
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Siege Dragon", "Creature — Dragon", oldB359Dragon, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(theirWall) {
		t.Error("an opposing Wall survived")
	}
	if !g.Battlefield.Contains(myWall) || !g.Battlefield.Contains(bear) {
		t.Error("my own Wall or a non-Wall was destroyed")
	}
}

func TestOldB359SiegeDragonAttackDamagesNonflyersUnlessTheDefenderHasAWall(t *testing.T) {
	for _, withWall := range []bool{false, true} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
		dragon := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
			InstanceID: uuid.New(), Name: "Siege Dragon", TypeLine: "Creature — Dragon", OracleID: oldB359Dragon,
			Power: 5, Toughness: 5, Keywords: []string{"flying"}, Owner: me.ID, Controller: me.ID,
		})
		bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 5)
		flyer := effectstest.Pr7bCreature(g, opp.ID, "Flyer", 2, "{2}", nil, "flying")
		if withWall {
			effectstest.B12Creature(g, opp.ID, "Wall", "Creature — Wall", 0, 5)
		}
		effectstest.DeclareAttack(t, g, opp.ID, dragon)
		effectstest.LockInAttacks(t, g)
		effectstest.PassPriorityAroundTable(t, g)
		wantBear := 2
		if withWall {
			wantBear = 0
		}
		if got := effectstest.DamageOn(t, g, bear); got != wantBear {
			t.Errorf("wall=%v: the Bear has %d damage, want %d", withWall, got, wantBear)
		}
		if got := effectstest.DamageOn(t, g, flyer); got != 0 {
			t.Errorf("wall=%v: a flyer took %d damage", withWall, got)
		}
	}
}

// --- Blastfire Bolt ------------------------------------------------------------

func TestOldB359BlastfireBoltDamagesAndDestroysItsEquipment(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	target := effectstest.B12Creature(g, opp.ID, "Target", "Creature — Bear", 2, 7)
	other := effectstest.B12Creature(g, opp.ID, "Other", "Creature — Bear", 2, 7)
	sword := effectstest.PushCatalogPermanent(g, opp.ID, "Sword", "Artifact — Equipment", "", false)
	spare := effectstest.PushCatalogPermanent(g, opp.ID, "Spare Sword", "Artifact — Equipment", "", false)
	oldB359Attach(g, sword, target)
	oldB359Attach(g, spare, other)
	_ = me
	effectstest.CastCatalogSpell(t, g, "Blastfire Bolt", "Instant", oldB359Blastfire, effectstest.CardRefs(target))
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.DamageOn(t, g, target); got != 5 {
		t.Errorf("the target has %d damage, want 5", got)
	}
	if g.Battlefield.Contains(sword) {
		t.Error("the Equipment attached to the target survived")
	}
	if !g.Battlefield.Contains(spare) {
		t.Error("an Equipment attached to another creature was destroyed")
	}
}

// --- Illusory Angel ------------------------------------------------------------

func TestOldB359IllusoryAngelCanBeCastOnlyAfterAnotherSpell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	if err := effectstest.CastCatalogSpellErrStrict(t, g, "Illusory Angel", "Creature — Angel Illusion", oldB359Angel, nil); err == nil {
		t.Fatal("cast with no other spell this turn")
	}
	effectstest.CastCatalogSpell(t, g, "Other Spell", "Sorcery", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if err := effectstest.CastCatalogSpellErrStrict(t, g, "Illusory Angel", "Creature — Angel Illusion", oldB359Angel, nil); err != nil {
		t.Fatalf("refused after casting another spell: %v", err)
	}
}

// --- Thundering Giant ----------------------------------------------------------

func TestOldB359ThunderingGiantHasHaste(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	giant := effectstest.PushCatalogPermanent(g, me.ID, "Thundering Giant", "Creature — Giant", oldB359Giant, true)
	if !effectstest.HasKeywordOf(g, giant, "haste") {
		t.Error("the Giant has no haste")
	}
}

// --- Might Makes Right ---------------------------------------------------------

func TestOldB359MightMakesRightStealsOnlyWhileYouHoldTheGreatestPower(t *testing.T) {
	for _, strongest := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
		effectstest.PushCatalogPermanent(g, me.ID, "Might Makes Right", "Enchantment", oldB359Might, false)
		effectstest.B12Creature(g, me.ID, "Big", "Creature — Giant", 5, 5)
		theirs := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
			InstanceID: uuid.New(), Name: "Theirs", TypeLine: "Creature — Bear", Power: 3, Toughness: 3,
			Tapped: true, Owner: opp.ID, Controller: opp.ID,
		})
		if !strongest {
			effectstest.B12Creature(g, opp.ID, "Bigger", "Creature — Giant", 6, 6)
		}
		effectstest.AdvanceTo(t, g, game.StepBeginCombat)
		if p := effectstest.LatestPickTarget(g, me.ID); p != nil {
			effectstest.PickCard(t, g, me.ID, theirs)
		}
		effectstest.PassPriorityAroundTable(t, g)
		got := effectstest.ControllerOf(t, g, theirs)
		if strongest && got != me.ID {
			t.Errorf("strongest: the creature is controlled by %s, want me", got)
		}
		if !strongest && got != opp.ID {
			t.Errorf("not strongest: the creature changed hands")
		}
	}
}
