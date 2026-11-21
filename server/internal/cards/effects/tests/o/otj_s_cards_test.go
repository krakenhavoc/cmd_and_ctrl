package testso

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// otj_s_cards_test.go — Outlaws of Thunder Junction (OTJ), the S cards
// (#4468). Each test plays the real catalog entry through the engine and
// asserts what the printed text says.

const (
	otjSSandstormVergeOracle    = "de417a82-8f03-4d7e-aee7-48f7d7eba61a"
	otjSScalestormOracle        = "55ede1d4-593d-433c-943a-4db12eeb5fa8"
	otjSScorchingShotOracle     = "e13113ba-1923-493c-a48d-ca320b7adbde"
	otjSSeizeOracle             = "db070312-9a27-435f-8b23-e538441103ec"
	otjSSlickSequenceOracle     = "4ac8c118-3942-4c92-856c-20d764283650"
	otjSShowOffOracle           = "74c7a029-59aa-4511-8e4c-f8f167cda4e2"
	otjSVaultBusterOracle       = "2085642e-29be-4e13-81dc-1f2ff6c095eb"
	otjSSouredSpringsOracle     = "e579edd7-4f6c-4f22-a72f-0a20d7a698a2"
	otjSSpringSplasherOracle    = "f1136a22-b202-4688-a44c-a9979c91261a"
	otjSStagecoachOracle        = "4b4589ff-d141-49a4-a0bd-1c73fc3651ef"
	otjSSterlingHoundOracle     = "3178c1f9-f644-4926-ac9a-d40e4db79955"
	otjSSterlingKeykeeperOracle = "f893d3d6-efef-4394-8e15-e01deed72b4f"
	otjSSterlingSupplierOracle  = "ab23cca2-b59e-4d34-80de-98bd08985275"
	otjSStingerbackOracle       = "5cc6ca73-c38f-4ef8-a2c9-4f0342eebb73"
)

func TestOtjSEveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		otjSSandstormVergeOracle:    "Sandstorm Verge",
		otjSScalestormOracle:        "Scalestorm Summoner",
		otjSScorchingShotOracle:     "Scorching Shot",
		otjSSeizeOracle:             "Seize the Secrets",
		otjSSlickSequenceOracle:     "Slick Sequence",
		otjSShowOffOracle:           "Slickshot Show-Off",
		otjSVaultBusterOracle:       "Slickshot Vault-Buster",
		otjSSouredSpringsOracle:     "Soured Springs",
		otjSSpringSplasherOracle:    "Spring Splasher",
		otjSStagecoachOracle:        "Stagecoach Security",
		otjSSterlingHoundOracle:     "Sterling Hound",
		otjSSterlingKeykeeperOracle: "Sterling Keykeeper",
		otjSSterlingSupplierOracle:  "Sterling Supplier",
		otjSStingerbackOracle:       "Stingerback Terror",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s (%s) is not registered", name, oracle)
			continue
		}
		if spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
	for oracle, cost := range map[string]string{
		otjSShowOffOracle:     "{1}{R}",
		otjSStagecoachOracle:  "{3}{W}",
		otjSStingerbackOracle: "{2}{R}",
	} {
		if sa := effectstest.PlotOf(oracle); sa == nil || sa.Cost != cost {
			t.Errorf("%s: plot %+v, want %s", oracle, sa, cost)
		}
	}
}

// --- Sandstorm Verge ---------------------------------------------------------

func TestOtjSSandstormVergeStopsATargetFromBlocking(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	verge := effectstest.PushCatalogPermanent(g, me.ID, "Sandstorm Verge", "Land — Desert", otjSSandstormVergeOracle, false)
	attacker := effectstest.B12Creature(g, me.ID, "Raider", "Creature — Human", 2, 2)
	wall := effectstest.B12Creature(g, opp.ID, "Wall", "Creature — Wall", 0, 4)
	other := effectstest.B12Creature(g, opp.ID, "Other", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	effectstest.B06AddMana(me, "C", "C", "C")
	effectstest.B16Activate(t, g, me.ID, verge, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(wall)})
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.B16Tapped(t, g, verge) {
		t.Error("{T} is part of the cost")
	}
	effectstest.BrAttack(t, g, attacker)
	effectstest.BrRefusal(t, g.DeclareBlocker(wall, attacker), game.BlockReasonCantBlock)
	if err := g.DeclareBlocker(other, attacker); err != nil {
		t.Fatalf("an untargeted creature still blocks: %v", err)
	}
	spec, _ := effects.Lookup(otjSSandstormVergeOracle)
	if len(spec.ManaAbilities) != 1 || spec.ManaAbilities[0].Produced != "{C}" {
		t.Errorf("mana row %+v, want {C}", spec.ManaAbilities)
	}
}

