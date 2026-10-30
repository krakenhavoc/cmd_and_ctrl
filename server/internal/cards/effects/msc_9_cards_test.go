package effects

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// msc_9_cards_test.go — Marvel Super Heroes Commander, ninth batch
// (#4455). Each test plays the real catalog entry and asserts what its
// text says.

const (
	msc9AbsorbingManOracle  = "a1afb796-1b0b-4f41-9c8f-5320c9af5989"
	msc9AncestralOracle     = "a45bb088-16fd-4d91-8fd6-3c52f1c24748"
	msc9ArmedAssailantOracl = "53d7722a-b60b-4520-abce-f31b1e6bba00"
	msc9AstonishingOracle   = "24059283-edbd-4817-9f35-4be02f07b8f8"
	msc9QuinjetOracle       = "905e4bd1-f5f2-4691-b13f-ce2628d26a38"
	msc9BastOracle          = "e3233686-7703-460c-80ee-8cd37b3014ff"
	msc9BionicBlowOracle    = "27683e43-7c3f-45b3-8ca4-968428000315"
	msc9PantherOracle       = "7ba593fd-93db-452f-ad80-7f0f37afb09d"
	msc9BlurOracle          = "b4a9b328-f976-4390-be9a-711ef357a50d"
	msc9BrawnOracle         = "19910a90-d334-4d99-89f2-cd81cea3c824"
	msc9TeamLeaderOracle    = "916cf34a-fd2c-47fe-bebb-f7b52b3dabed"
)

func TestMsc9EveryCardIsRegisteredFull(t *testing.T) {
	want := map[string]string{
		msc9AbsorbingManOracle:  "Absorbing Man and Titania",
		msc9AncestralOracle:     "Ancestral Communion",
		msc9ArmedAssailantOracl: "Armed Assailant",
		msc9AstonishingOracle:   "Astonishing Spider-Man",
		msc9QuinjetOracle:       "Avengers Quinjet",
		msc9BastOracle:          "Bast, Panther Goddess",
		msc9BionicBlowOracle:    "Bionic Blow",
		msc9PantherOracle:       "Black Panther, Most Dangerous",
		msc9BlurOracle:          "Blur of Heroism",
		msc9BrawnOracle:         "Brawn, Amadeus Cho",
		msc9TeamLeaderOracle:    "Captain America, Team Leader",
	}
	for oracle, name := range want {
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

func msc9Hand(g *game.Game, p *game.Player, name, typeLine, oracle string) uuid.UUID {
	id := uuid.New()
	g.WithWriteLock(func() {
		p.Hand.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, OracleID: oracle,
			Owner: p.ID, Controller: p.ID, Power: 2, Toughness: 2})
	})
	return id
}

func msc9Attach(t *testing.T, g *game.Game, equipment, host uuid.UUID) {
	t.Helper()
	g.WithWriteLock(func() {
		if err := g.AttachForEffect(equipment, game.TargetRef{Kind: game.TargetCard, ID: host}); err != nil {
			t.Fatalf("AttachForEffect: %v", err)
		}
	})
}

// --- Absorbing Man and Titania -------------------------------------------

func TestMsc9AbsorbingManDoublesYourCreaturesDamageOnly(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Push(g, me.ID, "Absorbing Man and Titania", "Legendary Creature — Human Villain", msc9AbsorbingManOracle, 4, 5)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	before := opp.Life
	attackWith(t, g, opp.ID, bear)
	passPriorityAroundTable(t, g)
	if opp.Life != before-4 {
		t.Errorf("a 2-power attacker: %d -> %d, want -4", before, opp.Life)
	}
	before = opp.Life
	b12Bolt(t, g, game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID})
	if opp.Life != before-3 {
		t.Errorf("a Bolt is not a creature source: %d -> %d, want -3", before, opp.Life)
	}
	theirs := b12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	mine := me.Life
	g.WithWriteLock(func() { _ = g.DealDamageToPlayerForEffect(theirs, me.ID, 2) })
	if me.Life != mine-2 {
		t.Errorf("an opponent's creature is not doubled: %d -> %d, want -2", mine, me.Life)
	}
}

