package heuristic_test

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/heuristic"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// net_removal_test.go is #2679: removal priced net of what its target's
// controller gets back. Each test fails with NetRemoval off.

func withoutNetRemoval() *heuristic.Policy {
	cfg := heuristic.DefaultConfig()
	cfg.NetRemoval = false
	return heuristic.NewWithConfig(cfg)
}

// removalReturning is an instant whose one target clause declares what
// its controller is given back.
func removalReturning(id, name, cost string, r protocol.TargetReturnView) protocol.CardView {
	c := spell(id, 0, name, cost)
	c.Purpose = targetEntries(protocol.TargetPurposeView{Slot: 0, Returns: &r})
	return c
}

func threeThree() protocol.TargetReturnView {
	return protocol.TargetReturnView{CreatureTokens: 1, TokenPower: 3, TokenToughness: 3}
}

// Review game 1, seq 66: Rapid Hybridization on a 1/1 hands its
// controller a 3/3. It is held, and still cast on a creature worth more
// than the 3/3.
func TestRemovalThatHandsBackABiggerBodyIsHeld(t *testing.T) {
	hybrid, small, big := cardID(1), cardID(2), cardID(3)
	v := newView([]protocol.PlayerView{
		newSeat(0, withHand(removalReturning(hybrid, "Rapid Hybridization", "{U}", threeThree()))),
		newSeat(1),
	}, withBattlefield(append(manaLands(2, 0, 100),
		creature(small, 1, "Archivist", 1, 1),
		creature(big, 1, "Dragon", 6, 6, keywords("flying")))...),
		withTurn(4, 0, "precombat_main"))
	atSmall := castMove(t, 0, hybrid, "Hybridize the 1/1", cardTarget(small))
	atBig := castMove(t, 0, hybrid, "Hybridize the 6/6", cardTarget(big))
	pass := passMove(0)

	on, off := heuristic.New(), withoutNetRemoval()
	cfg := on.Config()
	token := cfg.Weights.CreatureValue(&protocol.CardView{Power: 3, Toughness: 3, IsToken: true})
	gift := token * cfg.RemovalConfidence * cfg.LeaderBoost

	in := input(0, v, pass, atSmall)
	if got, want := rankValue(t, off, in, atSmall.Label)-rankValue(t, on, in, atSmall.Label), gift; !nearly(got, want) {
		t.Errorf("the 3/3 costs the cast %.3f, want %.3f (its body × RemovalConfidence × LeaderBoost)", got, want)
	}
	if d := decide(t, on, in); in.Moves[d.Index].Label != pass.Label {
		t.Errorf("chose %q, want to hold Rapid Hybridization rather than trade a 1/1 for a 3/3", in.Moves[d.Index].Label)
	}
	if d := decide(t, off, in); in.Moves[d.Index].Label != atSmall.Label {
		t.Errorf("with NetRemoval off chose %q, want the old cast on the 1/1", in.Moves[d.Index].Label)
	}

	in = input(0, v, pass, atSmall, atBig)
	if d := decide(t, on, in); in.Moves[d.Index].Label != atBig.Label {
		t.Errorf("chose %q, want Rapid Hybridization on the 6/6 flier", in.Moves[d.Index].Label)
	}
}

// Swords to Plowshares gives life equal to the target's power, and Path
// to Exile a land, each taken off the removal on its own scale.
func TestRemovalReturnsLifeAndLand(t *testing.T) {
	swords, path, bear := cardID(1), cardID(2), cardID(3)
	v := newView([]protocol.PlayerView{
		newSeat(0, withHand(
			removalReturning(swords, "Swords to Plowshares", "{W}", protocol.TargetReturnView{LifeEqualToPower: true}),
			removalReturning(path, "Path to Exile", "{W}", protocol.TargetReturnView{Lands: 1}),
		)),
		newSeat(1),
	}, withBattlefield(append(manaLands(2, 0, 100), creature(bear, 1, "Bear", 4, 2))...),
		withTurn(4, 0, "precombat_main"))
	atSwords := castMove(t, 0, swords, "Swords the bear", cardTarget(bear))
	atPath := castMove(t, 0, path, "Path the bear", cardTarget(bear))
	in := input(0, v, passMove(0), atSwords, atPath)

	on, off := heuristic.New(), withoutNetRemoval()
	cfg := on.Config()
	scale := cfg.RemovalConfidence * cfg.LeaderBoost
	if got, want := rankValue(t, off, in, atSwords.Label)-rankValue(t, on, in, atSwords.Label), 4*cfg.Weights.Life*scale; !nearly(got, want) {
		t.Errorf("4 life back costs Swords %.3f, want %.3f", got, want)
	}
	if got, want := rankValue(t, off, in, atPath.Label)-rankValue(t, on, in, atPath.Label), cfg.Weights.ManaSource*scale; !nearly(got, want) {
		t.Errorf("a land back costs Path %.3f, want %.3f", got, want)
	}
	if rankValue(t, on, in, atSwords.Label) <= rankValue(t, on, in, atPath.Label) {
		t.Error("Swords (4 life back) should beat Path (a land back) on the same 4/2")
	}
}

