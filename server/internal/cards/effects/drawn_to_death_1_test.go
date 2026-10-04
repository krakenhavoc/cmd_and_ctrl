package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// drawn_to_death_1_test.go — S58 PR 8, the Drawn To Death deck's
// hand-size punishers and Swamp counts. #2063.

const (
	ironMaidenOracle          = "e0351638-9224-4779-8f7c-f2b11d134a5f"
	viselingOracle            = "05b9dccb-2364-4bca-a12e-8cb54a3f3a0b"
	theRackOracle             = "3d873e1d-4fac-42c4-bb31-77e76099e1ef"
	darkSuspicionsOracle      = "cf5cf174-d976-4cd8-9721-c30fe784fb70"
	masterOfTheFeastOracle    = "01dc86f3-5ebb-4b10-bf68-8fc9f232c724"
	indulgentTormentorOracle  = "8b202c63-c961-4590-958f-d17e76610ab5"
	entropicSpecterOracle     = "a9c44443-85c7-4d0a-82ab-36137151e4e5"
	glassesOfUrzaOracle       = "af7fabf4-8d55-4b06-9c21-472f4a5775b4"
	defileOracle              = "49dddec0-d958-4810-8c6e-225fc8118c8f"
	consumingCorruptionOracle = "3bc0bd36-70be-4180-9fb3-b10ce054107b"
	magusOfTheCoffersOracle   = "23dd895c-92bd-4af0-8b1a-7d76ca49178f"
	cryptGhastOracle          = "a3c8d817-7949-4dae-b9f5-f9d952479270"
	nirkanaRevenantOracle     = "6dff1def-5b94-40c8-a942-dea8743ab47c"
	bogWitchOracle            = "85573cba-07ae-4421-a167-a8569f85c0f7"
)

// d2dSetHand makes a seat's hand exactly n filler cards.
func d2dSetHand(p *game.Player, n int) {
	p.Hand.Cards = nil
	for i := 0; i < n; i++ {
		p.Hand.PushTop(game.Card{InstanceID: uuid.New(), Name: "filler", Owner: p.ID, Controller: p.ID})
	}
}

// --- Iron Maiden and Viseling ----------------------------------------

func TestIronMaidenAndViselingDamageAnOpponentsOversizedHand(t *testing.T) {
	for _, tc := range []struct {
		name, oracle, typeLine string
	}{
		{"Iron Maiden", ironMaidenOracle, "Artifact"},
		{"Viseling", viselingOracle, "Artifact Creature — Phyrexian Construct"},
	} {
		for _, hand := range []int{3, 4, 7} {
			g := newCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			pushPermanentForTest(g, me.ID, tc.name, tc.oracle, tc.typeLine)
			advanceToUpkeepOf(t, g, 1)
			d2dSetHand(opp, hand)
			start := opp.Life
			passPriorityAroundTable(t, g)
			want := hand - 4
			if want < 0 {
				want = 0
			}
			if opp.Life != start-want {
				t.Errorf("%s, hand of %d: life %d -> %d, want -%d", tc.name, hand, start, opp.Life, want)
			}
		}
	}
}

func TestIronMaidenIgnoresItsControllersOwnUpkeep(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[1]
	pushPermanentForTest(g, me.ID, "Iron Maiden", ironMaidenOracle, "Artifact")
	advanceToUpkeepOf(t, g, 1)
	d2dSetHand(me, 9)
	start := me.Life
	passPriorityAroundTable(t, g)
	if me.Life != start {
		t.Errorf("the controller's own upkeep dealt %d", start-me.Life)
	}
}

// --- The Rack ---------------------------------------------------------

func TestTheRackPunishesOnlyTheChosenOpponent(t *testing.T) {
	g := newCatalogGame(t)
	me, chosen, other := g.Seats[0], g.Seats[1], g.Seats[2]
	castCatalogSpell(t, g, "The Rack", "Artifact", theRackOracle, nil)
	passPriorityAroundTable(t, g)
	prompt := pendingOfKind(g, game.PendingChoiceOptionPick)
	if prompt == nil {
		t.Fatalf("no as-enters prompt: %+v", g.PendingChoices)
	}
	for _, opt := range prompt.PickOptions {
		if opt.Player == me.ID {
			t.Error("choose an OPPONENT: the controller was offered")
		}
	}
	answerChosenPlayer(t, g, me.ID, chosen.ID)

	advanceToUpkeepOf(t, g, 2)
	d2dSetHand(other, 0)
	start := other.Life
	passPriorityAroundTable(t, g)
	if other.Life != start {
		t.Errorf("an unchosen opponent took %d", start-other.Life)
	}

	advanceToUpkeepOf(t, g, 1)
	d2dSetHand(chosen, 1)
	start = chosen.Life
	passPriorityAroundTable(t, g)
	if chosen.Life != start-2 {
		t.Errorf("hand of 1: life %d -> %d, want -2 (3 minus 1)", start, chosen.Life)
	}
}

