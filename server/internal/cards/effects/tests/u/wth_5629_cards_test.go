package testsu

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// wthCastWithModes casts a catalog spell from the active player's hand in
// a main phase and returns the announce error.
func wthCastWithModes(t *testing.T, g *game.Game, name, typeLine, oracle string, modes []int, targets []game.TargetRef) error {
	t.Helper()
	active := g.Seats[g.Turn.ActiveSeat]
	id := effectstest.B13HandCard(active, name, typeLine, oracle)
	effectstest.AdvanceToMain(t, g)
	return g.CastSpell(active.ID, id, game.CastSpellParams{Modes: modes, Targets: targets})
}

// --- every card is in the catalog ----------------------------------------

func TestWth5629EveryCardIsRegisteredFull(t *testing.T) {
	for _, oracle := range []string{
		effectstest.WthDuskriderFalconOracle, effectstest.WthRazortoothRatsOracle, effectstest.WthBenalishKnightOracle, effectstest.WthCloudDjinnOracle,
		effectstest.WthBloodrockCyclopsOracle, effectstest.WthForiysianBrigadeOracle, effectstest.WthTolarianDrakeOracle, effectstest.WthShadowRiderOracle,
		effectstest.WthSerrasBlessingOracle, effectstest.WthUktabiEfreetOracle, effectstest.WthGallowbraidOracle, effectstest.WthMorinfenOracle,
		effectstest.WthArcticWolvesOracle, effectstest.WthFledglingDjinnOracle, effectstest.WthStripedBearsOracle, effectstest.WthMerfolkTradersOracle,
		effectstest.WthLavaHoundsOracle, effectstest.WthSageOwlOracle, effectstest.WthOdylicWraithOracle, effectstest.WthBogardanFirefiendOracle,
		effectstest.WthAbyssalGatekeeperOracle, effectstest.WthHurloonShamanOracle, effectstest.WthTolarianSerpentOracle, effectstest.WthMischievousPoltergeistOracle,
		effectstest.WthHeavyBallistaOracle, effectstest.WthSouthernPaladinOracle, effectstest.WthTouchstoneOracle, effectstest.WthInfernalTributeOracle,
		effectstest.WthArgivianFindOracle, effectstest.WthRelearnOracle, effectstest.WthArgivianRestorationOracle, effectstest.WthFitOfRageOracle,
		effectstest.WthGuidedStrikeOracle, effectstest.WthThunderboltOracle, effectstest.WthFatalBlowOracle, effectstest.WthTendrilsOfDespairOracle,
		effectstest.WthAbjureOracle, effectstest.WthGerrardsWisdomOracle, effectstest.WthBlossomingWreathOracle, effectstest.WthNaturesResurgenceOracle,
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not in the catalog", oracle)
			continue
		}
		if spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: completeness %v with %d caveats, want Full with none", spec.Name, spec.Completeness, len(spec.Caveats))
		}
	}
}

// --- printed keywords -----------------------------------------------------------

func TestWth5629KeywordCreatures(t *testing.T) {
	cases := []struct {
		name, typeLine, oracle string
		keywords               []string
	}{
		{"Duskrider Falcon", "Creature — Bird", effectstest.WthDuskriderFalconOracle, []string{"flying", "protection from black"}},
		{"Razortooth Rats", "Creature — Rat", effectstest.WthRazortoothRatsOracle, []string{"fear"}},
		{"Benalish Knight", "Creature — Human Knight", effectstest.WthBenalishKnightOracle, []string{"flash", effectstest.WthFirstStrike}},
		{"Cloud Djinn", "Creature — Djinn", effectstest.WthCloudDjinnOracle, []string{"flying"}},
		{"Tolarian Drake", "Creature — Drake", effectstest.WthTolarianDrakeOracle, []string{"flying", game.KeywordPhasing}},
		{"Gallowbraid", effectstest.WthHorrorType, effectstest.WthGallowbraidOracle, []string{"trample"}},
		{"Morinfen", effectstest.WthHorrorType, effectstest.WthMorinfenOracle, []string{"flying"}},
		{"Fledgling Djinn", "Creature — Djinn", effectstest.WthFledglingDjinnOracle, []string{"flying"}},
		{"Lava Hounds", "Creature — Dog", effectstest.WthLavaHoundsOracle, []string{"haste"}},
		{"Sage Owl", "Creature — Bird", effectstest.WthSageOwlOracle, []string{"flying"}},
		{"Odylic Wraith", "Creature — Wraith", effectstest.WthOdylicWraithOracle, []string{"swampwalk"}},
		{"Mischievous Poltergeist", "Creature — Spirit", effectstest.WthMischievousPoltergeistOracle, []string{"flying"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			id := effectstest.B12Push(g, g.Seats[0].ID, tc.name, tc.typeLine, tc.oracle, 1, 1)
			abilities := effectstest.EffectiveAbilities(t, g, id)
			for _, kw := range tc.keywords {
				if !effectstest.HasAbility(abilities, kw) {
					t.Errorf("%s has no %s", tc.name, kw)
				}
			}
		})
	}
}

