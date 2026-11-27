package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b682_test.go — Mercadian Masques, "Activated abilities", batch 682
// (#5619). Every test plays the real catalog entry through the engine.

const (
	o682Official  = "c152b677-5e75-43da-be1d-876b7b83e2f4"
	o682Scouts    = "c71eb9b5-2017-4e6c-8140-f9c8af65893f"
	o682Distort   = "cc9cbbbc-8b02-42ef-a4d8-e2837e742c88"
	o682Caravan   = "d13aff7c-f0ee-44e2-ae1c-62cd86ad46a6"
	o682Dawn      = "d2783a37-b6de-4184-b094-e1a23f185a94"
	o682Elder     = "e8d18dc7-f44b-47f6-8400-3f563c2f19b7"
	o682Toymaker  = "eb80f233-76b5-4c72-8a4a-f22e56314690"
	o682Witch     = "ee608650-410a-4d6b-be6e-63d7955cff39"
	o682Breaker   = "f3413f57-229e-49dd-a776-6905b80b2b12"
	o682Archive   = "f8c7da00-69e1-459d-9b66-6de649c5c511"
	o682Runner    = "f97df8f4-a7ef-45f9-9563-b76b4d2b03f5"
	o682Alabaster = "f9f52651-cc64-48c5-bc18-cd5f646dfb1f"
)

func TestS82OldB682EveryCardIsRegisteredFull(t *testing.T) {
	for _, oracle := range []string{o682Official, o682Scouts, o682Distort, o682Caravan, o682Dawn, o682Elder, o682Toymaker, o682Witch, o682Breaker, o682Archive, o682Runner, o682Alabaster} {
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

func o682Perm(g *game.Game, owner uuid.UUID, name, typeLine, oracle string, p, tough int) uuid.UUID {
	return effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: typeLine, OracleID: oracle,
		Power: p, Toughness: tough, Owner: owner, Controller: owner,
	})
}

func o682Hand(g *game.Game, p *game.Player) uuid.UUID {
	id := uuid.New()
	g.WithWriteLock(func() {
		p.Hand.PushTop(game.Card{InstanceID: id, Name: "Spare", TypeLine: "Sorcery", Owner: p.ID, Controller: p.ID})
	})
	return id
}

// o682Top puts a card on top of the player's library.
func o682Top(g *game.Game, p *game.Player, name, typeLine string) uuid.UUID {
	id := uuid.New()
	g.WithWriteLock(func() {
		p.Library.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, Owner: p.ID, Controller: p.ID})
	})
	return id
}

func o682Activate(g *game.Game, me *game.Player, src uuid.UUID, params game.ActivateAbilityParams) error {
	params.Strict = true
	return g.ActivateCatalogAbility(me.ID, src, 0, params)
}

