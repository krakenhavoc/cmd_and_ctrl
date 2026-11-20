package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// tsp_b1_cards_test.go — Time Spiral, batch 1 (#5596, S82): thirty-nine
// cards built from seams the engine already has. Each test plays the real
// catalog entry.

const (
	tspB1MightSliverOracle          = "92f5f80c-de1a-44e8-93f1-e405241c1e82"
	tspB1BonesplitterSliverOracle   = "a0f6834c-faa9-4d90-adad-847a8b337fe6"
	tspB1WatcherSliverOracle        = "9a538bd7-e665-4896-95df-bbab1e28f75e"
	tspB1TwoHeadedSliverOracle      = "be857d6e-5bf2-491b-ab46-963b4d45adee"
	tspB1FurySliverOracle           = "44623693-51d6-49ad-8cd7-140505caf02f"
	tspB1SpinneretSliverOracle      = "fbdec63e-0679-47e9-b82a-daa56df50481"
	tspB1ShadowSliverOracle         = "46ddc2e1-386f-4aaf-b37b-37b1708e0d9c"
	tspB1SidewinderSliverOracle     = "de8ccd32-f537-4a11-8dc4-e48836a0b139"
	tspB1FirewakeSliverOracle       = "b5a3c09d-e822-4e26-b569-7f5830a39e6a"
	tspB1TelekineticSliverOracle    = "31ec6a73-f742-4dc6-809d-28d3c461c2a0"
	tspB1ScreechingSliverOracle     = "88485575-e491-4ab7-a020-e52dd92ae085"
	tspB1PsionicSliverOracle        = "ea79d4ee-c7b5-4687-9e33-22f98e46e8ec"
	tspB1QuilledSliverOracle        = "beb39342-5272-41a4-a470-9f2bf0cd0f92"
	tspB1MindlashSliverOracle       = "7a08b371-1353-476d-ae72-379594f801da"
	tspB1BasalSliverOracle          = "18e7ba0d-88a5-466b-9228-d3f5aae8f83e"
	tspB1PlagueSliverOracle         = "ee10dc06-26e1-4f50-a0a6-6d039e71b75f"
	tspB1HarmonicSliverOracle       = "1c06c093-18ab-49e6-a82d-38e826d25016"
	tspB1WheelOfFateOracle          = "15e4938d-1cd3-4670-9de5-2dc10abe4c19"
	tspB1RiftwingCloudskateOracle   = "ae2e6ffd-fb93-4c47-88e7-e84460eefbe5"
	tspB1DeepSeaKrakenOracle        = "54de4a7f-86f4-4029-80f7-bc25d6509cc8"
	tspB1ErrantEphemeronOracle      = "c77b89ec-4514-4e62-a508-2012f38a3e15"
	tspB1DurkwoodBalothOracle       = "1732cd31-1393-426f-a2fe-9b6b3f27136b"
	tspB1KeldonHalberdierOracle     = "a1472a2d-e6f7-4d65-8bb2-4ecfae2d2764"
	tspB1CorpulentCorpseOracle      = "b3a84b75-a6a2-4521-a9f1-92596796c703"
	tspB1DuskriderPeregrineOracle   = "2fdde5f7-a235-43a4-88a5-2faaf5900bd6"
	tspB1MindstabOracle             = "338c06ec-47b5-47e7-9fe1-39ff381b36a5"
	tspB1PlunderOracle              = "8a515f67-f868-494b-8bd7-de63fe14477b"
	tspB1DivineCongregationOracle   = "a2fc892b-3840-45a0-bb1c-0d2cf7d14f2f"
	tspB1AncientGrudgeOracle        = "306593b0-6ea8-476f-a3f2-e17876c1bab4"
	tspB1StranglingSootOracle       = "6c9d3ad0-1ede-41de-a1b4-15de9c728216"
	tspB1ThrillOfTheHuntOracle      = "7b1f0b78-806f-4a7d-809b-d2bc60cf1b01"
	tspB1DarkWitheringOracle        = "ac2720d2-b010-4fa9-8c94-b13680b75aac"
	tspB1CallToTheNetherworldOracle = "a28ef7a5-4b47-4429-838b-a64a7c0701d0"
	tspB1WipeAwayOracle             = "ad9b6737-22c4-45cc-976b-32743ed48462"
	tspB1StonewoodInvocationOracle  = "35f09a11-1d9b-4dfd-9695-b809c871f10d"
	tspB1WordOfSeizingOracle        = "b80d2841-03b0-42d3-ac9b-51e793b1dc24"
	tspB1PenumbraSpiderOracle       = "47678f44-b595-492e-af00-4e73f941e24e"
	tspB1AssassinateOracle          = "ce979f36-84f2-4419-b400-971263494cc6"
	tspB1SproutOracle               = "424c6f5e-b386-47e9-b3fe-25b263097d40"
)

