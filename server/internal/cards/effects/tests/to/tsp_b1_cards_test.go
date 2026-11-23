package teststo

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
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

// --- helpers -----------------------------------------------------------------

// tspB1Suspend suspends a card from the active seat's hand with `mana`
// floating, then leaves it one time counter from its free cast.
func tspB1Suspend(t *testing.T, g *game.Game, name, typeLine, oracle, mana string, n int) uuid.UUID {
	t.Helper()
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	id := effectstest.HandCardForTest(me, name, typeLine, oracle)
	for i := range me.Hand.Cards {
		if me.Hand.Cards[i].InstanceID == id {
			me.Hand.Cards[i].Power, me.Hand.Cards[i].Toughness = 4, 4 // read only by a creature
		}
	}
	effectstest.FloatMana(t, g, me, mana)
	if err := g.PerformSpecialAction(me.ID, id, game.SpecialActionSuspend, game.SpecialActionParams{Strict: true}); err != nil {
		t.Fatalf("suspend %s: %v", name, err)
	}
	if got := effectstest.CounterCountInExile(g, id, game.CounterTime); got != n {
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
	effectstest.AdvanceToUpkeepOf(t, g, seat)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestChoiceOfKind(g, game.PendingChoiceResolutionCast) == nil {
		t.Fatal("the last time counter coming off offered no cast")
	}
	params.Strict, params.FromZone = true, "exile"
	if err := g.CastSpell(me.ID, id, params); err != nil {
		t.Fatalf("the free cast from suspend: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
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
			g := effectstest.NewCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			mine := effectstest.TspB1Sliver(g, me.ID, "My Sliver", 1, 1)
			theirs := effectstest.TspB1Sliver(g, opp.ID, "Their Sliver", 1, 1)
			bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
			effectstest.B12Push(g, me.ID, tc.name, "Creature — Sliver", tc.oracle, 2, 2)
			for _, id := range []uuid.UUID{mine, theirs} {
				if p, tough := effectstest.EffectivePower(t, g, id), effectstest.EffectiveToughness(t, g, id); p != 1+tc.p || tough != 1+tc.tough {
					t.Errorf("a Sliver is %d/%d, want %d/%d", p, tough, 1+tc.p, 1+tc.tough)
				}
			}
			if p, tough := effectstest.EffectivePower(t, g, bear), effectstest.EffectiveToughness(t, g, bear); p != 2 || tough != 2 {
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
			g := effectstest.NewCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			theirs := effectstest.TspB1Sliver(g, opp.ID, "Their Sliver", 1, 1)
			bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
			lord := effectstest.B12Push(g, me.ID, tc.name, "Creature — Sliver", tc.oracle, 1, 1)
			if !effectstest.HasKeywordOf(g, theirs, tc.keyword) || !effectstest.HasKeywordOf(g, lord, tc.keyword) {
				t.Errorf("want every Sliver, an opponent's and this one included, to have %s", tc.keyword)
			}
			if effectstest.HasKeywordOf(g, bear, tc.keyword) {
				t.Errorf("a Bear has %s", tc.keyword)
			}
		})
	}
}

// A blocker without flanking shrinks when it blocks a Sliver given
// flanking by Sidewinder Sliver.
func TestTspB1SidewinderSliverFlankingShrinksTheBlocker(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	attacker := effectstest.TspB1Sliver(g, me.ID, "My Sliver", 2, 2)
	effectstest.B12Push(g, opp.ID, "Sidewinder Sliver", "Creature — Sliver", tspB1SidewinderSliverOracle, 1, 1)
	blocker := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 3)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(attacker, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	effectstest.LockInAttacks(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(blocker, attacker); err != nil {
		t.Fatalf("DeclareBlocker: %v", err)
	}
	effectstest.LockInBlocks(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.EffectivePower(t, g, blocker), effectstest.EffectiveToughness(t, g, blocker); p != 1 || tough != 2 {
		t.Errorf("the blocker is %d/%d, want 1/2 after flanking", p, tough)
	}
}

// --- sliver granted abilities ------------------------------------------------

func TestTspB1FirewakeSliverSacrificesASliverToPumpAnother(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	fodder := effectstest.TspB1Sliver(g, me.ID, "Fodder Sliver", 1, 1)
	other := effectstest.TspB1Sliver(g, me.ID, "Other Sliver", 1, 1)
	effectstest.B12Push(g, me.ID, "Firewake Sliver", "Creature — Sliver", tspB1FirewakeSliverOracle, 1, 1)
	effectstest.AdvanceToMain(t, g)
	effectstest.FloatMana(t, g, me, "{C}")
	effectstest.TspB1ActivateGranted(t, g, me.ID, fodder, effectstest.CardRefs(other))
	if g.Battlefield.Contains(fodder) {
		t.Error("the Sliver that activated the ability was not sacrificed")
	}
	if p, tough := effectstest.EffectivePower(t, g, other), effectstest.EffectiveToughness(t, g, other); p != 3 || tough != 3 {
		t.Errorf("the target Sliver is %d/%d, want 3/3", p, tough)
	}
}

func TestTspB1TelekineticSliverTapsATargetPermanent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	sliver := effectstest.TspB1Sliver(g, me.ID, "My Sliver", 1, 1)
	effectstest.B12Push(g, me.ID, "Telekinetic Sliver", "Creature — Sliver", tspB1TelekineticSliverOracle, 2, 2)
	rock := effectstest.B12Permanent(g, opp.ID, "Their Rock", "Artifact")
	effectstest.AdvanceToMain(t, g)
	effectstest.TspB1ActivateGranted(t, g, me.ID, sliver, effectstest.CardRefs(rock))
	if !effectstest.B16Tapped(t, g, rock) || !effectstest.B16Tapped(t, g, sliver) {
		t.Error("want the target and the Sliver both tapped")
	}
}

func TestTspB1ScreechingSliverMillsTheTargetPlayer(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	sliver := effectstest.TspB1Sliver(g, me.ID, "My Sliver", 1, 1)
	effectstest.B12Push(g, me.ID, "Screeching Sliver", "Creature — Sliver", tspB1ScreechingSliverOracle, 1, 1)
	effectstest.AdvanceToMain(t, g)
	before := opp.Graveyard.Size()
	effectstest.TspB1ActivateGranted(t, g, me.ID, sliver, effectstest.PlayerRefs(opp.ID))
	if got := opp.Graveyard.Size(); got != before+1 {
		t.Errorf("opponent's graveyard %d -> %d, want one card milled", before, got)
	}
}

func TestTspB1PsionicSliverHitsTheTargetAndItself(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	big := effectstest.TspB1Sliver(g, me.ID, "Big Sliver", 4, 4)
	small := effectstest.TspB1Sliver(g, me.ID, "Small Sliver", 1, 1)
	effectstest.B12Push(g, me.ID, "Psionic Sliver", "Creature — Sliver", tspB1PsionicSliverOracle, 2, 2)
	effectstest.AdvanceToMain(t, g)
	before := opp.Life
	effectstest.TspB1ActivateGranted(t, g, me.ID, big, effectstest.PlayerRefs(opp.ID))
	if opp.Life != before-2 {
		t.Errorf("opponent life %d -> %d, want 2 damage", before, opp.Life)
	}
	if !g.Battlefield.Contains(big) {
		t.Fatal("a 4/4 Sliver died to 3 damage")
	}
	effectstest.TspB1ActivateGranted(t, g, me.ID, small, effectstest.PlayerRefs(opp.ID))
	if g.Battlefield.Contains(small) {
		t.Error("a 1/1 Sliver survived 3 damage to itself")
	}
}

func TestTspB1QuilledSliverPingsOnlyACreatureInCombat(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	sliver := effectstest.B12Push(g, opp.ID, "Quilled Sliver", "Creature — Sliver", tspB1QuilledSliverOracle, 1, 1)
	attacker := effectstest.B12Creature(g, me.ID, "My Elf", "Creature — Elf", 1, 1)
	idle := effectstest.B12Creature(g, me.ID, "My Idler", "Creature — Elf", 1, 1)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(attacker, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	effectstest.LockInAttacks(t, g)
	idx, ref := effectstest.GrantedActivatedIndex(t, g, sliver)
	if idx < 0 {
		t.Fatal("Quilled Sliver has no granted ability")
	}
	if err := g.ActivateCatalogAbility(opp.ID, sliver, idx, game.ActivateAbilityParams{Ref: ref, Targets: effectstest.CardRefs(idle)}); err == nil {
		t.Error("a creature that is neither attacking nor blocking was targeted")
	}
	effectstest.B16Activate(t, g, opp.ID, sliver, idx, game.ActivateAbilityParams{Ref: ref, Targets: effectstest.CardRefs(attacker)})
	if g.Battlefield.Contains(attacker) {
		t.Error("the attacking 1/1 survived")
	}
}

func TestTspB1MindlashSliverMakesEachPlayerDiscard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	sliver := effectstest.TspB1Sliver(g, me.ID, "My Sliver", 1, 1)
	effectstest.B12Push(g, me.ID, "Mindlash Sliver", "Creature — Sliver", tspB1MindlashSliverOracle, 1, 1)
	effectstest.AdvanceToMain(t, g)
	effectstest.FloatMana(t, g, me, "{C}")
	idx, ref := effectstest.GrantedActivatedIndex(t, g, sliver)
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
	effectstest.PassPriorityAroundTable(t, g)
	// Each player picks one card of their own, in turn order.
	for i := 0; i < 8; i++ {
		answered := false
		for _, p := range g.Seats {
			if c := effectstest.DiscardChoiceFor(g, p.ID); c != nil && p.Hand.Size() > 0 {
				effectstest.AnswerDiscard(t, g, p.ID, p.Hand.Cards[0].InstanceID)
				answered = true
			}
		}
		if !answered {
			break
		}
		effectstest.PassPriorityAroundTable(t, g)
	}
	for _, p := range g.Seats {
		if got := p.Graveyard.Size(); got != before[p.ID]+1 {
			t.Errorf("%s's graveyard %d -> %d, want one discarded card", p.Name, before[p.ID], got)
		}
	}
}

func TestTspB1BasalSliverSacrificesASliverForBlackMana(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := effectstest.TspB1Sliver(g, me.ID, "My Sliver", 1, 1)
	theirs := effectstest.TspB1Sliver(g, opp.ID, "Their Sliver", 1, 1)
	effectstest.B12Push(g, me.ID, "Basal Sliver", "Creature — Sliver", tspB1BasalSliverOracle, 2, 2)
	if effectstest.GrantedManaRowsOf(g, theirs) != 1 {
		t.Error("an opponent's Sliver has no granted mana ability")
	}
	if err := effectstest.TapGrantedMana(t, g, me.ID, mine, ""); err != nil {
		t.Fatalf("sacrifice for mana: %v", err)
	}
	if g.Battlefield.Contains(mine) {
		t.Error("the Sliver was not sacrificed")
	}
	if got := effectstest.Batch01PoolColors(me); len(got) != 2 || got[0] != "B" || got[1] != "B" {
		t.Errorf("pool %v, want [B B]", got)
	}
}

func TestTspB1PlagueSliverHurtsEachSliversControllerInTheirUpkeep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	owner, other := g.Seats[1], g.Seats[2]
	effectstest.TspB1Sliver(g, owner.ID, "Owner's Sliver", 1, 1)
	effectstest.B12Push(g, owner.ID, "Plague Sliver", "Creature — Sliver", tspB1PlagueSliverOracle, 5, 5)
	effectstest.TspB1Sliver(g, other.ID, "Other's Sliver", 1, 1)
	before, otherBefore := owner.Life, other.Life
	effectstest.AdvanceToUpkeepOf(t, g, 1)
	effectstest.PassPriorityAroundTable(t, g)
	if owner.Life != before-2 {
		t.Errorf("owner life %d -> %d, want 2 damage (two Slivers)", before, owner.Life)
	}
	if other.Life != otherBefore {
		t.Error("another player's Sliver hurt them in someone else's upkeep")
	}
}

func TestTspB1HarmonicSliverDestroysOnEachSliverEntering(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	first := effectstest.B12Permanent(g, opp.ID, "Their Relic", "Artifact")
	second := effectstest.B12Permanent(g, opp.ID, "Their Aura", "Enchantment")
	effectstest.CastHoldCreature(t, g, "Harmonic Sliver", "Creature — Sliver", tspB1HarmonicSliverOracle, 1, 1)
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.AnswerPickTarget(t, g, first)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(first) {
		t.Fatal("Harmonic Sliver's own entry destroyed nothing")
	}
	effectstest.CastHoldCreature(t, g, "Muscle Sliver", "Creature — Sliver", "", 1, 1)
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.AnswerPickTarget(t, g, second)
	effectstest.PassPriorityAroundTable(t, g)
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
			effectstest.S40SuspendAction(t, tc.oracle, tc.n, tc.cost)
		})
	}
}

