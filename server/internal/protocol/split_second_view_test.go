package protocol

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// split_second_view_test.go — #1519. Split second (CR 702.61) is read
// off the card now, so it is on for real the moment a Krosan Grip is
// cast — and the view has to say what the engine and the bot
// enumerator say. Three surfaces had to agree:
//
//   - the move list: no cast, no activation, but mana and pass;
//   - `castable_here` on a graveyard card: it had no split-second
//     input at all, and lit a flashback instant the engine would
//     refuse;
//   - `timing_closed` on an INSTANT-speed ability row: it only ever
//     spoke for sorcery-speed and loyalty rows, so a Goblin
//     Bombardment row stayed live under split second.
//
// Both stamps now read the engine's two timing functions, which refuse
// under split second, so the property is checked the way
// TestViewAndEnumeratorAgreeOnCastTiming checks it: view and
// enumerator, same seat, before and after.

const oracleKrosanGrip = "3e39224c-72ce-4ecc-aa17-12c071ea1f3e"

func TestViewAndEnumeratorAgreeUnderSplitSecond(t *testing.T) {
	const oracleFlashbackInstant = "test-split-second-flashback-instant"
	g := busyTable(t, 1)
	me := g.Seats[g.Turn.ActiveSeat]
	them := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]

	withCastableZones(t, map[string][]game.ZoneKind{oracleFlashbackInstant: {game.ZoneGraveyard}})
	withAltCostsByOracle(t, map[string][]game.AlternativeCost{
		oracleFlashbackInstant: {{
			Key: "flashback", Label: "Flashback {R}", ManaCost: "{R}",
			FromZone: game.ZoneGraveyard,
		}},
	})
	me.Graveyard.Cards = nil
	flash := graveyardCard(me, "Split Second Test Instant", oracleFlashbackInstant)
	me.Graveyard.Cards[0].TypeLine = "Instant"
	knownToEveryone(g, me.Graveyard, flash)

	var bombardment, theirBombardment uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.OracleID != oracleGoblinBombardment {
			continue
		}
		switch c.Controller {
		case me.ID:
			bombardment = c.InstanceID
		case them.ID:
			theirBombardment = c.InstanceID
		}
	}
	if bombardment == uuid.Nil || theirBombardment == uuid.Nil {
		t.Fatal("setup: busyTable seats no Goblin Bombardment")
	}
	// busyTable's put stamps no knowers, and a per-seat frame redacts
	// an unknown card's rows; the battlefield is public in play.
	knownToEveryone(g, g.Battlefield, bombardment)

	type answer struct {
		castableHere, offeredFromGraveyard bool
		timingClosed                       bool
		kinds                              map[legal.Kind]int
		activatesBombardment               bool
	}
	read := func() answer {
		t.Helper()
		v := ViewOfGameFor(g, me.ID.String())
		a := answer{kinds: map[legal.Kind]int{}}
		a.castableHere = cardInSeatZone(t, v.Seats[me.Seat].Graveyard, flash).CastableHere
		a.offeredFromGraveyard = offeredFromGraveyard(g, me.ID, flash)
		rows := cardInSeatZone(t, v.Battlefield, bombardment).ActivatedAbilities
		if len(rows) == 0 {
			t.Fatal("Goblin Bombardment has no ability row")
		}
		a.timingClosed = rows[0].TimingClosed
		for _, m := range v.LegalMoves {
			a.kinds[m.Kind]++
			if m.Kind == legal.KindActivate && m.Source == bombardment {
				a.activatesBombardment = true
			}
		}
		return a
	}

	before := read()
	if !before.castableHere || !before.offeredFromGraveyard {
		t.Fatalf("setup: the flashback instant is not a cast surface before split second (view %v, enumerator %v)",
			before.castableHere, before.offeredFromGraveyard)
	}
	if before.timingClosed || !before.activatesBombardment {
		t.Fatalf("setup: the Bombardment row is not live before split second (timing_closed %v, move %v)",
			before.timingClosed, before.activatesBombardment)
	}

	grip := put(me.Hand, me, game.Card{
		Name: "Krosan Grip", TypeLine: "Instant", ManaCost: "{2}{G}", OracleID: oracleKrosanGrip,
	})
	if err := g.CastSpell(me.ID, grip, game.CastSpellParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: theirBombardment}},
	}); err != nil {
		t.Fatalf("cast Krosan Grip: %v", err)
	}

	v := ViewOfGameFor(g, me.ID.String())
	if !v.SplitSecondActive {
		t.Fatal("split_second_active is false with Krosan Grip on the stack")
	}
	var stamped bool
	for _, it := range v.StackItems {
		if it.ID == grip.String() && it.SplitSecond {
			stamped = true
		}
	}
	if !stamped {
		t.Error("the Grip's stack item does not carry split_second on the wire")
	}

	after := read()
	if after.castableHere != after.offeredFromGraveyard {
		t.Errorf("castable_here=%v but the enumerator says %v", after.castableHere, after.offeredFromGraveyard)
	}
	if after.castableHere {
		t.Error("a graveyard instant is still a cast surface under split second")
	}
	if after.timingClosed == after.activatesBombardment {
		t.Errorf("timing_closed=%v but the enumerator offers the activation: %v", after.timingClosed, after.activatesBombardment)
	}
	if !after.timingClosed {
		t.Error("an instant-speed ability row is live under split second")
	}
	if n := after.kinds[legal.KindCast] + after.kinds[legal.KindActivate]; n != 0 {
		t.Errorf("%d cast/activate moves offered under split second", n)
	}
	if after.kinds[legal.KindPass] == 0 {
		t.Error("pass is not offered under split second")
	}
	if after.kinds[legal.KindMana] == 0 {
		t.Error("CR 702.61b: no mana move offered under split second")
	}
}
