package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/deck"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// goldbug_humanitys_ally_test.go — ADR 0152 (#1856): More Than Meets
// the Eye, convert and living metal end to end on the proof card. The
// card goes down the import road (deck.ToGameCard) for the reason
// disturb_test.go gives: its back face only exists on an imported card.

func goldbugRow() cards.Card {
	row := disturbRow(goldbugOracleID, []string{"W", "U"},
		disturbFace{"Goldbug, Humanity's Ally", "Legendary Artifact Creature — Robot", "{1}{W}{U}", "3", "3"},
		disturbFace{"Goldbug, Scrappy Scout", "Legendary Artifact — Vehicle", "", "1", "3"})
	row.Keywords = []string{"Living metal", "Convert", "More Than Meets the Eye"}
	row.CardFaces[0].OracleText = "More Than Meets the Eye {W}{U} (You may cast this card converted for {W}{U}.)\n" +
		"Prevent all combat damage that would be dealt to attacking Humans you control.\n" +
		"Whenever you cast your second spell each turn, convert Goldbug."
	row.CardFaces[1].OracleText = "Living metal (During your turn, this Vehicle is also a creature.)\n" +
		"Human spells you control can't be countered.\n" +
		"Whenever Goldbug and at least one Human attack, draw a card and convert Goldbug."
	return row
}

const mtmte = "more_than_meets_the_eye"

// goldbugToMain imports Goldbug into the active seat's hand and walks
// to the precombat main phase with an empty pool.
func goldbugToMain(t *testing.T, g *game.Game) (*game.Player, uuid.UUID) {
	t.Helper()
	me := g.Seats[g.Turn.ActiveSeat]
	id := importToHand(goldbugRow(), me)
	advanceToMain(t, g)
	me.ManaPool = nil
	return me, id
}

// castGoldbugConverted casts Goldbug for More Than Meets the Eye with
// {W}{U} in the pool, strictly, and resolves it.
func castGoldbugConverted(t *testing.T, g *game.Game, me *game.Player, id uuid.UUID) {
	t.Helper()
	b06AddMana(me, "W", "U")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{AlternativeCost: mtmte, Strict: true}); err != nil {
		t.Fatalf("cast Goldbug converted: %v", err)
	}
	passPriorityAroundTable(t, g)
}

func goldbugNow(t *testing.T, g *game.Game, id uuid.UUID) game.Card {
	t.Helper()
	var c game.Card
	var ok bool
	g.ReadSnapshot(func() { c, ok = battlefieldCard(g, id) })
	if !ok {
		t.Fatal("Goldbug is not on the battlefield")
	}
	return c
}

// CR 702.162a, 712.11a, 712.8c, 712.13: the converted cast is the back
// face on the stack for {W}{U}, with the front face's mana value, and
// the permanent enters back face up.
func TestGoldbugCastConvertedFromHand(t *testing.T) {
	g := newCatalogGame(t)
	me, id := goldbugToMain(t, g)
	b06AddMana(me, "W", "U")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{AlternativeCost: mtmte, Strict: true}); err != nil {
		t.Fatalf("cast Goldbug converted: %v", err)
	}
	if len(me.ManaPool) != 0 {
		t.Errorf("the converted cast left %v floating, want {W}{U} spent", me.ManaPool)
	}
	zone, spell := cardWhere(g, me, id)
	if zone != "stack" || spell.ActiveFace != 1 || spell.Name != "Goldbug, Scrappy Scout" {
		t.Fatalf("the spell is in %s as face %d (%q), want the back face on the stack", zone, spell.ActiveFace, spell.Name)
	}
	g.ReadSnapshot(func() {
		if mv, ok := g.ManaValueForEffect(spell); !ok || mv != 3 {
			t.Errorf("the converted spell's mana value = (%d, %v), want 3 from {1}{W}{U} (CR 712.8c)", mv, ok)
		}
	})
	passPriorityAroundTable(t, g)
	c := goldbugNow(t, g, id)
	if c.ActiveFace != 1 || c.Name != "Goldbug, Scrappy Scout" {
		t.Fatalf("Goldbug entered as face %d (%q), want the back face", c.ActiveFace, c.Name)
	}
}