func TestTspB1WheelOfFateIsCastFromSuspendAndRefillsEveryHand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
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
	g := effectstest.NewCatalogGame(t)
	id := tspB1Suspend(t, g, "Errant Ephemeron", "Creature — Illusion", tspB1ErrantEphemeronOracle, "{C}{U}", 4)
	tspB1FreeCastAtUpkeep(t, g, id, game.CastSpellParams{})
	if !g.Battlefield.Contains(id) {
		t.Fatal("the Ephemeron is not on the battlefield")
	}
	if !effectstest.HasKeywordOf(g, id, "haste") || !effectstest.HasKeywordOf(g, id, "flying") {
		t.Error("a creature cast from suspend has haste (CR 702.62a), and this one flies")
	}
}

func TestTspB1RiftwingCloudskateBouncesATargetPermanent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	rock := effectstest.B12Permanent(g, opp.ID, "Their Rock", "Artifact")
	effectstest.CastHoldCreature(t, g, "Riftwing Cloudskate", "Creature — Illusion", tspB1RiftwingCloudskateOracle, 2, 2)
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.AnswerPickTarget(t, g, rock)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(rock) || !opp.Hand.Contains(rock) {
		t.Error("the target was not returned to its owner's hand")
	}
}

func TestTspB1DeepSeaKrakenLosesACounterWhenAnOpponentCasts(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	id := effectstest.HandCardForTest(me, "Deep-Sea Kraken", "Creature — Kraken", tspB1DeepSeaKrakenOracle)
	effectstest.FloatMana(t, g, me, "{C}{C}{U}")
	if err := g.PerformSpecialAction(me.ID, id, game.SpecialActionSuspend, game.SpecialActionParams{Strict: true}); err != nil {
		t.Fatalf("suspend: %v", err)
	}
	// My own spell does not count.
	effectstest.CastCatalogSpell(t, g, "My Instant", "Instant", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CounterCountInExile(g, id, game.CounterTime); got != 9 {
		t.Fatalf("%d time counters after my own spell, want 9", got)
	}
	spell := uuid.New()
	g.WithWriteLock(func() {
		opp.Hand.PushTop(game.Card{InstanceID: spell, Name: "Their Instant", TypeLine: "Instant", Owner: opp.ID, Controller: opp.ID})
	})
	if err := g.CastSpell(opp.ID, spell, game.CastSpellParams{}); err != nil {
		t.Fatalf("opponent casts: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CounterCountInExile(g, id, game.CounterTime); got != 8 {
		t.Errorf("%d time counters after an opponent's spell, want 8", got)
	}
}

func TestTspB1MindstabMakesTheTargetDiscardThree(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	effectstest.CastCatalogSpell(t, g, "Mindstab", "Sorcery", tspB1MindstabOracle, effectstest.PlayerRefs(opp.ID))
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Hand.Size() < 3 {
		t.Fatalf("the opponent holds %d cards", opp.Hand.Size())
	}
	before := opp.Graveyard.Size()
	effectstest.AnswerDiscard(t, g, opp.ID, opp.Hand.Cards[0].InstanceID, opp.Hand.Cards[1].InstanceID, opp.Hand.Cards[2].InstanceID)
	effectstest.PassPriorityAroundTable(t, g)
	if got := opp.Graveyard.Size(); got != before+3 {
		t.Errorf("the target's graveyard %d -> %d, want three discarded cards", before, got)
	}
}

func TestTspB1PlunderDestroysAnArtifactOrLand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	land := effectstest.B12Permanent(g, opp.ID, "Their Land", "Land")
	bear := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	if err := effectstest.CastCatalogSpellErr(t, g, "Plunder", "Sorcery", tspB1PlunderOracle, effectstest.CardRefs(bear)); err == nil {
		t.Error("Plunder targeted a creature")
	}
	effectstest.CastCatalogSpell(t, g, "Plunder", "Sorcery", tspB1PlunderOracle, effectstest.CardRefs(land))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(land) {
		t.Error("the land survived")
	}
}

func TestTspB1DivineCongregationCountsTheTargetsCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	for i := 0; i < 3; i++ {
		effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	}
	effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	before := me.Life
	effectstest.CastCatalogSpell(t, g, "Divine Congregation", "Sorcery", tspB1DivineCongregationOracle, effectstest.PlayerRefs(opp.ID))
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != before+6 {
		t.Errorf("life %d -> %d, want +6 for the target's three creatures", before, me.Life)
	}
}

