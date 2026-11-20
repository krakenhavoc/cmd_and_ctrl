package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// mor_5592_b1_cards_test.go — Morningtide, batch 1 (#5592). Each test
// plays the real catalog entry through the engine and asserts what its
// text says.

const (
	morB1ChangelingSentinelOracle     = "2977aae2-225d-4f3f-8bf7-535f542a7391"
	morB1GameTrailChangelingOracle    = "384611fb-aca2-4e2d-9125-44ea034de08b"
	morB1PricklyBoggartOracle         = "fa110e05-6759-4eeb-96eb-9b7a8e4a94c3"
	morB1MoongloveChangelingOracle    = "61b74f82-39f3-4618-b8e7-292058421d1b"
	morB1WarSpikeChangelingOracle     = "b6ddb093-1fff-4648-b7a1-141a3875599c"
	morB1FencerCliqueOracle           = "58374ee4-0497-4d12-847c-0f67555f8399"
	morB1BlightsoilDruidOracle        = "351091cb-1263-4202-9447-021c6019cff1"
	morB1MosquitoGuardOracle          = "6beaeec2-b4e9-4904-b817-c0118549d4a9"
	morB1BurrentonBombardierOracle    = "13ec2390-e08a-4645-99e9-bac7e41917bb"
	morB1EarthbrawnOracle             = "91edf44a-0f46-4936-909a-934e2c4fe176"
	morB1HuntingTriadOracle           = "66bf0c4b-0c36-4d63-a929-5a7427a07279"
	morB1MerrowWitsniperOracle        = "79f8378a-fb50-4178-be53-b6926bc1cba5"
	morB1FinalStingFaerieOracle       = "299765d8-72b9-4ada-a607-efb36aa0f5db"
	morB1MindShatterOracle            = "26809033-a823-4251-887e-0410a545f993"
	morB1VioletPallOracle             = "bf6ea336-6a4d-4653-acdd-f09af0b9f0e3"
	morB1StreamOfUnconsciousnessOracl = "7da64f81-64d0-4019-a208-9ee8b6782d85"
	morB1ShardVolleyOracle            = "3d19ab38-fb32-4f2b-b68c-28db4a6864c6"
	morB1VengefulFirebrandOracle      = "920e21c4-0116-416e-a375-bb887cb7b44a"
	morB1MothdustChangelingOracle     = "52237e6b-19f6-405c-a668-545c4ccb1b6d"
	morB1StingmoggieOracle            = "e868028f-4c6f-43d1-9cc6-6f3ce2df2ec3"
	morB1ShinewendOracle              = "8b3c4a6c-44ac-4718-ac1e-9d503523e89d"
	morB1FestercreepOracle            = "0f7f841a-1913-4396-9c94-bbd49dfdb9d1"
	morB1ShieldBearersOracle          = "8d038c81-b43f-4721-abda-5dacfc194ed3"
	morB1WeedPrunerPoplarOracle       = "40e8ade9-9cec-4983-ab11-d9e6278fb29f"
	morB1SeethingPathblazerOracle     = "574abab1-a722-45f5-a210-c1d4442a278a"
	morB1ScarbladeEliteOracle         = "e5fbd5c8-7b0c-4740-9430-456e86267966"
	morB1SpitebellowsOracle           = "fee18f57-4ae5-4df8-af3b-f82ed2157ac3"
	morB1NevermakerOracle             = "38a2ef3e-3311-4e0a-807d-e5ccdd05af49"
	morB1OffalsnoutOracle             = "f0a20c62-035d-408a-9385-8f2bc2a6d332"
	morB1WalkerOfTheGroveOracle       = "02d69413-e040-46b1-8f58-69bde666329d"
	morB1LatchkeyFaerieOracle         = "7bc4a0a0-f06e-4074-95a8-67a5d8fdfef9"
	morB1MorselTheftOracle            = "b58878ce-dd5a-4c43-882c-639de4f4a111"
	morB1ThievesFortuneOracle         = "e0bffcf3-8d34-461b-a1e4-ae47c1e7b19e"
	morB1WeirdingShamanOracle         = "ed8e0c38-64b1-48a5-9623-935d59ffaacc"
	morB1StonybrookSchoolmasterOracle = "7ef67bae-4af3-47b9-95d2-08f3c5a0b39d"
	morB1FeudkillersVerdictOracle     = "4dacd89b-5983-4905-9f6c-43aa6a12149d"
	morB1RhysTheExiledOracle          = "7613c50b-4568-4340-b64a-ab427aaffd51"
	morB1BorderlandBehemothOracle     = "378a78d8-9001-4aa5-a66b-2aca27fa8180"
	morB1StenchskipperOracle          = "0817d902-08c8-4ec3-9a84-2de65bc70e6a"
	morB1DailyRegimenOracle           = "3a869ec0-c3ea-49a8-8586-d7e54e62a734"
	morB1LunkErrantOracle             = "1c2445b4-e8be-452c-950d-f42e549daf9a"
	morB1GoldenCricketOracle          = "02abe85a-c603-4963-aba4-524cc80847ea"

	morB1BearType      = "Creature — Bear"
	morB1ShifterType   = "Creature — Shapeshifter"
	morB1ElementalType = "Creature — Elemental"
	morB1RogueType     = "Creature — Human Rogue"
)

