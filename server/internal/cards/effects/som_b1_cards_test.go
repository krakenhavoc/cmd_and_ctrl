package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// som_b1_cards_test.go — Scars of Mirrodin, batch 1 (#5577). Each test
// plays the real catalog entry and asserts what its text says.

const (
	somB1JournalOracle      = "f7907327-4df3-4bb2-a274-03d9e266af2f"
	somB1EzuriOracle        = "6734497c-16f0-4c4b-ba24-337333511fc6"
	somB1ThroneOracle       = "5a5f39ab-21b4-415f-aa1f-d8fad420d372"
	somB1CopperhornOracle   = "3d2ea1a3-72af-48bc-a8f1-98facc2eebd0"
	somB1TanglewurmOracle   = "a1091651-4e3e-4840-87e2-b80084770137"
	somB1ShieldOracle       = "dd70d439-bd60-43d1-ac26-13b744cc4a37"
	somB1TurnAsideOracle    = "4325ae73-af43-4a7d-a552-9909f0dc77db"
	somB1GalvanizerOracle   = "e070028e-468a-41e9-9ddd-a8cd01551266"
	somB1GlimmerpostOracle  = "92c9aad6-35ec-425d-be7d-393328992820"
	somB1RebuttalOracle     = "6dd5782d-2d1a-4d6b-8381-e15d0d5be24f"
	somB1DarksteelMyrOracle = "f90fe86c-5cca-483b-93c2-6e856fa01c88"
	somB1TanglecordOracle   = "423bebcc-8fa0-49c4-aa22-728bf51a535b"
	somB1DaisOracle         = "42b03dd7-e796-47c3-b596-87960d72818b"
	somB1RevokeOracle       = "289e1a7c-b301-4395-baf8-e9a1c6d38aab"
	somB1PraetorsOracle     = "f595c7f6-783f-4825-9c1a-e7f8edd380fd"
	somB1StrobeOracle       = "9bf4ee0f-6e28-4793-b251-f81ee444a119"
	somB1NecropedeOracle    = "e5021798-1f3d-42ed-be88-e76349ef46c6"
	somB1KembaOracle        = "499d5801-0c56-40f2-9a46-77025ed6cb87"
	somB1PerilousOracle     = "f1e8f746-1837-4ea8-9a07-e57946883353"
	somB1WithstandOracle    = "c462adf3-4ade-4b90-bdec-7247b080471f"
	somB1PropagatorOracle   = "d071e2fc-3be8-4168-afa9-ba34edad18b9"
	somB1FumeSpitterOracle  = "6556a15b-be4b-45ea-867d-3a8e26202b42"
	somB1CorpseCurOracle    = "8f59e4d9-67ba-4dc6-ae3e-183e63673318"
	somB1GraspOracle        = "494c819d-7d73-432b-8c96-4cb9dcd31094"
	somB1DisperseOracle     = "9d8fe3f8-5027-4f9e-999e-e277caa96478"
	somB1BluntOracle        = "dccd0533-fda0-41ff-b540-c4f1bb32fc95"
	somB1PinionsOracle      = "c8cdc08a-975b-475f-8e30-85d61f6f3a9b"
	somB1AxeOracle          = "3b7ea3ac-ac0a-40aa-b743-d153e1c47d8c"
	somB1GolemsHeartOracle  = "650890bc-aba6-45fb-8096-3a36b0a015b3"
	somB1GrindclockOracle   = "6b7de1c0-4b97-47e1-8ea7-ba55a8d16a9b"
	somB1StagOracle         = "9620c8a2-63a8-4151-b4fd-a15db191a7b0"
	somB1SliceOracle        = "a96d0065-b8c1-4843-8752-dce60bf83545"
	somB1PutrefaxOracle     = "3a336b30-335a-49d0-b192-c20f3670452b"
	somB1SpikeshotOracle    = "8307bfb0-3ab2-467f-b13c-ad4acadf319c"
	somB1SentinelOracle     = "b91c8946-591b-4c0d-a37e-36803df40db7"
	somB1BlackcleaveOracle  = "156d1a58-c970-4b81-b6c7-5a0caa1699cd"
	somB1VectorAspOracle    = "1950067a-cc57-42a8-b1ee-f03e1b50591b"
	somB1ChromeSteedOracle  = "8475f55e-473c-4df0-bae2-3cbbbde4ff12"
	somB1InvisimancerOracle = "e7425964-0028-4156-9037-b303b00e7e11"
	somB1LensOracle         = "72dfd729-ebba-4bb1-83eb-54c2051fea3b"
)

