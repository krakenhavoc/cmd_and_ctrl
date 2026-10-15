package effects

import (
	"strings"
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// conditional_purpose_test.go — ADR 0126's amendment of 2026-10-15
// (#3762): Purpose.Kicked and Purpose.FaceUp say what a card does only
// when its kicker is paid or when it is turned face up. The cards that
// needed them declare them, the wire carries them, and Register refuses
// a declaration a reader could not recognise.

func TestWaterspoutElementalDeclaresItsWipeAndSkipUnderKicked(t *testing.T) {
	s := specNamed(t, "Waterspout Elemental")
	p := s.Purpose
	// Nothing it does however it is cast: the unkicked cast is a body.
	base := p
	base.Kicked = nil
	if !base.IsZero() {
		t.Errorf("the card declares %+v outside Kicked: an unkicked cast would be priced by it", base)
	}
	k := p.Kicked
	if k == nil {
		t.Fatal("no Purpose.Kicked")
	}
	if k.Sweep.Matches != game.SweepCreatures || k.Sweep.How != game.SweepBounce {
		t.Errorf("kicked sweep = %+v, want a creature bounce", k.Sweep)
	}
	if k.SkipsTurns != 1 {
		t.Errorf("kicked SkipsTurns = %d, want 1", k.SkipsTurns)
	}
	// The registry audit (TestRealDumpPurposeAudit, rule 27) asks this.
	if !p.DeclaresSweep() {
		t.Error("DeclaresSweep is false for a card whose kicked cast is a wipe")
	}
}

func TestUnstableHulkDeclaresItsSkipUnderFaceUp(t *testing.T) {
	p := specNamed(t, "Unstable Hulk").Purpose
	if p.FaceUp == nil || p.FaceUp.SkipsTurns != 1 {
		t.Fatalf("Purpose.FaceUp = %+v, want a skipped turn", p.FaceUp)
	}
	base := p
	base.FaceUp = nil
	if !base.IsZero() {
		t.Errorf("the card declares %+v outside FaceUp: turning it up is not what casts it", base)
	}
	if p.DeclaresSweep() {
		t.Error("Unstable Hulk reads as a wipe")
	}
	// #3776: the +6/+6 and trample the trigger gives ride the same slot.
	pump := p.FaceUp.Pump
	if pump == nil || pump.Power != 6 || pump.Toughness != 6 || len(pump.Keywords) != 1 || pump.Keywords[0] != "trample" {
		t.Errorf("Purpose.FaceUp.Pump = %+v, want +6/+6 and trample", pump)
	}
}

// The wire: the hand card carries `kicked` and nothing the unkicked cast
// is priced by, and Priced() agrees.
func TestWaterspoutElementalKickedPurposeIsOnTheWire(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	id := handCardOf(g, me, "Waterspout Elemental", "Creature — Elemental", "{3}{U}{U}", waterspoutOracle)
	v := protocol.ViewOfGameFor(g, me.ID.String())
	var got *protocol.PurposeView
	for _, s := range v.Seats {
		for i := range s.Hand.Cards {
			if s.Hand.Cards[i].InstanceID == id.String() {
				got = s.Hand.Cards[i].Purpose
			}
		}
	}
	if got == nil || got.Kicked == nil {
		t.Fatalf("no purpose.kicked on the hand card: %+v", got)
	}
	if got.Kicked.Sweep == nil || got.Kicked.Sweep.How != "bounce" || got.Kicked.SkipsTurns != 1 {
		t.Errorf("kicked = %+v, want the bounce and the skip", got.Kicked)
	}
	if got.Sweep != nil || got.SkipsTurns != 0 {
		t.Errorf("the unkicked cast declares %+v", got)
	}
	if got.Priced() {
		t.Error("a purpose that is only Kicked is Priced: the unkicked cast would drop its proxy for it")
	}
}

// The wire: a face-down Hulk's CONTROLLER is told what turning it up
// does, and only that; an opponent is told nothing (CR 708.2a, 708.5).
func TestUnstableHulkFaceDownPurposeReachesOnlyItsController(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	id := castWithAltCost(t, g, "Unstable Hulk", "Creature — Goblin Mutant", unstableHulkOracle, "morph")
	passPriorityAroundTable(t, g)

	purposeOf := func(viewer string) (*protocol.PurposeView, bool) {
		v := protocol.ViewOfGameFor(g, viewer)
		for i := range v.Battlefield.Cards {
			if v.Battlefield.Cards[i].InstanceID == id.String() {
				return v.Battlefield.Cards[i].Purpose, v.Battlefield.Cards[i].FaceDown
			}
		}
		t.Fatal("the face-down Hulk is not on the battlefield")
		return nil, false
	}
	mine, down := purposeOf(me.ID.String())
	if !down {
		t.Fatal("the Hulk is not face down")
	}
	if mine == nil || mine.TurnedFaceUp == nil || mine.TurnedFaceUp.SkipsTurns != 1 {
		t.Fatalf("the controller's purpose = %+v, want turned_face_up with a skip", mine)
	}
	if pm := mine.TurnedFaceUp.Pump; pm == nil || pm.Power != 6 || pm.Toughness != 6 || len(pm.Keywords) != 1 || pm.Keywords[0] != "trample" {
		t.Errorf("turned_face_up.pump = %+v, want +6/+6 and trample (#3776)", pm)
	}
	stripped := *mine
	stripped.TurnedFaceUp = nil
	if stripped != (protocol.PurposeView{}) {
		t.Errorf("a face-down permanent has no text, yet the view says it %+v", stripped)
	}
	for _, o := range g.Seats {
		if o.ID == me.ID {
			continue
		}
		if theirs, _ := purposeOf(o.ID.String()); theirs != nil {
			t.Errorf("seat %s sees purpose %+v on an opponent's face-down permanent", o.ID, theirs)
		}
	}
}

func TestConditionalPurposeGuard(t *testing.T) {
	kicker := []game.AdditionalCost{Kicker("{U}")}
	twoKickers := []game.AdditionalCost{Kicker("{R}"), Kicker("{W}")}
	sweep := &game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepBounce}, SkipsTurns: 1}
	morph := []game.AlternativeCost{Morph("{1}{R}")}
	for _, c := range []struct {
		name string
		spec Spec
		want string
	}{
		{"kicked on a one-kicker card", Spec{Purpose: game.Purpose{Kicked: sweep}, OptionalCosts: kicker}, ""},
		{"face up on a morph card", Spec{Purpose: game.Purpose{FaceUp: &game.Purpose{SkipsTurns: 1}}, AlternativeCosts: morph}, ""},
		{"kicked with no optional cost", Spec{Purpose: game.Purpose{Kicked: sweep}}, "exactly one kicker"},
		{"kicked with two kickers", Spec{Purpose: game.Purpose{Kicked: sweep}, OptionalCosts: twoKickers}, "exactly one kicker"},
		{"kicked on a buyback", Spec{Purpose: game.Purpose{Kicked: sweep}, OptionalCosts: []game.AdditionalCost{{Optional: true, Key: "buyback"}}}, "exactly one kicker"},
		{"face up with a pump", Spec{Purpose: game.Purpose{FaceUp: &game.Purpose{Pump: &game.Pump{Power: 6, Toughness: 6, Keywords: []string{"trample"}}}}, AlternativeCosts: morph}, ""},
		{"face up with a pump that gives nothing", Spec{Purpose: game.Purpose{FaceUp: &game.Purpose{Pump: &game.Pump{}}}, AlternativeCosts: morph}, "gives nothing"},
		{"kicked with a pump", Spec{Purpose: game.Purpose{Kicked: &game.Purpose{Pump: &game.Pump{Power: 1}}}, OptionalCosts: kicker}, "Pump off a triggered"},
		{"face up with no face-down cast", Spec{Purpose: game.Purpose{FaceUp: &game.Purpose{SkipsTurns: 1}}}, "no face-down cast"},
		{"kicked that says nothing", Spec{Purpose: game.Purpose{Kicked: &game.Purpose{}}, OptionalCosts: kicker}, "says nothing"},
		{"kicked that nests", Spec{Purpose: game.Purpose{Kicked: &game.Purpose{Kicked: sweep}}, OptionalCosts: kicker}, "nests"},
		{"kicked with a target entry", Spec{Purpose: game.Purpose{Kicked: &game.Purpose{Targets: game.ForTargets(game.TargetPurpose{Slot: 0, Draws: 1})}}, OptionalCosts: kicker}, "Targets"},
		{"kicked with a negative amount", Spec{Purpose: game.Purpose{Kicked: &game.Purpose{SkipsTurns: -1}}, OptionalCosts: kicker}, "negative amount"},
		{"kicked with a payoff only a triggered row has", Spec{Purpose: game.Purpose{Kicked: &game.Purpose{DeathPayoff: true}}, OptionalCosts: kicker}, "DeathPayoff"},
	} {
		t.Run(c.name, func(t *testing.T) {
			c.spec.Name = "Test Card"
			var msg string
			func() {
				defer func() {
					if r := recover(); r != nil {
						msg, _ = r.(string)
					}
				}()
				checkSpecPurposes(c.spec)
			}()
			if c.want == "" && msg != "" {
				t.Fatalf("refused: %s", msg)
			}
			if c.want != "" && !strings.Contains(msg, c.want) {
				t.Fatalf("got %q, want a refusal naming %q", msg, c.want)
			}
		})
	}
	// Off the card slot, whatever the card is.
	for _, slot := range []purposeSlot{purposeOnMode, purposeOnAltCost, purposeOnActivated, purposeOnTriggered} {
		var msg string
		func() {
			defer func() {
				if r := recover(); r != nil {
					msg, _ = r.(string)
				}
			}()
			checkPurpose("Test Card", "row", slot, game.Purpose{Kicked: sweep})
		}()
		if !strings.Contains(msg, "off the card") {
			t.Errorf("slot %d: got %q, want a refusal off the card", slot, msg)
		}
	}
}
