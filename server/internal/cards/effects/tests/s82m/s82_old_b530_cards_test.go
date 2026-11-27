package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b530_cards_test.go — Shadowmoor, batch 530 (S82, #5591): damage.
// Each test plays the real catalog entry through the engine.

const (
	oldB530WurmOracle      = "37b7f044-f57b-4d13-82be-46fd01f1dba6"
	oldB530MentorOracle    = "380a1fbf-68bc-4456-9a67-f01c0174944c"
	oldB530TowerOracle     = "6f5d31d1-3493-4d1b-bbf9-1726e47507db"
	oldB530CanopyOracle    = "75f381a6-49f9-4caf-a96a-91c75640b9b0"
	oldB530InvocationOrcl  = "a45fd5cc-0085-4d4f-872b-b9251e8706cf"
	oldB530JawsOracle      = "af6a8149-a849-4e6a-bf9d-e6b4bb9b1a04"
	oldB530EmberGaleOracle = "d93633d6-8433-43c8-9224-a25b411bfd96"
	oldB530CrematorOracle  = "f8ea46e2-84ca-4dcb-b675-a2374bc967bc"
	oldB530GhastlordOracle = "fce163aa-c5d0-4147-8566-512321c0be8e"
)

func oldB530Table(t *testing.T) (*game.Game, *game.Player, *game.Player) {
	t.Helper()
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	return g, me, opp
}

func TestOldB530EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB530WurmOracle: "Witherscale Wurm", oldB530MentorOracle: "Corrosive Mentor",
		oldB530TowerOracle: "Tower Above", oldB530CanopyOracle: "Raking Canopy",
		oldB530InvocationOrcl: "Knollspine Invocation", oldB530JawsOracle: "Jaws of Stone",
		oldB530EmberGaleOracle: "Ember Gale", oldB530CrematorOracle: "Cragganwick Cremator",
		oldB530GhastlordOracle: "Ghastlord of Fugue",
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

func TestOldB530CorrosiveMentorGivesWitherToYourBlackCreaturesOnly(t *testing.T) {
	g, me, opp := oldB530Table(t)
	mentor := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Corrosive Mentor",
		TypeLine: "Creature — Elemental Rogue", OracleID: oldB530MentorOracle, Colors: []string{"B"},
		Power: 1, Toughness: 3, Owner: me.ID, Controller: me.ID})
	black := effectstest.B16Creature(g, me.ID, "Black Bear", "Creature — Bear", 2, 2, "B")
	green := effectstest.B16Creature(g, me.ID, "Green Bear", "Creature — Bear", 2, 2, "G")
	theirs := effectstest.B16Creature(g, opp.ID, "Their Black Bear", "Creature — Bear", 2, 2, "B")
	for name, tc := range map[string]struct {
		id   uuid.UUID
		want bool
	}{"the Mentor": {mentor, true}, "your black creature": {black, true}, "your green creature": {green, false}, "their black creature": {theirs, false}} {
		if got := effectstest.HasEffectiveKeyword(t, g, tc.id, game.KeywordWither); got != tc.want {
			t.Errorf("%s has wither = %v, want %v", name, got, tc.want)
		}
	}
}

func TestOldB530CragganwickCrematorDamagesByTheDiscardedCreatureCardsPower(t *testing.T) {
	for _, tc := range []struct {
		name string
		card *game.Card
		want int
	}{
		{"a creature card", &game.Card{Name: "Ogre", TypeLine: "Creature — Ogre", Power: 3, Toughness: 3}, 3},
		{"a noncreature card", &game.Card{Name: "Bolt", TypeLine: "Instant", ManaCost: "{R}"}, 0},
		{"an empty hand", nil, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g, me, opp := oldB530Table(t)
			me.Hand.Cards = nil
			if tc.card != nil {
				effectstest.RevealHand(me, *tc.card)
			}
			effectstest.AdvanceToMain(t, g)
			before := opp.Life
			effectstest.EnterFromHand(t, g, me.ID, "Cragganwick Cremator", "Creature — Giant Shaman", oldB530CrematorOracle)
			effectstest.PickPlayerTarget(t, g, me.ID, opp.ID)
			effectstest.PassPriorityAroundTable(t, g)
			if me.Hand.Size() != 0 {
				t.Errorf("hand size %d after the random discard, want 0", me.Hand.Size())
			}
			if opp.Life != before-tc.want {
				t.Errorf("opponent life %d → %d, want %d lost", before, opp.Life, tc.want)
			}
		})
	}
}

