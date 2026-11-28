package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b855_cards_test.go — Fallen Empires, Auras, other creatures, card
// draw, damage, instants and lands, batch 855 (#5637, S82 OLD). Every test plays
// the real catalog entry through the engine.

const (
	oldB855Retainer = "81a47019-1343-4a63-893a-0d8ba5e848d5"
	oldB855Mindstab = "43abe1f3-7b92-411a-bf93-e158f949685c"
	oldB855Derelor  = "43f2969e-46eb-4fd4-8b2d-ff81aafcc24a"
	oldB855Brassclw = "c8ce324f-aa58-48f9-9f56-435e49dfff0a"
	oldB855Ring     = "45964368-ea9e-415f-9faf-870c7bf550d4"
	oldB855Conch    = "b019cc20-d32d-4cdf-80cc-eafb3f78db88"
	oldB855Catapult = "2aae77ed-1a91-41f5-8dc2-e353d63a7586"
	oldB855Cloud    = "36d55084-4696-4d16-ab3a-b00f1456603f"
	oldB855Balm     = "45a40b4c-1d37-4d87-adf1-abf0ea4c86b4"
	oldB855Vale     = "76695b15-d0ba-41eb-85f1-52ba5d14b8ba"
)

func TestOldB855CardsAreRegisteredAndWhole(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB855Retainer: "Thrull Retainer", oldB855Mindstab: "Mindstab Thrull", oldB855Derelor: "Derelor",
		oldB855Brassclw: "Brassclaw Orcs", oldB855Ring: "Ring of Renewal", oldB855Conch: "Conch Horn",
		oldB855Catapult: "Dwarven Catapult", oldB855Cloud: "Spore Cloud", oldB855Balm: "Balm of Restoration",
		oldB855Vale: "Rainbow Vale",
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

func TestOldB855ThrullRetainerPumpsAndSacrificesForARegeneration(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	host := effectstest.B12Creature(g, me.ID, "Host", "Creature — Ogre", 2, 2)
	aura := effectstest.Enchant(t, g, "Thrull Retainer", oldB855Retainer, host)
	if p, tough := effectstest.EffectivePTOf(t, g, host); p != 3 || tough != 3 {
		t.Errorf("the host is %d/%d, want 3/3", p, tough)
	}
	effectstest.AdvanceToMain(t, g)
	effectstest.P7Activate(t, g, me, aura, 0, game.ActivateAbilityParams{Strict: true})
	if g.Battlefield.Contains(aura) {
		t.Fatal("the Aura was not sacrificed")
	}
	if p, tough := effectstest.EffectivePTOf(t, g, host); p != 2 || tough != 2 {
		t.Errorf("the host is %d/%d after the Aura left, want 2/2", p, tough)
	}
	effectstest.DestroyIt(t, g, host)
	if !g.Battlefield.Contains(host) {
		t.Error("the creature was destroyed despite its regeneration shield")
	}
}

func TestOldB855MindstabThrullMakesTheDefenderDiscardThree(t *testing.T) {
	for _, yes := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		thrull := effectstest.B12Push(g, me.ID, "Mindstab Thrull", "Creature — Thrull", oldB855Mindstab, 2, 2)
		for i := 0; i < 5; i++ {
			effectstest.HandCard(opp, "Spare", "Sorcery")
		}
		before := opp.Hand.Size()
		effectstest.DeclareAttack(t, g, opp.ID, thrull)
		effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
		effectstest.LockInBlocks(t, g)
		effectstest.PassPriorityAroundTable(t, g)
		effectstest.AnswerMayChoice(t, g, me.ID, yes)
		if yes {
			// The defending player chooses what to discard.
			for i := 0; i < 3; i++ {
				p := effectstest.LatestChooseCardsFor(g, opp.ID)
				if p == nil {
					break
				}
				if err := g.ResolveChooseCards(p.ID, opp.ID, []uuid.UUID{opp.Hand.Cards[0].InstanceID}); err != nil {
					t.Fatalf("ResolveChooseCards: %v", err)
				}
			}
		}
		effectstest.PassPriorityAroundTable(t, g)
		if got := g.Battlefield.Contains(thrull); got == yes {
			t.Errorf("answered %v: the Thrull on the battlefield = %v", yes, got)
		}
		want := before
		if yes {
			want = before - 3
		}
		if got := opp.Hand.Size(); got != want {
			t.Errorf("answered %v: the defender's hand is %d, want %d", yes, got, want)
		}
	}
}

