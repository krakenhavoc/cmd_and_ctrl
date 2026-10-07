package effects

import (
	"reflect"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// exert_attack_cards_test.go — ADR 0130 §11, PR 5: the other
// exert-as-it-attacks cards. Each is exerted through the declaration
// verb and the stack, and checked for its effect and for the untap
// skip.

const (
	ahnCropChampionOracle   = "c03d073e-ea59-4873-952a-bcf2730f1115"
	ahnCropCrasherOracle    = "5b3ef91a-5804-43cc-a947-f25f33b7b88c"
	battlefieldScavOracle   = "f88b0634-144a-4a33-a369-62df8edf05fe"
	bitterbladeOracle       = "55b308b2-6e5b-438e-afee-d9b7d4960322"
	championOfRhonasOracle  = "517502e4-4588-4327-b2c0-451f4531fecc"
	clockworkDroidOracle    = "0aa875dc-89ae-4e7a-b10d-aadde8e2b95c"
	devotedCropMateOracle   = "fcde4643-a212-4a9e-94c3-0a0a95841bc6"
	emberhornMinotaurOracle = "c2d258be-8634-4534-bb09-c064329f0940"
	gloryBoundOracle        = "b5582be1-6c2f-4052-80da-bf8f1913567b"
	gustWalkerOracle        = "48bf69ec-0030-44ee-9fbd-6c204b64da7d"
	hoodedBrawlerOracle     = "4e40a68a-5e0c-4f4f-9757-02f8139797e2"
	hydraTrainerOracle      = "c428cbe2-17fd-4bd2-9810-de2561519f14"
	khenraScrapperOracle    = "6ab8958b-10a2-4601-a609-7a372a47f36c"
	nefCropEntanglerOracle  = "04dc245b-011f-4207-ab5c-1528037200f5"
	rhetCropOracle          = "4853bb11-2994-4caf-9981-8a480d472004"
	rohirrimChargersOracle  = "8e638c52-e7b9-445b-adcf-77b58e06c2ba"
	sandstormCrasherOracle  = "d1597137-5609-4012-93bb-19c49c9cf2c3"
	tahCropEliteOracle      = "9af1750e-90cc-446a-9ba2-d44e03c9d01b"
	themberchaudOracle      = "a422a5b0-1aff-4d07-bcc5-edfbda09dc44"
	trueheartTwinsOracle    = "5d3be0d2-0940-4904-9f12-de592c990bdb"
	vizierOfTheTrueOracle   = "0e31d70e-2b5e-4c32-81cf-aed2a13e8aa5"
	watchfulNagaOracle      = "21932be9-cdfb-48d9-87c7-3d62eca98150"
	anepOracle              = "0a73289f-0f0b-4316-be21-8edcbb08bd7c"
)

// exertCard describes one PR 5 card for the shared tests.
type exertCard struct {
	name, oracle, typeLine string
	power, toughness       int
	completeness           Completeness
	// rows is how many triggered rows declare an exert stamp, by kind.
	linked, payoff int
}

var exertAttackCards = []exertCard{
	{"Ahn-Crop Champion", ahnCropChampionOracle, "Creature — Human Warrior", 4, 4, CompletenessFull, 1, 0},
	{"Ahn-Crop Crasher", ahnCropCrasherOracle, "Creature — Minotaur Warrior", 3, 2, CompletenessFull, 1, 0},
	{"Anep, Vizier of Hazoret", anepOracle, "Legendary Creature — Jackal Warrior", 4, 2, CompletenessFull, 1, 0},
	{"Battlefield Scavenger", battlefieldScavOracle, "Creature — Jackal Rogue", 2, 2, CompletenessFull, 0, 1},
	{"Bitterblade Warrior", bitterbladeOracle, "Creature — Jackal Warrior", 2, 2, CompletenessFull, 1, 0},
	{"Champion of Rhonas", championOfRhonasOracle, "Creature — Jackal Warrior", 3, 3, CompletenessFull, 1, 0},
	{"Clockwork Droid", clockworkDroidOracle, "Artifact Creature — Robot", 3, 1, CompletenessFull, 1, 0},
	{"Devoted Crop-Mate", devotedCropMateOracle, "Creature — Human Warrior", 3, 2, CompletenessFull, 1, 0},
	{"Emberhorn Minotaur", emberhornMinotaurOracle, "Creature — Minotaur Warrior", 4, 3, CompletenessFull, 1, 0},
	{"Glory-Bound Initiate", gloryBoundOracle, "Creature — Human Warrior", 3, 1, CompletenessFull, 1, 0},
	{"Gust Walker", gustWalkerOracle, "Creature — Human Wizard", 2, 2, CompletenessFull, 1, 0},
	{"Hooded Brawler", hoodedBrawlerOracle, "Creature — Snake Warrior", 3, 2, CompletenessFull, 1, 0},
	{"Hydra Trainer", hydraTrainerOracle, "Creature — Human Warrior", 1, 1, CompletenessFull, 1, 0},
	{"Khenra Scrapper", khenraScrapperOracle, "Creature — Jackal Warrior", 2, 3, CompletenessFull, 1, 0},
	{"Nef-Crop Entangler", nefCropEntanglerOracle, "Creature — Human Warrior", 2, 1, CompletenessFull, 1, 0},
	{"Rhet-Crop Spearmaster", rhetCropOracle, "Creature — Human Warrior", 3, 1, CompletenessFull, 1, 0},
	{"Rohirrim Chargers", rohirrimChargersOracle, "Creature — Human Knight", 4, 4, CompletenessFull, 0, 1},
	{"Sandstorm Crasher", sandstormCrasherOracle, "Creature — Minotaur Berserker Wizard", 3, 4, CompletenessCaveats, 1, 0},
	{"Tah-Crop Elite", tahCropEliteOracle, "Creature — Bird Warrior", 2, 2, CompletenessFull, 1, 0},
	{"Themberchaud", themberchaudOracle, "Legendary Creature — Dragon", 5, 5, CompletenessFull, 1, 0},
	{"Trueheart Twins", trueheartTwinsOracle, "Creature — Jackal Warrior", 4, 4, CompletenessFull, 0, 1},
	{"Vizier of the True", vizierOfTheTrueOracle, "Creature — Human Cleric", 3, 2, CompletenessFull, 0, 1},
	{"Watchful Naga", watchfulNagaOracle, "Creature — Snake Wizard", 2, 2, CompletenessFull, 1, 0},
}

// Every card is registered with the exert-as-it-attacks ability, the
// completeness it was written to, and its exert rows stamped by the
// helper that built them.
func TestExertAttackCardsDeclareTheirAbility(t *testing.T) {
	for _, c := range exertAttackCards {
		t.Run(c.name, func(t *testing.T) {
			spec, ok := Lookup(c.oracle)
			if !ok {
				t.Fatalf("%s is not registered", c.name)
			}
			if spec.ExertOnAttack == nil {
				t.Error("no ExertOnAttack")
			}
			if spec.Completeness != c.completeness {
				t.Errorf("completeness %v, want %v", spec.Completeness, c.completeness)
			}
			if c.completeness == CompletenessCaveats && len(spec.Caveats) == 0 {
				t.Error("a Caveats card must say what it leaves out")
			}
			linked, payoff := 0, 0
			for _, tr := range spec.Triggered {
				switch tr.Exert {
				case game.ExertRowLinked:
					linked++
				case game.ExertRowPayoff:
					payoff++
				}
			}
			if linked != c.linked || payoff != c.payoff {
				t.Errorf("exert rows linked=%d payoff=%d, want %d and %d", linked, payoff, c.linked, c.payoff)
			}
		})
	}
}

// The purposes the bot prices an exert by are the printed amounts.
func TestExertAttackCardsDeclareTheirPurposes(t *testing.T) {
	pump := func(p, tough int, kws ...string) game.Purpose {
		return game.Purpose{Pump: &game.Pump{Power: p, Toughness: tough, Keywords: kws}}
	}
	want := map[string]game.Purpose{
		bitterbladeOracle:       pump(1, 0, "deathtouch"),
		emberhornMinotaurOracle: pump(1, 1, "menace"),
		gloryBoundOracle:        pump(1, 3, "lifelink"),
		gustWalkerOracle:        pump(1, 1, "flying"),
		hoodedBrawlerOracle:     pump(2, 2),
		khenraScrapperOracle:    pump(2, 0),
		nefCropEntanglerOracle:  pump(1, 2),
		rhetCropOracle:          pump(1, 0, "first strike"),
		themberchaudOracle:      pump(0, 0, "flying"),
		watchfulNagaOracle:      {Draws: 1},
	}
	for oracle, p := range want {
		spec, _ := Lookup(oracle)
		found := false
		for _, tr := range spec.Triggered {
			if tr.Exert != game.ExertRowLinked {
				continue
			}
			found = true
			if !reflect.DeepEqual(tr.Purpose, p) {
				t.Errorf("%s: purpose %+v, want %+v", spec.Name, tr.Purpose, p)
			}
		}
		if !found {
			t.Errorf("%s has no linked exert row", spec.Name)
		}
	}
}

// pushExertAttacker seeds an exert card and returns the active seat,
// the defender and the creature.
func pushExertAttacker(t *testing.T, g *game.Game, c exertCard) (me, opp *game.Player, id uuid.UUID) {
	t.Helper()
	me = g.Seats[g.Turn.ActiveSeat]
	opp = g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	id = pushDiesCreatureForTest(g, me.ID, c.name, c.oracle, c.typeLine, c.power, c.toughness)
	return me, opp, id
}

func exertCardNamed(name string) exertCard {
	for _, c := range exertAttackCards {
		if c.name == name {
			return c
		}
	}
	panic("no exert card " + name)
}

// effectiveStats reads a battlefield creature's post-layer power,
// toughness and one keyword.
func effectiveStats(g *game.Game, id uuid.UUID, keyword string) (power, toughness int, has bool) {
	g.ReadSnapshot(func() {
		for i := range g.Battlefield.Cards {
			c := &g.Battlefield.Cards[i]
			if c.InstanceID != id {
				continue
			}
			eff := c.Effective()
			power, toughness = eff.Power, eff.Toughness
			has = keyword != "" && game.HasKeyword(c, keyword)
		}
	})
	return power, toughness, has
}

// The eight self-pump cards and Themberchaud's flying: exerted, the
// creature gets exactly the printed pump before blockers, and skips
// its controller's next untap. Not exerted, it gets nothing.
func TestExertSelfPumpCards(t *testing.T) {
	cases := []struct {
		name        string
		dp, dt      int
		keyword     string
		printedPlus string // a keyword the card already has, to prove nothing replaces it
	}{
		{"Bitterblade Warrior", 1, 0, "deathtouch", ""},
		{"Emberhorn Minotaur", 1, 1, "menace", ""},
		{"Glory-Bound Initiate", 1, 3, "lifelink", ""},
		{"Gust Walker", 1, 1, "flying", ""},
		{"Hooded Brawler", 2, 2, "", ""},
		{"Khenra Scrapper", 2, 0, "", ""},
		{"Nef-Crop Entangler", 1, 2, "", ""},
		{"Rhet-Crop Spearmaster", 1, 0, "first strike", ""},
		{"Themberchaud", 0, 0, "flying", ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			card := exertCardNamed(tc.name)
			g := newCatalogGame(t)
			me, opp, id := pushExertAttacker(t, g, card)
			declareExerting(t, g, id, opp.ID)
			lockInAttacks(t, g)
			if triggerOnStack(g, id) == nil {
				t.Fatal("the linked trigger goes on the stack with the declaration")
			}
			passPriorityAroundTable(t, g)
			p, tough, has := effectiveStats(g, id, tc.keyword)
			if p != card.power+tc.dp || tough != card.toughness+tc.dt {
				t.Errorf("%d/%d, want %d/%d", p, tough, card.power+tc.dp, card.toughness+tc.dt)
			}
			if tc.keyword != "" && !has {
				t.Errorf("did not gain %s", tc.keyword)
			}
			if !hasExerterSkip(g, id, me.ID) {
				t.Error("it won't untap during its controller's next untap step")
			}
		})
		t.Run(tc.name+" attacking without exerting", func(t *testing.T) {
			card := exertCardNamed(tc.name)
			g := newCatalogGame(t)
			me, opp, id := pushExertAttacker(t, g, card)
			advanceTo(t, g, game.StepDeclareAttackers)
			if err := g.DeclareAttacker(id, opp.ID); err != nil {
				t.Fatal(err)
			}
			lockInAttacks(t, g)
			passPriorityAroundTable(t, g)
			p, tough, has := effectiveStats(g, id, tc.keyword)
			if p != card.power || tough != card.toughness || (tc.keyword != "" && has) {
				t.Errorf("an unexerted attacker is %d/%d (keyword %v), want the printed %d/%d", p, tough, has, card.power, card.toughness)
			}
			if hasExerterSkip(g, id, me.ID) || exertedEvents(g, id) != 0 {
				t.Error("nothing exerted it")
			}
		})
	}
}

