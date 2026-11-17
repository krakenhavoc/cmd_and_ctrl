package effects

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// scg_b1_cards_test.go — Scourge, batch 1 (#5608, S82): thirty-nine cards
// built from seams the engine already has. Each test plays the real
// catalog entry.

const (
	scgB1AcceleratedMutationOracle = "f8a9c279-0f11-4a22-8651-f9caf013ca3c"
	scgB1AstralSteelOracle         = "ef8ac291-3789-4147-a11f-c62fd8566dbe"
	scgB1BladewingOracle           = "c22cc70a-9533-4bfc-b3ce-b2d43a73f0c1"
	scgB1BreakAsunderOracle        = "d2c53737-c265-46e3-a779-52c6b4f82d7d"
	scgB1CabalConditioningOracle   = "37a04d5a-9272-42bd-b448-c658375bbbc3"
	scgB1ChartoothCougarOracle     = "717152c3-2082-49aa-9274-a1a6abe54b65"
	scgB1DaruWarchiefOracle        = "3fce53fb-6914-4f43-b46b-18ca32ab21b2"
	scgB1DispersalShieldOracle     = "aefb1463-76fb-419e-8cda-8c5bfb88a3d2"
	scgB1DragonstormOracle         = "d74c3871-018a-4450-a613-f67703bb0d72"
	scgB1ElvishAberrationOracle    = "1d25b831-c1de-4f79-bbb3-1582eda96452"
	scgB1EnrageOracle              = "0f6e66d5-4f27-485b-999d-ee55c1e218b9"
	scgB1GoblinWarStrikeOracle     = "5944d5c7-65ab-425f-a453-ed5871ca3d73"
	scgB1GoblinWarchiefOracle      = "39882df0-c20f-469d-94ba-1617224e71a1"
	scgB1HinderingTouchOracle      = "90c6f338-fc41-4499-8f39-1cd202a2cd31"
	scgB1HuntingPackOracle         = "d3e1848a-c236-4352-a7be-b0b4699af968"
	scgB1KrosanDroverOracle        = "cd8ae400-4f40-42aa-a905-2787de443471"
	scgB1KrosanWarchiefOracle      = "88350861-244c-4634-9bc9-e9a79401bd02"
	scgB1MindsDesireOracle         = "b3739b5a-5731-4c64-a244-815a363b0d5c"
	scgB1MisguidedRageOracle       = "f5c77355-96dc-46af-a7de-f547a1368ff1"
	scgB1NobleTemplarOracle        = "e7079969-8d88-4f18-a872-2661bed6ff74"
	scgB1OneWithNatureOracle       = "e48d3de9-a5e0-4a69-8377-9e3ad5fc5b77"
	scgB1PemminsAuraOracle         = "ada2f380-bca7-4d64-9587-bab11c776ba8"
	scgB1PyrostaticPillarOracle    = "3b5b87da-5a00-449a-97ef-fa1e2c1c4b4b"
	scgB1ReapingTheGravesOracle    = "80c663a5-2794-4748-b47a-34352fb2112f"
	scgB1RewardTheFaithfulOracle   = "7fb7c848-377c-4476-9b3f-d95d145137f1"
	scgB1RushOfKnowledgeOracle     = "34282c99-68ef-4855-b683-7bb205fa6374"
	scgB1ScattershotOracle         = "ed3652da-0127-4c9e-9913-f20becdd0ba9"
	scgB1ShorelineRangerOracle     = "606fbf99-d3e3-4131-b991-3cb71c003834"
	scgB1SkulltapOracle            = "17c7878f-891f-40a6-80a7-47fcc7b8e29b"
	scgB1SliverOverlordOracle      = "762d864b-69bc-43ac-af2f-e18186010169"
	scgB1SparkSprayOracle          = "c37f1921-7ac3-4185-8579-1dfdfe647ea2"
	scgB1SproutingVinesOracle      = "a38ad1aa-fc12-49be-b9f7-fe287597a5ca"
	scgB1TemporalFissureOracle     = "8a9525f4-74a6-44f0-a1b1-d57937d7783b"
	scgB1TorrentOfFireOracle       = "00181784-9213-492a-ba8c-2028969b049e"
	scgB1UnburdenOracle            = "97acd3aa-ae4a-4e9d-a5b9-36eaba158735"
	scgB1UnspeakableSymbolOracle   = "a5690d5f-633c-4a1e-afba-5fd79dcbf20e"
	scgB1WingShardsOracle          = "b302bce5-d7ad-46d8-9dbf-8159e8709b4f"
	scgB1WipeCleanOracle           = "f5cbc4f3-4e73-405f-9daf-e5b7c73fa6ee"
	scgB1WirewoodGuardianOracle    = "61ade140-5b11-4882-82e5-a075fc80b260"
)