// A removal the bot aims at its own permanent nets nothing: the
// controller who gets the gift is the bot, and the pick keeps its old
// price.
func TestReturnsAreNotNettedOnTheBotsOwnPermanent(t *testing.T) {
	within, mine := cardID(1), cardID(2)
	v := newView([]protocol.PlayerView{
		newSeat(0, withHand(removalReturning(within, "Beast Within", "{2}{G}", threeThree()))),
		newSeat(1),
	}, withBattlefield(append(manaLands(3, 0, 100), creature(mine, 0, "Elf", 1, 1))...),
		withTurn(4, 0, "precombat_main"))
	own := castMove(t, 0, within, "Beast Within my elf", cardTarget(mine))
	in := input(0, v, passMove(0), own)
	if on, off := rankValue(t, heuristic.New(), in, own.Label), rankValue(t, withoutNetRemoval(), in, own.Label); !nearly(on, off) {
		t.Errorf("Beast Within on the bot's own elf priced %.3f, want the old %.3f", on, off)
	}
}

// Review game 2, seq 186: an opposing commander comes back from the
// command zone (CR 903.9a), so removing it takes the tax and a turn of
// the body, not the body. With a commander of its own to cast, the bot
// develops instead.
func TestRemovalOnACommanderIsPricedNetOfItsReturn(t *testing.T) {
	warp, theirs, twin, mine := cardID(1), cardID(2), cardID(3), cardID(4)
	theirCommander := creature(theirs, 1, "Y'shtola", 2, 4, keywords("flying", "lifelink"), commander())
	sameBody := creature(twin, 1, "Twin", 2, 4, keywords("flying", "lifelink"))
	myCommander := creature(mine, 0, "Mary Read", 3, 3, commander())
	myCommander.ManaCost = "{1}{U}{R}"
	v := newView([]protocol.PlayerView{
		newSeat(0, withHand(spell(warp, 0, "Chaos Warp", "{2}{R}"), myCommander)),
		newSeat(1),
	}, withBattlefield(append(manaLands(3, 0, 100), theirCommander, sameBody)...),
		withTurn(5, 0, "precombat_main"))
	atCommander := castMove(t, 0, warp, "Warp their commander", cardTarget(theirs))
	atTwin := castMove(t, 0, warp, "Warp the twin", cardTarget(twin))
	develop := castMove(t, 0, mine, "Cast my commander")

	on, off := heuristic.New(), withoutNetRemoval()
	cfg := on.Config()
	in := input(0, v, passMove(0), atCommander, atTwin, develop)

	if a, b := rankValue(t, off, in, atCommander.Label), rankValue(t, off, in, atTwin.Label); !nearly(a, b) {
		t.Fatalf("with NetRemoval off the commander %.3f and its twin %.3f should price the same", a, b)
	}
	back := 2*cfg.Weights.CommanderTax + cfg.DamageToOpponent*2
	full := cfg.Weights.CreatureValue(&theirCommander)
	if back >= full {
		t.Fatalf("fixture: the commander's return %.3f is not below its body %.3f", back, full)
	}
	want := rankValue(t, off, in, atCommander.Label) - (full-back)*cfg.RemovalConfidence*cfg.LeaderBoost
	if got := rankValue(t, on, in, atCommander.Label); !nearly(got, want) {
		t.Errorf("removal on their commander priced %.3f, want %.3f (tax and a turn of the body, not the body)", got, want)
	}
	if got := rankValue(t, on, in, atTwin.Label); !nearly(got, rankValue(t, off, in, atTwin.Label)) {
		t.Error("removal on a creature that is not a commander should keep its price")
	}
	in = input(0, v, passMove(0), atCommander, develop)
	if d := decide(t, on, in); in.Moves[d.Index].Label != develop.Label {
		t.Errorf("chose %q, want the bot's own commander over removing theirs", in.Moves[d.Index].Label)
	}
	if d := decide(t, off, in); in.Moves[d.Index].Label != atCommander.Label {
		t.Errorf("with NetRemoval off chose %q, want the old removal on their commander", in.Moves[d.Index].Label)
	}
}

// A commander the bot owns that an opponent controls comes back to the
// bot, so removing it keeps its full price; and declared damage that
// kills an opposing commander is netted like any other removal.
func TestCommanderNetIsForTheirCommandersAndCoversLethalDamage(t *testing.T) {
	shock, stolen, theirs := cardID(1), cardID(2), cardID(3)
	mineStolen := creature(stolen, 1, "My Commander", 2, 2, keywords("flying"), commander())
	mineStolen.Owner = seatID(0).String()
	theirCommander := creature(theirs, 1, "Their Commander", 2, 2, keywords("flying"), commander())
	v := newView([]protocol.PlayerView{
		newSeat(0, withHand(spell(shock, 0, "Shock", "{R}"))),
		newSeat(1),
	}, withBattlefield(append(manaLands(1, 0, 100), mineStolen, theirCommander)...),
		withTurn(5, 0, "precombat_main"))
	v.Seats[0].Hand.Cards[0].Purpose = targetEntries(protocol.TargetPurposeView{Slot: 0, Damage: 2})
	atStolen := castMove(t, 0, shock, "Shock my stolen commander", cardTarget(stolen))
	atTheirs := castMove(t, 0, shock, "Shock their commander", cardTarget(theirs))
	in := input(0, v, passMove(0), atStolen, atTheirs)

	on, off := heuristic.New(), withoutNetRemoval()
	if a, b := rankValue(t, on, in, atStolen.Label), rankValue(t, off, in, atStolen.Label); !nearly(a, b) {
		t.Errorf("Shock at the bot's own commander under an opponent's control priced %.3f, want the old %.3f", a, b)
	}
	if a, b := rankValue(t, on, in, atTheirs.Label), rankValue(t, off, in, atTheirs.Label); a >= b {
		t.Errorf("Shock that kills their commander priced %.3f, want below the old %.3f", a, b)
	}
}
