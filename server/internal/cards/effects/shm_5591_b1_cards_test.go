package effects

import (
	"errors"
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// shm_5591_b1_cards_test.go — Shadowmoor, batch 1 (#5591): two vanilla
// bodies, fifteen creatures whose whole text is keywords, and
// thirty-three cards whose text an existing catalog card already prints
// (restrictions, colour block rules, the Cohorts, self pumps, the
// Lieges, team grants, the Scarecrows and seven spells and triggers).
// Each test plays the real catalog entry through the engine.

const (
	shmB1LoamdraggerOracle   = "b7523824-8c3e-4735-8fbc-6a5f99d1e1ac"
	shmB1OldGhastbarkOracle  = "a5207cb5-7519-4802-85c5-ac8c4313693e"
	shmB1WaspLancerOracle    = "12a9a9a4-6b8c-48a1-833f-457a7ece8e7e"
	shmB1CavaliersOracle     = "69ae765d-03c9-46d9-bdad-8e5936df9616"
	shmB1ZealousOracle       = "508fdfb1-87d6-4342-a05a-01158f253896"
	shmB1PlumeveilOracle     = "14fa0771-cf2b-4563-98b2-ff4ee24bce21"
	shmB1RamGangOracle       = "30d2437a-87c9-4f88-8fb8-b686d6522677"
	shmB1GloomwidowOracle    = "d2546d25-07cd-4830-ba66-34bdbf8a245c"
	shmB1GatewardenOracle    = "439f7ac8-19f0-4227-82a7-48294adc2114"
	shmB1RustrazorOracle     = "58fd5a3c-83ca-4f6a-99db-21d5e03f2345"
	shmB1ScrapperOracle      = "508c673c-2b11-4b7d-8f4b-ff3a364eb4fe"
	shmB1SickleRipperOracle  = "d097b223-c558-4592-abb9-44dc8eb52a61"
	shmB1WildslayerOracle    = "c2f0b795-9f73-43db-aa8c-3b06963ba2b6"
	shmB1AxesharkOracle      = "2141b48d-5a04-40dd-a76b-457ac06b56c2"
	shmB1SafeholdEliteOracle = "68ca91ba-31fb-47e0-9b32-e4f3504cbbca"
	shmB1MaraudersOracle     = "423a1fc2-f623-4b5f-9a3e-7f63b2f3cf5d"
	shmB1OversoulOracle      = "da790c72-63e7-4955-9435-d073767f9bb5"

	shmB1GougerOracle       = "eaedc09f-43b8-42db-a645-0ac3baf1de12"
	shmB1InfiltratorOracle  = "2b43c7de-8c21-4137-b6b8-45965e4bb558"
	shmB1ManiacOracle       = "69ba7864-0d09-4517-ad1a-7efaf437ded0"
	shmB1CragtreadsOracle   = "969233f8-a644-4297-a2bf-b1ca7a58bb0b"
	shmB1RaidersOracle      = "d8863b64-110d-4f6b-9d0e-4662f2b399cb"
	shmB1DragoonOracle      = "e405800a-27e6-468a-938b-484346ee20cd"
	shmB1SootwalkersOracle  = "1492e778-b85e-4675-b48c-bc0a2eb1d105"
	shmB1RootcuttersOracle  = "7e2e96db-e4df-41de-b5a7-712c4a8fe7be"
	shmB1AshenCohortOracle  = "c627f1a8-9a57-4eb9-a8a5-2c285446fe97"
	shmB1BallyCohortOracle  = "7c415180-c179-474b-8c58-e7096bd2c763"
	shmB1BriarCohortOracle  = "598283f3-ae29-43b9-8743-6c065136dcad"
	shmB1CrabCohortOracle   = "470f0043-3ffe-470b-981f-425f0583a87e"
	shmB1MudCohortOracle    = "788bd9ae-b78e-4c86-bf12-d4e19780491a"
	shmB1DieflynOracle      = "4cd6ad78-7441-4dd7-b486-d9575e8e0947"
	shmB1FoxfireOakOracle   = "925775c6-b6d2-4793-a363-bd5aca871aa8"
	shmB1KorriganOracle     = "f281c47f-3587-4a3f-8726-ec59568fef94"
	shmB1ParapetOracle      = "c2ffce33-ab1a-40f1-959b-31583219ed1a"
	shmB1PyreChargerOracle  = "50bc464e-1c8d-455c-a870-ddbbc4618343"
	shmB1RuneCervinOracle   = "5c6956ad-8031-4708-82ed-75491755d2a4"
	shmB1BoartuskOracle     = "4fcffb56-5ab6-44e2-804e-8b3c55cc5415"
	shmB1ThistledownOracle  = "7b07356f-9c3f-4875-939e-aa64eb6b5ce2"
	shmB1RageOracle         = "03fb26b1-6d9c-455b-9b2f-a052a9222db7"
	shmB1RaptorOracle       = "1fd729f6-27ea-4b07-81f0-885f2c0c00eb"
	shmB1BlazethornOracle   = "5f7a4ffb-83f7-46e5-b0a7-c8b068af6969"
	shmB1ThornwatchOracle   = "fe01e07c-dbca-4598-a944-1b194df44260"
	shmB1WatchwingOracle    = "24bd190c-04bf-44d3-97ba-9dafd9adf7ae"
	shmB1ProcessionOracle   = "376271a6-80f0-47a6-990a-5ec5517aae42"
	shmB1JavelinOracle      = "58363090-bf89-4174-bbb2-636afc1c7a06"
	shmB1ScarOracle         = "4eafe158-abff-43e9-aeae-596430c5ce21"
	shmB1PunctureBoltOracle = "c8e53740-c633-42f8-9e02-3434d6dc0b30"
	shmB1HowlOracle         = "868b758c-e21c-4079-a1fc-4c89af106f4b"
	shmB1SprigganOracle     = "f80c8533-b29c-49ae-a241-ef4b73a9e327"
	shmB1CultbrandOracle    = "cbb030aa-ba4f-4d9e-920d-001a40982f5e"

	shmB1BearType = "Creature — Bear"
)

// shmB1Vanilla are the batch's bodies with no rules text.
var shmB1Vanilla = map[string]string{
	shmB1LoamdraggerOracle:  "Loamdragger Giant",
	shmB1OldGhastbarkOracle: "Old Ghastbark",
}

// shmB1Keyworded are the batch's creatures whose whole text is printed
// keywords.
var shmB1Keyworded = map[string]struct {
	name, typeLine string
	keywords       []string
}{
	shmB1WaspLancerOracle:    {"Wasp Lancer", "Creature — Faerie Soldier", []string{"flying"}},
	shmB1CavaliersOracle:     {"Wilt-Leaf Cavaliers", "Creature — Elf Knight", []string{"vigilance"}},
	shmB1ZealousOracle:       {"Zealous Guardian", "Creature — Kithkin Soldier", []string{"flash"}},
	shmB1PlumeveilOracle:     {"Plumeveil", "Creature — Elemental", []string{"flash", "defender", "flying"}},
	shmB1RamGangOracle:       {"Boggart Ram-Gang", "Creature — Goblin Warrior", []string{"haste", "wither"}},
	shmB1GloomwidowOracle:    {"Juvenile Gloomwidow", "Creature — Spider", []string{"reach", "wither"}},
	shmB1GatewardenOracle:    {"Oona's Gatewarden", "Creature — Faerie Soldier", []string{"defender", "flying", "wither"}},
	shmB1RustrazorOracle:     {"Rustrazor Butcher", "Creature — Goblin Warrior", []string{"first strike", "wither"}},
	shmB1ScrapperOracle:      {"Scuzzback Scrapper", "Creature — Goblin Warrior", []string{"wither"}},
	shmB1SickleRipperOracle:  {"Sickle Ripper", "Creature — Elemental Warrior", []string{"wither"}},
	shmB1WildslayerOracle:    {"Wildslayer Elves", "Creature — Elf Warrior", []string{"wither"}},
	shmB1AxesharkOracle:      {"Gravelgill Axeshark", "Creature — Merfolk Soldier", []string{game.KeywordPersist}},
	shmB1SafeholdEliteOracle: {"Safehold Elite", "Creature — Elf Scout", []string{game.KeywordPersist}},
	shmB1MaraudersOracle:     {"Scuzzback Marauders", "Creature — Goblin Warrior", []string{"trample", game.KeywordPersist}},
	shmB1OversoulOracle: {"Oversoul of Dusk", "Creature — Spirit Avatar",
		[]string{"protection from blue", "protection from black", "protection from red"}},
}

// shmB1Push seeds a catalog creature (or, with an empty oracle, a plain
// one) of the given colours under owner.
func shmB1Push(g *game.Game, owner uuid.UUID, name, typeLine, oracle string, power, toughness int, colors ...string) uuid.UUID {
	return pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: typeLine, OracleID: oracle,
		Power: power, Toughness: toughness, Colors: colors, Owner: owner, Controller: owner,
	})
}

