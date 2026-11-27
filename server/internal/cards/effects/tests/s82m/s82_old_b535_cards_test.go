package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b535_cards_test.go — Shadowmoor: Instants and sorceries, batch 535
// (#5591, S82 OLD). Every test plays the real catalog entry through the engine.

const (
	oldB535Worldpurge = "285ecbd5-ff7c-4919-ab0b-7fc05e5a0a0a"
	oldB535Advice     = "5063f015-bd1c-4e14-9888-63e202be2ac8"
	oldB535Din        = "6480ab32-3a29-4449-8ff7-c8e43c320ed3"
	oldB535Fate       = "7a07451f-70b3-4bfd-af54-6b4182369652"
	oldB535Swing      = "92d07c7c-be30-4bcc-a632-99986fc5315c"
	oldB535Repel      = "94a241bf-5709-4890-8f06-017833869d62"
	oldB535Reknit     = "963e36de-b803-417f-8166-4d328d358203"
	oldB535Gambit     = "ddf8d3d2-b0f9-43bc-9c9e-cb08fdedcc72"
)

func TestOldB535CardsAreRegisteredAndWhole(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB535Worldpurge: "Worldpurge",
		oldB535Advice:     "Advice from the Fae",
		oldB535Din:        "Din of the Fireherd",
		oldB535Fate:       "Fate Transfer",
		oldB535Swing:      "Wild Swing",
		oldB535Repel:      "Repel Intruders",
		oldB535Reknit:     "Reknit",
		oldB535Gambit:     "Thoughtweft Gambit",
	} {
		effectstest.RequireFullNoCaveats(t, oracle, name)
	}
}

// oldB535Cast casts a card from the active seat's hand with its mana floated,
// in a main phase, under strict rules.
func oldB535Cast(t *testing.T, g *game.Game, card game.Card, mana string, params game.CastSpellParams) (uuid.UUID, error) {
	t.Helper()
	me := g.Seats[g.Turn.ActiveSeat]
	card.InstanceID = uuid.New()
	card.Owner, card.Controller = me.ID, me.ID
	me.Hand.PushTop(card)
	effectstest.AdvanceToMain(t, g)
	effectstest.FloatForTest(g, me, mana)
	params.Strict = true
	return card.InstanceID, g.CastSpell(me.ID, card.InstanceID, params)
}

func TestOldB535ReknitRegeneratesTheTargetPermanent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	if _, err := oldB535Cast(t, g, game.Card{Name: "Reknit", TypeLine: "Instant", OracleID: oldB535Reknit, ManaCost: "{1}{G/W}", Colors: []string{"G", "W"}},
		"CG", game.CastSpellParams{Targets: effectstest.CardTarget(bear)}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.RegenShieldsOn(t, g, bear); got != 1 {
		t.Fatalf("%d regeneration shields, want 1", got)
	}
	effectstest.DestroyIt(t, g, bear)
	if !g.Battlefield.Contains(bear) {
		t.Error("the regenerated creature was destroyed")
	}
}