func TestScgB1CardsAreRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		scgB1AcceleratedMutationOracle: "Accelerated Mutation",
		scgB1AstralSteelOracle:         "Astral Steel",
		scgB1BladewingOracle:           "Bladewing the Risen",
		scgB1BreakAsunderOracle:        "Break Asunder",
		scgB1CabalConditioningOracle:   "Cabal Conditioning",
		scgB1ChartoothCougarOracle:     "Chartooth Cougar",
		scgB1DaruWarchiefOracle:        "Daru Warchief",
		scgB1DispersalShieldOracle:     "Dispersal Shield",
		scgB1DragonstormOracle:         "Dragonstorm",
		scgB1ElvishAberrationOracle:    "Elvish Aberration",
		scgB1EnrageOracle:              "Enrage",
		scgB1GoblinWarStrikeOracle:     "Goblin War Strike",
		scgB1GoblinWarchiefOracle:      "Goblin Warchief",
		scgB1HinderingTouchOracle:      "Hindering Touch",
		scgB1HuntingPackOracle:         "Hunting Pack",
		scgB1KrosanDroverOracle:        "Krosan Drover",
		scgB1KrosanWarchiefOracle:      "Krosan Warchief",
		scgB1MindsDesireOracle:         "Mind's Desire",
		scgB1MisguidedRageOracle:       "Misguided Rage",
		scgB1NobleTemplarOracle:        "Noble Templar",
		scgB1OneWithNatureOracle:       "One with Nature",
		scgB1PemminsAuraOracle:         "Pemmin's Aura",
		scgB1PyrostaticPillarOracle:    "Pyrostatic Pillar",
		scgB1ReapingTheGravesOracle:    "Reaping the Graves",
		scgB1RewardTheFaithfulOracle:   "Reward the Faithful",
		scgB1RushOfKnowledgeOracle:     "Rush of Knowledge",
		scgB1ScattershotOracle:         "Scattershot",
		scgB1ShorelineRangerOracle:     "Shoreline Ranger",
		scgB1SkulltapOracle:            "Skulltap",
		scgB1SliverOverlordOracle:      "Sliver Overlord",
		scgB1SparkSprayOracle:          "Spark Spray",
		scgB1SproutingVinesOracle:      "Sprouting Vines",
		scgB1TemporalFissureOracle:     "Temporal Fissure",
		scgB1TorrentOfFireOracle:       "Torrent of Fire",
		scgB1UnburdenOracle:            "Unburden",
		scgB1UnspeakableSymbolOracle:   "Unspeakable Symbol",
		scgB1WingShardsOracle:          "Wing Shards",
		scgB1WipeCleanOracle:           "Wipe Clean",
		scgB1WirewoodGuardianOracle:    "Wirewood Guardian",
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

// scgB1Settle passes priority until the stack is empty or a choice other
// than a re-target offer is waiting. Each storm copy's CR 707.10c offer is
// answered with the next of `retargets`, in order, and the number of
// offers answered is returned: one offer per copy is the per-copy shape.
func scgB1Settle(t *testing.T, g *game.Game, chooser uuid.UUID, retargets ...game.TargetRef) int {
	t.Helper()
	seen := 0
	for i := 0; i < 64; i++ {
		if p := latestPickTarget(g, chooser); p != nil {
			if seen >= len(retargets) {
				t.Fatalf("re-target offer %d, but only %d answers were given", seen+1, len(retargets))
			}
			if err := g.ResolvePickTarget(p.ID, chooser, retargets[seen]); err != nil {
				t.Fatalf("ResolvePickTarget: %v", err)
			}
			seen++
			continue
		}
		if stackFullyEmpty(g) {
			return seen
		}
		if err := g.PassPriority(); err != nil {
			if errors.Is(err, game.ErrChoicePending) {
				return seen
			}
			t.Fatalf("PassPriority: %v", err)
		}
	}
	t.Fatal("the stack never settled")
	return seen
}

// scgB1GraveyardCreature puts a creature card into `p`'s graveyard.
func scgB1GraveyardCreature(p *game.Player, name, typeLine string) uuid.UUID {
	id := uuid.New()
	p.Graveyard.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: typeLine, Power: 2, Toughness: 2,
		Owner: p.ID, Controller: p.ID,
	})
	return id
}

// scgB1ExiledNamed reports whether a card with this name is in exile.
func scgB1ExiledNamed(g *game.Game, name string) bool {
	for _, c := range g.Exile.Cards {
		if c.Name == name {
			return true
		}
	}
	return false
}

// scgB1Typecycle typecycles a card from the active seat's hand with the
// ability at `idx`, paying {2}, and fetches `land` from a library that also
// holds a land of another type.
func scgB1Typecycle(t *testing.T, name, typeLine, oracle string, idx int, land, landType, other, otherType string) {
	t.Helper()
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	seedSearchLibrary(me,
		game.Card{Name: other, TypeLine: otherType},
		game.Card{Name: land, TypeLine: landType},
	)
	id := pushCatalogHandCard(me, name, typeLine, oracle)
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{C}{C}"); err != nil {
		t.Fatal(err)
	}
	if err := g.ActivateCatalogAbility(me.ID, id, idx, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("%s: typecycle: %v", name, err)
	}
	if !me.Graveyard.Contains(id) {
		t.Fatalf("%s was not discarded to typecycle", name)
	}
	passPriorityAroundTable(t, g)
	c := searchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatalf("%s: no search prompt", name)
	}
	if searchOptionNamed(g, c, other) != uuid.Nil {
		t.Errorf("%s offered %s", name, other)
	}
	answerSearchNamed(t, g, me.ID, land)
	if !b02bHandHasNamed(me, land) {
		t.Errorf("%s did not put %s into the hand", name, land)
	}
}

// --- storm -------------------------------------------------------------------

func TestScgB1AstralSteelCopyPumpsTheSameCreatureAgain(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	elf := b12Creature(g, me.ID, "My Elf", "Creature — Elf", 1, 1)
	stormFiller(t, g)
	castCatalogSpell(t, g, "Astral Steel", "Instant", scgB1AstralSteelOracle, cardRefs(bear))
	if n := scgB1Settle(t, g, me.ID, cardRef(bear)); n != 1 {
		t.Errorf("%d re-target offers, want 1 (one copy)", n)
	}
	if p, tough := effectivePower(t, g, bear), effectiveToughness(t, g, bear); p != 4 || tough != 6 {
		t.Errorf("the Bear is %d/%d, want 4/6 (+1/+2 twice)", p, tough)
	}
	if p, tough := effectivePower(t, g, elf), effectiveToughness(t, g, elf); p != 1 || tough != 1 {
		t.Errorf("the Elf is %d/%d, want 1/1 (no copy was aimed at it)", p, tough)
	}
}