// shmB1Seats is the active seat and the next one.
func shmB1Seats(g *game.Game) (*game.Player, *game.Player) {
	return g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
}

func TestShmB1EveryCardIsRegisteredFull(t *testing.T) {
	all := map[string]string{
		shmB1GougerOracle: "Ashenmoor Gouger", shmB1InfiltratorOracle: "Inkfathom Infiltrator",
		shmB1ManiacOracle: "Tattermunge Maniac", shmB1CragtreadsOracle: "Barrenton Cragtreads",
		shmB1RaidersOracle: "Mudbrawler Raiders", shmB1DragoonOracle: "Raven's Run Dragoon",
		shmB1SootwalkersOracle: "Sootwalkers", shmB1RootcuttersOracle: "Wanderbrine Rootcutters",
		shmB1AshenCohortOracle: "Ashenmoor Cohort", shmB1BallyCohortOracle: "Ballynock Cohort",
		shmB1BriarCohortOracle: "Briarberry Cohort", shmB1CrabCohortOracle: "Crabapple Cohort",
		shmB1MudCohortOracle: "Mudbrawler Cohort", shmB1DieflynOracle: "Blistering Dieflyn",
		shmB1FoxfireOakOracle: "Foxfire Oak", shmB1KorriganOracle: "Loch Korrigan",
		shmB1ParapetOracle: "Parapet Watchers", shmB1PyreChargerOracle: "Pyre Charger",
		shmB1RuneCervinOracle: "Rune-Cervin Rider", shmB1BoartuskOracle: "Boartusk Liege",
		shmB1ThistledownOracle: "Thistledown Liege", shmB1RageOracle: "Rage Reflection",
		shmB1RaptorOracle: "Windbrisk Raptor", shmB1BlazethornOracle: "Blazethorn Scarecrow",
		shmB1ThornwatchOracle: "Thornwatch Scarecrow", shmB1WatchwingOracle: "Watchwing Scarecrow",
		shmB1ProcessionOracle: "Spectral Procession", shmB1JavelinOracle: "Flame Javelin",
		shmB1ScarOracle: "Scar", shmB1PunctureBoltOracle: "Puncture Bolt",
		shmB1HowlOracle: "Howl of the Night Pack", shmB1SprigganOracle: "Hungry Spriggan",
		shmB1CultbrandOracle: "Cultbrand Cinder",
	}
	for oracle, name := range shmB1Vanilla {
		all[oracle] = name
	}
	for oracle, k := range shmB1Keyworded {
		all[oracle] = k.name
	}
	if len(all) != 50 {
		t.Fatalf("the batch lists %d cards, want 50", len(all))
	}
	for oracle, name := range all {
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

// --- bodies and keywords -----------------------------------------------------

func TestShmB1VanillaBodiesDeclareNoAbilities(t *testing.T) {
	for oracle, name := range shmB1Vanilla {
		spec, _ := Lookup(oracle)
		if spec.OnResolve != nil || len(spec.Triggered) != 0 || len(spec.Static) != 0 ||
			len(spec.Activated) != 0 || len(spec.PrintedKeywords) != 0 || len(spec.BlockRules) != 0 {
			t.Errorf("%s is a vanilla body and declares no abilities", name)
		}
	}
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	giant := shmB1Push(g, me.ID, "Loamdragger Giant", "Creature — Giant Warrior", shmB1LoamdraggerOracle, 7, 6, "R", "G")
	bear := shmB1Push(g, opp.ID, "Bear", shmB1BearType, "", 2, 2, "G")
	if err := porBlock(t, g, giant, opp.ID, bear); err != nil {
		t.Errorf("a vanilla Loamdragger Giant couldn't be blocked: %v", err)
	}
	if p, tough := ptOf(t, g, giant); p != 7 || tough != 6 {
		t.Errorf("Loamdragger Giant is %d/%d, want 7/6", p, tough)
	}
}

func TestShmB1KeywordedCreaturesHaveTheirKeywords(t *testing.T) {
	for oracle, k := range shmB1Keyworded {
		t.Run(k.name, func(t *testing.T) {
			spec, _ := Lookup(oracle)
			if !slices.Equal(spec.PrintedKeywords, k.keywords) {
				t.Errorf("printed keywords %v, want %v", spec.PrintedKeywords, k.keywords)
			}
			if spec.OnResolve != nil || len(spec.Triggered) != 0 || len(spec.Static) != 0 || len(spec.Activated) != 0 {
				t.Errorf("%s declares more than its keywords", k.name)
			}
			g := newCatalogGame(t)
			id := shmB1Push(g, g.Seats[0].ID, k.name, k.typeLine, oracle, 2, 2)
			for _, kw := range k.keywords {
				if !hasKeywordOnBattlefield(t, g, id, kw) {
					t.Errorf("%s has no %s on the battlefield", k.name, kw)
				}
			}
		})
	}
}

// Wither (CR 702.80a): each wither creature's damage to a creature is
// -1/-1 counters, never marked damage.
func TestShmB1WitherDealsDamageAsMinusCounters(t *testing.T) {
	for oracle, k := range shmB1Keyworded {
		if !slices.Contains(k.keywords, "wither") {
			continue
		}
		t.Run(k.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			src := shmB1Push(g, me.ID, k.name, k.typeLine, oracle, 2, 2)
			victim := shmB1Push(g, opp.ID, "Their Bear", shmB1BearType, "", 4, 4)
			g.WithWriteLock(func() { _ = g.DealDamageToCreatureForEffect(src, victim, 2) })
			g.RunStateChecksForTest()
			if n := countersOn(g, victim, game.CounterMinusOne); n != 2 {
				t.Errorf("-1/-1 counters %d, want 2", n)
			}
			if d := damageOn(t, g, victim); d != 0 {
				t.Errorf("damage marked %d, want 0", d)
			}
		})
	}
}

// Persist (CR 702.79a): the creature comes back once with a -1/-1
// counter, and dies for good the second time.
func TestShmB1PersistReturnsOnce(t *testing.T) {
	for oracle, k := range shmB1Keyworded {
		if !slices.Contains(k.keywords, game.KeywordPersist) {
			continue
		}
		t.Run(k.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[0]
			id := shmB1Push(g, me.ID, k.name, k.typeLine, oracle, 3, 3)
			b25Destroy(g, id)
			passPriorityAroundTable(t, g)
			if !onBattlefield(g, id) || countersOn(g, id, game.CounterMinusOne) != 1 {
				t.Fatalf("on battlefield %v with %d -1/-1 counters, want true and 1",
					onBattlefield(g, id), countersOn(g, id, game.CounterMinusOne))
			}
			b25Destroy(g, id)
			passPriorityAroundTable(t, g)
			if onBattlefield(g, id) {
				t.Error("it persisted with a -1/-1 counter on it")
			}
		})
	}
}

