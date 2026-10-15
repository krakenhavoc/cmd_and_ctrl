package effects

import (
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// improvise_cards_test.go — the cards that print improvise (CR 702.126),
// Aether Revolt onward (#3654). The cost itself is tested in
// improvise_test.go; this file pins that each card declares it and that
// the keyword behaves as printed on real cards.

const (
	foundryAssemblerOracle   = "8242f4f5-c51d-42dc-a6ae-67060368fed9"
	reverseEngineerOracle    = "f56596be-e12f-4ab6-af61-3ef8dc87134c"
	ironheartOracle          = "df3d5b86-7b75-4df9-8985-7159cde12261"
	organicExtinctionOracle  = "d9e24d7c-0907-45c1-b5d3-9fb16a929868"
	maverickThopteristOracle = "71ba6f1b-2fa6-4d4c-8778-f0765b1a5d8e"
	battleAtTheBridgeOracle  = "3d63ad3a-4e9f-4968-a576-4be063f4ecbf"
	slyRequisitionerOracle   = "15ea4698-86b4-47c5-a9a7-60af16063a61"
	universalSurvOracle      = "c5a5c17d-d1e4-4d23-9d2b-1c2febdcb0e7"
)

var improviseCardOracles = map[string]string{
	"8242f4f5-c51d-42dc-a6ae-67060368fed9": "Foundry Assembler",
	"dbf4e56e-57dd-4f41-95a9-8b9b096274ec": "Enraged Giant",
	"9bb1de3f-b6af-4ee2-b896-a05aae50e238": "Sweatworks Brawler",
	"74dcff4f-c126-4bbb-b6f0-3f3315bc95fe": "Wind-Kin Raiders",
	"900ae045-0160-4e85-b283-e35399c2acec": "Bastion Inventor",
	"87a72c3d-e787-4dd8-b934-4fff4573af6c": "Freejam Regent",
	"93491d5c-8579-42ed-a700-e5c6f4238dea": "Barricade Breaker",
	"0f909cfd-c643-48d2-983a-da2e38cd2bff": "Fen Hauler",
	"55f7b005-31ea-4865-a8f3-460d15bed795": "Arc Reactor",
	"f56596be-e12f-4ab6-af61-3ef8dc87134c": "Reverse Engineer",
	"c5a5c17d-d1e4-4d23-9d2b-1c2febdcb0e7": "Universal Surveillance",
	"d9e24d7c-0907-45c1-b5d3-9fb16a929868": "Organic Extinction",
	"c4920e95-a6fc-4373-a0d6-244f51851364": "Metallic Rebuke",
	"71ba6f1b-2fa6-4d4c-8778-f0765b1a5d8e": "Maverick Thopterist",
	"97b559fe-b09d-4883-9509-c45ace2d8487": "Herald of Anguish",
	"15ea4698-86b4-47c5-a9a7-60af16063a61": "Sly Requisitioner",
	"df3d5b86-7b75-4df9-8985-7159cde12261": "Ironheart, Clever Champion",
	"3d63ad3a-4e9f-4968-a576-4be063f4ecbf": "Battle at the Bridge",
	"152b91c9-cc07-4ca8-944f-9bc2242a2283": "Whir of Invention",
	"3d3553df-0036-47d9-b0e3-4a3261a0f73c": "Synth Infiltrator",
	kappaCannoneerOracle:                   "Kappa Cannoneer",
}

func TestEveryNativeImproviseCardDeclaresTheImproviseCost(t *testing.T) {
	for oracle, name := range improviseCardOracles {
		spec, ok := Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.TapCost == nil || spec.TapCost.Key != "improvise" {
			t.Errorf("%s: TapCost = %+v, want the improvise cost", name, spec.TapCost)
		}
	}
}

func impCard(owner uuid.UUID, name, typeLine, cost, oracle string) game.Card {
	return game.Card{InstanceID: uuid.New(), Name: name, TypeLine: typeLine, ManaCost: cost, OracleID: oracle,
		Owner: owner, Controller: owner}
}

func impHand(g *game.Game, p *game.Player, c game.Card) uuid.UUID {
	g.WithWriteLock(func() { p.Hand.PushTop(c) })
	return c.InstanceID
}

func TestFoundryAssemblerIsPaidForByFiveArtifactsOnly(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	arts := impArtifacts(g, me.ID, 6)
	id := impHand(g, me, impCard(me.ID, "Foundry Assembler", "Artifact Creature — Assembly-Worker", "{5}", foundryAssemblerOracle))

	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, TapIDs: arts}); err == nil {
		t.Fatal("six artifacts were accepted for {5}")
	}
	// A tapped artifact cannot be tapped again.
	g.WithWriteLock(func() { findBattlefieldCardForTest(g, arts[0]).Tapped = true })
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, TapIDs: arts[:5]}); err == nil {
		t.Fatal("an already tapped artifact was tapped for improvise")
	}
	g.WithWriteLock(func() { findBattlefieldCardForTest(g, arts[0]).Tapped = false })
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, TapIDs: arts[:5]}); err != nil {
		t.Fatalf("five artifacts for {5}: %v", err)
	}
	if !cgTapped(g, arts[:5]) || cgTapped(g, arts[5:]) {
		t.Error("exactly the five named artifacts should be tapped")
	}
}

