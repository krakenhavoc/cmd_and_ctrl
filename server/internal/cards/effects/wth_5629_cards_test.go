package effects

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// wth_5629_cards_test.go — a batch of Weatherlight cards (#5629). Every
// test plays the real catalog entry through the engine.

const (
	wthDuskriderFalconOracle        = "d87ab916-5d49-4840-8d5b-805af800a83c"
	wthRazortoothRatsOracle         = "72e9a207-dad9-4c51-8040-e90a36bd35e6"
	wthBenalishKnightOracle         = "32b401e9-163f-4917-a728-fc63b25ef602"
	wthCloudDjinnOracle             = "4914d38d-ed6b-4531-b72c-06f0968c44e9"
	wthBloodrockCyclopsOracle       = "06a3fe24-7e1f-4f98-ba2f-dbb5d38f4b9d"
	wthForiysianBrigadeOracle       = "7261dc6d-7669-4682-977b-1304f57a201d"
	wthTolarianDrakeOracle          = "d999f2a3-5cb4-4f12-a79a-c5db0a300620"
	wthShadowRiderOracle            = "a518b31a-7e9a-460d-839f-39cb1de4f570"
	wthSerrasBlessingOracle         = "49cdd05c-feeb-4c24-9d88-069f5d2e08c3"
	wthUktabiEfreetOracle           = "2fd9a37f-151a-4501-9d8e-bc67cc902592"
	wthGallowbraidOracle            = "4a4418fa-7574-4027-aa09-4ca789be15ba"
	wthMorinfenOracle               = "d69f3aaf-8462-432d-afaf-4176ac04962b"
	wthArcticWolvesOracle           = "c9e7218f-07f2-46f3-bc63-eb5c0d32538b"
	wthFledglingDjinnOracle         = "2c73ef77-ae58-401f-8747-538c4cd075d0"
	wthStripedBearsOracle           = "ac856e91-1b88-464b-9420-dc311ce814ea"
	wthMerfolkTradersOracle         = "8af6646a-10ba-448b-bd5e-48feb70de5e5"
	wthLavaHoundsOracle             = "86c0dc85-d146-4dfa-819e-f78835bfffff"
	wthSageOwlOracle                = "d47fc902-51f7-4ad1-8ee4-c973e32192b8"
	wthOdylicWraithOracle           = "613fa223-856e-49c4-8952-224895e00b5e"
	wthBogardanFirefiendOracle      = "1d61f3a7-5861-479b-8312-f2d308c35b68"
	wthAbyssalGatekeeperOracle      = "4954181c-cb46-4cac-adc3-2c8b7693a211"
	wthHurloonShamanOracle          = "9fc74fc9-bb02-4d9e-bb8e-f44f219d8946"
	wthTolarianSerpentOracle        = "684e18bf-1390-4d12-8e7a-e8562db64dbb"
	wthMischievousPoltergeistOracle = "c12b5b9a-a46e-4635-8603-62c2ca5ee58e"
	wthHeavyBallistaOracle          = "b4219836-c802-49e2-9a66-f03ca8f13895"
	wthSouthernPaladinOracle        = "cab96aa8-13e3-4dd4-9d44-9e298f807388"
	wthTouchstoneOracle             = "92382bdb-84cf-4456-8ca1-8a19badae7ad"
	wthInfernalTributeOracle        = "a8bf79d2-29fd-4d2f-ac07-2e61d3886bc0"
	wthArgivianFindOracle           = "0b6c74a7-d7d3-4c05-84b1-bcb580bb476f"
	wthRelearnOracle                = "13cb7d87-ff12-4742-8f54-f3175536db40"
	wthArgivianRestorationOracle    = "d887919a-68d5-455d-977c-4f1f0b4b3141"
	wthFitOfRageOracle              = "69d08521-74e1-4215-9ed4-3f12137332a4"
	wthGuidedStrikeOracle           = "a215e813-df7e-4f21-9718-eb264eb62813"
	wthThunderboltOracle            = "6ffae2fa-511a-45bf-93c4-5d53c85324f5"
	wthFatalBlowOracle              = "16774611-c004-4383-b4e1-4a43a9b1f4aa"
	wthTendrilsOfDespairOracle      = "3b00f415-2496-4333-be89-601fd837fe7c"
	wthAbjureOracle                 = "dd97e457-07eb-4918-b780-ae7ff42a37e7"
	wthGerrardsWisdomOracle         = "3e30e8f2-f437-4620-a3bc-dd9c29ae6570"
	wthBlossomingWreathOracle       = "fbd99a88-8926-455e-8e23-58fed6d14719"
	wthNaturesResurgenceOracle      = "276f49ee-7cf3-4a4d-9b14-74aedcbef69f"

	wthBearType    = "Creature — Bear"
	wthRaiderType  = "Creature — Human"
	wthHorrorType  = "Legendary Creature — Phyrexian Horror"
	wthFirstStrike = "first strike"
)