// Flash (CR 702.8a): cast at the beginning of combat.
func TestShmB1FlashCreaturesAreCastInCombat(t *testing.T) {
	for _, tc := range []struct{ name, typeLine, oracle string }{
		{"Zealous Guardian", "Creature — Kithkin Soldier", shmB1ZealousOracle},
		{"Plumeveil", "Creature — Elemental", shmB1PlumeveilOracle},
		{"Thistledown Liege", "Creature — Kithkin Knight", shmB1ThistledownOracle},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			id := uuid.New()
			me.Hand.PushTop(game.Card{
				InstanceID: id, Name: tc.name, TypeLine: tc.typeLine, OracleID: tc.oracle,
				Power: 1, Toughness: 3, Owner: me.ID, Controller: me.ID,
			})
			advanceTo(t, g, game.StepBeginCombat)
			if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
				t.Fatalf("%s at the beginning of combat: %v", tc.name, err)
			}
			passPriorityAroundTable(t, g)
			if !g.Battlefield.Contains(id) {
				t.Errorf("%s did not resolve", tc.name)
			}
		})
	}
}

// Protection from red (CR 702.16b): a red spell can't target Oversoul of
// Dusk; it can target a creature without the protection.
func TestShmB1OversoulOfDuskCantBeTargetedByRed(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me, opp := shmB1Seats(g)
	oversoul := shmB1Push(g, opp.ID, "Oversoul of Dusk", "Creature — Spirit Avatar", shmB1OversoulOracle, 5, 5, "G", "W")
	bear := shmB1Push(g, opp.ID, "Bear", shmB1BearType, "", 2, 2, "G")
	bolt := func() uuid.UUID {
		id := uuid.New()
		me.Hand.PushTop(game.Card{
			InstanceID: id, Name: "Lightning Bolt", TypeLine: "Instant", ManaCost: "{R}", Colors: []string{"R"},
			OracleID: lightningBoltOracle, Owner: me.ID, Controller: me.ID,
		})
		me.ManaPool.AddMana(game.ManaToken{Color: "R"})
		return id
	}
	if err := g.CastSpell(me.ID, bolt(), game.CastSpellParams{Targets: cardTarget(oversoul)}); err == nil {
		t.Error("a red spell targeted a creature with protection from red")
	}
	if err := g.CastSpell(me.ID, bolt(), game.CastSpellParams{Targets: cardTarget(bear)}); err != nil {
		t.Errorf("Lightning Bolt couldn't target the Bear: %v", err)
	}
}