func TestOldB535ThoughtweftGambitTapsOpponentsCreaturesAndUntapsYours(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	effectstest.TapForTest(t, g, mine)
	if _, err := oldB535Cast(t, g, game.Card{Name: "Thoughtweft Gambit", TypeLine: "Instant", OracleID: oldB535Gambit, ManaCost: "{4}{W/U}{W/U}", Colors: []string{"W", "U"}},
		"CCCCWU", game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.IsTapped(g, mine) {
		t.Error("your creature is still tapped")
	}
	if !effectstest.IsTapped(g, theirs) {
		t.Error("the opponent's creature is untapped")
	}
}

func TestOldB535FateTransferMovesEveryKindOfCounter(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	from := effectstest.B12Creature(g, me.ID, "Donor", "Creature — Bear", 2, 2)
	to := effectstest.B12Creature(g, me.ID, "Taker", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() {
		_ = g.AddCounterForEffect(from, game.CounterPlusOne, 2)
		_ = g.AddCounterForEffect(from, "charge", 1)
	})
	card := game.Card{Name: "Fate Transfer", TypeLine: "Instant", OracleID: oldB535Fate, ManaCost: "{1}{U/B}", Colors: []string{"U", "B"}}
	if _, err := oldB535Cast(t, g, card, "CU", game.CastSpellParams{Targets: effectstest.CardRefs(from, from)}); err == nil {
		t.Fatal("the same creature was accepted as both targets")
	}
	if _, err := oldB535Cast(t, g, card, "CU", game.CastSpellParams{Targets: effectstest.CardRefs(from, to)}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	for kind, want := range map[string][2]int{game.CounterPlusOne: {0, 2}, "charge": {0, 1}} {
		if got := effectstest.CountersOn(g, from, kind); got != want[0] {
			t.Errorf("donor has %d %s counters, want %d", got, kind, want[0])
		}
		if got := effectstest.CountersOn(g, to, kind); got != want[1] {
			t.Errorf("taker has %d %s counters, want %d", got, kind, want[1])
		}
	}
}

func TestOldB535WildSwingDestroysExactlyOneOfThreeNonenchantmentPermanents(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a := effectstest.B12Creature(g, opp.ID, "A", "Creature — Bear", 2, 2)
	b := effectstest.B12Permanent(g, opp.ID, "B", "Artifact")
	c := effectstest.B12Creature(g, me.ID, "C", "Creature — Bear", 2, 2)
	ench := effectstest.B12Permanent(g, opp.ID, "Ench", "Enchantment")
	card := game.Card{Name: "Wild Swing", TypeLine: "Sorcery", OracleID: oldB535Swing, ManaCost: "{3}{R}", Colors: []string{"R"}}
	if _, err := oldB535Cast(t, g, card, "CCCR", game.CastSpellParams{Targets: effectstest.CardRefs(a, b, ench)}); err == nil {
		t.Fatal("an enchantment was accepted as a target")
	}
	if _, err := oldB535Cast(t, g, card, "CCCR", game.CastSpellParams{Targets: effectstest.CardRefs(a, b)}); err == nil {
		t.Fatal("two targets were accepted")
	}
	if _, err := oldB535Cast(t, g, card, "CCCR", game.CastSpellParams{Targets: effectstest.CardRefs(a, b, c)}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	gone := 0
	for _, id := range []uuid.UUID{a, b, c} {
		if !g.Battlefield.Contains(id) {
			gone++
		}
	}
	if gone != 1 {
		t.Errorf("%d of the three targets were destroyed, want exactly 1", gone)
	}
	if !g.Battlefield.Contains(ench) {
		t.Error("the enchantment was destroyed")
	}
}

func TestOldB535RepelIntrudersMakesKithkinForWhiteAndCountersForBlue(t *testing.T) {
	card := game.Card{Name: "Repel Intruders", TypeLine: "Instant", OracleID: oldB535Repel, ManaCost: "{3}{W/U}", Colors: []string{"W", "U"}}
	cast := func(t *testing.T, mana string, withSpell bool) (g *game.Game, spell uuid.UUID) {
		g = effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		var params game.CastSpellParams
		if withSpell {
			effectstest.AdvanceToMain(t, g)
			spell = effectstest.HandCardFull(me, "Bear Spell", "Creature — Bear", "{1}", "", nil)
			if err := g.CastSpell(me.ID, spell, game.CastSpellParams{}); err != nil {
				t.Fatalf("cast the creature spell: %v", err)
			}
			params.Targets = effectstest.CardTarget(spell)
		}
		if _, err := oldB535Cast(t, g, card, mana, params); err != nil {
			t.Fatalf("cast: %v", err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		return g, spell
	}
	t.Run("white only", func(t *testing.T) {
		g, spell := cast(t, "CCCW", true)
		if n := len(effectstest.TokensNamed(g, "Kithkin Soldier")); n != 2 {
			t.Errorf("%d Kithkin Soldiers, want 2", n)
		}
		if !g.Battlefield.Contains(spell) {
			t.Error("the creature spell was countered without {U} spent")
		}
	})
	t.Run("blue only", func(t *testing.T) {
		g, spell := cast(t, "CCCU", true)
		if n := len(effectstest.TokensNamed(g, "Kithkin Soldier")); n != 0 {
			t.Errorf("%d Kithkin Soldiers, want 0", n)
		}
		if g.Battlefield.Contains(spell) {
			t.Error("the creature spell was not countered")
		}
	})
	t.Run("both", func(t *testing.T) {
		g, spell := cast(t, "CCWU", true)
		if n := len(effectstest.TokensNamed(g, "Kithkin Soldier")); n != 2 {
			t.Errorf("%d Kithkin Soldiers, want 2", n)
		}
		if g.Battlefield.Contains(spell) {
			t.Error("the creature spell was not countered")
		}
	})
	t.Run("no target chosen", func(t *testing.T) {
		g, _ := cast(t, "CCCW", false)
		if n := len(effectstest.TokensNamed(g, "Kithkin Soldier")); n != 2 {
			t.Errorf("%d Kithkin Soldiers, want 2", n)
		}
	})
}

func TestOldB535DinOfTheFireherdMakesAnElementalAndAnEdictForEachBlackAndRedCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Black Bear", TypeLine: "Creature — Bear",
		Colors: []string{"B"}, Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID})
	var theirCreatures, theirLands []uuid.UUID
	for _, n := range []string{"One", "Two", "Three", "Four"} {
		theirCreatures = append(theirCreatures, effectstest.B12Creature(g, opp.ID, n, "Creature — Bear", 2, 2))
	}
	for _, n := range []string{"Swamp", "Mountain", "Forest"} {
		theirLands = append(theirLands, effectstest.PushLandFor(g, opp.ID, n, "Basic Land — "+n))
	}
	// The token is black and red: two black creatures and one red creature in all.
	if _, err := oldB535Cast(t, g, game.Card{Name: "Din of the Fireherd", TypeLine: "Sorcery", OracleID: oldB535Din, ManaCost: "{5}{B/R}{B/R}{B/R}", Colors: []string{"B", "R"}},
		"CCCCCBBB", game.CastSpellParams{Targets: []game.TargetRef{effectstest.PlayerTarget(opp.ID)}}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	for i := 0; i < 6; i++ {
		c := effectstest.SacrificeChoiceFor(g, opp.ID)
		if c == nil {
			break
		}
		effectstest.AnswerSacrifice(t, g, opp.ID, effectstest.SacrificeOptions(c)[0])
		effectstest.PassPriorityAroundTable(t, g)
	}
	if n := len(effectstest.TokensNamed(g, "Elemental")); n != 1 {
		t.Fatalf("%d Elemental tokens, want 1", n)
	}
	creatures, lands := 0, 0
	for _, id := range theirCreatures {
		if g.Battlefield.Contains(id) {
			creatures++
		}
	}
	for _, id := range theirLands {
		if g.Battlefield.Contains(id) {
			lands++
		}
	}
	if creatures != 2 {
		t.Errorf("the opponent kept %d of 4 creatures, want 2 (two black creatures)", creatures)
	}
	if lands != 2 {
		t.Errorf("the opponent kept %d of 3 lands, want 2 (one red creature)", lands)
	}
}

func TestOldB535AdviceFromTheFaeTakesTwoWithMoreCreaturesAndOneOtherwise(t *testing.T) {
	for _, tc := range []struct {
		name     string
		creature bool
		want     int
	}{{"more creatures", true, 2}, {"tied for creatures", false, 1}} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[0]
			effectstest.SeedLibrary(me, "A", "B", "C", "D", "E", "F", "G")
			if tc.creature {
				effectstest.B12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
			}
			before := me.Hand.Size()
			if _, err := oldB535Cast(t, g, game.Card{Name: "Advice from the Fae", TypeLine: "Sorcery", OracleID: oldB535Advice, ManaCost: "{2/U}{2/U}{2/U}", Colors: []string{"U"}},
				"UUU", game.CastSpellParams{}); err != nil {
				t.Fatalf("cast for {U}{U}{U}: %v", err)
			}
			effectstest.PassPriorityAroundTable(t, g)
			pick := effectstest.ChooseCardsChoiceFor(g, me.ID)
			if pick == nil {
				t.Fatal("no choose-cards prompt")
			}
			if pick.ChooseMax != tc.want {
				t.Errorf("the prompt takes %d cards, want %d", pick.ChooseMax, tc.want)
			}
			effectstest.AnswerChooseCards(t, g, me.ID, pick.ChooseCards[:tc.want]...)
			effectstest.PassPriorityAroundTable(t, g)
			if me.Hand.Size() != before+tc.want {
				t.Errorf("hand %d, want %d", me.Hand.Size(), before+tc.want)
			}
		})
	}
}