func TestOtjSSandstormVergeActivatesOnlyAsASorcery(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	verge := effectstest.PushCatalogPermanent(g, me.ID, "Sandstorm Verge", "Land — Desert", otjSSandstormVergeOracle, false)
	wall := effectstest.B12Creature(g, opp.ID, "Wall", "Creature — Wall", 0, 4)
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	effectstest.B06AddMana(me, "C", "C", "C")
	if err := g.ActivateCatalogAbility(me.ID, verge, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(wall)}); err == nil {
		t.Error("activated outside a main phase")
	}
}

// --- Scalestorm Summoner -----------------------------------------------------

func TestOtjSScalestormSummonerMakesADinosaurOnlyWithAPowerFourCreature(t *testing.T) {
	for _, tc := range []struct {
		name string
		big  bool
		want int
	}{
		{"no power 4 creature", false, 0},
		{"a power 4 creature", true, 1},
	} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		summoner := effectstest.B12Push(g, me.ID, "Scalestorm Summoner", "Creature — Human Warlock", otjSScalestormOracle, 3, 3)
		if tc.big {
			effectstest.B12Creature(g, me.ID, "My Giant", "Creature — Giant", 4, 4)
		}
		effectstest.AdvanceToMain(t, g)
		effectstest.LciB1Attack(t, g, summoner, opp.ID)
		var dinos []uuid.UUID
		for _, c := range g.Battlefield.Cards {
			if c.Name == "Dinosaur" && c.IsToken() && c.Controller == me.ID {
				dinos = append(dinos, c.InstanceID)
			}
		}
		if len(dinos) != tc.want {
			t.Fatalf("%s: %d Dinosaurs, want %d", tc.name, len(dinos), tc.want)
		}
		if tc.want == 1 {
			if p, tt := effectstest.PtOf(t, g, dinos[0]); p != 3 || tt != 1 {
				t.Errorf("the Dinosaur is %d/%d, want 3/1", p, tt)
			}
			if c := effectstest.LayeredCard(t, g, dinos[0]); len(c.Colors) != 1 || c.Colors[0] != "R" {
				t.Errorf("the Dinosaur's colours are %v, want red", c.Colors)
			}
		}
	}
}

// --- Scorching Shot ----------------------------------------------------------

func TestOtjSScorchingShotDealsFiveToACreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	big := effectstest.B12Creature(g, opp.ID, "Their Giant", "Creature — Giant", 6, 6)
	effectstest.AdvanceToMain(t, g)
	effectstest.CastCatalogSpell(t, g, "Scorching Shot", "Sorcery", otjSScorchingShotOracle, effectstest.B16TargetCard(big))
	effectstest.PassPriorityAroundTable(t, g)
	if got := otjAC2Damage(t, g, big); got != 5 {
		t.Errorf("the 6/6 has %d damage, want 5", got)
	}
}

// --- Seize the Secrets -------------------------------------------------------

func TestOtjSSeizeTheSecretsCostsOneLessAfterACrimeAndDrawsTwo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	if got := effectstest.SelfPricedMV(t, g, me, otjSSeizeOracle, "Sorcery", "{2}{U}"); got != 3 {
		t.Errorf("no crime: %d, want 3", got)
	}
	effectstest.CastCatalogSpell(t, g, "Lightning Bolt", "Instant", effectstest.LightningBoltOracle, effectstest.B16TargetCard(mine))
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.SelfPricedMV(t, g, me, otjSSeizeOracle, "Sorcery", "{2}{U}"); got != 3 {
		t.Errorf("targeting your own creature is no crime: %d, want 3", got)
	}
	otjB1BoltOpponent(t, g, opp.ID)
	if got := effectstest.SelfPricedMV(t, g, me, otjSSeizeOracle, "Sorcery", "{2}{U}"); got != 2 {
		t.Errorf("after a crime: %d, want 2", got)
	}

	effectstest.SeedLibrary(me, "One", "Two", "Three")
	hand := me.Hand.Size()
	effectstest.CastCatalogSpell(t, g, "Seize the Secrets", "Sorcery", otjSSeizeOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 2 {
		t.Errorf("drew %d, want 2", got)
	}
}

