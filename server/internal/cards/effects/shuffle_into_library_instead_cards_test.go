package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// shuffle_into_library_instead_cards_test.go — #2038's replacement
// siblings: Darksteel Colossus, Progenitus and Legacy Weapon print
// Blightsteel Colossus's "If ~ would be put into a graveyard from
// anywhere, reveal ~ and shuffle it into its owner's library instead"
// (ShuffleIntoOwnersLibraryInstead). The tests prove what the cards add
// beside it: a competing Rest in Peace is the affected card's
// controller's choice (CR 616.1), and a commander sent to the library
// is still offered the command zone (CR 903.9b, 616.2).

const (
	darksteelColossusOracle = "1b09d0cf-403c-4a15-aeee-602a1bdaf0c1"
	progenitusOracle        = "2d7e00b6-12f0-4b03-82a6-50e3d1b5395e"
	legacyWeaponOracle      = "aa1a248d-2f76-4f15-b065-08f299af07b9"
)

func darksteelColossusCard(owner uuid.UUID) game.Card {
	return game.Card{InstanceID: uuid.New(), Name: "Darksteel Colossus", TypeLine: "Artifact Creature — Golem",
		OracleID: darksteelColossusOracle, ManaCost: "{11}", Power: 11, Toughness: 11, Owner: owner, Controller: owner}
}

// sacrificeNow sacrifices a permanent outside any resolution.
func sacrificeNow(t *testing.T, g *game.Game, controller, id uuid.UUID) {
	t.Helper()
	g.WithWriteLock(func() {
		if err := (SacrificePermanent{Target: id}).Apply(NewContext(g, &game.StackItem{Controller: controller})); err != nil {
			t.Fatalf("sacrifice: %v", err)
		}
	})
}

// Discarded or milled, each card is shuffled into its owner's library
// and never reaches the graveyard.
func TestShuffleInsteadCardsFromHandAndLibrary(t *testing.T) {
	for _, card := range []struct{ name, typeLine, oracle string }{
		{"Darksteel Colossus", "Artifact Creature — Golem", darksteelColossusOracle},
		{"Progenitus", "Legendary Creature — Hydra Avatar", progenitusOracle},
		{"Legacy Weapon", "Legendary Artifact", legacyWeaponOracle},
	} {
		t.Run(card.name+" discarded", func(t *testing.T) {
			g := newCatalogGame(t)
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
			passPriorityAroundTable(t, g)
			if me.Graveyard.Contains(id) || !me.Library.Contains(id) {
				t.Errorf("in graveyard %v, in library %v; want it shuffled into the library", me.Graveyard.Contains(id), me.Library.Contains(id))
			}
			if n := blightsteelShuffleCount(g.Events[cut:], me.ID); n != 1 {
				t.Errorf("shuffles = %d, want 1", n)
			}
		})
		t.Run(card.name+" milled", func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[0]
			id := uuid.New()
			me.Library.PushTop(game.Card{InstanceID: id, Name: card.name, TypeLine: card.typeLine, OracleID: card.oracle, Owner: me.ID, Controller: me.ID})
			g.WithWriteLock(func() {
				if err := g.MillNForEffect(me.ID, 1); err != nil {
					t.Fatalf("mill: %v", err)
				}
			})
			passPriorityAroundTable(t, g)
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
			g := newCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			rip := seedReplacementPermanent(g, restInPeaceOracle, "Rest in Peace", opp.ID)
			colossus := pushBattlefieldCardWithTimestamp(g, darksteelColossusCard(me.ID))
			sacrificeNow(t, g, me.ID, colossus)
			prompt := latestChoiceOfKindFor(g, game.PendingChoiceReplacementOrder, me.ID)
			if prompt == nil {
				t.Fatal("no CR 616 prompt for the Colossus's controller")
			}
			own := replacementIDForSource(t, g, prompt.ReplacementEffectIDs, colossus)
			theirs := replacementIDForSource(t, g, prompt.ReplacementEffectIDs, rip)
			order := []game.ReplacementEffectID{theirs, own}
			if tc.colossus {
				order = []game.ReplacementEffectID{own, theirs}
			}
			if err := g.ResolveReplacementOrder(prompt.ID, me.ID, order); err != nil {
				t.Fatalf("ResolveReplacementOrder: %v", err)
			}
			passPriorityAroundTable(t, g)
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
	g := newCatalogGame(t)
	me := g.Seats[0]
	id := pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Progenitus",
		TypeLine: "Legendary Creature — Hydra Avatar", OracleID: progenitusOracle, Power: 10, Toughness: 10,
		Owner: me.ID, Controller: me.ID, IsCommander: true})
	sacrificeNow(t, g, me.ID, id)
	offer := latestChoiceOfKindFor(g, game.PendingChoiceOptionalReplacement, me.ID)
	if offer == nil {
		t.Fatal("the commander was not offered the command zone")
	}
	if err := g.ResolveOptionalReplacement(offer.ID, me.ID, true); err != nil {
		t.Fatalf("ResolveOptionalReplacement: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !me.Command.Contains(id) {
		t.Errorf("Progenitus is not in the command zone (library %v, graveyard %v)", me.Library.Contains(id), me.Graveyard.Contains(id))
	}
}

// Legacy Weapon's ability exiles a target permanent.
func TestLegacyWeaponExilesTargetPermanent(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	weapon := pushCatalogPermanent(g, me.ID, "Legacy Weapon", "Legendary Artifact", legacyWeaponOracle, false)
	victim := pr7Creature(g, opp.ID, "Victim", 2, "G")
	toMain(t, g)
	pr7Activate(t, g, me.ID, weapon, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{{Kind: game.TargetCard, ID: victim}}})
	if g.Battlefield.Contains(victim) || !g.Exile.Contains(victim) {
		t.Errorf("victim on battlefield %v, in exile %v; want exiled", g.Battlefield.Contains(victim), g.Exile.Contains(victim))
	}
}