func TestAnImprovisedArtifactCreatureCannotAttack(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	tapper := pushCatalogPermanent(g, me.ID, "Tapped Golem", "Artifact Creature — Golem", "", false)
	fresh := pushCatalogPermanent(g, me.ID, "Fresh Golem", "Artifact Creature — Golem", "", false)
	rest := impArtifacts(g, me.ID, 4)
	id := impHand(g, me, impCard(me.ID, "Foundry Assembler", "Artifact Creature — Assembly-Worker", "{5}", foundryAssemblerOracle))
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, TapIDs: append([]uuid.UUID{tapper}, rest...)}); err != nil {
		t.Fatalf("improvising with an artifact creature: %v", err)
	}
	passPriorityAroundTable(t, g)

	advanceTo(t, g, game.StepDeclareAttackers)
	opp := g.Seats[1]
	declared, err := g.DeclareAttackers([]game.AttackDeclaration{
		{Attacker: tapper, Target: opp.ID},
		{Attacker: fresh, Target: opp.ID},
	})
	if err != nil {
		t.Fatalf("DeclareAttackers: %v", err)
	}
	if len(declared) != 1 || declared[0] != fresh {
		t.Errorf("declared %v, want only the untapped artifact creature %v (the improvised one is tapped)", declared, fresh)
	}
}

func TestReverseEngineerImprovisesTheGenericAndDrawsThree(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	arts := impArtifacts(g, me.ID, 3)
	id := impHand(g, me, impCard(me.ID, "Reverse Engineer", "Sorcery", "{3}{U}{U}", reverseEngineerOracle))

	// Three artifacts pay {3}; the {U}{U} still needs blue mana.
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, TapIDs: arts}); err == nil {
		t.Fatal("artifacts paid {U}{U}")
	}
	fillPoolColored(me, "U", 2)
	handBefore := me.Hand.Size()
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, TapIDs: arts}); err != nil {
		t.Fatalf("improvising Reverse Engineer: %v", err)
	}
	passPriorityAroundTable(t, g)
	// The spell left the hand and three cards arrived.
	if got := me.Hand.Size() - handBefore; got != 2 {
		t.Errorf("hand changed by %d, want +2 (three drawn, the spell cast)", got)
	}
}

func TestUniversalSurveillanceArtifactsPayTheGenericX(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	arts := impArtifacts(g, me.ID, 3)
	id := impHand(g, me, impCard(me.ID, "Universal Surveillance", "Sorcery", "{X}{U}{U}{U}", universalSurvOracle))
	fillPoolColored(me, "U", 3)
	handBefore := me.Hand.Size()
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, XValue: 3, TapIDs: arts}); err != nil {
		t.Fatalf("X=3 paid by three artifacts: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size() - handBefore; got != 2 {
		t.Errorf("hand changed by %d, want +2 (X=3 drawn, the spell cast)", got)
	}
}

func TestIronheartGivesNoncreatureSpellsImproviseButNotCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	pushCatalogPermanent(g, me.ID, "Ironheart, Clever Champion", "Legendary Artifact Creature — Human Hero", ironheartOracle, false)
	arts := impArtifacts(g, me.ID, 2)

	bear := impBear(me.ID)
	impHand(g, me, bear)
	fillPoolColored(me, "G", 1)
	if err := g.CastSpell(me.ID, bear.InstanceID, game.CastSpellParams{Strict: true, TapIDs: arts[:1]}); err == nil {
		t.Error("a creature spell was improvised under Ironheart")
	}

	div := impHand(g, me, impCard(me.ID, "Divination", "Sorcery", "{2}{U}", divinationOracle))
	fillPoolColored(me, "U", 2)
	if err := g.CastSpell(me.ID, div, game.CastSpellParams{Strict: true, TapIDs: arts}); err != nil {
		t.Fatalf("improvising a noncreature spell with Ironheart: %v", err)
	}
	if !cgTapped(g, arts) {
		t.Error("the improvising artifacts were not tapped")
	}
}