// wthGraveyardCard puts a card of the given type line into p's graveyard.
func wthGraveyardCard(p *game.Player, name, typeLine string) uuid.UUID {
	id := uuid.New()
	p.Graveyard.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, Owner: p.ID, Controller: p.ID})
	return id
}

// wthCastWithModes casts a catalog spell from the active player's hand in
// a main phase and returns the announce error.
func wthCastWithModes(t *testing.T, g *game.Game, name, typeLine, oracle string, modes []int, targets []game.TargetRef) error {
	t.Helper()
	active := g.Seats[g.Turn.ActiveSeat]
	id := b13HandCard(active, name, typeLine, oracle)
	advanceToMain(t, g)
	return g.CastSpell(active.ID, id, game.CastSpellParams{Modes: modes, Targets: targets})
}

// --- every card is in the catalog ----------------------------------------

func TestWth5629EveryCardIsRegisteredFull(t *testing.T) {
	for _, oracle := range []string{
		wthDuskriderFalconOracle, wthRazortoothRatsOracle, wthBenalishKnightOracle, wthCloudDjinnOracle,
		wthBloodrockCyclopsOracle, wthForiysianBrigadeOracle, wthTolarianDrakeOracle, wthShadowRiderOracle,
		wthSerrasBlessingOracle, wthUktabiEfreetOracle, wthGallowbraidOracle, wthMorinfenOracle,
		wthArcticWolvesOracle, wthFledglingDjinnOracle, wthStripedBearsOracle, wthMerfolkTradersOracle,
		wthLavaHoundsOracle, wthSageOwlOracle, wthOdylicWraithOracle, wthBogardanFirefiendOracle,
		wthAbyssalGatekeeperOracle, wthHurloonShamanOracle, wthTolarianSerpentOracle, wthMischievousPoltergeistOracle,
		wthHeavyBallistaOracle, wthSouthernPaladinOracle, wthTouchstoneOracle, wthInfernalTributeOracle,
		wthArgivianFindOracle, wthRelearnOracle, wthArgivianRestorationOracle, wthFitOfRageOracle,
		wthGuidedStrikeOracle, wthThunderboltOracle, wthFatalBlowOracle, wthTendrilsOfDespairOracle,
		wthAbjureOracle, wthGerrardsWisdomOracle, wthBlossomingWreathOracle, wthNaturesResurgenceOracle,
	} {
		spec, ok := Lookup(oracle)
		if !ok {
			t.Errorf("%s is not in the catalog", oracle)
			continue
		}
		if spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
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
		{"Duskrider Falcon", "Creature — Bird", wthDuskriderFalconOracle, []string{"flying", "protection from black"}},
		{"Razortooth Rats", "Creature — Rat", wthRazortoothRatsOracle, []string{"fear"}},
		{"Benalish Knight", "Creature — Human Knight", wthBenalishKnightOracle, []string{"flash", wthFirstStrike}},
		{"Cloud Djinn", "Creature — Djinn", wthCloudDjinnOracle, []string{"flying"}},
		{"Tolarian Drake", "Creature — Drake", wthTolarianDrakeOracle, []string{"flying", game.KeywordPhasing}},
		{"Gallowbraid", wthHorrorType, wthGallowbraidOracle, []string{"trample"}},
		{"Morinfen", wthHorrorType, wthMorinfenOracle, []string{"flying"}},
		{"Fledgling Djinn", "Creature — Djinn", wthFledglingDjinnOracle, []string{"flying"}},
		{"Lava Hounds", "Creature — Dog", wthLavaHoundsOracle, []string{"haste"}},
		{"Sage Owl", "Creature — Bird", wthSageOwlOracle, []string{"flying"}},
		{"Odylic Wraith", "Creature — Wraith", wthOdylicWraithOracle, []string{"swampwalk"}},
		{"Mischievous Poltergeist", "Creature — Spirit", wthMischievousPoltergeistOracle, []string{"flying"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			id := b12Push(g, g.Seats[0].ID, tc.name, tc.typeLine, tc.oracle, 1, 1)
			abilities := effectiveAbilities(t, g, id)
			for _, kw := range tc.keywords {
				if !hasAbility(abilities, kw) {
					t.Errorf("%s has no %s", tc.name, kw)
				}
			}
		})
	}
}