func TestSomB1EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		somB1JournalOracle:      "Venser's Journal",
		somB1EzuriOracle:        "Ezuri, Renegade Leader",
		somB1ThroneOracle:       "Throne of Geth",
		somB1CopperhornOracle:   "Copperhorn Scout",
		somB1TanglewurmOracle:   "Bellowing Tanglewurm",
		somB1ShieldOracle:       "Accorder's Shield",
		somB1TurnAsideOracle:    "Turn Aside",
		somB1GalvanizerOracle:   "Myr Galvanizer",
		somB1GlimmerpostOracle:  "Glimmerpost",
		somB1RebuttalOracle:     "Stoic Rebuttal",
		somB1DarksteelMyrOracle: "Darksteel Myr",
		somB1TanglecordOracle:   "Wall of Tanglecord",
		somB1DaisOracle:         "Culling Dais",
		somB1RevokeOracle:       "Revoke Existence",
		somB1PraetorsOracle:     "Hand of the Praetors",
		somB1StrobeOracle:       "Assault Strobe",
		somB1NecropedeOracle:    "Necropede",
		somB1KembaOracle:        "Kemba, Kha Regent",
		somB1PerilousOracle:     "Perilous Myr",
		somB1WithstandOracle:    "Withstand Death",
		somB1PropagatorOracle:   "Myr Propagator",
		somB1FumeSpitterOracle:  "Fume Spitter",
		somB1CorpseCurOracle:    "Corpse Cur",
		somB1GraspOracle:        "Grasp of Darkness",
		somB1DisperseOracle:     "Disperse",
		somB1BluntOracle:        "Blunt the Assault",
		somB1PinionsOracle:      "Bladed Pinions",
		somB1AxeOracle:          "Darksteel Axe",
		somB1GolemsHeartOracle:  "Golem's Heart",
		somB1GrindclockOracle:   "Grindclock",
		somB1StagOracle:         "Glimmerpoint Stag",
		somB1SliceOracle:        "Slice in Twain",
		somB1PutrefaxOracle:     "Putrefax",
		somB1SpikeshotOracle:    "Spikeshot Elder",
		somB1SentinelOracle:     "Darksteel Sentinel",
		somB1BlackcleaveOracle:  "Blackcleave Goblin",
		somB1VectorAspOracle:    "Vector Asp",
		somB1ChromeSteedOracle:  "Chrome Steed",
		somB1InvisimancerOracle: "Neurok Invisimancer",
		somB1LensOracle:         "Infiltration Lens",
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

// --- shared test helpers --------------------------------------------------------------

// somB1Equip puts an Equipment on the battlefield and attaches it to host.
func somB1Equip(t *testing.T, g *game.Game, owner uuid.UUID, name, oracle string, host uuid.UUID) uuid.UUID {
	t.Helper()
	eq := b12Push(g, owner, name, "Artifact — Equipment", oracle, 0, 0)
	attachForTest(t, g, eq, host)
	return eq
}

// somB1CastWithKeywords casts a non-catalog spell that prints keywords
// from the active seat's hand.
func somB1CastWithKeywords(t *testing.T, g *game.Game, name, typeLine string, keywords ...string) uuid.UUID {
	t.Helper()
	active := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	active.Hand.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: typeLine, Power: 1, Toughness: 1,
		Keywords: keywords, Owner: active.ID, Controller: active.ID,
	})
	toMainForCost(t, g)
	if err := g.CastSpell(active.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("CastSpell %s: %v", name, err)
	}
	return id
}

// --- printed keywords ------------------------------------------------------------------

func TestSomB1CardsHaveTheirPrintedKeywords(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, oracle string
		keywords               []string
	}{
		{"Darksteel Myr", "Artifact Creature — Myr", somB1DarksteelMyrOracle, []string{"indestructible"}},
		{"Darksteel Sentinel", "Artifact Creature — Golem", somB1SentinelOracle, []string{"flash", "vigilance", "indestructible"}},
		{"Blackcleave Goblin", "Creature — Phyrexian Goblin Zombie", somB1BlackcleaveOracle, []string{"haste", "infect"}},
		{"Putrefax", "Creature — Phyrexian Horror", somB1PutrefaxOracle, []string{"trample", "haste", "infect"}},
		{"Bellowing Tanglewurm", "Creature — Wurm", somB1TanglewurmOracle, []string{"intimidate"}},
		{"Wall of Tanglecord", "Artifact Creature — Wall", somB1TanglecordOracle, []string{"defender"}},
		{"Necropede", "Artifact Creature — Phyrexian Insect", somB1NecropedeOracle, []string{"infect"}},
		{"Corpse Cur", "Artifact Creature — Phyrexian Dog", somB1CorpseCurOracle, []string{"infect"}},
		{"Hand of the Praetors", "Creature — Phyrexian Zombie", somB1PraetorsOracle, []string{"infect"}},
		{"Glimmerpoint Stag", "Creature — Elk", somB1StagOracle, []string{"vigilance"}},
	} {
		g := newCatalogGame(t)
		id := b12Push(g, g.Seats[0].ID, tc.name, tc.typeLine, tc.oracle, 2, 2)
		for _, kw := range tc.keywords {
			if !hasKeywordOnBattlefield(t, g, id, kw) {
				t.Errorf("%s lacks %s", tc.name, kw)
			}
		}
	}
	g := newCatalogGame(t)
	axe := b12Push(g, g.Seats[0].ID, "Darksteel Axe", "Artifact — Equipment", somB1AxeOracle, 0, 0)
	if !hasKeywordOnBattlefield(t, g, axe, "indestructible") {
		t.Error("Darksteel Axe is not indestructible")
	}
}

// --- Venser's Journal ------------------------------------------------------------------

func TestSomB1VensersJournalGainsLifePerCardInHand(t *testing.T) {
	spec, _ := Lookup(somB1JournalOracle)
	if !spec.NoMaxHandSize {
		t.Error("Venser's Journal does not lift the maximum hand size")
	}
	g := newCatalogGame(t)
	me := g.Seats[1]
	b12Push(g, me.ID, "Venser's Journal", "Artifact — Book", somB1JournalOracle, 0, 0)
	advanceToUpkeepOf(t, g, 1)
	d2dSetHand(me, 6)
	start := me.Life
	passPriorityAroundTable(t, g)
	if me.Life != start+6 {
		t.Errorf("life %d → %d, want +6 for six cards in hand", start, me.Life)
	}
}