func TestS82OldB682CorruptOfficialRegeneratesAndMakesTheDefenderDiscardAtRandomWhenBlocked(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	off := effectstest.B12Push(g, me.ID, "Corrupt Official", "Creature — Human Minion", o682Official, 3, 1)
	wall := effectstest.B12Creature(g, opp.ID, "Wall", "Creature — Wall", 0, 4)
	held := effectstest.PushHandCardOfType(opp, "Held", "Sorcery")
	hand, yard := opp.Hand.Size(), opp.Graveyard.Size()
	effectstest.UnblockedAttackIntoBlockers(t, g, off)
	if err := g.DeclareBlocker(wall, off); err != nil {
		t.Fatalf("DeclareBlocker: %v", err)
	}
	effectstest.LockInBlocks(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.DiscardChoiceFor(g, opp.ID) != nil {
		t.Error("the discard was a choice, not random")
	}
	_ = held
	if opp.Hand.Size() != hand-1 || opp.Graveyard.Size() != yard+1 {
		t.Errorf("hand %d→%d, graveyard %d→%d, want one card discarded", hand, opp.Hand.Size(), yard, opp.Graveyard.Size())
	}
}

func TestS82OldB682CorruptOfficialDoesNothingWhenUnblockedAndRegenerates(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	off := effectstest.B12Push(g, me.ID, "Corrupt Official", "Creature — Human Minion", o682Official, 3, 1)
	held := effectstest.PushHandCardOfType(opp, "Held", "Sorcery")
	effectstest.UnblockedAttackIntoBlockers(t, g, off)
	if err := g.FinishBlocks(opp.ID); err != nil {
		t.Fatalf("FinishBlocks: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !opp.Hand.Contains(held) {
		t.Error("an unblocked Corrupt Official made the defender discard")
	}
	// Its regeneration row.
	g2 := effectstest.NewCatalogGame(t)
	me2, opp2 := g2.Seats[0], g2.Seats[1]
	effectstest.AdvanceToMain(t, g2)
	off2 := effectstest.B12Push(g2, me2.ID, "Corrupt Official", "Creature — Human Minion", o682Official, 3, 1)
	src := effectstest.B12Creature(g2, opp2.ID, "Source", "Creature — Ogre", 5, 5)
	effectstest.FloatForTest(g2, me2, "CCB")
	effectstest.ApaActivate(t, g2, me2, off2, 0, game.ActivateAbilityParams{})
	effectstest.Cmdr2B08Damage(g2, src, off2, 5)
	effectstest.PassPriorityAroundTable(t, g2)
	g2.RunStateChecksForTest()
	if _, ok := effectstest.BattlefieldCard(g2, off2); !ok {
		t.Error("the regenerated Corrupt Official died")
	}
}

func TestS82OldB682RappellingScoutsGainsProtectionFromTheChosenColor(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	sc := effectstest.B12Push(g, me.ID, "Rappelling Scouts", "Creature — Human Rebel Scout", o682Scouts, 1, 4)
	effectstest.FloatForTest(g, me, "CCW")
	effectstest.ApaActivate(t, g, me, sc, 0, game.ActivateAbilityParams{})
	effectstest.AnswerColor(t, g, me.ID, "R")
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasEffectiveKeyword(t, g, sc, "protection from red") {
		t.Error("the Scouts did not gain protection from red")
	}
}

func TestS82OldB682WallOfDistortionMakesATargetPlayerDiscardAtSorcerySpeedOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	wall := o682Perm(g, me.ID, "Wall of Distortion", "Creature — Wall", o682Distort, 1, 3)
	held := effectstest.PushHandCardOfType(opp, "Held", "Sorcery")
	foe := []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}
	effectstest.FloatForTest(g, me, "CCB")
	effectstest.ApaActivate(t, g, me, wall, 0, game.ActivateAbilityParams{Targets: foe})
	effectstest.AnswerDiscard(t, g, opp.ID, held)
	if !opp.Graveyard.Contains(held) {
		t.Error("the target player did not discard")
	}
	// Not at instant speed: the end step is not a sorcery window.
	g2 := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToEndStep(t, g2)
	me2, opp2 := g2.Seats[g2.Turn.ActiveSeat], g2.Seats[1]
	wall2 := o682Perm(g2, me2.ID, "Wall of Distortion", "Creature — Wall", o682Distort, 1, 3)
	effectstest.FloatForTest(g2, me2, "CCB")
	if err := o682Activate(g2, me2, wall2, game.ActivateAbilityParams{Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp2.ID}}}); err == nil {
		t.Error("activated outside a sorcery window")
	}
}

func TestS82OldB682AerialCaravanExilesTheTopCardForYouToPlay(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	top := o682Top(g, me, "Top Land", "Basic Land — Forest")
	car := effectstest.B12Push(g, me.ID, "Aerial Caravan", "Creature — Human Soldier", o682Caravan, 4, 3)
	effectstest.FloatForTest(g, me, "CUU")
	effectstest.ApaActivate(t, g, me, car, 0, game.ActivateAbilityParams{})
	if !g.Exile.Contains(top) {
		t.Fatal("the top card was not exiled")
	}
	if err := g.CastSpell(me.ID, top, game.CastSpellParams{FromZone: "exile", Strict: true}); err != nil {
		t.Fatalf("the exiled land is not playable: %v", err)
	}
	if effectstest.FindBattlefieldCardForTest(g, top) == nil {
		t.Error("the exiled land did not reach the battlefield")
	}
}