// Duskrider Falcon's protection from black: Fatal Blow, a black spell,
// can't target it even after it was dealt damage this turn.
func TestWth5629DuskriderFalconCantBeTargetedByBlack(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me, opp := top100ActiveSeat(g), top100Opponent(g)
	falcon := b12Push(g, opp.ID, "Duskrider Falcon", "Creature — Bird", wthDuskriderFalconOracle, 1, 3)
	hmlSetColors(g, falcon, "W")
	pinger := b12Creature(g, me.ID, "Pinger", wthRaiderType, 1, 1)
	b27Damage(g, pinger, falcon, 1)
	g.RunStateChecksForTest()
	blow := uuid.New()
	me.Hand.PushTop(game.Card{
		InstanceID: blow, Name: "Fatal Blow", TypeLine: "Instant", ManaCost: "{B}", Colors: []string{"B"},
		OracleID: wthFatalBlowOracle, Owner: me.ID, Controller: me.ID,
	})
	me.ManaPool.AddMana(game.ManaToken{Color: "B"})
	if err := g.CastSpell(me.ID, blow, game.CastSpellParams{Targets: cardTarget(falcon)}); err == nil {
		t.Error("a black spell targeted a creature with protection from black")
	}
}

// Fear (CR 702.36b): only an artifact or black creature can block the Rats.
func TestWth5629RazortoothRatsHaveFear(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	rats := b12Push(g, me.ID, "Razortooth Rats", "Creature — Rat", wthRazortoothRatsOracle, 2, 1)
	bear := b12Creature(g, opp.ID, "Bear", wthBearType, 2, 2)
	hmlSetColors(g, bear, "G")
	zombie := b12Creature(g, opp.ID, "Zombie", "Creature — Zombie", 2, 2)
	hmlSetColors(g, zombie, "B")
	brAttack(t, g, rats)
	if err := g.DeclareBlocker(bear, rats); err == nil {
		t.Error("a green creature blocked a creature with fear")
	}
	if err := g.DeclareBlocker(zombie, rats); err != nil {
		t.Errorf("a black creature can't block the Rats: %v", err)
	}
}

// Flash: Benalish Knight is cast outside a main phase.
func TestWth5629BenalishKnightHasFlash(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	id := b13HandCard(me, "Benalish Knight", "Creature — Human Knight", wthBenalishKnightOracle)
	advanceTo(t, g, game.StepBeginCombat)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("Benalish Knight at the beginning of combat: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(id) {
		t.Error("Benalish Knight did not resolve")
	}
}

// --- combat statics ---------------------------------------------------------------

func TestWth5629CloudDjinnBlocksOnlyFlyers(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	walker := b12Creature(g, me.ID, "Walker", wthRaiderType, 2, 2)
	flyer := apaPush(g, me.ID, me.ID, game.Card{Name: "Flyer", TypeLine: "Creature — Bird", Power: 1, Toughness: 1, Keywords: []string{"flying"}})
	djinn := b12Push(g, opp.ID, "Cloud Djinn", "Creature — Djinn", wthCloudDjinnOracle, 5, 4)
	brAttack(t, g, walker, flyer)
	if err := g.DeclareBlocker(djinn, walker); err == nil {
		t.Error("Cloud Djinn blocked a creature without flying")
	}
	if err := g.DeclareBlocker(djinn, flyer); err != nil {
		t.Errorf("Cloud Djinn can't block a flyer: %v", err)
	}
}

func TestWth5629BloodrockCyclopsMustAttack(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me, opp := g.Seats[seat], g.Seats[(seat+1)%len(g.Seats)]
	cyclops := b12Push(g, me.ID, "Bloodrock Cyclops", "Creature — Cyclops", wthBloodrockCyclopsOracle, 3, 3)
	advanceToDeclareAttackersOf(t, g, seat)
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
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	brigade := b12Push(g, opp.ID, "Foriysian Brigade", "Creature — Human Soldier", wthForiysianBrigadeOracle, 2, 4)
	if got := mbCapacity(t, g, brigade); got != 2 {
		t.Fatalf("capacity = %d, want 2", got)
	}
	atks := mbAttack(t, g, me, opp, 3)
	if err := mbBlock(g, brigade, atks...); err == nil {
		t.Fatal("the Brigade blocked three attackers")
	}
	if err := mbBlock(g, brigade, atks[0], atks[1]); err != nil {
		t.Fatalf("blocking two: %v", err)
	}
	if got := mbBlocked(t, g, brigade); len(got) != 2 {
		t.Errorf("blocks %d attackers, want 2", len(got))
	}
}