func TestMorB1EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		morB1ChangelingSentinelOracle:     "Changeling Sentinel",
		morB1GameTrailChangelingOracle:    "Game-Trail Changeling",
		morB1PricklyBoggartOracle:         "Prickly Boggart",
		morB1MoongloveChangelingOracle:    "Moonglove Changeling",
		morB1WarSpikeChangelingOracle:     "War-Spike Changeling",
		morB1FencerCliqueOracle:           "Fencer Clique",
		morB1BlightsoilDruidOracle:        "Blightsoil Druid",
		morB1MosquitoGuardOracle:          "Mosquito Guard",
		morB1BurrentonBombardierOracle:    "Burrenton Bombardier",
		morB1EarthbrawnOracle:             "Earthbrawn",
		morB1HuntingTriadOracle:           "Hunting Triad",
		morB1MerrowWitsniperOracle:        "Merrow Witsniper",
		morB1FinalStingFaerieOracle:       "Final-Sting Faerie",
		morB1MindShatterOracle:            "Mind Shatter",
		morB1VioletPallOracle:             "Violet Pall",
		morB1StreamOfUnconsciousnessOracl: "Stream of Unconsciousness",
		morB1ShardVolleyOracle:            "Shard Volley",
		morB1VengefulFirebrandOracle:      "Vengeful Firebrand",
		morB1MothdustChangelingOracle:     "Mothdust Changeling",
		morB1StingmoggieOracle:            "Stingmoggie",
		morB1ShinewendOracle:              "Shinewend",
		morB1FestercreepOracle:            "Festercreep",
		morB1ShieldBearersOracle:          "Burrenton Shield-Bearers",
		morB1WeedPrunerPoplarOracle:       "Weed-Pruner Poplar",
		morB1SeethingPathblazerOracle:     "Seething Pathblazer",
		morB1ScarbladeEliteOracle:         "Scarblade Elite",
		morB1SpitebellowsOracle:           "Spitebellows",
		morB1NevermakerOracle:             "Nevermaker",
		morB1OffalsnoutOracle:             "Offalsnout",
		morB1WalkerOfTheGroveOracle:       "Walker of the Grove",
		morB1LatchkeyFaerieOracle:         "Latchkey Faerie",
		morB1MorselTheftOracle:            "Morsel Theft",
		morB1ThievesFortuneOracle:         "Thieves' Fortune",
		morB1WeirdingShamanOracle:         "Weirding Shaman",
		morB1StonybrookSchoolmasterOracle: "Stonybrook Schoolmaster",
		morB1FeudkillersVerdictOracle:     "Feudkiller's Verdict",
		morB1RhysTheExiledOracle:          "Rhys the Exiled",
		morB1BorderlandBehemothOracle:     "Borderland Behemoth",
		morB1StenchskipperOracle:          "Stenchskipper",
		morB1DailyRegimenOracle:           "Daily Regimen",
		morB1LunkErrantOracle:             "Lunk Errant",
		morB1GoldenCricketOracle:          "Order of the Golden Cricket",
	} {
		spec, ok := Lookup(oracle)
		if !ok {
			t.Errorf("%s (%s) is not in the catalog", name, oracle)
			continue
		}
		if spec.Name != name {
			t.Errorf("%s is registered as %q", name, spec.Name)
		}
		if spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: completeness %v with %d caveats, want Full with none", name, spec.Completeness, len(spec.Caveats))
		}
	}
}

// --- helpers ------------------------------------------------------------------

// morB1Opp is the seat after the active one.
func morB1Opp(g *game.Game) *game.Player {
	return g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
}

// morB1PT is a battlefield permanent's power and toughness after the layers.
func morB1PT(t *testing.T, g *game.Game, id uuid.UUID) (int, int) {
	t.Helper()
	g.WithWriteLock(func() { g.RecomputeLayersIfStaleLocked() })
	c := findBattlefieldCardByID(g, id)
	if c == nil {
		t.Fatalf("%s is not on the battlefield", id)
	}
	eff := c.Effective()
	return eff.Power, eff.Toughness
}

// morB1Tokens counts the tokens named `name` that `owner` controls.
func morB1Tokens(g *game.Game, owner uuid.UUID, name string) int {
	n := 0
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == owner && IsToken(c) && c.Name == name {
			n++
		}
	}
	return n
}

// morB1Activate activates row `row` of a catalog permanent and settles
// the stack.
func morB1Activate(t *testing.T, g *game.Game, who *game.Player, id uuid.UUID, row int, params game.ActivateAbilityParams) {
	t.Helper()
	if err := g.ActivateCatalogAbility(who.ID, id, row, params); err != nil {
		t.Fatalf("activate row %d: %v", row, err)
	}
	passPriorityAroundTable(t, g)
}

// morB1CastAndFind casts a catalog creature spell from the active seat's
// hand, settles the stack and returns the permanent it became.
func morB1CastAndFind(t *testing.T, g *game.Game, name, typeLine, oracle string) uuid.UUID {
	t.Helper()
	castCatalogSpell(t, g, name, typeLine, oracle, nil)
	passPriorityAroundTable(t, g)
	for _, c := range g.Battlefield.Cards {
		if c.OracleID == oracle {
			return c.InstanceID
		}
	}
	t.Fatalf("%s is not on the battlefield", name)
	return uuid.Nil
}

// morB1ConnectWithARogue has a Rogue the active seat controls deal combat
// damage to the next seat, then walks to the postcombat main phase, so
// prowl is open.
func morB1ConnectWithARogue(t *testing.T, g *game.Game) {
	t.Helper()
	me := g.Seats[g.Turn.ActiveSeat]
	rogue := b12Creature(g, me.ID, "Rogue", morB1RogueType, 2, 2)
	attackWith(t, g, morB1Opp(g).ID, rogue)
	passPriorityAroundTable(t, g)
	advanceTo(t, g, game.StepPostcombatMain)
}

// --- keywords and changelings -------------------------------------------------