func TestTspB1CardsAreRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		tspB1MightSliverOracle:          "Might Sliver",
		tspB1BonesplitterSliverOracle:   "Bonesplitter Sliver",
		tspB1WatcherSliverOracle:        "Watcher Sliver",
		tspB1TwoHeadedSliverOracle:      "Two-Headed Sliver",
		tspB1FurySliverOracle:           "Fury Sliver",
		tspB1SpinneretSliverOracle:      "Spinneret Sliver",
		tspB1ShadowSliverOracle:         "Shadow Sliver",
		tspB1SidewinderSliverOracle:     "Sidewinder Sliver",
		tspB1FirewakeSliverOracle:       "Firewake Sliver",
		tspB1TelekineticSliverOracle:    "Telekinetic Sliver",
		tspB1ScreechingSliverOracle:     "Screeching Sliver",
		tspB1PsionicSliverOracle:        "Psionic Sliver",
		tspB1QuilledSliverOracle:        "Quilled Sliver",
		tspB1MindlashSliverOracle:       "Mindlash Sliver",
		tspB1BasalSliverOracle:          "Basal Sliver",
		tspB1PlagueSliverOracle:         "Plague Sliver",
		tspB1HarmonicSliverOracle:       "Harmonic Sliver",
		tspB1WheelOfFateOracle:          "Wheel of Fate",
		tspB1RiftwingCloudskateOracle:   "Riftwing Cloudskate",
		tspB1DeepSeaKrakenOracle:        "Deep-Sea Kraken",
		tspB1ErrantEphemeronOracle:      "Errant Ephemeron",
		tspB1DurkwoodBalothOracle:       "Durkwood Baloth",
		tspB1KeldonHalberdierOracle:     "Keldon Halberdier",
		tspB1CorpulentCorpseOracle:      "Corpulent Corpse",
		tspB1DuskriderPeregrineOracle:   "Duskrider Peregrine",
		tspB1MindstabOracle:             "Mindstab",
		tspB1PlunderOracle:              "Plunder",
		tspB1DivineCongregationOracle:   "Divine Congregation",
		tspB1AncientGrudgeOracle:        "Ancient Grudge",
		tspB1StranglingSootOracle:       "Strangling Soot",
		tspB1ThrillOfTheHuntOracle:      "Thrill of the Hunt",
		tspB1DarkWitheringOracle:        "Dark Withering",
		tspB1CallToTheNetherworldOracle: "Call to the Netherworld",
		tspB1WipeAwayOracle:             "Wipe Away",
		tspB1StonewoodInvocationOracle:  "Stonewood Invocation",
		tspB1WordOfSeizingOracle:        "Word of Seizing",
		tspB1PenumbraSpiderOracle:       "Penumbra Spider",
		tspB1AssassinateOracle:          "Assassinate",
		tspB1SproutOracle:               "Sprout",
	} {
		spec, ok := Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

// --- helpers -----------------------------------------------------------------

// tspB1Sliver seeds a non-catalog Sliver creature that has been under its
// controller's control since the turn began.
func tspB1Sliver(g *game.Game, owner uuid.UUID, name string, power, toughness int) uuid.UUID {
	return pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: "Creature — Sliver",
		Power: power, Toughness: toughness, Owner: owner, Controller: owner,
	})
}

// tspB1ActivateGranted activates the first granted activated ability of
// `card` for `controller` and settles the stack.
func tspB1ActivateGranted(t *testing.T, g *game.Game, controller, card uuid.UUID, targets []game.TargetRef) {
	t.Helper()
	idx, ref := grantedActivatedIndex(t, g, card)
	if idx < 0 {
		t.Fatal("the Sliver has no granted activated ability")
	}
	b16Activate(t, g, controller, card, idx, game.ActivateAbilityParams{Ref: ref, Targets: targets})
}

// tspB1Suspend suspends a card from the active seat's hand with `mana`
// floating, then leaves it one time counter from its free cast.
func tspB1Suspend(t *testing.T, g *game.Game, name, typeLine, oracle, mana string, n int) uuid.UUID {
	t.Helper()
	me := g.Seats[g.Turn.ActiveSeat]
	advanceTo(t, g, game.StepPrecombatMain)
	id := handCardForTest(me, name, typeLine, oracle)
	for i := range me.Hand.Cards {
		if me.Hand.Cards[i].InstanceID == id {
			me.Hand.Cards[i].Power, me.Hand.Cards[i].Toughness = 4, 4 // read only by a creature
		}
	}
	floatMana(t, g, me, mana)
	if err := g.PerformSpecialAction(me.ID, id, game.SpecialActionSuspend, game.SpecialActionParams{Strict: true}); err != nil {
		t.Fatalf("suspend %s: %v", name, err)
	}
	if got := counterCountInExile(g, id, game.CounterTime); got != n {
		t.Fatalf("%s has %d time counters after suspending, want %d", name, got, n)
	}
	if n > 1 {
		g.WithWriteLock(func() {
			if err := g.AddCounterForEffect(id, game.CounterTime, -(n - 1)); err != nil {
				t.Fatalf("remove time counters: %v", err)
			}
		})
	}
	return id
}

