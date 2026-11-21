package testsex

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func TestFdnSEveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		effectstest.FdnSSearslicerOracle: "Searslicer Goblin",
		effectstest.FdnSFollyOracle:      "Seeker's Folly",
		effectstest.FdnSDowserOracle:     "Shipwreck Dowser",
		effectstest.FdnSSireOracle:       "Sire of Seven Deaths",
		effectstest.FdnSArcherOracle:     "Skeleton Archer",
		effectstest.FdnSBuccaneerOracle:  "Skyship Buccaneer",
		effectstest.FdnSSlagstormOracle:  "Slagstorm",
		effectstest.FdnSSowerOracle:      "Sower of Chaos",
		effectstest.FdnSLagacOracle:      "Spitfire Lagac",
		effectstest.FdnSStarlightOracle:  "Starlight Snare",
		effectstest.FdnSStasisOracle:     "Stasis Snare",
		effectstest.FdnSSpyOracle:        "Storm Fleet Spy",
		effectstest.FdnSStrixOracle:      "Strix Lookout",
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

func TestFdnSPrintedKeywords(t *testing.T) {
	for _, c := range []struct {
		name, typeLine, oracle string
		power, toughness       int
		keywords               []string
	}{
		{"Sire of Seven Deaths", "Creature — Eldrazi", effectstest.FdnSSireOracle, 7, 7, []string{"reach", "first strike", "vigilance", "menace", "trample", "lifelink"}},
		{"Skyship Buccaneer", "Creature — Human Pirate", effectstest.FdnSBuccaneerOracle, 4, 3, []string{"flying"}},
		{"Strix Lookout", "Creature — Bird", effectstest.FdnSStrixOracle, 1, 2, []string{"flying", "vigilance"}},
		{"Shipwreck Dowser", "Creature — Merfolk Wizard", effectstest.FdnSDowserOracle, 3, 3, []string{game.KeywordProwess}},
	} {
		g := effectstest.NewCatalogGame(t)
		id := effectstest.B12Push(g, g.Seats[0].ID, c.name, c.typeLine, c.oracle, c.power, c.toughness)
		effectstest.AssertKeywords(t, g, id, c.keywords...)
	}
	spec, _ := effects.Lookup(effectstest.FdnSStasisOracle)
	if len(spec.PrintedKeywords) != 1 || spec.PrintedKeywords[0] != "flash" {
		t.Errorf("Stasis Snare keywords %v, want flash", spec.PrintedKeywords)
	}
}

// fdnSCastCreature casts a catalog creature with its printed body from
// the active player's hand in the current main phase, without resolving
// it.
func fdnSCastCreature(t *testing.T, g *game.Game, name, typeLine, oracle string, power, toughness int) uuid.UUID {
	t.Helper()
	active := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	active.Hand.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: typeLine, OracleID: oracle,
		Power: power, Toughness: toughness, Owner: active.ID, Controller: active.ID,
	})
	effectstest.AdvanceToMain(t, g)
	if err := g.CastSpell(active.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("CastSpell %s: %v", name, err)
	}
	return id
}

// fdnSCastMode casts a catalog sorcery choosing one mode.
func fdnSCastMode(t *testing.T, g *game.Game, name, oracle string, mode int, targets []game.TargetRef) {
	t.Helper()
	active := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	active.Hand.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: "Sorcery", OracleID: oracle,
		Owner: active.ID, Controller: active.ID,
	})
	effectstest.AdvanceToMain(t, g)
	if err := g.CastSpell(active.ID, id, game.CastSpellParams{Modes: []int{mode}, Targets: targets}); err != nil {
		t.Fatalf("CastSpell %s mode %d: %v", name, mode, err)
	}
}

// --- Raid: Storm Fleet Spy, Skyship Buccaneer, Searslicer Goblin ----------------

func TestFdnSRaidDrawsOnlyAfterAnAttack(t *testing.T) {
	for _, c := range []struct {
		name, typeLine, oracle string
		power, toughness       int
	}{
		{"Storm Fleet Spy", "Creature — Human Pirate", effectstest.FdnSSpyOracle, 2, 2},
		{"Skyship Buccaneer", "Creature — Human Pirate", effectstest.FdnSBuccaneerOracle, 4, 3},
	} {
		for _, attacked := range []bool{false, true} {
			name := c.name + "/no attack"
			if attacked {
				name = c.name + "/attacked"
			}
			t.Run(name, func(t *testing.T) {
				g := effectstest.NewCatalogGame(t)
				me, opp := g.Seats[0], g.Seats[1]
				effectstest.SeedLibrary(me, "First", "Second", "Third")
				bear := effectstest.PushVanillaCreature(g, me.ID, "Bear", 2, 2)
				effectstest.AdvanceToMain(t, g)
				if attacked {
					effectstest.DeclareAttack(t, g, opp.ID, bear)
					effectstest.PassPriorityAroundTable(t, g)
					effectstest.MkAdvanceTo(t, g, game.StepPostcombatMain)
				}
				before := effectstest.HandSize(me)
				fdnSCastCreature(t, g, c.name, c.typeLine, c.oracle, c.power, c.toughness)
				effectstest.PassPriorityAroundTable(t, g)
				effectstest.PassPriorityAroundTable(t, g)
				want := before
				if attacked {
					want++
				}
				if got := effectstest.HandSize(me); got != want {
					t.Errorf("hand %d, want %d", got, want)
				}
			})
		}
	}
}