// Shadow Rider's flanking (CR 702.25a): a blocker without flanking gets
// -1/-1 until end of turn; one with flanking does not.
func TestWth5629ShadowRiderFlanks(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	rider := apaPush(g, me.ID, me.ID, game.Card{Name: "Shadow Rider", OracleID: wthShadowRiderOracle, TypeLine: "Creature — Knight", Power: 3, Toughness: 3, Keywords: []string{"flanking"}})
	plain := apaPush(g, opp.ID, opp.ID, apaCreature("Plain", "", 2, 2))
	flanker := apaPush(g, opp.ID, opp.ID, game.Card{Name: "Flanker", TypeLine: "Creature — Test", Power: 2, Toughness: 2, Keywords: []string{"flanking"}})
	g.WithWriteLock(func() {
		for i, b := range []uuid.UUID{plain, flanker} {
			g.EmitEvent(game.Event{Kind: game.EventBlock, Actor: opp.ID, CardID: b, Target: rider, Amount: i + 1})
		}
	})
	passPriorityAroundTable(t, g)
	if got := apaLive(g, plain).CurrentToughness(); got != 1 {
		t.Errorf("the blocker without flanking has toughness %d, want 1", got)
	}
	if got := apaLive(g, flanker).CurrentToughness(); got != 2 {
		t.Errorf("the blocker with flanking has toughness %d, want 2", got)
	}
}

func TestWth5629SerrasBlessingGivesYourCreaturesVigilance(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushCatalogPermanent(g, me.ID, "Serra's Blessing", "Enchantment", wthSerrasBlessingOracle, false)
	mine := b12Creature(g, me.ID, "Mine", wthBearType, 2, 2)
	theirs := b12Creature(g, opp.ID, "Theirs", wthBearType, 2, 2)
	if !hasEffectiveKeyword(t, g, mine, "vigilance") {
		t.Error("your creature has no vigilance")
	}
	if hasEffectiveKeyword(t, g, theirs, "vigilance") {
		t.Error("an opponent's creature has vigilance")
	}
	brAttack(t, g, mine)
	if hmlTapped(t, g, mine) {
		t.Error("attacking tapped a creature with vigilance")
	}
}

// Tolarian Drake phases out in its controller's next untap step, and not
// in another player's.
func TestWth5629TolarianDrakePhasesOut(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	drake := b12Push(g, me.ID, "Tolarian Drake", "Creature — Drake", wthTolarianDrakeOracle, 2, 4)
	advanceToMainOf(t, g, (seat+1)%len(g.Seats))
	if !g.Battlefield.Contains(drake) {
		t.Fatal("it phased out during another player's untap step")
	}
	advanceToMainOf(t, g, seat)
	if g.Battlefield.Contains(drake) || !g.PhasedOut.Contains(drake) {
		t.Error("it did not phase out during its controller's untap step")
	}
}

// --- cumulative upkeep ------------------------------------------------------------

func TestWth5629CumulativeUpkeepKeptOnlyWhilePaid(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, oracle, mana string
	}{
		{"Uktabi Efreet", "Creature — Efreet", wthUktabiEfreetOracle, "{G}"},
		{"Arctic Wolves", "Creature — Wolf", wthArcticWolvesOracle, "{C}{C}"},
		{"Gallowbraid", wthHorrorType, wthGallowbraidOracle, ""},
		{"Morinfen", wthHorrorType, wthMorinfenOracle, ""},
	} {
		t.Run(tc.name+" paid", func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[0]
			id := b12Push(g, me.ID, tc.name, tc.typeLine, tc.oracle, 2, 2)
			walkToUpkeepOfSeat0(t, g)
			passPriorityAroundTable(t, g)
			if got := ageCountersOn(t, g, id); got != 1 {
				t.Fatalf("%d age counters, want 1", got)
			}
			life := me.Life
			if tc.mana != "" {
				dcMana(t, g, me.ID, tc.mana)
			}
			answerPayUnless(t, g, me.ID, true)
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
			g := newCatalogGame(t)
			me := g.Seats[0]
			id := b12Push(g, me.ID, tc.name, tc.typeLine, tc.oracle, 2, 2)
			walkToUpkeepOfSeat0(t, g)
			passPriorityAroundTable(t, g)
			answerPayUnless(t, g, me.ID, false)
			if g.Battlefield.Contains(id) || !me.Graveyard.Contains(id) {
				t.Errorf("declining the upkeep did not sacrifice %s", tc.name)
			}
		})
	}
}

// --- triggered abilities ----------------------------------------------------------

