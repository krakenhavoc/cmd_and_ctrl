package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b088_cards_test.go — Commander Legends, batch 088 (#5507). Every
// test plays the real catalog entry through the engine.

const (
	cmrMolderBeastOracle       = "2b732c6a-f09c-4ac0-ae3a-1d8820b69aba"
	cmrDhundOperativeOracle    = "e025c802-9862-4dee-8c2a-b875eb001714"
	cmrVargusWrathOracle       = "e73d72f6-5572-47e0-90c9-dc1c146029d7"
	cmrGiltLeafWinnowerOracle  = "53b99846-3f80-4e9d-ad3c-52a4906b8823"
	cmrWildCelebrantsOracle    = "695661ac-6f50-4a0e-9831-a0e285bab4bf"
	cmrNoxiousDragonOracle     = "fe75b364-9e65-499d-9763-76844cf88cf7"
	cmrTrenchBehemothOracle    = "dc2b1863-b45c-4213-ad89-5e39432daca7"
	cmrIorRuinExpeditionOracle = "fc338ba2-c3a4-45d6-895f-af7f72bf7a94"
	cmrScrapdiverSerpentOracle = "6f45ac24-e153-4ec1-a41f-989f9547d409"
	cmrTransguildOracle        = "98334bfa-c516-4c20-bdc5-9e32e7127adc"
	cmrHorizonStoneOracle      = "add46292-470e-47fa-a261-4884d02e65fe"
	cmrSimicCharmOracle        = "c767aa2e-0a95-401f-8e36-d90a87648eab"
	cmrPlagueReaverOracle      = "eef50ef8-958b-4f5c-b261-16926d9fe672"
)

func TestCmr088EveryCardIsFull(t *testing.T) {
	for _, o := range []string{cmrMolderBeastOracle, cmrDhundOperativeOracle, cmrVargusWrathOracle, cmrGiltLeafWinnowerOracle,
		cmrWildCelebrantsOracle, cmrNoxiousDragonOracle, cmrTrenchBehemothOracle, cmrIorRuinExpeditionOracle,
		cmrScrapdiverSerpentOracle, cmrTransguildOracle, cmrHorizonStoneOracle, cmrSimicCharmOracle, cmrPlagueReaverOracle} {
		spec, ok := effects.Lookup(o)
		if !ok || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s not registered Full without caveats", o)
		}
	}
}

// Molder Beast: +2/+0 for each artifact put into a graveyard from the battlefield, anyone's.
func TestCmr088MolderBeastGrowsForEveryArtifactThatDies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	beast := effectstest.B12Push(g, me.ID, "Molder Beast", "Creature — Beast", cmrMolderBeastOracle, 5, 3)
	mine := effectstest.B12Permanent(g, me.ID, "My Rock", "Artifact")
	theirs := effectstest.B12Permanent(g, opp.ID, "Their Rock", "Artifact")
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() { _ = g.SacrificePermanentForEffect(bear) })
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	if p, _ := effectstest.EffectivePTOf(t, g, beast); p != 5 {
		t.Fatalf("a nonartifact creature dying changed the Beast: power %d", p)
	}
	for i, id := range []uuid.UUID{mine, theirs} {
		g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(id) })
		g.RunStateChecksForTest()
		effectstest.PassPriorityAroundTable(t, g)
		if p, to := effectstest.EffectivePTOf(t, g, beast); p != 5+2*(i+1) || to != 3 {
			t.Errorf("after %d artifacts the Beast is %d/%d, want %d/3", i+1, p, to, 5+2*(i+1))
		}
	}
}