func TestFdnSSearslicerGoblinMakesAGoblinAtYourEndStepOnlyAfterAnAttack(t *testing.T) {
	for _, attacked := range []bool{false, true} {
		name := "no attack"
		if attacked {
			name = "attacked"
		}
		t.Run(name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			effectstest.B12Push(g, me.ID, "Searslicer Goblin", "Creature — Goblin Warrior", effectstest.FdnSSearslicerOracle, 2, 1)
			bear := effectstest.PushVanillaCreature(g, me.ID, "Bear", 2, 2)
			effectstest.AdvanceToMain(t, g)
			if attacked {
				effectstest.DeclareAttack(t, g, opp.ID, bear)
				effectstest.PassPriorityAroundTable(t, g)
			}
			effectstest.AdvanceToEndStep(t, g)
			effectstest.PassPriorityAroundTable(t, g)
			want := 0
			if attacked {
				want = 1
			}
			if got := effectstest.CountOnBattlefield(g, "Goblin", me.ID); got != want {
				t.Errorf("%d Goblin tokens, want %d", got, want)
			}
		})
	}
}

// --- Skeleton Archer ----------------------------------------------------------

func TestFdnSSkeletonArcherPingsAnyTargetOnEntry(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	life := opp.Life
	effectstest.CastHoldCreature(t, g, "Skeleton Archer", "Creature — Skeleton Archer", effectstest.FdnSArcherOracle, 3, 3)
	effectstest.B04WaitForPick(t, g, me.ID)
	pick := effectstest.LatestPickTarget(g, me.ID)
	if err := g.ResolvePickTarget(pick.ID, me.ID, game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID}); err != nil {
		t.Fatalf("ResolvePickTarget: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-1 {
		t.Errorf("opponent at %d, want %d", opp.Life, life-1)
	}
}

func TestFdnSSkeletonArcherKillsAOneToughnessCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	elf := effectstest.B12Creature(g, opp.ID, "Their Elf", "Creature — Elf", 1, 1)
	effectstest.CastHoldCreature(t, g, "Skeleton Archer", "Creature — Skeleton Archer", effectstest.FdnSArcherOracle, 3, 3)
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.PickCard(t, g, me.ID, elf)
	effectstest.PassPriorityAroundTable(t, g)
	g.RunStateChecksForTest()
	if g.Battlefield.Contains(elf) {
		t.Error("a 1/1 survived 1 damage")
	}
}

// --- Spitfire Lagac -----------------------------------------------------------

func TestFdnSSpitfireLagacPingsEachOpponentOnLandfall(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B12Push(g, me.ID, "Spitfire Lagac", "Creature — Lizard", effectstest.FdnSLagacOracle, 3, 4)
	before := effectstest.B27Lives(g)
	effectstest.PlayLandFromHand(t, g, "Mountain", "")
	effectstest.PassPriorityAroundTable(t, g)
	for i := 1; i < len(g.Seats); i++ {
		if got := g.Seats[i].Life; got != before[i]-1 {
			t.Errorf("opponent %d: %d -> %d, want -1", i, before[i], got)
		}
	}
	if me.Life != before[0] {
		t.Error("the controller is not an opponent")
	}
}

// --- Strix Lookout ------------------------------------------------------------

func TestFdnSStrixLookoutLootsForOneBlue(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	strix := effectstest.B12Push(g, me.ID, "Strix Lookout", "Creature — Bird", effectstest.FdnSStrixOracle, 1, 2)
	effectstest.SeedLibrary(me, "Drawn")
	effectstest.AdvanceToMain(t, g)
	effectstest.B06AddMana(me, "C")
	if err := g.ActivateCatalogAbility(me.ID, strix, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("activated without blue mana")
	}
	effectstest.B06AddMana(me, "U")
	hand, grave := effectstest.HandSize(me), me.Graveyard.Size()
	effectstest.B16Activate(t, g, me.ID, strix, 0, game.ActivateAbilityParams{Strict: true})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.DiscardFromHand(t, g, me.ID)
	if !effectstest.IsTapped(g, strix) {
		t.Error("the Lookout is not tapped")
	}
	if effectstest.HandSize(me) != hand || me.Graveyard.Size() != grave+1 {
		t.Errorf("hand %d -> %d, graveyard %d -> %d; want the same hand and one more card in the graveyard",
			hand, effectstest.HandSize(me), grave, me.Graveyard.Size())
	}
}