// The printed cast is still the front face for its printed cost.
func TestGoldbugPrintedCastEntersFrontFaceUp(t *testing.T) {
	g := newCatalogGame(t)
	me, id := goldbugToMain(t, g)
	b06AddMana(me, "W", "U")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true}); err == nil {
		t.Fatal("the printed {1}{W}{U} cast was paid with {W}{U}")
	}
	b06AddMana(me, "C")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("printed cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if c := goldbugNow(t, g, id); c.ActiveFace != 0 || c.Name != "Goldbug, Humanity's Ally" {
		t.Errorf("the printed cast entered as face %d (%q), want the front face", c.ActiveFace, c.Name)
	}
}

// From the command zone, with the commander tax on top of {W}{U}
// (CR 903.8): the tax applies whatever price is paid.
func TestGoldbugCastConvertedFromTheCommandZonePaysTheTax(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	commander := deck.ToGameCard(goldbugRow(), true)
	commander.Owner, commander.Controller = me.ID, me.ID
	me.Command.PushTop(commander)
	id := commander.InstanceID
	advanceToMain(t, g)
	me.ManaPool = nil

	b06AddMana(me, "W", "U")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{FromZone: "command", AlternativeCost: mtmte, Strict: true}); err != nil {
		t.Fatalf("first converted cast from the command zone: %v", err)
	}
	if zone, spell := cardWhere(g, me, id); zone != "stack" || spell.ActiveFace != 1 {
		t.Fatalf("the commander is in %s as face %d, want the back face on the stack", zone, spell.ActiveFace)
	}
	passPriorityAroundTable(t, g)
	if c := goldbugNow(t, g, id); c.ActiveFace != 1 {
		t.Fatalf("the commander entered as face %d, want the back face", c.ActiveFace)
	}

	// Back to the command zone; the second cast owes {2} more.
	g.WithWriteLock(func() {
		if _, err := game.MoveCard(g.Battlefield, me.Command, id); err != nil {
			t.Fatalf("MoveCard: %v", err)
		}
	})
	me.ManaPool = nil
	b06AddMana(me, "W", "U")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{FromZone: "command", AlternativeCost: mtmte, Strict: true}); err == nil {
		t.Fatal("the second converted cast skipped the commander tax")
	}
	b06AddMana(me, "C", "C")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{FromZone: "command", AlternativeCost: mtmte, Strict: true}); err != nil {
		t.Fatalf("second converted cast with the tax: %v", err)
	}
	if len(me.ManaPool) != 0 {
		t.Errorf("the taxed cast left %v floating, want {2}{W}{U} spent", me.ManaPool)
	}
}