// Duskrider Falcon's protection from black: Fatal Blow, a black spell,
// can't target it even after it was dealt damage this turn.
func TestWth5629DuskriderFalconCantBeTargetedByBlack(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := effectstest.Top100ActiveSeat(g), effectstest.Top100Opponent(g)
	falcon := effectstest.B12Push(g, opp.ID, "Duskrider Falcon", "Creature — Bird", effectstest.WthDuskriderFalconOracle, 1, 3)
	effectstest.HmlSetColors(g, falcon, "W")
	pinger := effectstest.B12Creature(g, me.ID, "Pinger", effectstest.WthRaiderType, 1, 1)
	effectstest.B27Damage(g, pinger, falcon, 1)
	g.RunStateChecksForTest()
	blow := uuid.New()
	me.Hand.PushTop(game.Card{
		InstanceID: blow, Name: "Fatal Blow", TypeLine: "Instant", ManaCost: "{B}", Colors: []string{"B"},
		OracleID: effectstest.WthFatalBlowOracle, Owner: me.ID, Controller: me.ID,
	})
	me.ManaPool.AddMana(game.ManaToken{Color: "B"})
	if err := g.CastSpell(me.ID, blow, game.CastSpellParams{Targets: effectstest.CardTarget(falcon)}); err == nil {
		t.Error("a black spell targeted a creature with protection from black")
	}
}

// Fear (CR 702.36b): only an artifact or black creature can block the Rats.
func TestWth5629RazortoothRatsHaveFear(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	rats := effectstest.B12Push(g, me.ID, "Razortooth Rats", "Creature — Rat", effectstest.WthRazortoothRatsOracle, 2, 1)
	bear := effectstest.B12Creature(g, opp.ID, "Bear", effectstest.WthBearType, 2, 2)
	effectstest.HmlSetColors(g, bear, "G")
	zombie := effectstest.B12Creature(g, opp.ID, "Zombie", "Creature — Zombie", 2, 2)
	effectstest.HmlSetColors(g, zombie, "B")
	effectstest.BrAttack(t, g, rats)
	if err := g.DeclareBlocker(bear, rats); err == nil {
		t.Error("a green creature blocked a creature with fear")
	}
	if err := g.DeclareBlocker(zombie, rats); err != nil {
		t.Errorf("a black creature can't block the Rats: %v", err)
	}
}

// Flash: Benalish Knight is cast outside a main phase.
func TestWth5629BenalishKnightHasFlash(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	id := effectstest.B13HandCard(me, "Benalish Knight", "Creature — Human Knight", effectstest.WthBenalishKnightOracle)
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("Benalish Knight at the beginning of combat: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(id) {
		t.Error("Benalish Knight did not resolve")
	}
}

// --- combat statics ---------------------------------------------------------------

func TestWth5629CloudDjinnBlocksOnlyFlyers(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	walker := effectstest.B12Creature(g, me.ID, "Walker", effectstest.WthRaiderType, 2, 2)
	flyer := effectstest.ApaPush(g, me.ID, me.ID, game.Card{Name: "Flyer", TypeLine: "Creature — Bird", Power: 1, Toughness: 1, Keywords: []string{"flying"}})
	djinn := effectstest.B12Push(g, opp.ID, "Cloud Djinn", "Creature — Djinn", effectstest.WthCloudDjinnOracle, 5, 4)
	effectstest.BrAttack(t, g, walker, flyer)
	if err := g.DeclareBlocker(djinn, walker); err == nil {
		t.Error("Cloud Djinn blocked a creature without flying")
	}
	if err := g.DeclareBlocker(djinn, flyer); err != nil {
		t.Errorf("Cloud Djinn can't block a flyer: %v", err)
	}
}