// --- Slick Sequence ----------------------------------------------------------

func TestOtjSSlickSequenceDrawsOnlyAfterAnotherSpell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.SeedLibrary(me, "One", "Two", "Three")
	effectstest.AdvanceToMain(t, g)
	life := opp.Life
	hand := me.Hand.Size()
	effectstest.CastCatalogSpell(t, g, "Slick Sequence", "Instant", otjSSlickSequenceOracle, effectstest.B16TargetPlayer(opp.ID))
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-2 {
		t.Errorf("the opponent went %d → %d, want 2 damage", life, opp.Life)
	}
	if got := me.Hand.Size() - hand; got != 0 {
		t.Errorf("the first spell this turn drew %d", got)
	}

	hand = me.Hand.Size()
	effectstest.CastCatalogSpell(t, g, "Slick Sequence", "Instant", otjSSlickSequenceOracle, effectstest.B16TargetPlayer(opp.ID))
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-4 {
		t.Errorf("the opponent is at %d, want %d", opp.Life, life-4)
	}
	if got := me.Hand.Size() - hand; got != 1 {
		t.Errorf("another spell was cast this turn: drew %d, want 1", got)
	}
}

// --- Slickshot Show-Off ------------------------------------------------------

func TestOtjSShowOffGetsTwoForEachNoncreatureSpell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	showOff := effectstest.B12Push(g, me.ID, "Slickshot Show-Off", "Creature — Bird Wizard", otjSShowOffOracle, 1, 2)
	for _, kw := range []string{"flying", "haste"} {
		if !effectstest.HasEffectiveKeyword(t, g, showOff, kw) {
			t.Errorf("%s is printed", kw)
		}
	}
	effectstest.B24CastNoncreature(t, g, "Opt", "Instant", "{U}")
	effectstest.PassPriorityAroundTable(t, g)
	if p, tt := effectstest.PtOf(t, g, showOff); p != 3 || tt != 2 {
		t.Errorf("after a noncreature spell: %d/%d, want 3/2", p, tt)
	}
	effectstest.CastHoldCreature(t, g, "Bear", "Creature — Bear", "", 2, 2)
	effectstest.PassPriorityAroundTable(t, g)
	if p, _ := effectstest.PtOf(t, g, showOff); p != 3 {
		t.Errorf("a creature spell pumped it: power %d, want 3", p)
	}
	effectstest.AdvanceToMainOf(t, g, 1)
	if p, _ := effectstest.PtOf(t, g, showOff); p != 1 {
		t.Errorf("until end of turn only: power %d", p)
	}
}

// --- Slickshot Vault-Buster --------------------------------------------------

func TestOtjSVaultBusterGetsTwoAfterACrimeThisTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	buster := effectstest.B12Push(g, me.ID, "Slickshot Vault-Buster", "Creature — Human Rogue", otjSVaultBusterOracle, 1, 4)
	mine := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	if !effectstest.HasEffectiveKeyword(t, g, buster, "vigilance") {
		t.Error("vigilance is printed")
	}
	effectstest.AdvanceToMain(t, g)
	effectstest.CastCatalogSpell(t, g, "Lightning Bolt", "Instant", effectstest.LightningBoltOracle, effectstest.B16TargetCard(mine))
	effectstest.PassPriorityAroundTable(t, g)
	if p, _ := effectstest.PtOf(t, g, buster); p != 1 {
		t.Fatalf("targeting your own creature is no crime: power %d", p)
	}
	otjB1BoltOpponent(t, g, opp.ID)
	if p, tt := effectstest.PtOf(t, g, buster); p != 3 || tt != 4 {
		t.Fatalf("after a crime: %d/%d, want 3/4", p, tt)
	}
	effectstest.AdvanceToMainOf(t, g, 1)
	if p, _ := effectstest.PtOf(t, g, buster); p != 1 {
		t.Errorf("a crime last turn does not count this turn: power %d", p)
	}
}