// --- Ezuri, Renegade Leader ------------------------------------------------------------

func TestSomB1EzuriPumpsYourElvesAndRegeneratesAnother(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := seatsForCombat(g)
	advanceToMain(t, g)
	ezuri := b12Push(g, me.ID, "Ezuri, Renegade Leader", "Legendary Creature — Elf Warrior", somB1EzuriOracle, 2, 2)
	elf := b12Creature(g, me.ID, "Elf", "Creature — Elf", 1, 1)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirElf := b12Creature(g, opp.ID, "Their Elf", "Creature — Elf", 1, 1)

	if err := ravB2Activate(t, g, me.ID, ezuri, 0, game.ActivateAbilityParams{Targets: cardRefs(ezuri)}, "G"); err == nil {
		t.Error("Ezuri regenerated itself")
	}
	ravB2MustActivate(t, g, me.ID, ezuri, 0, game.ActivateAbilityParams{Targets: cardRefs(elf)}, "G")
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(elf); err != nil {
			t.Fatal(err)
		}
	})
	if !g.Battlefield.Contains(elf) {
		t.Error("the regenerated Elf was destroyed")
	}

	ravB2MustActivate(t, g, me.ID, ezuri, 1, game.ActivateAbilityParams{}, "C", "C", "G", "G", "G")
	for id, want := range map[uuid.UUID][2]int{ezuri: {5, 5}, elf: {4, 4}, bear: {2, 2}, theirElf: {1, 1}} {
		if p, tough := effectivePTOf(t, g, id); p != want[0] || tough != want[1] {
			t.Errorf("%s is %d/%d, want %d/%d", b12Card(t, g, id).Name, p, tough, want[0], want[1])
		}
	}
	if !hasKeywordOnBattlefield(t, g, elf, "trample") || hasKeywordOnBattlefield(t, g, bear, "trample") {
		t.Error("trample went to the wrong creatures")
	}
}

// --- Throne of Geth --------------------------------------------------------------------

func TestSomB1ThroneOfGethSacrificesAnArtifactToProliferate(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	throne := b12Push(g, me.ID, "Throne of Geth", "Artifact", somB1ThroneOracle, 0, 0)
	trinket := b12Permanent(g, me.ID, "Trinket", "Artifact")
	bear := counterBear(g, me.ID, map[string]int{game.CounterPlusOne: 1})
	if err := ravB2Activate(t, g, me.ID, throne, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{bear}}); err == nil {
		t.Fatal("Throne of Geth sacrificed a creature")
	}
	if err := ravB2Activate(t, g, me.ID, throne, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{trinket}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if g.Battlefield.Contains(trinket) {
		t.Error("the artifact was not sacrificed")
	}
	passPriorityAroundTable(t, g)
	answerProliferate(t, g, bear)
	passPriorityAroundTable(t, g)
	if got := counterCount(g, bear, game.CounterPlusOne); got != 2 {
		t.Errorf("+1/+1 counters = %d, want 2", got)
	}

	// The card's ruling: the Throne may pay its own cost.
	g2 := newCatalogGame(t)
	advanceToMain(t, g2)
	throne2 := b12Push(g2, g2.Seats[0].ID, "Throne of Geth", "Artifact", somB1ThroneOracle, 0, 0)
	if err := ravB2Activate(t, g2, g2.Seats[0].ID, throne2, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{throne2}}); err != nil {
		t.Errorf("the Throne could not sacrifice itself: %v", err)
	}
}

// --- Copperhorn Scout ------------------------------------------------------------------

func TestSomB1CopperhornScoutUntapsYourOtherCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := seatsForCombat(g)
	scout := b12Push(g, me.ID, "Copperhorn Scout", "Creature — Elf Scout", somB1CopperhornOracle, 1, 1)
	friend := b12Creature(g, me.ID, "Friend", "Creature — Bear", 2, 2)
	foe := b12Creature(g, opp.ID, "Foe", "Creature — Bear", 2, 2)
	declareAttack(t, g, opp.ID, scout, friend)
	g.WithWriteLock(func() { findBattlefieldCardForTest(g, foe).Tapped = true })
	passPriorityAroundTable(t, g)
	if isTapped(g, friend) {
		t.Error("the other attacker is still tapped")
	}
	if !isTapped(g, scout) {
		t.Error("the Scout untapped itself")
	}
	if !isTapped(g, foe) {
		t.Error("an opponent's creature was untapped")
	}
}

// --- Bellowing Tanglewurm ----------------------------------------------------------------

func TestSomB1TanglewurmGivesYourOtherGreenCreaturesIntimidate(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := seatsForCombat(g)
	b12Push(g, me.ID, "Bellowing Tanglewurm", "Creature — Wurm", somB1TanglewurmOracle, 4, 4)
	green := pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Green", TypeLine: "Creature — Elf",
		Colors: []string{"G"}, Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID})
	red := pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Red", TypeLine: "Creature — Goblin",
		Colors: []string{"R"}, Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID})
	theirs := pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Their Green", TypeLine: "Creature — Elf",
		Colors: []string{"G"}, Power: 1, Toughness: 1, Owner: opp.ID, Controller: opp.ID})
	if !hasKeywordOnBattlefield(t, g, green, "intimidate") {
		t.Error("your other green creature lacks intimidate")
	}
	if hasKeywordOnBattlefield(t, g, red, "intimidate") || hasKeywordOnBattlefield(t, g, theirs, "intimidate") {
		t.Error("intimidate went to a nongreen creature or an opponent's")
	}
}