func TestMorB1PrintedKeywords(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, oracle, keyword string
	}{
		{"Changeling Sentinel", morB1ShifterType, morB1ChangelingSentinelOracle, "vigilance"},
		{"Game-Trail Changeling", morB1ShifterType, morB1GameTrailChangelingOracle, "trample"},
		{"Prickly Boggart", "Creature — Goblin Rogue", morB1PricklyBoggartOracle, "fear"},
		{"Fencer Clique", "Creature — Faerie Soldier", morB1FencerCliqueOracle, "flying"},
		{"Mosquito Guard", "Creature — Kithkin Soldier", morB1MosquitoGuardOracle, "first strike"},
		{"Burrenton Bombardier", "Creature — Kithkin Soldier", morB1BurrentonBombardierOracle, "flying"},
		{"Final-Sting Faerie", "Creature — Faerie Assassin", morB1FinalStingFaerieOracle, "flying"},
		{"Nevermaker", morB1ElementalType, morB1NevermakerOracle, "flying"},
		{"Offalsnout", morB1ElementalType, morB1OffalsnoutOracle, "flash"},
		{"Shinewend", morB1ElementalType, morB1ShinewendOracle, "flying"},
		{"Latchkey Faerie", "Creature — Faerie Rogue", morB1LatchkeyFaerieOracle, "flying"},
		{"Borderland Behemoth", "Creature — Giant Warrior", morB1BorderlandBehemothOracle, "trample"},
		{"Stenchskipper", morB1ElementalType, morB1StenchskipperOracle, "flying"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			id := b12Push(g, g.Seats[0].ID, tc.name, tc.typeLine, tc.oracle, 1, 1)
			if !hasAbility(effectiveAbilities(t, g, id), tc.keyword) {
				t.Errorf("%s has no %s", tc.name, tc.keyword)
			}
		})
	}
}

func TestMorB1ChangelingsAreEveryCreatureType(t *testing.T) {
	for _, tc := range []struct{ name, oracle string }{
		{"Changeling Sentinel", morB1ChangelingSentinelOracle},
		{"Game-Trail Changeling", morB1GameTrailChangelingOracle},
		{"Moonglove Changeling", morB1MoongloveChangelingOracle},
		{"War-Spike Changeling", morB1WarSpikeChangelingOracle},
		{"Mothdust Changeling", morB1MothdustChangelingOracle},
	} {
		g := newCatalogGame(t)
		id := b12Push(g, g.Seats[0].ID, tc.name, morB1ShifterType, tc.oracle, 2, 2)
		if !lrwB1HasType(t, g, id, "Kithkin") || !lrwB1HasType(t, g, id, "Giant") {
			t.Errorf("%s is not every creature type", tc.name)
		}
	}
}

// --- activated abilities ------------------------------------------------------

func TestMorB1ChangelingsGainTheirKeyword(t *testing.T) {
	for _, tc := range []struct{ name, oracle, keyword string }{
		{"Moonglove Changeling", morB1MoongloveChangelingOracle, "deathtouch"},
		{"War-Spike Changeling", morB1WarSpikeChangelingOracle, "first strike"},
	} {
		g := newCatalogGame(t)
		advanceToMain(t, g)
		me := g.Seats[g.Turn.ActiveSeat]
		id := b12Push(g, me.ID, tc.name, morB1ShifterType, tc.oracle, 2, 2)
		if hasKeywordOnBattlefield(t, g, id, tc.keyword) {
			t.Fatalf("%s has %s before activating", tc.name, tc.keyword)
		}
		morB1Activate(t, g, me, id, 0, game.ActivateAbilityParams{})
		if !hasKeywordOnBattlefield(t, g, id, tc.keyword) {
			t.Errorf("%s did not gain %s", tc.name, tc.keyword)
		}
	}
}

func TestMorB1MothdustChangelingTapsACreatureToFly(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	moth := b12Push(g, me.ID, "Mothdust Changeling", morB1ShifterType, morB1MothdustChangelingOracle, 1, 1)
	bear := b12Creature(g, me.ID, "Bear", morB1BearType, 2, 2)
	morB1Activate(t, g, me, moth, 0, game.ActivateAbilityParams{TapIDs: []uuid.UUID{bear}})
	if tapped, _ := battlefieldCardTapped(g, bear); !tapped {
		t.Error("the Bear was not tapped to pay the cost")
	}
	if !hasKeywordOnBattlefield(t, g, moth, "flying") {
		t.Error("Mothdust Changeling did not gain flying")
	}
}

func TestMorB1FencerCliqueGoesOnTopOfItsLibrary(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	clique := b12Push(g, me.ID, "Fencer Clique", "Creature — Faerie Soldier", morB1FencerCliqueOracle, 3, 2)
	morB1Activate(t, g, me, clique, 0, game.ActivateAbilityParams{})
	top, err := me.Library.Top()
	if err != nil || top.InstanceID != clique {
		t.Error("Fencer Clique is not on top of its owner's library")
	}
}

func TestMorB1BlightsoilDruidPaysLifeForGreen(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	druid := b12Push(g, me.ID, "Blightsoil Druid", "Creature — Elf Druid", morB1BlightsoilDruidOracle, 1, 2)
	life := me.Life
	if err := g.ActivateManaAbility(me.ID, druid, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("mana ability: %v", err)
	}
	if me.ManaPool.CountColor("G") != 1 {
		t.Errorf("pool %v, want {G}", me.ManaPool)
	}
	if me.Life != life-1 {
		t.Errorf("life %d, want %d", me.Life, life-1)
	}
}