func TestCmr088DhundOperativeNeedsAnArtifact(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	op := effectstest.B12Push(g, me.ID, "Dhund Operative", "Creature — Human Rogue", cmrDhundOperativeOracle, 2, 2)
	if p, _ := effectstest.EffectivePTOf(t, g, op); p != 2 || effectstest.HasEffectiveKeyword(t, g, op, "deathtouch") {
		t.Fatalf("without an artifact: power %d, deathtouch %v", p, effectstest.HasEffectiveKeyword(t, g, op, "deathtouch"))
	}
	rock := effectstest.B12Permanent(g, me.ID, "Rock", "Artifact")
	if p, _ := effectstest.EffectivePTOf(t, g, op); p != 3 || !effectstest.HasEffectiveKeyword(t, g, op, "deathtouch") {
		t.Errorf("with an artifact: power %d, deathtouch %v, want 3 and deathtouch", p, effectstest.HasEffectiveKeyword(t, g, op, "deathtouch"))
	}
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(rock) })
	g.RunStateChecksForTest()
	if p, _ := effectstest.EffectivePTOf(t, g, op); p != 2 || effectstest.HasEffectiveKeyword(t, g, op, "deathtouch") {
		t.Errorf("after the artifact left: power %d, deathtouch %v", p, effectstest.HasEffectiveKeyword(t, g, op, "deathtouch"))
	}
}

// Captain Vargus Wrath: your Pirates get +1/+1 for each time you've cast a commander
// from the command zone this game.
func TestCmr088CaptainVargusWrathPumpsPiratesPerCommanderCast(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	captain := effectstest.B12Push(g, me.ID, "Captain Vargus Wrath", "Legendary Creature — Orc Pirate", cmrVargusWrathOracle, 1, 1)
	pirate := effectstest.B12Creature(g, me.ID, "Deckhand", "Creature — Human Pirate", 2, 2)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirPirate := effectstest.B12Creature(g, opp.ID, "Their Pirate", "Creature — Human Pirate", 2, 2)
	g.WithWriteLock(func() { me.CommanderCasts = map[uuid.UUID]int{uuid.New(): 2, uuid.New(): 1} })
	effectstest.AttackWithTrigger(t, g, captain, opp.ID, uuid.Nil)
	if p, to := effectstest.EffectivePTOf(t, g, captain); p != 4 || to != 4 {
		t.Errorf("Vargus Wrath is %d/%d, want 4/4", p, to)
	}
	if p, to := effectstest.EffectivePTOf(t, g, pirate); p != 5 || to != 5 {
		t.Errorf("my other Pirate is %d/%d, want 5/5", p, to)
	}
	if p, _ := effectstest.EffectivePTOf(t, g, bear); p != 2 {
		t.Errorf("a non-Pirate got the pump: power %d", p)
	}
	if p, _ := effectstest.EffectivePTOf(t, g, theirPirate); p != 2 {
		t.Errorf("an opponent's Pirate got the pump: power %d", p)
	}
}

func TestCmr088GiltLeafWinnowerDestroysANonElfWithUnequalPowerAndToughness(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	square := effectstest.B12Creature(g, opp.ID, "Square", "Creature — Bear", 3, 3)
	lopsided := effectstest.B12Creature(g, opp.ID, "Lopsided", "Creature — Bear", 4, 1)
	elf := effectstest.B12Creature(g, opp.ID, "Elf", "Creature — Elf Warrior", 4, 1)
	effectstest.CastHoldCreature(t, g, "Gilt-Leaf Winnower", "Creature — Elf Warrior", cmrGiltLeafWinnowerOracle, 4, 3)
	p := effectstest.PickTargetPrompt(t, g)
	if effectstest.HasID(p.PickTargetCards, square) || effectstest.HasID(p.PickTargetCards, elf) {
		t.Errorf("illegal targets offered: %v", p.PickTargetCards)
	}
	if !effectstest.HasID(p.PickTargetCards, lopsided) {
		t.Fatalf("the 4/1 non-Elf was not offered: %v", p.PickTargetCards)
	}
	effectstest.AnswerPickTarget(t, g, lopsided)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(lopsided) {
		t.Error("the creature survived")
	}
}

func TestCmr088GiltLeafWinnowerMayDecline(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	lopsided := effectstest.B12Creature(g, opp.ID, "Lopsided", "Creature — Bear", 4, 1)
	effectstest.CastHoldCreature(t, g, "Gilt-Leaf Winnower", "Creature — Elf Warrior", cmrGiltLeafWinnowerOracle, 4, 3)
	effectstest.AnswerPickTarget(t, g, lopsided)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(lopsided) {
		t.Error("the creature was destroyed although the controller declined")
	}
	if !effectstest.HasEffectiveKeyword(t, g, effectstest.FindBattlefieldByName(g, "Gilt-Leaf Winnower"), "menace") {
		t.Error("no menace")
	}
}