// --- Dark Suspicions --------------------------------------------------

func TestDarkSuspicionsComparesTheTwoHands(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushPermanentForTest(g, me.ID, "Dark Suspicions", darkSuspicionsOracle, "Enchantment")
	d2dSetHand(me, 2)
	advanceToUpkeepOf(t, g, 1)
	d2dSetHand(opp, 5)
	start := opp.Life
	passPriorityAroundTable(t, g)
	if opp.Life != start-3 {
		t.Errorf("5 cards against 2: life %d -> %d, want -3", start, opp.Life)
	}

	advanceToUpkeepOf(t, g, 2)
	small := g.Seats[2]
	d2dSetHand(small, 1)
	start = small.Life
	passPriorityAroundTable(t, g)
	if small.Life != start {
		t.Errorf("a smaller hand lost %d life", start-small.Life)
	}
}

// --- Master of the Feast ----------------------------------------------

func TestMasterOfTheFeastMakesEachOpponentDraw(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[1]
	pushCatalogPermanent(g, me.ID, "Master of the Feast", "Enchantment Creature — Demon", masterOfTheFeastOracle, false)
	advanceToUpkeepOf(t, g, 1)
	before := make([]int, len(g.Seats))
	for i, s := range g.Seats {
		before[i] = s.Hand.Size()
	}
	passPriorityAroundTable(t, g)
	for i, s := range g.Seats {
		want := before[i] + 1
		if s.ID == me.ID {
			want = before[i]
		}
		if s.Hand.Size() != want {
			t.Errorf("seat %d hand %d, want %d", i, s.Hand.Size(), want)
		}
	}
}

// --- Indulgent Tormentor ----------------------------------------------

func TestIndulgentTormentorLetsYouDrawWhenTheyDecline(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[1], g.Seats[2]
	pushCatalogPermanent(g, me.ID, "Indulgent Tormentor", "Creature — Demon", indulgentTormentorOracle, false)
	b12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	advanceToUpkeepOf(t, g, 1)
	pickPlayer(t, g, me.ID, opp.ID)
	passPriorityAroundTable(t, g)
	prompt := latestOptionPickFor(g, opp.ID)
	if prompt == nil || len(prompt.PickOptions) != 3 {
		t.Fatalf("the opponent is offered draw / sacrifice / pay 3 life: %+v", prompt)
	}
	before := me.Hand.Size()
	answerOptionPick(t, g, opp.ID, 0)
	if me.Hand.Size() != before+1 {
		t.Errorf("hand %d -> %d, want +1", before, me.Hand.Size())
	}
}

func TestIndulgentTormentorPayingLifeStopsTheDraw(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[1], g.Seats[2]
	pushCatalogPermanent(g, me.ID, "Indulgent Tormentor", "Creature — Demon", indulgentTormentorOracle, false)
	advanceToUpkeepOf(t, g, 1)
	pickPlayer(t, g, me.ID, opp.ID)
	passPriorityAroundTable(t, g)
	prompt := latestOptionPickFor(g, opp.ID)
	if prompt == nil || len(prompt.PickOptions) != 2 {
		t.Fatalf("with no creature the sacrifice is not offered: %+v", prompt)
	}
	before, life := me.Hand.Size(), opp.Life
	answerOptionPick(t, g, opp.ID, 1)
	if opp.Life != life-3 || me.Hand.Size() != before {
		t.Errorf("life %d -> %d, hand %d -> %d; want -3 life and no draw", life, opp.Life, before, me.Hand.Size())
	}
}

func TestIndulgentTormentorSacrificeStopsTheDraw(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[1], g.Seats[2]
	pushCatalogPermanent(g, me.ID, "Indulgent Tormentor", "Creature — Demon", indulgentTormentorOracle, false)
	bear := b12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	advanceToUpkeepOf(t, g, 1)
	pickPlayer(t, g, me.ID, opp.ID)
	passPriorityAroundTable(t, g)
	before := me.Hand.Size()
	answerOptionPick(t, g, opp.ID, 1)
	if c := latestChooseCardsFor(g, opp.ID); c != nil {
		if err := g.ResolveChooseCards(c.ID, opp.ID, []uuid.UUID{bear}); err != nil {
			t.Fatalf("sacrifice: %v", err)
		}
	}
	if wpOnBattlefield(g, bear) {
		t.Error("the Bear was not sacrificed")
	}
	if me.Hand.Size() != before {
		t.Errorf("hand %d -> %d: a sacrifice should stop the draw", before, me.Hand.Size())
	}
}

