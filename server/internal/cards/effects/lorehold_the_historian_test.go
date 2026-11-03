package effects

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// lorehold_the_historian_test.go — Lorehold, the Historian (#4685): "Each
// instant and sorcery card in your hand has miracle {2}" and the loot in
// each opponent's upkeep, end to end on the real catalog card. The engine
// half is game/miracle.go (GrantedMiracleTrigger, the price on the
// permission); molecule_man_test.go pins the granted keyword's road in
// general, and this file the filter, the {2} price, the source scoping and
// the price riding the trigger.

// The instant every miracle test here draws: {R}{R}{R}, so an empty pool
// or two generic mana tell the {2} miracle and the printed cost apart.
const (
	lhInstantName = "Triple Bolt"
	lhInstantType = "Instant"
	lhInstantCost = "{R}{R}{R}"
)

func loreholdOnTable(g *game.Game, owner uuid.UUID) uuid.UUID {
	return pushCatalogPermanent(g, owner, "Lorehold, the Historian", "Legendary Creature — Elder Dragon", loreholdTheHistorianOracleID, false)
}

// miraclePromptCount is how many reveal prompts `chooser` has open.
func miraclePromptCount(g *game.Game, chooser uuid.UUID) int {
	n := 0
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceTriggerPrompt && c.Chooser == chooser &&
			strings.Contains(c.Reason, "miracle cost") {
			n++
		}
	}
	return n
}

func TestLoreholdTheHistorianIsRegisteredFull(t *testing.T) {
	spec, ok := Lookup(loreholdTheHistorianOracleID)
	if !ok {
		t.Fatal("Lorehold, the Historian is not registered")
	}
	if spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
		t.Errorf("completeness %s, caveats %v", spec.Completeness, spec.Caveats)
	}
	def := buildDef(spec)
	// One printed trigger (the loot) and the granted keyword's reveal.
	reveals, others := 0, 0
	for _, tr := range def.Triggered {
		if tr.Key == game.GrantedMiracleTriggerLabel {
			reveals++
		} else {
			others++
		}
	}
	if reveals != 1 || others != 1 {
		t.Errorf("triggers: %d granted-miracle reveals and %d others, want 1 and 1", reveals, others)
	}
}

// The whole road: the first instant drawn this turn is offered a reveal
// naming {2}; the trigger resolves; the cast costs exactly {2} at instant
// speed, and the printed {R}{R}{R} is not repriced by the live grant.
func TestLoreholdFirstInstantDrawnIsCastForTwo(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[1]
	loreholdOnTable(g, me.ID)
	spell := drawMiracleInOwnDrawStep(t, g, 1, lhInstantName, lhInstantType, lhInstantCost, "")

	prompt := miraclePrompt(g, me.ID)
	if prompt == nil {
		t.Fatal("no reveal prompt for the first instant drawn this turn")
	}
	if !strings.Contains(prompt.Reason, "miracle cost {2}") {
		t.Errorf("prompt %q does not name the miracle cost {2}", prompt.Reason)
	}
	answerLatestTriggerPrompt(t, g, me.ID, true)
	resolveTopWithoutPassingHolder(t, g)
	if !hasMiracleGrantUnder(me, spell, game.AltCostKeyGrantedMiracle) {
		t.Fatal("the resolved trigger granted no miracle cast")
	}

	// An empty pool cannot pay {2}.
	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{Strict: true, AlternativeCost: game.AltCostKeyGrantedMiracle}); err == nil {
		t.Fatal("the miracle {2} cast was accepted with an empty pool")
	}
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{C}{C}"); err != nil {
		t.Fatal(err)
	}
	// Two generic mana is not {R}{R}{R}: the live grant does not reprice
	// the printed cast.
	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{Strict: true}); err == nil {
		t.Fatal("the printed {R}{R}{R} cast was paid with {C}{C} while the miracle grant was live")
	}
	if !me.Hand.Contains(spell) {
		t.Fatal("a refused cast took the card out of hand")
	}
	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{Strict: true, AlternativeCost: game.AltCostKeyGrantedMiracle}); err != nil {
		t.Fatalf("the miracle {2} cast: %v", err)
	}
	if me.Hand.Contains(spell) {
		t.Error("the cast card is still in hand")
	}
	if len(me.ManaPool) != 0 {
		t.Errorf("the {2} cast left %d mana in the pool", len(me.ManaPool))
	}
}