// --- Soured Springs ----------------------------------------------------------

func TestOtjSSouredSpringsEntersTappedPingsAnOpponentAndTapsForBlue(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[2]
	before := opp.Life
	land := effectstest.B12PlayFromHand(t, g, "Soured Springs", "Land — Desert", otjSSouredSpringsOracle, game.CastSpellParams{})
	if !effectstest.B16Tapped(t, g, land) {
		t.Fatal("the Desert enters tapped")
	}
	if p := effectstest.LatestPickTarget(g, me.ID); p == nil || effectstest.HasID(p.PickTargetPlayers, me.ID) || !effectstest.HasID(p.PickTargetPlayers, opp.ID) {
		t.Fatalf("the trigger asks for an opponent, never you: %+v", p)
	}
	effectstest.B16PickPlayer(t, g, me.ID, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != before-1 {
		t.Errorf("target opponent takes 1: %d → %d", before, opp.Life)
	}
	g.WithWriteLock(func() { _ = g.UntapTargetForEffect(land) })
	effectstest.B28TapForMana(t, g, me.ID, land, "B")
	if got := effectstest.PoolColors(me); len(got) != 1 || got[0] != "B" {
		t.Errorf("tapped for {B}: pool %v", got)
	}
}

// --- Spring Splasher ---------------------------------------------------------

func TestOtjSSpringSplasherShrinksACreatureOfTheDefendingPlayer(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, other := g.Seats[0], g.Seats[1], g.Seats[2]
	splasher := effectstest.B12Push(g, me.ID, "Spring Splasher", "Creature — Frog Beast", otjSSpringSplasherOracle, 2, 1)
	theirs := effectstest.B12Creature(g, opp.ID, "Ogre", "Creature — Ogre", 4, 4)
	elsewhere := effectstest.B12Creature(g, other.ID, "Other Ogre", "Creature — Ogre", 4, 4)
	mine := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(splasher, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	effectstest.LockInAttacks(t, g)
	p := effectstest.LatestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("no target prompt for the attack trigger")
	}
	for _, bad := range []uuid.UUID{elsewhere, mine} {
		if err := g.ResolvePickTarget(p.ID, me.ID, effectstest.CardRef(bad)); err == nil {
			t.Errorf("targeted %s, which the defending player does not control", effectstest.LayeredCard(t, g, bad).Name)
		}
	}
	pick := effectstest.CardRef(theirs)
	effectstest.AssistCardSettle(t, g, me.ID, &pick, true)
	if pw, tt := effectstest.PtOf(t, g, theirs); pw != 1 || tt != 4 {
		t.Errorf("the target is %d/%d, want 1/4", pw, tt)
	}
	effectstest.AdvanceToMainOf(t, g, 1)
	if pw, _ := effectstest.PtOf(t, g, theirs); pw != 4 {
		t.Errorf("until end of turn only: power %d", pw)
	}
}

// --- Stagecoach Security -----------------------------------------------------

func TestOtjSStagecoachSecurityPumpsTheTeamAndGivesVigilance(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	security := effectstest.CastHoldCreature(t, g, "Stagecoach Security", "Creature — Human Soldier", otjSStagecoachOracle, 4, 5)
	effectstest.PassPriorityAroundTable(t, g)
	if p, tt := effectstest.PtOf(t, g, bear); p != 3 || tt != 3 {
		t.Errorf("your Bear is %d/%d, want 3/3", p, tt)
	}
	if p, tt := effectstest.PtOf(t, g, security); p != 5 || tt != 6 {
		t.Errorf("the Security is %d/%d, want 5/6", p, tt)
	}
	for _, id := range []uuid.UUID{bear, security} {
		if !effectstest.HasEffectiveKeyword(t, g, id, "vigilance") {
			t.Errorf("%s lacks vigilance", effectstest.LayeredCard(t, g, id).Name)
		}
	}
	if p, _ := effectstest.PtOf(t, g, theirs); p != 2 || effectstest.HasEffectiveKeyword(t, g, theirs, "vigilance") {
		t.Error("an opponent's creature was affected")
	}
	effectstest.AdvanceToMainOf(t, g, 1)
	if p, _ := effectstest.PtOf(t, g, bear); p != 2 || effectstest.HasEffectiveKeyword(t, g, bear, "vigilance") {
		t.Error("until end of turn only")
	}
}

// --- Sterling Hound ----------------------------------------------------------

func TestOtjSSterlingHoundSurveilsTwoWhenItEnters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.SeedLibrary(me, "Top A", "Top B", "Third")
	effectstest.CastHoldCreature(t, g, "Sterling Hound", "Artifact Creature — Dog", otjSSterlingHoundOracle, 3, 2)
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.SurveilChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("entering did not queue a surveil")
	}
	if len(c.ScryCards) != 2 {
		t.Fatalf("looked at %d cards, want 2 (surveil 2)", len(c.ScryCards))
	}
	if err := g.ResolveSurveil(c.ID, me.ID, []uuid.UUID{c.ScryCards[0]}, []uuid.UUID{c.ScryCards[1]}); err != nil {
		t.Fatalf("ResolveSurveil: %v", err)
	}
	if !me.Graveyard.Contains(c.ScryCards[0]) {
		t.Error("the surveilled card did not reach the graveyard")
	}
}

