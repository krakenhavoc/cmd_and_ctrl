package testscr

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// d1657PickTargetView is the wire projection of an open pick_target
// prompt — what the client's picker reads.
func d1657PickTargetView(t *testing.T, g *game.Game, viewer uuid.UUID, p *game.PendingChoice) *protocol.LegalTargetsView {
	t.Helper()
	view := protocol.ViewOfGameFor(g, viewer.String())
	for _, c := range view.PendingChoices {
		if c.ID == p.ID.String() {
			return c.PickTarget
		}
	}
	t.Fatalf("pending choice %s not on the wire", p.ID)
	return nil
}

// The amount is the lands you control AS THE TRIGGER IS PUT ON THE
// STACK: lands that arrive while the prompt is open, and lands that
// leave before resolution, change nothing (Ureni's ruling).
func TestUreniDividesTheLandsYouControlFixedAtAnnounce(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a := effectstest.B12Creature(g, opp.ID, "A", "Creature — Wall", 0, 30)
	b := effectstest.B12Creature(g, opp.ID, "B", "Creature — Wall", 0, 30)
	mine := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Wall", 0, 30)

	p := effectstest.D1657UreniEnters(t, g, 3)
	if d := game.PickTargetDivideForEffect(p); d == nil || d.TotalFor(0) != 3 {
		t.Fatalf("three lands: the prompt divides %+v, want 3", d)
	}
	if pt := d1657PickTargetView(t, g, me.ID, p); pt == nil || pt.Divide == nil || pt.Divide.Total != 3 || pt.Divide.UpTo {
		t.Fatalf("pick_target.divide on the wire: %+v", pt)
	}
	if effectstest.HasID(p.PickTargetCards, mine) {
		t.Error("only creatures and planeswalkers your opponents control")
	}

	// Two more lands while the prompt is open: still 3.
	effectstest.AcLands(g, me.ID, 2, "Forest")
	if err := g.ResolvePickTargetsDivided(p.ID, me.ID, effectstest.CardRefs(a, b), map[uuid.UUID]int{a: 3, b: 2}); !errors.Is(err, game.ErrInvalidParam) {
		t.Fatalf("a 5-point split after the amount was fixed at 3: %v, want ErrInvalidParam", err)
	}
	effectstest.B17PickCardsDivided(t, g, me.ID, map[uuid.UUID]int{a: 1, b: 2}, a, b)

	// And a land that leaves before resolution changes nothing either.
	for _, c := range g.Battlefield.Cards {
		if c.Controller == me.ID && c.IsLand() {
			id := c.InstanceID
			g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(id) })
			break
		}
	}
	effectstest.PassPriorityAroundTable(t, g)
	if d1, d2 := effectstest.E2Card(t, g, a).DamageMarked, effectstest.E2Card(t, g, b).DamageMarked; d1 != 1 || d2 != 2 {
		t.Errorf("Ureni dealt %d/%d, want the announced 1/2", d1, d2)
	}
}

// "You cannot choose more than X targets": one land, two targets is a
// division the gate refuses; one target takes the whole amount.
func TestUreniCannotChooseMoreTargetsThanLands(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a := effectstest.B12Creature(g, opp.ID, "A", "Creature — Wall", 0, 30)
	b := effectstest.B12Creature(g, opp.ID, "B", "Creature — Wall", 0, 30)
	p := effectstest.D1657UreniEnters(t, g, 1)
	if err := g.ResolvePickTargetsDivided(p.ID, me.ID, effectstest.CardRefs(a, b), map[uuid.UUID]int{a: 1, b: 1}); !errors.Is(err, game.ErrInvalidParam) {
		t.Fatalf("two targets for one land: %v, want ErrInvalidParam", err)
	}
	effectstest.B17PickCards(t, g, me.ID, a)
	effectstest.PassPriorityAroundTable(t, g)
	if d := effectstest.E2Card(t, g, a).DamageMarked; d != 1 {
		t.Errorf("the lone target takes the whole 1: %d", d)
	}
}