// Ahn-Crop Champion untaps the other creatures you control, the
// attackers included, and stays tapped itself.
func TestAhnCropChampionUntapsOtherCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, champion := pushExertAttacker(t, g, exertCardNamed("Ahn-Crop Champion"))
	bear := pushVanillaCreature(g, me.ID, "Bear", 2, 2)
	theirs := pushVanillaCreature(g, opp.ID, "Their Bear", 2, 2)
	findBattlefieldCardForTest(g, theirs).Tapped = true

	advanceTo(t, g, game.StepDeclareAttackers)
	if _, err := g.DeclareAttackersWith([]game.AttackDeclaration{
		{Attacker: champion, Target: opp.ID, Exert: true},
		{Attacker: bear, Target: opp.ID},
	}, game.DeclareAttackersParams{}); err != nil {
		t.Fatal(err)
	}
	passPriorityAroundTable(t, g)
	if isTapped(g, bear) {
		t.Error("the attacking Bear untaps and stays attacking")
	}
	if findBattlefieldCardForTest(g, bear).AttackingTarget != opp.ID {
		t.Error("untapping an attacker doesn't remove it from combat")
	}
	if !isTapped(g, champion) || !isTapped(g, theirs) {
		t.Error("the Champion itself and the opponent's creatures stay tapped")
	}
	if !hasExerterSkip(g, champion, me.ID) {
		t.Error("the Champion won't untap during your next untap step")
	}
}