// A sorcery is reached too, and its {2} cast is made in the draw step.
func TestLoreholdFirstSorceryDrawnIsCastForTwoInTheDrawStep(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[1]
	loreholdOnTable(g, me.ID)
	spell := drawMiracleInOwnDrawStep(t, g, 1, "Pyre Sweep", "Sorcery", "{4}{R}{R}", "")
	revealAndResolve(t, g, me)
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{C}{C}"); err != nil {
		t.Fatal(err)
	}
	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{Strict: true, AlternativeCost: game.AltCostKeyGrantedMiracle}); err != nil {
		t.Fatalf("the sorcery's miracle {2} cast in the draw step: %v", err)
	}
}

// "Instant and sorcery card": a creature card drawn first gets nothing.
func TestLoreholdACreatureCardGetsNothing(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[1]
	loreholdOnTable(g, me.ID)
	bear := drawMiracleInOwnDrawStep(t, g, 1, "Grizzly Bears", "Creature — Bear", "{1}{G}", "")
	if miraclePrompt(g, me.ID) != nil {
		t.Error("a creature card was offered a miracle reveal")
	}
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{C}{C}"); err != nil {
		t.Fatal(err)
	}
	if err := g.CastSpell(me.ID, bear, game.CastSpellParams{Strict: true, AlternativeCost: game.AltCostKeyGrantedMiracle}); err == nil {
		t.Error("a creature card was cast for miracle {2}")
	}
}

// "The first card you drew this turn": the instant drawn second is not.
func TestLoreholdTheSecondCardDrawnIsNotOffered(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[1]
	loreholdOnTable(g, me.ID)
	drawMiracleInOwnDrawStep(t, g, 1, "Mountain", "Basic Land — Mountain", "", "")
	for g.Turn.Step != game.StepPrecombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatal(err)
		}
	}
	spell := pushMiracleOnTop(g, me, lhInstantName, lhInstantType, lhInstantCost, "")
	if err := g.DrawCard(me.ID); err != nil {
		t.Fatalf("DrawCard: %v", err)
	}
	if !me.Hand.Contains(spell) {
		t.Fatal("the second draw did not draw the instant")
	}
	if miraclePrompt(g, me.ID) != nil {
		t.Error("the second card drawn this turn was offered a miracle reveal")
	}
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{C}{C}"); err != nil {
		t.Fatal(err)
	}
	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{Strict: true, AlternativeCost: game.AltCostKeyGrantedMiracle}); err == nil {
		t.Error("the second card drawn was cast for miracle {2}")
	}
}

// No Lorehold, no miracle; and an opponent's Lorehold gives me nothing.
// The draw is mine on seat 0's turn, so no upkeep loot is in the way.
func TestLoreholdGoneOrAnOpponentsMeansNoOffer(t *testing.T) {
	for _, tc := range []struct {
		name  string
		place func(t *testing.T, g *game.Game)
	}{
		{"none", func(*testing.T, *game.Game) {}},
		{"opponent's", func(_ *testing.T, g *game.Game) { loreholdOnTable(g, g.Seats[2].ID) }},
		{"destroyed before the draw", func(t *testing.T, g *game.Game) {
			lh := loreholdOnTable(g, g.Seats[1].ID)
			g.WithWriteLock(func() {
				if err := g.DestroyPermanentForEffect(lh); err != nil {
					t.Errorf("destroying Lorehold: %v", err)
				}
			})
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[1]
			tc.place(t, g)
			spell := pushMiracleOnTop(g, me, lhInstantName, lhInstantType, lhInstantCost, "")
			if err := g.DrawCard(me.ID); err != nil {
				t.Fatalf("DrawCard: %v", err)
			}
			if miraclePrompt(g, me.ID) != nil {
				t.Error("a miracle reveal was offered without my Lorehold")
			}
			if err := g.AddManaForEffect(me.ID, uuid.Nil, "{C}{C}"); err != nil {
				t.Fatal(err)
			}
			if err := g.CastSpell(me.ID, spell, game.CastSpellParams{Strict: true, AlternativeCost: game.AltCostKeyGrantedMiracle}); err == nil {
				t.Error("the miracle {2} cast was accepted without my Lorehold")
			}
		})
	}
}

// "You may reveal": a "no" is refused the {2} claim.
func TestLoreholdDeclinedRevealIsRefused(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[1]
	loreholdOnTable(g, me.ID)
	spell := drawMiracleInOwnDrawStep(t, g, 1, lhInstantName, lhInstantType, lhInstantCost, "")
	answerLatestTriggerPrompt(t, g, me.ID, false)
	if !stackFullyEmpty(g) {
		t.Error("a declined reveal put a trigger on the stack")
	}
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{C}{C}"); err != nil {
		t.Fatal(err)
	}
	err := g.CastSpell(me.ID, spell, game.CastSpellParams{Strict: true, AlternativeCost: game.AltCostKeyGrantedMiracle})
	if !errors.Is(err, game.ErrAltCostNotGranted) {
		t.Errorf("miracle cast without a reveal: %v, want ErrAltCostNotGranted", err)
	}
	if !me.Hand.Contains(spell) {
		t.Error("a refused cast took the card out of hand")
	}
}