func TestCmr088WildCelebrantsMayDestroyAnArtifact(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	rock := effectstest.B12Permanent(g, opp.ID, "Their Rock", "Artifact")
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.CastHoldCreature(t, g, "Wild Celebrants", "Creature — Satyr", cmrWildCelebrantsOracle, 5, 3)
	p := effectstest.PickTargetPrompt(t, g)
	if effectstest.HasID(p.PickTargetCards, bear) {
		t.Error("a nonartifact creature was offered")
	}
	effectstest.AnswerPickTarget(t, g, rock)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(rock) {
		t.Error("the artifact survived")
	}
}

func TestCmr088NoxiousDragonDestroysACreatureWithManaValueThreeOrLessWhenItDies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	cheap := effectstest.B12Push(g, opp.ID, "Cheap", "Creature — Bear", "", 2, 2)
	dear := effectstest.B12Push(g, opp.ID, "Dear", "Creature — Bear", "", 2, 2)
	g.WithWriteLock(func() {
		effectstest.FindBattlefieldCardForTest(g, cheap).ManaCost = "{2}{G}"
		effectstest.FindBattlefieldCardForTest(g, dear).ManaCost = "{3}{G}"
	})
	dragon := effectstest.B12Push(g, me.ID, "Noxious Dragon", "Creature — Dragon", cmrNoxiousDragonOracle, 4, 4)
	effectstest.B27Kill(g, dragon)
	g.RunStateChecksForTest()
	effectstest.B04WaitForPick(t, g, me.ID)
	p := effectstest.PickTargetPrompt(t, g)
	if effectstest.HasID(p.PickTargetCards, dear) || !effectstest.HasID(p.PickTargetCards, cheap) {
		t.Fatalf("targets offered %v, want only the creature with mana value 3 or less", p.PickTargetCards)
	}
	effectstest.AnswerPickTarget(t, g, cheap)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(cheap) {
		t.Error("the cheap creature survived")
	}
}