// tspB1FreeCastAtUpkeep walks to the owner's next upkeep, lets the last
// time counter come off and casts the card from exile for free.
func tspB1FreeCastAtUpkeep(t *testing.T, g *game.Game, id uuid.UUID, params game.CastSpellParams) {
	t.Helper()
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	advanceToUpkeepOf(t, g, seat)
	passPriorityAroundTable(t, g)
	if latestChoiceOfKind(g, game.PendingChoiceResolutionCast) == nil {
		t.Fatal("the last time counter coming off offered no cast")
	}
	params.Strict, params.FromZone = true, "exile"
	if err := g.CastSpell(me.ID, id, params); err != nil {
		t.Fatalf("the free cast from suspend: %v", err)
	}
	passPriorityAroundTable(t, g)
}

// tspB1Loot discards `card` from the active seat's otherwise-empty hand
// through Faithless Looting, so a madness card is exiled and offered.
func tspB1Loot(t *testing.T, g *game.Game, name, typeLine, oracle string) uuid.UUID {
	t.Helper()
	me := g.Seats[g.Turn.ActiveSeat]
	g.WithWriteLock(func() { me.Hand.Cards = nil })
	id := pushCatalogHandCard(me, name, typeLine, oracle)
	spare := pushCatalogHandCard(me, "Spare", "Sorcery", "")
	castCatalogSpell(t, g, "Faithless Looting", "Sorcery", faithlessLootingOracl, nil)
	passPriorityAroundTable(t, g)
	answerChooseCards(t, g, me.ID, id, spare)
	passPriorityAroundTable(t, g)
	if !g.Exile.Contains(id) {
		t.Fatalf("%s was not discarded into exile (CR 702.35a)", name)
	}
	if latestChoiceOfKind(g, game.PendingChoiceResolutionCast) == nil {
		t.Fatalf("%s: the madness trigger offered no cast", name)
	}
	return id
}

// --- sliver anthems and keyword grants ---------------------------------------

func TestTspB1SliverAnthemsReachEverySliverCreature(t *testing.T) {
	for _, tc := range []struct {
		name, oracle string
		p, tough     int
	}{
		{"Might Sliver", tspB1MightSliverOracle, 2, 2},
		{"Bonesplitter Sliver", tspB1BonesplitterSliverOracle, 2, 0},
		{"Watcher Sliver", tspB1WatcherSliverOracle, 0, 2},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			mine := tspB1Sliver(g, me.ID, "My Sliver", 1, 1)
			theirs := tspB1Sliver(g, opp.ID, "Their Sliver", 1, 1)
			bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
			b12Push(g, me.ID, tc.name, "Creature — Sliver", tc.oracle, 2, 2)
			for _, id := range []uuid.UUID{mine, theirs} {
				if p, tough := effectivePower(t, g, id), effectiveToughness(t, g, id); p != 1+tc.p || tough != 1+tc.tough {
					t.Errorf("a Sliver is %d/%d, want %d/%d", p, tough, 1+tc.p, 1+tc.tough)
				}
			}
			if p, tough := effectivePower(t, g, bear), effectiveToughness(t, g, bear); p != 2 || tough != 2 {
				t.Errorf("the Bear is %d/%d, want 2/2", p, tough)
			}
		})
	}
}

func TestTspB1SliverKeywordGrantsReachEverySliverCreature(t *testing.T) {
	for _, tc := range []struct{ name, oracle, keyword string }{
		{"Two-Headed Sliver", tspB1TwoHeadedSliverOracle, "menace"},
		{"Fury Sliver", tspB1FurySliverOracle, "double strike"},
		{"Spinneret Sliver", tspB1SpinneretSliverOracle, "reach"},
		{"Shadow Sliver", tspB1ShadowSliverOracle, "shadow"},
		{"Sidewinder Sliver", tspB1SidewinderSliverOracle, game.KeywordFlanking},
		{"Firewake Sliver", tspB1FirewakeSliverOracle, "haste"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			theirs := tspB1Sliver(g, opp.ID, "Their Sliver", 1, 1)
			bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
			lord := b12Push(g, me.ID, tc.name, "Creature — Sliver", tc.oracle, 1, 1)
			if !hasKeywordOf(g, theirs, tc.keyword) || !hasKeywordOf(g, lord, tc.keyword) {
				t.Errorf("want every Sliver, an opponent's and this one included, to have %s", tc.keyword)
			}
			if hasKeywordOf(g, bear, tc.keyword) {
				t.Errorf("a Bear has %s", tc.keyword)
			}
		})
	}
}