// --- Sower of Chaos -----------------------------------------------------------

func TestFdnSSowerOfChaosStopsATargetCreatureBlocking(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	sower := effectstest.B12Push(g, me.ID, "Sower of Chaos", "Creature — Devil", effectstest.FdnSSowerOracle, 4, 3)
	wall := effectstest.B12Creature(g, opp.ID, "Their Wall", "Creature — Wall", 0, 5)
	effectstest.AdvanceToMain(t, g)
	effectstest.B06AddMana(me, "C", "C", "R")
	effectstest.B16Activate(t, g, me.ID, sower, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardRefs(wall)})
	effectstest.PassPriorityAroundTable(t, g)
	if r := effectstest.EffectiveOf(t, g, wall).Restrictions; r&game.CantBlock == 0 {
		t.Errorf("restrictions %v, want CantBlock", r)
	}
	if effectstest.IsTapped(g, sower) {
		t.Error("the ability has no {T} in its cost")
	}
}

// --- Starlight Snare ----------------------------------------------------------

func TestFdnSStarlightSnareTapsAndHoldsItsCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	host := effectstest.PushPermanentForTest(g, me.ID, "Host", "", "Creature — Bear")
	aura := effectstest.CastCatalogSpell(t, g, "Starlight Snare", effectstest.AuraTypeLine, effectstest.FdnSStarlightOracle, effectstest.CardRefs(host))
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.AttachmentHostOf(t, g, aura); got.Kind != game.TargetCard || got.ID != host {
		t.Fatalf("Starlight Snare attached to %+v, want host %s", got, host)
	}
	if c, _ := effectstest.BattlefieldCard(g, host); !c.Tapped {
		t.Fatal("the enters trigger did not tap the enchanted creature")
	}
	effectstest.AdvanceToUpkeepOf(t, g, 0)
	if c, _ := effectstest.BattlefieldCard(g, host); !c.Tapped {
		t.Error("the enchanted creature untapped during its controller's untap step")
	}
}

// --- Stasis Snare -------------------------------------------------------------

func TestFdnSStasisSnareExilesUntilItLeaves(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	bear := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	life := me.Life

	snare := effectstest.CastHoldCreature(t, g, "Stasis Snare", "Enchantment", effectstest.FdnSStasisOracle, 0, 0)
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.PickCard(t, g, me.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) || effectstest.OnBattlefieldNamed(g, "Their Bear") != 0 {
		t.Fatal("the creature was not exiled")
	}
	if me.Life != life {
		t.Errorf("life %d -> %d: the Snare gains no life", life, me.Life)
	}

	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(snare) })
	g.RunStateChecksForTest()
	if effectstest.TmtB1Count(g, opp.ID, "Their Bear") != 1 {
		t.Error("the creature did not return when the Snare left")
	}
}

func TestFdnSStasisSnareOffersOnlyAnOpponentsCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	mine := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	effectstest.CastHoldCreature(t, g, "Stasis Snare", "Enchantment", effectstest.FdnSStasisOracle, 0, 0)
	if p := effectstest.LatestPickTarget(g, me.ID); p != nil {
		if err := g.ResolvePickTarget(p.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: mine}); err == nil {
			t.Error("my own creature was accepted as the target")
		}
	}
	if !g.Battlefield.Contains(mine) {
		t.Error("my own creature was exiled")
	}
}

// --- Slagstorm ----------------------------------------------------------------

func TestFdnSSlagstormDealsThreeToEachCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	mine := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 3, 3)
	big := effectstest.B12Creature(g, opp.ID, "Their Ox", "Creature — Ox", 4, 4)
	before := effectstest.B27Lives(g)
	fdnSCastMode(t, g, "Slagstorm", effectstest.FdnSSlagstormOracle, 0, nil)
	effectstest.PassPriorityAroundTable(t, g)
	g.RunStateChecksForTest()
	if g.Battlefield.Contains(theirs) || g.Battlefield.Contains(mine) {
		t.Error("a creature with toughness 3 or less survived 3 damage")
	}
	if !g.Battlefield.Contains(big) || effectstest.DamageMarkedOn(g, big) != 3 {
		t.Error("the 4/4 was not dealt exactly 3 damage")
	}
	for i, l := range effectstest.B27Lives(g) {
		if l != before[i] {
			t.Errorf("seat %d life %d, want an unchanged %d", i, l, before[i])
		}
	}
}