// CR 702.161a: during its controller's turn the Vehicle is an artifact
// creature with its printed 1/3, and on an opponent's turn it is not a
// creature at all. A +1/+1 counter shows on its own turn.
func TestGoldbugLivingMetalIsACreatureOnlyOnItsControllersTurn(t *testing.T) {
	g := newCatalogGame(t)
	me, id := goldbugToMain(t, g)
	castGoldbugConverted(t, g, me, id)
	g.WithWriteLock(func() {
		if err := g.AddCounterForEffect(id, game.CounterPlusOne, 1); err != nil {
			t.Fatalf("AddCounterForEffect: %v", err)
		}
	})

	g.WithWriteLock(func() { g.RecomputeLayersIfStaleLocked() })
	c := goldbugNow(t, g, id)
	if !c.IsCreature() || !c.IsArtifact() {
		t.Fatalf("on its controller's turn Goldbug, Scrappy Scout is %q, want an artifact creature", c.Effective().Types)
	}
	if p, tough := c.CurrentPower(), c.CurrentToughness(); p != 2 || tough != 4 {
		t.Errorf("on its controller's turn it is %d/%d, want 2/4 (1/3 printed, one +1/+1 counter)", p, tough)
	}
	// Cast this turn, it is summoning sick (CR 302.6).
	advanceTo(t, g, game.StepDeclareAttackers)
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	if err := g.DeclareAttacker(id, opp.ID); err == nil {
		t.Error("a Goldbug cast converted this turn attacked without haste")
	}

	mine := g.Turn.ActiveSeat
	for i := 0; i < 200 && g.Turn.ActiveSeat == mine; i++ {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	g.WithWriteLock(func() { g.RecomputeLayersIfStaleLocked() })
	c = goldbugNow(t, g, id)
	if c.IsCreature() {
		t.Errorf("on an opponent's turn Goldbug, Scrappy Scout is still a creature: %q", c.Effective().Types)
	}
	if !c.IsArtifact() {
		t.Errorf("on an opponent's turn it stopped being an artifact: %q", c.Effective().Types)
	}
}

// Goldbug's second-spell trigger converts it.
func TestGoldbugSecondSpellConvertsIt(t *testing.T) {
	g := newCatalogGame(t)
	me, id := goldbugToMain(t, g)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast Goldbug: %v", err)
	}
	passPriorityAroundTable(t, g)
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	castCatalogSpell(t, g, "Lightning Bolt", "Instant", "4457ed35-7c10-48c8-9776-456485fdf070",
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	passPriorityAroundTable(t, g)
	if c := goldbugNow(t, g, id); c.ActiveFace != 1 || c.TurnOvers != 1 {
		t.Fatalf("after the second spell Goldbug is face %d with %d turn-overs, want converted once", c.ActiveFace, c.TurnOvers)
	}
	var said bool
	for _, ev := range g.Events {
		if ev.Kind == game.EventTransform && ev.CardID == id && ev.Converted {
			said = true
		}
	}
	if !said {
		t.Error("no EventTransform marked Converted")
	}
}

// CR 701.28e: Goldbug's "convert Goldbug" is ignored once Goldbug has
// already turned over since the trigger went on the stack — once, or
// twice back to the same face.
func TestGoldbugConvertTriggerIsIgnoredAfterAnotherTurnOver(t *testing.T) {
	for _, flips := range []int{1, 2} {
		g := newCatalogGame(t)
		me, id := goldbugToMain(t, g)
		if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
			t.Fatalf("cast Goldbug: %v", err)
		}
		passPriorityAroundTable(t, g)
		opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
		castCatalogSpell(t, g, "Lightning Bolt", "Instant", "4457ed35-7c10-48c8-9776-456485fdf070",
			[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
		// In response, something else turns Goldbug over.
		g.WithWriteLock(func() {
			for i := 0; i < flips; i++ {
				if err := g.ConvertPermanentForEffect(id); err != nil {
					t.Fatalf("convert in response: %v", err)
				}
			}
		})
		passPriorityAroundTable(t, g)
		c := goldbugNow(t, g, id)
		if c.TurnOvers != flips || c.ActiveFace != flips%2 {
			t.Errorf("%d flips in response: Goldbug turned over %d times and shows face %d; the trigger should have been ignored",
				flips, c.TurnOvers, c.ActiveFace)
		}
	}
}

// "Prevent all combat damage that would be dealt to attacking Humans
// you control": a Human attacker takes none, a non-Human attacker does.
func TestGoldbugPreventsCombatDamageToAttackingHumans(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushCatalogPermanent(g, me.ID, "Goldbug, Humanity's Ally", "Legendary Artifact Creature — Robot", goldbugOracleID, false)
	blocker := uuid.New()
	g.Battlefield.PushTop(game.Card{
		InstanceID: blocker, Name: "Blocker", TypeLine: "Creature — Bear",
		Power: 2, Toughness: 2, Owner: opp.ID, Controller: opp.ID,
	})
	attacker := func(typeLine string) uuid.UUID {
		id := uuid.New()
		g.Battlefield.PushTop(game.Card{
			InstanceID: id, Name: typeLine, TypeLine: typeLine,
			Power: 2, Toughness: 5, Owner: me.ID, Controller: me.ID, AttackingTarget: opp.ID,
		})
		return id
	}
	human, bear := attacker("Creature — Human Soldier"), attacker("Creature — Bear")
	for _, id := range []uuid.UUID{human, bear} {
		if err := g.MarkCombatDamage(blocker, id, 2); err != nil {
			t.Fatalf("MarkCombatDamage: %v", err)
		}
	}
	h, _ := battlefieldCard(g, human)
	b, _ := battlefieldCard(g, bear)
	if h.DamageMarked != 0 {
		t.Errorf("the attacking Human took %d, want 0", h.DamageMarked)
	}
	if b.DamageMarked != 2 {
		t.Errorf("the attacking Bear took %d, want 2", b.DamageMarked)
	}
}

// "Human spells you control can't be countered" on the back face.
func TestGoldbugScrappyScoutStopsHumanSpellsBeingCountered(t *testing.T) {
	g := newCatalogGame(t)
	me, id := goldbugToMain(t, g)
	castGoldbugConverted(t, g, me, id)
	human := castCatalogSpell(t, g, "Human Recruit", "Creature — Human Soldier", "", nil)
	g.WithWriteLock(func() { _ = g.CounterTargetForEffect(human) })
	if zone, _ := cardWhere(g, me, human); zone != "stack" {
		t.Errorf("a Human spell under Goldbug, Scrappy Scout was countered (now in %s)", zone)
	}
	passPriorityAroundTable(t, g)
	bear := castCatalogSpell(t, g, "Bear Recruit", "Creature — Bear", "", nil)
	g.WithWriteLock(func() { _ = g.CounterTargetForEffect(bear) })
	if zone, _ := cardWhere(g, me, bear); zone == "stack" {
		t.Error("a non-Human spell could not be countered")
	}
}

// "Whenever Goldbug and at least one Human attack, draw a card and
// convert Goldbug": it attacks as a creature on its own turn (living
// metal), and the trigger needs a Human beside it.
func TestGoldbugScrappyScoutAttacksWithAHumanDrawsAndConverts(t *testing.T) {
	for _, withHuman := range []bool{true, false} {
		g := newCatalogGame(t)
		me, id := goldbugToMain(t, g)
		castGoldbugConverted(t, g, me, id)
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == id {
				g.Battlefield.Cards[i].SummonedThisTurn = false
			}
		}
		human := pushCatalogPermanent(g, me.ID, "Human Soldier", "Creature — Human Soldier", "", false)
		opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
		advanceTo(t, g, game.StepDeclareAttackers)
		if err := g.DeclareAttacker(id, opp.ID); err != nil {
			t.Fatalf("Goldbug, Scrappy Scout can't attack on its own turn: %v", err)
		}
		if withHuman {
			if err := g.DeclareAttacker(human, opp.ID); err != nil {
				t.Fatalf("DeclareAttacker Human: %v", err)
			}
		}
		hand := me.Hand.Size()
		lockInAttacks(t, g)
		passPriorityAroundTable(t, g)
		c := goldbugNow(t, g, id)
		if withHuman {
			if me.Hand.Size() != hand+1 || c.ActiveFace != 0 {
				t.Errorf("with a Human: hand %d → %d, face %d; want a card drawn and Goldbug converted to its front face",
					hand, me.Hand.Size(), c.ActiveFace)
			}
			if c.AttackingTarget != opp.ID {
				t.Error("converting took Goldbug out of combat")
			}
		} else if me.Hand.Size() != hand || c.ActiveFace != 1 {
			t.Errorf("alone: hand %d → %d, face %d; want nothing", hand, me.Hand.Size(), c.ActiveFace)
		}
	}
}

// The keywords are canonical: the badge, and the coverage signal.
func TestTransformersKeywordsAreCanonical(t *testing.T) {
	for _, kw := range []string{"More Than Meets the Eye", "Living metal"} {
		if got, ok := game.CanonicalKeywords(kw); !ok || len(got) != 1 {
			t.Errorf("CanonicalKeywords(%q) = %v, %v", kw, got, ok)
		}
	}
}