// A blocker without flanking shrinks when it blocks a Sliver given
// flanking by Sidewinder Sliver.
func TestTspB1SidewinderSliverFlankingShrinksTheBlocker(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	attacker := tspB1Sliver(g, me.ID, "My Sliver", 2, 2)
	b12Push(g, opp.ID, "Sidewinder Sliver", "Creature — Sliver", tspB1SidewinderSliverOracle, 1, 1)
	blocker := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 3)
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(attacker, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	lockInAttacks(t, g)
	passPriorityAroundTable(t, g)
	advanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(blocker, attacker); err != nil {
		t.Fatalf("DeclareBlocker: %v", err)
	}
	lockInBlocks(t, g)
	passPriorityAroundTable(t, g)
	if p, tough := effectivePower(t, g, blocker), effectiveToughness(t, g, blocker); p != 1 || tough != 2 {
		t.Errorf("the blocker is %d/%d, want 1/2 after flanking", p, tough)
	}
}

// --- sliver granted abilities ------------------------------------------------

func TestTspB1FirewakeSliverSacrificesASliverToPumpAnother(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	fodder := tspB1Sliver(g, me.ID, "Fodder Sliver", 1, 1)
	other := tspB1Sliver(g, me.ID, "Other Sliver", 1, 1)
	b12Push(g, me.ID, "Firewake Sliver", "Creature — Sliver", tspB1FirewakeSliverOracle, 1, 1)
	advanceToMain(t, g)
	floatMana(t, g, me, "{C}")
	tspB1ActivateGranted(t, g, me.ID, fodder, cardRefs(other))
	if g.Battlefield.Contains(fodder) {
		t.Error("the Sliver that activated the ability was not sacrificed")
	}
	if p, tough := effectivePower(t, g, other), effectiveToughness(t, g, other); p != 3 || tough != 3 {
		t.Errorf("the target Sliver is %d/%d, want 3/3", p, tough)
	}
}

func TestTspB1TelekineticSliverTapsATargetPermanent(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	sliver := tspB1Sliver(g, me.ID, "My Sliver", 1, 1)
	b12Push(g, me.ID, "Telekinetic Sliver", "Creature — Sliver", tspB1TelekineticSliverOracle, 2, 2)
	rock := b12Permanent(g, opp.ID, "Their Rock", "Artifact")
	advanceToMain(t, g)
	tspB1ActivateGranted(t, g, me.ID, sliver, cardRefs(rock))
	if !b16Tapped(t, g, rock) || !b16Tapped(t, g, sliver) {
		t.Error("want the target and the Sliver both tapped")
	}
}

func TestTspB1ScreechingSliverMillsTheTargetPlayer(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	sliver := tspB1Sliver(g, me.ID, "My Sliver", 1, 1)
	b12Push(g, me.ID, "Screeching Sliver", "Creature — Sliver", tspB1ScreechingSliverOracle, 1, 1)
	advanceToMain(t, g)
	before := opp.Graveyard.Size()
	tspB1ActivateGranted(t, g, me.ID, sliver, playerRefs(opp.ID))
	if got := opp.Graveyard.Size(); got != before+1 {
		t.Errorf("opponent's graveyard %d -> %d, want one card milled", before, got)
	}
}

func TestTspB1PsionicSliverHitsTheTargetAndItself(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	big := tspB1Sliver(g, me.ID, "Big Sliver", 4, 4)
	small := tspB1Sliver(g, me.ID, "Small Sliver", 1, 1)
	b12Push(g, me.ID, "Psionic Sliver", "Creature — Sliver", tspB1PsionicSliverOracle, 2, 2)
	advanceToMain(t, g)
	before := opp.Life
	tspB1ActivateGranted(t, g, me.ID, big, playerRefs(opp.ID))
	if opp.Life != before-2 {
		t.Errorf("opponent life %d -> %d, want 2 damage", before, opp.Life)
	}
	if !g.Battlefield.Contains(big) {
		t.Fatal("a 4/4 Sliver died to 3 damage")
	}
	tspB1ActivateGranted(t, g, me.ID, small, playerRefs(opp.ID))
	if g.Battlefield.Contains(small) {
		t.Error("a 1/1 Sliver survived 3 damage to itself")
	}
}