func TestOldB855DerelorTaxesYourBlackSpellsOnly(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, cost, oracle, mana string
		colors                             []string
		opponent, refused                  bool
	}{
		{"Dark Ritual", "Instant", "{B}", "", "B", []string{"B"}, false, true},
		{"Dark Ritual", "Instant", "{B}", "", "BB", []string{"B"}, false, false},
		{"Giant Growth", "Instant", "{G}", "", "G", []string{"G"}, false, false},
		{"Dark Ritual", "Instant", "{B}", "", "B", []string{"B"}, true, false},
	} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
		effectstest.PushCatalogPermanent(g, me.ID, "Derelor", "Creature — Thrull", oldB855Derelor, false)
		effectstest.AdvanceToMain(t, g)
		caster := me
		if tc.opponent {
			caster = opp
		}
		effectstest.FloatForTest(g, caster, tc.mana)
		id := uuid.New()
		g.WithWriteLock(func() {
			caster.Hand.PushTop(game.Card{InstanceID: id, Name: tc.name, TypeLine: tc.typeLine, ManaCost: tc.cost, Colors: tc.colors, Owner: caster.ID, Controller: caster.ID})
		})
		err := g.CastSpell(caster.ID, id, game.CastSpellParams{Strict: true})
		if (err != nil) != tc.refused {
			t.Errorf("%s with %s (opponent %v): error %v, want refused %v", tc.name, tc.mana, tc.opponent, err, tc.refused)
		}
	}
}

func TestOldB855BrassclawOrcsCantBlockPowerTwoOrMore(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	weak := effectstest.B12Creature(g, me.ID, "Weak", "Creature — Bear", 1, 1)
	strong := effectstest.B12Creature(g, me.ID, "Strong", "Creature — Bear", 2, 2)
	orcs := effectstest.B12Push(g, opp.ID, "Brassclaw Orcs", "Creature — Orc", oldB855Brassclw, 3, 2)
	effectstest.DeclareAttack(t, g, opp.ID, weak, strong)
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(orcs, strong); err == nil {
		t.Fatal("the Orcs blocked a creature with power 2")
	}
	if err := g.DeclareBlocker(orcs, weak); err != nil {
		t.Fatalf("the Orcs could not block a 1-power creature: %v", err)
	}
}

func TestOldB855RingOfRenewalDiscardsAtRandomThenDrawsTwo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	ring := effectstest.PushCatalogPermanent(g, me.ID, "Ring of Renewal", "Artifact", oldB855Ring, false)
	effectstest.SeedLibrary(me, "A", "B", "C")
	effectstest.AdvanceToMain(t, g)
	for i := 0; i < 3; i++ {
		effectstest.HandCard(me, "Spare", "Sorcery")
	}
	hand, yard := me.Hand.Size(), me.Graveyard.Size()
	effectstest.FillPool(me, 5)
	effectstest.P7Activate(t, g, me, ring, 0, game.ActivateAbilityParams{Strict: true})
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand %d → %d, want +1 (one discarded, two drawn)", hand, got)
	}
	if got := me.Graveyard.Size(); got != yard+1 {
		t.Errorf("graveyard %d → %d, want +1", yard, got)
	}
	if !effectstest.IsTapped(g, ring) {
		t.Error("the Ring did not tap")
	}
}

func TestOldB855ConchHornDrawsTwoPutsOneBackAndIsSacrificed(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	horn := effectstest.PushCatalogPermanent(g, me.ID, "Conch Horn", "Artifact", oldB855Conch, false)
	for i := 0; i < 3; i++ {
		me.Library.PushTop(game.NewCard("Filler", me.ID))
	}
	effectstest.AdvanceToMain(t, g)
	hand, library := me.Hand.Size(), me.Library.Size()
	effectstest.FillPool(me, 1)
	if err := g.ActivateCatalogAbility(me.ID, horn, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.LatestChooseCardsFor(g, me.ID)
	if pick == nil {
		t.Fatal("no put-back prompt")
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{me.Hand.Cards[0].InstanceID}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	if g.Battlefield.Contains(horn) {
		t.Error("the Horn was not sacrificed")
	}
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand %d → %d, want +1", hand, got)
	}
	if got := me.Library.Size(); got != library-1 {
		t.Errorf("library %d → %d, want -1", library, got)
	}
}

func TestOldB855DwarvenCatapultDividesXAmongTheirCreaturesRoundedDown(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	a := effectstest.B12Creature(g, opp.ID, "A", "Creature — Bear", 1, 3)
	b := effectstest.B12Creature(g, opp.ID, "B", "Creature — Bear", 1, 3)
	c := effectstest.B12Creature(g, opp.ID, "C", "Creature — Bear", 1, 3)
	mine := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Bear", 1, 3)
	effectstest.FloatForTest(g, me, "RRRRRRRR")
	id := effectstest.CastFromHandForTest(t, g, me, "Dwarven Catapult", "Instant", "{X}{R}", oldB855Catapult, game.CastSpellParams{Strict: true, XValue: 7, Targets: effectstest.PlayerRefs(opp.ID)})
	_ = id
	effectstest.PassPriorityAroundTable(t, g)
	for _, cr := range []uuid.UUID{a, b, c} {
		if got := effectstest.DamageMarkedOn(g, cr); got != 2 {
			t.Errorf("a creature took %d, want 2 (7 divided by 3, rounded down)", got)
		}
	}
	if got := effectstest.DamageMarkedOn(g, mine); got != 0 {
		t.Errorf("the caster's own creature took %d", got)
	}
}

func TestOldB855DwarvenCatapultDealsNothingWhenTheShareIsZero(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	a := effectstest.B12Creature(g, opp.ID, "A", "Creature — Bear", 1, 3)
	effectstest.B12Creature(g, opp.ID, "B", "Creature — Bear", 1, 3)
	effectstest.FloatForTest(g, me, "RRR")
	effectstest.CastFromHandForTest(t, g, me, "Dwarven Catapult", "Instant", "{X}{R}", oldB855Catapult, game.CastSpellParams{Strict: true, XValue: 1, Targets: effectstest.PlayerRefs(opp.ID)})
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.DamageMarkedOn(g, a); got != 0 {
		t.Errorf("a creature took %d from X=1 among two, want 0", got)
	}
}

func TestOldB855SporeCloudTapsBlockersAndPreventsCombatDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	atk := effectstest.B12Creature(g, me.ID, "Attacker", "Creature — Bear", 3, 3)
	blocker := effectstest.B12Creature(g, opp.ID, "Blocker", "Creature — Bear", 3, 3)
	bystander := effectstest.B12Creature(g, opp.ID, "Bystander", "Creature — Bear", 3, 3)
	effectstest.DeclareAttack(t, g, opp.ID, atk)
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(blocker, atk); err != nil {
		t.Fatal(err)
	}
	effectstest.FinishBlockDeclarations(t, g)
	effectstest.FloatForTest(g, me, "CGG")
	effectstest.CastFromHandForTest(t, g, me, "Spore Cloud", "Instant", "{1}{G}{G}", oldB855Cloud, game.CastSpellParams{Strict: true})
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.IsTapped(g, blocker) {
		t.Error("the blocking creature was not tapped")
	}
	if effectstest.IsTapped(g, bystander) {
		t.Error("a creature that was not blocking was tapped")
	}
	if err := g.MarkCombatDamage(atk, blocker, 3); err != nil {
		t.Fatalf("MarkCombatDamage: %v", err)
	}
	if got := effectstest.DamageMarkedOn(g, blocker); got != 0 {
		t.Errorf("combat damage marked = %d, want 0", got)
	}
}