func TestScgB1HinderingTouchTaxesEachCopySeparately(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	spell := castCatalogSpell(t, g, "A Bear", "Creature — Bear", "", nil)
	castCatalogSpell(t, g, "Hindering Touch", "Instant", scgB1HinderingTouchOracle, cardTarget(spell))
	if n := scgB1Settle(t, g, me.ID, cardRef(spell)); n != 1 {
		t.Fatalf("%d re-target offers, want 1 (the Bear spell was cast first)", n)
	}
	// The copy taxes first; paying it does not pay the original.
	fillPool(me, 2)
	if c := answerPayUnless(t, g, me.ID, true); c == nil {
		t.Fatal("the copy asked for nothing")
	}
	if !g.Stack.Contains(spell) {
		t.Fatal("the spell was countered though the copy's {2} was paid")
	}
	scgB1Settle(t, g, me.ID)
	if c := answerPayUnless(t, g, me.ID, false); c == nil {
		t.Fatal("the original asked for nothing")
	}
	scgB1Settle(t, g, me.ID)
	if g.Stack.Contains(spell) || g.Battlefield.Contains(spell) {
		t.Error("the original did not counter the spell when its {2} went unpaid")
	}
}

func TestScgB1HuntingPackMakesABeastPerResolution(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	stormFiller(t, g)
	stormFiller(t, g)
	castCatalogSpell(t, g, "Hunting Pack", "Instant", scgB1HuntingPackOracle, nil)
	if n := scgB1Settle(t, g, me.ID); n != 0 {
		t.Errorf("%d re-target offers for a spell with no target", n)
	}
	if got := countOnBattlefield(g, "Beast", me.ID); got != 3 {
		t.Errorf("%d Beasts, want 3 (the spell and two copies)", got)
	}
}

func TestScgB1ReapingTheGravesCopyReturnsASecondCreature(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	a := scgB1GraveyardCreature(me, "Old Bear", "Creature — Bear")
	b := scgB1GraveyardCreature(me, "Old Wolf", "Creature — Wolf")
	stormFiller(t, g)
	castCatalogSpell(t, g, "Reaping the Graves", "Instant", scgB1ReapingTheGravesOracle, cardRefs(a))
	scgB1Settle(t, g, me.ID, cardRef(b))
	if !b02bHandHasNamed(me, "Old Bear") || !b02bHandHasNamed(me, "Old Wolf") {
		t.Error("want both creature cards returned, one by the spell and one by its copy")
	}
}

func TestScgB1ScattershotCopiesHitDifferentCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], ons2Opponent(g)
	a := b12Creature(g, opp.ID, "Their Elf", "Creature — Elf", 1, 1)
	b := b12Creature(g, opp.ID, "Their Goblin", "Creature — Goblin", 1, 1)
	stormFiller(t, g)
	castCatalogSpell(t, g, "Scattershot", "Instant", scgB1ScattershotOracle, cardRefs(a))
	scgB1Settle(t, g, me.ID, cardRef(b))
	if g.Battlefield.Contains(a) || g.Battlefield.Contains(b) {
		t.Error("want both 1/1s dead, one to the spell and one to its copy")
	}
}

func TestScgB1SproutingVinesSearchesOncePerResolution(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	seedSearchLibrary(me,
		game.Card{Name: "Grizzly", TypeLine: "Creature — Bear"},
		game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"},
		game.Card{Name: "Island", TypeLine: "Basic Land — Island"},
	)
	stormFiller(t, g)
	castCatalogSpell(t, g, "Sprouting Vines", "Instant", scgB1SproutingVinesOracle, nil)
	scgB1Settle(t, g, me.ID)
	c := searchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("the copy did not search")
	}
	if searchOptionNamed(g, c, "Grizzly") != uuid.Nil {
		t.Error("a nonland card was offered")
	}
	answerSearchNamed(t, g, me.ID, "Forest")
	scgB1Settle(t, g, me.ID)
	answerSearchNamed(t, g, me.ID, "Island")
	scgB1Settle(t, g, me.ID)
	if !b02bHandHasNamed(me, "Forest") || !b02bHandHasNamed(me, "Island") {
		t.Error("want two basic lands in hand, one per resolution")
	}
}

func TestScgB1TemporalFissureCopyBouncesASecondPermanent(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], ons2Opponent(g)
	a := b12Creature(g, opp.ID, "Their Elf", "Creature — Elf", 1, 1)
	b := b12Permanent(g, opp.ID, "Their Relic", "Artifact")
	stormFiller(t, g)
	castCatalogSpell(t, g, "Temporal Fissure", "Sorcery", scgB1TemporalFissureOracle, cardRefs(a))
	scgB1Settle(t, g, me.ID, cardRef(b))
	if g.Battlefield.Contains(a) || g.Battlefield.Contains(b) {
		t.Error("want both permanents bounced")
	}
	if !b02bHandHasNamed(opp, "Their Elf") || !b02bHandHasNamed(opp, "Their Relic") {
		t.Error("the permanents did not reach their owner's hand")
	}
}

func TestScgB1WingShardsTakesOnlyAnAttacker(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], ons2Opponent(g)
	attacker := b12Creature(g, opp.ID, "Their Attacker", "Creature — Bear", 2, 2)
	idle := b12Creature(g, opp.ID, "Their Idler", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == attacker {
				g.Battlefield.Cards[i].AttackingTarget = me.ID
			}
		}
	})
	castCatalogSpell(t, g, "Wing Shards", "Instant", scgB1WingShardsOracle, playerRefs(opp.ID))
	scgB1Settle(t, g, me.ID)
	if c := sacrificeChoiceFor(g, opp.ID); c != nil {
		if edictOffers(sacrificeOptions(c), idle) {
			t.Error("a creature that is not attacking was offered")
		}
		answerSacrifice(t, g, opp.ID, attacker)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(attacker) || !g.Battlefield.Contains(idle) {
		t.Error("want the attacker sacrificed and the idle creature kept")
	}
}