func TestOrganicExtinctionDestroysNonartifactCreaturesOnly(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	advanceToMain(t, g)
	elf := pushCatalogPermanent(g, me.ID, "Elf", "Creature — Elf", "", false)
	golem := pushCatalogPermanent(g, me.ID, "Golem", "Artifact Creature — Golem", "", false)
	theirs := pushCatalogPermanent(g, opp.ID, "Ogre", "Creature — Ogre", "", false)
	arts := impArtifacts(g, me.ID, 8)
	id := impHand(g, me, impCard(me.ID, "Organic Extinction", "Sorcery", "{8}{W}{W}", organicExtinctionOracle))
	fillPoolColored(me, "W", 2)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, TapIDs: arts}); err != nil {
		t.Fatalf("improvising Organic Extinction: %v", err)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(elf) || g.Battlefield.Contains(theirs) {
		t.Error("nonartifact creatures survived")
	}
	if !g.Battlefield.Contains(golem) {
		t.Error("an artifact creature was destroyed")
	}
}

func TestMaverickThopteristMakesTwoFlyingThopters(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	arts := impArtifacts(g, me.ID, 3)
	id := impHand(g, me, impCard(me.ID, "Maverick Thopterist", "Creature — Human Artificer", "{3}{U}{R}", maverickThopteristOracle))
	fillPoolColored(me, "U", 1)
	fillPoolColored(me, "R", 1)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, TapIDs: arts}); err != nil {
		t.Fatalf("improvising Maverick Thopterist: %v", err)
	}
	passPriorityAroundTable(t, g)
	passPriorityAroundTable(t, g)
	n := 0
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == me.ID && c.Name == "Thopter" {
			n++
		}
	}
	if n != 2 {
		t.Errorf("%d Thopters, want 2", n)
	}
}

func TestBattleAtTheBridgeShrinksAndGains(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	victim := pushVanillaCreature(g, g.Seats[1].ID, "Ogre", 4, 4)
	arts := impArtifacts(g, me.ID, 2)
	id := impHand(g, me, impCard(me.ID, "Battle at the Bridge", "Sorcery", "{X}{B}", battleAtTheBridgeOracle))
	fillPoolColored(me, "B", 1)
	life := me.Life
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, XValue: 2, TapIDs: arts, Targets: b42CardTarget(victim)}); err != nil {
		t.Fatalf("X=2 paid by two artifacts and {B}: %v", err)
	}
	passPriorityAroundTable(t, g)
	if p := effectivePower(t, g, victim); p != 2 {
		t.Errorf("power %d, want 2", p)
	}
	if me.Life != life+2 {
		t.Errorf("life %d, want %d", me.Life, life+2)
	}
}

func TestSlyRequisitionerMakesAServoWhenANontokenArtifactDies(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	pushCatalogPermanent(g, me.ID, "Sly Requisitioner", "Creature — Human Artificer", slyRequisitionerOracle, false)
	trinket := pushCatalogPermanent(g, me.ID, "Trinket", "Artifact", "", false)
	servos := func() int {
		n := 0
		for _, c := range g.BattlefieldCardsForEffect() {
			if c.Controller == me.ID && strings.Contains(c.Name, "Servo") {
				n++
			}
		}
		return n
	}
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(trinket) })
	passPriorityAroundTable(t, g)
	if got := servos(); got != 1 {
		t.Fatalf("%d Servos after a nontoken artifact died, want 1", got)
	}
	// The Servo is a token: its death makes no more.
	for _, c := range g.BattlefieldCardsForEffect() {
		if strings.Contains(c.Name, "Servo") {
			id := c.InstanceID
			g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(id) })
		}
	}
	passPriorityAroundTable(t, g)
	if got := servos(); got != 0 {
		t.Errorf("%d Servos after the token died, want 0", got)
	}
}

// The bot enumerator prices the printed cost and offers no tap payments,
// as for convoke: improvise artifacts alone do not make a spell castable
// in its move list.
func TestEnumeratorOffersNoImproviseTaps(t *testing.T) {
	g, me, _ := tapTable(t)
	impArtifacts(g, me.ID, 5)
	card := handCardOf(g, me, "Foundry Assembler", "Artifact Creature — Assembly-Worker", "{5}", foundryAssemblerOracle)
	cast := func() []legal.Move {
		var out []legal.Move
		for _, m := range legal.EnumerateFor(g, me.ID) {
			if m.Type == legal.TypeCastSpell && m.Source == card {
				out = append(out, m)
			}
		}
		return out
	}
	if got := cast(); len(got) != 0 {
		t.Fatalf("a cast was enumerated with no mana, paid by artifacts alone: %v", got)
	}
	fillPoolColored(me, "C", 5)
	got := cast()
	if len(got) == 0 {
		t.Fatal("no cast enumerated once the printed cost is payable")
	}
	for _, m := range got {
		if strings.Contains(string(m.Params), "tap_ids") {
			t.Errorf("enumerated an improvise tap payment: %s", m.Params)
		}
	}
}