// The price rides the trigger (CR 113.7a): Lorehold leaving after the card
// is revealed but before the trigger resolves does not take the {2} cast
// away, and the cast still costs {2}, not nothing.
func TestLoreholdLeavingAfterTheRevealKeepsTheCast(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[1]
	lh := loreholdOnTable(g, me.ID)
	spell := drawMiracleInOwnDrawStep(t, g, 1, lhInstantName, lhInstantType, lhInstantCost, "")
	answerLatestTriggerPrompt(t, g, me.ID, true)
	if stackFullyEmpty(g) {
		t.Fatal("the revealed trigger is not on the stack")
	}
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(lh); err != nil {
			t.Errorf("destroying Lorehold: %v", err)
		}
	})
	resolveTopWithoutPassingHolder(t, g)
	if !hasMiracleGrantUnder(me, spell, game.AltCostKeyGrantedMiracle) {
		t.Fatal("the trigger granted no cast once Lorehold had left")
	}
	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{Strict: true, AlternativeCost: game.AltCostKeyGrantedMiracle}); err == nil {
		t.Fatal("the miracle cast with Lorehold gone was accepted for nothing")
	}
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{C}{C}"); err != nil {
		t.Fatal(err)
	}
	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{Strict: true, AlternativeCost: game.AltCostKeyGrantedMiracle}); err != nil {
		t.Fatalf("the miracle {2} cast after Lorehold left: %v", err)
	}
}

// Two Loreholds (the second a nonlegendary copy, as Spark Double makes,
// so the legend rule keeps both) give the card two identical miracles:
// one reveal, one {2}.
func TestLoreholdTwoLoreholdsMakeOneReveal(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[1]
	loreholdOnTable(g, me.ID)
	pushCatalogPermanent(g, me.ID, "Lorehold, the Historian", "Creature — Elder Dragon", loreholdTheHistorianOracleID, false)
	spell := drawMiracleInOwnDrawStep(t, g, 1, lhInstantName, lhInstantType, lhInstantCost, "")
	if n := miraclePromptCount(g, me.ID); n != 1 {
		t.Fatalf("%d reveal prompts under two Loreholds, want 1", n)
	}
	revealAndResolve(t, g, me)
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{C}{C}"); err != nil {
		t.Fatal(err)
	}
	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{Strict: true, AlternativeCost: game.AltCostKeyGrantedMiracle}); err != nil {
		t.Fatalf("the miracle {2} cast under two Loreholds: %v", err)
	}
}

// Each granting permanent fires for the cards ITS static reaches: beside
// Molecule Man ("nonland"), a creature drawn first is revealed once (Molecule
// Man's), and an instant twice (each at its own price).
func TestLoreholdBesideMoleculeManRevealsOnlyWhatEachReaches(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[1]
	loreholdOnTable(g, me.ID)
	moleculeManOnTable(g, me.ID)
	drawMiracleInOwnDrawStep(t, g, 1, "Grizzly Bears", "Creature — Bear", "{1}{G}", "")
	if n := miraclePromptCount(g, me.ID); n != 1 {
		t.Errorf("a creature drawn under Lorehold and Molecule Man raised %d reveals, want 1", n)
	}
	if p := miraclePrompt(g, me.ID); p != nil && !strings.Contains(p.Reason, "{0}") {
		t.Errorf("the creature's reveal %q is not Molecule Man's {0}", p.Reason)
	}

	g2 := newCatalogGame(t)
	me2 := g2.Seats[1]
	loreholdOnTable(g2, me2.ID)
	moleculeManOnTable(g2, me2.ID)
	drawMiracleInOwnDrawStep(t, g2, 1, lhInstantName, lhInstantType, lhInstantCost, "")
	if n := miraclePromptCount(g2, me2.ID); n != 2 {
		t.Errorf("an instant drawn under Lorehold and Molecule Man raised %d reveals, want 2", n)
	}
}

// "On any player's turn": the first instant I draw in an opponent's turn
// is offered, and cast there for {2}.
func TestLoreholdOnAnOpponentsTurn(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[1] // seat 0 is active
	loreholdOnTable(g, me.ID)
	spell := pushMiracleOnTop(g, me, lhInstantName, lhInstantType, lhInstantCost, "")
	if err := g.DrawCard(me.ID); err != nil {
		t.Fatalf("DrawCard: %v", err)
	}
	revealAndResolve(t, g, me)
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{C}{C}"); err != nil {
		t.Fatal(err)
	}
	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{Strict: true, AlternativeCost: game.AltCostKeyGrantedMiracle}); err != nil {
		t.Fatalf("the miracle {2} cast on an opponent's turn: %v", err)
	}
}