// Ahn-Crop Crasher: the chosen creature can't block; other creatures
// still can.
func TestAhnCropCrasherStopsATargetFromBlocking(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, crasher := pushExertAttacker(t, g, exertCardNamed("Ahn-Crop Crasher"))
	wall := pushVanillaCreature(g, opp.ID, "Wall", 0, 4)
	other := pushVanillaCreature(g, opp.ID, "Other Wall", 0, 4)

	declareExerting(t, g, crasher, opp.ID)
	lockInAttacks(t, g)
	pickTriggerTarget(t, g, me.ID, wall)
	passPriorityAroundTable(t, g)
	advanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(wall, crasher); err == nil {
		t.Error("the targeted Wall can't block this turn")
	}
	if err := g.DeclareBlocker(other, crasher); err != nil {
		t.Errorf("another creature still blocks: %v", err)
	}
	if !hasExerterSkip(g, crasher, me.ID) {
		t.Error("the Crasher skips its next untap")
	}
}

// Clockwork Droid can't be blocked this turn and scries 1.
func TestClockworkDroidIsUnblockableAndScries(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, droid := pushExertAttacker(t, g, exertCardNamed("Clockwork Droid"))
	blocker := pushVanillaCreature(g, opp.ID, "Wall", 0, 4)

	declareExerting(t, g, droid, opp.ID)
	lockInAttacks(t, g)
	passPriorityAroundTable(t, g)
	if scryChoiceFor(g, me.ID) == nil {
		t.Fatal("scry 1")
	}
	answerScryKeepAll(t, g, me.ID)
	passPriorityAroundTable(t, g)
	advanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(blocker, droid); err == nil {
		t.Error("the exerted Droid can't be blocked")
	}
	if !hasExerterSkip(g, droid, me.ID) {
		t.Error("the Droid skips its next untap")
	}
}

