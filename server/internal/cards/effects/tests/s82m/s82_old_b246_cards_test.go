package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// S82 Amonkhet batch 246 (#5535): pump and combat tricks.

const (
	akh246Favor    = "2071b712-2539-4797-8359-f9adb31889ae"
	akh246Lizard   = "46a69558-fad1-4a04-9bb6-8f971bbbb038"
	akh246Strider  = "b9dcdf5b-0fdb-4d8e-8ddf-08e44f67018d"
	akh246Shed     = "e3390776-66cc-4e0a-a034-701793b05db8"
	akh246Name     = "f4a8756d-94df-40bc-a363-2938ada0a808"
	akh246Captain  = "fd561a9a-45d2-4e9b-82df-5d4b24cca24a"
	akh246Zombie   = "Creature — Zombie"
	akh246BearType = "Creature — Bear"
)

func TestAkh246CardsAreRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		akh246Favor: "Hazoret's Favor", akh246Lizard: "Thresher Lizard", akh246Strider: "Grim Strider",
		akh246Shed: "Shed Weakness", akh246Name: "In Oketra's Name", akh246Captain: "Honored Crop-Captain",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

// akh246SetHand gives `p` exactly n cards in hand, through real draws so the
// layer cache sees the hand change.
func akh246SetHand(t *testing.T, g *game.Game, p *game.Player, n int) {
	t.Helper()
	for p.Hand.Size() > n {
		g.WithWriteLock(func() { _ = g.DiscardRandomForEffect(p.ID, 1) })
	}
	for p.Hand.Size() < n {
		effectstest.PushLibraryCardForTest(p, game.Card{Name: "Filler", TypeLine: "Land"})
		if err := g.DrawCard(p.ID); err != nil {
			t.Fatalf("draw: %v", err)
		}
	}
}

func TestAkh246ThresherLizardGetsPlusOnePlusTwoWithOneOrFewerCardsInHand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	lizard := effectstest.PushTestPermanent(g, me.ID, game.Card{Name: "Thresher Lizard", OracleID: akh246Lizard,
		TypeLine: "Creature — Lizard", Power: 3, Toughness: 2})
	for _, c := range []struct{ hand, power, toughness int }{{3, 3, 2}, {2, 3, 2}, {1, 4, 4}, {0, 4, 4}, {2, 3, 2}} {
		akh246SetHand(t, g, me, c.hand)
		if p, tough := effectstest.EffectivePTOf(t, g, lizard); p != c.power || tough != c.toughness {
			t.Errorf("with %d cards in hand: %d/%d, want %d/%d", c.hand, p, tough, c.power, c.toughness)
		}
	}
}

func TestAkh246ThresherLizardReadsOnlyItsControllersHand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	lizard := effectstest.PushTestPermanent(g, me.ID, game.Card{Name: "Thresher Lizard", OracleID: akh246Lizard,
		TypeLine: "Creature — Lizard", Power: 3, Toughness: 2})
	akh246SetHand(t, g, me, 4)
	akh246SetHand(t, g, opp, 0)
	if p, tough := effectstest.EffectivePTOf(t, g, lizard); p != 3 || tough != 2 {
		t.Errorf("an empty opposing hand pumped it to %d/%d", p, tough)
	}
}

func TestAkh246GrimStriderShrinksForEachCardInHand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	strider := effectstest.PushTestPermanent(g, me.ID, game.Card{Name: "Grim Strider", OracleID: akh246Strider,
		TypeLine: "Creature — Horror", Power: 6, Toughness: 6})
	for _, n := range []int{0, 2, 5, 1} {
		akh246SetHand(t, g, me, n)
		if p, tough := effectstest.EffectivePTOf(t, g, strider); p != 6-n || tough != 6-n {
			t.Errorf("with %d cards in hand: %d/%d, want %d/%d", n, p, tough, 6-n, 6-n)
		}
	}
}

func TestAkh246ShedWeaknessPumpsAndMayRemoveAMinusOneCounter(t *testing.T) {
	for _, tc := range []struct {
		name   string
		accept bool
		want   int
	}{{"accept", true, 0}, {"decline", false, 1}} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[0]
			bear := effectstest.B12Creature(g, me.ID, "Bear", akh246BearType, 3, 3)
			effectstest.PutCounters(t, g, bear, game.CounterMinusOne, 1)
			p0, t0 := effectstest.EffectivePTOf(t, g, bear)
			effectstest.CastCatalogSpell(t, g, "Shed Weakness", "Instant", akh246Shed, effectstest.CardTarget(bear))
			effectstest.PassPriorityAroundTable(t, g)
			effectstest.AnswerMayChoice(t, g, me.ID, tc.accept)
			effectstest.PassPriorityAroundTable(t, g)
			if got := effectstest.CounterCount(g, bear, game.CounterMinusOne); got != tc.want {
				t.Errorf("-1/-1 counters %d, want %d", got, tc.want)
			}
			// +2/+2 until end of turn either way (the fixture's base size ignores counters).
			wantP, wantT := p0+2, t0+2
			if p, tough := effectstest.EffectivePTOf(t, g, bear); p != wantP || tough != wantT {
				t.Errorf("Bear is %d/%d, want %d/%d", p, tough, wantP, wantT)
			}
		})
	}
}