func TestS82OldB682DawnstriderPreventsAllCombatDamageForADiscard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	dawn := o682Perm(g, me.ID, "Dawnstrider", "Creature — Dryad Spellshaper", o682Dawn, 1, 1)
	bear := effectstest.SeedBear(g, me.ID)
	effectstest.FloatForTest(g, me, "G")
	if err := o682Activate(g, me, dawn, game.ActivateAbilityParams{}); err == nil {
		t.Fatal("activated without discarding")
	}
	spare := o682Hand(g, me)
	effectstest.ApaActivate(t, g, me, dawn, 0, game.ActivateAbilityParams{DiscardIDs: []uuid.UUID{spare}})
	life := opp.Life
	effectstest.AttackWith(t, g, opp.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life {
		t.Errorf("life %d → %d, want the combat damage prevented", life, opp.Life)
	}
}

func TestS82OldB682ReveredElderShieldsItselfOneDamageAtATime(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	elder := o682Perm(g, me.ID, "Revered Elder", "Creature — Human Cleric", o682Elder, 1, 2)
	src := effectstest.B12Creature(g, opp.ID, "Source", "Creature — Ogre", 2, 2)
	effectstest.FloatForTest(g, me, "C")
	effectstest.ApaActivate(t, g, me, elder, 0, game.ActivateAbilityParams{})
	effectstest.Cmdr2B08Damage(g, src, elder, 2)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.DamageMarkedOn(g, elder); n != 1 {
		t.Errorf("%d damage marked, want 1 (2 less the 1 prevented)", n)
	}
}

func TestS82OldB682ToymakerAnimatesANoncreatureArtifactToItsManaValue(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	toy := o682Perm(g, me.ID, "Toymaker", "Artifact Creature — Spellshaper", o682Toymaker, 1, 1)
	rock := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Big Rock", TypeLine: "Artifact", ManaCost: "{4}", Owner: me.ID, Controller: me.ID})
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.FloatForTest(g, me, "C")
	spare := o682Hand(g, me)
	if err := o682Activate(g, me, toy, game.ActivateAbilityParams{Targets: effectstest.CardRefs(bear), DiscardIDs: []uuid.UUID{spare}}); err == nil {
		t.Fatal("a creature was a legal target")
	}
	effectstest.ApaActivate(t, g, me, toy, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(rock), DiscardIDs: []uuid.UUID{spare}})
	if p, tough := effectstest.PtOf(t, g, rock); p != 4 || tough != 4 {
		t.Errorf("the artifact is %d/%d, want 4/4", p, tough)
	}
	c, _ := effectstest.BattlefieldCard(g, rock)
	if !c.IsCreature() {
		t.Error("the artifact is not a creature")
	}
}

func TestS82OldB682CacklingWitchPumpsByTheAnnouncedX(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	witch := o682Perm(g, me.ID, "Cackling Witch", "Creature — Human Spellshaper", o682Witch, 1, 1)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.FloatForTest(g, me, "CCCB")
	spare := o682Hand(g, me)
	effectstest.ApaActivate(t, g, me, witch, 0, game.ActivateAbilityParams{XValue: 3, Targets: effectstest.CardRefs(bear), DiscardIDs: []uuid.UUID{spare}})
	if p, tough := effectstest.PtOf(t, g, bear); p != 5 || tough != 2 {
		t.Errorf("the creature is %d/%d, want 5/2", p, tough)
	}
}