// Watchful Naga draws a card when exerted.
func TestWatchfulNagaDrawsWhenExerted(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, naga := pushExertAttacker(t, g, exertCardNamed("Watchful Naga"))
	hand := me.Hand.Size()
	declareExerting(t, g, naga, opp.ID)
	lockInAttacks(t, g)
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand %d, want %d", got, hand+1)
	}
	if !hasExerterSkip(g, naga, me.ID) {
		t.Error("the Naga skips its next untap")
	}
}

// Tah-Crop Elite gives every creature you control +1/+1, and only
// those.
func TestTahCropEliteGivesTheTeamAPump(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, elite := pushExertAttacker(t, g, exertCardNamed("Tah-Crop Elite"))
	mine := pushVanillaCreature(g, me.ID, "Bear", 2, 2)
	theirs := pushVanillaCreature(g, opp.ID, "Their Bear", 2, 2)
	declareExerting(t, g, elite, opp.ID)
	lockInAttacks(t, g)
	passPriorityAroundTable(t, g)
	if p, tough, _ := effectiveStats(g, mine, ""); p != 3 || tough != 3 {
		t.Errorf("my Bear %d/%d, want 3/3", p, tough)
	}
	if p, tough, _ := effectiveStats(g, elite, ""); p != 3 || tough != 3 {
		t.Errorf("the Elite %d/%d, want 3/3", p, tough)
	}
	if p, tough, _ := effectiveStats(g, theirs, ""); p != 2 || tough != 2 {
		t.Errorf("their Bear %d/%d, want 2/2", p, tough)
	}
	if !hasExerterSkip(g, elite, me.ID) {
		t.Error("the Elite skips its next untap")
	}
}

