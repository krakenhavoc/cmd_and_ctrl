package heuristic_test

import (
	"context"
	"strings"
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/heuristic"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// temporary_effect_test.go — #4302: a change to a target that ends with
// the turn, priced by whose permanent it is and what it does this turn.
// The whole-engine version, with the real Heliod, Sun-Crowned, is
// aiseat/heliod_grant_bot_test.go.

// granter is a seat-0 enchantment with one activated row, "{1}{W}: …
// target creature …", whose target entry declares a change until end of
// turn.
func granter(id, label, effect string, pump *protocol.PumpView) protocol.CardView {
	c := protocol.CardView{
		InstanceID: id, Name: "Granter", Owner: seatID(0).String(), Controller: seatID(0).String(),
		TypeLine: "Enchantment", ManaCost: "{2}{W}", KnownByYou: true,
	}
	c.ActivatedAbilities = []protocol.ActivatedAbilityView{{
		Index: 0, Label: label, ManaCost: "{1}{W}",
		Purpose: &protocol.PurposeView{Targets: &[]protocol.TargetPurposeView{{Slot: 0, UntilEndOfTurn: effect, Pump: pump}}},
	}}
	return c
}

func lifelinkGranter(id string) protocol.CardView {
	return granter(id, "{1}{W}: Another target creature gains lifelink until end of turn.", "benefit",
		&protocol.PumpView{Keywords: []string{"lifelink"}})
}

func grantMove(t *testing.T, src, target string) legal.Move {
	return activateMove(t, 0, src, "Granter: grant targeting "+target, &legal.MoveCost{Mana: "{1}{W}"}, cardTarget(target))
}

func twoPlains() []protocol.CardView {
	return []protocol.CardView{land(cardID(91), 0), land(cardID(92), 0)}
}

func decideOn(t *testing.T, pol *heuristic.Policy, v protocol.GameView, moves ...legal.Move) (int, string) {
	t.Helper()
	d, err := pol.Decide(context.Background(), input(0, v, moves...))
	if err != nil {
		t.Fatalf("Decide: %v", err)
	}
	return d.Index, d.Reason
}

// TestTemporaryBenefitIsNeverGivenToAnOpponent: the only creature on
// the board is an opponent's, in the bot's own main phase and in
// combat. Lifelink would be that opponent's gain, so the bot passes,
// and the move is priced below passing. Under the baseline the same
// pick was removal.
func TestTemporaryBenefitIsNeverGivenToAnOpponent(t *testing.T) {
	src, theirs := cardID(1), cardID(2)
	for _, step := range []string{"precombat_main", "declare_blockers", "end"} {
		t.Run(step, func(t *testing.T) {
			board := append(twoPlains(), lifelinkGranter(src), creature(theirs, 1, "Opposing Bear", 3, 3))
			v := newView([]protocol.PlayerView{newSeat(0), newSeat(1)}, withBattlefield(board...), withTurn(3, 0, step))
			moves := []legal.Move{passMove(0), grantMove(t, src, theirs)}
			if i, why := decideOn(t, heuristic.New(), v, moves...); i != 0 {
				t.Errorf("the bot gave an opponent's creature lifelink: %s", why)
			}
			if c := heuristic.New().Rank(context.Background(), input(0, v, moves...)); c[len(c)-1].Index != 1 || c[len(c)-1].Value >= 0 {
				t.Errorf("the grant is priced %+v, want below passing", c)
			}
		})
	}
	t.Run("the baseline prices it as removal", func(t *testing.T) {
		board := append(twoPlains(), lifelinkGranter(src), creature(theirs, 1, "Opposing Bear", 3, 3))
		v := newView([]protocol.PlayerView{newSeat(0), newSeat(1)}, withBattlefield(board...), withTurn(3, 0, "precombat_main"))
		if i, _ := decideOn(t, heuristic.NewWithConfig(heuristic.BaselineConfig()), v, passMove(0), grantMove(t, src, theirs)); i != 1 {
			t.Errorf("the baseline passed; it is meant to keep the old price, an opposing pick removed")
		}
	})
}

// TestTemporaryBenefitGoesOnItsOwnAttacker: in the bot's declare
// blockers step, its own 2/2 attacking, an opponent's untapped 3/3 and
// a creature of its own at home. The grant goes on the attacker.
func TestTemporaryBenefitGoesOnItsOwnAttacker(t *testing.T) {
	src, theirs, attacker, home := cardID(1), cardID(2), cardID(3), cardID(4)
	board := append(twoPlains(), lifelinkGranter(src),
		creature(theirs, 1, "Opposing Bear", 3, 3),
		creature(attacker, 0, "Attacker", 2, 2, attacking(1)),
		creature(home, 0, "Home Guard", 4, 4))
	v := newView([]protocol.PlayerView{newSeat(0), newSeat(1)}, withBattlefield(board...), withTurn(3, 0, "declare_blockers"))
	moves := []legal.Move{passMove(0), grantMove(t, src, theirs), grantMove(t, src, home), grantMove(t, src, attacker)}
	i, why := decideOn(t, heuristic.New(), v, moves...)
	if i != 3 {
		t.Fatalf("picked %d (%s), want the grant on its own attacker", i, why)
	}
	if !strings.Contains(why, "spare mana in combat") {
		t.Errorf("reason %q, want the spare-mana bar", why)
	}
}

// TestTemporaryBenefitWaitsWhenItsManaIsNeeded: the same combat with a
// {1}{W} creature in hand and two lands. The grant would spend the mana
// the second main phase casts with, so it does not clear the response
// bar, and the bot passes.
func TestTemporaryBenefitWaitsWhenItsManaIsNeeded(t *testing.T) {
	src, attacker := cardID(1), cardID(3)
	cat := creature(cardID(5), 0, "Test Lion", 2, 2)
	cat.ManaCost = "{1}{W}"
	board := append(twoPlains(), lifelinkGranter(src), creature(attacker, 0, "Attacker", 2, 2, attacking(1)))
	v := newView([]protocol.PlayerView{newSeat(0, withHand(cat)), newSeat(1)}, withBattlefield(board...), withTurn(3, 0, "declare_blockers"))
	if i, why := decideOn(t, heuristic.New(), v, passMove(0), grantMove(t, src, attacker)); i != 0 {
		t.Errorf("the bot spent its main-phase mana in combat: %s", why)
	}
}

// TestTemporaryBenefitLosesToACast: the bot's own main phase, two
// lands, a {1}{W} creature to cast and a creature of its own the grant
// could go on. Lifelink on a creature that is not about to deal damage
// is worth nothing, and the cast develops the board.
func TestTemporaryBenefitLosesToACast(t *testing.T) {
	src, mine := cardID(1), cardID(3)
	cat := creature(cardID(5), 0, "Test Lion", 2, 2)
	cat.ManaCost = "{1}{W}"
	board := append(twoPlains(), lifelinkGranter(src), creature(mine, 0, "Mine", 2, 2, sick()))
	v := newView([]protocol.PlayerView{newSeat(0, withHand(cat)), newSeat(1)}, withBattlefield(board...), withTurn(3, 0, "precombat_main"))
	moves := []legal.Move{passMove(0), grantMove(t, src, mine), castMove(t, 0, cat.InstanceID, "Cast Test Lion")}
	if i, why := decideOn(t, heuristic.New(), v, moves...); i != 2 {
		t.Errorf("picked %d (%s), want the cast", i, why)
	}
	// And with nothing to cast it is not worth the mana either.
	v = newView([]protocol.PlayerView{newSeat(0), newSeat(1)}, withBattlefield(board...), withTurn(3, 0, "precombat_main"))
	if i, why := decideOn(t, heuristic.New(), v, moves[:2]...); i != 0 {
		t.Errorf("the bot granted lifelink to a creature that deals no damage this turn: %s", why)
	}
}

// TestTemporaryBenefitAddsNothingTwice: the attacker already has
// lifelink (CR 702.15f), so a second grant is below passing whatever
// the mana.
func TestTemporaryBenefitAddsNothingTwice(t *testing.T) {
	src, attacker := cardID(1), cardID(3)
	board := append(twoPlains(), lifelinkGranter(src), creature(attacker, 0, "Attacker", 2, 2, attacking(1), keywords("lifelink")))
	v := newView([]protocol.PlayerView{newSeat(0), newSeat(1)}, withBattlefield(board...), withTurn(3, 0, "declare_blockers"))
	if i, why := decideOn(t, heuristic.New(), v, passMove(0), grantMove(t, src, attacker)); i != 0 {
		t.Errorf("the bot granted lifelink to a creature that has it: %s", why)
	}
}

// TestNoAmountGrantIsMadeOncePerCreature: "can't be blocked this turn"
// shows nothing on the creature afterwards, so the policy remembers the
// grant it made and does not buy it again for the same attacker.
func TestNoAmountGrantIsMadeOncePerCreature(t *testing.T) {
	src, attacker := cardID(1), cardID(3)
	g := granter(src, "{1}{U}: Target creature can't be blocked this turn.", "benefit", nil)
	board := append(twoPlains(), land(cardID(93), 0), land(cardID(94), 0), g, creature(attacker, 0, "Attacker", 2, 2, attacking(1)))
	v := newView([]protocol.PlayerView{newSeat(0), newSeat(1)}, withBattlefield(board...), withTurn(3, 0, "declare_attackers"))
	pol := heuristic.New()
	if i, why := decideOn(t, pol, v, passMove(0), grantMove(t, src, attacker)); i != 1 {
		t.Fatalf("picked %d (%s), want the grant on its attacker", i, why)
	}
	if i, why := decideOn(t, pol, v, passMove(0), grantMove(t, src, attacker)); i != 0 {
		t.Errorf("the bot bought the same grant twice: %s", why)
	}
	// A new turn forgets it.
	v.Turn.Seq++
	if i, why := decideOn(t, pol, v, passMove(0), grantMove(t, src, attacker)); i != 1 {
		t.Errorf("picked %d (%s) on a new turn, want the grant again", i, why)
	}
}

// TestTemporaryHarmIsNeverAimedAtItsOwnCreature: "-1/-1 until end of
// turn". On an opponent's 1/1 it kills (CR 704.5f) and is removal; on
// the bot's own it is a creature lost; on an opponent's 3/3 out of
// combat it does nothing that lasts.
func TestTemporaryHarmIsNeverAimedAtItsOwnCreature(t *testing.T) {
	src, mine, small, big := cardID(1), cardID(3), cardID(4), cardID(5)
	g := granter(src, "{1}{W}: Target creature gets -1/-1 until end of turn.", "harm", &protocol.PumpView{Power: -1, Toughness: -1})
	board := append(twoPlains(), g, creature(mine, 0, "Mine", 1, 1),
		creature(small, 1, "Their Thopter", 1, 1), creature(big, 1, "Their Bear", 3, 3))
	v := newView([]protocol.PlayerView{newSeat(0), newSeat(1)}, withBattlefield(board...), withTurn(3, 0, "precombat_main"))
	if i, why := decideOn(t, heuristic.New(), v, passMove(0), grantMove(t, src, mine), grantMove(t, src, small)); i != 2 {
		t.Errorf("picked %d (%s), want the opponent's 1/1", i, why)
	}
	if i, why := decideOn(t, heuristic.New(), v, passMove(0), grantMove(t, src, mine)); i != 0 {
		t.Errorf("the bot shrank its own creature: %s", why)
	}
	if i, why := decideOn(t, heuristic.New(), v, passMove(0), grantMove(t, src, big)); i != 0 {
		t.Errorf("the bot spent its mana on -1/-1 to a 3/3 out of combat: %s", why)
	}
}

// TestHasteGrantLetsASickCreatureAttack: haste before combat on the
// bot's summoning-sick 3/3 into an open board is an attack it would not
// otherwise have; after combat it is nothing.
func TestHasteGrantLetsASickCreatureAttack(t *testing.T) {
	src, mine := cardID(1), cardID(3)
	g := granter(src, "{1}{W}: Target creature gains haste until end of turn.", "benefit", &protocol.PumpView{Keywords: []string{"haste"}})
	board := append(twoPlains(), g, creature(mine, 0, "Fresh Ogre", 3, 3, sick()))
	for step, want := range map[string]int{"precombat_main": 1, "postcombat_main": 0} {
		v := newView([]protocol.PlayerView{newSeat(0), newSeat(1)}, withBattlefield(board...), withTurn(3, 0, step))
		if i, why := decideOn(t, heuristic.New(), v, passMove(0), grantMove(t, src, mine)); i != want {
			t.Errorf("%s: picked %d (%s), want %d", step, i, why, want)
		}
	}
}
