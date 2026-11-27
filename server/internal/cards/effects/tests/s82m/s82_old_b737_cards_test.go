package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b737_cards_test.go — Exodus, "Activated abilities", batch 737
// (#5626). Every test plays the real catalog entry through the engine.

const (
	o737CatBurglar = "16aea5b3-9f43-4852-b8b8-80e87909b794"
	o737Mystic     = "32d8f7c8-8680-4901-b31b-e0cb66e37548"
	o737Bag        = "5801f773-e214-413d-968d-b84bf99f0642"
	o737Matter     = "656afde2-cdf3-4907-a0fc-7ac93f5d3e03"
	o737Keeper     = "86dcdffd-c14f-4e06-a74f-d2c727f55b1c"
	o737Licid      = "ae2804c6-fb38-4845-ae21-7f41a766cbba"
	o737Skyshaper  = "b5642460-883c-4e38-8552-9bea4e50d8e8"
	o737Surgeon    = "ead5a2bf-f48a-43a2-89bb-df3cc587c843"
)

func TestS82OldB737EveryCardIsRegisteredFull(t *testing.T) {
	for _, oracle := range []string{o737CatBurglar, o737Mystic, o737Bag, o737Matter, o737Keeper, o737Licid, o737Skyshaper, o737Surgeon} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not in the catalog", oracle)
			continue
		}
		if spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: completeness %v with %d caveats, want Full with none", spec.Name, spec.Completeness, len(spec.Caveats))
		}
	}
}

func o737Activate(g *game.Game, me *game.Player, src uuid.UUID, params game.ActivateAbilityParams) error {
	params.Strict = true
	return g.ActivateCatalogAbility(me.ID, src, 0, params)
}

func o737Foe(opp *game.Player) []game.TargetRef {
	return []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}
}

func TestS82OldB737CatBurglarMakesATargetPlayerDiscardAtSorcerySpeedOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	cat := effectstest.B12Push(g, me.ID, "Cat Burglar", "Creature — Kor Rogue Minion", o737CatBurglar, 2, 2)
	held := effectstest.PushHandCardOfType(opp, "Held", "Sorcery")
	effectstest.FloatForTest(g, me, "CCB")
	effectstest.ApaActivate(t, g, me, cat, 0, game.ActivateAbilityParams{Targets: o737Foe(opp)})
	effectstest.AnswerDiscard(t, g, opp.ID, held)
	if !opp.Graveyard.Contains(held) {
		t.Error("the target player did not discard")
	}
	g2 := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToEndStep(t, g2)
	me2, opp2 := g2.Seats[g2.Turn.ActiveSeat], g2.Seats[1]
	cat2 := effectstest.B12Push(g2, me2.ID, "Cat Burglar", "Creature — Kor Rogue Minion", o737CatBurglar, 2, 2)
	effectstest.FloatForTest(g2, me2, "CCB")
	if err := o737Activate(g2, me2, cat2, game.ActivateAbilityParams{Targets: o737Foe(opp2)}); err == nil {
		t.Error("activated outside a sorcery window")
	}
}

func TestS82OldB737RootwaterMysticLooksAtTheTopCardAndLeavesItThere(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	top := effectstest.SsCard(opp.Library, opp.ID, "Their Top", "Sorcery", 0, 0)
	mystic := effectstest.B12Push(g, me.ID, "Rootwater Mystic", "Creature — Merfolk Wizard", o737Mystic, 1, 1)
	effectstest.FloatForTest(g, me, "CU")
	effectstest.ApaActivate(t, g, me, mystic, 0, game.ActivateAbilityParams{Targets: o737Foe(opp)})
	effectstest.PassPriorityAroundTable(t, g)
	if !opp.Library.Contains(top) {
		t.Error("the looked-at card left the library")
	}
	if c, err := opp.Library.Top(); err != nil || c.InstanceID != top {
		t.Error("the looked-at card is no longer on top")
	}
}

func TestS82OldB737MedicineBagRegeneratesForADiscard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	bag := effectstest.B12Push(g, me.ID, "Medicine Bag", "Artifact", o737Bag, 0, 0)
	victim := effectstest.B12Creature(g, me.ID, "Victim", "Creature — Bear", 2, 2)
	src := effectstest.B12Creature(g, opp.ID, "Source", "Creature — Ogre", 5, 5)
	effectstest.FloatForTest(g, me, "C")
	if err := o737Activate(g, me, bag, game.ActivateAbilityParams{Targets: effectstest.CardRefs(victim)}); err == nil {
		t.Fatal("activated without discarding")
	}
	spare := effectstest.PushHandCardOfType(me, "Spare", "Sorcery")
	effectstest.ApaActivate(t, g, me, bag, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(victim), DiscardIDs: []uuid.UUID{spare}})
	if !me.Graveyard.Contains(spare) {
		t.Error("the discard was not paid")
	}
	effectstest.Cmdr2B08Damage(g, src, victim, 5)
	effectstest.PassPriorityAroundTable(t, g)
	g.RunStateChecksForTest()
	if _, ok := effectstest.BattlefieldCard(g, victim); !ok {
		t.Error("the regenerated creature died")
	}
}