// --- restrictions and block rules ---------------------------------------------

func TestShmB1CantBlockAndCantBeBlocked(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, oracle string
		want                   game.Restriction
	}{
		{"Ashenmoor Gouger", "Creature — Elemental Warrior", shmB1GougerOracle, game.CantBlock},
		{"Inkfathom Infiltrator", "Creature — Merfolk Rogue", shmB1InfiltratorOracle, game.CantBlock | game.CantBeBlocked},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			id := shmB1Push(g, g.Seats[0].ID, tc.name, tc.typeLine, tc.oracle, 2, 2)
			assertRestrictions(t, g, id, tc.want)
		})
	}
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	attacker := shmB1Push(g, me.ID, "Bear", shmB1BearType, "", 2, 2, "G")
	gouger := shmB1Push(g, opp.ID, "Ashenmoor Gouger", "Creature — Elemental Warrior", shmB1GougerOracle, 4, 4, "B", "R")
	if err := porBlock(t, g, attacker, opp.ID, gouger); err == nil {
		t.Error("Ashenmoor Gouger blocked")
	}
	g = newCatalogGame(t)
	me, opp = g.Seats[0], g.Seats[1]
	infiltrator := shmB1Push(g, me.ID, "Inkfathom Infiltrator", "Creature — Merfolk Rogue", shmB1InfiltratorOracle, 2, 1, "U", "B")
	bear := shmB1Push(g, opp.ID, "Bear", shmB1BearType, "", 2, 2, "G")
	if err := porBlock(t, g, infiltrator, opp.ID, bear); err == nil {
		t.Error("Inkfathom Infiltrator was blocked")
	}
}