// Trench Behemoth: return a land to untap it and give it hexproof.
func TestCmr088TrenchBehemothReturnsALandToUntapAndGainHexproof(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	behemoth := effectstest.PushCatalogPermanent(g, me.ID, "Trench Behemoth", "Creature — Kraken", cmrTrenchBehemothOracle, false)
	land := effectstest.SeedLand(g, me.ID, "Island", "Basic Land — Island", "")
	g.WithWriteLock(func() { effectstest.FindBattlefieldCardForTest(g, behemoth).Tapped = true })
	effectstest.AdvanceToMain(t, g)
	if err := g.ActivateCatalogAbility(me.ID, behemoth, 0, game.ActivateAbilityParams{Strict: true, ReturnIDs: []uuid.UUID{behemoth}}); err == nil {
		t.Fatal("returned a nonland permanent as the cost")
	}
	if err := g.ActivateCatalogAbility(me.ID, behemoth, 0, game.ActivateAbilityParams{Strict: true, ReturnIDs: []uuid.UUID{land}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(land) || !me.Hand.Contains(land) {
		t.Error("the land was not returned to its owner's hand")
	}
	if effectstest.TappedForTest(t, g, behemoth) {
		t.Error("the Behemoth stayed tapped")
	}
	if !effectstest.HasEffectiveKeyword(t, g, behemoth, "hexproof") {
		t.Error("the Behemoth has no hexproof")
	}
}

// Landfall: target creature an opponent controls must attack in its controller's next combat.
func TestCmr088TrenchBehemothLandfallForcesAnOpposingCreatureToAttack(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Trench Behemoth", "Creature — Kraken", cmrTrenchBehemothOracle, false)
	mine := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	effectstest.CastCatalogSpell(t, g, "Forest", "Basic Land — Forest", "", nil)
	p := effectstest.PickTargetPrompt(t, g)
	if effectstest.HasID(p.PickTargetCards, mine) || !effectstest.HasID(p.PickTargetCards, theirs) {
		t.Fatalf("targets offered %v, want only an opponent's creature", p.PickTargetCards)
	}
	effectstest.AnswerPickTarget(t, g, theirs)
	effectstest.PassPriorityAroundTable(t, g)
	required := false
	g.ReadSnapshot(func() {
		c := effectstest.BattlefieldCardFor(g, theirs)
		required = len(c.Effective().AttackRequirements) > 0
	})
	if !required {
		t.Error("the creature has no attack requirement")
	}
}

func TestCmr088IorRuinExpeditionCountsLandsAndDrawsTwo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	ruin := effectstest.PushCatalogPermanent(g, me.ID, "Ior Ruin Expedition", "Enchantment", cmrIorRuinExpeditionOracle, false)
	effectstest.AdvanceToMain(t, g)
	for i := 0; i < 3; i++ {
		effectstest.PlayLandFromHand(t, g, "Forest", "")
		effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, i != 1) // decline the second
		effectstest.PassPriorityAroundTable(t, g)
	}
	if got := effectstest.CounterCount(g, ruin, "quest"); got != 2 {
		t.Fatalf("%d quest counters, want 2 (one landfall was declined)", got)
	}
	if err := g.ActivateCatalogAbility(me.ID, ruin, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("activated with only two quest counters")
	}
	effectstest.PlayLandFromHand(t, g, "Forest", "")
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	hand := me.Hand.Size()
	if err := g.ActivateCatalogAbility(me.ID, ruin, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(ruin) {
		t.Error("the enchantment was not sacrificed")
	}
	if me.Hand.Size() != hand+2 {
		t.Errorf("hand %d, want %d", me.Hand.Size(), hand+2)
	}
}

func TestCmr088ScrapdiverSerpentIsUnblockableWhileTheDefenderControlsAnArtifact(t *testing.T) {
	for _, artifact := range []bool{false, true} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		serpent := effectstest.B12Push(g, me.ID, "Scrapdiver Serpent", "Creature — Serpent", cmrScrapdiverSerpentOracle, 5, 5)
		wall := effectstest.B12Creature(g, opp.ID, "Their Wall", "Creature — Wall", 0, 4)
		if artifact {
			effectstest.B12Permanent(g, opp.ID, "Their Rock", "Artifact")
		} else {
			effectstest.B12Permanent(g, me.ID, "My Rock", "Artifact") // only the defender's artifacts count
		}
		effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
		if err := g.DeclareAttacker(serpent, opp.ID); err != nil {
			t.Fatalf("DeclareAttacker: %v", err)
		}
		effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
		err := g.DeclareBlocker(wall, serpent)
		if artifact && err == nil {
			t.Error("the Serpent was blocked although the defender controls an artifact")
		}
		if !artifact && err != nil {
			t.Errorf("the Serpent could not be blocked without a defending artifact: %v", err)
		}
	}
}

func TestCmr088TransguildPromenadeEntersTappedAndIsPaidForOrSacrificed(t *testing.T) {
	for _, pay := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		effectstest.AdvanceToMain(t, g)
		land := effectstest.PlayLandFromHand(t, g, "Transguild Promenade", cmrTransguildOracle)
		effectstest.PassPriorityAroundTable(t, g)
		if c, ok := effectstest.BattlefieldCard(g, land); !ok || !c.Tapped {
			t.Fatal("the land did not enter tapped")
		}
		if pay {
			effectstest.DcMana(t, g, me.ID, "{C}")
		}
		effectstest.AnswerPayUnless(t, g, me.ID, pay)
		effectstest.PassPriorityAroundTable(t, g)
		if g.Battlefield.Contains(land) != pay {
			t.Errorf("paid=%v: the land on the battlefield = %v", pay, g.Battlefield.Contains(land))
		}
	}
}

// Horizon Stone: unspent mana becomes colorless instead of being lost.
func TestCmr088HorizonStoneTurnsUnspentManaColorless(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushCatalogPermanent(g, me.ID, "Horizon Stone", "Artifact", cmrHorizonStoneOracle, false)
	effectstest.AdvanceToMain(t, g)
	effectstest.FillPoolColored(me, "R", 2)
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	got := effectstest.PoolColors(me)
	if len(got) != 2 {
		t.Fatalf("pool after the phase %v, want two mana kept", got)
	}
	for _, c := range got {
		if c != "C" {
			t.Errorf("kept mana is %s, want colorless", c)
		}
	}
}

