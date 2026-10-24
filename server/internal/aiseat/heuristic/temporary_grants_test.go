package heuristic_test

import (
	"context"
	"strings"
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/heuristic"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// temporary_grants_test.go — #4312, the follow-up to #4302: a spell's
// change until end of turn priced as an activated row's is, a grant aimed
// at the bot's own permanent priced as the save it is when the stack would
// remove that permanent, and evasion priced before attacks. The
// whole-engine version of the save, with the real Mother of Runes and
// Murder, is aiseat/mother_of_runes_save_bot_test.go.

// trick is a seat-0 instant whose one target entry declares a change
// until end of turn.
func trick(id, name, cost, effect string, pump *protocol.PumpView) protocol.CardView {
	c := spell(id, 0, name, cost)
	c.Purpose = &protocol.PurposeView{Targets: &[]protocol.TargetPurposeView{{Slot: 0, UntilEndOfTurn: effect, Pump: pump}}}
	return c
}

func giantGrowth(id string) protocol.CardView {
	return trick(id, "Giant Growth", "{G}", "benefit", &protocol.PumpView{Power: 3, Toughness: 3})
}

func noTemporary() *heuristic.Policy {
	return withConfig(func(c *heuristic.Config) { c.PriceTemporaryEffects = false })
}

// TestPumpSpellIsNeverCastAtAnOpponent: Giant Growth in hand, the only
// creature on the board an opponent's 3/3, in the bot's main phase and in
// its combat. "+3/+3" is that opponent's gain, so the bot keeps the card.
// Before #4312 the spell declared nothing and the pick was priced as
// removal: heuristic-notemporary still casts it there.
func TestPumpSpellIsNeverCastAtAnOpponent(t *testing.T) {
	gg, theirs := giantGrowth(cardID(90)), cardID(2)
	for _, step := range []string{"precombat_main", "declare_blockers"} {
		t.Run(step, func(t *testing.T) {
			v := newView([]protocol.PlayerView{newSeat(0, withHand(gg)), newSeat(1)},
				withBattlefield(land(cardID(91), 0), creature(theirs, 1, "Opposing Bear", 3, 3)), withTurn(3, 0, step))
			moves := []legal.Move{passMove(0), castMove(t, 0, gg.InstanceID, "Cast Giant Growth", cardTarget(theirs))}
			if i, why := decideOn(t, heuristic.New(), v, moves...); i != 0 {
				t.Errorf("the bot cast Giant Growth on an opponent's creature: %s", why)
			}
			if c := heuristic.New().Rank(context.Background(), input(0, v, moves...)); c[len(c)-1].Index != 1 || c[len(c)-1].Value >= 0 {
				t.Errorf("the cast is priced %+v, want below passing", c)
			}
		})
	}
	t.Run("without the pricing it is removal", func(t *testing.T) {
		v := newView([]protocol.PlayerView{newSeat(0, withHand(gg)), newSeat(1)},
			withBattlefield(land(cardID(91), 0), creature(theirs, 1, "Opposing Bear", 3, 3)), withTurn(3, 0, "precombat_main"))
		if i, _ := decideOn(t, noTemporary(), v, passMove(0), castMove(t, 0, gg.InstanceID, "Cast Giant Growth", cardTarget(theirs))); i != 1 {
			t.Errorf("heuristic-notemporary passed; it is meant to keep the old price, an opposing pick removed")
		}
	})
}

// TestPumpSpellGoesOnTheAttackThatKills: the bot's 2/2 attacks an open
// board and the defender is on 4. Giant Growth on the attacker is the
// kill, and is cast in the declare attackers step.
func TestPumpSpellGoesOnTheAttackThatKills(t *testing.T) {
	gg, mine, theirs := giantGrowth(cardID(90)), cardID(3), cardID(2)
	v := newView([]protocol.PlayerView{newSeat(0, withHand(gg)), newSeat(1, withLife(4))},
		withBattlefield(land(cardID(91), 0), creature(mine, 0, "Attacker", 2, 2, attacking(1)),
			creature(theirs, 1, "Tapped Bear", 3, 3, tapped())),
		withTurn(5, 0, "declare_attackers"))
	moves := []legal.Move{passMove(0),
		castMove(t, 0, gg.InstanceID, "Cast Giant Growth at theirs", cardTarget(theirs)),
		castMove(t, 0, gg.InstanceID, "Cast Giant Growth at mine", cardTarget(mine))}
	if i, why := decideOn(t, heuristic.New(), v, moves...); i != 2 {
		t.Errorf("picked %d (%s), want Giant Growth on the attacker that kills", i, why)
	}
}

// TestShrinkSpellIsRemovalOnlyWhenItKills: Dismember's -5/-5. On an
// opponent's 4/4 it kills (CR 704.5f) and is removal; on an opponent's
// 7/7 outside combat it does nothing that lasts; on the bot's own
// creature it is a creature lost.
func TestShrinkSpellIsRemovalOnlyWhenItKills(t *testing.T) {
	dm := trick(cardID(90), "Dismember", "{1}{B}{B}", "harm", &protocol.PumpView{Power: -5, Toughness: -5})
	small, big, mine := cardID(2), cardID(4), cardID(3)
	board := withBattlefield(land(cardID(91), 0), land(cardID(92), 0), land(cardID(93), 0),
		creature(small, 1, "Their Ogre", 4, 4), creature(big, 1, "Their Wurm", 7, 7), creature(mine, 0, "Mine", 2, 2))
	v := newView([]protocol.PlayerView{newSeat(0, withHand(dm)), newSeat(1)}, board, withTurn(3, 0, "precombat_main"))
	cast := func(target string) legal.Move {
		return castMove(t, 0, dm.InstanceID, "Cast Dismember", cardTarget(target))
	}
	if i, why := decideOn(t, heuristic.New(), v, passMove(0), cast(big), cast(small), cast(mine)); i != 2 {
		t.Errorf("picked %d (%s), want Dismember on the 4/4 it kills", i, why)
	}
	if i, why := decideOn(t, heuristic.New(), v, passMove(0), cast(big)); i != 0 {
		t.Errorf("the bot cast -5/-5 at a 7/7 it does not kill: %s", why)
	}
	if i, why := decideOn(t, heuristic.New(), v, passMove(0), cast(mine)); i != 0 {
		t.Errorf("the bot shrank its own creature: %s", why)
	}
}

// --- protection as a save ------------------------------------------------

// motherOfRunes is a seat-0 Mother of Runes, untapped and able to tap:
// "{T}: Target creature you control gains protection from the color of
// your choice until end of turn."
func motherOfRunes(id string) protocol.CardView {
	c := creature(id, 0, "Mother of Runes", 1, 1)
	c.ActivatedAbilities = []protocol.ActivatedAbilityView{{
		Index: 0, TapCost: true,
		Label: "{T}: Target creature you control gains protection from the color of your choice until end of turn.",
		Purpose: &protocol.PurposeView{Targets: &[]protocol.TargetPurposeView{{
			Slot: 0, UntilEndOfTurn: "benefit", Pump: &protocol.PumpView{Keywords: []string{"protection"}},
		}}},
	}}
	return c
}

func motherMove(t *testing.T, src, target string) legal.Move {
	return activateMove(t, 0, src, "Mother of Runes: protect "+target, nil, cardTarget(target))
}

// murderOn is an opponent's black Murder on the stack aimed at target.
func murderOn(target string) viewOpt {
	c, it := removal(1, cardID(80), target)
	c = colored(c, "B")
	return func(v *protocol.GameView) {
		withStack(c)(v)
		withStackItems(it)(v)
	}
}

// TestProtectionGrantSavesFromTargetedRemoval: an opponent's Murder
// targets the bot's 4/4 on the opponent's turn. Mother of Runes on it is
// the save, worth about the 4/4, and the bot makes it. With no Murder, or
// with the save pricing off, it is worth nothing and the bot passes.
func TestProtectionGrantSavesFromTargetedRemoval(t *testing.T) {
	mom, ogre := cardID(1), cardID(3)
	board := withBattlefield(motherOfRunes(mom), creature(ogre, 0, "Ogre", 4, 4), creature(cardID(2), 1, "Their Bear", 2, 2))
	moves := []legal.Move{passMove(0), motherMove(t, mom, ogre)}

	v := newView([]protocol.PlayerView{newSeat(0), newSeat(1)}, board, withTurn(4, 1, "precombat_main"), murderOn(ogre))
	if i, why := decideOn(t, heuristic.New(), v, moves...); i != 1 {
		t.Fatalf("picked %d (%s), want Mother of Runes on the Murder's target", i, why)
	}
	ranked := heuristic.New().Rank(context.Background(), input(0, v, moves...))
	if ranked[0].Index != 1 || ranked[0].Value < 2.0 {
		t.Errorf("the save is priced %+v, want about the 4/4's value", ranked)
	}

	// A 1/1's save is worth less than the response bar, and the tap is
	// one the bot gets back in its untap step: the spare-mana bar.
	soldier := cardID(6)
	small := newView([]protocol.PlayerView{newSeat(0), newSeat(1)},
		withBattlefield(motherOfRunes(mom), creature(soldier, 0, "Soldier", 1, 1)), withTurn(4, 1, "precombat_main"), murderOn(soldier))
	if i, why := decideOn(t, heuristic.New(), small, passMove(0), motherMove(t, mom, soldier)); i != 1 || !strings.Contains(why, "spare mana for a save") {
		t.Errorf("picked %d (%s), want Mother of Runes on the Soldier past the spare-mana bar", i, why)
	}

	quiet := newView([]protocol.PlayerView{newSeat(0), newSeat(1)}, board, withTurn(4, 1, "precombat_main"))
	if i, why := decideOn(t, heuristic.New(), quiet, moves...); i != 0 {
		t.Errorf("with nothing on the stack the bot protected its creature anyway: %s", why)
	}
	off := withConfig(func(c *heuristic.Config) { c.PriceGrantSaves = false })
	if i, why := decideOn(t, off, v, moves...); i != 0 {
		t.Errorf("with PriceGrantSaves off the grant is meant to buy nothing out of combat: %s", why)
	}
}

// TestProtectionGrantDoesNotStackASecondSave: the bot's own Mother of
// Runes activation is already on the stack above the Murder, aimed at the
// same creature. A second grant saves nothing more.
func TestProtectionGrantDoesNotStackASecondSave(t *testing.T) {
	mom, mom2, ogre := cardID(1), cardID(5), cardID(3)
	board := withBattlefield(motherOfRunes(mom), motherOfRunes(mom2), creature(ogre, 0, "Ogre", 4, 4))
	mine := protocol.StackItemView{ID: "item-mom", Kind: "ability", Controller: seatID(0).String(), SourceCardID: mom,
		Targets: []protocol.TargetRefView{{Kind: "card", ID: ogre}}}
	v := newView([]protocol.PlayerView{newSeat(0), newSeat(1)}, board, withTurn(4, 1, "precombat_main"),
		murderOn(ogre), withStackItems(mine))
	if i, why := decideOn(t, heuristic.New(), v, passMove(0), motherMove(t, mom2, ogre)); i != 0 {
		t.Errorf("the bot stacked a second save on the same creature: %s", why)
	}
}

// TestHexproofSpellIsNoAnswerToASweep: Blossoming Defense's hexproof
// against a declared Wrath of God. The Wrath targets nothing, so hexproof
// keeps nothing (CR 702.11b), and the card stays in hand. Against a
// targeted Murder it is the save.
func TestHexproofSpellIsNoAnswerToASweep(t *testing.T) {
	bd := trick(cardID(90), "Blossoming Defense", "{G}", "benefit", &protocol.PumpView{Power: 2, Toughness: 2, Keywords: []string{"hexproof"}})
	ogre := cardID(3)
	board := withBattlefield(land(cardID(91), 0), creature(ogre, 0, "Ogre", 4, 4))
	cast := castMove(t, 0, bd.InstanceID, "Cast Blossoming Defense", cardTarget(ogre))

	wrath := spell(cardID(81), 1, "Wrath of God", "{2}{W}{W}")
	wrath.TypeLine = "Sorcery"
	wrath.Purpose = &protocol.PurposeView{Sweep: &protocol.SweepView{Matches: "creatures", How: "destroy"}}
	sweep := newView([]protocol.PlayerView{newSeat(0, withHand(bd)), newSeat(1)}, board, withTurn(4, 1, "precombat_main"),
		withStack(wrath), withStackItems(protocol.StackItemView{ID: "item-wrath", Kind: "spell", Controller: seatID(1).String(), SourceCardID: wrath.InstanceID}))
	if i, why := decideOn(t, heuristic.New(), sweep, passMove(0), cast); i != 0 {
		t.Errorf("the bot answered a Wrath with hexproof: %s", why)
	}

	targeted := newView([]protocol.PlayerView{newSeat(0, withHand(bd)), newSeat(1)}, board, withTurn(4, 1, "precombat_main"), murderOn(ogre))
	if i, why := decideOn(t, heuristic.New(), targeted, passMove(0), cast); i != 1 {
		t.Errorf("picked %d (%s), want Blossoming Defense in answer to the Murder", i, why)
	}
}

// TestProtectionNamesTheColourOfTheRemoval: Mother of Runes resolves with
// the opponent's black Murder still aimed at the bot's creature and a red
// 6/6 of theirs attacking. The colour to name is black, the Murder's: the
// 6/6 is not what the grant was bought against.
func TestProtectionNamesTheColourOfTheRemoval(t *testing.T) {
	ogre := cardID(3)
	v := newView([]protocol.PlayerView{newSeat(0), newSeat(1)},
		withBattlefield(creature(ogre, 0, "Ogre", 4, 4), colored(creature(cardID(4), 1, "Their Dragon", 6, 6, attacking(0)), "R")),
		withTurn(4, 1, "declare_attackers"), murderOn(ogre), withColorChoice(0, "protect"))
	in := input(0, v, colorMoves(t, 0)...)
	if got := chose(t, in, decide(t, heuristic.New(), in)); got != "choose B" {
		t.Errorf("the bot named %q, want the Murder's black", got)
	}
}

// --- evasion before attacks ------------------------------------------------

func rogueGranter(id string) protocol.CardView {
	return granter(id, "{4}, {T}: Target creature can't be blocked this turn", "benefit", &protocol.PumpView{CantBeBlocked: true})
}

// TestUnblockableGrantIsWorthItsAttackBeforeCombat: the bot's 3/3 can
// attack, and the opponent's untapped 4/4 would block it. "Can't be
// blocked this turn" before attacks is three damage through, so the bot
// makes the grant in its first main phase. With the evasion pricing off,
// or with nothing that could block, it is worth nothing.
func TestUnblockableGrantIsWorthItsAttackBeforeCombat(t *testing.T) {
	src, mine := cardID(1), cardID(3)
	walled := withBattlefield(append(twoPlains(), rogueGranter(src), creature(mine, 0, "Bear", 3, 3),
		creature(cardID(2), 1, "Their Wall", 4, 4))...)
	moves := []legal.Move{passMove(0), grantMove(t, src, mine)}
	v := newView([]protocol.PlayerView{newSeat(0), newSeat(1)}, walled, withTurn(3, 0, "precombat_main"))
	if i, why := decideOn(t, heuristic.New(), v, moves...); i != 1 {
		t.Errorf("picked %d (%s), want the grant on the creature about to attack", i, why)
	}
	off := withConfig(func(c *heuristic.Config) { c.PriceEvasionGrants = false })
	if i, why := decideOn(t, off, v, moves...); i != 0 {
		t.Errorf("with PriceEvasionGrants off the grant is meant to be worth nothing before combat: %s", why)
	}
	open := newView([]protocol.PlayerView{newSeat(0), newSeat(1)},
		withBattlefield(append(twoPlains(), rogueGranter(src), creature(mine, 0, "Bear", 3, 3))...), withTurn(3, 0, "precombat_main"))
	if i, why := decideOn(t, heuristic.New(), open, moves...); i != 0 {
		t.Errorf("the bot made a creature unblockable with nothing to block it: %s", why)
	}
	// After blockers are declared it changes nothing (CR 509.1h).
	blocked := newView([]protocol.PlayerView{newSeat(0), newSeat(1)},
		withBattlefield(append(twoPlains(), rogueGranter(src), creature(mine, 0, "Bear", 3, 3, attacking(1)),
			creature(cardID(2), 1, "Their Wall", 4, 4))...), withTurn(3, 0, "declare_blockers"))
	if c := heuristic.New().Rank(context.Background(), input(0, blocked, moves...)); c[0].Index == 1 && c[0].Value > 0.6 {
		t.Errorf("after blockers the grant is priced %+v, want no more than CombatGrant", c)
	}
}

// TestTrampleGrantIsWorthTheDamageOverAChump: the bot's 6/6 can attack and
// the opponent, on 8 life, would chump it with a 1/1. Trample before
// attacks is the five that still get through (CR 702.19b).
func TestTrampleGrantIsWorthTheDamageOverAChump(t *testing.T) {
	src, mine := cardID(1), cardID(3)
	g := granter(src, "{1}{G}: Target creature gains trample until end of turn.", "benefit", &protocol.PumpView{Keywords: []string{"trample"}})
	board := withBattlefield(append(twoPlains(), g, creature(mine, 0, "Wurm", 6, 6), creature(cardID(2), 1, "Their Squirrel", 1, 1))...)
	v := newView([]protocol.PlayerView{newSeat(0), newSeat(1, withLife(8))}, board, withTurn(3, 0, "precombat_main"))
	moves := []legal.Move{passMove(0), grantMove(t, src, mine)}
	on := heuristic.New().Rank(context.Background(), input(0, v, moves...))
	offPol := withConfig(func(c *heuristic.Config) { c.PriceEvasionGrants = false })
	off := offPol.Rank(context.Background(), input(0, v, moves...))
	value := func(c []heuristic.Candidate) float64 {
		for _, x := range c {
			if x.Index == 1 {
				return x.Value
			}
		}
		return 0
	}
	if value(on) <= value(off)+1.0 {
		t.Errorf("trample before attacks is priced %.2f, and %.2f with the evasion pricing off; want the five damage over the chump between them", value(on), value(off))
	}
	if i, why := decideOn(t, heuristic.New(), v, moves...); i != 1 {
		t.Errorf("picked %d (%s), want the trample grant before attacks", i, why)
	}
}