func TestMorB1VengefulFirebrandHasteNeedsAWarriorCard(t *testing.T) {
	for _, warrior := range []bool{false, true} {
		g := newCatalogGame(t)
		advanceToMain(t, g)
		me := g.Seats[g.Turn.ActiveSeat]
		if warrior {
			pushGraveyardCardTyped(me, "Dead Warrior", "Creature — Human Warrior")
		} else {
			pushGraveyardCardTyped(me, "Dead Bear", morB1BearType)
		}
		brand := b12Push(g, me.ID, "Vengeful Firebrand", "Creature — Elemental Warrior", morB1VengefulFirebrandOracle, 5, 2)
		if got := hasKeywordOnBattlefield(t, g, brand, "haste"); got != warrior {
			t.Errorf("Warrior card in graveyard = %v: haste = %v", warrior, got)
		}
		if warrior {
			morB1Activate(t, g, me, brand, 0, game.ActivateAbilityParams{})
			if p, _ := morB1PT(t, g, brand); p != 6 {
				t.Errorf("power %d after {R}, want 6", p)
			}
		}
	}
}

func TestMorB1SeethingPathblazerSacrificesAnElemental(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	blazer := b12Push(g, me.ID, "Seething Pathblazer", "Creature — Elemental Warrior", morB1SeethingPathblazerOracle, 2, 2)
	fodder := b12Creature(g, me.ID, "Fodder", morB1ElementalType, 1, 1)
	morB1Activate(t, g, me, blazer, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{fodder}})
	if g.Battlefield.Contains(fodder) {
		t.Error("the Elemental was not sacrificed")
	}
	if p, _ := morB1PT(t, g, blazer); p != 4 {
		t.Errorf("power %d, want 4", p)
	}
	if !hasKeywordOnBattlefield(t, g, blazer, "first strike") {
		t.Error("no first strike")
	}
}

func TestMorB1ScarbladeEliteExilesAnAssassinToDestroy(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me, opp := g.Seats[g.Turn.ActiveSeat], morB1Opp(g)
	elite := b12Push(g, me.ID, "Scarblade Elite", "Creature — Elf Assassin", morB1ScarbladeEliteOracle, 2, 2)
	bear := pushGraveyardCardTyped(me, "Dead Bear", morB1BearType)
	assassin := pushGraveyardCardTyped(me, "Dead Assassin", "Creature — Human Assassin")
	victim := b12Creature(g, opp.ID, "Victim", morB1BearType, 2, 2)
	if err := g.ActivateCatalogAbility(me.ID, elite, 0, game.ActivateAbilityParams{
		ExileIDs: []uuid.UUID{bear}, Targets: cardTarget(victim),
	}); err == nil {
		t.Fatal("a non-Assassin card paid the cost")
	}
	morB1Activate(t, g, me, elite, 0, game.ActivateAbilityParams{ExileIDs: []uuid.UUID{assassin}, Targets: cardTarget(victim)})
	if !g.Exile.Contains(assassin) {
		t.Error("the Assassin card was not exiled")
	}
	if g.Battlefield.Contains(victim) {
		t.Error("the target was not destroyed")
	}
}

func TestMorB1WeirdingShamanMakesTwoGoblinRogues(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	shaman := b12Push(g, me.ID, "Weirding Shaman", "Creature — Goblin Shaman", morB1WeirdingShamanOracle, 2, 1)
	goblin := b12Creature(g, me.ID, "Goblin", "Creature — Goblin", 1, 1)
	morB1Activate(t, g, me, shaman, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{goblin}})
	if g.Battlefield.Contains(goblin) {
		t.Error("the Goblin was not sacrificed")
	}
	if got := morB1Tokens(g, me.ID, "Goblin Rogue"); got != 2 {
		t.Errorf("%d Goblin Rogue tokens, want 2", got)
	}
}

func TestMorB1DailyRegimenCountersTheEnchantedCreature(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := b12Creature(g, me.ID, "Bear", morB1BearType, 2, 2)
	aura := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Daily Regimen", TypeLine: "Enchantment — Aura", OracleID: morB1DailyRegimenOracle,
		Owner: me.ID, Controller: me.ID, AttachedTo: game.TargetRef{Kind: game.TargetCard, ID: bear},
	})
	morB1Activate(t, g, me, aura, 0, game.ActivateAbilityParams{})
	if got := countersOn(g, bear, game.CounterPlusOne); got != 1 {
		t.Errorf("%d +1/+1 counters on the enchanted creature, want 1", got)
	}
}

// --- reinforce -----------------------------------------------------------------

func TestMorB1ReinforceFromTheHand(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, oracle string
		mana                   []string
		n                      int
	}{
		{"Mosquito Guard", "Creature — Kithkin Soldier", morB1MosquitoGuardOracle, []string{"W", "C"}, 1},
		{"Burrenton Bombardier", "Creature — Kithkin Soldier", morB1BurrentonBombardierOracle, []string{"W", "C", "C"}, 2},
		{"Earthbrawn", "Instant", morB1EarthbrawnOracle, []string{"G", "C"}, 1},
		{"Hunting Triad", "Kindred Sorcery — Elf", morB1HuntingTriadOracle, []string{"G", "C", "C", "C"}, 3},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			advanceToMain(t, g)
			bear := b12Creature(g, me.ID, "Bear", morB1BearType, 2, 2)
			card := b13HandCard(me, tc.name, tc.typeLine, tc.oracle)
			apaMana(me, tc.mana...)
			apaActivate(t, g, me, card, 0, game.ActivateAbilityParams{Targets: cardTarget(bear)})
			if got := countersOn(g, bear, game.CounterPlusOne); got != tc.n {
				t.Errorf("%d +1/+1 counters, want %d", got, tc.n)
			}
			if b12ZoneOf(g, card) != game.ZoneGraveyard {
				t.Error("the card was not discarded to pay for reinforce")
			}
		})
	}
}

// --- spells --------------------------------------------------------------------