func TestWth5629BloodrockCyclopsMustAttack(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me, opp := g.Seats[seat], g.Seats[(seat+1)%len(g.Seats)]
	cyclops := effectstest.B12Push(g, me.ID, "Bloodrock Cyclops", "Creature — Cyclops", effectstest.WthBloodrockCyclopsOracle, 3, 3)
	effectstest.AdvanceToDeclareAttackersOf(t, g, seat)
	var re *game.AttackRequirementError
	if err := g.PassPriority(); !errors.As(err, &re) || re.Attacker != cyclops {
		t.Fatalf("pass with the Cyclops home = %v, want a refusal naming it", err)
	}
	if err := g.DeclareAttacker(cyclops, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	if err := g.PassPriority(); err != nil {
		t.Fatalf("pass with the Cyclops attacking: %v", err)
	}
}

func TestWth5629ForiysianBrigadeBlocksTwo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	brigade := effectstest.B12Push(g, opp.ID, "Foriysian Brigade", "Creature — Human Soldier", effectstest.WthForiysianBrigadeOracle, 2, 4)
	if got := effectstest.MbCapacity(t, g, brigade); got != 2 {
		t.Fatalf("capacity = %d, want 2", got)
	}
	atks := effectstest.MbAttack(t, g, me, opp, 3)
	if err := effectstest.MbBlock(g, brigade, atks...); err == nil {
		t.Fatal("the Brigade blocked three attackers")
	}
	if err := effectstest.MbBlock(g, brigade, atks[0], atks[1]); err != nil {
		t.Fatalf("blocking two: %v", err)
	}
	if got := effectstest.MbBlocked(t, g, brigade); len(got) != 2 {
		t.Errorf("blocks %d attackers, want 2", len(got))
	}
}

// Shadow Rider's flanking (CR 702.25a): a blocker without flanking gets
// -1/-1 until end of turn; one with flanking does not.
func TestWth5629ShadowRiderFlanks(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	rider := effectstest.ApaPush(g, me.ID, me.ID, game.Card{Name: "Shadow Rider", OracleID: effectstest.WthShadowRiderOracle, TypeLine: "Creature — Knight", Power: 3, Toughness: 3, Keywords: []string{"flanking"}})
	plain := effectstest.ApaPush(g, opp.ID, opp.ID, effectstest.ApaCreature("Plain", "", 2, 2))
	flanker := effectstest.ApaPush(g, opp.ID, opp.ID, game.Card{Name: "Flanker", TypeLine: "Creature — Test", Power: 2, Toughness: 2, Keywords: []string{"flanking"}})
	g.WithWriteLock(func() {
		for i, b := range []uuid.UUID{plain, flanker} {
			g.EmitEvent(game.Event{Kind: game.EventBlock, Actor: opp.ID, CardID: b, Target: rider, Amount: i + 1})
		}
	})
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.ApaLive(g, plain).CurrentToughness(); got != 1 {
		t.Errorf("the blocker without flanking has toughness %d, want 1", got)
	}
	if got := effectstest.ApaLive(g, flanker).CurrentToughness(); got != 2 {
		t.Errorf("the blocker with flanking has toughness %d, want 2", got)
	}
}

func TestWth5629SerrasBlessingGivesYourCreaturesVigilance(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Serra's Blessing", "Enchantment", effectstest.WthSerrasBlessingOracle, false)
	mine := effectstest.B12Creature(g, me.ID, "Mine", effectstest.WthBearType, 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", effectstest.WthBearType, 2, 2)
	if !effectstest.HasEffectiveKeyword(t, g, mine, "vigilance") {
		t.Error("your creature has no vigilance")
	}
	if effectstest.HasEffectiveKeyword(t, g, theirs, "vigilance") {
		t.Error("an opponent's creature has vigilance")
	}
	effectstest.BrAttack(t, g, mine)
	if effectstest.HmlTapped(t, g, mine) {
		t.Error("attacking tapped a creature with vigilance")
	}
}

// Tolarian Drake phases out in its controller's next untap step, and not
// in another player's.
func TestWth5629TolarianDrakePhasesOut(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	drake := effectstest.B12Push(g, me.ID, "Tolarian Drake", "Creature — Drake", effectstest.WthTolarianDrakeOracle, 2, 4)
	effectstest.AdvanceToMainOf(t, g, (seat+1)%len(g.Seats))
	if !g.Battlefield.Contains(drake) {
		t.Fatal("it phased out during another player's untap step")
	}
	effectstest.AdvanceToMainOf(t, g, seat)
	if g.Battlefield.Contains(drake) || !g.PhasedOut.Contains(drake) {
		t.Error("it did not phase out during its controller's untap step")
	}
}

// --- cumulative upkeep ------------------------------------------------------------

func TestWth5629CumulativeUpkeepKeptOnlyWhilePaid(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, oracle, mana string
	}{
		{"Uktabi Efreet", "Creature — Efreet", effectstest.WthUktabiEfreetOracle, "{G}"},
		{"Arctic Wolves", "Creature — Wolf", effectstest.WthArcticWolvesOracle, "{C}{C}"},
		{"Gallowbraid", effectstest.WthHorrorType, effectstest.WthGallowbraidOracle, ""},
		{"Morinfen", effectstest.WthHorrorType, effectstest.WthMorinfenOracle, ""},
	} {
		t.Run(tc.name+" paid", func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[0]
			id := effectstest.B12Push(g, me.ID, tc.name, tc.typeLine, tc.oracle, 2, 2)
			effectstest.WalkToUpkeepOfSeat0(t, g)
			effectstest.PassPriorityAroundTable(t, g)
			if got := effectstest.AgeCountersOn(t, g, id); got != 1 {
				t.Fatalf("%d age counters, want 1", got)
			}
			life := me.Life
			if tc.mana != "" {
				effectstest.DcMana(t, g, me.ID, tc.mana)
			}
			effectstest.AnswerPayUnless(t, g, me.ID, true)
			if !g.Battlefield.Contains(id) {
				t.Fatalf("paying did not keep %s", tc.name)
			}
			if tc.mana == "" && me.Life != life-1 {
				t.Errorf("life %d after paying, want %d", me.Life, life-1)
			}
			if n := me.ManaPool.Total(); n != 0 {
				t.Errorf("%d mana left in the pool after paying the upkeep", n)
			}
		})
		t.Run(tc.name+" declined", func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[0]
			id := effectstest.B12Push(g, me.ID, tc.name, tc.typeLine, tc.oracle, 2, 2)
			effectstest.WalkToUpkeepOfSeat0(t, g)
			effectstest.PassPriorityAroundTable(t, g)
			effectstest.AnswerPayUnless(t, g, me.ID, false)
			if g.Battlefield.Contains(id) || !me.Graveyard.Contains(id) {
				t.Errorf("declining the upkeep did not sacrifice %s", tc.name)
			}
		})
	}
}