// The bot's move list and the view see the granted miracle: nothing before
// the reveal, the {2} cast after it.
func TestLoreholdEnumeratorAndViewSeeTheGrantedMiracle(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[1]
	loreholdOnTable(g, me.ID)
	spell := drawMiracleInOwnDrawStep(t, g, 1, lhInstantName, lhInstantType, lhInstantCost, "")
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{C}{C}"); err != nil {
		t.Fatal(err)
	}
	miracleMoves := func() int {
		n := 0
		for _, m := range legal.EnumerateFor(g, me.ID) {
			if m.Type != "cast_spell" || m.Source != spell {
				continue
			}
			var p struct {
				AlternativeCost string `json:"alternative_cost"`
			}
			if err := json.Unmarshal(m.Params, &p); err != nil {
				t.Fatalf("move params: %v", err)
			}
			if p.AlternativeCost == game.AltCostKeyGrantedMiracle {
				n++
			}
		}
		return n
	}
	if n := miracleMoves(); n != 0 {
		t.Fatalf("the miracle cast was offered before any reveal: %d moves", n)
	}
	answerLatestTriggerPrompt(t, g, me.ID, true)
	resolveTopWithoutPassingHolder(t, g)
	if n := miracleMoves(); n == 0 {
		t.Error("the enumerator did not offer the {2} miracle cast with the grant live")
	}

	view := protocol.ViewOfGameFor(g, me.ID.String())
	found := false
	for _, seat := range view.Seats {
		if seat.ID != me.ID.String() {
			continue
		}
		for _, c := range seat.Hand.Cards {
			if c.InstanceID != spell.String() {
				continue
			}
			for _, ac := range c.AlternativeCosts {
				if ac.Key == game.AltCostKeyGrantedMiracle && !ac.TimingClosed {
					found = true
				}
			}
		}
	}
	if !found {
		t.Error("the view does not list the live {2} miracle offer on the revealed card")
	}
}

// --- the loot -------------------------------------------------------------

// "At the beginning of each opponent's upkeep, you may discard a card. If
// you do, draw a card": in each opponent's upkeep, a discard is a draw.
func TestLoreholdLootsInEachOpponentsUpkeep(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[1]
	loreholdOnTable(g, me.ID)
	pushHandCardOfType(me, "Fodder", "Instant")
	pushHandCardOfType(me, "Fodder", "Instant")
	for _, opp := range []int{2, 3} {
		advanceToUpkeepOfSeat(t, g, opp)
		passPriorityAroundTable(t, g)
		hand, library, yard := me.Hand.Size(), me.Library.Size(), me.Graveyard.Size()
		if discardChoiceFor(g, me.ID) == nil {
			t.Fatalf("seat %d's upkeep: no discard offer", opp)
		}
		discardFromHand(t, g, me.ID)
		passPriorityAroundTable(t, g)
		if me.Graveyard.Size() != yard+1 || me.Library.Size() != library-1 || me.Hand.Size() != hand {
			t.Errorf("seat %d's upkeep: hand %d→%d, library %d→%d, graveyard %d→%d; want a discard and a draw",
				opp, hand, me.Hand.Size(), library, me.Library.Size(), yard, me.Graveyard.Size())
		}
	}
}

// "If you do": declining draws nothing.
func TestLoreholdDecliningTheLootDrawsNothing(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[1]
	loreholdOnTable(g, me.ID)
	pushHandCardOfType(me, "Fodder", "Instant")
	advanceToUpkeepOfSeat(t, g, 2)
	passPriorityAroundTable(t, g)
	hand, library := me.Hand.Size(), me.Library.Size()
	answerDiscard(t, g, me.ID)
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand || me.Library.Size() != library {
		t.Errorf("declining: hand %d→%d, library %d→%d; want nothing", hand, me.Hand.Size(), library, me.Library.Size())
	}
}

// "Each opponent's upkeep", not mine.
func TestLoreholdDoesNotLootInYourOwnUpkeep(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[1]
	loreholdOnTable(g, me.ID)
	pushHandCardOfType(me, "Fodder", "Instant")
	advanceToUpkeepOfSeat(t, g, 1)
	passPriorityAroundTable(t, g)
	if discardChoiceFor(g, me.ID) != nil {
		t.Error("Lorehold offered its loot in its controller's own upkeep")
	}
}