// CR 508.1d: Tattermunge Maniac can't stay home while it can attack.
func TestShmB1TattermungeManiacAttacksEachCombat(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me, opp := g.Seats[seat], g.Seats[(seat+1)%len(g.Seats)]
	maniac := shmB1Push(g, me.ID, "Tattermunge Maniac", "Creature — Goblin Warrior", shmB1ManiacOracle, 2, 1, "R", "G")
	advanceToDeclareAttackersOf(t, g, seat)
	var re *game.AttackRequirementError
	if err := g.PassPriority(); !errors.As(err, &re) || re.Attacker != maniac {
		t.Fatalf("pass with the Maniac home = %v, want a refusal naming it", err)
	}
	if err := g.DeclareAttacker(maniac, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	if err := g.PassPriority(); err != nil {
		t.Fatalf("pass with the Maniac attacking: %v", err)
	}
}

// "Can't be blocked by <colour> creatures": that colour's blocker is
// refused, another colour's is not.
func TestShmB1CantBeBlockedByAColour(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, oracle string
		refused, allowed       string
	}{
		{"Barrenton Cragtreads", "Creature — Kithkin Scout", shmB1CragtreadsOracle, "R", "W"},
		{"Mudbrawler Raiders", "Creature — Goblin Warrior", shmB1RaidersOracle, "U", "R"},
		{"Raven's Run Dragoon", "Creature — Elf Knight", shmB1DragoonOracle, "B", "G"},
		{"Sootwalkers", "Creature — Elemental Rogue", shmB1SootwalkersOracle, "W", "B"},
		{"Wanderbrine Rootcutters", "Creature — Merfolk Rogue", shmB1RootcuttersOracle, "G", "U"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			attacker := shmB1Push(g, me.ID, tc.name, tc.typeLine, tc.oracle, 3, 3)
			refused := shmB1Push(g, opp.ID, "Refused", shmB1BearType, "", 2, 2, tc.refused)
			allowed := shmB1Push(g, opp.ID, "Allowed", shmB1BearType, "", 2, 2, tc.allowed)
			brAttack(t, g, attacker)
			if err := g.DeclareBlocker(refused, attacker); err == nil {
				t.Errorf("a %s creature blocked %s", tc.refused, tc.name)
			}
			if err := g.DeclareBlocker(allowed, attacker); err != nil {
				t.Errorf("a %s creature couldn't block %s: %v", tc.allowed, tc.name, err)
			}
		})
	}
}

// --- statics -------------------------------------------------------------------

// The Cohorts get +1/+1 only while their controller controls ANOTHER
// creature of the colour: not themselves, not an opponent's, not another
// colour's.
func TestShmB1CohortsNeedAnotherCreatureOfTheirColour(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, oracle string
		power, tough           int
		color, other           string
	}{
		{"Ashenmoor Cohort", "Creature — Elemental Warrior", shmB1AshenCohortOracle, 4, 3, "B", "G"},
		{"Ballynock Cohort", "Creature — Kithkin Soldier", shmB1BallyCohortOracle, 2, 2, "W", "U"},
		{"Briarberry Cohort", "Creature — Faerie Soldier", shmB1BriarCohortOracle, 1, 1, "U", "B"},
		{"Crabapple Cohort", "Creature — Treefolk Warrior", shmB1CrabCohortOracle, 4, 4, "G", "R"},
		{"Mudbrawler Cohort", "Creature — Goblin Warrior", shmB1MudCohortOracle, 1, 1, "R", "W"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			cohort := shmB1Push(g, me.ID, tc.name, tc.typeLine, tc.oracle, tc.power, tc.tough, tc.color)
			check := func(when string, bonus int) {
				t.Helper()
				if p, tough := ptOf(t, g, cohort); p != tc.power+bonus || tough != tc.tough+bonus {
					t.Errorf("%s: %d/%d, want %d/%d", when, p, tough, tc.power+bonus, tc.tough+bonus)
				}
			}
			check("alone", 0)
			shmB1Push(g, opp.ID, "Their Creature", shmB1BearType, "", 1, 1, tc.color)
			check("with an opponent's creature of its colour", 0)
			shmB1Push(g, me.ID, "Off Colour", shmB1BearType, "", 1, 1, tc.other)
			check("with my creature of another colour", 0)
			shmB1Push(g, me.ID, "Same Colour", shmB1BearType, "", 1, 1, tc.color)
			check("with another creature of its colour", 1)
		})
	}
}

func TestShmB1CohortKeywords(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, oracle, keyword string
	}{
		{"Ballynock Cohort", "Creature — Kithkin Soldier", shmB1BallyCohortOracle, "first strike"},
		{"Briarberry Cohort", "Creature — Faerie Soldier", shmB1BriarCohortOracle, "flying"},
		{"Mudbrawler Cohort", "Creature — Goblin Warrior", shmB1MudCohortOracle, "haste"},
	} {
		g := newCatalogGame(t)
		id := shmB1Push(g, g.Seats[0].ID, tc.name, tc.typeLine, tc.oracle, 1, 1)
		if !hasKeywordOnBattlefield(t, g, id, tc.keyword) {
			t.Errorf("%s has no %s", tc.name, tc.keyword)
		}
	}
}

