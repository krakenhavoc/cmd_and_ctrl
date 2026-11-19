package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_bns_b40_test.go — Foundations Jumpstart bonus cards, S82 BNS batch 40.

const (
	bnsSunhomeStalwart   = "340d6218-d351-4b2e-80da-7d4afa77190a"
	bnsSunmanePegasus    = "4f7e7c59-e0a9-4b32-96b9-b1abf52f5ef7"
	bnsSunspearShikari   = "c49db060-2d4b-4dc8-8e54-7606e079c9fe"
	bnsScythecat         = "75d38954-6148-4d95-bf4b-cb82efcbe8df"
	bnsTethmosHighPriest = "f922150c-8f79-412b-a611-59064e3e3cea"
	bnsGuildEnforcer     = "84c0db59-255d-41af-a3d4-3f3548f1195f"
	bnsBloodlord         = "b04ba676-a1f5-4216-9d93-3c475b504bfa"
	bnsWaveMystic        = "121a1efa-03f2-4b5f-8e72-539de028457e"
	bnsTukatongue        = "c7dbd008-e941-497e-8305-80c53b1dffe2"
	bnsUkudCobra         = "d38cbb37-0ddd-4768-bb05-f35b38730e44"
	bnsUnderseaInvader   = "ff12dfee-7140-44bd-a62e-5f8ab98e93c6"
)

func TestS82BnsB40CardsAreRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		bnsSunhomeStalwart: "Sunhome Stalwart", bnsSunmanePegasus: "Sunmane Pegasus",
		bnsSunspearShikari: "Sunspear Shikari", bnsScythecat: "Territorial Scythecat",
		bnsTethmosHighPriest: "Tethmos High Priest", bnsGuildEnforcer: "Thieves' Guild Enforcer",
		bnsBloodlord: "Thirsting Bloodlord", bnsWaveMystic: "Towering-Wave Mystic",
		bnsTukatongue: "Tukatongue Thallid", bnsUkudCobra: "Ukud Cobra",
		bnsUnderseaInvader: "Undersea Invader",
	} {
		spec, ok := Lookup(oracle)
		if !ok || spec.Name != name || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: registered=%v spec=%+v", name, ok, spec)
		}
	}
}

// bnsSettle passes priority until the stack is empty, answering any target
// prompt with `pick`.
func bnsSettle(t *testing.T, g *game.Game, chooser uuid.UUID, pick *game.TargetRef) {
	t.Helper()
	for i := 0; i < 24; i++ {
		if p := latestPickTarget(g, chooser); p != nil {
			if pick == nil {
				t.Fatalf("unexpected pick_target prompt")
			}
			if err := g.ResolvePickTarget(p.ID, p.Chooser, *pick); err != nil {
				t.Fatalf("ResolvePickTarget: %v", err)
			}
			continue
		}
		if len(g.PendingChoices) > 0 || stackFullyEmpty(g) {
			return
		}
		passPriorityAroundTable(t, g)
	}
	t.Fatal("did not settle")
}

func bnsFillLibrary(p *game.Player, n int) {
	for i := 0; i < n; i++ {
		p.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: "Filler", TypeLine: "Sorcery", Owner: p.ID, Controller: p.ID})
	}
}

func bnsHasKeyword(t *testing.T, g *game.Game, id uuid.UUID, kw string) bool {
	t.Helper()
	c := layeredCard(t, g, id)
	return game.HasKeyword(&c, kw)
}

func TestS82BnsB40SunhomeStalwartMentorsALesserAttacker(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	stalwart := b12Push(g, me.ID, "Sunhome Stalwart", "Creature — Human Soldier", bnsSunhomeStalwart, 2, 2)
	small := b12Creature(g, me.ID, "Small", "Creature — Bear", 1, 1)
	big := b12Creature(g, me.ID, "Big", "Creature — Bear", 3, 3)
	assertKeywords(t, g, stalwart, "first strike")
	advanceTo(t, g, game.StepBeginCombat)
	passPriorityAroundTable(t, g)
	declareAttack(t, g, opp.ID, stalwart, small, big)
	if p := latestPickTarget(g, me.ID); p != nil {
		if err := g.ResolvePickTarget(p.ID, p.Chooser, game.TargetRef{Kind: game.TargetCard, ID: small}); err != nil {
			t.Fatal(err)
		}
	}
	answerAnyTriggerOrderPrompt(t, g, me.ID)
	passPriorityAroundTable(t, g)
	if got := countersOn(g, small, game.CounterPlusOne); got != 1 {
		t.Errorf("lesser attacker has %d counters, want 1", got)
	}
	if got := countersOn(g, big, game.CounterPlusOne) + countersOn(g, stalwart, game.CounterPlusOne); got != 0 {
		t.Errorf("an attacker without lesser power got %d counters", got)
	}
}