func TestWth5629EntersAndDraws(t *testing.T) {
	for _, tc := range []struct{ name, typeLine, oracle string }{
		{"Striped Bears", wthBearType, wthStripedBearsOracle},
		{"Arctic Wolves", "Creature — Wolf", wthArcticWolvesOracle},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			advanceToMain(t, g)
			hand := me.Hand.Size()
			id := castHoldCreature(t, g, tc.name, tc.typeLine, tc.oracle, 2, 2)
			passPriorityAroundTable(t, g)
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
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	hand := me.Hand.Size()
	castHoldCreature(t, g, "Merfolk Traders", "Creature — Merfolk", wthMerfolkTradersOracle, 1, 2)
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand is %d before the discard, want %d", got, hand+1)
	}
	if n := discardOwed(g, me.ID); n != 1 {
		t.Errorf("owes %d discards, want 1", n)
	}
}

func TestWth5629LavaHoundsDealFourToYou(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	life := me.Life
	castHoldCreature(t, g, "Lava Hounds", "Creature — Dog", wthLavaHoundsOracle, 4, 4)
	passPriorityAroundTable(t, g)
	if me.Life != life-4 {
		t.Errorf("life %d, want %d", me.Life, life-4)
	}
}

func TestWth5629SageOwlLooksAtTheTopFour(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	castHoldCreature(t, g, "Sage Owl", "Creature — Bird", wthSageOwlOracle, 1, 1)
	passPriorityAroundTable(t, g)
	c := lookAtTopChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no look-at prompt")
	}
	if len(c.ScryCards) != 4 {
		t.Errorf("looking at %d cards, want 4", len(c.ScryCards))
	}
}

func TestWth5629OdylicWraithMakesTheDamagedPlayerDiscard(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, other := blbB1Seats(g)
	wraith := b12Push(g, me.ID, "Odylic Wraith", "Creature — Wraith", wthOdylicWraithOracle, 2, 2)
	bear := b12Creature(g, opp.ID, "Bear", wthBearType, 2, 2)
	d2dSetHand(opp, 3)
	d2dSetHand(other, 3)
	dealNoncombatDamage(g, me.ID, wraith, bear)
	passPriorityAroundTable(t, g)
	if n := discardOwed(g, opp.ID); n != 0 {
		t.Fatalf("damage to a creature asked its controller to discard %d", n)
	}
	dealCombatDamage(g, me.ID, wraith, opp.ID)
	passPriorityAroundTable(t, g)
	if n := discardOwed(g, opp.ID); n != 1 {
		t.Fatalf("the damaged player owes %d discards, want 1", n)
	}
	if n := discardOwed(g, other.ID); n != 0 {
		t.Errorf("a player not dealt damage owes %d discards", n)
	}
}

func TestWth5629FledglingDjinnDamagesYouEachUpkeep(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b12Push(g, me.ID, "Fledgling Djinn", "Creature — Djinn", wthFledglingDjinnOracle, 2, 2)
	walkToUpkeepOfSeat0(t, g)
	life := me.Life
	passPriorityAroundTable(t, g)
	if me.Life != life-1 {
		t.Errorf("life %d, want %d", me.Life, life-1)
	}
}

func TestWth5629TolarianSerpentMillsSevenEachUpkeep(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b12Push(g, me.ID, "Tolarian Serpent", "Creature — Serpent", wthTolarianSerpentOracle, 7, 7)
	walkToUpkeepOfSeat0(t, g)
	library, graveyard := me.Library.Size(), me.Graveyard.Size()
	passPriorityAroundTable(t, g)
	if got := library - me.Library.Size(); got != 7 {
		t.Errorf("milled %d, want 7", got)
	}
	if got := me.Graveyard.Size() - graveyard; got != 7 {
		t.Errorf("%d cards reached the graveyard, want 7", got)
	}
}

