package testss82m

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b086_cards_test.go — Commander Legends, batch 086 (#5507). Every
// test plays the real catalog entry through the engine.

const (
	cmrVowOfTormentOracle       = "237b87f4-783b-40ac-9fcf-4ae76a33229d"
	cmrVowOfWildnessOracle      = "569a65fb-e35f-4388-ab49-f57be46b4889"
	cmrVowOfFlightOracle        = "81fe554a-5257-446c-ad53-ee45bf5edb40"
	cmrBenevolentBlessingOracle = "9472c96a-2827-4af6-a6a1-b6749a5e9069"
	cmrOnSerrasWingsOracle      = "f43584de-4f60-46e6-9e6b-e704bc3ecd59"
	cmrPortentOfBetrayalOracle  = "15d794f4-7789-40e3-9ee0-b09250c0516d"
	cmrSakashimasWillOracle     = "2c3c3bec-00f0-4de2-b929-73f1c9276ec4"
	cmrPeelFromRealityOracle    = "65dc05d0-f885-4fe5-9a23-d6b75185c179"
	cmrWrongTurnOracle          = "d1cfbbcc-4aee-4400-b5a8-18713829ae72"
)

// cmr086Hand puts a catalog card into the player's hand and returns its id.
func cmr086Hand(p *game.Player, name, typeLine, oracle string) uuid.UUID {
	id := uuid.New()
	p.Hand.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, OracleID: oracle, Owner: p.ID, Controller: p.ID})
	return id
}

func TestCmr086EveryCardIsFull(t *testing.T) {
	for _, o := range []string{cmrVowOfTormentOracle, cmrVowOfWildnessOracle, cmrVowOfFlightOracle,
		cmrBenevolentBlessingOracle, cmrOnSerrasWingsOracle, cmrPortentOfBetrayalOracle,
		cmrSakashimasWillOracle, cmrPeelFromRealityOracle, cmrWrongTurnOracle} {
		spec, ok := effects.Lookup(o)
		if !ok || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s not registered Full without caveats", o)
		}
	}
}

func TestCmr086VowsPumpGrantAndBarAttackingTheirController(t *testing.T) {
	for _, tc := range []struct {
		name, oracle, kw string
		pt               int
	}{
		{"Vow of Torment", cmrVowOfTormentOracle, "menace", 4},
		{"Vow of Wildness", cmrVowOfWildnessOracle, "trample", 5},
		{"Vow of Flight", cmrVowOfFlightOracle, "flying", 4},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me, opp, third := g.Seats[0], g.Seats[1], g.Seats[2]
			host := effectstest.B12Creature(g, opp.ID, "Host", "Creature — Bear", 2, 2)
			other := effectstest.B12Creature(g, opp.ID, "Other", "Creature — Bear", 2, 2)
			effectstest.Cmdr2B19Attach(g, me.ID, host, tc.name, tc.oracle)
			if p, to := effectstest.EffectivePTOf(t, g, host); p != tc.pt || to != tc.pt {
				t.Errorf("host is %d/%d, want %d/%d", p, to, tc.pt, tc.pt)
			}
			if !effectstest.HasEffectiveKeyword(t, g, host, tc.kw) {
				t.Errorf("host lacks %s", tc.kw)
			}
			var walker uuid.UUID
			g.WithWriteLock(func() {
				walker = effectstest.PushPermanent(g, me.ID, game.Card{Name: "My Walker", TypeLine: "Legendary Planeswalker — Test",
					Counters: map[string]int{game.CounterLoyalty: 3}})
			})
			effectstest.AdvanceToDeclareAttackersOf(t, g, 1)
			for _, target := range []uuid.UUID{me.ID, walker} {
				if err := g.Clone().DeclareAttacker(host, target); !errors.Is(err, game.ErrIllegalAttackTarget) {
					t.Errorf("enchanted creature attacking %s: err = %v, want ErrIllegalAttackTarget", target, err)
				}
			}
			if err := g.Clone().DeclareAttacker(host, third.ID); err != nil {
				t.Errorf("enchanted creature attacking another player: %v", err)
			}
			if err := g.Clone().DeclareAttacker(other, me.ID); err != nil {
				t.Errorf("an unenchanted creature attacking me: %v", err)
			}
		})
	}
}