func TestScgB1DragonstormPutsADragonPerResolution(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	seedSearchLibrary(me,
		game.Card{Name: "Grizzly", TypeLine: "Creature — Bear", Power: 2, Toughness: 2},
		game.Card{Name: "Dragon Fire", TypeLine: "Kindred Instant — Dragon"},
		game.Card{Name: "Shivan", TypeLine: "Creature — Dragon", Power: 5, Toughness: 5},
		game.Card{Name: "Hoard Idol", TypeLine: "Kindred Artifact — Dragon"},
	)
	stormFiller(t, g)
	castCatalogSpell(t, g, "Dragonstorm", "Sorcery", scgB1DragonstormOracle, nil)
	scgB1Settle(t, g, me.ID)
	c := searchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("the copy did not search")
	}
	if searchOptionNamed(g, c, "Dragon Fire") != uuid.Nil || searchOptionNamed(g, c, "Grizzly") != uuid.Nil {
		t.Error("a Dragon instant or a non-Dragon was offered")
	}
	answerSearchNamed(t, g, me.ID, "Shivan")
	scgB1Settle(t, g, me.ID)
	answerSearchNamed(t, g, me.ID, "Hoard Idol")
	scgB1Settle(t, g, me.ID)
	if countOnBattlefield(g, "Shivan", me.ID) != 1 || countOnBattlefield(g, "Hoard Idol", me.ID) != 1 {
		t.Error("want both Dragon permanent cards on the battlefield")
	}
}

func TestScgB1MindsDesireExilesAFreeCardPerResolution(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	before := map[uuid.UUID]bool{}
	for _, c := range g.Exile.Cards {
		before[c.InstanceID] = true
	}
	stormFiller(t, g)
	castCatalogSpell(t, g, "Mind's Desire", "Sorcery", scgB1MindsDesireOracle, nil)
	scgB1Settle(t, g, me.ID)
	free := 0
	for _, c := range g.Exile.Cards {
		if before[c.InstanceID] {
			continue
		}
		if perm := g.CastPermissionOnCardByIDForEffect(c.InstanceID); perm != nil && perm.Player == me.ID && perm.Cost == "{0}" {
			free++
		}
	}
	if free != 2 {
		t.Errorf("%d exiled cards may be played for free, want 2 (the spell and one copy)", free)
	}
}

// --- cycling spells ----------------------------------------------------------

func TestScgB1CyclingSpellsCycle(t *testing.T) {
	for _, tc := range []struct{ name, typeLine, oracle, mana string }{
		{"Spark Spray", "Instant", scgB1SparkSprayOracle, "{R}"},
		{"Wipe Clean", "Instant", scgB1WipeCleanOracle, "{C}{C}{C}"},
		{"Break Asunder", "Sorcery", scgB1BreakAsunderOracle, "{C}{C}"},
		{"Unburden", "Sorcery", scgB1UnburdenOracle, "{C}{C}"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			seedLibrary(g.Seats[g.Turn.ActiveSeat], "A", "B")
			ons1Cycle(t, g, tc.name, tc.typeLine, tc.oracle, tc.mana)
		})
	}
}

func TestScgB1SparkSprayDealsOne(t *testing.T) {
	g := newCatalogGame(t)
	opp := ons2Opponent(g)
	life := opp.Life
	castCatalogSpell(t, g, "Spark Spray", "Instant", scgB1SparkSprayOracle, playerRefs(opp.ID))
	passPriorityAroundTable(t, g)
	if opp.Life != life-1 {
		t.Errorf("life %d -> %d, want -1", life, opp.Life)
	}
}

func TestScgB1WipeCleanExilesAnEnchantment(t *testing.T) {
	g := newCatalogGame(t)
	opp := ons2Opponent(g)
	ench := b12Permanent(g, opp.ID, "Their Shrine", "Enchantment")
	castCatalogSpell(t, g, "Wipe Clean", "Instant", scgB1WipeCleanOracle, cardRefs(ench))
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(ench) || !scgB1ExiledNamed(g, "Their Shrine") {
		t.Error("the enchantment was not exiled")
	}
}

func TestScgB1BreakAsunderDestroysAnArtifact(t *testing.T) {
	g := newCatalogGame(t)
	opp := ons2Opponent(g)
	relic := b12Permanent(g, opp.ID, "Their Relic", "Artifact")
	castCatalogSpell(t, g, "Break Asunder", "Sorcery", scgB1BreakAsunderOracle, cardRefs(relic))
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(relic) {
		t.Error("the artifact was not destroyed")
	}
}

func TestScgB1UnburdenDiscardsTwoOfTheirChoice(t *testing.T) {
	g := newCatalogGame(t)
	opp := ons2Opponent(g)
	a := pushToHand(g, opp.ID, "Card A", "Sorcery")
	b := pushToHand(g, opp.ID, "Card B", "Sorcery")
	castCatalogSpell(t, g, "Unburden", "Sorcery", scgB1UnburdenOracle, playerRefs(opp.ID))
	passPriorityAroundTable(t, g)
	answerDiscard(t, g, opp.ID, a, b)
	if !opp.Graveyard.Contains(a) || !opp.Graveyard.Contains(b) {
		t.Error("the two chosen cards were not discarded")
	}
}

// --- typecycling creatures ---------------------------------------------------