func TestMorB1EarthbrawnPumpsTheTarget(t *testing.T) {
	g := newCatalogGame(t)
	bear := b12Creature(g, g.Seats[0].ID, "Bear", morB1BearType, 2, 2)
	castCatalogSpell(t, g, "Earthbrawn", "Instant", morB1EarthbrawnOracle, cardTarget(bear))
	passPriorityAroundTable(t, g)
	if p, tough := morB1PT(t, g, bear); p != 5 || tough != 5 {
		t.Errorf("Bear %d/%d, want 5/5", p, tough)
	}
}

func TestMorB1HuntingTriadMakesThreeElfWarriors(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	castCatalogSpell(t, g, "Hunting Triad", "Kindred Sorcery — Elf", morB1HuntingTriadOracle, nil)
	passPriorityAroundTable(t, g)
	if got := morB1Tokens(g, me.ID, "Elf Warrior"); got != 3 {
		t.Errorf("%d Elf Warrior tokens, want 3", got)
	}
}

func TestMorB1MindShatterDiscardsXAtRandom(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], morB1Opp(g)
	for _, n := range []string{"A", "B", "C", "D"} {
		opp.Hand.PushTop(game.Card{InstanceID: uuid.New(), Name: n, TypeLine: "Instant", Owner: opp.ID, Controller: opp.ID})
	}
	hand := opp.Hand.Size()
	advanceToMain(t, g)
	id := b13HandCard(me, "Mind Shatter", "Sorcery", morB1MindShatterOracle)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{XValue: 2, Targets: []game.TargetRef{playerTarget(opp.ID)}}); err != nil {
		t.Fatalf("cast Mind Shatter: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := opp.Hand.Size(); got != hand-2 {
		t.Errorf("hand %d, want %d", got, hand-2)
	}
}

func TestMorB1VioletPallDestroysANonblackCreatureAndMakesAFaerie(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], morB1Opp(g)
	bear := b12Creature(g, opp.ID, "Bear", morB1BearType, 2, 2)
	hmlSetColors(g, bear, "G")
	zombie := b12Creature(g, opp.ID, "Zombie", "Creature — Zombie", 2, 2)
	hmlSetColors(g, zombie, "B")
	if err := castCatalogSpellErr(t, g, "Violet Pall", "Kindred Instant — Faerie", morB1VioletPallOracle, cardTarget(zombie)); err == nil {
		t.Fatal("a black creature was accepted as the target")
	}
	castCatalogSpell(t, g, "Violet Pall", "Kindred Instant — Faerie", morB1VioletPallOracle, cardTarget(bear))
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) {
		t.Error("the nonblack creature survived")
	}
	if got := morB1Tokens(g, me.ID, "Faerie Rogue"); got != 1 {
		t.Errorf("%d Faerie Rogue tokens, want 1", got)
	}
}

func TestMorB1StreamOfUnconsciousnessDrawsWithAWizard(t *testing.T) {
	for _, wizard := range []bool{false, true} {
		g := newCatalogGame(t)
		advanceToMain(t, g)
		me, opp := g.Seats[g.Turn.ActiveSeat], morB1Opp(g)
		bear := b12Creature(g, opp.ID, "Bear", morB1BearType, 3, 3)
		if wizard {
			b12Creature(g, me.ID, "Wizard", "Creature — Merfolk Wizard", 1, 1)
		}
		hand := me.Hand.Size()
		castCatalogSpell(t, g, "Stream of Unconsciousness", "Kindred Instant — Wizard", morB1StreamOfUnconsciousnessOracl, cardTarget(bear))
		passPriorityAroundTable(t, g)
		if p, _ := morB1PT(t, g, bear); p != -1 {
			t.Errorf("wizard=%v: power %d, want -1", wizard, p)
		}
		want := hand
		if wizard {
			want++
		}
		if got := me.Hand.Size(); got != want {
			t.Errorf("wizard=%v: hand %d, want %d", wizard, got, want)
		}
	}
}

func TestMorB1ShardVolleySacrificesALand(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me, opp := g.Seats[g.Turn.ActiveSeat], morB1Opp(g)
	land := seedLandOnBattlefield(g, me.ID, "Mountain", "Basic Land — Mountain")
	life := opp.Life
	id := b13HandCard(me, "Shard Volley", "Instant", morB1ShardVolleyOracle)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{
		SacrificeIDs: []uuid.UUID{land}, Targets: []game.TargetRef{playerTarget(opp.ID)},
	}); err != nil {
		t.Fatalf("cast Shard Volley: %v", err)
	}
	if g.Battlefield.Contains(land) {
		t.Error("the land was not sacrificed as the spell was cast")
	}
	passPriorityAroundTable(t, g)
	if opp.Life != life-3 {
		t.Errorf("life %d, want %d", opp.Life, life-3)
	}
}

func TestMorB1FeudkillersVerdictMakesAGiantWhenAhead(t *testing.T) {
	for _, ahead := range []bool{true, false} {
		g := newCatalogGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		life := me.Life
		if !ahead {
			for _, p := range g.Seats {
				if p.ID != me.ID {
					p.Life = life + 20
				}
			}
		}
		castCatalogSpell(t, g, "Feudkiller's Verdict", "Kindred Sorcery — Giant", morB1FeudkillersVerdictOracle, nil)
		passPriorityAroundTable(t, g)
		if me.Life != life+10 {
			t.Errorf("ahead=%v: life %d, want %d", ahead, me.Life, life+10)
		}
		want := 0
		if ahead {
			want = 1
		}
		if got := morB1Tokens(g, me.ID, "Giant Warrior"); got != want {
			t.Errorf("ahead=%v: %d Giant Warrior tokens, want %d", ahead, got, want)
		}
	}
}