// Undo with the prompt open keeps the fixed amount (the frame is
// cloned with it), and undo / a snapshot round trip with the trigger on
// the stack keep the announced division.
func TestUreniDivisionSurvivesUndoAndSnapshot(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a := effectstest.B12Creature(g, opp.ID, "A", "Creature — Wall", 0, 30)
	b := effectstest.B12Creature(g, opp.ID, "B", "Creature — Wall", 0, 30)
	effectstest.D1657UreniEnters(t, g, 4)

	undo := g.Clone()
	effectstest.AcLands(g, me.ID, 3, "Forest")
	g.RestoreFrom(undo)
	if d := game.PickTargetDivideForEffect(effectstest.LatestPickTarget(g, me.ID)); d == nil || d.TotalFor(0) != 4 {
		t.Fatalf("after undo the prompt divides %+v, want 4", d)
	}
	effectstest.B17PickCardsDivided(t, g, me.ID, map[uuid.UUID]int{a: 3, b: 1}, a, b)

	blob, err := json.Marshal(g.CaptureSnapshot())
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var round game.GameSnapshot
	if err := json.Unmarshal(blob, &round); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	restored, err := round.Restore()
	if err != nil {
		t.Fatalf("restore: %v", err)
	}

	onStack := g.Clone()
	effectstest.PassPriorityAroundTable(t, g)
	g.RestoreFrom(onStack)
	effectstest.PassPriorityAroundTable(t, g)
	if d1, d2 := effectstest.E2Card(t, g, a).DamageMarked, effectstest.E2Card(t, g, b).DamageMarked; d1 != 3 || d2 != 1 {
		t.Errorf("after undo: %d/%d, want 3/1", d1, d2)
	}
	effectstest.PassPriorityAroundTable(t, restored)
	if d1, d2 := effectstest.E2Card(t, restored, a).DamageMarked, effectstest.E2Card(t, restored, b).DamageMarked; d1 != 3 || d2 != 1 {
		t.Errorf("after a snapshot round trip: %d/%d, want 3/1", d1, d2)
	}
}

// Orca grows as other creatures die, and its death trigger divides the
// power it LAST had — counters included — among any number of targets.
func TestOrcaDiesDividingItsLastKnownPower(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	orca := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Orca, Siege Demon", TypeLine: "Legendary Creature — Demon",
		OracleID: effectstest.D1657OrcaOracle, Power: 5, Toughness: 5, Owner: me.ID, Controller: me.ID,
	})
	fodder := effectstest.B12Creature(g, opp.ID, "Fodder", "Creature — Goblin", 1, 1)
	a := effectstest.B12Creature(g, opp.ID, "A", "Creature — Wall", 0, 30)

	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(fodder) })
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CounterCount(g, orca, game.CounterPlusOne); got != 1 {
		t.Fatalf("another creature died: %d counters on Orca, want 1", got)
	}

	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(orca) })
	effectstest.PassPriorityAroundTable(t, g)
	p := effectstest.LatestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("Orca's death trigger asks for targets")
	}
	if d := game.PickTargetDivideForEffect(p); d == nil || d.TotalFor(0) != 6 {
		t.Fatalf("a 5/5 with one counter divides %+v, want 6", d)
	}
	before := opp.Life
	refs := []game.TargetRef{{Kind: game.TargetCard, ID: a}, {Kind: game.TargetPlayer, ID: opp.ID}}
	if err := g.ResolvePickTargetsDivided(p.ID, me.ID, refs, map[uuid.UUID]int{a: 2, opp.ID: 2}); !errors.Is(err, game.ErrInvalidParam) {
		t.Fatalf("a split of 4 out of 6: %v, want ErrInvalidParam", err)
	}
	if err := g.ResolvePickTargetsDivided(p.ID, me.ID, refs, map[uuid.UUID]int{a: 2, opp.ID: 4}); err != nil {
		t.Fatalf("a legal split: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if d := effectstest.E2Card(t, g, a).DamageMarked; d != 2 {
		t.Errorf("A took %d, want 2", d)
	}
	if got := opp.Life; got != before-4 {
		t.Errorf("opponent at %d, want %d", got, before-4)
	}
}