func TestScgB1TypecyclingFetchesTheNamedLandType(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, oracle string
		idx                    int
		land, landType         string
	}{
		{"Noble Templar", "Creature — Human Cleric Soldier", scgB1NobleTemplarOracle, 0, "Plains", "Basic Land — Plains"},
		{"Shoreline Ranger", "Creature — Bird Soldier Ranger", scgB1ShorelineRangerOracle, 0, "Island", "Basic Land — Island"},
		{"Wirewood Guardian", "Creature — Elf Mutant", scgB1WirewoodGuardianOracle, 0, "Forest", "Basic Land — Forest"},
		{"Elvish Aberration", "Creature — Elf Mutant", scgB1ElvishAberrationOracle, 0, "Forest", "Basic Land — Forest"},
		{"Chartooth Cougar", "Creature — Cat Beast", scgB1ChartoothCougarOracle, 1, "Mountain", "Basic Land — Mountain"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			scgB1Typecycle(t, tc.name, tc.typeLine, tc.oracle, tc.idx, tc.land, tc.landType, "Swamp", "Basic Land — Swamp")
		})
	}
}

func TestScgB1TypecyclersHaveTheirKeywords(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	templar := b12Push(g, me.ID, "Noble Templar", "Creature — Human Cleric Soldier", scgB1NobleTemplarOracle, 3, 6)
	ranger := b12Push(g, me.ID, "Shoreline Ranger", "Creature — Bird Soldier Ranger", scgB1ShorelineRangerOracle, 3, 4)
	if !hasEffectiveKeyword(t, g, templar, "vigilance") {
		t.Error("Noble Templar lacks vigilance")
	}
	if !hasEffectiveKeyword(t, g, ranger, "flying") {
		t.Error("Shoreline Ranger lacks flying")
	}
}

func TestScgB1ChartoothCougarFirebreathes(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	cougar := b12Push(g, me.ID, "Chartooth Cougar", "Creature — Cat Beast", scgB1ChartoothCougarOracle, 4, 4)
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{R}"); err != nil {
		t.Fatal(err)
	}
	b16Activate(t, g, me.ID, cougar, 0, game.ActivateAbilityParams{})
	passPriorityAroundTable(t, g)
	if p := effectivePower(t, g, cougar); p != 5 {
		t.Errorf("power %d, want 5", p)
	}
}

func TestScgB1ElvishAberrationTapsForThreeGreen(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	elf := b12Push(g, me.ID, "Elvish Aberration", "Creature — Elf Mutant", scgB1ElvishAberrationOracle, 4, 5)
	b28TapForMana(t, g, me.ID, elf, "")
	if got := poolColors(me); len(got) != 3 || got[0] != "G" || got[1] != "G" || got[2] != "G" {
		t.Errorf("{T}: Add {G}{G}{G}: pool %v", got)
	}
}

// --- the greatest mana value among permanents you control --------------------

func TestScgB1RushOfKnowledgeDrawsTheGreatestManaValue(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], ons2Opponent(g)
	b27Push(g, me.ID, "Big", "Creature — Beast", "", "{3}{G}{G}", 5, 5, "G")
	b27Push(g, me.ID, "Small", "Artifact", "", "{1}", 0, 0)
	b27Push(g, opp.ID, "Huge", "Artifact", "", "{9}", 0, 0)
	castCatalogSpell(t, g, "Rush of Knowledge", "Sorcery", scgB1RushOfKnowledgeOracle, nil)
	before := len(me.Hand.Cards)
	passPriorityAroundTable(t, g)
	if got := len(me.Hand.Cards); got != before+5 {
		t.Errorf("hand %d -> %d, want five cards (an opponent's permanent does not count)", before, got)
	}
}

func TestScgB1TorrentOfFireDealsTheGreatestManaValue(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], ons2Opponent(g)
	b27Push(g, me.ID, "Four", "Enchantment", "", "{2}{R}{R}", 0, 0)
	life := opp.Life
	castCatalogSpell(t, g, "Torrent of Fire", "Sorcery", scgB1TorrentOfFireOracle, playerRefs(opp.ID))
	passPriorityAroundTable(t, g)
	if opp.Life != life-4 {
		t.Errorf("life %d -> %d, want -4", life, opp.Life)
	}
}

func TestScgB1AcceleratedMutationPumpsByTheGreatestManaValue(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	b27Push(g, me.ID, "Three", "Enchantment", "", "{1}{G}{G}", 0, 0)
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	castCatalogSpell(t, g, "Accelerated Mutation", "Instant", scgB1AcceleratedMutationOracle, cardRefs(bear))
	passPriorityAroundTable(t, g)
	if p, tough := effectivePower(t, g, bear), effectiveToughness(t, g, bear); p != 5 || tough != 5 {
		t.Errorf("the Bear is %d/%d, want 5/5", p, tough)
	}
}

func TestScgB1DispersalShieldComparesTheSpellsManaValue(t *testing.T) {
	for _, tc := range []struct {
		cost      string
		countered bool
	}{
		{"{1}{G}", true},
		{"{2}{G}", true},
		{"{3}{G}", false},
	} {
		t.Run(tc.cost, func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			b27Push(g, me.ID, "Three", "Enchantment", "", "{1}{U}{U}", 0, 0)
			spell := castFromHand(t, g, game.Card{Name: "A Bear", TypeLine: "Creature — Bear", ManaCost: tc.cost, Power: 2, Toughness: 2}, game.CastSpellParams{})
			castCatalogSpell(t, g, "Dispersal Shield", "Instant", scgB1DispersalShieldOracle, cardTarget(spell))
			passPriorityAroundTable(t, g)
			if got := !g.Battlefield.Contains(spell); got != tc.countered {
				t.Errorf("a %s spell against a greatest mana value of 3: countered %v, want %v", tc.cost, got, tc.countered)
			}
		})
	}
}