func TestOldB535WorldpurgeReturnsEverythingKeepsSevenCardsAndEmptiesPools(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	creature := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	var token uuid.UUID
	g.WithWriteLock(func() {
		made, err := g.CreateTokensForEffect(me.ID, effects.TokenCard("1/1 white Kithkin Soldier"), 1, game.TokenEntryOptions{})
		if err != nil || len(made) != 1 {
			t.Fatalf("token: %v %v", made, err)
		}
		token = made[0]
	})
	land := effectstest.PushLandFor(g, opp.ID, "Island", "Basic Land — Island")
	for me.Hand.Size() < 9 {
		effectstest.HandCardFull(me, "Filler", "Sorcery", "{1}", "", nil)
	}
	for opp.Hand.Size() < 3 {
		effectstest.HandCardFull(opp, "Filler", "Sorcery", "{1}", "", nil)
	}
	effectstest.SeedLibrary(me, "L1", "L2")
	library := me.Library.Size()
	oppHand := opp.Hand.Size()
	if _, err := oldB535Cast(t, g, game.Card{Name: "Worldpurge", TypeLine: "Sorcery", OracleID: oldB535Worldpurge, ManaCost: "{4}{W/U}{W/U}{W/U}{W/U}", Colors: []string{"W", "U"}},
		"CCCCWWWWGG", game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{creature, token, land} {
		if g.Battlefield.Contains(id) {
			t.Errorf("permanent %s is still on the battlefield", id)
		}
	}
	// 9 in hand + the Bear came back = 10 cards; keep seven, shuffle three.
	pick := effectstest.ChooseCardsChoiceFor(g, me.ID)
	if pick == nil {
		t.Fatal("no prompt to choose seven cards")
	}
	if len(pick.ChooseCards) != 10 {
		t.Errorf("%d candidates, want 10: the returned token is not a card", len(pick.ChooseCards))
	}
	if pick.ChooseMax != 7 || pick.ChooseMin != 7 {
		t.Errorf("the prompt takes %d-%d cards, want exactly 7", pick.ChooseMin, pick.ChooseMax)
	}
	effectstest.AnswerChooseCards(t, g, me.ID, pick.ChooseCards[:7]...)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != 7 {
		t.Errorf("hand %d, want 7", me.Hand.Size())
	}
	if me.Library.Size() != library+3 {
		t.Errorf("library %d, want %d", me.Library.Size(), library+3)
	}
	if want := oppHand + 1; opp.Hand.Size() != want {
		t.Errorf("the opponent's hand is %d, want %d (the land came back; no shuffling at seven or fewer)", opp.Hand.Size(), want)
	}
	for _, p := range g.Seats {
		if effectstest.PoolSize(g, p.ID) != 0 {
			t.Errorf("a player kept unspent mana")
		}
	}
}