// --- triggered abilities ----------------------------------------------------------

func TestWth5629EntersAndDraws(t *testing.T) {
	for _, tc := range []struct{ name, typeLine, oracle string }{
		{"Striped Bears", effectstest.WthBearType, effectstest.WthStripedBearsOracle},
		{"Arctic Wolves", "Creature — Wolf", effectstest.WthArcticWolvesOracle},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			effectstest.AdvanceToMain(t, g)
			hand := me.Hand.Size()
			id := effectstest.CastHoldCreature(t, g, tc.name, tc.typeLine, tc.oracle, 2, 2)
			effectstest.PassPriorityAroundTable(t, g)
			if !g.Battlefield.Contains(id) {
				t.Fatalf("%s did not resolve", tc.name)
			}
			if got := me.Hand.Size(); got != hand+1 {
				t.Errorf("hand is %d, want %d", got, hand+1)
			}
		})
	}
}

func TestWth5629MerfolkTradersLoots(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	hand := me.Hand.Size()
	effectstest.CastHoldCreature(t, g, "Merfolk Traders", "Creature — Merfolk", effectstest.WthMerfolkTradersOracle, 1, 2)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand is %d before the discard, want %d", got, hand+1)
	}
	if n := effectstest.DiscardOwed(g, me.ID); n != 1 {
		t.Errorf("owes %d discards, want 1", n)
	}
}

func TestWth5629LavaHoundsDealFourToYou(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	life := me.Life
	effectstest.CastHoldCreature(t, g, "Lava Hounds", "Creature — Dog", effectstest.WthLavaHoundsOracle, 4, 4)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life-4 {
		t.Errorf("life %d, want %d", me.Life, life-4)
	}
}

func TestWth5629SageOwlLooksAtTheTopFour(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.CastHoldCreature(t, g, "Sage Owl", "Creature — Bird", effectstest.WthSageOwlOracle, 1, 1)
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.LookAtTopChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no look-at prompt")
	}
	if len(c.ScryCards) != 4 {
		t.Errorf("looking at %d cards, want 4", len(c.ScryCards))
	}
}

func TestWth5629OdylicWraithMakesTheDamagedPlayerDiscard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, other := effectstest.BlbB1Seats(g)
	wraith := effectstest.B12Push(g, me.ID, "Odylic Wraith", "Creature — Wraith", effectstest.WthOdylicWraithOracle, 2, 2)
	bear := effectstest.B12Creature(g, opp.ID, "Bear", effectstest.WthBearType, 2, 2)
	effectstest.D2dSetHand(opp, 3)
	effectstest.D2dSetHand(other, 3)
	effectstest.DealNoncombatDamage(g, me.ID, wraith, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.DiscardOwed(g, opp.ID); n != 0 {
		t.Fatalf("damage to a creature asked its controller to discard %d", n)
	}
	effectstest.DealCombatDamage(g, me.ID, wraith, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.DiscardOwed(g, opp.ID); n != 1 {
		t.Fatalf("the damaged player owes %d discards, want 1", n)
	}
	if n := effectstest.DiscardOwed(g, other.ID); n != 0 {
		t.Errorf("a player not dealt damage owes %d discards", n)
	}
}

func TestWth5629FledglingDjinnDamagesYouEachUpkeep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B12Push(g, me.ID, "Fledgling Djinn", "Creature — Djinn", effectstest.WthFledglingDjinnOracle, 2, 2)
	effectstest.WalkToUpkeepOfSeat0(t, g)
	life := me.Life
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life-1 {
		t.Errorf("life %d, want %d", me.Life, life-1)
	}
}