func TestScgB1RewardTheFaithfulGivesEachTargetTheSameLife(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, other := g.Seats[g.Turn.ActiveSeat], ons2Opponent(g), g.Seats[(g.Turn.ActiveSeat+2)%len(g.Seats)]
	b27Push(g, me.ID, "Three", "Enchantment", "", "{1}{W}{W}", 0, 0)
	mine, theirs, untouched := me.Life, opp.Life, other.Life
	castCatalogSpell(t, g, "Reward the Faithful", "Instant", scgB1RewardTheFaithfulOracle, playerRefs(me.ID, opp.ID))
	passPriorityAroundTable(t, g)
	if me.Life != mine+3 || opp.Life != theirs+3 || other.Life != untouched {
		t.Errorf("life %d/%d/%d, want +3, +3 and unchanged", me.Life-mine, opp.Life-theirs, other.Life-untouched)
	}
}

func TestScgB1CabalConditioningDiscardsTheGreatestManaValue(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], ons2Opponent(g)
	b27Push(g, me.ID, "Two", "Enchantment", "", "{1}{B}", 0, 0)
	a := pushToHand(g, opp.ID, "Card A", "Sorcery")
	b := pushToHand(g, opp.ID, "Card B", "Sorcery")
	before := len(opp.Hand.Cards)
	castCatalogSpell(t, g, "Cabal Conditioning", "Sorcery", scgB1CabalConditioningOracle, playerRefs(opp.ID))
	passPriorityAroundTable(t, g)
	answerDiscard(t, g, opp.ID, a, b)
	if got := len(opp.Hand.Cards); got != before-2 {
		t.Errorf("hand %d -> %d, want two discarded", before, got)
	}
}

// --- tribal cost reducers and lords ------------------------------------------

func TestScgB1WarchiefDiscounts(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, oracle string
		spellType, cost        string
		lands                  int
		ok                     bool
	}{
		{"Goblin Warchief", "Creature — Goblin Warrior", scgB1GoblinWarchiefOracle, "Creature — Goblin", "{2}{R}", 2, true},
		{"Goblin Warchief", "Creature — Goblin Warrior", scgB1GoblinWarchiefOracle, "Creature — Bear", "{2}{R}", 2, false},
		{"Daru Warchief", "Creature — Human Soldier", scgB1DaruWarchiefOracle, "Creature — Human Soldier", "{2}{R}", 2, true},
		{"Daru Warchief", "Creature — Human Soldier", scgB1DaruWarchiefOracle, "Creature — Bear", "{2}{R}", 2, false},
		{"Krosan Warchief", "Creature — Beast", scgB1KrosanWarchiefOracle, "Creature — Beast", "{2}{R}", 2, true},
		{"Krosan Warchief", "Creature — Beast", scgB1KrosanWarchiefOracle, "Creature — Bear", "{2}{R}", 2, false},
		{"Krosan Drover", "Creature — Elf", scgB1KrosanDroverOracle, "Creature — Wurm", "{5}{R}", 4, true},
		{"Krosan Drover", "Creature — Elf", scgB1KrosanDroverOracle, "Creature — Wurm", "{4}{R}", 3, false},
		{"Krosan Drover", "Creature — Elf", scgB1KrosanDroverOracle, "Sorcery", "{5}{R}", 4, false},
	} {
		g := newCatalogGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		b27Push(g, me.ID, tc.name, tc.typeLine, tc.oracle, "{1}", 1, 1)
		_, err := trekB37CastFromHand(t, g, game.Card{Name: "Spell", TypeLine: tc.spellType, ManaCost: tc.cost, Power: 1, Toughness: 1}, tc.lands)
		if (err == nil) != tc.ok {
			t.Errorf("%s: %s %s with %d Mountains: cast error %v, want ok=%v", tc.name, tc.spellType, tc.cost, tc.lands, err, tc.ok)
		}
	}
}

func TestScgB1GoblinWarchiefGivesYourGoblinsHaste(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	chief := b12Push(g, me.ID, "Goblin Warchief", "Creature — Goblin Warrior", scgB1GoblinWarchiefOracle, 2, 2)
	goblin := b12Creature(g, me.ID, "My Goblin", "Creature — Goblin", 1, 1)
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Goblin", "Creature — Goblin", 1, 1)
	if !hasEffectiveKeyword(t, g, chief, "haste") || !hasEffectiveKeyword(t, g, goblin, "haste") {
		t.Error("your Goblins, the Warchief included, lack haste")
	}
	if hasEffectiveKeyword(t, g, bear, "haste") || hasEffectiveKeyword(t, g, theirs, "haste") {
		t.Error("a non-Goblin or an opponent's Goblin has haste")
	}
}

func TestScgB1DaruWarchiefPumpsYourSoldiers(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	chief := b12Push(g, me.ID, "Daru Warchief", "Creature — Human Soldier", scgB1DaruWarchiefOracle, 1, 1)
	soldier := b12Creature(g, me.ID, "My Soldier", "Creature — Human Soldier", 1, 1)
	theirs := b12Creature(g, opp.ID, "Their Soldier", "Creature — Human Soldier", 1, 1)
	for _, c := range []struct {
		id          uuid.UUID
		power, toug int
	}{{chief, 2, 3}, {soldier, 2, 3}, {theirs, 1, 1}} {
		if p, tough := effectivePower(t, g, c.id), effectiveToughness(t, g, c.id); p != c.power || tough != c.toug {
			t.Errorf("%v is %d/%d, want %d/%d", c.id, p, tough, c.power, c.toug)
		}
	}
}