// Each Liege pumps your other creatures of either colour, +2/+2 for one of
// both, and never itself or an opponent's creature.
func TestShmB1LiegesPumpTheirTwoColours(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, oracle, keyword string
		power, tough                    int
		a, b                            string
	}{
		{"Boartusk Liege", "Creature — Goblin Knight", shmB1BoartuskOracle, "trample", 3, 4, "R", "G"},
		{"Thistledown Liege", "Creature — Kithkin Knight", shmB1ThistledownOracle, "flash", 1, 3, "W", "U"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			liege := shmB1Push(g, me.ID, tc.name, tc.typeLine, tc.oracle, tc.power, tc.tough, tc.a, tc.b)
			first := shmB1Push(g, me.ID, "First", shmB1BearType, "", 2, 2, tc.a)
			second := shmB1Push(g, me.ID, "Second", shmB1BearType, "", 2, 2, tc.b)
			both := shmB1Push(g, me.ID, "Both", shmB1BearType, "", 2, 2, tc.a, tc.b)
			neither := shmB1Push(g, me.ID, "Neither", shmB1BearType, "", 2, 2, "B")
			theirs := shmB1Push(g, opp.ID, "Theirs", shmB1BearType, "", 2, 2, tc.a)
			for _, want := range []struct {
				id   uuid.UUID
				who  string
				p, t int
			}{
				{liege, "the Liege", tc.power, tc.tough},
				{first, "my first-colour creature", 3, 3},
				{second, "my second-colour creature", 3, 3},
				{both, "my two-colour creature", 4, 4},
				{neither, "my creature of neither colour", 2, 2},
				{theirs, "the opponent's creature", 2, 2},
			} {
				if p, tough := ptOf(t, g, want.id); p != want.p || tough != want.t {
					t.Errorf("%s is %d/%d, want %d/%d", want.who, p, tough, want.p, want.t)
				}
			}
			if !hasKeywordOnBattlefield(t, g, liege, tc.keyword) {
				t.Errorf("%s has no %s", tc.name, tc.keyword)
			}
		})
	}
}

func TestShmB1RageReflectionGivesYourCreaturesDoubleStrike(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	shmB1Push(g, me.ID, "Rage Reflection", "Enchantment", shmB1RageOracle, 0, 0, "R")
	mine := shmB1Push(g, me.ID, "Mine", shmB1BearType, "", 2, 2, "G")
	theirs := shmB1Push(g, opp.ID, "Theirs", shmB1BearType, "", 2, 2, "G")
	if !hasKeywordOnBattlefield(t, g, mine, "double strike") {
		t.Error("my creature has no double strike")
	}
	if hasKeywordOnBattlefield(t, g, theirs, "double strike") {
		t.Error("the opponent's creature has double strike")
	}
}

// Windbrisk Raptor's lifelink reaches only my creatures while they attack.
func TestShmB1WindbriskRaptorGivesAttackersLifelink(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := seatsForCombat(g)
	raptor := shmB1Push(g, me.ID, "Windbrisk Raptor", "Creature — Bird", shmB1RaptorOracle, 5, 7, "W")
	attacker := shmB1Push(g, me.ID, "Attacker", shmB1BearType, "", 2, 2, "G")
	home := shmB1Push(g, me.ID, "Home", shmB1BearType, "", 2, 2, "G")
	if !hasKeywordOnBattlefield(t, g, raptor, "flying") {
		t.Error("Windbrisk Raptor has no flying")
	}
	if hasKeywordOnBattlefield(t, g, attacker, "lifelink") {
		t.Fatal("a creature has lifelink before it attacks")
	}
	declareAttack(t, g, opp.ID, attacker)
	passPriorityAroundTable(t, g)
	if !hasAbility(effectiveAbilities(t, g, attacker), "lifelink") {
		t.Error("my attacking creature has no lifelink")
	}
	if hasAbility(effectiveAbilities(t, g, home), "lifelink") {
		t.Error("my creature that stayed home has lifelink")
	}
}

