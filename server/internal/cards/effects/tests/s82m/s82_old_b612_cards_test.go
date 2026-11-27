package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b612_cards_test.go — Eighth Edition, activated abilities (S82 OLD batch
// 612, #5607). Every test plays the real catalog entry through the engine.

const (
	oldB612OrcishSpy         = "0d058991-3751-4afe-80c9-79f0c6e10282"
	oldB612CoastalHornclaw   = "27d672a7-f73b-483d-80a3-a4cd2a9855f2"
	oldB612FleetingImage     = "623f3efb-5431-42d4-98aa-9caac40e4e83"
	oldB612HornedTroll       = "6ecc61ff-719a-405d-b02a-01607729d05c"
	oldB612CallOfTheWild     = "882ee72e-2149-4778-a81c-8bf09beb5d66"
	oldB612SamiteHealer      = "95a0ca48-d924-47f4-86ed-42c673ee778c"
	oldB612DistortingLens    = "9c9579c5-7edf-41ca-8595-713ca9245576"
	oldB612DisruptingScepter = "cd30a128-4b23-476e-8066-2272fdc395d9"
	oldB612Rhox              = "f6f4c684-2915-4a58-9c4b-487387f7cf73"
)

func TestS82OldB612EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB612OrcishSpy: "Orcish Spy", oldB612CoastalHornclaw: "Coastal Hornclaw",
		oldB612FleetingImage: "Fleeting Image", oldB612HornedTroll: "Horned Troll",
		oldB612CallOfTheWild: "Call of the Wild", oldB612SamiteHealer: "Samite Healer",
		oldB612DistortingLens: "Distorting Lens", oldB612DisruptingScepter: "Disrupting Scepter",
		oldB612Rhox: "Rhox",
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

func TestS82OldB612OrcishSpyLooksAtTheTopThreeOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	spy := effectstest.B12Push(g, me.ID, "Orcish Spy", "Creature — Orc Rogue", oldB612OrcishSpy, 1, 1)
	opp.Library.Cards = nil
	ids := make([]uuid.UUID, 5)
	for i := range ids {
		ids[i] = uuid.New()
		opp.Library.PushTop(game.Card{InstanceID: ids[i], Name: "Card", TypeLine: "Instant", Owner: opp.ID, Controller: opp.ID})
	}
	effectstest.B16Activate(t, g, me.ID, spy, 0, game.ActivateAbilityParams{Targets: effectstest.PlayerRefs(opp.ID)})
	effectstest.PassPriorityAroundTable(t, g)
	size := len(opp.Library.Cards)
	for i := 0; i < size; i++ {
		c := opp.Library.Cards[i]
		top := i >= size-3
		if c.IsKnownTo(me.ID) != top {
			t.Errorf("library card %d from the bottom: known to me = %v, want %v", i, c.IsKnownTo(me.ID), top)
		}
		if c.IsKnownTo(g.Seats[2].ID) {
			t.Errorf("library card %d is known to a bystander", i)
		}
	}
	if size != 5 {
		t.Errorf("the library has %d cards, want it untouched at 5", size)
	}
}

func TestS82OldB612CoastalHornclawSacrificesALandForFlying(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	claw := effectstest.B12Push(g, me.ID, "Coastal Hornclaw", "Creature — Bird", oldB612CoastalHornclaw, 3, 3)
	if effectstest.HasKeywordOnBattlefield(t, g, claw, "flying") {
		t.Fatal("flying before the ability")
	}
	if err := g.ActivateCatalogAbility(me.ID, claw, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Error("activated with no land to sacrifice")
	}
	land := effectstest.PushLandFor(g, me.ID, "Island", "Basic Land — Island")
	effectstest.B16Activate(t, g, me.ID, claw, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{land}})
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(land) {
		t.Error("the land was not sacrificed")
	}
	if !effectstest.HasKeywordOnBattlefield(t, g, claw, "flying") {
		t.Error("no flying after the ability resolved")
	}
}

