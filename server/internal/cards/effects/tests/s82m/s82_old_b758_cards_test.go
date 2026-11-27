package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// S82 OLD batch 758 (Tempest, #5628): Jackal Pup, Repentance, Flowstone
// Salamander, Thumbscrews, Heartwood Giant, the two licids and Rootbreaker
// Wurm.

const (
	oldB758Pup        = "993e41e6-0643-4a1f-8ada-b7a4aae486d9"
	oldB758Repentance = "9a345a67-0c6c-4f6f-b3c9-d525199cdbfa"
	oldB758Salamander = "a0e40eb1-a909-405b-8fb4-3759df8edf7f"
	oldB758Thumbs     = "a8f1f55a-7a85-4338-b4d1-3d28f10f69a7"
	oldB758Giant      = "b7d7418d-222b-4ae0-b0d9-9ea71d6ad800"
	oldB758Leeching   = "bcb5af89-a304-4de0-86af-b3f4cfabafc4"
	oldB758Stinging   = "c6e48eaf-02f8-43b3-90bf-970d4e412fa6"
	oldB758Wurm       = "d3edbb47-6892-4853-badc-cc01499d4e55"
)

func TestOldB758EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB758Pup: "Jackal Pup", oldB758Repentance: "Repentance", oldB758Salamander: "Flowstone Salamander",
		oldB758Thumbs: "Thumbscrews", oldB758Giant: "Heartwood Giant", oldB758Leeching: "Leeching Licid",
		oldB758Stinging: "Stinging Licid", oldB758Wurm: "Rootbreaker Wurm",
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

func TestOldB758RootbreakerWurmTramples(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	id := effectstest.PushCatalogPermanent(g, g.Seats[0].ID, "Rootbreaker Wurm", "Creature — Wurm", oldB758Wurm, false)
	if !effectstest.HasEffectiveKeyword(t, g, id, "trample") {
		t.Error("no trample")
	}
}

func TestOldB758JackalPupReflectsTheDamageToItsController(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pup := effectstest.B12Push(g, me.ID, "Jackal Pup", "Creature — Jackal", oldB758Pup, 2, 3)
	src := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	life, theirs := me.Life, opp.Life
	effectstest.DamageAndSettle(t, g, src, pup, 2)
	if me.Life != life-2 {
		t.Errorf("my life %d → %d, want 2 damage", life, me.Life)
	}
	if opp.Life != theirs {
		t.Errorf("their life changed %d → %d", theirs, opp.Life)
	}
}

func TestOldB758JackalPupStillDealsItsDamageWhenItDies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pup := effectstest.B12Push(g, me.ID, "Jackal Pup", "Creature — Jackal", oldB758Pup, 2, 1)
	src := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	life := me.Life
	effectstest.DamageAndSettle(t, g, src, pup, 5)
	if _, ok := effectstest.CardOnBattlefield(g, pup); ok {
		t.Error("the Pup survived 5 damage")
	}
	if me.Life != life-5 {
		t.Errorf("my life %d → %d, want 5 damage from the dead Pup", life, me.Life)
	}
}

func TestOldB758RepentanceHurtsACreatureByItsOwnPower(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	big := effectstest.B12Creature(g, opp.ID, "Big", "Creature — Bear", 3, 4)
	effectstest.CastCatalogSpell(t, g, "Repentance", "Sorcery", oldB758Repentance, []game.TargetRef{{Kind: game.TargetCard, ID: big}})
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.DamageMarkedOn(g, big); got != 3 {
		t.Errorf("damage on the creature %d, want 3 (its power)", got)
	}
}

func TestOldB758FlowstoneSalamanderTargetsOnlyItsBlockers(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	sal := effectstest.PushCatalogPermanent(g, me.ID, "Flowstone Salamander", "Creature — Salamander", oldB758Salamander, false)
	blocker := effectstest.B12Creature(g, opp.ID, "Blocker", "Creature — Bear", 2, 3)
	bystander := effectstest.B12Creature(g, opp.ID, "Bystander", "Creature — Bear", 2, 3)
	effectstest.CrtBlock(g, blocker, sal, opp.ID)
	if err := effectstest.CrtActivateRaw(t, g, me.ID, sal, 0, effectstest.LtCardTarget(bystander), "R"); err == nil {
		t.Error("targeted a creature that is not blocking it")
	}
	if err := effectstest.CrtActivateRaw(t, g, me.ID, sal, 0, effectstest.LtCardTarget(blocker), "R"); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.DamageMarkedOn(g, blocker); got != 1 {
		t.Errorf("damage on the blocker %d, want 1", got)
	}
}