func TestAkh246ShedWeaknessAsksNothingWithoutACounter(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.B12Creature(g, me.ID, "Bear", akh246BearType, 2, 2)
	effectstest.CastCatalogSpell(t, g, "Shed Weakness", "Instant", akh246Shed, effectstest.CardTarget(bear))
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestConfirmFor(g, me.ID) != nil {
		t.Error("asked to remove a counter from a creature that has none")
	}
	if p, tough := effectstest.EffectivePTOf(t, g, bear); p != 4 || tough != 4 {
		t.Errorf("Bear is %d/%d, want 4/4", p, tough)
	}
}

func TestAkh246InOketrasNameGivesZombiesTwoOneAndOthersOneOne(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	zombie := effectstest.B12Creature(g, me.ID, "Zed", akh246Zombie, 2, 2)
	bear := effectstest.B12Creature(g, me.ID, "Bear", akh246BearType, 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Zombie", akh246Zombie, 2, 2)
	effectstest.CastCatalogSpell(t, g, "In Oketra's Name", "Instant", akh246Name, nil)
	effectstest.PassPriorityAroundTable(t, g)
	for _, c := range []struct {
		name         string
		id           uuid.UUID
		power, tough int
	}{{"my Zombie", zombie, 4, 3}, {"my Bear", bear, 3, 3}, {"their Zombie", theirs, 2, 2}} {
		if p, tough := effectstest.EffectivePTOf(t, g, c.id); p != c.power || tough != c.tough {
			t.Errorf("%s is %d/%d, want %d/%d", c.name, p, tough, c.power, c.tough)
		}
	}
}

func TestAkh246InOketrasNameLeavesLaterCreaturesAlone(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B12Creature(g, me.ID, "Bear", akh246BearType, 2, 2)
	effectstest.CastCatalogSpell(t, g, "In Oketra's Name", "Instant", akh246Name, nil)
	effectstest.PassPriorityAroundTable(t, g)
	late := effectstest.B12Creature(g, me.ID, "Late", akh246BearType, 2, 2)
	if p, tough := effectstest.EffectivePTOf(t, g, late); p != 2 || tough != 2 {
		t.Errorf("a creature that arrived later is %d/%d, want 2/2", p, tough)
	}
}

func TestAkh246HonoredCropCaptainPumpsOtherAttackersOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	captain := effectstest.PushCatalogPermanent(g, me.ID, "Honored Crop-Captain", "Creature — Human Warrior", akh246Captain, false)
	bear := effectstest.PushCatalogPermanent(g, me.ID, "Bear", akh246BearType, "", false)
	idle := effectstest.PushCatalogPermanent(g, me.ID, "Idle", akh246BearType, "", false)
	p0, t0 := effectstest.EffectivePTOf(t, g, captain)
	b0, bt0 := effectstest.EffectivePTOf(t, g, bear)
	i0, _ := effectstest.EffectivePTOf(t, g, idle)
	effectstest.AdvanceToDeclareAttackersOf(t, g, 0)
	if err := g.DeclareAttacker(captain, opp.ID); err != nil {
		t.Fatalf("attack with the Captain: %v", err)
	}
	if err := g.DeclareAttacker(bear, opp.ID); err != nil {
		t.Fatalf("attack with the Bear: %v", err)
	}
	effectstest.LockInAttacks(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.EffectivePTOf(t, g, captain); p != p0 || tough != t0 {
		t.Errorf("the Captain is %d/%d, want %d/%d", p, tough, p0, t0)
	}
	if p, tough := effectstest.EffectivePTOf(t, g, bear); p != b0+1 || tough != bt0 {
		t.Errorf("the other attacker is %d/%d, want %d/%d", p, tough, b0+1, bt0)
	}
	if p, _ := effectstest.EffectivePTOf(t, g, idle); p != i0 {
		t.Errorf("a creature that did not attack got +%d power", p-i0)
	}
}

func TestAkh246HazoretsFavorPumpsGivesHasteAndSacrificesAtTheEndStep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	effectstest.PushCatalogPermanent(g, me.ID, "Hazoret's Favor", "Enchantment", akh246Favor, false)
	bear := effectstest.PushCatalogPermanent(g, me.ID, "Bear", akh246BearType, "", true)
	p0, t0 := effectstest.EffectivePTOf(t, g, bear)
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	if prompt := effectstest.LatestPickTarget(g, me.ID); prompt != nil {
		effectstest.PickCard(t, g, me.ID, bear)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.EffectivePTOf(t, g, bear); p != p0+2 || tough != t0 {
		t.Fatalf("the creature is %d/%d, want %d/%d", p, tough, p0+2, t0)
	}
	if !effectstest.HasKeywordOf(g, bear, "haste") {
		t.Error("the creature did not gain haste")
	}
	effectstest.AdvanceToEndStep(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.BattlefieldHas(g, bear) {
		t.Error("the creature was not sacrificed at the beginning of the next end step")
	}
}

func TestAkh246HazoretsFavorDeclinedDoesNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	effectstest.PushCatalogPermanent(g, me.ID, "Hazoret's Favor", "Enchantment", akh246Favor, false)
	bear := effectstest.PushCatalogPermanent(g, me.ID, "Bear", akh246BearType, "", false)
	p0, _ := effectstest.EffectivePTOf(t, g, bear)
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	if prompt := effectstest.LatestPickTarget(g, me.ID); prompt != nil {
		effectstest.PickCard(t, g, me.ID, bear)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if p, _ := effectstest.EffectivePTOf(t, g, bear); p != p0 {
		t.Errorf("declining still pumped the creature to %d power", p)
	}
	effectstest.AdvanceToEndStep(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.BattlefieldHas(g, bear) {
		t.Error("the creature was sacrificed although the player declined")
	}
}