func TestOldB530JawsOfStoneDividesAsManyDamageAsYouHaveMountains(t *testing.T) {
	g, me, opp := oldB530Table(t)
	for i := 0; i < 3; i++ {
		effectstest.B12Permanent(g, me.ID, "Mountain", "Basic Land — Mountain")
	}
	effectstest.B12Permanent(g, me.ID, "Forest", "Basic Land — Forest")
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Jaws of Stone", TypeLine: "Sorcery", OracleID: oldB530JawsOracle, Owner: me.ID, Controller: me.ID})
	targets := []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}, {Kind: game.TargetCard, ID: bear}}
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Targets: targets, Distribution: map[uuid.UUID]int{opp.ID: 2, bear: 2}}); err == nil {
		t.Fatal("a division of 4 was accepted with three Mountains")
	}
	before := opp.Life
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Targets: targets, Distribution: map[uuid.UUID]int{opp.ID: 1, bear: 2}}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != before-1 {
		t.Errorf("opponent life %d → %d, want 1 lost", before, opp.Life)
	}
	if g.Battlefield.Contains(bear) {
		t.Error("the Bear took 2 and should have died")
	}
}

func TestOldB530TowerAbovePumpsAndMakesAnOpponentBlockTheAttacker(t *testing.T) {
	g, me, opp := oldB530Table(t)
	mine := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	effectstest.CastCatalogSpell(t, g, "Tower Above", "Sorcery", oldB530TowerOracle, effectstest.CardTarget(mine))
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.PtOf(t, g, mine); p != 6 || tough != 6 {
		t.Errorf("the creature is %d/%d, want 6/6", p, tough)
	}
	for _, kw := range []string{"trample", game.KeywordWither} {
		if !effectstest.HasEffectiveKeyword(t, g, mine, kw) {
			t.Errorf("the creature has no %s", kw)
		}
	}
	effectstest.DeclareAttack(t, g, opp.ID, mine)
	effectstest.PickTriggerTarget(t, g, me.ID, theirs)
	effectstest.PassPriorityAroundTable(t, g)
	recs := effectstest.BlocksAttackerRecords(g)
	if len(recs) != 1 || recs[0].Objects[0].ID != mine {
		t.Errorf("records = %+v, want one requiring a block of the attacker", recs)
	}
}

func TestOldB530TowerAboveGrantsNoAttackTriggerToOtherCreatures(t *testing.T) {
	g, me, opp := oldB530Table(t)
	mine := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	other := effectstest.B12Creature(g, me.ID, "Other", "Creature — Bear", 2, 2)
	effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	effectstest.CastCatalogSpell(t, g, "Tower Above", "Sorcery", oldB530TowerOracle, effectstest.CardTarget(mine))
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.DeclareAttack(t, g, opp.ID, other)
	if effectstest.LatestPickTarget(g, me.ID) != nil {
		t.Error("an attacker that was not the target triggered the granted ability")
	}
}

