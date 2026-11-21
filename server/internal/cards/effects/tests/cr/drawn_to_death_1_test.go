package testscr

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// --- Iron Maiden and Viseling ----------------------------------------

func TestIronMaidenAndViselingDamageAnOpponentsOversizedHand(t *testing.T) {
	for _, tc := range []struct {
		name, oracle, typeLine string
	}{
		{"Iron Maiden", effectstest.IronMaidenOracle, "Artifact"},
		{"Viseling", effectstest.ViselingOracle, "Artifact Creature — Phyrexian Construct"},
	} {
		for _, hand := range []int{3, 4, 7} {
			g := effectstest.NewCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			effectstest.PushPermanentForTest(g, me.ID, tc.name, tc.oracle, tc.typeLine)
			effectstest.AdvanceToUpkeepOf(t, g, 1)
			effectstest.D2dSetHand(opp, hand)
			start := opp.Life
			effectstest.PassPriorityAroundTable(t, g)
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
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[1]
	effectstest.PushPermanentForTest(g, me.ID, "Iron Maiden", effectstest.IronMaidenOracle, "Artifact")
	effectstest.AdvanceToUpkeepOf(t, g, 1)
	effectstest.D2dSetHand(me, 9)
	start := me.Life
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != start {
		t.Errorf("the controller's own upkeep dealt %d", start-me.Life)
	}
}

// --- The Rack ---------------------------------------------------------

func TestTheRackPunishesOnlyTheChosenOpponent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, chosen, other := g.Seats[0], g.Seats[1], g.Seats[2]
	effectstest.CastCatalogSpell(t, g, "The Rack", "Artifact", effectstest.TheRackOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	prompt := effectstest.PendingOfKind(g, game.PendingChoiceOptionPick)
	if prompt == nil {
		t.Fatalf("no as-enters prompt: %+v", g.PendingChoices)
	}
	for _, opt := range prompt.PickOptions {
		if opt.Player == me.ID {
			t.Error("choose an OPPONENT: the controller was offered")
		}
	}
	effectstest.AnswerChosenPlayer(t, g, me.ID, chosen.ID)

	effectstest.AdvanceToUpkeepOf(t, g, 2)
	effectstest.D2dSetHand(other, 0)
	start := other.Life
	effectstest.PassPriorityAroundTable(t, g)
	if other.Life != start {
		t.Errorf("an unchosen opponent took %d", start-other.Life)
	}

	effectstest.AdvanceToUpkeepOf(t, g, 1)
	effectstest.D2dSetHand(chosen, 1)
	start = chosen.Life
	effectstest.PassPriorityAroundTable(t, g)
	if chosen.Life != start-2 {
		t.Errorf("hand of 1: life %d -> %d, want -2 (3 minus 1)", start, chosen.Life)
	}
}

// --- Dark Suspicions --------------------------------------------------

func TestDarkSuspicionsComparesTheTwoHands(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushPermanentForTest(g, me.ID, "Dark Suspicions", effectstest.DarkSuspicionsOracle, "Enchantment")
	effectstest.D2dSetHand(me, 2)
	effectstest.AdvanceToUpkeepOf(t, g, 1)
	effectstest.D2dSetHand(opp, 5)
	start := opp.Life
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != start-3 {
		t.Errorf("5 cards against 2: life %d -> %d, want -3", start, opp.Life)
	}

	effectstest.AdvanceToUpkeepOf(t, g, 2)
	small := g.Seats[2]
	effectstest.D2dSetHand(small, 1)
	start = small.Life
	effectstest.PassPriorityAroundTable(t, g)
	if small.Life != start {
		t.Errorf("a smaller hand lost %d life", start-small.Life)
	}
}

// --- Master of the Feast ----------------------------------------------

func TestMasterOfTheFeastMakesEachOpponentDraw(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Master of the Feast", "Enchantment Creature — Demon", effectstest.MasterOfTheFeastOracle, false)
	effectstest.AdvanceToUpkeepOf(t, g, 1)
	before := make([]int, len(g.Seats))
	for i, s := range g.Seats {
		before[i] = s.Hand.Size()
	}
	effectstest.PassPriorityAroundTable(t, g)
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
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[1], g.Seats[2]
	effectstest.PushCatalogPermanent(g, me.ID, "Indulgent Tormentor", "Creature — Demon", effectstest.IndulgentTormentorOracle, false)
	effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceToUpkeepOf(t, g, 1)
	effectstest.PickPlayer(t, g, me.ID, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	prompt := effectstest.LatestOptionPickFor(g, opp.ID)
	if prompt == nil || len(prompt.PickOptions) != 3 {
		t.Fatalf("the opponent is offered draw / sacrifice / pay 3 life: %+v", prompt)
	}
	before := me.Hand.Size()
	effectstest.AnswerOptionPick(t, g, opp.ID, 0)
	if me.Hand.Size() != before+1 {
		t.Errorf("hand %d -> %d, want +1", before, me.Hand.Size())
	}
}

func TestIndulgentTormentorPayingLifeStopsTheDraw(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[1], g.Seats[2]
	effectstest.PushCatalogPermanent(g, me.ID, "Indulgent Tormentor", "Creature — Demon", effectstest.IndulgentTormentorOracle, false)
	effectstest.AdvanceToUpkeepOf(t, g, 1)
	effectstest.PickPlayer(t, g, me.ID, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	prompt := effectstest.LatestOptionPickFor(g, opp.ID)
	if prompt == nil || len(prompt.PickOptions) != 2 {
		t.Fatalf("with no creature the sacrifice is not offered: %+v", prompt)
	}
	before, life := me.Hand.Size(), opp.Life
	effectstest.AnswerOptionPick(t, g, opp.ID, 1)
	if opp.Life != life-3 || me.Hand.Size() != before {
		t.Errorf("life %d -> %d, hand %d -> %d; want -3 life and no draw", life, opp.Life, before, me.Hand.Size())
	}
}

func TestIndulgentTormentorSacrificeStopsTheDraw(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[1], g.Seats[2]
	effectstest.PushCatalogPermanent(g, me.ID, "Indulgent Tormentor", "Creature — Demon", effectstest.IndulgentTormentorOracle, false)
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceToUpkeepOf(t, g, 1)
	effectstest.PickPlayer(t, g, me.ID, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	before := me.Hand.Size()
	effectstest.AnswerOptionPick(t, g, opp.ID, 1)
	if c := effectstest.LatestChooseCardsFor(g, opp.ID); c != nil {
		if err := g.ResolveChooseCards(c.ID, opp.ID, []uuid.UUID{bear}); err != nil {
			t.Fatalf("sacrifice: %v", err)
		}
	}
	if effectstest.WpOnBattlefield(g, bear) {
		t.Error("the Bear was not sacrificed")
	}
	if me.Hand.Size() != before {
		t.Errorf("hand %d -> %d: a sacrifice should stop the draw", before, me.Hand.Size())
	}
}

// --- Entropic Specter -------------------------------------------------

func TestEntropicSpecterIsAsBigAsTheChosenPlayersHand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, other := g.Seats[0], g.Seats[1], g.Seats[2]
	effectstest.D2dSetHand(opp, 4)
	effectstest.D2dSetHand(other, 9)
	id := effectstest.CastCatalogSpell(t, g, "Entropic Specter", "Creature — Specter Spirit", effectstest.EntropicSpecterOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerChosenPlayer(t, g, me.ID, opp.ID)

	if p, tough := effectstest.EffectivePTOf(t, g, id); p != 4 || tough != 4 {
		t.Fatalf("P/T = %d/%d, want 4/4 (the chosen player's hand, not the other's)", p, tough)
	}
	// A hand change shows at once.
	g.WithWriteLock(func() {
		if err := g.DrawNForEffect(opp.ID, 2); err != nil {
			t.Fatalf("draw: %v", err)
		}
	})
	if p, _ := effectstest.EffectivePTOf(t, g, id); p != 6 {
		t.Errorf("after two draws power = %d, want 6", p)
	}
}

func TestEntropicSpecterDiscardsOnAnyDamageToAPlayer(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.D2dSetHand(opp, 3)
	id := effectstest.CastCatalogSpell(t, g, "Entropic Specter", "Creature — Specter Spirit", effectstest.EntropicSpecterOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerChosenPlayer(t, g, me.ID, opp.ID)
	g.WithWriteLock(func() {
		if err := g.DealDamageToPlayerForEffect(id, opp.ID, 1); err != nil {
			t.Fatalf("damage: %v", err)
		}
	})
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.DiscardOwed(g, opp.ID) != 1 {
		t.Fatalf("the damaged player owes a discard: %+v", g.PendingChoices)
	}
}

// --- Glasses of Urza --------------------------------------------------

func TestGlassesOfUrzaShowsTheHandToItsControllerOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, other := g.Seats[0], g.Seats[1], g.Seats[2]
	glasses := effectstest.PushCatalogPermanent(g, me.ID, "Glasses of Urza", "Artifact", effectstest.GlassesOfUrzaOracle, false)
	effectstest.D2dSetHand(opp, 2)
	if err := g.ActivateCatalogAbility(me.ID, glasses, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
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
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	for i := 0; i < 3; i++ {
		effectstest.B31Push(g, me.ID, "Swamp", "Basic Land — Swamp", "", "", 0, 0)
	}
	effectstest.B31Push(g, opp.ID, "Swamp", "Basic Land — Swamp", "", "", 0, 0)
	bear := effectstest.B12Creature(g, opp.ID, "Ogre", "Creature — Ogre", 5, 5)
	effectstest.CastCatalogSpell(t, g, "Defile", "Instant", effectstest.DefileOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: bear}})
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.EffectivePTOf(t, g, bear); p != 2 || tough != 2 {
		t.Errorf("P/T = %d/%d, want 2/2 (three of MY Swamps, not theirs)", p, tough)
	}
}

func TestConsumingCorruptionDamagesAndGainsBySwamps(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	for i := 0; i < 3; i++ {
		effectstest.B31Push(g, me.ID, "Swamp", "Basic Land — Swamp", "", "", 0, 0)
	}
	ogre := effectstest.B12Creature(g, opp.ID, "Ogre", "Creature — Ogre", 5, 5)
	start := me.Life
	effectstest.CastCatalogSpell(t, g, "Consuming Corruption", "Instant", effectstest.ConsumingCorruptionOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: ogre}})
	effectstest.PassPriorityAroundTable(t, g)
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
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	magus := effectstest.SeedPermanentWithOracle(g, me.ID, "Magus of the Coffers", "Creature — Human Wizard", effectstest.MagusOfTheCoffersOracle)
	for i := 0; i < 3; i++ {
		effectstest.SeedManaLand(g, me.ID, "Swamp", "Basic Land — Swamp", "B")
	}
	me.ManaPool.AddMana(game.ManaToken{Color: "C"}, game.ManaToken{Color: "C"})
	if err := g.ActivateManaAbility(me.ID, magus, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if got := effectstest.SortedPool(me); len(got) != 3 || got[0] != "B" {
		t.Errorf("pool = %v, want three {B}", got)
	}
}

// --- Crypt Ghast and Nirkana Revenant ----------------------------------

func TestCryptGhastAndNirkanaRevenantDoubleOnlyYourSwamps(t *testing.T) {
	for _, tc := range []struct{ name, typeLine, oracle string }{
		{"Crypt Ghast", "Creature — Spirit", effectstest.CryptGhastOracle},
		{"Nirkana Revenant", "Creature — Vampire Shade", effectstest.NirkanaRevenantOracle},
	} {
		g := effectstest.NewCatalogGame(t)
		me, them := g.Seats[0], g.Seats[1]
		effectstest.PushCatalogPermanent(g, me.ID, tc.name, tc.typeLine, tc.oracle, false)
		swamp := effectstest.B31Push(g, me.ID, "Swamp", "Basic Land — Swamp", "", "", 0, 0)
		forest := effectstest.B31Push(g, me.ID, "Forest", "Basic Land — Forest", "", "", 0, 0)
		theirs := effectstest.B31Push(g, them.ID, "Swamp", "Basic Land — Swamp", "", "", 0, 0)

		effectstest.TapForMana(t, g, me.ID, swamp)
		if got := effectstest.SortedPool(me); len(got) != 2 || got[0] != "B" || got[1] != "B" {
			t.Errorf("%s: Swamp pool = %v, want {B}{B}", tc.name, got)
		}
		g.WithWriteLock(func() { me.ManaPool = nil })
		effectstest.TapForMana(t, g, me.ID, forest)
		if got := effectstest.SortedPool(me); len(got) != 1 || got[0] != "G" {
			t.Errorf("%s: Forest pool = %v, want just {G}", tc.name, got)
		}
		effectstest.TapForMana(t, g, them.ID, theirs)
		if got := effectstest.SortedPool(them); len(got) != 1 {
			t.Errorf("%s: an opponent's Swamp pool = %v, want one {B}", tc.name, got)
		}
	}
}

func TestCryptGhastExtortsOnACast(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushCatalogPermanent(g, me.ID, "Crypt Ghast", "Creature — Spirit", effectstest.CryptGhastOracle, false)
	lives := effectstest.B17Life(g)

	effectstest.E2CastSpellOfColor(t, g, "Some Spell", "Sorcery", "U")
	for i := 0; i < 8 && !effectstest.HasPayUnlessFor(g, me.ID); i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatal(err)
		}
	}
	if !effectstest.HasPayUnlessFor(g, me.ID) {
		t.Fatal("casting a spell offers the extort payment")
	}
	me.ManaPool.AddMana(game.ManaToken{Color: "B"})
	effectstest.AnswerPayUnless(t, g, me.ID, true)
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
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	witch := effectstest.PushCatalogPermanent(g, me.ID, "Bog Witch", "Creature — Human Spellshaper", effectstest.BogWitchOracle, false)
	pitch := effectstest.CcHandCard(me, "Pitch Me", "Instant", "{3}{R}")
	me.ManaPool.AddMana(game.ManaToken{Color: "B"})
	if err := g.ActivateManaAbility(me.ID, witch, 0, game.ManaAbilityParams{
		DiscardIDs: []uuid.UUID{pitch},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if got := effectstest.SortedPool(me); len(got) != 3 || got[0] != "B" {
		t.Errorf("pool = %v, want {B}{B}{B} (the {B} was spent)", got)
	}
	if me.Hand.Contains(pitch) || !me.Graveyard.Contains(pitch) {
		t.Error("the card was not discarded")
	}
}