func TestWth5629BogardanFirefiendDealsTwoWhenItDies(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	fiend := b12Push(g, me.ID, "Bogardan Firefiend", "Creature — Elemental Spirit", wthBogardanFirefiendOracle, 2, 1)
	ogre := b12Creature(g, opp.ID, "Ogre", "Creature — Ogre", 3, 3)
	b27Kill(g, fiend)
	g.RunStateChecksForTest()
	answerPickTarget(t, g, ogre)
	passPriorityAroundTable(t, g)
	if got := findCardAnywhere(t, g, ogre).DamageMarked; got != 2 {
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
		{"Abyssal Gatekeeper", "Creature — Horror", wthAbyssalGatekeeperOracle, wthBearType},
		{"Hurloon Shaman", "Creature — Minotaur Shaman", wthHurloonShamanOracle, "Basic Land — Forest"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			id := b12Push(g, g.Seats[0].ID, tc.name, tc.typeLine, tc.oracle, 2, 2)
			var victims []uuid.UUID
			for _, s := range g.Seats {
				victims = append(victims, b12Permanent(g, s.ID, "Victim", tc.victimType))
			}
			b27Kill(g, id)
			g.RunStateChecksForTest()
			passPriorityAroundTable(t, g)
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
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := top100ActiveSeat(g)
	ghost := b12Push(g, me.ID, "Mischievous Poltergeist", "Creature — Spirit", wthMischievousPoltergeistOracle, 1, 1)
	life := me.Life
	apaActivate(t, g, me, ghost, 0, game.ActivateAbilityParams{})
	if me.Life != life-1 {
		t.Errorf("life %d after paying, want %d", me.Life, life-1)
	}
	b27Kill(g, ghost)
	if !g.Battlefield.Contains(ghost) {
		t.Error("the shielded Poltergeist was destroyed")
	}
}

func TestWth5629HeavyBallistaShootsOnlyACombatant(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	ballista := b12Push(g, me.ID, "Heavy Ballista", "Creature — Human Soldier", wthHeavyBallistaOracle, 2, 3)
	raider := b12Creature(g, me.ID, "Raider", wthRaiderType, 2, 3)
	idle := b12Creature(g, me.ID, "Idle", wthRaiderType, 2, 3)
	brAttack(t, g, raider)
	if err := g.ActivateCatalogAbility(me.ID, ballista, 0, game.ActivateAbilityParams{Strict: true, Targets: cardTarget(idle)}); err == nil {
		t.Fatal("the Ballista targeted a creature that isn't in combat")
	}
	apaActivate(t, g, me, ballista, 0, game.ActivateAbilityParams{Targets: cardTarget(raider)})
	if got := findCardAnywhere(t, g, raider).DamageMarked; got != 2 {
		t.Errorf("the attacker has %d damage, want 2", got)
	}
	if !hmlTapped(t, g, ballista) {
		t.Error("the Ballista did not tap")
	}
}

func TestWth5629SouthernPaladinDestroysOnlyRed(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me, opp := top100ActiveSeat(g), top100Opponent(g)
	paladin := b12Push(g, me.ID, "Southern Paladin", "Creature — Human Knight", wthSouthernPaladinOracle, 3, 3)
	goblin := b12Creature(g, opp.ID, "Goblin", "Creature — Goblin", 1, 1)
	hmlSetColors(g, goblin, "R")
	elf := b12Creature(g, opp.ID, "Elf", "Creature — Elf", 1, 1)
	hmlSetColors(g, elf, "G")
	apaMana(me, "W", "W")
	if err := g.ActivateCatalogAbility(me.ID, paladin, 0, game.ActivateAbilityParams{Strict: true, Targets: cardTarget(elf)}); err == nil {
		t.Fatal("the Paladin targeted a green creature")
	}
	apaActivate(t, g, me, paladin, 0, game.ActivateAbilityParams{Targets: cardTarget(goblin)})
	if g.Battlefield.Contains(goblin) {
		t.Error("the red creature survived")
	}
}

func TestWth5629TouchstoneTapsAnArtifactYouDontControl(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me, opp := top100ActiveSeat(g), top100Opponent(g)
	stone := pushCatalogPermanent(g, me.ID, "Touchstone", "Artifact", wthTouchstoneOracle, false)
	mine := b12Permanent(g, me.ID, "My Rock", "Artifact")
	theirs := b12Permanent(g, opp.ID, "Their Rock", "Artifact")
	if err := g.ActivateCatalogAbility(me.ID, stone, 0, game.ActivateAbilityParams{Strict: true, Targets: cardTarget(mine)}); err == nil {
		t.Fatal("Touchstone targeted its controller's own artifact")
	}
	apaActivate(t, g, me, stone, 0, game.ActivateAbilityParams{Targets: cardTarget(theirs)})
	if !hmlTapped(t, g, theirs) {
		t.Error("the opponent's artifact is untapped")
	}
}

func TestWth5629InfernalTributeSacrificesANontokenToDraw(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := top100ActiveSeat(g)
	tribute := pushCatalogPermanent(g, me.ID, "Infernal Tribute", "Enchantment", wthInfernalTributeOracle, false)
	token := b12Creature(g, me.ID, "Spirit", "Token Creature — Spirit", 1, 1)
	bear := b12Creature(g, me.ID, "Bear", wthBearType, 2, 2)
	apaMana(me, "C", "C")
	if err := g.ActivateCatalogAbility(me.ID, tribute, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{token}}); err == nil {
		t.Fatal("a token paid the sacrifice")
	}
	hand := me.Hand.Size()
	apaActivate(t, g, me, tribute, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{bear}})
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
		{"Argivian Find", "Instant", wthArgivianFindOracle, "Enchantment", "Creature — Bear"},
		{"Relearn", "Sorcery", wthRelearnOracle, "Instant", "Artifact"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			advanceToMain(t, g)
			me := top100ActiveSeat(g)
			good := wthGraveyardCard(me, "Good", tc.good)
			bad := wthGraveyardCard(me, "Bad", tc.bad)
			if err := castCatalogSpellErr(t, g, tc.name, tc.typeLine, tc.oracle, cardTarget(bad)); err == nil {
				t.Fatalf("%s targeted a %s card", tc.name, tc.bad)
			}
			castCatalogSpell(t, g, tc.name, tc.typeLine, tc.oracle, cardTarget(good))
			passPriorityAroundTable(t, g)
			if !me.Hand.Contains(good) || me.Graveyard.Contains(good) {
				t.Error("the card did not return to hand")
			}
		})
	}
}