func TestScgB1KrosanWarchiefRegeneratesABeast(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	chief := b12Push(g, me.ID, "Krosan Warchief", "Creature — Beast", scgB1KrosanWarchiefOracle, 2, 2)
	beast := b12Creature(g, me.ID, "My Beast", "Creature — Beast", 3, 3)
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{G}{C}"); err != nil {
		t.Fatal(err)
	}
	if err := g.ActivateCatalogAbility(me.ID, chief, 0, game.ActivateAbilityParams{Targets: cardRefs(bear)}); err == nil {
		t.Fatal("a non-Beast was a legal target")
	}
	b16Activate(t, g, me.ID, chief, 0, game.ActivateAbilityParams{Targets: cardRefs(beast)})
	passPriorityAroundTable(t, g)
	b27Kill(g, beast)
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(beast) {
		t.Error("the regenerated Beast was destroyed")
	}
}

// --- the rest ----------------------------------------------------------------

func TestScgB1GoblinWarStrikeCountsYourGoblins(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], ons2Opponent(g)
	b12Creature(g, me.ID, "Goblin One", "Creature — Goblin", 1, 1)
	b12Creature(g, me.ID, "Goblin Two", "Creature — Goblin Warrior", 1, 1)
	b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	b12Creature(g, opp.ID, "Their Goblin", "Creature — Goblin", 1, 1)
	life := opp.Life
	castCatalogSpell(t, g, "Goblin War Strike", "Sorcery", scgB1GoblinWarStrikeOracle, playerRefs(opp.ID))
	passPriorityAroundTable(t, g)
	if opp.Life != life-2 {
		t.Errorf("life %d -> %d, want -2 (two Goblins you control)", life, opp.Life)
	}
}

func TestScgB1UnspeakableSymbolTradesLifeForCounters(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	symbol := pushCatalogPermanent(g, me.ID, "Unspeakable Symbol", "Enchantment", scgB1UnspeakableSymbolOracle, false)
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	life := me.Life
	for i := 0; i < 2; i++ {
		b16Activate(t, g, me.ID, symbol, 0, game.ActivateAbilityParams{Targets: cardRefs(bear)})
		passPriorityAroundTable(t, g)
	}
	if me.Life != life-6 {
		t.Errorf("life %d -> %d, want -6 (3 per activation)", life, me.Life)
	}
	if n := countersOn(g, bear, game.CounterPlusOne); n != 2 {
		t.Errorf("%d +1/+1 counters, want 2", n)
	}
}

func TestScgB1PyrostaticPillarBurnsCheapSpells(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	pushCatalogPermanent(g, me.ID, "Pyrostatic Pillar", "Enchantment", scgB1PyrostaticPillarOracle, false)
	life := me.Life
	castFromHand(t, g, game.Card{Name: "Cheap", TypeLine: "Sorcery", ManaCost: "{2}{R}"}, game.CastSpellParams{})
	passPriorityAroundTable(t, g)
	if me.Life != life-2 {
		t.Errorf("a mana value 3 spell: life %d -> %d, want -2", life, me.Life)
	}
	life = me.Life
	castFromHand(t, g, game.Card{Name: "Dear", TypeLine: "Sorcery", ManaCost: "{2}{R}{R}"}, game.CastSpellParams{})
	passPriorityAroundTable(t, g)
	if me.Life != life {
		t.Errorf("a mana value 4 spell: life %d -> %d, want unchanged", life, me.Life)
	}
}

func TestScgB1BladewingReturnsADragonPermanentCard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	dragon := scgB1GraveyardCreature(me, "Old Dragon", "Creature — Dragon")
	castFromHand(t, g, game.Card{Name: "Bladewing the Risen", TypeLine: "Legendary Creature — Zombie Dragon",
		OracleID: scgB1BladewingOracle, Power: 4, Toughness: 4}, game.CastSpellParams{})
	passPriorityAroundTable(t, g)
	if latestPickTarget(g, me.ID) != nil {
		pickTriggerTarget(t, g, me.ID, dragon)
	}
	passPriorityAroundTable(t, g)
	answerLatestTriggerPrompt(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	if countOnBattlefield(g, "Old Dragon", me.ID) != 1 {
		t.Error("the Dragon card did not return to the battlefield")
	}
}

func TestScgB1BladewingPumpsEveryDragon(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], ons2Opponent(g)
	advanceToMain(t, g)
	bladewing := b12Push(g, me.ID, "Bladewing the Risen", "Legendary Creature — Zombie Dragon", scgB1BladewingOracle, 4, 4)
	if !hasEffectiveKeyword(t, g, bladewing, "flying") {
		t.Error("Bladewing lacks flying")
	}
	theirs := b12Creature(g, opp.ID, "Their Dragon", "Creature — Dragon", 2, 2)
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{B}{R}"); err != nil {
		t.Fatal(err)
	}
	b16Activate(t, g, me.ID, bladewing, 0, game.ActivateAbilityParams{})
	passPriorityAroundTable(t, g)
	if effectivePower(t, g, bladewing) != 5 || effectivePower(t, g, theirs) != 3 || effectivePower(t, g, bear) != 2 {
		t.Error("want every Dragon creature +1/+1, any player's, and nothing else")
	}
}

func TestScgB1SliverOverlordTutorsASliver(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	overlord := b12Push(g, me.ID, "Sliver Overlord", "Legendary Creature — Sliver Mutant", scgB1SliverOverlordOracle, 7, 7)
	seedSearchLibrary(me,
		game.Card{Name: "Grizzly", TypeLine: "Creature — Bear"},
		game.Card{Name: "Muscle Sliver", TypeLine: "Creature — Sliver"},
	)
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{C}{C}{C}"); err != nil {
		t.Fatal(err)
	}
	b16Activate(t, g, me.ID, overlord, 0, game.ActivateAbilityParams{})
	passPriorityAroundTable(t, g)
	c := searchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no search prompt")
	}
	if searchOptionNamed(g, c, "Grizzly") != uuid.Nil {
		t.Error("a non-Sliver was offered")
	}
	answerSearchNamed(t, g, me.ID, "Muscle Sliver")
	if !b02bHandHasNamed(me, "Muscle Sliver") {
		t.Error("the Sliver did not reach the hand")
	}
}