// Trueheart Twins pumps once per exert, for itself and for any other
// creature its controller exerts.
func TestTrueheartTwinsTriggersOnEveryExert(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, twins := pushExertAttacker(t, g, exertCardNamed("Trueheart Twins"))
	bear := pushVanillaCreature(g, me.ID, "Bear", 2, 2)
	avenger := pushDiesCreatureForTest(g, me.ID, "Oketra's Avenger", oketrasAvengerOracle, "Creature — Human Warrior", 3, 1)
	theirs := pushVanillaCreature(g, opp.ID, "Their Bear", 2, 2)

	advanceTo(t, g, game.StepDeclareAttackers)
	if _, err := g.DeclareAttackersWith([]game.AttackDeclaration{
		{Attacker: twins, Target: opp.ID},
		{Attacker: avenger, Target: opp.ID, Exert: true},
	}, game.DeclareAttackersParams{}); err != nil {
		t.Fatal(err)
	}
	answerAnyTriggerOrderPrompt(t, g, me.ID)
	passPriorityAroundTable(t, g)
	if p, _, _ := effectiveStats(g, bear, ""); p != 3 {
		t.Errorf("my Bear has power %d, want 3: one exert, one +1/+0", p)
	}
	if p, _, _ := effectiveStats(g, twins, ""); p != 5 {
		t.Errorf("the Twins have power %d, want 5", p)
	}
	if p, _, _ := effectiveStats(g, theirs, ""); p != 2 {
		t.Errorf("their Bear has power %d, want 2", p)
	}
	if exertedEvents(g, twins) != 0 {
		t.Error("the Twins attacked without exerting")
	}
}

// Vizier of the True taps a creature an opponent controls when you
// exert any creature, its own included.
func TestVizierOfTheTrueTapsAnOpponentsCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, vizier := pushExertAttacker(t, g, exertCardNamed("Vizier of the True"))
	theirs := pushVanillaCreature(g, opp.ID, "Their Bear", 2, 2)
	mine := pushVanillaCreature(g, me.ID, "My Bear", 2, 2)

	declareExerting(t, g, vizier, opp.ID)
	lockInAttacks(t, g)
	if err := g.ResolvePickTarget(latestPickTarget(g, me.ID).ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: mine}); err == nil {
		t.Fatal("your own creature is not a legal target")
	}
	pickTriggerTarget(t, g, me.ID, theirs)
	passPriorityAroundTable(t, g)
	if !isTapped(g, theirs) {
		t.Error("the opponent's creature is tapped")
	}
	if !hasExerterSkip(g, vizier, me.ID) {
		t.Error("the Vizier skips its next untap")
	}
}

// Battlefield Scavenger: the rummage is "may discard, if you do draw".
func TestBattlefieldScavengerRummages(t *testing.T) {
	for _, discard := range []bool{true, false} {
		name := "declined"
		if discard {
			name = "discarded"
		}
		t.Run(name, func(t *testing.T) {
			g := newCatalogGame(t)
			me, opp, scav := pushExertAttacker(t, g, exertCardNamed("Battlefield Scavenger"))
			hand, graves := me.Hand.Size(), me.Graveyard.Size()
			pitch := me.Hand.Cards[0].InstanceID
			declareExerting(t, g, scav, opp.ID)
			lockInAttacks(t, g)
			passPriorityAroundTable(t, g)
			if discard {
				answerDiscard(t, g, me.ID, pitch)
			} else {
				answerDiscard(t, g, me.ID)
			}
			passPriorityAroundTable(t, g)
			if got := me.Hand.Size(); got != hand {
				t.Errorf("hand %d, want %d (discard one and draw one, or neither)", got, hand)
			}
			wantGraves := graves
			if discard {
				wantGraves++
			}
			if got := me.Graveyard.Size(); got != wantGraves {
				t.Errorf("graveyard %d, want %d", got, wantGraves)
			}
			if !hasExerterSkip(g, scav, me.ID) {
				t.Error("the Scavenger skips its next untap")
			}
		})
	}
}