// --- Equipment -------------------------------------------------------------------------

func TestSomB1EquipmentBoostsAndGrants(t *testing.T) {
	for _, tc := range []struct {
		name, oracle string
		power, tough int
		keywords     []string
	}{
		{"Accorder's Shield", somB1ShieldOracle, 2, 5, []string{"vigilance"}},
		{"Bladed Pinions", somB1PinionsOracle, 2, 2, []string{"flying", "first strike"}},
		{"Darksteel Axe", somB1AxeOracle, 4, 2, nil},
	} {
		g := newCatalogGame(t)
		me := g.Seats[0]
		advanceToMain(t, g)
		bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
		eq := b12Push(g, me.ID, tc.name, "Artifact — Equipment", tc.oracle, 0, 0)
		fillPool(me, 3)
		equipTo(t, g, me.ID, eq, bear)
		if p, tough := effectivePTOf(t, g, bear); p != tc.power || tough != tc.tough {
			t.Errorf("%s: the creature is %d/%d, want %d/%d", tc.name, p, tough, tc.power, tc.tough)
		}
		for _, kw := range tc.keywords {
			if !hasKeywordOnBattlefield(t, g, bear, kw) {
				t.Errorf("%s: the creature lacks %s", tc.name, kw)
			}
		}
	}
}

// --- Infiltration Lens -------------------------------------------------------------------

func TestSomB1InfiltrationLensDrawsTwoPerBlocker(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := seatsForCombat(g)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	somB1Equip(t, g, me.ID, "Infiltration Lens", somB1LensOracle, bear)
	a := b12Creature(g, opp.ID, "Wall A", "Creature — Wall", 0, 4)
	b := b12Creature(g, opp.ID, "Wall B", "Creature — Wall", 0, 4)
	declareAttack(t, g, opp.ID, bear)
	advanceTo(t, g, game.StepDeclareBlockers)
	for _, w := range []uuid.UUID{a, b} {
		if err := g.DeclareBlocker(w, bear); err != nil {
			t.Fatalf("DeclareBlocker: %v", err)
		}
	}
	hand := me.Hand.Size()
	lockInBlocks(t, g)
	answerLatestTriggerPrompt(t, g, me.ID, true)
	answerLatestTriggerPrompt(t, g, me.ID, false)
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 2 {
		t.Errorf("drew %d cards, want 2 (one yes, one no)", got)
	}
}

// --- Turn Aside ------------------------------------------------------------------------

func TestSomB1TurnAsideCountersOnlyASpellAimedAtYourPermanent(t *testing.T) {
	for _, aimAtResponder := range []bool{true, false} {
		g := newCatalogGame(t)
		caster, responder := g.Seats[0], g.Seats[1]
		toMainForCost(t, g)
		theirs := b12Creature(g, responder.ID, "Responder's Bear", "Creature — Bear", 2, 2)
		own := b12Creature(g, caster.ID, "Caster's Bear", "Creature — Bear", 2, 2)
		aim := own
		if aimAtResponder {
			aim = theirs
		}
		shock := castCatalogSpell(t, g, "Shock", "Instant", "", cardRefs(aim))
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
		spell := handCardFull(responder, "Turn Aside", "Instant", "", somB1TurnAsideOracle, []string{"U"})
		err := g.CastSpell(responder.ID, spell, game.CastSpellParams{Targets: cardRefs(shock)})
		if !aimAtResponder {
			if err == nil {
				t.Error("Turn Aside targeted a spell aimed at another player's permanent")
			}
			continue
		}
		if err != nil {
			t.Fatalf("casting Turn Aside: %v", err)
		}
		passPriorityAroundTable(t, g)
		if g.Stack.Contains(shock) {
			t.Error("the spell aimed at the responder's permanent was not countered")
		}
	}
}

// --- Myr Galvanizer ----------------------------------------------------------------------

func TestSomB1MyrGalvanizerPumpsAndUntapsYourOtherMyr(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := seatsForCombat(g)
	advanceToMain(t, g)
	galv := b12Push(g, me.ID, "Myr Galvanizer", "Artifact Creature — Myr", somB1GalvanizerOracle, 2, 2)
	myr := b12Creature(g, me.ID, "Plain Myr", "Artifact Creature — Myr", 1, 1)
	theirMyr := b12Creature(g, opp.ID, "Their Myr", "Artifact Creature — Myr", 1, 1)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	for id, want := range map[uuid.UUID]int{galv: 2, myr: 2, theirMyr: 1, bear: 2} {
		if p, _ := effectivePTOf(t, g, id); p != want {
			t.Errorf("%s has power %d, want %d", b12Card(t, g, id).Name, p, want)
		}
	}
	g.WithWriteLock(func() {
		findBattlefieldCardForTest(g, myr).Tapped = true
		findBattlefieldCardForTest(g, bear).Tapped = true
		findBattlefieldCardForTest(g, theirMyr).Tapped = true
	})
	ravB2MustActivate(t, g, me.ID, galv, 0, game.ActivateAbilityParams{}, "C")
	if isTapped(g, myr) {
		t.Error("your other Myr is still tapped")
	}
	if !isTapped(g, bear) || !isTapped(g, theirMyr) {
		t.Error("a non-Myr or an opponent's Myr was untapped")
	}
	if !isTapped(g, galv) {
		t.Error("the Galvanizer did not tap as the cost")
	}
}

