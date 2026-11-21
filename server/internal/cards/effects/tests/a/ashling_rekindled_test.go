package testsa

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// ashling_rekindled_test.go — Ashling, Rekindled // Ashling, Rimebound
// (#4694) on the imported card, through the real steps and the real
// stack. The mana restriction itself is pinned in
// game/mana_restriction_mana_value_test.go.

const ashlingOracle = "31fc5d8b-94bc-4688-8a84-f15800e717a8"

func ashlingRow() cards.Card {
	return cards.Card{
		ID:       uuid.New(),
		Name:     "Ashling, Rekindled // Ashling, Rimebound",
		Layout:   "transform",
		TypeLine: "Legendary Creature — Elemental Sorcerer // Legendary Creature — Elemental Wizard",
		OracleID: uuid.MustParse(ashlingOracle),
		Keywords: []string{"Transform"},
		CardFaces: []cards.CardFace{
			{Name: "Ashling, Rekindled", TypeLine: "Legendary Creature — Elemental Sorcerer", ManaCost: "{1}{R}", Colors: []string{"R"}, Power: "1", Toughness: "3",
				OracleText: "Whenever this creature enters or transforms into Ashling, Rekindled, you may discard a card. If you do, draw a card.\nAt the beginning of your first main phase, you may pay {U}. If you do, transform Ashling."},
			{Name: "Ashling, Rimebound", TypeLine: "Legendary Creature — Elemental Wizard", Colors: []string{"U"}, Power: "1", Toughness: "3",
				OracleText: "Whenever this creature transforms into Ashling, Rimebound and at the beginning of your first main phase, add two mana of any one color. Spend this mana only to cast spells with mana value 4 or greater.\nAt the beginning of your first main phase, you may pay {R}. If you do, transform Ashling."},
		},
	}
}

// ashlingPut puts the imported card onto p's battlefield, front face up.
func ashlingPut(t *testing.T, g *game.Game, p *game.Player) uuid.UUID {
	t.Helper()
	id := effectstest.ImportToHand(ashlingRow(), p)
	c, ok := effectstest.CardInZone(p.Hand, id)
	if !ok {
		t.Fatal("Ashling did not reach the hand")
	}
	p.Hand.Remove(id)
	effectstest.PushBattlefieldCardWithTimestamp(g, c)
	return id
}

// ashlingAssertRestrictedPool checks the pool is exactly n mana of
// `color`, each spendable only on a spell with mana value 4 or greater.
func ashlingAssertRestrictedPool(t *testing.T, p *game.Player, color string, n int) {
	t.Helper()
	if got := p.ManaPool.Total(); got != n {
		t.Fatalf("pool holds %d mana, want %d", got, n)
	}
	want := game.ManaRestrictManaValueAtLeast(4)
	for _, tok := range p.ManaPool {
		if tok.Color != color || len(tok.Restrictions) != 1 || tok.Restrictions[0] != want {
			t.Fatalf("pool token %+v, want %s restricted to %q", tok, color, want)
		}
	}
}

// Entering: the rummage is "you may discard a card. If you do, draw a
// card".
func TestAshlingRekindledRummagesAsItEnters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.ImportAndCast(t, g, ashlingRow(), me)
	effectstest.MonSettle(t, g)
	hand, graves := me.Hand.Size(), me.Graveyard.Size()
	effectstest.AnswerDiscard(t, g, me.ID, me.Hand.Cards[0].InstanceID)
	effectstest.MonSettle(t, g)
	if me.Hand.Size() != hand || me.Graveyard.Size() != graves+1 {
		t.Errorf("hand %d, graveyard %d: want %d and %d (one discarded, one drawn)", me.Hand.Size(), me.Graveyard.Size(), hand, graves+1)
	}
}