func TestScgB1SliverOverlordStealsASliver(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], ons2Opponent(g)
	advanceToMain(t, g)
	overlord := b12Push(g, me.ID, "Sliver Overlord", "Legendary Creature — Sliver Mutant", scgB1SliverOverlordOracle, 7, 7)
	sliver := b12Creature(g, opp.ID, "Their Sliver", "Creature — Sliver", 1, 1)
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{C}{C}{C}"); err != nil {
		t.Fatal(err)
	}
	b16Activate(t, g, me.ID, overlord, 1, game.ActivateAbilityParams{Targets: cardRefs(sliver)})
	passPriorityAroundTable(t, g)
	c, ok := lookupBattlefield(g, sliver)
	if !ok || c.Controller != me.ID {
		t.Error("the Sliver did not change control")
	}
}

func TestScgB1SkulltapSacrificesThenDrawsTwo(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	if _, err := tryCastWithSacrifice(g, "Skulltap", "Sorcery", scgB1SkulltapOracle, []uuid.UUID{bear}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	if g.Battlefield.Contains(bear) {
		t.Fatal("the creature was not sacrificed as the spell was cast")
	}
	before := len(me.Hand.Cards)
	passPriorityAroundTable(t, g)
	if got := len(me.Hand.Cards); got != before+2 {
		t.Errorf("hand %d -> %d, want two cards drawn", before, got)
	}
}

func TestScgB1MisguidedRageSacrificesAPermanentOfTheirChoice(t *testing.T) {
	g := newCatalogGame(t)
	opp := ons2Opponent(g)
	land := b12Permanent(g, opp.ID, "Their Island", "Basic Land — Island")
	bear := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	castCatalogSpell(t, g, "Misguided Rage", "Sorcery", scgB1MisguidedRageOracle, playerRefs(opp.ID))
	passPriorityAroundTable(t, g)
	c := sacrificeChoiceFor(g, opp.ID)
	if c == nil {
		t.Fatal("no sacrifice prompt")
	}
	if opts := sacrificeOptions(c); !edictOffers(opts, land) || !edictOffers(opts, bear) {
		t.Fatalf("the prompt offers %v, want the land and the creature", opts)
	}
	answerSacrifice(t, g, opp.ID, land)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(land) || !g.Battlefield.Contains(bear) {
		t.Error("want only the chosen land sacrificed")
	}
}

func TestScgB1EnragePumpsPowerByX(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	castFromHand(t, g, game.Card{Name: "Enrage", TypeLine: "Instant", OracleID: scgB1EnrageOracle, ManaCost: "{X}{R}"},
		game.CastSpellParams{XValue: 3, Targets: cardRefs(bear)})
	passPriorityAroundTable(t, g)
	if p, tough := effectivePower(t, g, bear), effectiveToughness(t, g, bear); p != 5 || tough != 2 {
		t.Errorf("the Bear is %d/%d, want 5/2", p, tough)
	}
}

func TestScgB1PemminsAuraAbilitiesReachTheHost(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	aura := auraCast(t, g, "Pemmin's Aura", scgB1PemminsAuraOracle, bear)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == bear {
				g.Battlefield.Cards[i].Tapped = true
			}
		}
	})
	activate := func(idx int) {
		t.Helper()
		if err := g.AddManaForEffect(me.ID, uuid.Nil, "{U}"); err != nil {
			t.Fatal(err)
		}
		b16Activate(t, g, me.ID, aura, idx, game.ActivateAbilityParams{})
		passPriorityAroundTable(t, g)
	}
	activate(0)
	if c, _ := lookupBattlefield(g, bear); c.Tapped {
		t.Error("{U}: the host was not untapped")
	}
	activate(1)
	if !hasEffectiveKeyword(t, g, bear, "flying") {
		t.Error("{U}: the host did not gain flying")
	}
	activate(2)
	if !hasEffectiveKeyword(t, g, bear, "shroud") {
		t.Error("{U}: the host did not gain shroud")
	}
	activate(3)
	answerMayChoice(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	if p, tough := effectivePower(t, g, bear), effectiveToughness(t, g, bear); p != 3 || tough != 1 {
		t.Errorf("after +1/-1 the host is %d/%d, want 3/1", p, tough)
	}
	activate(3)
	answerMayChoice(t, g, me.ID, false)
	passPriorityAroundTable(t, g)
	if p, tough := effectivePower(t, g, bear), effectiveToughness(t, g, bear); p != 2 || tough != 2 {
		t.Errorf("after -1/+1 the host is %d/%d, want 2/2", p, tough)
	}
}

func TestScgB1OneWithNatureFetchesATappedBasic(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], ons2Opponent(g)
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	auraCast(t, g, "One with Nature", scgB1OneWithNatureOracle, bear)
	seedSearchLibrary(me,
		game.Card{Name: "Grizzly", TypeLine: "Creature — Bear"},
		game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"},
	)
	dealCombatDamageToPlayer(g, bear, opp.ID, 2)
	passPriorityAroundTable(t, g)
	answerLatestTriggerPrompt(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	c := searchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no search prompt")
	}
	if searchOptionNamed(g, c, "Grizzly") != uuid.Nil {
		t.Error("a nonland card was offered")
	}
	answerSearchNamed(t, g, me.ID, "Forest")
	found := false
	for _, card := range g.Battlefield.Cards {
		if card.Name == "Forest" && card.Controller == me.ID {
			found = true
			if !card.Tapped {
				t.Error("the Forest entered untapped")
			}
		}
	}
	if !found {
		t.Error("the Forest is not on the battlefield")
	}
}