func TestOldB530WitherscaleWurmGivesWitherToTheCreatureItBlocksOrIsBlockedBy(t *testing.T) {
	g, me, opp := oldB530Table(t)
	wurm := effectstest.B12Push(g, me.ID, "Witherscale Wurm", "Creature — Wurm", oldB530WurmOracle, 9, 9)
	attacker := effectstest.B12Creature(g, opp.ID, "Attacker", "Creature — Bear", 2, 2)
	bystander := effectstest.B12Creature(g, opp.ID, "Bystander", "Creature — Bear", 2, 2)
	blocker := effectstest.B12Creature(g, opp.ID, "Blocker", "Creature — Bear", 2, 2)
	effectstest.WoeS2Block(g, me.ID, wurm, attacker)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.WoeS2Block(g, opp.ID, blocker, wurm)
	effectstest.PassPriorityAroundTable(t, g)
	for name, tc := range map[string]struct {
		id   uuid.UUID
		want bool
	}{"the creature it blocked": {attacker, true}, "the creature that blocked it": {blocker, true}, "a bystander": {bystander, false}, "the Wurm": {wurm, false}} {
		if got := effectstest.HasEffectiveKeyword(t, g, tc.id, game.KeywordWither); got != tc.want {
			t.Errorf("%s has wither = %v, want %v", name, got, tc.want)
		}
	}
}

func TestOldB530WitherscaleWurmShedsItsMinusCountersWhenItHurtsAnOpponent(t *testing.T) {
	g, me, opp := oldB530Table(t)
	wurm := effectstest.B12Push(g, me.ID, "Witherscale Wurm", "Creature — Wurm", oldB530WurmOracle, 9, 9)
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(wurm, game.CounterMinusOne, 3) })
	if n := effectstest.CountersOn(g, wurm, game.CounterMinusOne); n != 3 {
		t.Fatalf("setup: %d counters", n)
	}
	effectstest.DealCombatDamageToPlayer(g, wurm, opp.ID, 6)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountersOn(g, wurm, game.CounterMinusOne); n != 0 {
		t.Errorf("%d -1/-1 counters left, want none", n)
	}
	if p, tough := effectstest.PtOf(t, g, wurm); p != 9 || tough != 9 {
		t.Errorf("the Wurm is %d/%d, want 9/9", p, tough)
	}
}

func TestOldB530RakingCanopyHitsOnlyFliersThatAttackYou(t *testing.T) {
	for _, tc := range []struct {
		name   string
		flying bool
		atMe   bool
		dies   bool
	}{
		{"a flier attacking you", true, true, true},
		{"a ground creature attacking you", false, true, false},
		{"a flier attacking someone else", true, false, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g, me, opp := oldB530Table(t)
			third := g.Seats[(opp.Seat+1)%len(g.Seats)]
			if third.ID == me.ID {
				third = g.Seats[(opp.Seat+2)%len(g.Seats)]
			}
			effectstest.B12Push(g, third.ID, "Raking Canopy", "Enchantment", oldB530CanopyOracle, 0, 0)
			card := game.Card{InstanceID: uuid.New(), Name: "Walker", TypeLine: "Creature — Bear",
				Power: 3, Toughness: 4, Owner: opp.ID, Controller: opp.ID}
			if tc.flying {
				card.Name, card.TypeLine, card.Keywords = "Flier", "Creature — Bird", []string{"flying"}
			}
			attacker := effectstest.PushBattlefieldCardWithTimestamp(g, card)
			g.WithWriteLock(func() { g.Turn.ActiveSeat = opp.Seat })
			target := third.ID
			if !tc.atMe {
				target = me.ID
			}
			effectstest.DeclareAttack(t, g, target, attacker)
			effectstest.PassPriorityAroundTable(t, g)
			if dealt := effectstest.Pr6Marked(g, attacker) == 4 || !g.Battlefield.Contains(attacker); dealt != tc.dies {
				t.Errorf("took the 4 damage = %v, want %v", dealt, tc.dies)
			}
		})
	}
}