func TestWth5629TolarianSerpentMillsSevenEachUpkeep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B12Push(g, me.ID, "Tolarian Serpent", "Creature — Serpent", effectstest.WthTolarianSerpentOracle, 7, 7)
	effectstest.WalkToUpkeepOfSeat0(t, g)
	library, graveyard := me.Library.Size(), me.Graveyard.Size()
	effectstest.PassPriorityAroundTable(t, g)
	if got := library - me.Library.Size(); got != 7 {
		t.Errorf("milled %d, want 7", got)
	}
	if got := me.Graveyard.Size() - graveyard; got != 7 {
		t.Errorf("%d cards reached the graveyard, want 7", got)
	}
}

func TestWth5629BogardanFirefiendDealsTwoWhenItDies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	fiend := effectstest.B12Push(g, me.ID, "Bogardan Firefiend", "Creature — Elemental Spirit", effectstest.WthBogardanFirefiendOracle, 2, 1)
	ogre := effectstest.B12Creature(g, opp.ID, "Ogre", "Creature — Ogre", 3, 3)
	effectstest.B27Kill(g, fiend)
	g.RunStateChecksForTest()
	effectstest.AnswerPickTarget(t, g, ogre)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.FindCardAnywhere(t, g, ogre).DamageMarked; got != 2 {
		t.Errorf("the target has %d damage, want 2", got)
	}
}

// Each player has exactly one candidate, so no one is asked; every
// candidate, the controller's included, is sacrificed.
func TestWth5629DiesEdicts(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, oracle string
		victimType             string
	}{
		{"Abyssal Gatekeeper", "Creature — Horror", effectstest.WthAbyssalGatekeeperOracle, effectstest.WthBearType},
		{"Hurloon Shaman", "Creature — Minotaur Shaman", effectstest.WthHurloonShamanOracle, "Basic Land — Forest"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			id := effectstest.B12Push(g, g.Seats[0].ID, tc.name, tc.typeLine, tc.oracle, 2, 2)
			var victims []uuid.UUID
			for _, s := range g.Seats {
				victims = append(victims, effectstest.B12Permanent(g, s.ID, "Victim", tc.victimType))
			}
			effectstest.B27Kill(g, id)
			g.RunStateChecksForTest()
			effectstest.PassPriorityAroundTable(t, g)
			for i, v := range victims {
				if g.Battlefield.Contains(v) {
					t.Errorf("seat %d kept its permanent", i)
				}
			}
		})
	}
}

// --- activated abilities -------------------------------------------------------------

func TestWth5629MischievousPoltergeistRegeneratesForLife(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := effectstest.Top100ActiveSeat(g)
	ghost := effectstest.B12Push(g, me.ID, "Mischievous Poltergeist", "Creature — Spirit", effectstest.WthMischievousPoltergeistOracle, 1, 1)
	life := me.Life
	effectstest.ApaActivate(t, g, me, ghost, 0, game.ActivateAbilityParams{})
	if me.Life != life-1 {
		t.Errorf("life %d after paying, want %d", me.Life, life-1)
	}
	effectstest.B27Kill(g, ghost)
	if !g.Battlefield.Contains(ghost) {
		t.Error("the shielded Poltergeist was destroyed")
	}
}

func TestWth5629HeavyBallistaShootsOnlyACombatant(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	ballista := effectstest.B12Push(g, me.ID, "Heavy Ballista", "Creature — Human Soldier", effectstest.WthHeavyBallistaOracle, 2, 3)
	raider := effectstest.B12Creature(g, me.ID, "Raider", effectstest.WthRaiderType, 2, 3)
	idle := effectstest.B12Creature(g, me.ID, "Idle", effectstest.WthRaiderType, 2, 3)
	effectstest.BrAttack(t, g, raider)
	if err := g.ActivateCatalogAbility(me.ID, ballista, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardTarget(idle)}); err == nil {
		t.Fatal("the Ballista targeted a creature that isn't in combat")
	}
	effectstest.ApaActivate(t, g, me, ballista, 0, game.ActivateAbilityParams{Targets: effectstest.CardTarget(raider)})
	if got := effectstest.FindCardAnywhere(t, g, raider).DamageMarked; got != 2 {
		t.Errorf("the attacker has %d damage, want 2", got)
	}
	if !effectstest.HmlTapped(t, g, ballista) {
		t.Error("the Ballista did not tap")
	}
}