func TestTspB1QuilledSliverPingsOnlyACreatureInCombat(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	sliver := b12Push(g, opp.ID, "Quilled Sliver", "Creature — Sliver", tspB1QuilledSliverOracle, 1, 1)
	attacker := b12Creature(g, me.ID, "My Elf", "Creature — Elf", 1, 1)
	idle := b12Creature(g, me.ID, "My Idler", "Creature — Elf", 1, 1)
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(attacker, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	lockInAttacks(t, g)
	idx, ref := grantedActivatedIndex(t, g, sliver)
	if idx < 0 {
		t.Fatal("Quilled Sliver has no granted ability")
	}
	if err := g.ActivateCatalogAbility(opp.ID, sliver, idx, game.ActivateAbilityParams{Ref: ref, Targets: cardRefs(idle)}); err == nil {
		t.Error("a creature that is neither attacking nor blocking was targeted")
	}
	b16Activate(t, g, opp.ID, sliver, idx, game.ActivateAbilityParams{Ref: ref, Targets: cardRefs(attacker)})
	if g.Battlefield.Contains(attacker) {
		t.Error("the attacking 1/1 survived")
	}
}

func TestTspB1MindlashSliverMakesEachPlayerDiscard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	sliver := tspB1Sliver(g, me.ID, "My Sliver", 1, 1)
	b12Push(g, me.ID, "Mindlash Sliver", "Creature — Sliver", tspB1MindlashSliverOracle, 1, 1)
	advanceToMain(t, g)
	floatMana(t, g, me, "{C}")
	idx, ref := grantedActivatedIndex(t, g, sliver)
	if idx < 0 {
		t.Fatal("the Sliver has no granted ability")
	}
	before := map[uuid.UUID]int{}
	for _, p := range g.Seats {
		before[p.ID] = p.Graveyard.Size()
	}
	if err := g.ActivateCatalogAbility(me.ID, sliver, idx, game.ActivateAbilityParams{Ref: ref}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if g.Battlefield.Contains(sliver) {
		t.Error("the Sliver was not sacrificed")
	}
	before[me.ID]++ // the sacrificed Sliver
	passPriorityAroundTable(t, g)
	// Each player picks one card of their own, in turn order.
	for i := 0; i < 8; i++ {
		answered := false
		for _, p := range g.Seats {
			if c := discardChoiceFor(g, p.ID); c != nil && p.Hand.Size() > 0 {
				answerDiscard(t, g, p.ID, p.Hand.Cards[0].InstanceID)
				answered = true
			}
		}
		if !answered {
			break
		}
		passPriorityAroundTable(t, g)
	}
	for _, p := range g.Seats {
		if got := p.Graveyard.Size(); got != before[p.ID]+1 {
			t.Errorf("%s's graveyard %d -> %d, want one discarded card", p.Name, before[p.ID], got)
		}
	}
}

func TestTspB1BasalSliverSacrificesASliverForBlackMana(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := tspB1Sliver(g, me.ID, "My Sliver", 1, 1)
	theirs := tspB1Sliver(g, opp.ID, "Their Sliver", 1, 1)
	b12Push(g, me.ID, "Basal Sliver", "Creature — Sliver", tspB1BasalSliverOracle, 2, 2)
	if grantedManaRowsOf(g, theirs) != 1 {
		t.Error("an opponent's Sliver has no granted mana ability")
	}
	if err := tapGrantedMana(t, g, me.ID, mine, ""); err != nil {
		t.Fatalf("sacrifice for mana: %v", err)
	}
	if g.Battlefield.Contains(mine) {
		t.Error("the Sliver was not sacrificed")
	}
	if got := batch01PoolColors(me); len(got) != 2 || got[0] != "B" || got[1] != "B" {
		t.Errorf("pool %v, want [B B]", got)
	}
}

func TestTspB1PlagueSliverHurtsEachSliversControllerInTheirUpkeep(t *testing.T) {
	g := newCatalogGame(t)
	owner, other := g.Seats[1], g.Seats[2]
	tspB1Sliver(g, owner.ID, "Owner's Sliver", 1, 1)
	b12Push(g, owner.ID, "Plague Sliver", "Creature — Sliver", tspB1PlagueSliverOracle, 5, 5)
	tspB1Sliver(g, other.ID, "Other's Sliver", 1, 1)
	before, otherBefore := owner.Life, other.Life
	advanceToUpkeepOf(t, g, 1)
	passPriorityAroundTable(t, g)
	if owner.Life != before-2 {
		t.Errorf("owner life %d -> %d, want 2 damage (two Slivers)", before, owner.Life)
	}
	if other.Life != otherBefore {
		t.Error("another player's Sliver hurt them in someone else's upkeep")
	}
}

func TestTspB1HarmonicSliverDestroysOnEachSliverEntering(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	first := b12Permanent(g, opp.ID, "Their Relic", "Artifact")
	second := b12Permanent(g, opp.ID, "Their Aura", "Enchantment")
	castHoldCreature(t, g, "Harmonic Sliver", "Creature — Sliver", tspB1HarmonicSliverOracle, 1, 1)
	b04WaitForPick(t, g, me.ID)
	answerPickTarget(t, g, first)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(first) {
		t.Fatal("Harmonic Sliver's own entry destroyed nothing")
	}
	castHoldCreature(t, g, "Muscle Sliver", "Creature — Sliver", "", 1, 1)
	b04WaitForPick(t, g, me.ID)
	answerPickTarget(t, g, second)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(second) {
		t.Error("another Sliver entering destroyed nothing")
	}
}

// --- suspend -----------------------------------------------------------------

func TestTspB1SuspendCardsDeclareTheirPrintedNumbers(t *testing.T) {
	for _, tc := range []struct {
		name, oracle string
		n            int
		cost         string
	}{
		{"Wheel of Fate", tspB1WheelOfFateOracle, 4, "{1}{R}"},
		{"Riftwing Cloudskate", tspB1RiftwingCloudskateOracle, 3, "{1}{U}"},
		{"Deep-Sea Kraken", tspB1DeepSeaKrakenOracle, 9, "{2}{U}"},
		{"Errant Ephemeron", tspB1ErrantEphemeronOracle, 4, "{1}{U}"},
		{"Durkwood Baloth", tspB1DurkwoodBalothOracle, 5, "{G}"},
		{"Keldon Halberdier", tspB1KeldonHalberdierOracle, 4, "{R}"},
		{"Corpulent Corpse", tspB1CorpulentCorpseOracle, 5, "{B}"},
		{"Duskrider Peregrine", tspB1DuskriderPeregrineOracle, 3, "{1}{W}"},
		{"Mindstab", tspB1MindstabOracle, 4, "{B}"},
		{"Plunder", tspB1PlunderOracle, 4, "{1}{R}"},
		{"Divine Congregation", tspB1DivineCongregationOracle, 5, "{1}{W}"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			s40SuspendAction(t, tc.oracle, tc.n, tc.cost)
		})
	}
}