func TestMorB1ThievesFortuneTakesOneOfFour(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	ids := seedLibrary(me, "A", "B", "C", "D")
	castCatalogSpell(t, g, "Thieves' Fortune", "Kindred Instant — Rogue", morB1ThievesFortuneOracle, nil)
	passPriorityAroundTable(t, g)
	pick := latestChooseCardsFor(g, me.ID)
	if pick == nil {
		t.Fatal("no take prompt")
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{ids[1]}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	if !me.Hand.Contains(ids[1]) {
		t.Error("the picked card is not in hand")
	}
	if order := putInLibraryChoiceFor(g, me.ID); order != nil {
		if err := g.ResolvePutInLibrary(order.ID, me.ID, []uuid.UUID{ids[0], ids[2], ids[3]}, nil); err != nil {
			t.Fatalf("ResolvePutInLibrary: %v", err)
		}
	}
	if got := libraryBottomIDs(me, 3); len(got) != 3 || me.Hand.Contains(ids[0]) {
		t.Error("the other three did not go to the bottom")
	}
}

// --- prowl ---------------------------------------------------------------------

func TestMorB1LatchkeyFaerieDrawsOnlyWhenProwled(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	morB1ConnectWithARogue(t, g)
	id := uuid.New()
	me.Hand.PushTop(game.Card{
		InstanceID: id, Name: "Latchkey Faerie", TypeLine: "Creature — Faerie Rogue", ManaCost: "{3}{U}",
		OracleID: morB1LatchkeyFaerieOracle, Colors: []string{"U"}, Power: 3, Toughness: 1, Owner: me.ID, Controller: me.ID,
	})
	hand := me.Hand.Size()
	floatForTest(g, me, "CCU")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, AlternativeCost: "prowl"}); err != nil {
		t.Fatalf("prowl was refused: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(id) {
		t.Fatal("Latchkey Faerie did not resolve")
	}
	if got := me.Hand.Size(); got != hand {
		t.Errorf("hand %d, want %d: the Faerie left and a card was drawn", got, hand)
	}

	// Hard-cast, it draws nothing.
	g = newCatalogGame(t)
	me = g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	hand = me.Hand.Size()
	castHoldCreature(t, g, "Latchkey Faerie", "Creature — Faerie Rogue", morB1LatchkeyFaerieOracle, 3, 1)
	if got := me.Hand.Size(); got != hand {
		t.Errorf("hard-cast: hand %d, want %d", got, hand)
	}
}

func TestMorB1MorselTheftDrainsAndDrawsWhenProwled(t *testing.T) {
	for _, prowl := range []bool{false, true} {
		g := newCatalogGame(t)
		me, opp := g.Seats[g.Turn.ActiveSeat], morB1Opp(g)
		if prowl {
			morB1ConnectWithARogue(t, g)
		} else {
			advanceToMain(t, g)
		}
		myLife, theirLife := me.Life, opp.Life
		id := b13HandCard(me, "Morsel Theft", "Kindred Sorcery — Rogue", morB1MorselTheftOracle)
		hand := me.Hand.Size()
		params := game.CastSpellParams{Targets: []game.TargetRef{playerTarget(opp.ID)}}
		if prowl {
			floatForTest(g, me, "CB")
			params.Strict = true
			params.AlternativeCost = "prowl"
		}
		if err := g.CastSpell(me.ID, id, params); err != nil {
			t.Fatalf("prowl=%v: cast: %v", prowl, err)
		}
		passPriorityAroundTable(t, g)
		if opp.Life != theirLife-3 || me.Life != myLife+3 {
			t.Errorf("prowl=%v: lives %d and %d, want %d and %d", prowl, opp.Life, me.Life, theirLife-3, myLife+3)
		}
		want := hand - 1
		if prowl {
			want = hand
		}
		if got := me.Hand.Size(); got != want {
			t.Errorf("prowl=%v: hand %d, want %d", prowl, got, want)
		}
	}
}

// --- enters, leaves and evoke --------------------------------------------------

func TestMorB1MerrowWitsniperMillsTargetPlayer(t *testing.T) {
	g := newCatalogGame(t)
	opp := morB1Opp(g)
	top, err := opp.Library.Top()
	if err != nil {
		t.Fatal(err)
	}
	castHoldCreature(t, g, "Merrow Witsniper", "Creature — Merfolk Rogue", morB1MerrowWitsniperOracle, 1, 1)
	ch := pickTargetPrompt(t, g)
	if err := g.ResolvePickTarget(ch.ID, ch.Chooser, playerTarget(opp.ID)); err != nil {
		t.Fatalf("ResolvePickTarget: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !opp.Graveyard.Contains(top.InstanceID) {
		t.Error("the top card of the target's library was not milled")
	}
}

func TestMorB1FinalStingFaerieDestroysADamagedCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], morB1Opp(g)
	pinger := b12Creature(g, me.ID, "Pinger", morB1BearType, 1, 1)
	hurt := b12Creature(g, opp.ID, "Hurt", morB1BearType, 3, 3)
	fresh := b12Creature(g, opp.ID, "Fresh", morB1BearType, 3, 3)
	b27Damage(g, pinger, hurt, 1)
	g.RunStateChecksForTest()
	castHoldCreature(t, g, "Final-Sting Faerie", "Creature — Faerie Assassin", morB1FinalStingFaerieOracle, 2, 2)
	ch := pickTargetPrompt(t, g)
	if err := g.ResolvePickTarget(ch.ID, ch.Chooser, game.TargetRef{Kind: game.TargetCard, ID: fresh}); err == nil {
		t.Fatal("an undamaged creature was accepted as the target")
	}
	answerPickTarget(t, g, hurt)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(hurt) {
		t.Error("the damaged creature survived")
	}
	if !g.Battlefield.Contains(fresh) {
		t.Error("the undamaged creature was destroyed")
	}
}