// The Scarecrows grant themselves each keyword only while their controller
// controls a creature of its colour.
func TestShmB1ScarecrowsReadTheColoursYouControl(t *testing.T) {
	for _, tc := range []struct {
		name, oracle string
		a, kwA       string
		b, kwB       string
	}{
		{"Blazethorn Scarecrow", shmB1BlazethornOracle, "R", "haste", "G", "wither"},
		{"Thornwatch Scarecrow", shmB1ThornwatchOracle, "G", "wither", "W", "vigilance"},
		{"Watchwing Scarecrow", shmB1WatchwingOracle, "W", "vigilance", "U", "flying"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			sc := shmB1Push(g, me.ID, tc.name, "Artifact Creature — Scarecrow", tc.oracle, 3, 3)
			if hasKeywordOnBattlefield(t, g, sc, tc.kwA) || hasKeywordOnBattlefield(t, g, sc, tc.kwB) {
				t.Fatal("alone, the Scarecrow has a granted keyword")
			}
			shmB1Push(g, me.ID, "First", shmB1BearType, "", 1, 1, tc.a)
			if !hasKeywordOnBattlefield(t, g, sc, tc.kwA) {
				t.Errorf("with a %s creature it has no %s", tc.a, tc.kwA)
			}
			if hasKeywordOnBattlefield(t, g, sc, tc.kwB) {
				t.Errorf("with no %s creature it has %s", tc.b, tc.kwB)
			}
			shmB1Push(g, opp.ID, "Theirs", shmB1BearType, "", 1, 1, tc.b)
			if hasKeywordOnBattlefield(t, g, sc, tc.kwB) {
				t.Errorf("an opponent's %s creature gave it %s", tc.b, tc.kwB)
			}
			shmB1Push(g, me.ID, "Second", shmB1BearType, "", 1, 1, tc.b)
			if !hasKeywordOnBattlefield(t, g, sc, tc.kwB) {
				t.Errorf("with a %s creature it has no %s", tc.b, tc.kwB)
			}
		})
	}
}

// --- activated rows -------------------------------------------------------------

func TestShmB1SelfPumpRows(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, oracle string
		power, tough           int
		mana                   string
		wantP, wantT           int
	}{
		{"Blistering Dieflyn", "Creature — Imp", shmB1DieflynOracle, 0, 1, "R", 1, 1},
		{"Foxfire Oak", "Creature — Treefolk Shaman", shmB1FoxfireOakOracle, 3, 6, "RGR", 6, 6},
		{"Loch Korrigan", "Creature — Spirit", shmB1KorriganOracle, 1, 1, "U", 2, 2},
		{"Parapet Watchers", "Creature — Kithkin Soldier", shmB1ParapetOracle, 2, 2, "W", 2, 3},
		{"Pyre Charger", "Creature — Elemental Warrior", shmB1PyreChargerOracle, 1, 1, "R", 2, 1},
		{"Rune-Cervin Rider", "Creature — Elf Knight", shmB1RuneCervinOracle, 2, 2, "GW", 3, 3},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me, _ := shmB1Seats(g)
			advanceToMain(t, g)
			id := shmB1Push(g, me.ID, tc.name, tc.typeLine, tc.oracle, tc.power, tc.tough)
			if err := conB1Activate(t, g, me, id, 0, "", game.ActivateAbilityParams{}); err == nil {
				t.Fatal("activated with no mana")
			}
			if err := conB1Activate(t, g, me, id, 0, tc.mana, game.ActivateAbilityParams{}); err != nil {
				t.Fatalf("activate: %v", err)
			}
			passPriorityAroundTable(t, g)
			if p, tough := ptOf(t, g, id); p != tc.wantP || tough != tc.wantT {
				t.Errorf("%d/%d, want %d/%d", p, tough, tc.wantP, tc.wantT)
			}
		})
	}
}

func TestShmB1PumpCreatureKeywords(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, oracle, keyword string
	}{
		{"Blistering Dieflyn", "Creature — Imp", shmB1DieflynOracle, "flying"},
		{"Pyre Charger", "Creature — Elemental Warrior", shmB1PyreChargerOracle, "haste"},
		{"Rune-Cervin Rider", "Creature — Elf Knight", shmB1RuneCervinOracle, "flying"},
		{"Hungry Spriggan", "Creature — Goblin Warrior", shmB1SprigganOracle, "trample"},
	} {
		g := newCatalogGame(t)
		id := shmB1Push(g, g.Seats[0].ID, tc.name, tc.typeLine, tc.oracle, 1, 1)
		if !hasKeywordOnBattlefield(t, g, id, tc.keyword) {
			t.Errorf("%s has no %s", tc.name, tc.keyword)
		}
	}
}

// --- spells ----------------------------------------------------------------------

func TestShmB1SpectralProcessionMakesThreeFlyingSpirits(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := shmB1Seats(g)
	castCatalogSpell(t, g, "Spectral Procession", "Sorcery", shmB1ProcessionOracle, nil)
	passPriorityAroundTable(t, g)
	if n, _ := countTokensNamed(g, me.ID, "Spirit"); n != 3 {
		t.Fatalf("%d Spirit tokens, want 3", n)
	}
	for _, c := range tokensNamed(g, "Spirit") {
		if !hasKeywordOnBattlefield(t, g, c.InstanceID, "flying") {
			t.Error("a Spirit token has no flying")
		}
	}
}