func TestTspB1WheelOfFateIsCastFromSuspendAndRefillsEveryHand(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	id := tspB1Suspend(t, g, "Wheel of Fate", "Sorcery", tspB1WheelOfFateOracle, "{C}{R}", 4)
	tspB1FreeCastAtUpkeep(t, g, id, game.CastSpellParams{})
	if !me.Graveyard.Contains(id) {
		t.Error("the resolved Wheel of Fate is not in its owner's graveyard")
	}
	for _, p := range g.Seats {
		if p.Hand.Size() != 7 {
			t.Errorf("%s has %d cards, want 7", p.Name, p.Hand.Size())
		}
	}
}

func TestTspB1ErrantEphemeronArrivesFromSuspendWithHaste(t *testing.T) {
	g := newCatalogGame(t)
	id := tspB1Suspend(t, g, "Errant Ephemeron", "Creature — Illusion", tspB1ErrantEphemeronOracle, "{C}{U}", 4)
	tspB1FreeCastAtUpkeep(t, g, id, game.CastSpellParams{})
	if !g.Battlefield.Contains(id) {
		t.Fatal("the Ephemeron is not on the battlefield")
	}
	if !hasKeywordOf(g, id, "haste") || !hasKeywordOf(g, id, "flying") {
		t.Error("a creature cast from suspend has haste (CR 702.62a), and this one flies")
	}
}

func TestTspB1RiftwingCloudskateBouncesATargetPermanent(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	rock := b12Permanent(g, opp.ID, "Their Rock", "Artifact")
	castHoldCreature(t, g, "Riftwing Cloudskate", "Creature — Illusion", tspB1RiftwingCloudskateOracle, 2, 2)
	b04WaitForPick(t, g, me.ID)
	answerPickTarget(t, g, rock)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(rock) || !opp.Hand.Contains(rock) {
		t.Error("the target was not returned to its owner's hand")
	}
}

