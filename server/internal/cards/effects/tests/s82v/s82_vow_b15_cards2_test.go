package testss82v

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// --- Runo Stromkirk // Krothuss ---------------------------------------------

func vowB15RunoRow() cards.Card {
	return effectstest.VowB15Row(vowB15Runo, "Runo Stromkirk", "Legendary Creature — Vampire Cleric", "{1}{U}{B}", "1", "4",
		"Krothuss, Lord of the Deep", "Legendary Creature — Kraken Horror", "3", "5", []string{"U", "B"})
}

func TestVowB15RunoEntersAndTucksACreatureCardFromTheGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := game.Card{InstanceID: uuid.New(), Name: "Dead Bear", TypeLine: "Creature — Bear", Owner: me.ID, Controller: me.ID}
	rock := game.Card{InstanceID: uuid.New(), Name: "Dead Rock", TypeLine: "Artifact", Owner: me.ID, Controller: me.ID}
	me.Graveyard.PushTop(bear)
	me.Graveyard.PushTop(rock)
	effectstest.ImportAndCast(t, g, vowB15RunoRow(), me)
	effectstest.MonSettle(t, g)
	pick := effectstest.LatestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("entering asked for no target")
	}
	if err := g.ResolvePickTargets(pick.ID, me.ID, []game.TargetRef{{Kind: game.TargetCard, ID: rock.InstanceID}}); err == nil {
		t.Fatal("a noncreature card was a legal target")
	}
	effectstest.AnswerTriggerTargets(t, g, me.ID, bear.InstanceID)
	effectstest.PassPriorityAroundTable(t, g)
	if z := effectstest.VowB15Zone(g, bear.InstanceID); z != game.ZoneLibrary {
		t.Fatalf("the creature card is in %q, want the library", z)
	}
	if top, err := me.Library.Top(); err != nil || top.InstanceID != bear.InstanceID {
		t.Errorf("the creature card is not on top of the library")
	}
}

func TestVowB15RunoTransformsWhenYouRevealAMVSixCreatureOnTop(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	runo := effectstest.VowB15Put(t, g, vowB15RunoRow(), me)
	effectstest.VowB15Library(me, "Small Creature", "Creature — Bear", "{3}{G}")
	effectstest.BeginUpkeep(t, g, me.ID)
	if f := effectstest.IwFace(g, runo); f != 0 {
		t.Fatalf("a mana value 4 creature turned Runo to face %d", f)
	}
	effectstest.VowB15Library(me, "Spell", "Sorcery", "{7}")
	effectstest.BeginUpkeep(t, g, me.ID)
	if f := effectstest.IwFace(g, runo); f != 0 {
		t.Fatalf("a sorcery turned Runo to face %d", f)
	}
	big := effectstest.VowB15Library(me, "Big Creature", "Creature — Wurm", "{5}{G}")
	effectstest.BeginUpkeep(t, g, me.ID)
	effectstest.AnswerChooseCards(t, g, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if f := effectstest.IwFace(g, runo); f != 0 {
		t.Fatalf("declining the reveal still turned Runo to face %d", f)
	}
	effectstest.BeginUpkeep(t, g, me.ID)
	effectstest.AnswerChooseCards(t, g, me.ID, big)
	effectstest.PassPriorityAroundTable(t, g)
	if f := effectstest.IwFace(g, runo); f != 1 {
		t.Fatalf("revealing a mana value 6 creature left Runo on face %d", f)
	}
	if z := effectstest.VowB15Zone(g, big); z != game.ZoneLibrary {
		t.Errorf("the revealed card moved to %q, want to stay on top of the library", z)
	}
}

func TestVowB15KrothussCopiesAnAttackerAndTwiceForAKraken(t *testing.T) {
	for _, tc := range []struct {
		typeLine string
		want     int
	}{{"Creature — Bear", 1}, {"Creature — Serpent", 2}} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		k := effectstest.VowB15Put(t, g, vowB15RunoRow(), me)
		effectstest.IwTransform(t, g, k)
		other := effectstest.B12Creature(g, me.ID, "Other", tc.typeLine, 2, 2)
		effectstest.DeclareAttack(t, g, opp.ID, k, other)
		effectstest.AnswerTriggerTargets(t, g, me.ID, other)
		effectstest.PassAndSettleAttackSplit(t, g, me.ID)
		effectstest.PassPriorityAroundTable(t, g)
		n := 0
		for _, c := range g.Battlefield.Cards {
			if c.Name == "Other" && c.IsToken() && c.AttackingTarget != uuid.Nil && c.Tapped {
				n++
			}
		}
		if n != tc.want {
			t.Errorf("%s: %d tapped and attacking token copies, want %d", tc.typeLine, n, tc.want)
		}
	}
}

// --- Soulcipher Board // Cipherbound Spirit ---------------------------------