// Lathiel's "up to": a split that falls short is legal, one that runs
// over is not, no targets at all is the clause's Min 0, and life gained
// in response changes nothing.
func TestLathielUpToDivisionMayFallShortButNotOver(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Lathiel, the Bounteous Dawn", "Legendary Creature — Unicorn", effectstest.B33LathielOracle, 2, 2)
	a := effectstest.B12Creature(g, me.ID, "Bear A", "Creature — Bear", 2, 2)
	b := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)

	effectstest.AdvanceToMainOf(t, g, 1)
	g.WithWriteLock(func() { _ = g.ChangePlayerLifeForEffect(uuid.Nil, me.ID, 3) })
	effectstest.AdvanceToEndStepOf(t, g, 1)
	effectstest.B04WaitForPick(t, g, me.ID)
	p := effectstest.LatestPickTarget(g, me.ID)
	if pt := d1657PickTargetView(t, g, me.ID, p); pt == nil || pt.Divide == nil || pt.Divide.Total != 3 || !pt.Divide.UpTo {
		t.Fatalf("pick_target.divide on the wire: %+v, want up to 3", pt)
	}
	if err := g.ResolvePickTargetsDivided(p.ID, me.ID, effectstest.CardRefs(a, b), map[uuid.UUID]int{a: 2, b: 2}); !errors.Is(err, game.ErrInvalidParam) {
		t.Fatalf("four counters out of up to three: %v, want ErrInvalidParam", err)
	}
	if err := g.ResolvePickTargetsDivided(p.ID, me.ID, effectstest.CardRefs(a, b), map[uuid.UUID]int{a: 1, b: 0}); !errors.Is(err, game.ErrInvalidParam) {
		t.Fatalf("a chosen target given 0: %v, want ErrInvalidParam", err)
	}
	effectstest.B17PickCardsDivided(t, g, me.ID, map[uuid.UUID]int{a: 1, b: 1}, a, b)
	// Life gained in response: the division stands.
	g.WithWriteLock(func() { _ = g.ChangePlayerLifeForEffect(uuid.Nil, me.ID, 5) })
	effectstest.PassPriorityAroundTable(t, g)
	if ca, cb := effectstest.CounterCount(g, a, "+1/+1"), effectstest.CounterCount(g, b, "+1/+1"); ca != 1 || cb != 1 {
		t.Errorf("counters %d/%d, want the announced 1/1", ca, cb)
	}

	// Next turn: no targets at all.
	effectstest.AdvanceToMainOf(t, g, 2)
	g.WithWriteLock(func() { _ = g.ChangePlayerLifeForEffect(uuid.Nil, me.ID, 2) })
	effectstest.AdvanceToEndStepOf(t, g, 2)
	effectstest.B04WaitForPick(t, g, me.ID)
	p = effectstest.LatestPickTarget(g, me.ID)
	if err := g.ResolvePickTargets(p.ID, me.ID, nil); err != nil {
		t.Fatalf("no targets: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if ca, cb := effectstest.CounterCount(g, a, "+1/+1"), effectstest.CounterCount(g, b, "+1/+1"); ca != 1 || cb != 1 {
		t.Errorf("no targets chosen, yet counters moved: %d/%d", ca, cb)
	}
}