func TestTspB1DeepSeaKrakenLosesACounterWhenAnOpponentCasts(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceTo(t, g, game.StepPrecombatMain)
	id := handCardForTest(me, "Deep-Sea Kraken", "Creature — Kraken", tspB1DeepSeaKrakenOracle)
	floatMana(t, g, me, "{C}{C}{U}")
	if err := g.PerformSpecialAction(me.ID, id, game.SpecialActionSuspend, game.SpecialActionParams{Strict: true}); err != nil {
		t.Fatalf("suspend: %v", err)
	}
	// My own spell does not count.
	castCatalogSpell(t, g, "My Instant", "Instant", "", nil)
	passPriorityAroundTable(t, g)
	if got := counterCountInExile(g, id, game.CounterTime); got != 9 {
		t.Fatalf("%d time counters after my own spell, want 9", got)
	}
	spell := uuid.New()
	g.WithWriteLock(func() {
		opp.Hand.PushTop(game.Card{InstanceID: spell, Name: "Their Instant", TypeLine: "Instant", Owner: opp.ID, Controller: opp.ID})
	})
	if err := g.CastSpell(opp.ID, spell, game.CastSpellParams{}); err != nil {
		t.Fatalf("opponent casts: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := counterCountInExile(g, id, game.CounterTime); got != 8 {
		t.Errorf("%d time counters after an opponent's spell, want 8", got)
	}
}

func TestTspB1MindstabMakesTheTargetDiscardThree(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	castCatalogSpell(t, g, "Mindstab", "Sorcery", tspB1MindstabOracle, playerRefs(opp.ID))
	passPriorityAroundTable(t, g)
	if opp.Hand.Size() < 3 {
		t.Fatalf("the opponent holds %d cards", opp.Hand.Size())
	}
	before := opp.Graveyard.Size()
	answerDiscard(t, g, opp.ID, opp.Hand.Cards[0].InstanceID, opp.Hand.Cards[1].InstanceID, opp.Hand.Cards[2].InstanceID)
	passPriorityAroundTable(t, g)
	if got := opp.Graveyard.Size(); got != before+3 {
		t.Errorf("the target's graveyard %d -> %d, want three discarded cards", before, got)
	}
}

func TestTspB1PlunderDestroysAnArtifactOrLand(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	land := b12Permanent(g, opp.ID, "Their Land", "Land")
	bear := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	if err := castCatalogSpellErr(t, g, "Plunder", "Sorcery", tspB1PlunderOracle, cardRefs(bear)); err == nil {
		t.Error("Plunder targeted a creature")
	}
	castCatalogSpell(t, g, "Plunder", "Sorcery", tspB1PlunderOracle, cardRefs(land))
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(land) {
		t.Error("the land survived")
	}
}

func TestTspB1DivineCongregationCountsTheTargetsCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	for i := 0; i < 3; i++ {
		b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	}
	b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	before := me.Life
	castCatalogSpell(t, g, "Divine Congregation", "Sorcery", tspB1DivineCongregationOracle, playerRefs(opp.ID))
	passPriorityAroundTable(t, g)
	if me.Life != before+6 {
		t.Errorf("life %d -> %d, want +6 for the target's three creatures", before, me.Life)
	}
}

// --- flashback ---------------------------------------------------------------

func TestTspB1AncientGrudgeFlashesBackForGreen(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	rock := b12Permanent(g, opp.ID, "Their Rock", "Artifact")
	id := seedGraveyardCard(t, g, "Ancient Grudge", "Instant", tspB1AncientGrudgeOracle)
	floatMana(t, g, me, "{G}")
	if err := s82MidB11Flashback(g, id, game.CastSpellParams{Targets: cardRefs(rock)}); err != nil {
		t.Fatalf("flashback: %v", err)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(rock) {
		t.Error("the artifact survived")
	}
	if !g.Exile.Contains(id) {
		t.Error("a flashback spell is exiled as it resolves (CR 702.34a)")
	}
}

func TestTspB1StranglingSootTakesOnlyToughnessThreeOrLess(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	small := b12Creature(g, opp.ID, "Small", "Creature — Bear", 3, 3)
	big := b12Creature(g, opp.ID, "Big", "Creature — Bear", 3, 4)
	if err := castCatalogSpellErr(t, g, "Strangling Soot", "Instant", tspB1StranglingSootOracle, cardRefs(big)); err == nil {
		t.Error("a toughness-4 creature was targeted")
	}
	castCatalogSpell(t, g, "Strangling Soot", "Instant", tspB1StranglingSootOracle, cardRefs(small))
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(small) {
		t.Error("the toughness-3 creature survived")
	}
}

func TestTspB1ThrillOfTheHuntFlashesBackForWhite(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	id := seedGraveyardCard(t, g, "Thrill of the Hunt", "Instant", tspB1ThrillOfTheHuntOracle)
	floatMana(t, g, me, "{W}")
	if err := s82MidB11Flashback(g, id, game.CastSpellParams{Targets: cardRefs(bear)}); err != nil {
		t.Fatalf("flashback: %v", err)
	}
	passPriorityAroundTable(t, g)
	if p, tough := effectivePower(t, g, bear), effectiveToughness(t, g, bear); p != 3 || tough != 4 {
		t.Errorf("the Bear is %d/%d, want 3/4", p, tough)
	}
}

// --- madness -----------------------------------------------------------------

func TestTspB1DarkWitheringIsCastForItsMadnessCost(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	id := tspB1Loot(t, g, "Dark Withering", "Instant", tspB1DarkWitheringOracle)
	floatMana(t, g, me, "{B}")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{
		Strict: true, FromZone: "exile", AlternativeCost: game.AltCostKeyMadness, Targets: cardRefs(bear),
	}); err != nil {
		t.Fatalf("the madness cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) {
		t.Error("the nonblack creature survived")
	}
}

func TestTspB1CallToTheNetherworldReturnsOnlyABlackCreatureCard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	black := uuid.New()
	green := uuid.New()
	g.WithWriteLock(func() {
		me.Graveyard.PushTop(game.Card{InstanceID: black, Name: "Black Zombie", TypeLine: "Creature — Zombie",
			ManaCost: "{1}{B}", Colors: []string{"B"}, Owner: me.ID, Controller: me.ID})
		me.Graveyard.PushTop(game.Card{InstanceID: green, Name: "Green Bear", TypeLine: "Creature — Bear",
			ManaCost: "{1}{G}", Colors: []string{"G"}, Owner: me.ID, Controller: me.ID})
	})
	if err := castCatalogSpellErr(t, g, "Call to the Netherworld", "Sorcery", tspB1CallToTheNetherworldOracle, cardRefs(green)); err == nil {
		t.Error("a green creature card was targeted")
	}
	id := tspB1Loot(t, g, "Call to the Netherworld", "Sorcery", tspB1CallToTheNetherworldOracle)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{
		Strict: true, FromZone: "exile", AlternativeCost: game.AltCostKeyMadness, Targets: cardRefs(black),
	}); err != nil {
		t.Fatalf("the free madness cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(black) {
		t.Error("the black creature card was not returned to hand")
	}
}