// --- flashback ---------------------------------------------------------------

func TestTspB1AncientGrudgeFlashesBackForGreen(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	rock := effectstest.B12Permanent(g, opp.ID, "Their Rock", "Artifact")
	id := effectstest.SeedGraveyardCard(t, g, "Ancient Grudge", "Instant", tspB1AncientGrudgeOracle)
	effectstest.FloatMana(t, g, me, "{G}")
	if err := effectstest.S82MidB11Flashback(g, id, game.CastSpellParams{Targets: effectstest.CardRefs(rock)}); err != nil {
		t.Fatalf("flashback: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(rock) {
		t.Error("the artifact survived")
	}
	if !g.Exile.Contains(id) {
		t.Error("a flashback spell is exiled as it resolves (CR 702.34a)")
	}
}

func TestTspB1StranglingSootTakesOnlyToughnessThreeOrLess(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	small := effectstest.B12Creature(g, opp.ID, "Small", "Creature — Bear", 3, 3)
	big := effectstest.B12Creature(g, opp.ID, "Big", "Creature — Bear", 3, 4)
	if err := effectstest.CastCatalogSpellErr(t, g, "Strangling Soot", "Instant", tspB1StranglingSootOracle, effectstest.CardRefs(big)); err == nil {
		t.Error("a toughness-4 creature was targeted")
	}
	effectstest.CastCatalogSpell(t, g, "Strangling Soot", "Instant", tspB1StranglingSootOracle, effectstest.CardRefs(small))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(small) {
		t.Error("the toughness-3 creature survived")
	}
}

func TestTspB1ThrillOfTheHuntFlashesBackForWhite(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	id := effectstest.SeedGraveyardCard(t, g, "Thrill of the Hunt", "Instant", tspB1ThrillOfTheHuntOracle)
	effectstest.FloatMana(t, g, me, "{W}")
	if err := effectstest.S82MidB11Flashback(g, id, game.CastSpellParams{Targets: effectstest.CardRefs(bear)}); err != nil {
		t.Fatalf("flashback: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.EffectivePower(t, g, bear), effectstest.EffectiveToughness(t, g, bear); p != 3 || tough != 4 {
		t.Errorf("the Bear is %d/%d, want 3/4", p, tough)
	}
}

// --- madness -----------------------------------------------------------------

func TestTspB1DarkWitheringIsCastForItsMadnessCost(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	id := effectstest.TspB1Loot(t, g, "Dark Withering", "Instant", tspB1DarkWitheringOracle)
	effectstest.FloatMana(t, g, me, "{B}")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{
		Strict: true, FromZone: "exile", AlternativeCost: game.AltCostKeyMadness, Targets: effectstest.CardRefs(bear),
	}); err != nil {
		t.Fatalf("the madness cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) {
		t.Error("the nonblack creature survived")
	}
}

func TestTspB1CallToTheNetherworldReturnsOnlyABlackCreatureCard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	black := uuid.New()
	green := uuid.New()
	g.WithWriteLock(func() {
		me.Graveyard.PushTop(game.Card{InstanceID: black, Name: "Black Zombie", TypeLine: "Creature — Zombie",
			ManaCost: "{1}{B}", Colors: []string{"B"}, Owner: me.ID, Controller: me.ID})
		me.Graveyard.PushTop(game.Card{InstanceID: green, Name: "Green Bear", TypeLine: "Creature — Bear",
			ManaCost: "{1}{G}", Colors: []string{"G"}, Owner: me.ID, Controller: me.ID})
	})
	if err := effectstest.CastCatalogSpellErr(t, g, "Call to the Netherworld", "Sorcery", tspB1CallToTheNetherworldOracle, effectstest.CardRefs(green)); err == nil {
		t.Error("a green creature card was targeted")
	}
	id := effectstest.TspB1Loot(t, g, "Call to the Netherworld", "Sorcery", tspB1CallToTheNetherworldOracle)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{
		Strict: true, FromZone: "exile", AlternativeCost: game.AltCostKeyMadness, Targets: effectstest.CardRefs(black),
	}); err != nil {
		t.Fatalf("the free madness cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(black) {
		t.Error("the black creature card was not returned to hand")
	}
}

// --- split second ------------------------------------------------------------

func TestTspB1WipeAwayBouncesUnderSplitSecond(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	rock := effectstest.B12Permanent(g, opp.ID, "Their Rock", "Artifact")
	spell := effectstest.CastCatalogSpell(t, g, "Wipe Away", "Instant", tspB1WipeAwayOracle, effectstest.CardRefs(rock))
	effectstest.RequireSplitSecondOn(t, g, spell)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(rock) || !opp.Hand.Contains(rock) {
		t.Error("the target was not returned to its owner's hand")
	}
}

func TestTspB1StonewoodInvocationPumpsAndShroudsUnderSplitSecond(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	spell := effectstest.CastCatalogSpell(t, g, "Stonewood Invocation", "Instant", tspB1StonewoodInvocationOracle, effectstest.CardRefs(bear))
	effectstest.RequireSplitSecondOn(t, g, spell)
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.EffectivePower(t, g, bear), effectstest.EffectiveToughness(t, g, bear); p != 7 || tough != 7 {
		t.Errorf("the Bear is %d/%d, want 7/7", p, tough)
	}
	if !effectstest.HasKeywordOf(g, bear, "shroud") {
		t.Error("the Bear did not gain shroud")
	}
}

func TestTspB1WordOfSeizingUntapsAndTakesAPermanent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() { effectstest.B16Tap(g, bear) })
	spell := effectstest.CastCatalogSpell(t, g, "Word of Seizing", "Instant", tspB1WordOfSeizingOracle, effectstest.CardRefs(bear))
	effectstest.RequireSplitSecondOn(t, g, spell)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.ControllerOf(t, g, bear) != me.ID {
		t.Error("control of the permanent was not gained")
	}
	if effectstest.B16Tapped(t, g, bear) {
		t.Error("the permanent was not untapped")
	}
	if !effectstest.HasKeywordOf(g, bear, "haste") {
		t.Error("the permanent did not gain haste")
	}
}