// --- Glimmerpost -----------------------------------------------------------------------

func TestSomB1GlimmerpostGainsLifePerLocus(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	seedLandOnBattlefield(g, me.ID, "Cloudpost", "Land — Locus")
	seedLandOnBattlefield(g, g.Seats[1].ID, "Their Cloudpost", "Land — Locus")
	start := me.Life
	b12PlayFromHand(t, g, "Glimmerpost", "Land — Locus", somB1GlimmerpostOracle, game.CastSpellParams{})
	passPriorityAroundTable(t, g)
	if me.Life != start+3 {
		t.Errorf("life %d → %d, want +3 for three Loci", start, me.Life)
	}
}

// --- Stoic Rebuttal ----------------------------------------------------------------------

func TestSomB1StoicRebuttalIsCheaperWithMetalcraft(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b12Permanent(g, me.ID, "A", "Artifact")
	b12Permanent(g, me.ID, "B", "Artifact")
	if got := selfPricedMV(t, g, me, somB1RebuttalOracle, "Instant", "{1}{U}{U}"); got != 3 {
		t.Errorf("with two artifacts it costs %d, want 3", got)
	}
	b12Permanent(g, me.ID, "C", "Artifact")
	if got := selfPricedMV(t, g, me, somB1RebuttalOracle, "Instant", "{1}{U}{U}"); got != 2 {
		t.Errorf("with three artifacts it costs %d, want 2", got)
	}
}

// --- Wall of Tanglecord, Vector Asp --------------------------------------------------------

func TestSomB1SelfGrantsLastUntilEndOfTurn(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, oracle, colour, keyword string
	}{
		{"Wall of Tanglecord", "Artifact Creature — Wall", somB1TanglecordOracle, "G", "reach"},
		{"Vector Asp", "Artifact Creature — Phyrexian Snake", somB1VectorAspOracle, "B", "infect"},
	} {
		g := newCatalogGame(t)
		me := g.Seats[0]
		id := b12Push(g, me.ID, tc.name, tc.typeLine, tc.oracle, 1, 1)
		if hasKeywordOnBattlefield(t, g, id, tc.keyword) {
			t.Fatalf("%s has %s before activating", tc.name, tc.keyword)
		}
		ravB2MustActivate(t, g, me.ID, id, 0, game.ActivateAbilityParams{}, tc.colour)
		if !hasKeywordOnBattlefield(t, g, id, tc.keyword) {
			t.Errorf("%s did not gain %s", tc.name, tc.keyword)
		}
		dcCleanup(g)
		if hasKeywordOnBattlefield(t, g, id, tc.keyword) {
			t.Errorf("%s kept %s past the cleanup step", tc.name, tc.keyword)
		}
	}
}

// --- Culling Dais --------------------------------------------------------------------------

func TestSomB1CullingDaisBanksCreaturesThenDraws(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	dais := b12Push(g, me.ID, "Culling Dais", "Artifact", somB1DaisOracle, 0, 0)
	for i := 0; i < 2; i++ {
		bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
		ravB2MustActivate(t, g, me.ID, dais, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{bear}})
		if g.Battlefield.Contains(bear) {
			t.Fatal("the creature was not sacrificed")
		}
		g.WithWriteLock(func() { findBattlefieldCardForTest(g, dais).Tapped = false })
	}
	if got := counterCount(g, dais, game.CounterCharge); got != 2 {
		t.Fatalf("charge counters = %d, want 2", got)
	}
	hand := me.Hand.Size()
	ravB2MustActivate(t, g, me.ID, dais, 1, game.ActivateAbilityParams{}, "C")
	if g.Battlefield.Contains(dais) {
		t.Error("the Dais was not sacrificed")
	}
	if got := me.Hand.Size() - hand; got != 2 {
		t.Errorf("drew %d, want 2", got)
	}
}

// --- Grindclock ----------------------------------------------------------------------------

func TestSomB1GrindclockMillsItsCount(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	clock := b12Push(g, me.ID, "Grindclock", "Artifact", somB1GrindclockOracle, 0, 0)
	for i := 0; i < 3; i++ {
		ravB2MustActivate(t, g, me.ID, clock, 0, game.ActivateAbilityParams{})
		g.WithWriteLock(func() { findBattlefieldCardForTest(g, clock).Tapped = false })
	}
	lib := opp.Library.Size()
	ravB2MustActivate(t, g, me.ID, clock, 1, game.ActivateAbilityParams{Targets: []game.TargetRef{playerTarget(opp.ID)}})
	if got := lib - opp.Library.Size(); got != 3 {
		t.Errorf("milled %d, want 3", got)
	}
}

// --- Revoke Existence, Slice in Twain --------------------------------------------------------

func TestSomB1RevokeExistenceExilesAnArtifact(t *testing.T) {
	g := newCatalogGame(t)
	_, opp := seatsForCombat(g)
	art := b12Permanent(g, opp.ID, "Their Artifact", "Artifact")
	bear := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	if err := castCatalogSpellErr(t, g, "Revoke Existence", "Sorcery", somB1RevokeOracle, cardTarget(bear)); err == nil {
		t.Fatal("Revoke Existence targeted a creature")
	}
	castCatalogSpell(t, g, "Revoke Existence", "Sorcery", somB1RevokeOracle, cardTarget(art))
	passPriorityAroundTable(t, g)
	if b12ZoneOf(g, art) != game.ZoneExile {
		t.Errorf("the artifact is in %q, want exile", b12ZoneOf(g, art))
	}
}