// Champion of Rhonas puts a creature card from hand onto the
// battlefield, not attacking, and the "may" is a real decline.
func TestChampionOfRhonasPutsACreatureFromHand(t *testing.T) {
	for _, put := range []bool{true, false} {
		name := "declined"
		if put {
			name = "put"
		}
		t.Run(name, func(t *testing.T) {
			g := newCatalogGame(t)
			me, opp, champion := pushExertAttacker(t, g, exertCardNamed("Champion of Rhonas"))
			giant := handCardForTest(me, "Giant", "Creature — Giant", "")
			spell := handCardForTest(me, "Bolt", "Instant", "")
			declareExerting(t, g, champion, opp.ID)
			lockInAttacks(t, g)
			passPriorityAroundTable(t, g)
			if put {
				answerChooseCards(t, g, me.ID, giant)
			} else {
				answerChooseCards(t, g, me.ID)
			}
			passPriorityAroundTable(t, g)
			c := findBattlefieldCardForTest(g, giant)
			if put {
				if c == nil || c.Tapped || c.AttackingTarget != uuid.Nil {
					t.Errorf("the Giant should be on the battlefield untapped and not attacking: %+v", c)
				}
			} else if c != nil {
				t.Error("declined: the Giant stays in hand")
			}
			if findBattlefieldCardForTest(g, spell) != nil {
				t.Error("only a creature card can be put")
			}
			if !hasExerterSkip(g, champion, me.ID) {
				t.Error("Champion of Rhonas skips its next untap")
			}
		})
	}
}

// Devoted Crop-Mate returns a creature card with mana value 2 or less
// from your graveyard.
func TestDevotedCropMateReturnsASmallCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, mate := pushExertAttacker(t, g, exertCardNamed("Devoted Crop-Mate"))
	small := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: small, Name: "Small", TypeLine: "Creature — Elf", ManaCost: "{1}{G}", Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID})
	big := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: big, Name: "Big", TypeLine: "Creature — Giant", ManaCost: "{2}{G}", Power: 3, Toughness: 3, Owner: me.ID, Controller: me.ID})

	declareExerting(t, g, mate, opp.ID)
	lockInAttacks(t, g)
	pick := latestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("the trigger asks for its target")
	}
	if err := g.ResolvePickTarget(pick.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: big}); err == nil {
		t.Fatal("mana value 3 is not a legal target")
	}
	pickTriggerTarget(t, g, me.ID, small)
	passPriorityAroundTable(t, g)
	if c := findBattlefieldCardForTest(g, small); c == nil || c.Controller != me.ID {
		t.Fatal("the creature card returns to the battlefield under your control")
	}
	if !hasExerterSkip(g, mate, me.ID) {
		t.Error("the Crop-Mate skips its next untap")
	}
}

// Hydra Trainer pumps the target by every counter on permanents you
// control, and only those.
func TestHydraTrainerCountsCountersOnYourPermanents(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, trainer := pushExertAttacker(t, g, exertCardNamed("Hydra Trainer"))
	a := pushVanillaCreature(g, me.ID, "Counter Bear", 2, 2)
	findBattlefieldCardForTest(g, a).Counters = map[string]int{game.CounterPlusOne: 2}
	b := pushVanillaCreature(g, me.ID, "Other Bear", 2, 2)
	findBattlefieldCardForTest(g, b).Counters = map[string]int{"charge": 1}
	theirs := pushVanillaCreature(g, opp.ID, "Their Bear", 2, 2)
	findBattlefieldCardForTest(g, theirs).Counters = map[string]int{game.CounterPlusOne: 5}

	declareExerting(t, g, trainer, opp.ID)
	lockInAttacks(t, g)
	pickTriggerTarget(t, g, me.ID, trainer)
	passPriorityAroundTable(t, g)
	// X = 2 + 1 = 3; the opponent's five counters don't count.
	if p, tough, _ := effectiveStats(g, trainer, ""); p != 4 || tough != 4 {
		t.Errorf("the Trainer is %d/%d, want 4/4 (X = 3)", p, tough)
	}
	if !hasExerterSkip(g, trainer, me.ID) {
		t.Error("the Trainer skips its next untap")
	}
}

