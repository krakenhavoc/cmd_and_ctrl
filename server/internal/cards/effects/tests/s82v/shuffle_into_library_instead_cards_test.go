package testss82v

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Discarded or milled, each card is shuffled into its owner's library
// and never reaches the graveyard.
func TestShuffleInsteadCardsFromHandAndLibrary(t *testing.T) {
	for _, card := range []struct{ name, typeLine, oracle string }{
		{"Darksteel Colossus", "Artifact Creature — Golem", effectstest.DarksteelColossusOracle},
		{"Progenitus", "Legendary Creature — Hydra Avatar", effectstest.ProgenitusOracle},
		{"Legacy Weapon", "Legendary Artifact", effectstest.LegacyWeaponOracle},
	} {
		t.Run(card.name+" discarded", func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[0]
			me.Hand.Cards = nil
			id := uuid.New()
			me.Hand.PushTop(game.Card{InstanceID: id, Name: card.name, TypeLine: card.typeLine, OracleID: card.oracle, Owner: me.ID, Controller: me.ID})
			cut := len(g.Events)
			g.WithWriteLock(func() {
				if err := g.DiscardRandomForEffect(me.ID, 1); err != nil {
					t.Fatalf("discard: %v", err)
				}
			})
			effectstest.PassPriorityAroundTable(t, g)
			if me.Graveyard.Contains(id) || !me.Library.Contains(id) {
				t.Errorf("in graveyard %v, in library %v; want it shuffled into the library", me.Graveyard.Contains(id), me.Library.Contains(id))
			}
			if n := effectstest.BlightsteelShuffleCount(g.Events[cut:], me.ID); n != 1 {
				t.Errorf("shuffles = %d, want 1", n)
			}
		})
		t.Run(card.name+" milled", func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[0]
			id := uuid.New()
			me.Library.PushTop(game.Card{InstanceID: id, Name: card.name, TypeLine: card.typeLine, OracleID: card.oracle, Owner: me.ID, Controller: me.ID})
			g.WithWriteLock(func() {
				if err := g.MillNForEffect(me.ID, 1); err != nil {
					t.Fatalf("mill: %v", err)
				}
			})
			effectstest.PassPriorityAroundTable(t, g)
			if me.Graveyard.Contains(id) || !me.Library.Contains(id) {
				t.Errorf("in graveyard %v, in library %v; want it shuffled into the library", me.Graveyard.Contains(id), me.Library.Contains(id))
			}
		})
	}
}

// Against an opponent's Rest in Peace the Colossus's controller chooses
// (CR 616.1): its own replacement first sends it to the library, where
// Rest in Peace no longer applies; Rest in Peace first exiles it.
func TestDarksteelColossusAgainstRestInPeaceIsItsControllersChoice(t *testing.T) {
	for _, tc := range []struct {
		name      string
		colossus  bool
		toLibrary bool
	}{
		{"its own replacement first", true, true},
		{"Rest in Peace first", false, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			rip := effectstest.SeedReplacementPermanent(g, effectstest.RestInPeaceOracle, "Rest in Peace", opp.ID)
			colossus := effectstest.PushBattlefieldCardWithTimestamp(g, effectstest.DarksteelColossusCard(me.ID))
			effectstest.SacrificeNow(t, g, me.ID, colossus)
			prompt := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceReplacementOrder, me.ID)
			if prompt == nil {
				t.Fatal("no CR 616 prompt for the Colossus's controller")
			}
			own := effectstest.ReplacementIDForSource(t, g, prompt.ReplacementEffectIDs, colossus)
			theirs := effectstest.ReplacementIDForSource(t, g, prompt.ReplacementEffectIDs, rip)
			order := []game.ReplacementEffectID{theirs, own}
			if tc.colossus {
				order = []game.ReplacementEffectID{own, theirs}
			}
			if err := g.ResolveReplacementOrder(prompt.ID, me.ID, order); err != nil {
				t.Fatalf("ResolveReplacementOrder: %v", err)
			}
			effectstest.PassPriorityAroundTable(t, g)
			if got := me.Library.Contains(colossus); got != tc.toLibrary {
				t.Errorf("in library %v, want %v", got, tc.toLibrary)
			}
			if got := g.Exile.Contains(colossus); got == tc.toLibrary {
				t.Errorf("in exile %v, want %v", got, !tc.toLibrary)
			}
			if me.Graveyard.Contains(colossus) {
				t.Error("the Colossus reached the graveyard")
			}
		})
	}
}

// Progenitus as a commander: its replacement sends it to the library,
// and that move is offered the command zone (CR 903.9b, 616.2).
func TestProgenitusCommanderIsOfferedTheCommandZone(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	id := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Progenitus",
		TypeLine: "Legendary Creature — Hydra Avatar", OracleID: effectstest.ProgenitusOracle, Power: 10, Toughness: 10,
		Owner: me.ID, Controller: me.ID, IsCommander: true})
	effectstest.SacrificeNow(t, g, me.ID, id)
	offer := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceOptionalReplacement, me.ID)
	if offer == nil {
		t.Fatal("the commander was not offered the command zone")
	}
	if err := g.ResolveOptionalReplacement(offer.ID, me.ID, true); err != nil {
		t.Fatalf("ResolveOptionalReplacement: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Command.Contains(id) {
		t.Errorf("Progenitus is not in the command zone (library %v, graveyard %v)", me.Library.Contains(id), me.Graveyard.Contains(id))
	}
}

// Legacy Weapon's ability exiles a target permanent.
func TestLegacyWeaponExilesTargetPermanent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	weapon := effectstest.PushCatalogPermanent(g, me.ID, "Legacy Weapon", "Legendary Artifact", effectstest.LegacyWeaponOracle, false)
	victim := effectstest.Pr7Creature(g, opp.ID, "Victim", 2, "G")
	effectstest.ToMain(t, g)
	effectstest.Pr7Activate(t, g, me.ID, weapon, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{{Kind: game.TargetCard, ID: victim}}})
	if g.Battlefield.Contains(victim) || !g.Exile.Contains(victim) {
		t.Errorf("victim on battlefield %v, in exile %v; want exiled", g.Battlefield.Contains(victim), g.Exile.Contains(victim))
	}
}