func TestMorB1SpitebellowsDealsSixWhenItLeaves(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	spite := b12Push(g, me.ID, "Spitebellows", morB1ElementalType, morB1SpitebellowsOracle, 6, 1)
	ogre := b12Creature(g, opp.ID, "Ogre", "Creature — Ogre", 5, 5)
	b27Kill(g, spite)
	g.RunStateChecksForTest()
	answerPickTarget(t, g, ogre)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(ogre) {
		t.Error("the target survived 6 damage")
	}
}

func TestMorB1NevermakerTucksANonlandPermanent(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	maker := b12Push(g, me.ID, "Nevermaker", morB1ElementalType, morB1NevermakerOracle, 2, 3)
	rock := b12Permanent(g, opp.ID, "Rock", "Artifact")
	b27Kill(g, maker)
	g.RunStateChecksForTest()
	answerPickTarget(t, g, rock)
	passPriorityAroundTable(t, g)
	top, err := opp.Library.Top()
	if err != nil || top.InstanceID != rock {
		t.Error("the artifact is not on top of its owner's library")
	}
}

func TestMorB1OffalsnoutExilesAGraveyardCardWhenItLeaves(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	snout := b12Push(g, me.ID, "Offalsnout", morB1ElementalType, morB1OffalsnoutOracle, 2, 2)
	dead := pushGraveyardCardTyped(opp, "Dead Bear", morB1BearType)
	b27Kill(g, snout)
	g.RunStateChecksForTest()
	answerPickTarget(t, g, dead)
	passPriorityAroundTable(t, g)
	if !g.Exile.Contains(dead) {
		t.Error("the graveyard card was not exiled")
	}
}

func TestMorB1WalkerOfTheGroveEvokedLeavesAnElemental(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	me.Hand.PushTop(game.Card{
		InstanceID: id, Name: "Walker of the Grove", TypeLine: morB1ElementalType, OracleID: morB1WalkerOfTheGroveOracle,
		Power: 7, Toughness: 7, Owner: me.ID, Controller: me.ID,
	})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{AlternativeCost: "evoke"}); err != nil {
		t.Fatalf("evoke: %v", err)
	}
	for i := 0; i < 4; i++ {
		passPriorityAroundTable(t, g)
	}
	if g.Battlefield.Contains(id) {
		t.Error("the evoked Walker stayed on the battlefield")
	}
	if got := morB1Tokens(g, me.ID, "Elemental"); got != 1 {
		t.Errorf("%d Elemental tokens, want 1", got)
	}
}

// --- counters -------------------------------------------------------------------

func TestMorB1StingmoggieDestroysAnArtifactOrLand(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], morB1Opp(g)
	moggie := morB1CastAndFind(t, g, "Stingmoggie", morB1ElementalType, morB1StingmoggieOracle)
	if got := countersOn(g, moggie, game.CounterPlusOne); got != 2 {
		t.Fatalf("enters with %d counters, want 2", got)
	}
	land := b12Permanent(g, opp.ID, "Their Land", "Land")
	morB1Activate(t, g, me, moggie, 0, game.ActivateAbilityParams{Targets: cardTarget(land)})
	if g.Battlefield.Contains(land) {
		t.Error("the land was not destroyed")
	}
	if got := countersOn(g, moggie, game.CounterPlusOne); got != 1 {
		t.Errorf("%d counters left, want 1", got)
	}
}

func TestMorB1ShinewendDestroysAnEnchantment(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], morB1Opp(g)
	wend := morB1CastAndFind(t, g, "Shinewend", morB1ElementalType, morB1ShinewendOracle)
	if got := countersOn(g, wend, game.CounterPlusOne); got != 1 {
		t.Fatalf("enters with %d counters, want 1", got)
	}
	ench := b12Permanent(g, opp.ID, "Their Enchantment", "Enchantment")
	morB1Activate(t, g, me, wend, 0, game.ActivateAbilityParams{Targets: cardTarget(ench)})
	if g.Battlefield.Contains(ench) {
		t.Error("the enchantment was not destroyed")
	}
}

func TestMorB1FestercreepShrinksEveryOtherCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], morB1Opp(g)
	creep := morB1CastAndFind(t, g, "Festercreep", morB1ElementalType, morB1FestercreepOracle)
	if got := countersOn(g, creep, game.CounterPlusOne); got != 1 {
		t.Fatalf("enters with %d counters, want 1", got)
	}
	mine := b12Creature(g, me.ID, "Mine", morB1BearType, 2, 2)
	theirs := b12Creature(g, opp.ID, "Theirs", morB1BearType, 2, 2)
	morB1Activate(t, g, me, creep, 0, game.ActivateAbilityParams{})
	for _, id := range []uuid.UUID{mine, theirs} {
		if p, tough := morB1PT(t, g, id); p != 1 || tough != 1 {
			t.Errorf("%d/%d, want 1/1", p, tough)
		}
	}
}

// --- combat ----------------------------------------------------------------------

func TestMorB1ShieldBearersGiveTargetPlusZeroPlusThree(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	sb := b12Push(g, me.ID, "Burrenton Shield-Bearers", "Creature — Kithkin Soldier", morB1ShieldBearersOracle, 3, 3)
	declareAttack(t, g, opp.ID, sb)
	b04WaitForPick(t, g, me.ID)
	answerPickTarget(t, g, sb)
	passPriorityAroundTable(t, g)
	if p, tough := morB1PT(t, g, sb); p != 3 || tough != 6 {
		t.Errorf("%d/%d, want 3/6", p, tough)
	}
}