func TestS82OldB612FleetingImageFliesAndBouncesItself(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	image := effectstest.B12Push(g, me.ID, "Fleeting Image", "Creature — Illusion", oldB612FleetingImage, 2, 1)
	if !effectstest.HasKeywordOnBattlefield(t, g, image, "flying") {
		t.Error("no flying")
	}
	effectstest.B06AddMana(me, "U", "C")
	effectstest.B16Activate(t, g, me.ID, image, 0, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(image) || !effectstest.HandHasID(me, image) {
		t.Error("the creature did not return to its owner's hand")
	}
}

func TestS82OldB612HornedTrollRegeneratesForG(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	troll := effectstest.B12Push(g, me.ID, "Horned Troll", "Creature — Troll", oldB612HornedTroll, 2, 2)
	effectstest.B06AddMana(me, "G")
	effectstest.B16Activate(t, g, me.ID, troll, 0, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.B27Kill(g, troll)
	g.RunStateChecksForTest()
	if !g.Battlefield.Contains(troll) {
		t.Error("the regeneration shield did not save the Troll")
	}
}

func TestS82OldB612CallOfTheWildPutsACreatureOntoTheBattlefieldElseTheGraveyard(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, zone string
	}{
		{"Top Bear", "Creature — Bear", "battlefield"},
		{"Top Bolt", "Instant", "graveyard"},
		{"Top Rock", "Artifact", "graveyard"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[0]
			effectstest.AdvanceToMain(t, g)
			wild := effectstest.PushCatalogPermanent(g, me.ID, "Call of the Wild", "Enchantment", oldB612CallOfTheWild, false)
			me.Library.Cards = nil
			under := uuid.New()
			me.Library.PushTop(game.Card{InstanceID: under, Name: "Under", TypeLine: "Basic Land — Forest", Owner: me.ID, Controller: me.ID})
			top := uuid.New()
			me.Library.PushTop(game.Card{InstanceID: top, Name: tc.name, TypeLine: tc.typeLine, Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID})
			effectstest.B06AddMana(me, "G", "G", "C", "C")
			effectstest.B16Activate(t, g, me.ID, wild, 0, game.ActivateAbilityParams{})
			effectstest.PassPriorityAroundTable(t, g)
			if z, _ := effectstest.CardWhere(g, me, top); z != tc.zone {
				t.Errorf("%s: the revealed card is in %s, want %s", tc.name, z, tc.zone)
			}
			if z, _ := effectstest.CardWhere(g, me, under); z != "library" {
				t.Errorf("the card beneath it moved to %s", z)
			}
		})
	}
}

func TestS82OldB612SamiteHealerPreventsTheNextOneDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	healer := effectstest.B12Push(g, me.ID, "Samite Healer", "Creature — Human Cleric", oldB612SamiteHealer, 1, 1)
	effectstest.B16Activate(t, g, me.ID, healer, 0, game.ActivateAbilityParams{Targets: effectstest.PlayerRefs(opp.ID)})
	effectstest.PassPriorityAroundTable(t, g)
	life := opp.Life
	g.WithWriteLock(func() { _ = g.DealDamageToPlayerForEffect(uuid.New(), opp.ID, 3) })
	if opp.Life != life-2 {
		t.Errorf("3 damage into a 1-shield: life %d → %d, want −2", life, opp.Life)
	}
	g.WithWriteLock(func() { _ = g.DealDamageToPlayerForEffect(uuid.New(), opp.ID, 1) })
	if opp.Life != life-3 {
		t.Errorf("the shield is spent: life %d → %d, want −3", life, opp.Life)
	}
}