func TestCmr088SimicCharmModes(t *testing.T) {
	t.Run("pump", func(t *testing.T) {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
		effectstest.AdvanceToMain(t, g)
		if err := g.CastSpell(me.ID, cmr088Hand(me, "Simic Charm", "Instant", cmrSimicCharmOracle), game.CastSpellParams{Modes: []int{0},
			Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear, Mode: 0}}}); err != nil {
			t.Fatalf("CastSpell: %v", err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		if p, to := effectstest.EffectivePTOf(t, g, bear); p != 5 || to != 5 {
			t.Errorf("the creature is %d/%d, want 5/5", p, to)
		}
	})
	t.Run("hexproof", func(t *testing.T) {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
		rock := effectstest.B12Permanent(g, me.ID, "Rock", "Artifact")
		effectstest.AdvanceToMain(t, g)
		if err := g.CastSpell(me.ID, cmr088Hand(me, "Simic Charm", "Instant", cmrSimicCharmOracle), game.CastSpellParams{Modes: []int{1}}); err != nil {
			t.Fatalf("CastSpell: %v", err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		for _, id := range []uuid.UUID{bear, rock} {
			if !effectstest.HasEffectiveKeyword(t, g, id, "hexproof") {
				t.Errorf("%s lacks hexproof", id)
			}
		}
	})
	t.Run("bounce", func(t *testing.T) {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
		theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
		effectstest.AdvanceToMain(t, g)
		if err := g.CastSpell(me.ID, cmr088Hand(me, "Simic Charm", "Instant", cmrSimicCharmOracle), game.CastSpellParams{Modes: []int{2},
			Targets: []game.TargetRef{{Kind: game.TargetCard, ID: theirs, Mode: 0}}}); err != nil {
			t.Fatalf("CastSpell: %v", err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		if g.Battlefield.Contains(theirs) || !opp.Hand.Contains(theirs) {
			t.Error("the creature was not returned to its owner's hand")
		}
	})
}

// cmr088Hand puts a catalog card into the player's hand and returns its id.
func cmr088Hand(p *game.Player, name, typeLine, oracle string) uuid.UUID {
	id := uuid.New()
	p.Hand.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, OracleID: oracle, Owner: p.ID, Controller: p.ID})
	return id
}

// Plague Reaver: sacrifice each other creature you control at your end step.
func TestCmr088PlagueReaverSacrificesTheOtherCreaturesAtYourEndStep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	reaver := effectstest.B12Push(g, me.ID, "Plague Reaver", "Creature — Beast", cmrPlagueReaverOracle, 6, 5)
	mine := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(mine) {
		t.Error("my other creature survived the end step")
	}
	if !g.Battlefield.Contains(reaver) || !g.Battlefield.Contains(theirs) {
		t.Error("the Reaver or an opponent's creature was sacrificed")
	}
}

// Plague Reaver: discard two and sacrifice it to give it to an opponent at their next upkeep.
func TestCmr088PlagueReaverComesBackUnderAnOpponentsControlAtTheirNextUpkeep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	reaver := effectstest.B12Push(g, me.ID, "Plague Reaver", "Creature — Beast", cmrPlagueReaverOracle, 6, 5)
	effectstest.D2dSetHand(me, 3)
	effectstest.AdvanceToMain(t, g)
	discards := []uuid.UUID{me.Hand.Cards[0].InstanceID, me.Hand.Cards[1].InstanceID}
	if err := g.ActivateCatalogAbility(me.ID, reaver, 0, game.ActivateAbilityParams{Strict: true,
		DiscardIDs: discards[:1], Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}}); err == nil {
		t.Fatal("activated with one discard")
	}
	if err := g.ActivateCatalogAbility(me.ID, reaver, 0, game.ActivateAbilityParams{Strict: true,
		DiscardIDs: discards, Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(reaver) || !me.Graveyard.Contains(reaver) {
		t.Fatal("the Reaver was not sacrificed")
	}
	effectstest.AdvanceToUpkeepOf(t, g, 1)
	effectstest.PassPriorityAroundTable(t, g)
	c, ok := effectstest.BattlefieldCard(g, reaver)
	if !ok {
		t.Fatal("the Reaver did not return at the opponent's upkeep")
	}
	if c.Controller != opp.ID {
		t.Errorf("the Reaver is controlled by %s, want the chosen opponent", c.Controller)
	}
}