// --- Ancestral Communion -------------------------------------------------

func msc9Commander(g *game.Game, owner uuid.UUID) {
	id := b12Creature(g, owner, "My Commander", "Legendary Creature — Human", 2, 2)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == id {
				g.Battlefield.Cards[i].IsCommander = true
			}
		}
	})
}

func TestMsc9AncestralCommunionReturnsAPermanentCard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	rock := batch01GraveyardCard(me, "Rock", "Artifact")
	spell := batch01GraveyardCard(me, "Bolt", "Instant")
	castCatalogSpell(t, g, "Ancestral Communion", "Sorcery", msc9AncestralOracle, []game.TargetRef{{Kind: game.TargetCard, ID: rock}})
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(rock) {
		t.Error("the permanent card returns to hand")
	}
	if !me.Graveyard.Contains(spell) {
		t.Error("the instant stays in the graveyard")
	}
	if g.Stack.Size() != 0 {
		t.Errorf("no commander, no copy: the stack holds %d", g.Stack.Size())
	}
	if err := castCatalogSpellErr(t, g, "Ancestral Communion", "Sorcery", msc9AncestralOracle, []game.TargetRef{{Kind: game.TargetCard, ID: spell}}); err == nil {
		t.Error("an instant card is not a legal target")
	}
}

func TestMsc9AncestralCommunionCopiesWhileYouControlYourCommander(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	msc9Commander(g, me.ID)
	a := batch01GraveyardCard(me, "Rock A", "Artifact")
	b := batch01GraveyardCard(me, "Rock B", "Artifact")
	castCatalogSpell(t, g, "Ancestral Communion", "Sorcery", msc9AncestralOracle, []game.TargetRef{{Kind: game.TargetCard, ID: a}})
	// The cast trigger goes on the stack above the spell and resolves
	// first, copying the spell; the copy asks for a new target.
	for i := 0; i < 60 && (g.Stack.Size() > 0 || len(g.PendingChoices) > 0); i++ {
		if ch := pickTargetChoiceForMsc9(g, me.ID); ch != nil {
			if err := g.ResolvePickTarget(ch.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: b}); err != nil {
				t.Fatalf("new target for the copy: %v", err)
			}
			continue
		}
		if ch := latestConfirmFor(g, me.ID); ch != nil {
			if err := g.ResolveConfirm(ch.ID, me.ID, true); err != nil {
				t.Fatalf("confirm: %v", err)
			}
			continue
		}
		if err := g.PassPriority(); err != nil && !errors.Is(err, game.ErrChoicePending) {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	if !me.Hand.Contains(a) || !me.Hand.Contains(b) {
		t.Errorf("the spell and its copy return both cards: a in hand %v, b in hand %v", me.Hand.Contains(a), me.Hand.Contains(b))
	}
}

func pickTargetChoiceForMsc9(g *game.Game, chooser uuid.UUID) *game.PendingChoice {
	return latestPickTarget(g, chooser)
}

// --- Armed Assailant -----------------------------------------------------

func TestMsc9ArmedAssailantGrowsAndGainsMenaceWhileEquipped(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	assailant := b12Push(g, me.ID, "Armed Assailant", "Creature — Human Villain", msc9ArmedAssailantOracl, 1, 3)
	assertKeywords(t, g, assailant, "deathtouch")
	if hasEffectiveKeyword(t, g, assailant, "menace") || effectivePower(t, g, assailant) != 1 {
		t.Fatalf("unequipped: power %d, menace %v", effectivePower(t, g, assailant), hasEffectiveKeyword(t, g, assailant, "menace"))
	}
	sword := b12Permanent(g, me.ID, "Sword", "Artifact — Equipment")
	msc9Attach(t, g, sword, assailant)
	if got := effectivePower(t, g, assailant); got != 3 {
		t.Errorf("equipped power = %d, want 3", got)
	}
	if !hasEffectiveKeyword(t, g, assailant, "menace") {
		t.Error("equipped, it has menace")
	}
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(sword) })
	if got := effectivePower(t, g, assailant); got != 1 {
		t.Errorf("the Equipment gone: power %d, want 1", got)
	}
	if hasEffectiveKeyword(t, g, assailant, "menace") {
		t.Error("the Equipment gone: no menace")
	}
}