func TestWth5629SouthernPaladinDestroysOnlyRed(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := effectstest.Top100ActiveSeat(g), effectstest.Top100Opponent(g)
	paladin := effectstest.B12Push(g, me.ID, "Southern Paladin", "Creature — Human Knight", effectstest.WthSouthernPaladinOracle, 3, 3)
	goblin := effectstest.B12Creature(g, opp.ID, "Goblin", "Creature — Goblin", 1, 1)
	effectstest.HmlSetColors(g, goblin, "R")
	elf := effectstest.B12Creature(g, opp.ID, "Elf", "Creature — Elf", 1, 1)
	effectstest.HmlSetColors(g, elf, "G")
	effectstest.ApaMana(me, "W", "W")
	if err := g.ActivateCatalogAbility(me.ID, paladin, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardTarget(elf)}); err == nil {
		t.Fatal("the Paladin targeted a green creature")
	}
	effectstest.ApaActivate(t, g, me, paladin, 0, game.ActivateAbilityParams{Targets: effectstest.CardTarget(goblin)})
	if g.Battlefield.Contains(goblin) {
		t.Error("the red creature survived")
	}
}

func TestWth5629TouchstoneTapsAnArtifactYouDontControl(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := effectstest.Top100ActiveSeat(g), effectstest.Top100Opponent(g)
	stone := effectstest.PushCatalogPermanent(g, me.ID, "Touchstone", "Artifact", effectstest.WthTouchstoneOracle, false)
	mine := effectstest.B12Permanent(g, me.ID, "My Rock", "Artifact")
	theirs := effectstest.B12Permanent(g, opp.ID, "Their Rock", "Artifact")
	if err := g.ActivateCatalogAbility(me.ID, stone, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardTarget(mine)}); err == nil {
		t.Fatal("Touchstone targeted its controller's own artifact")
	}
	effectstest.ApaActivate(t, g, me, stone, 0, game.ActivateAbilityParams{Targets: effectstest.CardTarget(theirs)})
	if !effectstest.HmlTapped(t, g, theirs) {
		t.Error("the opponent's artifact is untapped")
	}
}

func TestWth5629InfernalTributeSacrificesANontokenToDraw(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := effectstest.Top100ActiveSeat(g)
	tribute := effectstest.PushCatalogPermanent(g, me.ID, "Infernal Tribute", "Enchantment", effectstest.WthInfernalTributeOracle, false)
	token := effectstest.B12Creature(g, me.ID, "Spirit", "Token Creature — Spirit", 1, 1)
	bear := effectstest.B12Creature(g, me.ID, "Bear", effectstest.WthBearType, 2, 2)
	effectstest.ApaMana(me, "C", "C")
	if err := g.ActivateCatalogAbility(me.ID, tribute, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{token}}); err == nil {
		t.Fatal("a token paid the sacrifice")
	}
	hand := me.Hand.Size()
	effectstest.ApaActivate(t, g, me, tribute, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{bear}})
	if g.Battlefield.Contains(bear) {
		t.Error("the Bear was not sacrificed")
	}
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand is %d, want %d", got, hand+1)
	}
}

// --- the spells --------------------------------------------------------------------

func TestWth5629GraveyardReturnsToHand(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, oracle string
		good, bad              string
	}{
		{"Argivian Find", "Instant", effectstest.WthArgivianFindOracle, "Enchantment", "Creature — Bear"},
		{"Relearn", "Sorcery", effectstest.WthRelearnOracle, "Instant", "Artifact"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			effectstest.AdvanceToMain(t, g)
			me := effectstest.Top100ActiveSeat(g)
			good := effectstest.WthGraveyardCard(me, "Good", tc.good)
			bad := effectstest.WthGraveyardCard(me, "Bad", tc.bad)
			if err := effectstest.CastCatalogSpellErr(t, g, tc.name, tc.typeLine, tc.oracle, effectstest.CardTarget(bad)); err == nil {
				t.Fatalf("%s targeted a %s card", tc.name, tc.bad)
			}
			effectstest.CastCatalogSpell(t, g, tc.name, tc.typeLine, tc.oracle, effectstest.CardTarget(good))
			effectstest.PassPriorityAroundTable(t, g)
			if !me.Hand.Contains(good) || me.Graveyard.Contains(good) {
				t.Error("the card did not return to hand")
			}
		})
	}
}