// Anep exiles the top two cards and lets you play them.
func TestAnepExilesTwoCardsYouMayPlay(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, anep := pushExertAttacker(t, g, exertCardNamed("Anep, Vizier of Hazoret"))
	me.Library.Cards = nil
	second := uuid.New()
	me.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: "Third", TypeLine: "Instant", ManaCost: "{R}", Owner: me.ID, Controller: me.ID})
	me.Library.PushTop(game.Card{InstanceID: second, Name: "Second", TypeLine: "Instant", ManaCost: "{R}", Owner: me.ID, Controller: me.ID})
	first := uuid.New()
	me.Library.PushTop(game.Card{InstanceID: first, Name: "First", TypeLine: "Land", Owner: me.ID, Controller: me.ID})

	declareExerting(t, g, anep, opp.ID)
	lockInAttacks(t, g)
	passPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{first, second} {
		if !g.Exile.Contains(id) {
			t.Fatalf("the top two cards are exiled (%s missing)", id)
		}
		if perm := exiledPermission(g, id); !permissionLive(g, perm, me.ID) {
			t.Errorf("you may play %s from exile", id)
		}
	}
	if me.Library.Size() != 1 {
		t.Errorf("library %d, want 1", me.Library.Size())
	}
	if !hasExerterSkip(g, anep, me.ID) {
		t.Error("Anep skips its next untap")
	}
}

// Rohirrim Chargers reveals to an Equipment and puts it onto the
// battlefield attached to the exerted creature, the rest on the bottom.
// It sees another creature's exert too, and attaches to THAT creature.
func TestRohirrimChargersAttachesTheEquipmentToTheExertedCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, chargers := pushExertAttacker(t, g, exertCardNamed("Rohirrim Chargers"))
	avenger := pushDiesCreatureForTest(g, me.ID, "Oketra's Avenger", oketrasAvengerOracle, "Creature — Human Warrior", 3, 1)
	me.Library.Cards = nil
	me.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: "Deep", TypeLine: "Creature — Elf", Owner: me.ID, Controller: me.ID})
	sword := uuid.New()
	me.Library.PushTop(game.Card{InstanceID: sword, Name: "Short Sword", TypeLine: "Artifact — Equipment", Owner: me.ID, Controller: me.ID})
	me.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: "Filler Two", TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID})
	me.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: "Filler One", TypeLine: "Land", Owner: me.ID, Controller: me.ID})

	advanceTo(t, g, game.StepDeclareAttackers)
	if _, err := g.DeclareAttackersWith([]game.AttackDeclaration{
		{Attacker: chargers, Target: opp.ID},
		{Attacker: avenger, Target: opp.ID, Exert: true},
	}, game.DeclareAttackersParams{}); err != nil {
		t.Fatal(err)
	}
	answerAnyTriggerOrderPrompt(t, g, me.ID)
	passPriorityAroundTable(t, g)
	eq := findBattlefieldCardForTest(g, sword)
	if eq == nil {
		t.Fatal("the Equipment is put onto the battlefield")
	}
	if eq.AttachedTo.ID != avenger {
		t.Errorf("attached to %s, want the exerted Avenger", eq.AttachedTo.ID)
	}
	if got := me.Library.Size(); got != 3 {
		t.Errorf("library %d, want the 3 cards that weren't put onto the battlefield", got)
	}
	if exertedEvents(g, chargers) != 0 {
		t.Error("the Chargers attacked without exerting")
	}
}