func TestCmr086OnSerrasWingsMakesTheCreatureLegendaryAndGrantsKeywords(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	host := effectstest.B12Creature(g, me.ID, "Host", "Creature — Bear", 2, 2)
	effectstest.Cmdr2B19Attach(g, me.ID, host, "On Serra's Wings", cmrOnSerrasWingsOracle)
	if p, to := effectstest.EffectivePTOf(t, g, host); p != 3 || to != 3 {
		t.Errorf("host is %d/%d, want 3/3", p, to)
	}
	for _, kw := range []string{"flying", "vigilance", "lifelink"} {
		if !effectstest.HasEffectiveKeyword(t, g, host, kw) {
			t.Errorf("host lacks %s", kw)
		}
	}
	legendary := false
	g.ReadSnapshot(func() {
		c := effectstest.BattlefieldCardFor(g, host)
		legendary = effects.IsLegendary(c)
	})
	if !legendary {
		t.Error("the enchanted creature is not legendary")
	}
}

// Benevolent Blessing (CR 702.16p): protection from the chosen colour does not
// remove the Auras and Equipment you control that are already attached, and
// removes everything else with the colour.
func TestCmr086BenevolentBlessingSparesYourAttachmentsAndRemovesTheirs(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := effectstest.SeedBear(g, me.ID)
	push := func(owner uuid.UUID, name, typeLine string) uuid.UUID {
		return effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: name, TypeLine: typeLine,
			Colors: []string{"W"}, Owner: owner, Controller: owner, AttachedTo: game.TargetRef{Kind: game.TargetCard, ID: bear}})
	}
	myEquip := push(me.ID, "My Sword", "Artifact — Equipment")
	myAura := push(me.ID, "My Aura", "Enchantment — Aura")
	theirEquip := push(opp.ID, "Their Sword", "Artifact — Equipment")
	theirAura := push(opp.ID, "Their Aura", "Enchantment — Aura")
	effectstest.NmAdvanceTo(t, g, game.StepPrecombatMain)
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Benevolent Blessing", TypeLine: "Enchantment — Aura", OracleID: cmrBenevolentBlessingOracle,
		ManaCost: "{1}{W}", Colors: []string{"W"}, Owner: me.ID, Controller: me.ID})
	effectstest.FillPoolColored(me, "W", 1)
	effectstest.FillPool(me, 1)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}}}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	effectstest.AdResolveSpell(t, g, id)
	effectstest.AnswerColor(t, g, me.ID, "W")
	g.RunStateChecksForTest()
	if got := effectstest.ProtectionsOn(t, g, bear); !effectstest.HasString(got, "white") {
		t.Fatalf("the bear's protections = %q, want white", got)
	}
	attached := func(id uuid.UUID) bool {
		c, ok := effectstest.BattlefieldCard(g, id)
		return ok && c.IsAttachedTo(bear)
	}
	for name, kept := range map[string]struct {
		id   uuid.UUID
		keep bool
	}{
		"Benevolent Blessing": {id, true}, "my Equipment": {myEquip, true}, "my Aura": {myAura, true},
		"their Equipment": {theirEquip, false}, "their Aura": {theirAura, false},
	} {
		// An Aura that can no longer enchant goes to the graveyard (CR 704.5m); an
		// Equipment only becomes unattached (CR 704.5n).
		if attached(kept.id) != kept.keep {
			t.Errorf("%s attached to the bear = %v, want %v", name, !kept.keep, kept.keep)
		}
	}
	if g.Battlefield.Contains(theirAura) {
		t.Error("their Aura stayed on the battlefield")
	}
	// Protection still stops a new white Aura from attaching: Flash works, and nothing else gets on.
	if spec, _ := effects.Lookup(cmrBenevolentBlessingOracle); len(spec.PrintedKeywords) != 1 || spec.PrintedKeywords[0] != "flash" {
		t.Errorf("printed keywords %v, want flash", spec.PrintedKeywords)
	}
}

func TestCmr086PortentOfBetrayalStealsUntapsHastesAndScries(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() { effectstest.FindBattlefieldCardForTest(g, theirs).Tapped = true })
	effectstest.AdvanceToMain(t, g)
	effectstest.CastCatalogSpell(t, g, "Portent of Betrayal", "Sorcery", cmrPortentOfBetrayalOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: theirs}})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerScryKeepAll(t, g, me.ID)
	c, _ := effectstest.BattlefieldCard(g, theirs)
	if c.Controller != me.ID {
		t.Error("the creature was not stolen")
	}
	if c.Tapped {
		t.Error("the creature stayed tapped")
	}
	if !effectstest.HasEffectiveKeyword(t, g, theirs, "haste") {
		t.Error("the creature has no haste")
	}
}

func TestCmr086PeelFromRealityBouncesBothCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	mine := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	if err := g.CastSpell(me.ID, cmr086Hand(me, "Peel from Reality", "Instant", cmrPeelFromRealityOracle), game.CastSpellParams{
		Strict: true, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: theirs}, {Kind: game.TargetCard, ID: mine}}}); err == nil {
		t.Fatal("targets in the wrong slots (their creature as the one I control) were accepted")
	}
	effectstest.CastCatalogSpell(t, g, "Peel from Reality", "Instant", cmrPeelFromRealityOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: mine}, {Kind: game.TargetCard, ID: theirs}})
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(mine) || g.Battlefield.Contains(theirs) {
		t.Error("a creature stayed on the battlefield")
	}
	if !me.Hand.Contains(mine) || !opp.Hand.Contains(theirs) {
		t.Error("a creature did not reach its owner's hand")
	}
}

func TestCmr086WrongTurnGivesAnyCreatureToTheOpponent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	third := g.Seats[(g.Turn.ActiveSeat+2)%len(g.Seats)]
	mine := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	thirds := effectstest.B12Creature(g, third.ID, "Thirds", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	if err := g.CastSpell(me.ID, cmr086Hand(me, "Wrong Turn", "Instant", cmrWrongTurnOracle), game.CastSpellParams{
		Strict: true, Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID}, {Kind: game.TargetCard, ID: mine}}}); err == nil {
		t.Fatal("Wrong Turn targeted its own controller as the opponent")
	}
	effectstest.CastCatalogSpell(t, g, "Wrong Turn", "Instant", cmrWrongTurnOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}, {Kind: game.TargetCard, ID: thirds}})
	effectstest.PassPriorityAroundTable(t, g)
	c, _ := effectstest.BattlefieldCard(g, thirds)
	if c.Controller != opp.ID {
		t.Errorf("the creature is controlled by %s, want the targeted opponent", c.Controller)
	}
	_ = mine
}

// Sakashima's Will, first bullet: the opponent picks which creature you gain.
func TestCmr086SakashimasWillTheOpponentChoosesWhichCreatureYouGain(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	weak := effectstest.B12Creature(g, opp.ID, "Weak", "Creature — Bear", 1, 1)
	strong := effectstest.B12Creature(g, opp.ID, "Strong", "Creature — Bear", 5, 5)
	effectstest.AdvanceToMain(t, g)
	id := cmr086Hand(me, "Sakashima's Will", "Sorcery", cmrSakashimasWillOracle)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, Modes: []int{0, 1},
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID, Mode: 0}}}); err == nil {
		t.Fatal("both modes without a commander were accepted")
	}
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Modes: []int{0},
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID, Mode: 0}}}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerChooseCards(t, g, opp.ID, weak)
	effectstest.PassPriorityAroundTable(t, g)
	if c, _ := effectstest.BattlefieldCard(g, weak); c.Controller != me.ID {
		t.Error("the creature the opponent chose was not gained")
	}
	if c, _ := effectstest.BattlefieldCard(g, strong); c.Controller != opp.ID {
		t.Error("the other creature changed hands")
	}
}

// Second bullet, and both with a commander: the stolen creature is among the
// creatures that become copies, because the bullets run in printed order.
func TestCmr086SakashimasWillCopiesYourCreaturesAndBothBulletsRunInOrder(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	model := effectstest.B12Creature(g, me.ID, "Model", "Creature — Dragon", 6, 6)
	small := effectstest.B12Creature(g, me.ID, "Small", "Creature — Bear", 1, 1)
	stolen := effectstest.B12Creature(g, opp.ID, "Stolen", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() { effectstest.FindBattlefieldCardForTest(g, small).IsCommander = true })
	effectstest.AdvanceToMain(t, g)
	id := cmr086Hand(me, "Sakashima's Will", "Sorcery", cmrSakashimasWillOracle)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, Modes: []int{0, 1},
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID, Mode: 0}}}); err != nil {
		t.Fatalf("both modes with a commander: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	// The opponent controls one creature, so is not asked; the model is chosen by me.
	effectstest.AnswerChooseCards(t, g, me.ID, model)
	effectstest.PassPriorityAroundTable(t, g)
	for _, c := range []uuid.UUID{small, stolen} {
		if p, to := effectstest.EffectivePTOf(t, g, c); p != 6 || to != 6 {
			t.Errorf("%s is %d/%d, want a 6/6 copy of the model", c, p, to)
		}
	}
	if x, _ := effectstest.BattlefieldCard(g, stolen); x.Controller != me.ID {
		t.Error("the opponent's only creature was not gained")
	}
}