// --- Sterling Keykeeper ------------------------------------------------------

func TestOtjSSterlingKeykeeperTapsANonMountCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	keeper := effectstest.B12Push(g, me.ID, "Sterling Keykeeper", "Creature — Human Mercenary", otjSSterlingKeykeeperOracle, 2, 2)
	mount := effectstest.B12Creature(g, opp.ID, "Their Steed", "Creature — Horse Mount", 3, 3)
	bear := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	effectstest.B06AddMana(me, "C", "C")
	if err := g.ActivateCatalogAbility(me.ID, keeper, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(mount)}); err == nil {
		t.Error("a Mount was accepted as the target")
	}
	effectstest.B16Activate(t, g, me.ID, keeper, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(bear)})
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.B16Tapped(t, g, bear) {
		t.Error("the Bear is tapped")
	}
	if !effectstest.B16Tapped(t, g, keeper) {
		t.Error("{T} is part of the cost")
	}
}

// --- Sterling Supplier -------------------------------------------------------

func TestOtjSSterlingSupplierPutsACounterOnAnotherCreatureYouControl(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	supplier := effectstest.CastHoldCreature(t, g, "Sterling Supplier", "Creature — Bird Soldier", otjSSterlingSupplierOracle, 3, 4)
	if !effectstest.HasEffectiveKeyword(t, g, supplier, "flying") {
		t.Error("flying is printed")
	}
	p := effectstest.LatestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("the enters trigger asked for no target")
	}
	if effectstest.HasID(p.PickTargetCards, supplier) || effectstest.HasID(p.PickTargetCards, theirs) || !effectstest.HasID(p.PickTargetCards, mine) {
		t.Fatalf("another target creature you control: %v", p.PickTargetCards)
	}
	effectstest.PickCard(t, g, me.ID, mine)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.PlusOneCounters(g, mine); got != 1 {
		t.Errorf("%d +1/+1 counters on your Bear, want 1", got)
	}
}

// --- Stingerback Terror ------------------------------------------------------

func TestOtjSStingerbackTerrorShrinksForEachCardInYourHand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.SeedLibrary(me, "One", "Two", "Three")
	terror := effectstest.B12Push(g, me.ID, "Stingerback Terror", "Creature — Scorpion Dragon", otjSStingerbackOracle, 7, 7)
	for _, kw := range []string{"flying", "trample"} {
		if !effectstest.HasEffectiveKeyword(t, g, terror, kw) {
			t.Errorf("%s is printed", kw)
		}
	}
	n := me.Hand.Size()
	if p, tt := effectstest.PtOf(t, g, terror); p != 7-n || tt != 7-n {
		t.Fatalf("with %d cards in hand: %d/%d, want %d/%d", n, p, tt, 7-n, 7-n)
	}
	g.WithWriteLock(func() { _ = g.DrawNForEffect(me.ID, 2) })
	if p, tt := effectstest.PtOf(t, g, terror); p != 5-n || tt != 5-n {
		t.Errorf("two more cards in hand: %d/%d, want %d/%d", p, tt, 5-n, 5-n)
	}
}
