package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const syggOracle = "9cc91ac5-6f2b-4fc0-b810-cdc92eebe6de"

func syggRow() cards.Card {
	return cards.Card{
		ID:       uuid.New(),
		Name:     "Sygg, Wanderwine Wisdom // Sygg, Wanderbrine Shield",
		Layout:   "transform",
		TypeLine: "Legendary Creature — Merfolk Wizard // Legendary Creature — Merfolk Rogue",
		OracleID: uuid.MustParse(syggOracle),
		Keywords: []string{"Transform"},
		CardFaces: []cards.CardFace{
			{Name: "Sygg, Wanderwine Wisdom", TypeLine: "Legendary Creature — Merfolk Wizard", ManaCost: "{1}{U}", Colors: []string{"U"}, Power: "2", Toughness: "2",
				OracleText: "Sygg can't be blocked.\nWhenever this creature enters or transforms into Sygg, Wanderwine Wisdom, target creature gains \"Whenever this creature deals combat damage to a player or planeswalker, draw a card\" until end of turn.\nAt the beginning of your first main phase, you may pay {W}. If you do, transform Sygg."},
			{Name: "Sygg, Wanderbrine Shield", TypeLine: "Legendary Creature — Merfolk Rogue", Colors: []string{"W"}, Power: "2", Toughness: "2",
				OracleText: "Sygg can't be blocked.\nWhenever this creature transforms into Sygg, Wanderbrine Shield, target creature you control gains protection from each color until your next turn.\nAt the beginning of your first main phase, you may pay {U}. If you do, transform Sygg."},
		},
	}
}

func TestSyggIsRegisteredFull(t *testing.T) {
	for _, oracle := range []string{syggOracle, syggOracle + "#1"} {
		spec, ok := Lookup(oracle)
		if !ok || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
			t.Fatalf("%s: spec = %+v, %v", oracle, spec, ok)
		}
	}
}

// Entering gives the chosen creature the draw trigger: an opponent's
// creature draws its controller a card on combat damage, noncombat damage
// draws nothing, and a creature without the grant draws nothing.
func TestSyggWanderwineWisdomGrantsTheDrawToTheTarget(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	opp := g.Seats[(seat+1)%len(g.Seats)]
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	mine := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	importAndCast(t, g, syggRow(), me)
	monSettle(t, g)
	pickTriggerTarget(t, g, me.ID, theirs)
	monSettle(t, g)

	myHand, theirHand := me.Hand.Size(), opp.Hand.Size()
	dealCombatDamageToPlayer(g, mine, opp.ID, 2)
	monSettle(t, g)
	if me.Hand.Size() != myHand {
		t.Fatalf("a creature without the grant drew: hand %d → %d", myHand, me.Hand.Size())
	}
	b27Damage(g, theirs, me.ID, 2)
	monSettle(t, g)
	if opp.Hand.Size() != theirHand {
		t.Fatalf("noncombat damage drew: hand %d → %d", theirHand, opp.Hand.Size())
	}
	dealCombatDamageToPlayer(g, theirs, me.ID, 2)
	monSettle(t, g)
	if opp.Hand.Size() != theirHand+1 {
		t.Fatalf("the granted creature's controller drew %d, want 1", opp.Hand.Size()-theirHand)
	}
	if me.Hand.Size() != myHand {
		t.Errorf("Sygg's controller drew off the opponent's creature")
	}
}

// Paying {W} turns Sygg over; the back's trigger can't target an
// opponent's creature and gives yours protection from each color until
// your next turn; paying {U} then turns Sygg back, which grants the draw
// again.
func TestSyggTransformsProtectsAndTransformsBack(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	opp := g.Seats[(seat+1)%len(g.Seats)]
	id := importToHand(syggRow(), me)
	c, ok := cardInZone(me.Hand, id)
	if !ok {
		t.Fatal("Sygg did not reach the hand")
	}
	me.Hand.Remove(id)
	pushBattlefieldCardWithTimestamp(g, c)
	mine := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)

	if restrictionsOf(t, g, id)&game.CantBeBlocked == 0 {
		t.Fatal("Wanderwine Wisdom can be blocked")
	}
	advanceTo(t, g, game.StepPrecombatMain)
	monSettle(t, g)
	me.ManaPool.AddMana(game.ManaToken{Color: "W"})
	answerPayUnless(t, g, me.ID, true)
	monSettle(t, g)
	if f := iwFace(g, id); f != 1 {
		t.Fatalf("Sygg is on face %d after paying {W}, want the back face", f)
	}
	if restrictionsOf(t, g, id)&game.CantBeBlocked == 0 {
		t.Fatal("Wanderbrine Shield can be blocked")
	}
	pick := latestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("transforming into Wanderbrine Shield asked for no target")
	}
	if err := g.ResolvePickTarget(pick.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: theirs}); err == nil {
		t.Fatal("the protection targeted a creature you don't control")
	}
	pickTriggerTarget(t, g, me.ID, mine)
	monSettle(t, g)
	for _, color := range game.AllColors {
		if !hasEffectiveKeyword(t, g, mine, game.ProtectionFromColor(color)) {
			t.Errorf("the bear lacks protection from %s", color)
		}
	}
	if hasEffectiveKeyword(t, g, theirs, game.ProtectionFromColor("W")) {
		t.Error("the opponent's bear gained protection")
	}

	advanceToStepOf(t, g, (seat+1)%len(g.Seats), game.StepUpkeep)
	if !hasEffectiveKeyword(t, g, mine, game.ProtectionFromColor("U")) {
		t.Fatal("the protection ended before your next turn")
	}
	advanceToStepOf(t, g, seat, game.StepUpkeep)
	if hasEffectiveKeyword(t, g, mine, game.ProtectionFromColor("U")) {
		t.Fatal("the protection outlasted the start of your next turn")
	}
	advanceTo(t, g, game.StepPrecombatMain)
	monSettle(t, g)
	me.ManaPool.AddMana(game.ManaToken{Color: "U"})
	answerPayUnless(t, g, me.ID, true)
	monSettle(t, g)
	if f := iwFace(g, id); f != 0 {
		t.Fatalf("Sygg is on face %d after paying {U}, want the front face", f)
	}
	pickTriggerTarget(t, g, me.ID, mine)
	monSettle(t, g)
	hand := me.Hand.Size()
	dealCombatDamageToPlayer(g, mine, opp.ID, 2)
	monSettle(t, g)
	if me.Hand.Size() != hand+1 {
		t.Errorf("transforming into Wanderwine Wisdom granted no draw: hand %d → %d", hand, me.Hand.Size())
	}
}

// Declining the {W} keeps Sygg on its front face.
func TestSyggDeclinedStaysFrontFace(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	id := importToHand(syggRow(), me)
	c, _ := cardInZone(me.Hand, id)
	me.Hand.Remove(id)
	pushBattlefieldCardWithTimestamp(g, c)
	advanceTo(t, g, game.StepPrecombatMain)
	monSettle(t, g)
	answerPayUnless(t, g, me.ID, false)
	monSettle(t, g)
	if f := iwFace(g, id); f != 0 {
		t.Fatalf("Sygg turned over without the {W} paid (face %d)", f)
	}
}