func TestS82BnsB40SunmanePegasusGainsVigilanceAndLifelinkUntilEndOfTurn(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	peg := b12Push(g, me.ID, "Sunmane Pegasus", "Creature — Pegasus", bnsSunmanePegasus, 2, 3)
	advanceToMain(t, g)
	assertKeywords(t, g, peg, "flying")
	if bnsHasKeyword(t, g, peg, "vigilance") || bnsHasKeyword(t, g, peg, "lifelink") {
		t.Fatal("has vigilance or lifelink before the ability")
	}
	if err := g.ActivateCatalogAbility(me.ID, peg, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("activated with no mana")
	}
	floatForTest(g, me, "1W")
	if err := g.ActivateCatalogAbility(me.ID, peg, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !bnsHasKeyword(t, g, peg, "vigilance") || !bnsHasKeyword(t, g, peg, "lifelink") {
		t.Error("did not gain both keywords")
	}
}

func TestS82BnsB40SunspearShikariNeedsAnEquipment(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	shikari := b12Push(g, me.ID, "Sunspear Shikari", "Creature — Cat Soldier", bnsSunspearShikari, 2, 2)
	advanceToMain(t, g)
	if bnsHasKeyword(t, g, shikari, "first strike") || bnsHasKeyword(t, g, shikari, "lifelink") {
		t.Fatal("unequipped Shikari has first strike or lifelink")
	}
	equip := pushCatalogPermanent(g, me.ID, "Batterbone", "Artifact — Equipment", "389d459d-446a-4b84-82b2-a30fe6ced11f", false)
	floatForTest(g, me, "5")
	equipTo(t, g, me.ID, equip, shikari)
	if !bnsHasKeyword(t, g, shikari, "first strike") || !bnsHasKeyword(t, g, shikari, "lifelink") {
		t.Error("equipped Shikari lacks first strike or lifelink")
	}
	other := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	if bnsHasKeyword(t, g, other, "first strike") {
		t.Error("another creature gained first strike")
	}
}

func TestS82BnsB40ScythecatGrowsOnYourLandsOnly(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	cat := b12Push(g, me.ID, "Territorial Scythecat", "Creature — Cat", bnsScythecat, 2, 1)
	assertKeywords(t, g, cat, "trample")
	advanceToMain(t, g)
	b31PlayLand(t, g, "Forest", "Basic Land — Forest", "")
	passPriorityAroundTable(t, g)
	if got := countersOn(g, cat, game.CounterPlusOne); got != 1 {
		t.Fatalf("%d counters after landfall, want 1", got)
	}
	_ = opp
}

func TestS82BnsB40TethmosHighPriestReturnsATwoDropWhenTargeted(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	priest := b12Push(g, me.ID, "Tethmos High Priest", "Creature — Cat Cleric", bnsTethmosHighPriest, 2, 3)
	other := b12Creature(g, me.ID, "Other", "Creature — Bear", 2, 2)
	src := b12Creature(g, opp.ID, "Src", "Creature — Bear", 3, 3)
	_ = src
	cheap := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: cheap, Name: "Cheap", TypeLine: "Creature — Bear", ManaCost: "{1}{G}", Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID})
	dear := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: dear, Name: "Dear", TypeLine: "Creature — Giant", ManaCost: "{3}{G}", Power: 4, Toughness: 4, Owner: me.ID, Controller: me.ID})
	advanceToMain(t, g)
	// A spell that targets another creature does not trigger heroic.
	castCatalogSpell(t, g, "Djeru's Resolve", "Instant", pr7aDjerusResolve, []game.TargetRef{pr7aCard(other)})
	bnsSettle(t, g, me.ID, nil)
	if g.Battlefield.Contains(cheap) {
		t.Fatal("heroic triggered on a spell targeting another creature")
	}
	castCatalogSpell(t, g, "Djeru's Resolve", "Instant", pr7aDjerusResolve, []game.TargetRef{pr7aCard(priest)})
	bnsSettle(t, g, me.ID, &game.TargetRef{Kind: game.TargetCard, ID: cheap})
	if !g.Battlefield.Contains(cheap) {
		t.Error("the mana value 2 creature card did not return")
	}
	if g.Battlefield.Contains(dear) {
		t.Error("the mana value 4 creature card returned")
	}
}