func TestWth5629ArgivianRestorationReturnsAnArtifact(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := effectstest.Top100ActiveSeat(g)
	bear := effectstest.WthGraveyardCard(me, "Bear", effectstest.WthBearType)
	if err := effectstest.CastCatalogSpellErr(t, g, "Argivian Restoration", "Sorcery", effectstest.WthArgivianRestorationOracle, effectstest.CardTarget(bear)); err == nil {
		t.Fatal("Argivian Restoration targeted a creature card")
	}
	rock := effectstest.WthGraveyardCard(me, "Old Rock", "Artifact")
	effectstest.CastCatalogSpell(t, g, "Argivian Restoration", "Sorcery", effectstest.WthArgivianRestorationOracle, effectstest.CardTarget(rock))
	effectstest.PassPriorityAroundTable(t, g)
	if _, ok := effectstest.VisBattlefieldNamed(g, me.ID, "Old Rock"); !ok {
		t.Error("the artifact card did not return to the battlefield under your control")
	}
}

func TestWth5629PumpsGrantFirstStrike(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, oracle string
		p, tough, draws        int
	}{
		{"Fit of Rage", "Sorcery", effectstest.WthFitOfRageOracle, 5, 5, 0},
		{"Guided Strike", "Instant", effectstest.WthGuidedStrikeOracle, 3, 2, 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			effectstest.AdvanceToMain(t, g)
			me := effectstest.Top100ActiveSeat(g)
			bear := effectstest.B12Creature(g, me.ID, "Bear", effectstest.WthBearType, 2, 2)
			hand := me.Hand.Size()
			effectstest.CastCatalogSpell(t, g, tc.name, tc.typeLine, tc.oracle, effectstest.CardTarget(bear))
			effectstest.PassPriorityAroundTable(t, g)
			if p, tough := effectstest.PtOf(t, g, bear); p != tc.p || tough != tc.tough {
				t.Errorf("the Bear is %d/%d, want %d/%d", p, tough, tc.p, tc.tough)
			}
			if !effectstest.HasEffectiveKeyword(t, g, bear, effectstest.WthFirstStrike) {
				t.Error("the Bear has no first strike")
			}
			if got := me.Hand.Size() - hand; got != tc.draws {
				t.Errorf("drew %d, want %d", got, tc.draws)
			}
			effectstest.AdvanceToMainOf(t, g, (g.Turn.ActiveSeat+1)%len(g.Seats))
			if effectstest.HasEffectiveKeyword(t, g, bear, effectstest.WthFirstStrike) {
				t.Error("first strike outlasted the turn")
			}
		})
	}
}

func TestWth5629ThunderboltEachMode(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	opp := effectstest.Top100Opponent(g)
	life := opp.Life
	effectstest.CastModal(t, g, "Thunderbolt", "Instant", effectstest.WthThunderboltOracle, []int{0}, []game.TargetRef{effectstest.PlayerTarget(opp.ID)})
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-3 {
		t.Errorf("the player is at %d, want %d", opp.Life, life-3)
	}
	walker := effectstest.B12Creature(g, opp.ID, "Walker", effectstest.WthRaiderType, 5, 5)
	flyer := effectstest.ApaPush(g, opp.ID, opp.ID, game.Card{Name: "Flyer", TypeLine: "Creature — Bird", Power: 5, Toughness: 5, Keywords: []string{"flying"}})
	if err := wthCastWithModes(t, g, "Thunderbolt", "Instant", effectstest.WthThunderboltOracle, []int{1}, effectstest.CardTarget(walker)); err == nil {
		t.Fatal("the flying bullet targeted a creature without flying")
	}
	effectstest.CastModal(t, g, "Thunderbolt", "Instant", effectstest.WthThunderboltOracle, []int{1}, effectstest.CardTarget(flyer))
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.FindCardAnywhere(t, g, flyer).DamageMarked; got != 4 {
		t.Errorf("the flyer has %d damage, want 4", got)
	}
}

func TestWth5629FatalBlowDestroysADamagedCreaturePastRegeneration(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	opp := effectstest.Top100Opponent(g)
	fresh := effectstest.B12Creature(g, opp.ID, "Fresh", effectstest.WthBearType, 2, 2)
	boa := visShield(t, g, opp, 2, 3)
	effectstest.B27Damage(g, fresh, boa, 1)
	g.RunStateChecksForTest()
	if err := effectstest.CastCatalogSpellErr(t, g, "Fatal Blow", "Instant", effectstest.WthFatalBlowOracle, effectstest.CardTarget(fresh)); err == nil {
		t.Fatal("Fatal Blow targeted a creature dealt no damage this turn")
	}
	effectstest.CastCatalogSpell(t, g, "Fatal Blow", "Instant", effectstest.WthFatalBlowOracle, effectstest.CardTarget(boa))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(boa) {
		t.Error("a regeneration shield saved Fatal Blow's target")
	}
}