// --- Astonishing Spider-Man ----------------------------------------------

func TestMsc9AstonishingSpiderManDiscardsTheHandAndDrawsForEachDiscardedThisTurn(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	spidey := b12Push(g, me.ID, "Astonishing Spider-Man", "Legendary Creature — Spider Human Hero", msc9AstonishingOracle, 4, 4)
	assertKeywords(t, g, spidey, "reach")
	seedLibrary(me, "One", "Two", "Three", "Four", "Five", "Six", "Seven", "Eight", "Nine")
	g.WithWriteLock(func() { me.Hand.Cards = me.Hand.Cards[:3] })
	declareAttack(t, g, opp.ID, spidey)
	passPriorityAroundTable(t, g)
	answerMayChoice(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != 3 {
		t.Errorf("three cards discarded, three drawn: hand %d", got)
	}
	if got := g.TurnTallyFor(me.ID).CardsDiscarded; got != 3 {
		t.Errorf("discard tally %d, want 3", got)
	}
}

func TestMsc9AstonishingSpiderManCountsEarlierDiscardsAndMayDecline(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	spidey := b12Push(g, me.ID, "Astonishing Spider-Man", "Legendary Creature — Spider Human Hero", msc9AstonishingOracle, 4, 4)
	seedLibrary(me, "One", "Two", "Three", "Four", "Five", "Six", "Seven", "Eight", "Nine")
	g.WithWriteLock(func() { me.Hand.Cards = me.Hand.Cards[:2]; _ = g.DiscardRandomForEffect(me.ID, 1) })
	// One card was discarded before the attack; one is left in hand.
	declareAttack(t, g, opp.ID, spidey)
	passPriorityAroundTable(t, g)
	answerMayChoice(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != 2 {
		t.Errorf("two discarded this turn, two drawn: hand %d", got)
	}

	g2 := newCatalogGame(t)
	me2, opp2 := g2.Seats[0], g2.Seats[1]
	spidey2 := b12Push(g2, me2.ID, "Astonishing Spider-Man", "Legendary Creature — Spider Human Hero", msc9AstonishingOracle, 4, 4)
	before := me2.Hand.Size()
	declareAttack(t, g2, opp2.ID, spidey2)
	passPriorityAroundTable(t, g2)
	answerMayChoice(t, g2, me2.ID, false)
	passPriorityAroundTable(t, g2)
	if me2.Hand.Size() != before {
		t.Errorf("declined: hand %d -> %d", before, me2.Hand.Size())
	}
}

func TestMsc9AstonishingSpiderManWithNoHandDrawsNothing(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	spidey := b12Push(g, me.ID, "Astonishing Spider-Man", "Legendary Creature — Spider Human Hero", msc9AstonishingOracle, 4, 4)
	g.WithWriteLock(func() { me.Hand.Cards = nil })
	declareAttack(t, g, opp.ID, spidey)
	passPriorityAroundTable(t, g)
	if c := latestConfirmFor(g, me.ID); c != nil {
		answerMayChoice(t, g, me.ID, true)
	}
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != 0 {
		t.Errorf("nothing was discarded, so nothing is drawn: hand %d", got)
	}
}

// --- Avengers Quinjet ----------------------------------------------------

func TestMsc9QuinjetReturnsAHeroCreatureCardFromTheGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	hero := batch01GraveyardCard(me, "Hero", "Creature — Human Hero")
	villain := batch01GraveyardCard(me, "Villain", "Creature — Human Villain")
	castAndResolveCreature(t, g, "Avengers Quinjet", "Artifact — Vehicle", msc9QuinjetOracle)
	mode := latestChoiceOfKindFor(g, game.PendingChoiceModePick, me.ID)
	if mode == nil {
		t.Fatalf("no mode prompt from the enters trigger: %+v", g.PendingChoices)
	}
	if err := g.ResolveModePick(mode.ID, me.ID, []int{1}); err != nil {
		t.Fatalf("choosing the return bullet: %v", err)
	}
	if err := sgAnswerTargets(t, g, villain); !errors.Is(err, game.ErrIllegalTarget) {
		t.Fatalf("a non-Hero card: err = %v, want ErrIllegalTarget", err)
	}
	if err := sgAnswerTargets(t, g, hero); err != nil {
		t.Fatalf("a Hero card: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(hero) || !me.Graveyard.Contains(villain) {
		t.Error("only the Hero card returns")
	}
}

func TestMsc9QuinjetPutsAHeroCreatureFromHandOntoTheBattlefield(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	hero := msc9Hand(g, me, "Hero", "Creature — Human Hero", "")
	bear := msc9Hand(g, me, "Bear", "Creature — Bear", "")
	castAndResolveCreature(t, g, "Avengers Quinjet", "Artifact — Vehicle", msc9QuinjetOracle)
	mode := latestChoiceOfKindFor(g, game.PendingChoiceModePick, me.ID)
	if mode == nil {
		t.Fatalf("no mode prompt: %+v", g.PendingChoices)
	}
	if err := g.ResolveModePick(mode.ID, me.ID, []int{0}); err != nil {
		t.Fatalf("choosing the put bullet: %v", err)
	}
	passPriorityAroundTable(t, g)
	pick := chooseCardsChoiceFor(g, me.ID)
	if pick == nil {
		t.Fatal("the Hero put offered nothing")
	}
	for _, id := range pick.ChooseCards {
		if id == bear {
			t.Error("a non-Hero creature card is offered")
		}
	}
	answerChooseCards(t, g, me.ID, hero)
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(hero) {
		t.Error("the Hero enters")
	}
	if !me.Hand.Contains(bear) {
		t.Error("the Bear stays in hand")
	}
}

// --- Bast, Panther Goddess -----------------------------------------------

func TestMsc9BastNeedsThreeCreaturesToAttackOrBlock(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bast := b12Push(g, me.ID, "Bast, Panther Goddess", "Legendary Creature — Cat God", msc9BastOracle, 4, 4)
	assertKeywords(t, g, bast, "reach", "trample", "indestructible")
	other := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	advanceToMain(t, g)
	if c, ok := g.LookupCardForEffect(bast); !ok || !game.Restricted(&c, game.CantBlock) || !game.Restricted(&c, game.CantAttack) {
		t.Fatal("with two creatures Bast can neither attack nor block")
	}
	_ = other
	castHoldCreature(t, g, "Cub", "Creature — Cat", "", 1, 1)
	if c, ok := g.LookupCardForEffect(bast); !ok || game.Restricted(&c, game.CantBlock) || game.Restricted(&c, game.CantAttack) {
		t.Fatal("with three creatures Bast is free")
	}
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(bast, opp.ID); err != nil {
		t.Fatalf("Bast attacks with three creatures: %v", err)
	}
}

func TestMsc9BastPumpsAnAttackerByYourCreatureCount(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bast := b12Push(g, me.ID, "Bast, Panther Goddess", "Legendary Creature — Cat God", msc9BastOracle, 4, 4)
	b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	b12Creature(g, me.ID, "Cub", "Creature — Cat", 1, 1)
	b12Creature(g, me.ID, "Fox", "Creature — Fox", 1, 1)
	declareAttack(t, g, opp.ID, bast)
	if latestPickTarget(g, me.ID) != nil {
		pickCard(t, g, me.ID, bast)
	}
	passPriorityAroundTable(t, g)
	if got := effectivePower(t, g, bast); got != 8 {
		t.Errorf("four creatures: Bast is 4+4 = 8 power, got %d", got)
	}
	if got := effectiveToughness(t, g, bast); got != 8 {
		t.Errorf("toughness %d, want 8", got)
	}
}

// --- Bionic Blow ---------------------------------------------------------

func msc9CastBionicBlow(t *testing.T, g *game.Game, me *game.Player, x int, targets ...uuid.UUID) error {
	t.Helper()
	id := msc9Hand(g, me, "Bionic Blow", "Sorcery", msc9BionicBlowOracle)
	advanceToMain(t, g)
	refs := make([]game.TargetRef, len(targets))
	for i, tid := range targets {
		refs[i] = game.TargetRef{Kind: game.TargetCard, ID: tid}
	}
	return g.CastSpell(me.ID, id, game.CastSpellParams{XValue: x, Targets: refs})
}

func TestMsc9BionicBlowPumpsThenBitesAnotherCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := b12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	foe := b12Creature(g, opp.ID, "Foe", "Creature — Giant", 1, 9)
	if err := msc9CastBionicBlow(t, g, me, 3, mine, foe); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := effectivePower(t, g, mine); got != 5 {
		t.Errorf("2 +3 = 5 power, got %d", got)
	}
	if got := damageMarkedOn(g, foe); got != 5 {
		t.Errorf("the bite uses the pumped power: %d damage, want 5", got)
	}
}

func TestMsc9BionicBlowWithNoSecondTargetOnlyPumps(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	mine := b12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	if err := msc9CastBionicBlow(t, g, me, 2, mine); err != nil {
		t.Fatalf("cast with one target: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := effectivePower(t, g, mine); got != 4 {
		t.Errorf("power %d, want 4", got)
	}
	if got := damageMarkedOn(g, mine); got != 0 {
		t.Errorf("it bit itself for %d", got)
	}
}

func TestMsc9BionicBlowSecondTargetMustBeAnotherCreature(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	mine := b12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	if err := msc9CastBionicBlow(t, g, me, 1, mine, mine); err == nil {
		t.Error("the same creature in both slots was accepted")
	}
}

// --- Black Panther, Most Dangerous ---------------------------------------

func TestMsc9BlackPantherReflectsDamageAtAnyOtherTarget(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	panther := b27Push(g, me.ID, "Black Panther, Most Dangerous", "Legendary Creature — Human Warrior Hero", msc9PantherOracle, "{2}{W}", 3, 3, "W")
	foe := b12Creature(g, opp.ID, "Foe", "Creature — Bear", 2, 2)
	life := opp.Life
	b27Damage(g, foe, panther, 2)
	b04WaitForPick(t, g, me.ID)
	b16PickPlayer(t, g, me.ID, opp.ID)
	passPriorityAroundTable(t, g)
	if opp.Life != life-2 {
		t.Errorf("2 damage reflected at the opponent: %d -> %d", life, opp.Life)
	}
}

func TestMsc9BlackPantherPowerUpGrowsAndPumpsTheOthers(t *testing.T) {
	g := newCatalogGame(t)
	powerUpMain(t, g)
	me := g.Seats[0]
	panther := pushPowerUpPermanent(g, me.ID, "Black Panther, Most Dangerous", "Legendary Creature — Human Warrior Hero", "{2}{W}", msc9PantherOracle)
	friend := b12Creature(g, me.ID, "Friend", "Creature — Bear", 2, 2)
	foe := b12Creature(g, g.Seats[1].ID, "Foe", "Creature — Bear", 2, 2)
	floatMana(t, g, me, "{C}{C}{C}{C}{C}{W}{W}")
	if err := g.ActivateCatalogAbility(me.ID, panther, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("power up: %v", err)
	}
	resolveTopOfStack(t, g)
	if got := plusOnes(g, panther); got != 2 {
		t.Errorf("two counters, got %d", got)
	}
	if got := effectivePowerWithCounters(t, g, panther); got != 4 {
		t.Errorf("Panther is 2 +2 counters = 4, no pump on himself: %d", got)
	}
	if got := effectivePower(t, g, friend); got != 4 {
		t.Errorf("another creature you control gets +2/+2: power %d", got)
	}
	if got := effectivePower(t, g, foe); got != 2 {
		t.Errorf("an opponent's creature is untouched: power %d", got)
	}
}

// --- Blur of Heroism -----------------------------------------------------

func TestMsc9BlurDrawsWhenItEnters(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	before := me.Hand.Size()
	castCatalogSpell(t, g, "Blur of Heroism", "Enchantment", msc9BlurOracle, nil)
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != before+1 {
		t.Errorf("hand %d -> %d, want one card drawn", before, got)
	}
}

func TestMsc9BlurLetsYouFlashInHeroSpellsOnly(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushCatalogPermanent(g, me.ID, "Blur of Heroism", "Enchantment", msc9BlurOracle, false)
	hero := eacHandCard(g, me, "Hero", "Creature — Human Hero", "{1}")
	bear := eacHandCard(g, me, "Bear", "Creature — Bear", "{1}")
	ssToBeginCombat(t, g)
	if err := g.CastSpell(me.ID, bear, game.CastSpellParams{}); !errors.Is(err, game.ErrSorcerySpeedRequired) {
		t.Fatalf("a non-Hero creature at instant speed: err = %v, want ErrSorcerySpeedRequired", err)
	}
	if err := g.CastSpell(me.ID, hero, game.CastSpellParams{}); err != nil {
		t.Fatalf("a Hero creature at instant speed: %v", err)
	}
}

// --- Brawn, Amadeus Cho --------------------------------------------------

func TestMsc9BrawnDrawsAndPowerUpCountsTheHand(t *testing.T) {
	g := newCatalogGame(t)
	powerUpMain(t, g)
	me := g.Seats[0]
	before := me.Hand.Size()
	brawn := castPowerUpCreature(t, g, me, "Brawn, Amadeus Cho", "Legendary Creature — Gamma Scientist Hero", "{1}{G/U}", "{C}{G}", msc9BrawnOracle, 1)
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != before+1 {
		t.Fatalf("the enters trigger draws a card: hand %d -> %d", before, got)
	}
	hand := me.Hand.Size()
	floatMana(t, g, me, "{C}{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, brawn, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("power up for {3}: %v", err)
	}
	resolveTopOfStack(t, g)
	if got := plusOnes(g, brawn); got != hand {
		t.Errorf("one counter per card in hand: %d counters, hand %d", got, hand)
	}
	floatMana(t, g, me, "{C}{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, brawn, 0, game.ActivateAbilityParams{}); !errors.Is(err, game.ErrPowerUpUsed) {
		t.Errorf("second activation: err = %v, want ErrPowerUpUsed", err)
	}
}

// --- Captain America, Team Leader ----------------------------------------

func TestMsc9TeamLeaderBoostsAnotherHeroAndItself(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	cap := b12Push(g, me.ID, "Captain America, Team Leader", "Legendary Creature — Human Soldier Hero", msc9TeamLeaderOracle, 3, 3)
	hero := castHoldCreature(t, g, "Hero", "Creature — Human Hero", "", 2, 2)
	if got := counterOn(g, hero, game.CounterPlusOne); got != 1 {
		t.Errorf("the Hero gets a counter: %d", got)
	}
	if got := counterOn(g, cap, game.CounterPlusOne); got != 1 {
		t.Errorf("Captain America gets a counter: %d", got)
	}
	if !hasEffectiveKeyword(t, g, hero, "haste") || !hasEffectiveKeyword(t, g, hero, "vigilance") {
		t.Error("the Hero gains vigilance and haste")
	}
}

func TestMsc9TeamLeaderIgnoresNonHeroesAndOpposingHeroes(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	cap := b12Push(g, me.ID, "Captain America, Team Leader", "Legendary Creature — Human Soldier Hero", msc9TeamLeaderOracle, 3, 3)
	castHoldCreature(t, g, "Bear", "Creature — Bear", "", 2, 2)
	if got := counterOn(g, cap, game.CounterPlusOne); got != 0 {
		t.Errorf("a non-Hero entering: %d counters on Captain America", got)
	}
	b12Creature(g, opp.ID, "Their Hero", "Creature — Human Hero", 2, 2)
	passPriorityAroundTable(t, g)
	if got := counterOn(g, cap, game.CounterPlusOne); got != 0 {
		t.Errorf("an opposing Hero: %d counters on Captain America", got)
	}
}