func TestSomB1SliceInTwainDestroysAndDraws(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := seatsForCombat(g)
	ench := b12Permanent(g, opp.ID, "Their Enchantment", "Enchantment")
	hand := me.Hand.Size()
	castCatalogSpell(t, g, "Slice in Twain", "Instant", somB1SliceOracle, cardTarget(ench))
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(ench) || b12ZoneOf(g, ench) != game.ZoneGraveyard {
		t.Error("the enchantment was not destroyed")
	}
	// castCatalogSpell puts the spell in hand and casts it from there, so
	// the hand ends one card up: the draw.
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d → %d, want one draw", hand, me.Hand.Size())
	}
}

// --- Hand of the Praetors --------------------------------------------------------------------

func TestSomB1HandOfThePraetorsPumpsInfectAndPoisonsOnCast(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := seatsForCombat(g)
	hand := b12Push(g, me.ID, "Hand of the Praetors", "Creature — Phyrexian Zombie", somB1PraetorsOracle, 3, 2)
	infected := iceB1Push(g, me.ID, "Infected", "Creature — Horror", "", 1, 1, "infect")
	plain := b12Creature(g, me.ID, "Plain", "Creature — Bear", 1, 1)
	theirs := iceB1Push(g, opp.ID, "Their Infected", "Creature — Horror", "", 1, 1, "infect")
	for id, want := range map[uuid.UUID]int{hand: 3, infected: 2, plain: 1, theirs: 1} {
		if p, _ := effectivePTOf(t, g, id); p != want {
			t.Errorf("%s has power %d, want %d", b12Card(t, g, id).Name, p, want)
		}
	}

	somB1CastWithKeywords(t, g, "Plain Spell", "Creature — Bear")
	if passUntilPickTarget(t, g, me.ID) != nil {
		t.Fatal("a creature spell without infect triggered it")
	}
	passPriorityAroundTable(t, g)
	somB1CastWithKeywords(t, g, "Infect Spell", "Creature — Horror", "infect")
	pick := passUntilPickTarget(t, g, me.ID)
	if pick == nil {
		t.Fatal("casting a creature spell with infect did not trigger it")
	}
	if err := g.ResolvePickTarget(pick.ID, me.ID, playerTarget(opp.ID)); err != nil {
		t.Fatalf("ResolvePickTarget: %v", err)
	}
	passPriorityAroundTable(t, g)
	if poisonOn(opp) != 1 {
		t.Errorf("the target has %d poison, want 1", poisonOn(opp))
	}
}

// --- Assault Strobe, Withstand Death, Grasp of Darkness ----------------------------------------

func TestSomB1TargetedUntilEndOfTurnSpells(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, oracle, keyword string
		power, tough                    int
	}{
		{"Assault Strobe", "Sorcery", somB1StrobeOracle, "double strike", 5, 5},
		{"Withstand Death", "Instant", somB1WithstandOracle, "indestructible", 5, 5},
		{"Grasp of Darkness", "Instant", somB1GraspOracle, "", 1, 1},
	} {
		g := newCatalogGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		ogre := b12Creature(g, me.ID, "Ogre", "Creature — Ogre", 5, 5)
		castCatalogSpell(t, g, tc.name, tc.typeLine, tc.oracle, cardTarget(ogre))
		passPriorityAroundTable(t, g)
		if p, tough := effectivePTOf(t, g, ogre); p != tc.power || tough != tc.tough {
			t.Errorf("%s: the target is %d/%d, want %d/%d", tc.name, p, tough, tc.power, tc.tough)
		}
		if tc.keyword != "" && !hasKeywordOnBattlefield(t, g, ogre, tc.keyword) {
			t.Errorf("%s: the target lacks %s", tc.name, tc.keyword)
		}
		dcCleanup(g)
		if p, tough := effectivePTOf(t, g, ogre); p != 5 || tough != 5 {
			t.Errorf("%s: the effect outlived the cleanup step (%d/%d)", tc.name, p, tough)
		}
		if tc.keyword != "" && hasKeywordOnBattlefield(t, g, ogre, tc.keyword) {
			t.Errorf("%s: %s outlived the cleanup step", tc.name, tc.keyword)
		}
	}
}

// --- Disperse --------------------------------------------------------------------------------

func TestSomB1DisperseBouncesANonlandPermanent(t *testing.T) {
	g := newCatalogGame(t)
	_, opp := seatsForCombat(g)
	art := b12Permanent(g, opp.ID, "Their Artifact", "Artifact")
	land := seedLandOnBattlefield(g, opp.ID, "Forest", "Basic Land — Forest")
	if err := castCatalogSpellErr(t, g, "Disperse", "Instant", somB1DisperseOracle, cardTarget(land)); err == nil {
		t.Fatal("Disperse targeted a land")
	}
	castCatalogSpell(t, g, "Disperse", "Instant", somB1DisperseOracle, cardTarget(art))
	passPriorityAroundTable(t, g)
	if !opp.Hand.Contains(art) {
		t.Error("the permanent did not return to its owner's hand")
	}
}

// --- Blunt the Assault -------------------------------------------------------------------------