// --- the rest ----------------------------------------------------------------

func TestTspB1PenumbraSpiderLeavesABlackSpider(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	spider := effectstest.B12Push(g, me.ID, "Penumbra Spider", "Creature — Spider", tspB1PenumbraSpiderOracle, 2, 4)
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(spider); err != nil {
			t.Fatalf("destroy: %v", err)
		}
	})
	effectstest.PassPriorityAroundTable(t, g)
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
	if token.Power != 2 || token.Toughness != 4 || len(token.Colors) != 1 || token.Colors[0] != "B" || !effectstest.HasKeywordOf(g, token.InstanceID, "reach") {
		t.Errorf("token %+v, want a 2/4 black Spider with reach", token)
	}
}

func TestTspB1AssassinateTakesOnlyATappedCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	tapped := effectstest.B12Creature(g, opp.ID, "Tapped", "Creature — Bear", 2, 2)
	untapped := effectstest.B12Creature(g, opp.ID, "Untapped", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() { effectstest.B16Tap(g, tapped) })
	if err := effectstest.CastCatalogSpellErr(t, g, "Assassinate", "Sorcery", tspB1AssassinateOracle, effectstest.CardRefs(untapped)); err == nil {
		t.Error("an untapped creature was targeted")
	}
	effectstest.CastCatalogSpell(t, g, "Assassinate", "Sorcery", tspB1AssassinateOracle, effectstest.CardRefs(tapped))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(tapped) {
		t.Error("the tapped creature survived")
	}
}

func TestTspB1SproutMakesASaproling(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.CastCatalogSpell(t, g, "Sprout", "Instant", tspB1SproutOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountOnBattlefield(g, "Saproling", me.ID); got != 1 {
		t.Errorf("%d Saprolings, want 1", got)
	}
}