func TestS82OldB737MindOverMatterTapsOrUntapsAnArtifactCreatureOrLand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	mom := effectstest.B12Push(g, me.ID, "Mind Over Matter", "Enchantment", o737Matter, 0, 0)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	ench := effectstest.B12Push(g, me.ID, "Plain Aura", "Enchantment", "", 0, 0)
	spare := effectstest.PushHandCardOfType(me, "Spare", "Sorcery")
	if err := o737Activate(g, me, mom, game.ActivateAbilityParams{Targets: effectstest.CardRefs(ench), DiscardIDs: []uuid.UUID{spare}}); err == nil {
		t.Fatal("a plain enchantment was a legal target")
	}
	effectstest.ApaActivate(t, g, me, mom, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(bear), DiscardIDs: []uuid.UUID{spare}})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerOptionPick(t, g, me.ID, 0)
	effectstest.PassPriorityAroundTable(t, g)
	if tapped, _ := effectstest.BattlefieldCardTapped(g, bear); !tapped {
		t.Error("the creature was not tapped")
	}
}

func TestS82OldB737SkyshaperGivesYourCreaturesFlyingForTheTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	sky := effectstest.B12Push(g, me.ID, "Skyshaper", "Artifact", o737Skyshaper, 0, 0)
	mine := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	effectstest.ApaActivate(t, g, me, sky, 0, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.FindBattlefieldCardForTest(g, sky) != nil {
		t.Error("Skyshaper was not sacrificed")
	}
	if !effectstest.HasEffectiveKeyword(t, g, mine, "flying") {
		t.Error("your creature did not gain flying")
	}
	if effectstest.HasEffectiveKeyword(t, g, theirs, "flying") {
		t.Error("an opponent's creature gained flying")
	}
}

func TestS82OldB737ThrullSurgeonDiscardsTheCardYouChoose(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	for opp.Hand.Size() > 0 {
		opp.Hand.PopTop()
	}
	keep := effectstest.HandSpell(opp, "Keep", "Sorcery", "")
	toss := effectstest.HandSpell(opp, "Toss", "Sorcery", "")
	th := effectstest.B12Push(g, me.ID, "Thrull Surgeon", "Creature — Thrull", o737Surgeon, 1, 1)
	effectstest.FloatForTest(g, me, "CB")
	effectstest.ApaActivate(t, g, me, th, 0, game.ActivateAbilityParams{Targets: o737Foe(opp)})
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.OpenVariant(t, g)
	if len(pick.DiscardOptions) != 2 || pick.FromPlayer != opp.ID {
		t.Fatalf("the pick offers %d cards from %v, want the target's 2", len(pick.DiscardOptions), pick.FromPlayer)
	}
	if err := g.ResolvePendingChoice(pick.ID, me.ID, []uuid.UUID{toss}); err != nil {
		t.Fatalf("ResolvePendingChoice: %v", err)
	}
	if !opp.Graveyard.Contains(toss) || !opp.Hand.Contains(keep) {
		t.Errorf("toss in graveyard %v, keep in hand %v", opp.Graveyard.Contains(toss), opp.Hand.Contains(keep))
	}
	if effectstest.FindBattlefieldCardForTest(g, th) != nil {
		t.Error("the Surgeon was not sacrificed")
	}
}

func TestS82OldB737DominatingLicidTakesControlAndGivesItBack(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	licid := effectstest.B12Push(g, me.ID, "Dominating Licid", "Creature — Licid", o737Licid, 1, 1)
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.FloatForTest(g, me, "CUU")
	effectstest.ApaActivate(t, g, me, licid, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(bear)})
	effectstest.PassPriorityAroundTable(t, g)
	l := effectstest.B12Card(t, g, licid)
	if !l.IsAura() || l.IsCreature() || l.AttachedTo.ID != bear {
		t.Fatalf("the licid is not an Aura on the bear")
	}
	if c := effectstest.B12Card(t, g, bear); c.Controller != me.ID {
		t.Fatalf("the bear is controlled by %v, want the licid's controller", c.Controller)
	}
	if err := g.PerformSpecialAction(me.ID, licid, game.SpecialActionEndEffect, game.SpecialActionParams{Strict: true}); err == nil {
		t.Fatal("ended the effect without paying {U}")
	}
	effectstest.FloatForTest(g, me, "U")
	if err := g.PerformSpecialAction(me.ID, licid, game.SpecialActionEndEffect, game.SpecialActionParams{}); err != nil {
		t.Fatalf("end_effect: %v", err)
	}
	if c := effectstest.B12Card(t, g, bear); c.Controller != opp.ID {
		t.Errorf("after ending, the bear is controlled by %v, want its owner", c.Controller)
	}
	if l := effectstest.B12Card(t, g, licid); !l.IsCreature() || l.IsAura() {
		t.Error("after ending, the licid is not a creature again")
	}
}

func TestS82OldB737KeeperOfTheBeastsNeedsAnOpponentWithMoreCreaturesAsYouActivate(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	keeper := effectstest.B12Push(g, me.ID, "Keeper of the Beasts", "Creature — Human Wizard", o737Keeper, 1, 2)
	effectstest.FloatForTest(g, me, "G")
	// Equal counts (one each, the keeper being mine): not more.
	effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	if err := o737Activate(g, me, keeper, game.ActivateAbilityParams{Targets: o737Foe(opp)}); err == nil {
		t.Fatal("targeted an opponent who controls no more creatures than you")
	}
	// More: the activation is legal.
	second := effectstest.B12Creature(g, opp.ID, "Their Ox", "Creature — Ox", 2, 2)
	effectstest.ApaActivate(t, g, me, keeper, 0, game.ActivateAbilityParams{Targets: o737Foe(opp)})
	// As you activate: losing a creature in response does not stop it.
	if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneBattlefield}, game.ZoneRef{Kind: game.ZoneGraveyard, Owner: opp.ID}, second); err != nil {
		t.Fatalf("MoveCardByID: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountOnBattlefield(g, "Beast", me.ID); n != 1 {
		t.Errorf("%d Beast tokens, want 1 even though the count changed after activation", n)
	}
}