func TestWth5629TendrilsOfDespairSacrificesAndMakesTheOpponentDiscardTwo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := effectstest.Top100ActiveSeat(g), effectstest.Top100Opponent(g)
	victim := effectstest.B12Creature(g, me.ID, "Victim", effectstest.WthBearType, 2, 2)
	effectstest.D2dSetHand(opp, 4)
	id := effectstest.B13HandCard(me, "Tendrils of Despair", "Sorcery", effectstest.WthTendrilsOfDespairOracle)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Targets: []game.TargetRef{effectstest.PlayerTarget(opp.ID)}}); err == nil {
		t.Fatal("cast with no sacrifice")
	}
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Targets: []game.TargetRef{effectstest.PlayerTarget(opp.ID)}, SacrificeIDs: []uuid.UUID{victim}}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(victim) {
		t.Error("the creature was not sacrificed")
	}
	if n := effectstest.DiscardOwed(g, opp.ID); n != 2 {
		t.Errorf("the opponent owes %d discards, want 2", n)
	}
}

func TestWth5629AbjureSacrificesABluePermanentToCounter(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	red := effectstest.B12Creature(g, me.ID, "Red", "Creature — Goblin", 1, 1)
	effectstest.HmlSetColors(g, red, "R")
	blue := effectstest.B12Creature(g, me.ID, "Blue", "Creature — Merfolk", 1, 1)
	effectstest.HmlSetColors(g, blue, "U")
	bolt := effectstest.B13OpponentCasts(t, g, opp, "Their Bolt", "Instant", "", "{R}", nil)
	id := effectstest.B13HandCard(me, "Abjure", "Instant", effectstest.WthAbjureOracle)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Targets: effectstest.B16TargetCard(bolt), SacrificeIDs: []uuid.UUID{red}}); err == nil {
		t.Fatal("a red permanent paid the sacrifice")
	}
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Targets: effectstest.B16TargetCard(bolt), SacrificeIDs: []uuid.UUID{blue}}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(blue) {
		t.Error("the blue permanent was not sacrificed")
	}
	if !opp.Graveyard.Contains(bolt) {
		t.Error("the spell was not countered")
	}
}

func TestWth5629GerrardsWisdomGainsTwoPerCardInHand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := effectstest.Top100ActiveSeat(g)
	effectstest.D2dSetHand(me, 3)
	life := me.Life
	effectstest.CastCatalogSpell(t, g, "Gerrard's Wisdom", "Sorcery", effectstest.WthGerrardsWisdomOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life+6 {
		t.Errorf("life %d, want %d", me.Life, life+6)
	}
}

func TestWth5629BlossomingWreathCountsCreatureCards(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := effectstest.Top100ActiveSeat(g)
	effectstest.PushGraveyardCardForTest(me, "One")
	effectstest.PushGraveyardCardForTest(me, "Two")
	effectstest.WthGraveyardCard(me, "Rock", "Artifact")
	life := me.Life
	effectstest.CastCatalogSpell(t, g, "Blossoming Wreath", "Instant", effectstest.WthBlossomingWreathOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life+2 {
		t.Errorf("life %d, want %d", me.Life, life+2)
	}
}

func TestWth5629NaturesResurgenceDrawsPerCreatureCard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := effectstest.Top100ActiveSeat(g), effectstest.Top100Opponent(g)
	effectstest.PushGraveyardCardForTest(me, "One")
	effectstest.PushGraveyardCardForTest(me, "Two")
	effectstest.WthGraveyardCard(me, "Rock", "Artifact")
	effectstest.PushGraveyardCardForTest(opp, "Three")
	hands := map[uuid.UUID]int{}
	for _, s := range g.Seats {
		hands[s.ID] = s.Hand.Size()
	}
	effectstest.CastCatalogSpell(t, g, "Nature's Resurgence", "Sorcery", effectstest.WthNaturesResurgenceOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	want := map[uuid.UUID]int{me.ID: 2, opp.ID: 1}
	for _, s := range g.Seats {
		if got := s.Hand.Size() - hands[s.ID]; got != want[s.ID] {
			t.Errorf("%s drew %d, want %d", s.Name, got, want[s.ID])
		}
	}
}