func TestS82OldB682SaprazzanBreakerIsUnblockableOnlyWhenALandIsMilled(t *testing.T) {
	for _, tc := range []struct {
		name     string
		typeLine string
		blocked  bool
	}{{"land", "Basic Land — Island", false}, {"spell", "Sorcery", true}} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			effectstest.AdvanceToMain(t, g)
			me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
			top := o682Top(g, me, "Top", tc.typeLine)
			br := effectstest.B12Push(g, me.ID, "Saprazzan Breaker", "Creature — Beast", o682Breaker, 3, 3)
			blocker := effectstest.B12Creature(g, opp.ID, "Blocker", "Creature — Wall", 0, 4)
			effectstest.FloatForTest(g, me, "U")
			effectstest.ApaActivate(t, g, me, br, 0, game.ActivateAbilityParams{})
			if !me.Graveyard.Contains(top) {
				t.Fatal("the top card was not milled")
			}
			if tc.blocked {
				effectstest.AssertRestrictions(t, g, br, 0)
			} else {
				effectstest.AssertRestrictions(t, g, br, game.CantBeBlocked)
				effectstest.AttackIntoBlocks(t, g, opp.ID, br)
				if reason := effectstest.BlockRefusal(t, g, blocker, br); reason == "" {
					t.Error("the Breaker could be blocked after milling a land")
				}
			}
		})
	}
}

func TestS82OldB682KyrenArchiveExilesBlindAndReturnsEverythingItExiled(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	top := o682Top(g, me, "Top", "Sorcery")
	arch := o682Perm(g, me.ID, "Kyren Archive", "Artifact — Book", o682Archive, 0, 0)
	effectstest.AdvanceToUpkeepOf(t, g, 0)
	effectstest.AssistCardSettle(t, g, me.ID, nil, true)
	if !g.Exile.Contains(top) {
		t.Fatal("the top card was not exiled")
	}
	if c := effectstest.ExiledCard(g, top); c != nil && c.IsKnownTo(me.ID) {
		t.Error("the controller can look at a card exiled face down by Kyren Archive")
	}
	effectstest.AdvanceToMain(t, g)
	held := o682Hand(g, me)
	effectstest.FloatForTest(g, me, "CCCCC")
	effectstest.ApaActivate(t, g, me, arch, 0, game.ActivateAbilityParams{})
	if !me.Graveyard.Contains(held) {
		t.Error("the hand was not discarded as a cost")
	}
	if effectstest.FindBattlefieldCardForTest(g, arch) != nil {
		t.Error("the Archive was not sacrificed")
	}
	if !me.Hand.Contains(top) {
		t.Error("the exiled card did not return to its owner's hand")
	}
}

func TestS82OldB682BlockadeRunnerCantBeBlockedThisTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	run := effectstest.B12Push(g, me.ID, "Blockade Runner", "Creature — Merfolk", o682Runner, 2, 2)
	blocker := effectstest.B12Creature(g, opp.ID, "Blocker", "Creature — Wall", 0, 4)
	effectstest.FloatForTest(g, me, "U")
	effectstest.ApaActivate(t, g, me, run, 0, game.ActivateAbilityParams{})
	effectstest.AssertRestrictions(t, g, run, game.CantBeBlocked)
	effectstest.AttackIntoBlocks(t, g, opp.ID, run)
	if reason := effectstest.BlockRefusal(t, g, blocker, run); reason == "" {
		t.Error("the Runner could be blocked")
	}
}

func TestS82OldB682AlabasterWallShieldsAnyTargetFromOneDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	wall := o682Perm(g, me.ID, "Alabaster Wall", "Creature — Wall", o682Alabaster, 0, 4)
	src := effectstest.B12Creature(g, opp.ID, "Source", "Creature — Ogre", 3, 3)
	victim := effectstest.B12Creature(g, me.ID, "Victim", "Creature — Bear", 4, 4)
	effectstest.ApaActivate(t, g, me, wall, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(victim)})
	effectstest.Cmdr2B08Damage(g, src, victim, 3)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.DamageMarkedOn(g, victim); n != 2 {
		t.Errorf("%d damage marked, want 2 (3 less the 1 prevented)", n)
	}
}