func TestOldB855BalmOfRestorationGainsLifeOrPreventsDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	balm := effectstest.PushCatalogPermanent(g, me.ID, "Balm of Restoration", "Artifact", oldB855Balm, false)
	before := effectstest.LifeOf(g, me.ID)
	effectstest.B06AddMana(me, "C")
	effectstest.P7Activate(t, g, me, balm, 0, game.ActivateAbilityParams{Strict: true, Modes: []int{0}})
	if g.Battlefield.Contains(balm) {
		t.Error("the Balm was not sacrificed")
	}
	if got := effectstest.LifeOf(g, me.ID); got != before+2 {
		t.Errorf("life %d, want %d", got, before+2)
	}

	g2 := effectstest.NewCatalogGame(t)
	me2 := g2.Seats[0]
	effectstest.AdvanceToMain(t, g2)
	balm2 := effectstest.PushCatalogPermanent(g2, me2.ID, "Balm of Restoration", "Artifact", oldB855Balm, false)
	bear := effectstest.B12Creature(g2, me2.ID, "Bear", "Creature — Bear", 2, 4)
	effectstest.B06AddMana(me2, "C")
	effectstest.P7Activate(t, g2, me2, balm2, 0, game.ActivateAbilityParams{Strict: true, Modes: []int{1}, Targets: effectstest.CardTarget(bear)})
	if err := g2.MarkDamage(bear, 3); err != nil {
		t.Fatalf("MarkDamage: %v", err)
	}
	if got := effectstest.DamageMarkedOn(g2, bear); got != 1 {
		t.Errorf("the creature took %d of 3, want 1 (the next 2 prevented)", got)
	}
}

func TestOldB855RainbowValeMakesAnyColourAndGoesToAnOpponentAtTheEndStep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	vale := effectstest.PushCatalogPermanent(g, me.ID, "Rainbow Vale", "Land", oldB855Vale, false)
	if err := g.ActivateManaAbility(me.ID, vale, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("mana: %v", err)
	}
	effectstest.ResolveManaPickAnyColor(t, g, me.ID)
	if got := effectstest.PoolColors(me); len(got) != 1 {
		t.Errorf("pool %v, want one mana", got)
	}
	effectstest.AdvanceToEndStep(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerChoosePlayer(t, g, me.ID, g.Seats[2])
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.LayeredCard(t, g, vale).Controller; got != g.Seats[2].ID {
		t.Errorf("the land is controlled by %s, want the chosen opponent", got)
	}
}

func TestOldB855RainbowValeStaysIfItNeverTapped(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	vale := effectstest.PushCatalogPermanent(g, me.ID, "Rainbow Vale", "Land", oldB855Vale, false)
	effectstest.AdvanceToEndStep(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.LayeredCard(t, g, vale).Controller; got != me.ID {
		t.Error("an untapped Vale changed hands")
	}
}