func TestOldB758ThumbscrewsNeedsFiveOrMoreCards(t *testing.T) {
	for _, tc := range []struct{ hand, want int }{{5, 1}, {7, 1}, {4, 0}, {0, 0}} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		me.Hand.Cards = nil
		for i := 0; i < tc.hand; i++ {
			effectstest.PushCatalogHandCard(me, "Filler", "Sorcery", "")
		}
		effectstest.PushCatalogPermanent(g, me.ID, "Thumbscrews", "Artifact", oldB758Thumbs, false)
		before := opp.Life
		g.WithWriteLock(func() { g.EmitEvent(game.Event{Kind: game.EventBeginUpkeep, Actor: me.ID}) })
		if pick := effectstest.LatestPickTarget(g, me.ID); pick != nil {
			if tc.want == 0 {
				t.Errorf("hand %d: the ability asked for a target", tc.hand)
			}
			effectstest.AnswerTriggerTargets2(t, g, me.ID, opp.ID)
		}
		effectstest.PassPriorityAroundTable(t, g)
		if got := before - opp.Life; got != tc.want {
			t.Errorf("hand %d: %d damage, want %d", tc.hand, got, tc.want)
		}
	}
}

func TestOldB758HeartwoodGiantSacrificesAForestForTwoDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	giant := effectstest.PushCatalogPermanent(g, me.ID, "Heartwood Giant", "Creature — Giant", oldB758Giant, false)
	forest := effectstest.PushLandFor(g, me.ID, "Forest", "Basic Land — Forest")
	island := effectstest.PushLandFor(g, me.ID, "Island", "Basic Land — Island")
	target := []game.TargetRef{effectstest.PlayerTarget(opp.ID)}
	if err := g.ActivateCatalogAbility(me.ID, giant, 0, game.ActivateAbilityParams{Targets: target, SacrificeIDs: []uuid.UUID{island}, Strict: true}); err == nil {
		t.Error("sacrificed an Island to pay for the Giant")
	}
	before := opp.Life
	if err := g.ActivateCatalogAbility(me.ID, giant, 0, game.ActivateAbilityParams{Targets: target, SacrificeIDs: []uuid.UUID{forest}, Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(forest) {
		t.Error("the Forest was not sacrificed")
	}
	if opp.Life != before-2 {
		t.Errorf("life %d → %d, want 2 damage", before, opp.Life)
	}
	if !effectstest.TappedNow(g, giant) {
		t.Error("the Giant was not tapped")
	}
}

func oldB758Licid(t *testing.T, oracle, name, mana string) (g *game.Game, me, opp *game.Player, licid, bear uuid.UUID) {
	t.Helper()
	g = effectstest.NewCatalogGame(t)
	me, opp = g.Seats[0], g.Seats[1]
	licid = effectstest.B12Push(g, me.ID, name, "Creature — Licid", oracle, 1, 1)
	bear = effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	for _, c := range mana {
		effectstest.FillPoolColored(me, string(c), 1)
	}
	if err := g.ActivateCatalogAbility(me.ID, licid, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}}, Strict: true,
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	l := effectstest.B12Card(t, g, licid)
	if !l.IsAura() || l.IsCreature() || l.AttachedTo.ID != bear {
		t.Fatalf("not an Aura on the bear (aura %v, creature %v, attached %+v)", l.IsAura(), l.IsCreature(), l.AttachedTo)
	}
	return g, me, opp, licid, bear
}

func TestOldB758LeechingLicidPingsTheEnchantedCreaturesControllerEachUpkeep(t *testing.T) {
	g, me, opp, licid, _ := oldB758Licid(t, oldB758Leeching, "Leeching Licid", "BB")
	life, mine := opp.Life, me.Life
	effectstest.BeginUpkeep(t, g, opp.ID)
	if opp.Life != life-1 {
		t.Errorf("their life %d → %d, want 1 damage", life, opp.Life)
	}
	effectstest.BeginUpkeep(t, g, me.ID)
	if me.Life != mine {
		t.Errorf("my upkeep hurt me: %d → %d", mine, me.Life)
	}
	effectstest.FillPoolColored(me, "B", 1)
	if err := g.PerformSpecialAction(me.ID, licid, game.SpecialActionEndEffect, game.SpecialActionParams{}); err != nil {
		t.Fatalf("end effect: %v", err)
	}
	if l := effectstest.B12Card(t, g, licid); !l.IsCreature() || l.IsAura() {
		t.Errorf("after ending it is not a creature again (aura %v, creature %v)", l.IsAura(), l.IsCreature())
	}
	life = opp.Life
	effectstest.BeginUpkeep(t, g, opp.ID)
	if opp.Life != life {
		t.Errorf("the licid pinged after the effect ended: %d → %d", life, opp.Life)
	}
}

func TestOldB758LeechingLicidDoesNothingAsACreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Leeching Licid", "Creature — Licid", oldB758Leeching, 1, 1)
	life, mine := opp.Life, me.Life
	effectstest.BeginUpkeep(t, g, opp.ID)
	effectstest.BeginUpkeep(t, g, me.ID)
	if opp.Life != life || me.Life != mine {
		t.Errorf("an unattached licid pinged: opp %d → %d, me %d → %d", life, opp.Life, mine, me.Life)
	}
}

func TestOldB758StingingLicidHurtsWhenTheEnchantedCreatureTaps(t *testing.T) {
	g, me, opp, _, bear := oldB758Licid(t, oldB758Stinging, "Stinging Licid", "UU")
	life, mine := opp.Life, me.Life
	effectstest.TapForTest(t, g, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-2 {
		t.Errorf("their life %d → %d, want 2 damage", life, opp.Life)
	}
	if me.Life != mine {
		t.Errorf("my life changed %d → %d", mine, me.Life)
	}
}