func TestSomB1BluntTheAssaultGainsPerCreatureAndFogs(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := seatsForCombat(g)
	defender := b12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	attacker := b12Creature(g, opp.ID, "Theirs", "Creature — Bear", 3, 3)
	b12Creature(g, g.Seats[2].ID, "Third", "Creature — Bear", 2, 2)
	start := me.Life
	castCatalogSpell(t, g, "Blunt the Assault", "Instant", somB1BluntOracle, nil)
	passPriorityAroundTable(t, g)
	if me.Life != start+3 {
		t.Fatalf("life %d → %d, want +3 for three creatures", start, me.Life)
	}
	if err := g.MarkCombatDamage(attacker, defender, 3); err != nil {
		t.Fatalf("MarkCombatDamage: %v", err)
	}
	if got := b12Card(t, g, defender).DamageMarked; got != 0 {
		t.Errorf("combat damage marked %d, want 0", got)
	}
}

// --- Necropede, Perilous Myr ---------------------------------------------------------------------

func TestSomB1NecropedeMayShrinkACreatureAsItDies(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := seatsForCombat(g)
	pede := b12Push(g, me.ID, "Necropede", "Artifact Creature — Phyrexian Insect", somB1NecropedeOracle, 1, 1)
	ogre := b12Creature(g, opp.ID, "Ogre", "Creature — Ogre", 3, 3)
	killCreature(t, g, me.ID, pede)
	passPriorityAroundTable(t, g)
	pickCard(t, g, me.ID, ogre)
	answerLatestTriggerPrompt(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	if got := counterCount(g, ogre, game.CounterMinusOne); got != 1 {
		t.Errorf("-1/-1 counters on the target = %d, want 1", got)
	}
}

func TestSomB1PerilousMyrDealsTwoAsItDies(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := seatsForCombat(g)
	myr := b12Push(g, me.ID, "Perilous Myr", "Artifact Creature — Phyrexian Myr", somB1PerilousOracle, 1, 1)
	life := opp.Life
	killCreature(t, g, me.ID, myr)
	passPriorityAroundTable(t, g)
	pick := latestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("no target was asked for")
	}
	if err := g.ResolvePickTarget(pick.ID, me.ID, playerTarget(opp.ID)); err != nil {
		t.Fatalf("ResolvePickTarget: %v", err)
	}
	passPriorityAroundTable(t, g)
	if opp.Life != life-2 {
		t.Errorf("life %d → %d, want 2 damage", life, opp.Life)
	}
}

// --- Kemba, Kha Regent ---------------------------------------------------------------------------

func TestSomB1KembaMakesACatPerEquipment(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[1]
	kemba := b12Push(g, me.ID, "Kemba, Kha Regent", "Legendary Creature — Cat Cleric", somB1KembaOracle, 2, 4)
	for _, name := range []string{"Sword A", "Sword B"} {
		eq := b12Permanent(g, me.ID, name, "Artifact — Equipment")
		attachForTest(t, g, eq, kemba)
	}
	advanceToUpkeepOf(t, g, 1)
	passPriorityAroundTable(t, g)
	if got := tokensNamedControlledBy(g, "Cat", me.ID); got != 2 {
		t.Errorf("Cat tokens = %d, want 2", got)
	}
}

// The card's ruling: a Kemba that has left by resolution counts the
// Equipment attached to it as it last existed on the battlefield.
func TestSomB1KembaThatLeftCountsItsLastAttachments(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[1]
	kemba := b12Push(g, me.ID, "Kemba, Kha Regent", "Legendary Creature — Cat Cleric", somB1KembaOracle, 2, 4)
	eq := b12Permanent(g, me.ID, "Sword", "Artifact — Equipment")
	attachForTest(t, g, eq, kemba)
	advanceToUpkeepOf(t, g, 1)
	killCreature(t, g, me.ID, kemba)
	passPriorityAroundTable(t, g)
	if got := tokensNamedControlledBy(g, "Cat", me.ID); got != 1 {
		t.Errorf("Cat tokens = %d, want 1", got)
	}
}

// --- Myr Propagator ------------------------------------------------------------------------------

func TestSomB1MyrPropagatorCopiesItself(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	prop := b12Push(g, me.ID, "Myr Propagator", "Artifact Creature — Myr", somB1PropagatorOracle, 1, 1)
	ravB2MustActivate(t, g, me.ID, prop, 0, game.ActivateAbilityParams{}, "C", "C", "C")
	if got := len(tokensNamed(g, "Myr Propagator")); got != 1 {
		t.Errorf("Myr Propagator tokens = %d, want 1", got)
	}
}

// --- Fume Spitter --------------------------------------------------------------------------------

func TestSomB1FumeSpitterSacrificesForACounter(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := seatsForCombat(g)
	spitter := b12Push(g, me.ID, "Fume Spitter", "Creature — Phyrexian Horror", somB1FumeSpitterOracle, 1, 1)
	ogre := b12Creature(g, opp.ID, "Ogre", "Creature — Ogre", 3, 3)
	ravB2MustActivate(t, g, me.ID, spitter, 0, game.ActivateAbilityParams{Targets: cardRefs(ogre)})
	if g.Battlefield.Contains(spitter) {
		t.Error("Fume Spitter was not sacrificed")
	}
	if got := counterCount(g, ogre, game.CounterMinusOne); got != 1 {
		t.Errorf("-1/-1 counters = %d, want 1", got)
	}
}

// --- Corpse Cur ----------------------------------------------------------------------------------