func TestS82BnsB40GuildEnforcerMillsForItselfAndOtherRoguesAndGrowsWithAFullGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bnsFillLibrary(opp, 20)
	advanceToMain(t, g)
	enf := b11Cast(t, g, me, "Thieves' Guild Enforcer", "Creature — Human Rogue", bnsGuildEnforcer, 1, 1)
	if got := opp.Graveyard.Size(); got != 2 {
		t.Fatalf("opponent graveyard %d after the Enforcer entered, want 2", got)
	}
	b11Cast(t, g, me, "Rogue Friend", "Creature — Human Rogue", "", 1, 1)
	if got := opp.Graveyard.Size(); got != 4 {
		t.Fatalf("opponent graveyard %d after another Rogue entered, want 4", got)
	}
	b11Cast(t, g, me, "Bear", "Creature — Bear", "", 2, 2)
	if got := opp.Graveyard.Size(); got != 4 {
		t.Fatalf("a non-Rogue milled: graveyard %d, want 4", got)
	}
	if c := layeredCard(t, g, enf); c.CurrentPower() != 1 || c.CurrentToughness() != 1 || bnsHasKeyword(t, g, enf, "deathtouch") {
		t.Fatalf("Enforcer is %d/%d before eight cards", c.CurrentPower(), c.CurrentToughness())
	}
	for i := 0; i < 4; i++ {
		pushGraveyardCardWithTypeLine(g, opp.ID, "G", "Sorcery")
	}
	if c := layeredCard(t, g, enf); c.CurrentPower() != 3 || c.CurrentToughness() != 2 || !bnsHasKeyword(t, g, enf, "deathtouch") {
		t.Errorf("Enforcer is %d/%d deathtouch=%v with eight cards, want 3/2 deathtouch", c.CurrentPower(), c.CurrentToughness(), bnsHasKeyword(t, g, enf, "deathtouch"))
	}
	assertKeywords(t, g, enf, "flash")
}

func TestS82BnsB40ThirstingBloodlordPumpsOtherVampiresYouControl(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	lord := b12Push(g, me.ID, "Thirsting Bloodlord", "Creature — Vampire", bnsBloodlord, 3, 3)
	mine := b12Creature(g, me.ID, "Fang", "Creature — Vampire", 2, 2)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Fang", "Creature — Vampire", 2, 2)
	for id, want := range map[uuid.UUID]int{lord: 3, mine: 3, bear: 2, theirs: 2} {
		if c := layeredCard(t, g, id); c.CurrentPower() != want {
			t.Errorf("power %d, want %d", c.CurrentPower(), want)
		}
	}
}

func TestS82BnsB40ToweringWaveMysticMillsThatManyCards(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mystic := b12Push(g, me.ID, "Towering-Wave Mystic", "Creature — Merfolk Wizard", bnsWaveMystic, 2, 1)
	bnsFillLibrary(opp, 10)
	bnsFillLibrary(me, 10)
	other := b12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	advanceToMain(t, g)
	emitNoncombatDamage(g, other, mystic, 2)
	bnsSettle(t, g, me.ID, nil)
	if opp.Graveyard.Size() != 0 {
		t.Fatal("damage dealt to the Mystic triggered it")
	}
	emitNoncombatDamage(g, mystic, other, 2)
	bnsSettle(t, g, me.ID, &game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID})
	if got := opp.Graveyard.Size(); got != 2 {
		t.Errorf("target player milled %d, want 2", got)
	}
	dealCombatDamageToPlayer(g, mystic, opp.ID, 2)
	bnsSettle(t, g, me.ID, &game.TargetRef{Kind: game.TargetPlayer, ID: me.ID})
	if got := me.Graveyard.Size(); got != 2 {
		t.Errorf("the controller, targeted, milled %d, want 2", got)
	}
}

func TestS82BnsB40TukatongueThallidLeavesASaprolingWhenItDies(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	thallid := b12Push(g, me.ID, "Tukatongue Thallid", "Creature — Fungus", bnsTukatongue, 1, 1)
	advanceToMain(t, g)
	b11Die(t, g, me.ID, thallid)
	if n := cardsNamed(g, "Saproling", me.ID); n != 1 {
		t.Fatalf("%d Saprolings, want 1", n)
	}
}

func TestS82BnsB40UkudCobraHasDeathtouch(t *testing.T) {
	g := newCatalogGame(t)
	cobra := b12Push(g, g.Seats[0].ID, "Ukud Cobra", "Creature — Snake", bnsUkudCobra, 2, 5)
	assertKeywords(t, g, cobra, "deathtouch")
}

func TestS82BnsB40UnderseaInvaderHasFlashAndEntersTapped(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	id := b11Cast(t, g, me, "Undersea Invader", "Creature — Giant Rogue", bnsUnderseaInvader, 5, 6)
	if !battlefieldCardCopy(t, g, id).Tapped {
		t.Error("the Invader entered untapped")
	}
	assertKeywords(t, g, id, "flash")
}