func TestOldB530KnollspineInvocationDealsXAndNeedsAMatchingDiscard(t *testing.T) {
	g, me, opp := oldB530Table(t)
	inv := effectstest.B12Push(g, me.ID, "Knollspine Invocation", "Enchantment", oldB530InvocationOrcl, 0, 0)
	two := effectstest.PushHandCardWithManaCost(me, "Two Drop", "Creature — Bear", "{1}{G}")
	three := effectstest.PushHandCardWithManaCost(me, "Three Drop", "Creature — Bear", "{2}{G}")
	effectstest.B10AddMana(me, "R", "R")
	params := func(discard uuid.UUID) game.ActivateAbilityParams {
		return game.ActivateAbilityParams{XValue: 2, Strict: true, DiscardIDs: []uuid.UUID{discard},
			Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}}
	}
	if err := g.ActivateCatalogAbility(me.ID, inv, 0, params(three)); err == nil {
		t.Fatal("a mana value 3 card was accepted for X=2")
	}
	before := opp.Life
	if err := g.ActivateCatalogAbility(me.ID, inv, 0, params(two)); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != before-2 {
		t.Errorf("opponent life %d → %d, want 2 lost", before, opp.Life)
	}
	if me.Hand.Contains(two) {
		t.Error("the discarded card is still in hand")
	}
}

func TestOldB530GhastlordCantBeBlockedAndTakesACardYouChoose(t *testing.T) {
	g, me, opp := oldB530Table(t)
	ghast := effectstest.B12Push(g, me.ID, "Ghastlord of Fugue", "Creature — Spirit Avatar", oldB530GhastlordOracle, 4, 4)
	ids := effectstest.RevealHand(opp,
		game.Card{Name: "Keep", TypeLine: "Sorcery", ManaCost: "{1}"},
		game.Card{Name: "Take", TypeLine: "Creature — Bear", ManaCost: "{2}", Power: 2, Toughness: 2})
	effectstest.DealCombatDamageToPlayer(g, ghast, opp.ID, 4)
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.OpenVariant(t, g)
	if err := g.ResolvePendingChoice(pick.ID, me.ID, []uuid.UUID{ids[1]}); err != nil {
		t.Fatalf("pick: %v", err)
	}
	if !g.Exile.Contains(ids[1]) || g.Exile.Contains(ids[0]) {
		t.Error("the chosen card was not the one exiled")
	}
	if opp.Hand.Size() != 1 {
		t.Errorf("hand size %d, want 1", opp.Hand.Size())
	}
}

func TestOldB530EmberGaleBarsBlockersAndPingsWhiteAndBlueCreatures(t *testing.T) {
	g, me, opp := oldB530Table(t)
	white := effectstest.B16Creature(g, opp.ID, "White", "Creature — Bear", 2, 2, "W")
	blue := effectstest.B16Creature(g, opp.ID, "Blue", "Creature — Bear", 2, 2, "U")
	both := effectstest.B16Creature(g, opp.ID, "Both", "Creature — Bear", 2, 3, "W", "U")
	red := effectstest.B16Creature(g, opp.ID, "Red", "Creature — Bear", 2, 2, "R")
	mine := effectstest.B16Creature(g, me.ID, "Mine White", "Creature — Bear", 2, 2, "W")
	effectstest.AdvanceToMain(t, g)
	effectstest.CastCatalogSpell(t, g, "Ember Gale", "Sorcery", oldB530EmberGaleOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	effectstest.PassPriorityAroundTable(t, g)
	for name, tc := range map[string]struct {
		id   uuid.UUID
		want int
	}{"white": {white, 1}, "blue": {blue, 1}, "white and blue": {both, 1}, "red": {red, 0}, "your own white": {mine, 0}} {
		if got := effectstest.Pr6Marked(g, tc.id); got != tc.want {
			t.Errorf("%s creature has %d damage, want %d", name, got, tc.want)
		}
	}
	effectstest.AssertRestrictions(t, g, red, game.CantBlock)
	effectstest.AssertRestrictions(t, g, mine, 0)
}