func TestS82OldB612DistortingLensRecoloursAPermanentUntilEndOfTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	lens := effectstest.PushCatalogPermanent(g, me.ID, "Distorting Lens", "Artifact", oldB612DistortingLens, false)
	bear := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Green Bear", TypeLine: "Creature — Bear",
		Power: 2, Toughness: 2, Colors: []string{"G", "W"}, Owner: opp.ID, Controller: opp.ID})
	effectstest.B16Activate(t, g, me.ID, lens, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(bear)})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerColor(t, g, me.ID, "R")
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.EffectiveColors(t, g, bear); len(got) != 1 || got[0] != "R" {
		t.Errorf("the creature is %v, want only red", got)
	}
}

func TestS82OldB612DisruptingScepterMakesThemDiscardOnlyOnYourTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	scepter := effectstest.B12Push(g, me.ID, "Disrupting Scepter", "Artifact", oldB612DisruptingScepter, 0, 0)
	effectstest.D2dSetHand(opp, 2)
	keep, toss := opp.Hand.Cards[0].InstanceID, opp.Hand.Cards[1].InstanceID
	effectstest.B06AddMana(me, "C", "C", "C")
	effectstest.B16Activate(t, g, me.ID, scepter, 0, game.ActivateAbilityParams{Targets: effectstest.PlayerRefs(opp.ID)})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerDiscard(t, g, opp.ID, toss)
	if opp.Hand.Size() != 1 || !effectstest.HandHasID(opp, keep) {
		t.Errorf("hand after the discard: %d cards, want the one they kept", opp.Hand.Size())
	}

	// On an opponent's turn it cannot be activated.
	g2 := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g2)
	active := g2.Seats[g2.Turn.ActiveSeat]
	other := g2.Seats[(g2.Turn.ActiveSeat+1)%len(g2.Seats)]
	theirs := effectstest.B12Push(g2, other.ID, "Disrupting Scepter", "Artifact", oldB612DisruptingScepter, 0, 0)
	effectstest.B06AddMana(other, "C", "C", "C")
	if err := g2.ActivateCatalogAbility(other.ID, theirs, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.PlayerRefs(active.ID)}); err == nil {
		t.Error("activated on an opponent's turn")
	}
}

func TestS82OldB612BlockedRhoxMayAssignAsThoughUnblocked(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	rhox := effectstest.ApaPush(g, g.Seats[0].ID, g.Seats[0].ID, game.Card{
		Name: "Rhox", TypeLine: "Creature — Rhino Beast", OracleID: oldB612Rhox, Power: 5, Toughness: 5,
	})
	wall := effectstest.MscWlBlocker(g, "Wall", 0, 7)
	effectstest.MscWlBlockedCombat(t, g, []uuid.UUID{rhox}, map[uuid.UUID]uuid.UUID{wall: rhox})

	c := effectstest.MscWlDamagePrompt(t, g)
	if f := c.DamageAssignment; !f.AsThoughUnblocked || len(f.BlockerIDs) != 1 || f.AttackerPower != 5 {
		t.Fatalf("the prompt is %+v, want one blocker, power 5, as-though-unblocked", f)
	}
	life := opp.Life
	if err := effectstest.MscWlAnswer(t, g, c, 5, game.DamageAssignmentEntry{BlockerID: wall, Amount: 0}); err != nil {
		t.Fatalf("all of it to the player: %v", err)
	}
	if opp.Life != life-5 {
		t.Errorf("the defending player went %d -> %d, want -5", life, opp.Life)
	}
	if got := effectstest.Pr6Marked(g, wall); got != 0 {
		t.Errorf("the blocker was dealt %d damage, want none", got)
	}
}

func TestS82OldB612RhoxRegeneratesForTwoG(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	rhox := effectstest.B12Push(g, me.ID, "Rhox", "Creature — Rhino Beast", oldB612Rhox, 5, 5)
	effectstest.B06AddMana(me, "G", "C", "C")
	effectstest.B16Activate(t, g, me.ID, rhox, 0, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.B27Kill(g, rhox)
	g.RunStateChecksForTest()
	if !g.Battlefield.Contains(rhox) {
		t.Error("the regeneration shield did not save Rhox")
	}
}