func vowB15BoardRow() cards.Card {
	return effectstest.VowB15Row(vowB15Board, "Soulcipher Board", "Artifact", "{1}{U}", "", "",
		"Cipherbound Spirit", "Creature — Spirit", "3", "2", []string{"U"})
}

func TestVowB15BoardLosesAnOmenCounterPerCreatureCardAndTransformsAtNone(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	board := effectstest.ImportToBattlefield(t, g, vowB15BoardRow(), me)
	if n := effectstest.CountersOn(g, board, "omen"); n != 3 {
		t.Fatalf("enters with %d omen counters, want 3", n)
	}
	mill := func(typeLine string) {
		effectstest.VowB15Library(me, "Milled", typeLine, "{1}")
		g.WithWriteLock(func() {
			if err := g.MillNForEffect(me.ID, 1); err != nil {
				t.Fatalf("mill: %v", err)
			}
		})
		effectstest.PassPriorityAroundTable(t, g)
	}
	mill("Sorcery")
	if n := effectstest.CountersOn(g, board, "omen"); n != 3 {
		t.Fatalf("a sorcery card removed a counter: %d left", n)
	}
	mill("Creature — Bear")
	mill("Creature — Bear")
	if n := effectstest.CountersOn(g, board, "omen"); n != 1 {
		t.Fatalf("%d omen counters after two creature cards, want 1", n)
	}
	if f := effectstest.IwFace(g, board); f != 0 {
		t.Fatalf("transformed with a counter left: face %d", f)
	}
	// A creature card discarded is put into the graveyard from the hand.
	d := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: d, Name: "Discarded", TypeLine: "Creature — Bear", Owner: me.ID, Controller: me.ID})
	g.WithWriteLock(func() {
		if err := g.DiscardCardsKeyedForEffect(me.ID, []uuid.UUID{d}, uuid.Nil, game.KeyedTail{}); err != nil {
			t.Fatalf("discard: %v", err)
		}
	})
	effectstest.PassPriorityAroundTable(t, g)
	if f := effectstest.IwFace(g, board); f != 1 {
		t.Fatalf("no counters left but face %d, want the Spirit", f)
	}
}

func TestVowB15BoardLooksAtTwoAndBinsOne(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	board := effectstest.VowB15Put(t, g, vowB15BoardRow(), me)
	keep := effectstest.VowB15Library(me, "Keep", "Sorcery", "{1}")
	bin := effectstest.VowB15Library(me, "Bin", "Instant", "{1}")
	vowB15Mana(me, "U", "U")
	vowB15Activate(t, g, me, board, "{1}{U}, {T}: Look at the top two cards of your library. Put one of them into your graveyard.")
	effectstest.AnswerChooseCards(t, g, me.ID, bin)
	effectstest.PassPriorityAroundTable(t, g)
	if z := effectstest.VowB15Zone(g, bin); z != game.ZoneGraveyard {
		t.Errorf("the chosen card is in %q, want the graveyard", z)
	}
	if z := effectstest.VowB15Zone(g, keep); z != game.ZoneLibrary {
		t.Errorf("the other card is in %q, want the library", z)
	}
}

// --- Voltaic Visionary // Volt-Charged Berserker ----------------------------

func vowB15VisionaryRow() cards.Card {
	return effectstest.VowB15Row(vowB15Visioner, "Voltaic Visionary", "Creature — Human Wizard", "{1}{R}", "3", "1",
		"Volt-Charged Berserker", "Creature — Human Berserker", "4", "3", []string{"R"})
}

func TestVowB15VisionaryPaysTwoExilesAndTransformsWhenYouPlayThatCard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	v := effectstest.VowB15Put(t, g, vowB15VisionaryRow(), me)
	label := "{T}: This creature deals 2 damage to you. Exile the top card of your library. You may play that card this turn. Activate only as a sorcery."
	// A card another ability exiled does not transform it.
	free := effectstest.VowB15Library(me, "Free Bear", "Creature — Bear", "{0}")
	life := me.Life
	vowB15Activate(t, g, me, v, label)
	if me.Life != life-2 {
		t.Errorf("life %d, want %d", me.Life, life-2)
	}
	if z := effectstest.VowB15Zone(g, free); z != game.ZoneExile {
		t.Fatalf("the top card is in %q, want exile", z)
	}
	if f := effectstest.IwFace(g, v); f != 0 {
		t.Fatalf("transformed before the card was played: face %d", f)
	}
	if err := g.CastSpell(me.ID, free, game.CastSpellParams{Strict: true, FromZone: "exile"}); err != nil {
		t.Fatalf("cast the exiled card: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if f := effectstest.IwFace(g, v); f != 1 {
		t.Fatalf("playing the exiled card left the Visionary on face %d", f)
	}
}