func TestFdnSSlagstormDealsThreeToEachPlayer(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	bear := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	before := effectstest.B27Lives(g)
	fdnSCastMode(t, g, "Slagstorm", effectstest.FdnSSlagstormOracle, 1, nil)
	effectstest.PassPriorityAroundTable(t, g)
	for i, l := range effectstest.B27Lives(g) {
		if l != before[i]-3 {
			t.Errorf("seat %d life %d, want %d", i, l, before[i]-3)
		}
	}
	if !g.Battlefield.Contains(bear) || effectstest.DamageMarkedOn(g, bear) != 0 {
		t.Error("the player mode dealt damage to a creature")
	}
}

// --- Seeker's Folly -----------------------------------------------------------

func TestFdnSSeekersFollyMakesTargetOpponentDiscardTwo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	for _, n := range []string{"One", "Two", "Three"} {
		opp.Hand.PushTop(game.Card{InstanceID: uuid.New(), Name: n, TypeLine: "Instant", Owner: opp.ID, Controller: opp.ID})
	}
	hand, grave := effectstest.HandSize(opp), opp.Graveyard.Size()
	fdnSCastMode(t, g, "Seeker's Folly", effectstest.FdnSFollyOracle, 0, []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.DiscardChoiceFor(g, opp.ID)
	if c == nil {
		t.Fatal("no discard prompt for the target opponent")
	}
	if c.ChooseMax != 2 {
		t.Errorf("discard prompt asks for %d cards, want 2", c.ChooseMax)
	}
	effectstest.DiscardFromHand(t, g, opp.ID)
	if effectstest.HandSize(opp) != hand-2 || opp.Graveyard.Size() != grave+2 {
		t.Errorf("opponent hand %d -> %d, graveyard %d -> %d; want two cards moved", hand, effectstest.HandSize(opp), grave, opp.Graveyard.Size())
	}
}

func TestFdnSSeekersFollyShrinksOnlyYourOpponentsCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	elf := effectstest.B12Creature(g, opp.ID, "Their Elf", "Creature — Elf", 1, 1)
	ox := effectstest.B12Creature(g, opp.ID, "Their Ox", "Creature — Ox", 3, 3)
	mine := effectstest.B12Creature(g, me.ID, "My Elf", "Creature — Elf", 1, 1)
	fdnSCastMode(t, g, "Seeker's Folly", effectstest.FdnSFollyOracle, 1, nil)
	effectstest.PassPriorityAroundTable(t, g)
	g.RunStateChecksForTest()
	if g.Battlefield.Contains(elf) {
		t.Error("an opponent's 1/1 survived -1/-1")
	}
	if p, tough := effectstest.EffectivePower(t, g, ox), effectstest.EffectiveToughness(t, g, ox); p != 2 || tough != 2 {
		t.Errorf("the opponent's 3/3 is %d/%d, want 2/2", p, tough)
	}
	if !g.Battlefield.Contains(mine) {
		t.Error("your own creature was shrunk")
	}
}

// --- Shipwreck Dowser ---------------------------------------------------------

func TestFdnSShipwreckDowserReturnsAnInstantOrSorceryFromYourGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bolt := effectstest.GraveyardCardOf(g, me, "Bolt", "Instant", "{R}", "")
	bear := effectstest.GraveyardCardOf(g, me, "Dead Bear", "Creature — Bear", "{1}{G}", "")
	theirs := effectstest.GraveyardCardOf(g, opp, "Their Bolt", "Sorcery", "{R}", "")

	effectstest.CastAndResolveCreature(t, g, "Shipwreck Dowser", "Creature — Merfolk Wizard", effectstest.FdnSDowserOracle)
	prompt := effectstest.LatestPickTarget(g, me.ID)
	if prompt == nil {
		t.Fatal("no target pick")
	}
	if !effectstest.HasID(prompt.PickTargetCards, bolt) || effectstest.HasID(prompt.PickTargetCards, bear) || effectstest.HasID(prompt.PickTargetCards, theirs) {
		t.Errorf("legal set = %v, want only my instant or sorcery", prompt.PickTargetCards)
	}
	effectstest.PickCard(t, g, me.ID, bolt)
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(bolt) || me.Graveyard.Contains(bolt) {
		t.Error("the chosen card did not return to hand")
	}
}

// --- Sire of Seven Deaths -----------------------------------------------------

func TestFdnSSireOfSevenDeathsWardAsksForSevenLife(t *testing.T) {
	g, _, opp, sire := effectstest.WardTable(t, "Sire of Seven Deaths", "Creature — Eldrazi", effectstest.FdnSSireOracle)
	effectstest.CastAtWardedCreature(t, g, opp, sire)
	prompt := effectstest.PassUntilConfirmFor(t, g, opp.ID)
	if prompt == nil {
		t.Fatal("no ward prompt for the spell's controller")
	}
	if prompt.LifeCost != 7 {
		t.Errorf("prompt LifeCost = %d, want 7", prompt.LifeCost)
	}
}