// d1657MadnessJudgment discards Avacyn's Judgment to Faithless Looting
// and accepts the madness offer, leaving it castable from exile.
func d1657MadnessJudgment(t *testing.T, g *game.Game) uuid.UUID {
	t.Helper()
	me := g.Seats[g.Turn.ActiveSeat]
	g.WithWriteLock(func() { me.Hand.Cards = nil })
	judgment := effectstest.PushCatalogHandCard(me, "Avacyn's Judgment", "Sorcery", effectstest.D1657AvacynsOracle)
	spare := effectstest.PushCatalogHandCard(me, "Spare", "Sorcery", "")
	effectstest.CastCatalogSpell(t, g, "Faithless Looting", "Sorcery", effectstest.FaithlessLootingOracl, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerChooseCards(t, g, me.ID, judgment, spare)
	effectstest.PassPriorityAroundTable(t, g)
	offer := effectstest.LatestChoiceOfKind(g, game.PendingChoiceResolutionCast)
	if offer == nil {
		t.Fatal("the madness trigger offered no cast")
	}
	// #4250: the cast below answers the resolution_cast prompt.
	return judgment
}

// Avacyn's Judgment divides 2 when cast for {1}{R}, and the announced X
// when cast for its madness cost {X}{R}.
func TestAvacynsJudgmentDividesTwoOrXWithMadness(t *testing.T) {
	t.Run("hard cast divides 2", func(t *testing.T) {
		g := effectstest.NewCatalogGame(t)
		opp := g.Seats[1]
		a := effectstest.B12Creature(g, opp.ID, "A", "Creature — Wall", 0, 30)
		b := effectstest.B12Creature(g, opp.ID, "B", "Creature — Wall", 0, 30)
		if err := effectstest.D1658CastErr(t, g, "Avacyn's Judgment", "Sorcery", effectstest.D1657AvacynsOracle, game.CastSpellParams{
			Targets: effectstest.CardRefs(a, b), Distribution: map[uuid.UUID]int{a: 2, b: 1},
		}); !errors.Is(err, game.ErrInvalidParam) {
			t.Fatalf("3 out of 2: %v, want ErrInvalidParam", err)
		}
		effectstest.D1658Cast(t, g, "Avacyn's Judgment", "Sorcery", effectstest.D1657AvacynsOracle, game.CastSpellParams{
			Targets: effectstest.CardRefs(a, b), Distribution: map[uuid.UUID]int{a: 1, b: 1},
		})
		effectstest.PassPriorityAroundTable(t, g)
		if d1, d2 := effectstest.E2Card(t, g, a).DamageMarked, effectstest.E2Card(t, g, b).DamageMarked; d1 != 1 || d2 != 1 {
			t.Errorf("dealt %d/%d, want 1/1", d1, d2)
		}
	})

	t.Run("madness divides X", func(t *testing.T) {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
		a := effectstest.B12Creature(g, opp.ID, "A", "Creature — Wall", 0, 30)
		b := effectstest.B12Creature(g, opp.ID, "B", "Creature — Wall", 0, 30)
		judgment := d1657MadnessJudgment(t, g)

		// The wire quotes the madness offer's amount as X, and the
		// printed cast's as 2.
		view := protocol.ViewOfGameFor(g, me.ID.String())
		var madness *protocol.AlternativeCostView
		for i := range view.Exile.Cards {
			c := &view.Exile.Cards[i]
			if c.InstanceID != judgment.String() {
				continue
			}
			for j := range c.AlternativeCosts {
				if c.AlternativeCosts[j].Key == game.AltCostKeyMadness {
					madness = &c.AlternativeCosts[j]
				}
			}
		}
		if madness == nil || madness.LegalTargets == nil || madness.LegalTargets.Divide == nil || !madness.LegalTargets.Divide.FromX {
			t.Fatalf("the madness offer's divide should be from_x: %+v", madness)
		}

		// X=1 cannot be divided between two targets.
		if err := g.AddManaForEffect(me.ID, uuid.Nil, "{R}{R}{R}{R}"); err != nil {
			t.Fatalf("AddManaForEffect: %v", err)
		}
		err := g.CastSpell(me.ID, judgment, game.CastSpellParams{
			Strict: true, FromZone: "exile", AlternativeCost: game.AltCostKeyMadness, XValue: 1,
			Targets: effectstest.CardRefs(a, b), Distribution: map[uuid.UUID]int{a: 1, b: 1},
		})
		if !errors.Is(err, game.ErrInvalidParam) {
			t.Fatalf("madness X=1 over two targets: %v, want ErrInvalidParam", err)
		}
		if err := g.CastSpell(me.ID, judgment, game.CastSpellParams{
			Strict: true, FromZone: "exile", AlternativeCost: game.AltCostKeyMadness, XValue: 3,
			Targets: effectstest.CardRefs(a, b), Distribution: map[uuid.UUID]int{a: 1, b: 2},
		}); err != nil {
			t.Fatalf("madness X=3: %v", err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		if d1, d2 := effectstest.E2Card(t, g, a).DamageMarked, effectstest.E2Card(t, g, b).DamageMarked; d1 != 1 || d2 != 2 {
			t.Errorf("dealt %d/%d, want 1/2", d1, d2)
		}
	})
}

// The hand card's view quotes the printed cast's amount: 2.
func TestAvacynsJudgmentViewQuotesTwoInHand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B12Creature(g, g.Seats[1].ID, "A", "Creature — Wall", 0, 30)
	id := effectstest.PushCatalogHandCard(me, "Avacyn's Judgment", "Sorcery", effectstest.D1657AvacynsOracle)
	view := protocol.ViewOfGameFor(g, me.ID.String())
	for _, c := range view.Seats[0].Hand.Cards {
		if c.InstanceID != id.String() {
			continue
		}
		if c.LegalTargets == nil || c.LegalTargets.Divide == nil || c.LegalTargets.Divide.Total != 2 || c.LegalTargets.Divide.FromX {
			t.Fatalf("hand view divide: %+v, want total 2", c.LegalTargets)
		}
		return
	}
	t.Fatal("Avacyn's Judgment not in the hand view")
}

// An amount rule stands alone: a clause that also names a fixed or X
// amount is a card-file bug Register refuses at boot.
func TestRegisterRefusesARuleBesideAnAmount(t *testing.T) {
	effectstest.MustPanic(t, "names an amount rule AND", func() {
		d := effects.DivideBy(effects.DivideLandsYouControl)
		d.Total = 3
		effects.Register(effects.Spec{OracleID: "divide-test-rule-and-total", Name: "Rule And Total",
			Targets: effects.TargetAny().WithCount(0, 0).Dividing(d)})
	})
	effectstest.MustPanic(t, "names an amount rule AND", func() {
		d := effects.DivideBy(effects.DivideLandsYouControl)
		d.FromX = true
		effects.Register(effects.Spec{OracleID: "divide-test-rule-and-x", Name: "Rule And X",
			Targets: effects.TargetAny().WithCount(0, 0).Dividing(d)})
	})
}

// The activation path (CR 602.2b) fixes a ruled amount at activation,
// and the ability row on the wire quotes it.
func TestDividedAmountFixedAtActivation(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a := effectstest.B12Creature(g, opp.ID, "A", "Creature — Wall", 0, 30)
	b := effectstest.B12Creature(g, opp.ID, "B", "Creature — Wall", 0, 30)
	cannon := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Test Land Cannon", TypeLine: "Artifact",
		OracleID: effectstest.D1657LandCannonOracle, Owner: me.ID, Controller: me.ID,
	})
	effectstest.AcLands(g, me.ID, 2, "Wastes")
	effectstest.AdvanceToMain(t, g)

	if row := effectstest.ActivatedRowOf(t, g, me.ID, cannon, 0); row.LegalTargets == nil || row.LegalTargets.Divide == nil || row.LegalTargets.Divide.Total != 2 {
		t.Fatalf("the ability row's divide: %+v, want total 2", row.LegalTargets)
	}

	if err := g.ActivateCatalogAbility(me.ID, cannon, 0, game.ActivateAbilityParams{
		Targets: effectstest.CardRefs(a, b), Distribution: map[uuid.UUID]int{a: 2, b: 1},
	}); !errors.Is(err, game.ErrInvalidParam) {
		t.Fatalf("3 out of two lands: %v, want ErrInvalidParam", err)
	}
	if err := g.ActivateCatalogAbility(me.ID, cannon, 0, game.ActivateAbilityParams{
		Targets: effectstest.CardRefs(a, b), Distribution: map[uuid.UUID]int{a: 1, b: 1},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.AcLands(g, me.ID, 3, "Wastes")
	effectstest.PassPriorityAroundTable(t, g)
	if d1, d2 := effectstest.E2Card(t, g, a).DamageMarked, effectstest.E2Card(t, g, b).DamageMarked; d1 != 1 || d2 != 1 {
		t.Errorf("dealt %d/%d, want the announced 1/1", d1, d2)
	}
}

// A single-clause divided ACTIVATED ability (Mogg Mob) carries its
// amount on the ability row — before #1657 abilityClauseView dropped
// it, so the client never asked for the split the gate demands.
func TestMoggMobAbilityRowCarriesTheDivision(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B12Creature(g, g.Seats[1].ID, "A", "Creature — Wall", 0, 30)
	mob := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Mogg Mob", TypeLine: "Creature — Goblin",
		OracleID: moggMobOracle, Power: 3, Toughness: 3, Owner: me.ID, Controller: me.ID,
	})
	if row := effectstest.ActivatedRowOf(t, g, me.ID, mob, 0); row.LegalTargets == nil || row.LegalTargets.Divide == nil || row.LegalTargets.Divide.Total != 3 {
		t.Fatalf("Mogg Mob's ability row: %+v, want divide total 3", row.LegalTargets)
	}
}
