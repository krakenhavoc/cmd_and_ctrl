package heuristic_test

import (
	"context"
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/heuristic"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// cant_block_test.go — #4977: "can't block this turn" priced by the
// attack it opens before the defender's blocks, not as removal.

// daringDiscovery is a seat-0 sorcery declaring "up to three target
// creatures can't block this turn" and its discover 4.
func daringDiscovery(id string) protocol.CardView {
	c := spell(id, 0, "Daring Discovery", "{4}{R}")
	c.TypeLine = "Sorcery"
	c.Purpose = &protocol.PurposeView{Discover: 4, Targets: &[]protocol.TargetPurposeView{
		{Slot: 0, UntilEndOfTurn: "harm", Pump: &protocol.PumpView{CantBlock: true}},
	}}
	return c
}

func fiveMountains() []protocol.CardView {
	return []protocol.CardView{land(cardID(91), 0), land(cardID(92), 0), land(cardID(93), 0), land(cardID(94), 0), land(cardID(95), 0)}
}

// cantBlockOn is today's heuristic with PriceCantBlock on, which
// DefaultConfig ships off (#4977's arena); noCantBlock is the default.
func cantBlockOn() *heuristic.Policy {
	return withConfig(func(c *heuristic.Config) { c.PriceCantBlock = true })
}

func noCantBlock() *heuristic.Policy { return heuristic.New() }

// valueOf is move i's price in a ranking.
func valueOf(t *testing.T, pol *heuristic.Policy, v protocol.GameView, i int, moves ...legal.Move) float64 {
	t.Helper()
	for _, c := range pol.Rank(context.Background(), input(0, v, moves...)) {
		if c.Index == i {
			return c.Value
		}
	}
	t.Fatalf("move %d is not ranked", i)
	return 0
}

// walledIn is the bot's three 3/3s, able to attack, against an opponent's
// three untapped 4/4s that would block each of them at a loss.
func walledIn(extra ...protocol.CardView) []protocol.CardView {
	board := append(fiveMountains(),
		creature(cardID(3), 0, "Bear A", 3, 3), creature(cardID(4), 0, "Bear B", 3, 3), creature(cardID(5), 0, "Bear C", 3, 3),
		creature(cardID(6), 1, "Wall A", 4, 4), creature(cardID(7), 1, "Wall B", 4, 4), creature(cardID(8), 1, "Wall C", 4, 4))
	return append(board, extra...)
}

// TestDaringDiscoveryIsWorthTheAttacksItOpens: the issue's case. Before
// attacks, the three 4/4s unable to block let the three 3/3s through, so
// the cast is worth that attack (and its discover), and far more than the
// same cast after combat, where it is the discover alone. It is not three
// removal spells: with PriceCantBlock off the same cast is priced that
// way, well above the attack.
func TestDaringDiscoveryIsWorthTheAttacksItOpens(t *testing.T) {
	dd := daringDiscovery(cardID(90))
	all := castMove(t, 0, dd.InstanceID, "Cast Daring Discovery at all three", cardTarget(cardID(6)), cardTarget(cardID(7)), cardTarget(cardID(8)))
	one := castMove(t, 0, dd.InstanceID, "Cast Daring Discovery at one", cardTarget(cardID(6)))
	none := castMove(t, 0, dd.InstanceID, "Cast Daring Discovery for its discover")
	moves := []legal.Move{passMove(0), all, one, none}
	view := func(step string) protocol.GameView {
		return newView([]protocol.PlayerView{newSeat(0, withHand(dd)), newSeat(1)}, withBattlefield(walledIn()...), withTurn(5, 0, step))
	}
	pol := cantBlockOn()
	before := valueOf(t, pol, view("precombat_main"), 1, moves...)
	after := valueOf(t, pol, view("postcombat_main"), 1, moves...)
	discoverOnly := valueOf(t, pol, view("precombat_main"), 3, moves...)
	removal := valueOf(t, noCantBlock(), view("precombat_main"), 1, moves...)
	t.Logf("all three before attacks %.2f, after combat %.2f, no target %.2f, as three removals %.2f", before, after, discoverOnly, removal)
	if before < after+2.0 {
		t.Errorf("before attacks the cast is %.2f and after combat %.2f; want the three 3/3s' attack between them", before, after)
	}
	if after > discoverOnly+0.01 {
		t.Errorf("after combat the cast at three creatures is %.2f, the discover alone %.2f; want nothing for the targets", after, discoverOnly)
	}
	if removal < before+2.0 {
		t.Errorf("as removal the cast is %.2f, by its attack %.2f; want the removal price well above the attack", removal, before)
	}
	// One 4/4 that can't block opens nothing: the other two still block
	// the first two attackers at a loss, and the third is left alone.
	if v := valueOf(t, pol, view("precombat_main"), 2, moves...); v > discoverOnly+0.01 {
		t.Errorf("one of three blockers is priced %.2f over the discover alone (%.2f); want nothing", v-discoverOnly, discoverOnly)
	}
	if i, why := decideOn(t, pol, view("precombat_main"), moves...); i != 1 {
		t.Errorf("picked %d (%s), want the cast at all three blockers before attacks", i, why)
	}
}

// cantBlockRow is a seat-0 permanent with "{1}{R}: Target creature can't
// block this turn." (Fifty Feet of Rope's shape, at instant speed).
func cantBlockRow(id string) protocol.CardView {
	return granter(id, "{1}{R}: Target creature can't block this turn.", "harm", &protocol.PumpView{CantBlock: true})
}

// TestCantBlockIsNothingAfterBlocks: the bot's 3/3 attacks into an
// opponent's untapped 4/4. In the declare attackers step "can't block"
// on the 4/4 is the attack through, and the bot makes it. Once blockers
// are declared it changes nothing (CR 509.1b), on the opponent's turn the
// bot is not attacking, and outside combat it is gone at cleanup: in
// each the bot passes. Without PriceCantBlock it is removal everywhere.
func TestCantBlockIsNothingAfterBlocks(t *testing.T) {
	src, wall := cardID(1), cardID(2)
	board := func(step string, active int, opts ...protocol.CardView) protocol.GameView {
		cards := append(twoPlains(), cantBlockRow(src), creature(wall, 1, "Their Wall", 4, 4))
		cards = append(cards, opts...)
		return newView([]protocol.PlayerView{newSeat(0), newSeat(1)}, withBattlefield(cards...), withTurn(5, active, step))
	}
	moves := []legal.Move{passMove(0), grantMove(t, src, wall)}
	attacker := creature(cardID(3), 0, "Bear", 3, 3, attacking(1))
	if i, why := decideOn(t, cantBlockOn(), board("declare_attackers", 0, attacker), moves...); i != 1 {
		t.Errorf("picked %d (%s), want the 4/4 unable to block the attack", i, why)
	}
	for _, c := range []struct {
		name string
		view protocol.GameView
	}{
		{"after blockers", board("declare_blockers", 0, attacker)},
		{"after combat", board("postcombat_main", 0, creature(cardID(3), 0, "Bear", 3, 3, tapped()))},
		{"on the opponent's turn", board("precombat_main", 1, creature(cardID(3), 0, "Bear", 3, 3))},
	} {
		t.Run(c.name, func(t *testing.T) {
			if v := valueOf(t, cantBlockOn(), c.view, 1, moves...); v > 0.6 {
				t.Errorf("the restriction is priced %.2f, want nothing for it", v)
			}
			if v := valueOf(t, noCantBlock(), c.view, 1, moves...); v < 1.0 {
				t.Errorf("with PriceCantBlock off it is priced %.2f, want the old removal price", v)
			}
		})
	}
}

// TestCantBlockOpensTheLethalSwing: the opponent is on 6 with one
// untapped 1/1, and the bot has two 3/3s that can attack. With the 1/1
// unable to block, the all-in swing is lethal, so "can't block" before
// attacks is worth LethalBonus on top of the damage.
func TestCantBlockOpensTheLethalSwing(t *testing.T) {
	src, chump := cardID(1), cardID(2)
	cards := append(twoPlains(), cantBlockRow(src), creature(chump, 1, "Their Squirrel", 1, 1),
		creature(cardID(3), 0, "Bear A", 3, 3), creature(cardID(4), 0, "Bear B", 3, 3))
	v := newView([]protocol.PlayerView{newSeat(0), newSeat(1, withLife(6))}, withBattlefield(cards...), withTurn(5, 0, "precombat_main"))
	moves := []legal.Move{passMove(0), grantMove(t, src, chump)}
	cfg := heuristic.DefaultConfig()
	if got := valueOf(t, cantBlockOn(), v, 1, moves...); got < cfg.LethalBonus {
		t.Errorf("the lethal can't-block is priced %.2f, want at least LethalBonus (%.2f)", got, cfg.LethalBonus)
	}
}

// TestTragicSlipKeepsItsRemovalPrice: a harm with no printed amount and
// no restriction (Tragic Slip's -1/-1 or -13/-13) is priced as removal
// whatever PriceCantBlock says, before combat and after.
func TestTragicSlipKeepsItsRemovalPrice(t *testing.T) {
	slip := trick(cardID(90), "Tragic Slip", "{B}", "harm", nil)
	theirs := cardID(2)
	for _, step := range []string{"precombat_main", "postcombat_main"} {
		v := newView([]protocol.PlayerView{newSeat(0, withHand(slip)), newSeat(1)},
			withBattlefield(land(cardID(91), 0), creature(theirs, 1, "Their Ogre", 4, 4)), withTurn(5, 0, step))
		moves := []legal.Move{passMove(0), castMove(t, 0, slip.InstanceID, "Cast Tragic Slip", cardTarget(theirs))}
		on, off := valueOf(t, cantBlockOn(), v, 1, moves...), valueOf(t, noCantBlock(), v, 1, moves...)
		if on != off || on < 1.0 {
			t.Errorf("%s: Tragic Slip is priced %.2f with PriceCantBlock and %.2f without; want the same removal price", step, on, off)
		}
	}
}

// TestCantBlockShipsOff: the price lost in #4977's arena, so the default
// keeps a can't-block pick at the removal price and BaselineConfig never
// had it.
func TestCantBlockShipsOff(t *testing.T) {
	if heuristic.DefaultConfig().PriceCantBlock || heuristic.BaselineConfig().PriceCantBlock {
		t.Error("PriceCantBlock is on by default; #4977's arena measured it as a loss")
	}
}

// TestDuelTacticsKillIsStillRemoval: Duel Tactics deals 1 damage and
// says "It can't block this turn." On an opponent's 1/1 the damage kills
// it, which is priced as declared damage is, not as the attack the
// restriction opens: after combat it is still worth the kill.
func TestDuelTacticsKillIsStillRemoval(t *testing.T) {
	dt := spell(cardID(90), 0, "Duel Tactics", "{R}")
	dt.TypeLine = "Sorcery"
	dt.Purpose = &protocol.PurposeView{Targets: &[]protocol.TargetPurposeView{
		{Slot: 0, Damage: 1, UntilEndOfTurn: "harm", Pump: &protocol.PumpView{CantBlock: true}},
	}}
	squirrel := cardID(2)
	v := newView([]protocol.PlayerView{newSeat(0, withHand(dt)), newSeat(1)},
		withBattlefield(land(cardID(91), 0), creature(squirrel, 1, "Their Squirrel", 1, 1)), withTurn(5, 0, "postcombat_main"))
	moves := []legal.Move{passMove(0), castMove(t, 0, dt.InstanceID, "Cast Duel Tactics", cardTarget(squirrel))}
	on, off := valueOf(t, cantBlockOn(), v, 1, moves...), valueOf(t, noCantBlock(), v, 1, moves...)
	if on != off || on <= 0 {
		t.Errorf("Duel Tactics on a 1/1 after combat is priced %.2f with PriceCantBlock and %.2f without; want the same kill", on, off)
	}
}