func TestWth5629ArgivianRestorationReturnsAnArtifact(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := top100ActiveSeat(g)
	bear := wthGraveyardCard(me, "Bear", wthBearType)
	if err := castCatalogSpellErr(t, g, "Argivian Restoration", "Sorcery", wthArgivianRestorationOracle, cardTarget(bear)); err == nil {
		t.Fatal("Argivian Restoration targeted a creature card")
	}
	rock := wthGraveyardCard(me, "Old Rock", "Artifact")
	castCatalogSpell(t, g, "Argivian Restoration", "Sorcery", wthArgivianRestorationOracle, cardTarget(rock))
	passPriorityAroundTable(t, g)
	if _, ok := visBattlefieldNamed(g, me.ID, "Old Rock"); !ok {
		t.Error("the artifact card did not return to the battlefield under your control")
	}
}

func TestWth5629PumpsGrantFirstStrike(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, oracle string
		p, tough, draws        int
	}{
		{"Fit of Rage", "Sorcery", wthFitOfRageOracle, 5, 5, 0},
		{"Guided Strike", "Instant", wthGuidedStrikeOracle, 3, 2, 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			advanceToMain(t, g)
			me := top100ActiveSeat(g)
			bear := b12Creature(g, me.ID, "Bear", wthBearType, 2, 2)
			hand := me.Hand.Size()
			castCatalogSpell(t, g, tc.name, tc.typeLine, tc.oracle, cardTarget(bear))
			passPriorityAroundTable(t, g)
			if p, tough := ptOf(t, g, bear); p != tc.p || tough != tc.tough {
				t.Errorf("the Bear is %d/%d, want %d/%d", p, tough, tc.p, tc.tough)
			}
			if !hasEffectiveKeyword(t, g, bear, wthFirstStrike) {
				t.Error("the Bear has no first strike")
			}
			if got := me.Hand.Size() - hand; got != tc.draws {
				t.Errorf("drew %d, want %d", got, tc.draws)
			}
			advanceToMainOf(t, g, (g.Turn.ActiveSeat+1)%len(g.Seats))
			if hasEffectiveKeyword(t, g, bear, wthFirstStrike) {
				t.Error("first strike outlasted the turn")
			}
		})
	}
}

func TestWth5629ThunderboltEachMode(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	opp := top100Opponent(g)
	life := opp.Life
	castModal(t, g, "Thunderbolt", "Instant", wthThunderboltOracle, []int{0}, []game.TargetRef{playerTarget(opp.ID)})
	passPriorityAroundTable(t, g)
	if opp.Life != life-3 {
		t.Errorf("the player is at %d, want %d", opp.Life, life-3)
	}
	walker := b12Creature(g, opp.ID, "Walker", wthRaiderType, 5, 5)
	flyer := apaPush(g, opp.ID, opp.ID, game.Card{Name: "Flyer", TypeLine: "Creature — Bird", Power: 5, Toughness: 5, Keywords: []string{"flying"}})
	if err := wthCastWithModes(t, g, "Thunderbolt", "Instant", wthThunderboltOracle, []int{1}, cardTarget(walker)); err == nil {
		t.Fatal("the flying bullet targeted a creature without flying")
	}
	castModal(t, g, "Thunderbolt", "Instant", wthThunderboltOracle, []int{1}, cardTarget(flyer))
	passPriorityAroundTable(t, g)
	if got := findCardAnywhere(t, g, flyer).DamageMarked; got != 4 {
		t.Errorf("the flyer has %d damage, want 4", got)
	}
}