func TestVowB15VisionaryOnlyAtSorcerySpeedAndNotForOtherExiledCards(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	v := effectstest.VowB15Put(t, g, vowB15VisionaryRow(), me)
	effectstest.AdvanceTo(t, g, game.StepUpkeep)
	if err := g.ActivateCatalogAbility(me.ID, v, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Error("activated outside a main phase")
	}
	effectstest.AdvanceToMain(t, g)
	other := uuid.New()
	g.Exile.PushTop(game.Card{InstanceID: other, Name: "Other Exile", TypeLine: "Creature — Bear", ManaCost: "{0}", Owner: me.ID, Controller: me.ID})
	g.WithWriteLock(func() {
		g.GrantCastPermissionForEffect(game.CastPermission{Player: me.ID, Zone: game.ZoneExile, Filter: game.PermissionFilter{}})
	})
	if err := g.CastSpell(me.ID, other, game.CastSpellParams{FromZone: "exile"}); err != nil {
		t.Skipf("the control permission did not open the card: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if f := effectstest.IwFace(g, v); f != 0 {
		t.Errorf("a card not exiled with the Visionary turned it to face %d", f)
	}
}

// --- Jacob Hauken, Inspector // Hauken's Insight ----------------------------

func vowB15JacobRow() cards.Card {
	return effectstest.VowB15Row(vowB15Jacob, "Jacob Hauken, Inspector", "Legendary Creature — Human Advisor", "{1}{U}", "0", "2",
		"Hauken's Insight", "Legendary Enchantment", "", "", []string{"U"})
}

func TestVowB15JacobDrawsThenExilesFromHandFaceDownAndTransformsOnPayment(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	jacob := effectstest.VowB15Put(t, g, vowB15JacobRow(), me)
	drawn := effectstest.VowB15Library(me, "Drawn Card", "Sorcery", "{1}")
	label := "{T}: Draw a card, then exile a card from your hand face down. You may look at that card for as long as it remains exiled. You may pay {4}{U}{U}. If you do, transform Jacob Hauken."
	vowB15Activate(t, g, me, jacob, label)
	effectstest.AnswerChooseCards(t, g, me.ID, drawn)
	effectstest.PassPriorityAroundTable(t, g)
	c, ok := effectstest.ExileCard(g, drawn)
	if !ok {
		t.Fatalf("the chosen card is in %q, want exile", effectstest.VowB15Zone(g, drawn))
	}
	if !c.FaceDown || !c.KnownBy[me.ID] || c.KnownBy[opp.ID] {
		t.Errorf("the exiled card: face down %v, known by you %v, known by an opponent %v", c.FaceDown, c.KnownBy[me.ID], c.KnownBy[opp.ID])
	}
	if c.ExiledWith.ID != jacob {
		t.Errorf("the card is not linked to Jacob Hauken")
	}
	effectstest.AnswerPayUnless(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if f := effectstest.IwFace(g, jacob); f != 0 {
		t.Fatalf("declined payment but face %d", f)
	}
}

func TestVowB15JacobTransformsWhenYouPaySixAndInsightPlaysACardFreeOncePerTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	jacob := effectstest.VowB15Put(t, g, vowB15JacobRow(), me)
	first := effectstest.VowB15Library(me, "First Bear", "Creature — Bear", "{3}{G}")
	label := "{T}: Draw a card, then exile a card from your hand face down. You may look at that card for as long as it remains exiled. You may pay {4}{U}{U}. If you do, transform Jacob Hauken."
	vowB15Activate(t, g, me, jacob, label)
	effectstest.AnswerChooseCards(t, g, me.ID, first)
	vowB15Mana(me, "U", "U", "U", "U", "U", "U")
	effectstest.AnswerPayUnless(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if f := effectstest.IwFace(g, jacob); f != 1 {
		t.Fatalf("paid {4}{U}{U} but face %d", f)
	}
	// The card the front exiled is playable through the back face, free,
	// once during each of your turns.
	second := effectstest.VowB15Library(me, "Second Bear", "Creature — Bear", "{3}{G}")
	effectstest.BeginUpkeep(t, g, me.ID)
	if z := effectstest.VowB15Zone(g, second); z != game.ZoneExile {
		t.Fatalf("the upkeep exile left the top card in %q", z)
	}
	effectstest.AdvanceToMain(t, g)
	if err := g.CastSpell(me.ID, first, game.CastSpellParams{Strict: true, FromZone: "exile"}); err != nil {
		t.Fatalf("cast the front-face card free: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if _, ok := effectstest.BattlefieldCard(g, first); !ok {
		t.Fatal("the cast Bear did not resolve")
	}
	if err := g.CastSpell(me.ID, second, game.CastSpellParams{Strict: true, FromZone: "exile"}); err == nil {
		t.Error("a second free play in the same turn was allowed")
	}
}