// --- split second ------------------------------------------------------------

func TestTspB1WipeAwayBouncesUnderSplitSecond(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	rock := b12Permanent(g, opp.ID, "Their Rock", "Artifact")
	spell := castCatalogSpell(t, g, "Wipe Away", "Instant", tspB1WipeAwayOracle, cardRefs(rock))
	requireSplitSecondOn(t, g, spell)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(rock) || !opp.Hand.Contains(rock) {
		t.Error("the target was not returned to its owner's hand")
	}
}

func TestTspB1StonewoodInvocationPumpsAndShroudsUnderSplitSecond(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	spell := castCatalogSpell(t, g, "Stonewood Invocation", "Instant", tspB1StonewoodInvocationOracle, cardRefs(bear))
	requireSplitSecondOn(t, g, spell)
	passPriorityAroundTable(t, g)
	if p, tough := effectivePower(t, g, bear), effectiveToughness(t, g, bear); p != 7 || tough != 7 {
		t.Errorf("the Bear is %d/%d, want 7/7", p, tough)
	}
	if !hasKeywordOf(g, bear, "shroud") {
		t.Error("the Bear did not gain shroud")
	}
}

func TestTspB1WordOfSeizingUntapsAndTakesAPermanent(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() { b16Tap(g, bear) })
	spell := castCatalogSpell(t, g, "Word of Seizing", "Instant", tspB1WordOfSeizingOracle, cardRefs(bear))
	requireSplitSecondOn(t, g, spell)
	passPriorityAroundTable(t, g)
	if controllerOf(t, g, bear) != me.ID {
		t.Error("control of the permanent was not gained")
	}
	if b16Tapped(t, g, bear) {
		t.Error("the permanent was not untapped")
	}
	if !hasKeywordOf(g, bear, "haste") {
		t.Error("the permanent did not gain haste")
	}
}

// --- the rest ----------------------------------------------------------------

func TestTspB1PenumbraSpiderLeavesABlackSpider(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	spider := b12Push(g, me.ID, "Penumbra Spider", "Creature — Spider", tspB1PenumbraSpiderOracle, 2, 4)
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(spider); err != nil {
			t.Fatalf("destroy: %v", err)
		}
	})
	passPriorityAroundTable(t, g)
	var token game.Card
	found := false
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.Name == "Spider" && c.Controller == me.ID && c.IsToken() {
				token, found = c, true
			}
		}
	})
	if !found {
		t.Fatal("no Spider token was created")
	}
	if token.Power != 2 || token.Toughness != 4 || len(token.Colors) != 1 || token.Colors[0] != "B" || !hasKeywordOf(g, token.InstanceID, "reach") {
		t.Errorf("token %+v, want a 2/4 black Spider with reach", token)
	}
}

func TestTspB1AssassinateTakesOnlyATappedCreature(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	tapped := b12Creature(g, opp.ID, "Tapped", "Creature — Bear", 2, 2)
	untapped := b12Creature(g, opp.ID, "Untapped", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() { b16Tap(g, tapped) })
	if err := castCatalogSpellErr(t, g, "Assassinate", "Sorcery", tspB1AssassinateOracle, cardRefs(untapped)); err == nil {
		t.Error("an untapped creature was targeted")
	}
	castCatalogSpell(t, g, "Assassinate", "Sorcery", tspB1AssassinateOracle, cardRefs(tapped))
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(tapped) {
		t.Error("the tapped creature survived")
	}
}

func TestTspB1SproutMakesASaproling(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	castCatalogSpell(t, g, "Sprout", "Instant", tspB1SproutOracle, nil)
	passPriorityAroundTable(t, g)
	if got := countOnBattlefield(g, "Saproling", me.ID); got != 1 {
		t.Errorf("%d Saprolings, want 1", got)
	}
}