func TestShmB1FlameJavelinDealsFourToAnyTarget(t *testing.T) {
	g := newCatalogGame(t)
	_, opp := shmB1Seats(g)
	life := opp.Life
	castCatalogSpell(t, g, "Flame Javelin", "Instant", shmB1JavelinOracle, playerRefs(opp.ID))
	passPriorityAroundTable(t, g)
	if opp.Life != life-4 {
		t.Errorf("opponent life %d -> %d, want -4", life, opp.Life)
	}
	beast := shmB1Push(g, opp.ID, "Beast", shmB1BearType, "", 4, 4, "G")
	castCatalogSpell(t, g, "Flame Javelin", "Instant", shmB1JavelinOracle, cardTarget(beast))
	passPriorityAroundTable(t, g)
	if onBattlefield(g, beast) {
		t.Error("a 4/4 survived 4 damage")
	}
}

func TestShmB1ScarPutsAMinusCounterOnTheTarget(t *testing.T) {
	g := newCatalogGame(t)
	_, opp := shmB1Seats(g)
	bear := shmB1Push(g, opp.ID, "Bear", shmB1BearType, "", 2, 2, "G")
	castCatalogSpell(t, g, "Scar", "Instant", shmB1ScarOracle, cardTarget(bear))
	passPriorityAroundTable(t, g)
	if n := countersOn(g, bear, game.CounterMinusOne); n != 1 {
		t.Errorf("-1/-1 counters %d, want 1", n)
	}
	if p, tough := ptOf(t, g, bear); p != 1 || tough != 1 {
		t.Errorf("the Bear is %d/%d, want 1/1", p, tough)
	}
}

// Puncture Bolt deals 1 and then adds the counter: a 2/2 ends 1/1 with 1
// damage and dies; a 3/3 lives as a 2/2 with 1 damage.
func TestShmB1PunctureBoltDamagesThenAddsACounter(t *testing.T) {
	g := newCatalogGame(t)
	_, opp := shmB1Seats(g)
	small := shmB1Push(g, opp.ID, "Small", shmB1BearType, "", 2, 2, "G")
	big := shmB1Push(g, opp.ID, "Big", shmB1BearType, "", 3, 3, "G")
	castCatalogSpell(t, g, "Puncture Bolt", "Instant", shmB1PunctureBoltOracle, cardTarget(big))
	passPriorityAroundTable(t, g)
	if n := countersOn(g, big, game.CounterMinusOne); n != 1 {
		t.Errorf("-1/-1 counters %d, want 1", n)
	}
	if d := damageOn(t, g, big); d != 1 {
		t.Errorf("damage marked %d, want 1", d)
	}
	castCatalogSpell(t, g, "Puncture Bolt", "Instant", shmB1PunctureBoltOracle, cardTarget(small))
	passPriorityAroundTable(t, g)
	g.RunStateChecksForTest()
	if onBattlefield(g, small) {
		t.Error("a 2/2 survived 1 damage and a -1/-1 counter")
	}
}

func TestShmB1HowlOfTheNightPackMakesAWolfPerForest(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := shmB1Seats(g)
	for range 3 {
		b12Permanent(g, me.ID, "Forest", "Basic Land — Forest")
	}
	b12Permanent(g, me.ID, "Mountain", "Basic Land — Mountain")
	b12Permanent(g, opp.ID, "Forest", "Basic Land — Forest")
	castCatalogSpell(t, g, "Howl of the Night Pack", "Sorcery", shmB1HowlOracle, nil)
	passPriorityAroundTable(t, g)
	if n, _ := countTokensNamed(g, me.ID, "Wolf"); n != 3 {
		t.Errorf("%d Wolf tokens, want 3 (one per Forest I control)", n)
	}
	if n, _ := countTokensNamed(g, opp.ID, "Wolf"); n != 0 {
		t.Errorf("the opponent got %d Wolves", n)
	}
}

// --- triggers ----------------------------------------------------------------------

func TestShmB1HungrySprigganGetsPlusThreeWhenItAttacks(t *testing.T) {
	g := newCatalogGame(t)
	me, them := seatsForCombat(g)
	spriggan := shmB1Push(g, me.ID, "Hungry Spriggan", "Creature — Goblin Warrior", shmB1SprigganOracle, 1, 1, "G")
	declareAttack(t, g, them.ID, spriggan)
	passPriorityAroundTable(t, g)
	if p, tough := ptOf(t, g, spriggan); p != 4 || tough != 4 {
		t.Errorf("Hungry Spriggan is %d/%d while attacking, want 4/4", p, tough)
	}
}

func TestShmB1CultbrandCinderPutsACounterOnEntry(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := shmB1Seats(g)
	victim := shmB1Push(g, opp.ID, "Victim", shmB1BearType, "", 3, 3, "G")
	castHoldCreature(t, g, "Cultbrand Cinder", "Creature — Elemental Shaman", shmB1CultbrandOracle, 3, 3)
	b04WaitForPick(t, g, me.ID)
	pickCard(t, g, me.ID, victim)
	passPriorityAroundTable(t, g)
	if n := countersOn(g, victim, game.CounterMinusOne); n != 1 {
		t.Errorf("-1/-1 counters %d, want 1", n)
	}
}