func TestMorB1LunkErrantPumpsOnlyAlone(t *testing.T) {
	for _, alone := range []bool{true, false} {
		g := newCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		lunk := b12Push(g, me.ID, "Lunk Errant", "Creature — Giant Warrior", morB1LunkErrantOracle, 4, 4)
		attackers := []uuid.UUID{lunk}
		if !alone {
			attackers = append(attackers, b12Creature(g, me.ID, "Friend", morB1BearType, 2, 2))
		}
		declareAttack(t, g, opp.ID, attackers...)
		passPriorityAroundTable(t, g)
		want := 4
		if alone {
			want = 5
		}
		if p, tough := morB1PT(t, g, lunk); p != want || tough != want {
			t.Errorf("alone=%v: %d/%d, want %d/%d", alone, p, tough, want, want)
		}
		if got := hasKeywordOnBattlefield(t, g, lunk, "trample"); got != alone {
			t.Errorf("alone=%v: trample=%v", alone, got)
		}
	}
}

func TestMorB1GoldenCricketFliesOnlyWhenYouPay(t *testing.T) {
	for _, pay := range []bool{true, false} {
		g := newCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		cricket := b12Push(g, me.ID, "Order of the Golden Cricket", "Creature — Kithkin Knight", morB1GoldenCricketOracle, 2, 2)
		declareAttack(t, g, opp.ID, cricket)
		b06AddMana(me, "W")
		passPriorityAroundTable(t, g)
		answerPayUnless(t, g, me.ID, pay)
		passPriorityAroundTable(t, g)
		if got := hasEffectiveKeyword(t, g, cricket, "flying"); got != pay {
			t.Errorf("pay=%v: flying=%v", pay, got)
		}
	}
}

func TestMorB1RhysGainsLifeForEachElf(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	rhys := b12Push(g, me.ID, "Rhys the Exiled", "Legendary Creature — Elf Warrior", morB1RhysTheExiledOracle, 3, 2)
	b12Creature(g, me.ID, "Elf", "Creature — Elf", 1, 1)
	b12Creature(g, me.ID, "Bear", morB1BearType, 2, 2)
	life := me.Life
	declareAttack(t, g, opp.ID, rhys)
	passPriorityAroundTable(t, g)
	if me.Life != life+2 {
		t.Errorf("life %d, want %d", me.Life, life+2)
	}
}

func TestMorB1RhysRegeneratesBySacrificingAnElf(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	rhys := b12Push(g, me.ID, "Rhys the Exiled", "Legendary Creature — Elf Warrior", morB1RhysTheExiledOracle, 3, 2)
	elf := b12Creature(g, me.ID, "Elf", "Creature — Elf", 1, 1)
	morB1Activate(t, g, me, rhys, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{elf}})
	if g.Battlefield.Contains(elf) {
		t.Fatal("the Elf was not sacrificed")
	}
	lrwB1DestroyAndExpectRegenerated(t, g, rhys)
}

func TestMorB1StonybrookSchoolmasterMakesAWizardWhenTapped(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	master := b12Push(g, me.ID, "Stonybrook Schoolmaster", "Creature — Merfolk Wizard", morB1StonybrookSchoolmasterOracle, 1, 2)
	declareAttack(t, g, opp.ID, master)
	answerLatestTriggerPrompt(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	if got := morB1Tokens(g, me.ID, "Merfolk Wizard"); got != 1 {
		t.Errorf("%d Merfolk Wizard tokens, want 1", got)
	}
}

// --- statics and step triggers -----------------------------------------------------

func TestMorB1BorderlandBehemothGrowsPerOtherGiant(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	behemoth := b12Push(g, me.ID, "Borderland Behemoth", "Creature — Giant Warrior", morB1BorderlandBehemothOracle, 4, 4)
	b12Creature(g, me.ID, "Giant A", "Creature — Giant", 3, 3)
	b12Creature(g, me.ID, "Giant B", "Creature — Giant", 3, 3)
	b12Creature(g, g.Seats[1].ID, "Their Giant", "Creature — Giant", 3, 3)
	if p, tough := morB1PT(t, g, behemoth); p != 12 || tough != 12 {
		t.Errorf("%d/%d, want 12/12", p, tough)
	}
}

func TestMorB1StenchskipperNeedsAGoblin(t *testing.T) {
	for _, goblin := range []bool{false, true} {
		g := newCatalogGame(t)
		me := g.Seats[0]
		skipper := b12Push(g, me.ID, "Stenchskipper", morB1ElementalType, morB1StenchskipperOracle, 6, 5)
		if goblin {
			b12Creature(g, me.ID, "Goblin", "Creature — Goblin", 1, 1)
		}
		advanceToEndStepOf(t, g, 0)
		passPriorityAroundTable(t, g)
		if kept := g.Battlefield.Contains(skipper); kept != goblin {
			t.Errorf("controls a Goblin = %v: Stenchskipper on the battlefield = %v", goblin, kept)
		}
	}
}

func TestMorB1WeedPrunerPoplarShrinksAnotherCreatureAtUpkeep(t *testing.T) {
	g := newCatalogGame(t)
	owner := g.Seats[1]
	poplar := b12Push(g, owner.ID, "Weed-Pruner Poplar", "Creature — Treefolk Assassin", morB1WeedPrunerPoplarOracle, 3, 3)
	bear := b12Creature(g, g.Seats[0].ID, "Bear", morB1BearType, 2, 2)
	advanceToUpkeepOfSeat(t, g, 1)
	b04WaitForPick(t, g, owner.ID)
	ch := latestPickTarget(g, owner.ID)
	if err := g.ResolvePickTarget(ch.ID, owner.ID, game.TargetRef{Kind: game.TargetCard, ID: poplar}); err == nil {
		t.Fatal("the Poplar targeted itself")
	}
	answerPickTarget(t, g, bear)
	passPriorityAroundTable(t, g)
	if p, tough := morB1PT(t, g, bear); p != 1 || tough != 1 {
		t.Errorf("Bear %d/%d, want 1/1", p, tough)
	}
}