func TestWth5629FatalBlowDestroysADamagedCreaturePastRegeneration(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	opp := top100Opponent(g)
	fresh := b12Creature(g, opp.ID, "Fresh", wthBearType, 2, 2)
	boa := visShield(t, g, opp, 2, 3)
	b27Damage(g, fresh, boa, 1)
	g.RunStateChecksForTest()
	if err := castCatalogSpellErr(t, g, "Fatal Blow", "Instant", wthFatalBlowOracle, cardTarget(fresh)); err == nil {
		t.Fatal("Fatal Blow targeted a creature dealt no damage this turn")
	}
	castCatalogSpell(t, g, "Fatal Blow", "Instant", wthFatalBlowOracle, cardTarget(boa))
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(boa) {
		t.Error("a regeneration shield saved Fatal Blow's target")
	}
}

func TestWth5629TendrilsOfDespairSacrificesAndMakesTheOpponentDiscardTwo(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me, opp := top100ActiveSeat(g), top100Opponent(g)
	victim := b12Creature(g, me.ID, "Victim", wthBearType, 2, 2)
	d2dSetHand(opp, 4)
	id := b13HandCard(me, "Tendrils of Despair", "Sorcery", wthTendrilsOfDespairOracle)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Targets: []game.TargetRef{playerTarget(opp.ID)}}); err == nil {
		t.Fatal("cast with no sacrifice")
	}
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Targets: []game.TargetRef{playerTarget(opp.ID)}, SacrificeIDs: []uuid.UUID{victim}}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(victim) {
		t.Error("the creature was not sacrificed")
	}
	if n := discardOwed(g, opp.ID); n != 2 {
		t.Errorf("the opponent owes %d discards, want 2", n)
	}
}

func TestWth5629AbjureSacrificesABluePermanentToCounter(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	red := b12Creature(g, me.ID, "Red", "Creature — Goblin", 1, 1)
	hmlSetColors(g, red, "R")
	blue := b12Creature(g, me.ID, "Blue", "Creature — Merfolk", 1, 1)
	hmlSetColors(g, blue, "U")
	bolt := b13OpponentCasts(t, g, opp, "Their Bolt", "Instant", "", "{R}", nil)
	id := b13HandCard(me, "Abjure", "Instant", wthAbjureOracle)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Targets: b16TargetCard(bolt), SacrificeIDs: []uuid.UUID{red}}); err == nil {
		t.Fatal("a red permanent paid the sacrifice")
	}
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Targets: b16TargetCard(bolt), SacrificeIDs: []uuid.UUID{blue}}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(blue) {
		t.Error("the blue permanent was not sacrificed")
	}
	if !opp.Graveyard.Contains(bolt) {
		t.Error("the spell was not countered")
	}
}

func TestWth5629GerrardsWisdomGainsTwoPerCardInHand(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := top100ActiveSeat(g)
	d2dSetHand(me, 3)
	life := me.Life
	castCatalogSpell(t, g, "Gerrard's Wisdom", "Sorcery", wthGerrardsWisdomOracle, nil)
	passPriorityAroundTable(t, g)
	if me.Life != life+6 {
		t.Errorf("life %d, want %d", me.Life, life+6)
	}
}

func TestWth5629BlossomingWreathCountsCreatureCards(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := top100ActiveSeat(g)
	pushGraveyardCardForTest(me, "One")
	pushGraveyardCardForTest(me, "Two")
	wthGraveyardCard(me, "Rock", "Artifact")
	life := me.Life
	castCatalogSpell(t, g, "Blossoming Wreath", "Instant", wthBlossomingWreathOracle, nil)
	passPriorityAroundTable(t, g)
	if me.Life != life+2 {
		t.Errorf("life %d, want %d", me.Life, life+2)
	}
}

func TestWth5629NaturesResurgenceDrawsPerCreatureCard(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me, opp := top100ActiveSeat(g), top100Opponent(g)
	pushGraveyardCardForTest(me, "One")
	pushGraveyardCardForTest(me, "Two")
	wthGraveyardCard(me, "Rock", "Artifact")
	pushGraveyardCardForTest(opp, "Three")
	hands := map[uuid.UUID]int{}
	for _, s := range g.Seats {
		hands[s.ID] = s.Hand.Size()
	}
	castCatalogSpell(t, g, "Nature's Resurgence", "Sorcery", wthNaturesResurgenceOracle, nil)
	passPriorityAroundTable(t, g)
	want := map[uuid.UUID]int{me.ID: 2, opp.ID: 1}
	for _, s := range g.Seats {
		if got := s.Hand.Size() - hands[s.ID]; got != want[s.ID] {
			t.Errorf("%s drew %d, want %d", s.Name, got, want[s.ID])
		}
	}
}
