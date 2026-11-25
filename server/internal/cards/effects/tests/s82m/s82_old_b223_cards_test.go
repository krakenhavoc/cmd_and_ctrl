package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b223_cards_test.go — Ixalan raid, batch 223 (#5532). Each test plays
// the real catalog entry and asserts what its text says.

const (
	oldB223Tormentor = "ab735456-f831-4988-84c6-50ca14351990"
	oldB223Looter    = "427bf4eb-3dfd-4747-be92-4a4a3d96a228"
	oldB223Arsonist  = "36081fbb-413f-4924-a536-82add34e38be"
	oldB223Ruin      = "cd830c46-3444-488d-ae5b-0a2533158a13"
	oldB223Aerialist = "b128d0cf-77d9-4751-99c6-0f3bc3660318"
	oldB223Barrage   = "a08c599e-ee75-4d2b-ae9b-0987c233dd17"
	oldB223Shipwreck = "ead1accc-b90e-424a-a55f-2f4771726ce1"
	oldB223Wake      = "d5fd3060-358c-4a5b-aee9-1afcf6c506ba"
	oldB223Raider    = "d6a837d0-d4f9-45ed-bbd7-5e37a9ca3b56"
)

func TestOldB223CardsAreRegisteredAndWhole(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB223Tormentor: "Deadeye Tormentor", oldB223Looter: "Marauding Looter", oldB223Arsonist: "Storm Fleet Arsonist",
		oldB223Ruin: "Navigator's Ruin", oldB223Aerialist: "Storm Fleet Aerialist", oldB223Barrage: "Repeating Barrage",
		oldB223Shipwreck: "Shipwreck Looter", oldB223Wake: "Raiders' Wake", oldB223Raider: "Ruin Raider",
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

// oldB223Raid makes the active seat have attacked this turn, then moves to the
// postcombat main phase.
func oldB223Raid(t *testing.T, g *game.Game, attacked bool) (me, opp *game.Player) {
	t.Helper()
	me, opp = effectstest.SeatsForCombat(g)
	if attacked {
		attacker := effectstest.B12Creature(g, me.ID, "Attacker", "Creature — Bear", 2, 2)
		effectstest.DeclareAttack(t, g, opp.ID, attacker)
		effectstest.AdvanceTo(t, g, game.StepPostcombatMain)
	} else {
		effectstest.AdvanceToMain(t, g)
	}
	return me, opp
}

func TestOldB223DeadeyeTormentorDiscardsOnlyAfterYouAttacked(t *testing.T) {
	for _, attacked := range []bool{false, true} {
		g := effectstest.NewCatalogGame(t)
		me, opp := oldB223Raid(t, g, attacked)
		card := uuid.New()
		opp.Hand.PushTop(game.Card{InstanceID: card, Name: "Their Card", TypeLine: "Sorcery", Owner: opp.ID, Controller: opp.ID})
		hand := opp.Hand.Size()
		effectstest.CastAndResolveCreature(t, g, "Deadeye Tormentor", "Creature — Human Pirate", oldB223Tormentor)
		if attacked {
			effectstest.B04WaitForPick(t, g, me.ID)
			effectstest.PickPlayer(t, g, me.ID, opp.ID)
			effectstest.PassPriorityAroundTable(t, g)
			if effectstest.DiscardChoiceFor(g, opp.ID) != nil {
				effectstest.AnswerDiscard(t, g, opp.ID, card)
			}
		}
		effectstest.PassPriorityAroundTable(t, g)
		want := hand
		if attacked {
			want = hand - 1
		}
		if opp.Hand.Size() != want {
			t.Errorf("attacked=%v: opponent hand %d, want %d", attacked, opp.Hand.Size(), want)
		}
	}
}

func TestOldB223StormFleetAerialistGetsACounterOnlyAfterYouAttacked(t *testing.T) {
	for _, attacked := range []bool{false, true} {
		g := effectstest.NewCatalogGame(t)
		me, _ := oldB223Raid(t, g, attacked)
		id := effectstest.CastAndResolveCreature(t, g, "Storm Fleet Aerialist", "Creature — Human Pirate", oldB223Aerialist)
		effectstest.PassPriorityAroundTable(t, g)
		c := effectstest.FindBattlefieldCardByID(g, id)
		if c == nil {
			t.Fatalf("attacked=%v: the Aerialist is not on the battlefield", attacked)
		}
		want := 0
		if attacked {
			want = 1
		}
		if c.Counters[game.CounterPlusOne] != want || c.Controller != me.ID {
			t.Errorf("attacked=%v: %d +1/+1 counters, want %d", attacked, c.Counters[game.CounterPlusOne], want)
		}
	}
}

func TestOldB223StormFleetArsonistMakesTheOpponentSacrificeAfterRaid(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := oldB223Raid(t, g, true)
	victim := effectstest.B12Creature(g, opp.ID, "Their Only Permanent", "Creature — Bear", 2, 2)
	effectstest.CastAndResolveCreature(t, g, "Storm Fleet Arsonist", "Creature — Orc Pirate", oldB223Arsonist)
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.PickPlayer(t, g, me.ID, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.SacrificeChoiceFor(g, opp.ID) == nil {
		t.Fatal("the opponent was not asked what to sacrifice")
	}
	effectstest.AnswerSacrifice(t, g, opp.ID, victim)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(victim) {
		t.Error("the opponent kept their only permanent")
	}
}

func TestOldB223ShipwreckLooterLootsOnlyAfterRaid(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := oldB223Raid(t, g, true)
	effectstest.SeedLibrary(me, "a", "b", "c")
	keep := effectstest.B13HandCard(me, "Junk", "Sorcery", "")
	gy := me.Graveyard.Size()
	effectstest.CastAndResolveCreature(t, g, "Shipwreck Looter", "Creature — Human Pirate", oldB223Shipwreck)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerMayChoice(t, g, me.ID, true)
	if effectstest.DiscardChoiceFor(g, me.ID) != nil {
		effectstest.AnswerDiscard(t, g, me.ID, keep)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Graveyard.Size() != gy+1 {
		t.Errorf("graveyard %d, want %d (one discard)", me.Graveyard.Size(), gy+1)
	}
}

func TestOldB223MaraudingLooterLootsAtTheEndStepOnlyAfterRaid(t *testing.T) {
	for _, attacked := range []bool{false, true} {
		g := effectstest.NewCatalogGame(t)
		me, _ := effectstest.SeatsForCombat(g)
		effectstest.PushCatalogPermanent(g, me.ID, "Marauding Looter", "Creature — Human Pirate", oldB223Looter, false)
		effectstest.SeedLibrary(me, "a", "b", "c")
		if attacked {
			attacker := effectstest.B12Creature(g, me.ID, "Attacker", "Creature — Bear", 2, 2)
			effectstest.DeclareAttack(t, g, g.Seats[1].ID, attacker)
		}
		hand := me.Hand.Size()
		effectstest.AdvanceToEndStep(t, g)
		if !attacked {
			effectstest.PassPriorityAroundTable(t, g)
			if effectstest.LatestConfirmFor(g, me.ID) != nil || me.Hand.Size() != hand {
				t.Error("a loot was offered without an attack")
			}
			continue
		}
		effectstest.PassPriorityAroundTable(t, g)
		effectstest.AnswerMayChoice(t, g, me.ID, true)
		if effectstest.DiscardChoiceFor(g, me.ID) != nil {
			effectstest.AnswerDiscard(t, g, me.ID, me.Hand.Cards[0].InstanceID)
		}
		effectstest.PassPriorityAroundTable(t, g)
		if me.Hand.Size() != hand {
			t.Errorf("hand %d, want %d (draw one, discard one)", me.Hand.Size(), hand)
		}
	}
}

func TestOldB223RuinRaiderRevealsAndLosesLifeEqualToManaValue(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.SeatsForCombat(g)
	effectstest.PushCatalogPermanent(g, me.ID, "Ruin Raider", "Creature — Orc Pirate", oldB223Raider, false)
	top := effectstest.SeedLibraryTop(me, "Big Spell", "Sorcery", "{3}{R}")
	attacker := effectstest.B12Creature(g, me.ID, "Attacker", "Creature — Bear", 2, 2)
	effectstest.DeclareAttack(t, g, opp.ID, attacker)
	life := me.Life
	effectstest.AdvanceToEndStep(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(top) {
		t.Fatal("the top card did not reach the hand")
	}
	if me.Life != life-4 {
		t.Errorf("life %d -> %d, want -4 (mana value 4)", life, me.Life)
	}
}

func TestOldB223NavigatorsRuinMillsFourAfterRaid(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.SeatsForCombat(g)
	effectstest.PushCatalogPermanent(g, me.ID, "Navigator's Ruin", "Enchantment", oldB223Ruin, false)
	attacker := effectstest.B12Creature(g, me.ID, "Attacker", "Creature — Bear", 2, 2)
	effectstest.DeclareAttack(t, g, opp.ID, attacker)
	lib := opp.Library.Size()
	effectstest.AdvanceToEndStep(t, g)
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.PickPlayer(t, g, me.ID, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Library.Size() != lib-4 {
		t.Errorf("library %d -> %d, want -4", lib, opp.Library.Size())
	}
}

func TestOldB223RaidersWakeDiscardsAtTheEndStepAndDrainsForEachDiscard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.SeatsForCombat(g)
	effectstest.PushCatalogPermanent(g, me.ID, "Raiders' Wake", "Enchantment", oldB223Wake, false)
	card := uuid.New()
	opp.Hand.PushTop(game.Card{InstanceID: card, Name: "Their Card", TypeLine: "Sorcery", Owner: opp.ID, Controller: opp.ID})
	attacker := effectstest.B12Creature(g, me.ID, "Attacker", "Creature — Bear", 2, 2)
	effectstest.DeclareAttack(t, g, opp.ID, attacker)
	hand := opp.Hand.Size()
	effectstest.AdvanceToEndStep(t, g)
	life := opp.Life
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.PickPlayer(t, g, me.ID, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.DiscardChoiceFor(g, opp.ID) != nil {
		effectstest.AnswerDiscard(t, g, opp.ID, card)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := hand - opp.Hand.Size(); got != 1 {
		t.Errorf("opponent discarded %d, want 1", got)
	}
	if opp.Life != life-2 {
		t.Errorf("opponent life %d -> %d, want -2 for the discard", life, opp.Life)
	}
}

func TestOldB223RepeatingBarrageDamagesAndReturnsOnlyAfterRaid(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.SeatsForCombat(g)
	life := opp.Life
	effectstest.CastCatalogSpell(t, g, "Repeating Barrage", "Sorcery", oldB223Barrage,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-3 {
		t.Fatalf("opponent life %d -> %d, want -3", life, opp.Life)
	}
	var barrage uuid.UUID
	for _, c := range me.Graveyard.Cards {
		if c.OracleID == oldB223Barrage {
			barrage = c.InstanceID
		}
	}
	if barrage == uuid.Nil {
		t.Fatal("the Barrage is not in the graveyard")
	}
	effectstest.B10AddMana(me, "R", "R", "R", "R", "R")
	if err := g.ActivateCatalogAbility(me.ID, barrage, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("returned the Barrage without having attacked")
	}
	attacker := effectstest.B12Creature(g, me.ID, "Attacker", "Creature — Bear", 2, 2)
	effectstest.DeclareAttack(t, g, opp.ID, attacker)
	effectstest.AdvanceTo(t, g, game.StepPostcombatMain)
	effectstest.B10AddMana(me, "R", "R", "R", "R", "R")
	if err := g.ActivateCatalogAbility(me.ID, barrage, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate after attacking: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(barrage) {
		t.Error("the Barrage did not return to the hand")
	}
}