func TestSomB1CorpseCurReturnsOnlyAnInfectCreatureCard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	plain := pushGraveyardCardTyped(me, "Plain Bear", "Creature — Bear")
	infect := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: infect, Name: "Infect Bear", TypeLine: "Creature — Horror",
		Power: 1, Toughness: 1, Keywords: []string{"infect"}, Owner: me.ID, Controller: me.ID})
	castCatalogSpell(t, g, "Corpse Cur", "Artifact Creature — Phyrexian Dog", somB1CorpseCurOracle, nil)
	passPriorityAroundTable(t, g)
	pick := latestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("no target was asked for")
	}
	if hasID(pick.PickTargetCards, plain) {
		t.Error("a creature card without infect was offered")
	}
	pickCard(t, g, me.ID, infect)
	answerLatestTriggerPrompt(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(infect) {
		t.Error("the infect creature card did not return to hand")
	}
}

// --- Golem's Heart -------------------------------------------------------------------------------

func TestSomB1GolemsHeartGainsOnAnyArtifactSpell(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	b12Push(g, opp.ID, "Golem's Heart", "Artifact", somB1GolemsHeartOracle, 0, 0)
	life := opp.Life
	castCatalogSpell(t, g, "Trinket", "Artifact", "", nil)
	answerLatestTriggerPrompt(t, g, opp.ID, true)
	passPriorityAroundTable(t, g)
	if opp.Life != life+1 {
		t.Errorf("life %d → %d, want +1 when another player casts an artifact", life, opp.Life)
	}
	castCatalogSpell(t, g, "Bear", "Creature — Bear", "", nil)
	passPriorityAroundTable(t, g)
	if latestTriggerPrompt(g, opp.ID) != nil || opp.Life != life+1 {
		t.Error("a nonartifact spell triggered it")
	}
}

// --- Glimmerpoint Stag ---------------------------------------------------------------------------

func TestSomB1GlimmerpointStagFlickersUntilTheEndStep(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := seatsForCombat(g)
	ogre := b12Creature(g, opp.ID, "Ogre", "Creature — Ogre", 3, 3)
	castCatalogSpell(t, g, "Glimmerpoint Stag", "Creature — Elk", somB1StagOracle, nil)
	passPriorityAroundTable(t, g)
	pickCard(t, g, me.ID, ogre)
	passPriorityAroundTable(t, g)
	if b12ZoneOf(g, ogre) != game.ZoneExile {
		t.Fatalf("the target is in %q, want exile", b12ZoneOf(g, ogre))
	}
	advanceTo(t, g, game.StepEnd)
	passPriorityAroundTable(t, g)
	if namedOnBattlefield(g, "Ogre", opp.ID) != 1 {
		t.Error("the exiled card did not return under its owner's control at the end step")
	}
}

// --- Putrefax ------------------------------------------------------------------------------------

func TestSomB1PutrefaxIsSacrificedAtTheEndStep(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	fax := b12Push(g, me.ID, "Putrefax", "Creature — Phyrexian Horror", somB1PutrefaxOracle, 5, 3)
	advanceTo(t, g, game.StepEnd)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(fax) {
		t.Error("Putrefax survived the end step")
	}
}

// --- Spikeshot Elder -----------------------------------------------------------------------------

func TestSomB1SpikeshotElderDealsItsPower(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := seatsForCombat(g)
	elder := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Spikeshot Elder", TypeLine: "Creature — Goblin Shaman", OracleID: somB1SpikeshotOracle,
		Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID, Counters: map[string]int{game.CounterPlusOne: 2},
	})
	life := opp.Life
	ravB2MustActivate(t, g, me.ID, elder, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{playerTarget(opp.ID)}}, "C", "R", "R")
	if opp.Life != life-3 {
		t.Errorf("life %d → %d, want 3 damage from a 3-power Elder", life, opp.Life)
	}
}

// --- Chrome Steed --------------------------------------------------------------------------------

func TestSomB1ChromeSteedGrowsWithMetalcraft(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	steed := b12Push(g, me.ID, "Chrome Steed", "Artifact Creature — Horse", somB1ChromeSteedOracle, 2, 2)
	b12Permanent(g, me.ID, "A", "Artifact")
	b12Permanent(g, g.Seats[1].ID, "Theirs", "Artifact")
	if p, tough := effectivePTOf(t, g, steed); p != 2 || tough != 2 {
		t.Errorf("with two artifacts it is %d/%d, want 2/2", p, tough)
	}
	b := b12Permanent(g, me.ID, "B", "Artifact")
	if p, tough := effectivePTOf(t, g, steed); p != 4 || tough != 4 {
		t.Errorf("with three artifacts it is %d/%d, want 4/4", p, tough)
	}
	killCreature(t, g, me.ID, b)
	if p, tough := effectivePTOf(t, g, steed); p != 2 || tough != 2 {
		t.Errorf("back to two artifacts it is %d/%d, want 2/2", p, tough)
	}
}

// --- Neurok Invisimancer -------------------------------------------------------------------------

func TestSomB1NeurokInvisimancerCantBeBlockedAndLendsIt(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := seatsForCombat(g)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	castCatalogSpell(t, g, "Neurok Invisimancer", "Creature — Human Wizard", somB1InvisimancerOracle, nil)
	passPriorityAroundTable(t, g)
	pickCard(t, g, me.ID, bear)
	passPriorityAroundTable(t, g)
	wall := b12Creature(g, opp.ID, "Wall", "Creature — Wall", 0, 4)
	declareAttack(t, g, opp.ID, bear)
	advanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(wall, bear); err == nil {
		t.Error("the target creature was blocked")
	}
}