// With no Equipment in the library, everything is revealed and goes to
// the bottom: nothing enters.
func TestRohirrimChargersWithNoEquipmentRevealsTheLibrary(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, chargers := pushExertAttacker(t, g, exertCardNamed("Rohirrim Chargers"))
	me.Library.Cards = nil
	for _, n := range []string{"A", "B"} {
		me.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: n, TypeLine: "Land", Owner: me.ID, Controller: me.ID})
	}
	bf := len(g.Battlefield.Cards)
	declareExerting(t, g, chargers, opp.ID)
	lockInAttacks(t, g)
	passPriorityAroundTable(t, g)
	if len(g.Battlefield.Cards) != bf || me.Library.Size() != 2 {
		t.Errorf("battlefield %d→%d, library %d: nothing should have entered", bf, len(g.Battlefield.Cards), me.Library.Size())
	}
	if !hasExerterSkip(g, chargers, me.ID) {
		t.Error("the Chargers skip their next untap")
	}
}

// Sandstorm Crasher makes a tapped, attacking token copy of the target
// and sacrifices it at the next end step.
func TestSandstormCrasherCopiesAndSacrificesAtEndStep(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, crasher := pushExertAttacker(t, g, exertCardNamed("Sandstorm Crasher"))
	hydra := pushDiesCreatureForTest(g, me.ID, "Tah-Crop Elite", tahCropEliteOracle, "Creature — Bird Warrior", 2, 2)

	declareExerting(t, g, crasher, opp.ID)
	lockInAttacks(t, g)
	pickTriggerTarget(t, g, me.ID, hydra)
	passPriorityAroundTable(t, g)

	var token *game.Card
	for i := range g.Battlefield.Cards {
		c := &g.Battlefield.Cards[i]
		if c.Controller == me.ID && c.IsToken() && c.Name == "Tah-Crop Elite" {
			token = c
		}
	}
	if token == nil {
		t.Fatal("a token copy of Tah-Crop Elite")
	}
	if !token.Tapped || token.AttackingTarget != opp.ID {
		t.Errorf("the token should be tapped and attacking %s: tapped=%v attacking=%s", opp.ID, token.Tapped, token.AttackingTarget)
	}
	if !hasExerterSkip(g, crasher, me.ID) {
		t.Error("the Crasher skips its next untap")
	}
	id := token.InstanceID
	advanceTo(t, g, game.StepEnd)
	passPriorityAroundTable(t, g)
	if findBattlefieldCardForTest(g, id) != nil {
		t.Error("the token is sacrificed at the beginning of the next end step")
	}
}

// Themberchaud's enters trigger deals X to each other creature without
// flying and each player, X being the Mountains you control.
func TestThemberchaudEntersDealsDamageForEachMountain(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	for i := 0; i < 3; i++ {
		pushDiesCreatureForTest(g, me.ID, "Mountain", "", "Basic Land — Mountain", 0, 0)
	}
	pushDiesCreatureForTest(g, opp.ID, "Their Mountain", "", "Basic Land — Mountain", 0, 0)
	ground := pushVanillaCreature(g, opp.ID, "Ground Bear", 2, 9)
	flyer := pushVanillaCreature(g, opp.ID, "Flyer", 2, 9)
	findBattlefieldCardForTest(g, flyer).Keywords = append(findBattlefieldCardForTest(g, flyer).Keywords, "flying")
	mine := pushVanillaCreature(g, me.ID, "My Bear", 2, 9)

	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Themberchaud", OracleID: themberchaudOracle,
		TypeLine: "Legendary Creature — Dragon", Power: 5, Toughness: 5, Owner: me.ID, Controller: me.ID})
	lives := map[uuid.UUID]int{}
	for _, p := range g.Seats {
		lives[p.ID] = p.Life
	}
	g.WithWriteLock(func() {
		if _, err := g.PutFromHandOntoBattlefieldForEffect(id, game.HandEntryOptions{Controller: me.ID}); err != nil {
			t.Fatal(err)
		}
	})
	passPriorityAroundTable(t, g)
	if damageMarkedOn(g, ground) != 3 || damageMarkedOn(g, mine) != 3 {
		t.Errorf("ground creatures took %d and %d, want 3 (three Mountains you control)", damageMarkedOn(g, ground), damageMarkedOn(g, mine))
	}
	if damageMarkedOn(g, flyer) != 0 {
		t.Error("a creature with flying is spared")
	}
	if damageMarkedOn(g, id) != 0 {
		t.Error("\"each other creature\": Themberchaud is spared")
	}
	for _, p := range g.Seats {
		if p.Life != lives[p.ID]-3 {
			t.Errorf("%s at %d, want %d: each player, the controller included", p.Name, p.Life, lives[p.ID]-3)
		}
	}
}