// The front's first main phase: pay {U}, Ashling turns over, and turning
// into Ashling, Rimebound adds two mana of one colour that only a spell
// with mana value 4 or greater may spend.
func TestAshlingRekindledPaysUToTransformAndMakesRestrictedMana(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	id := ashlingPut(t, g, me)
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	effectstest.MonSettle(t, g)
	me.ManaPool.AddMana(game.ManaToken{Color: "U"})
	effectstest.AnswerPayUnless(t, g, me.ID, true)
	effectstest.MonSettle(t, g)
	if f := effectstest.IwFace(g, id); f != 1 {
		t.Fatalf("Ashling is on face %d after paying {U}, want the back face", f)
	}
	effectstest.AshlingPickColor(t, g, me.ID, "R")
	effectstest.MonSettle(t, g)
	ashlingAssertRestrictedPool(t, me, "R", 2)

	three := game.NewCard("Three Drop", me.ID)
	three.TypeLine, three.ManaCost = "Instant", "{1}{R}{R}"
	me.Hand.PushTop(three)
	var short *game.InsufficientManaError
	if err := g.CastSpell(me.ID, three.InstanceID, game.CastSpellParams{Strict: true}); !errors.As(err, &short) {
		t.Fatalf("a three-drop paid with Ashling's mana: err = %v, want InsufficientManaError", err)
	}
	me.ManaPool.AddMana(game.ManaToken{Color: "R"})
	me.ManaPool.AddMana(game.ManaToken{Color: "R"})
	four := game.NewCard("Four Drop", me.ID)
	four.TypeLine, four.ManaCost = "Instant", "{2}{R}{R}"
	me.Hand.PushTop(four)
	if err := g.CastSpell(me.ID, four.InstanceID, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("a four-drop with Ashling's mana and two more: %v", err)
	}
	if got := me.ManaPool.Total(); got != 0 {
		t.Errorf("after the four-drop the pool holds %d mana, want 0", got)
	}
}

// Declining keeps Ashling on its front face and adds nothing.
func TestAshlingRekindledDeclinedStaysFrontFace(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	id := ashlingPut(t, g, me)
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	effectstest.MonSettle(t, g)
	effectstest.AnswerPayUnless(t, g, me.ID, false)
	effectstest.MonSettle(t, g)
	if f := effectstest.IwFace(g, id); f != 0 {
		t.Fatalf("Ashling turned over without the {U} paid (face %d)", f)
	}
	if got := me.ManaPool.Total(); got != 0 {
		t.Errorf("pool holds %d mana, want none", got)
	}
}

// The back face's first main phase: the mana trigger and the {R} offer
// both trigger. The mana comes either way; paying {R} turns Ashling
// back, which triggers the front face's rummage. And the mana empties
// as the phase ends.
func TestAshlingRimeboundFirstMainPhase(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	seat := g.Turn.ActiveSeat
	// Onto the battlefield once this main phase has begun, so nothing of
	// Ashling's triggers for it; then turned over by an instruction.
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	id := ashlingPut(t, g, me)
	effectstest.IwTransform(t, g, id)
	effectstest.MonSettle(t, g)
	effectstest.AshlingPickColor(t, g, me.ID, "G")
	effectstest.MonSettle(t, g)
	ashlingAssertRestrictedPool(t, me, "G", 2)

	// The next turn of this seat: the pool emptied long ago.
	effectstest.AdvanceToStepOf(t, g, (seat+1)%len(g.Seats), game.StepUpkeep)
	if got := me.ManaPool.Total(); got != 0 {
		t.Fatalf("Ashling's mana is still in the pool on the next turn: %d", got)
	}
	effectstest.AdvanceToStepOf(t, g, seat, game.StepPrecombatMain)
	picked, offered, rummaged := false, false, false
prompts:
	for range 12 {
		effectstest.MonSettle(t, g)
		switch {
		case effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceMana, me.ID) != nil:
			effectstest.AshlingPickColor(t, g, me.ID, "U")
			picked = true
		case effectstest.HasPayUnlessFor(g, me.ID):
			me.ManaPool.AddMana(game.ManaToken{Color: "R"})
			effectstest.AnswerPayUnless(t, g, me.ID, true)
			offered = true
		case effectstest.DiscardChoiceFor(g, me.ID) != nil:
			effectstest.AnswerDiscard(t, g, me.ID)
			rummaged = true
		default:
			break prompts
		}
	}
	if !picked || !offered {
		t.Fatalf("picked %v, offered %v: want both first-main-phase triggers", picked, offered)
	}
	if f := effectstest.IwFace(g, id); f != 0 {
		t.Fatalf("Ashling is on face %d after paying {R}, want the front face", f)
	}
	if !rummaged {
		t.Error("turning into Ashling, Rekindled asked for no rummage")
	}
	ashlingAssertRestrictedPool(t, me, "U", 2)

	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	if got := me.ManaPool.Total(); got != 0 {
		t.Errorf("Ashling's mana did not empty as the main phase ended: %d", got)
	}
}