// --- Entropic Specter -------------------------------------------------

func TestEntropicSpecterIsAsBigAsTheChosenPlayersHand(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, other := g.Seats[0], g.Seats[1], g.Seats[2]
	d2dSetHand(opp, 4)
	d2dSetHand(other, 9)
	id := castCatalogSpell(t, g, "Entropic Specter", "Creature — Specter Spirit", entropicSpecterOracle, nil)
	passPriorityAroundTable(t, g)
	answerChosenPlayer(t, g, me.ID, opp.ID)

	if p, tough := effectivePTOf(t, g, id); p != 4 || tough != 4 {
		t.Fatalf("P/T = %d/%d, want 4/4 (the chosen player's hand, not the other's)", p, tough)
	}
	// A hand change shows at once.
	g.WithWriteLock(func() {
		if err := g.DrawNForEffect(opp.ID, 2); err != nil {
			t.Fatalf("draw: %v", err)
		}
	})
	if p, _ := effectivePTOf(t, g, id); p != 6 {
		t.Errorf("after two draws power = %d, want 6", p)
	}
}

func TestEntropicSpecterDiscardsOnAnyDamageToAPlayer(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	d2dSetHand(opp, 3)
	id := castCatalogSpell(t, g, "Entropic Specter", "Creature — Specter Spirit", entropicSpecterOracle, nil)
	passPriorityAroundTable(t, g)
	answerChosenPlayer(t, g, me.ID, opp.ID)
	g.WithWriteLock(func() {
		if err := g.DealDamageToPlayerForEffect(id, opp.ID, 1); err != nil {
			t.Fatalf("damage: %v", err)
		}
	})
	passPriorityAroundTable(t, g)
	if discardOwed(g, opp.ID) != 1 {
		t.Fatalf("the damaged player owes a discard: %+v", g.PendingChoices)
	}
}

// --- Glasses of Urza --------------------------------------------------

func TestGlassesOfUrzaShowsTheHandToItsControllerOnly(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, other := g.Seats[0], g.Seats[1], g.Seats[2]
	glasses := pushCatalogPermanent(g, me.ID, "Glasses of Urza", "Artifact", glassesOfUrzaOracle, false)
	d2dSetHand(opp, 2)
	if err := g.ActivateCatalogAbility(me.ID, glasses, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	for i := range opp.Hand.Cards {
		if !opp.Hand.Cards[i].IsKnownTo(me.ID) {
			t.Errorf("card %d is not known to the controller", i)
		}
		if opp.Hand.Cards[i].IsKnownTo(other.ID) {
			t.Errorf("card %d leaked to another seat", i)
		}
	}
}

// --- Defile and Consuming Corruption -----------------------------------

func TestDefileShrinksByYourSwamps(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	for i := 0; i < 3; i++ {
		b31Push(g, me.ID, "Swamp", "Basic Land — Swamp", "", "", 0, 0)
	}
	b31Push(g, opp.ID, "Swamp", "Basic Land — Swamp", "", "", 0, 0)
	bear := b12Creature(g, opp.ID, "Ogre", "Creature — Ogre", 5, 5)
	castCatalogSpell(t, g, "Defile", "Instant", defileOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: bear}})
	passPriorityAroundTable(t, g)
	if p, tough := effectivePTOf(t, g, bear); p != 2 || tough != 2 {
		t.Errorf("P/T = %d/%d, want 2/2 (three of MY Swamps, not theirs)", p, tough)
	}
}

func TestConsumingCorruptionDamagesAndGainsBySwamps(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	for i := 0; i < 3; i++ {
		b31Push(g, me.ID, "Swamp", "Basic Land — Swamp", "", "", 0, 0)
	}
	ogre := b12Creature(g, opp.ID, "Ogre", "Creature — Ogre", 5, 5)
	start := me.Life
	castCatalogSpell(t, g, "Consuming Corruption", "Instant", consumingCorruptionOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: ogre}})
	passPriorityAroundTable(t, g)
	if me.Life != start+3 {
		t.Errorf("life %d -> %d, want +3", start, me.Life)
	}
	var dmg int
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.InstanceID == ogre {
				dmg = c.DamageMarked
			}
		}
	})
	if dmg != 3 {
		t.Errorf("the Ogre has %d damage, want 3", dmg)
	}
}

// --- Magus of the Coffers ----------------------------------------------

func TestMagusOfTheCoffersScalesWithSwamps(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	magus := seedPermanentWithOracle(g, me.ID, "Magus of the Coffers", "Creature — Human Wizard", magusOfTheCoffersOracle)
	for i := 0; i < 3; i++ {
		seedManaLand(g, me.ID, "Swamp", "Basic Land — Swamp", "B")
	}
	me.ManaPool.AddMana(game.ManaToken{Color: "C"}, game.ManaToken{Color: "C"})
	if err := g.ActivateManaAbility(me.ID, magus, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if got := sortedPool(me); len(got) != 3 || got[0] != "B" {
		t.Errorf("pool = %v, want three {B}", got)
	}
}

// --- Crypt Ghast and Nirkana Revenant ----------------------------------

func TestCryptGhastAndNirkanaRevenantDoubleOnlyYourSwamps(t *testing.T) {
	for _, tc := range []struct{ name, typeLine, oracle string }{
		{"Crypt Ghast", "Creature — Spirit", cryptGhastOracle},
		{"Nirkana Revenant", "Creature — Vampire Shade", nirkanaRevenantOracle},
	} {
		g := newCatalogGame(t)
		me, them := g.Seats[0], g.Seats[1]
		pushCatalogPermanent(g, me.ID, tc.name, tc.typeLine, tc.oracle, false)
		swamp := b31Push(g, me.ID, "Swamp", "Basic Land — Swamp", "", "", 0, 0)
		forest := b31Push(g, me.ID, "Forest", "Basic Land — Forest", "", "", 0, 0)
		theirs := b31Push(g, them.ID, "Swamp", "Basic Land — Swamp", "", "", 0, 0)

		tapForMana(t, g, me.ID, swamp)
		if got := sortedPool(me); len(got) != 2 || got[0] != "B" || got[1] != "B" {
			t.Errorf("%s: Swamp pool = %v, want {B}{B}", tc.name, got)
		}
		g.WithWriteLock(func() { me.ManaPool = nil })
		tapForMana(t, g, me.ID, forest)
		if got := sortedPool(me); len(got) != 1 || got[0] != "G" {
			t.Errorf("%s: Forest pool = %v, want just {G}", tc.name, got)
		}
		tapForMana(t, g, them.ID, theirs)
		if got := sortedPool(them); len(got) != 1 {
			t.Errorf("%s: an opponent's Swamp pool = %v, want one {B}", tc.name, got)
		}
	}
}

func TestCryptGhastExtortsOnACast(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushCatalogPermanent(g, me.ID, "Crypt Ghast", "Creature — Spirit", cryptGhastOracle, false)
	lives := b17Life(g)

	e2CastSpellOfColor(t, g, "Some Spell", "Sorcery", "U")
	for i := 0; i < 8 && !hasPayUnlessFor(g, me.ID); i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatal(err)
		}
	}
	if !hasPayUnlessFor(g, me.ID) {
		t.Fatal("casting a spell offers the extort payment")
	}
	me.ManaPool.AddMana(game.ManaToken{Color: "B"})
	answerPayUnless(t, g, me.ID, true)
	for i, p := range g.Seats {
		want := lives[i] - 1
		if i == 0 {
			want = lives[i] + len(g.Seats) - 1
		}
		if p.Life != want {
			t.Errorf("seat %d life %d, want %d", i, p.Life, want)
		}
	}
}

// --- Bog Witch ----------------------------------------------------------

func TestBogWitchPitchesACardForThreeBlack(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	witch := pushCatalogPermanent(g, me.ID, "Bog Witch", "Creature — Human Spellshaper", bogWitchOracle, false)
	pitch := ccHandCard(me, "Pitch Me", "Instant", "{3}{R}")
	me.ManaPool.AddMana(game.ManaToken{Color: "B"})
	if err := g.ActivateManaAbility(me.ID, witch, 0, game.ManaAbilityParams{
		DiscardIDs: []uuid.UUID{pitch},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if got := sortedPool(me); len(got) != 3 || got[0] != "B" {
		t.Errorf("pool = %v, want {B}{B}{B} (the {B} was spent)", got)
	}
	if me.Hand.Contains(pitch) || !me.Graveyard.Contains(pitch) {
		t.Error("the card was not discarded")
	}
}
