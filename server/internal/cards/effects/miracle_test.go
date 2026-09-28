package effects

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// miracle_test.go — miracle (CR 702.94, #1665), end to end on real
// catalog cards. The keyword's engine half is game/miracle.go; what is
// pinned here is the whole road a player walks: draw it first, answer
// "Reveal Terminus for its miracle cost {W}?", let the trigger resolve,
// cast it for {W} — and every way that road is closed.

const (
	terminusOracle         = "3dd196b6-a85a-4e3e-bb57-ec34241f8117"
	thunderousWrathOracle  = "78260893-c443-44c8-ab45-ce86ef347d98"
	devastationTideOracle  = "4245ee98-2d4c-49d1-8d07-80760cae2bf9"
	reforgeTheSoulOracle   = "ece854f8-8c60-4f30-894f-2286d3dd61b9"
	entreatTheAngelsOracle = "b349f018-c20b-48b0-9e65-d5fd56b24b88"
	banishingStrokeOracle  = "a6898364-c29e-4b97-a500-344efa3ec24a"
)

// pushMiracleOnTop puts a catalog card on top of p's library, so the
// next draw is it.
func pushMiracleOnTop(g *game.Game, p *game.Player, name, typeLine, manaCost, oracle string) uuid.UUID {
	id := uuid.New()
	g.WithWriteLock(func() {
		p.Library.PushTop(game.Card{
			InstanceID: id,
			Name:       name,
			TypeLine:   typeLine,
			ManaCost:   manaCost,
			OracleID:   oracle,
			Owner:      p.ID,
			Controller: p.ID,
		})
	})
	return id
}

// miraclePrompt is the pending "Reveal X for its miracle cost?" prompt
// addressed to chooser, or nil.
func miraclePrompt(g *game.Game, chooser uuid.UUID) *game.PendingChoice {
	for i := len(g.PendingChoices) - 1; i >= 0; i-- {
		c := g.PendingChoices[i]
		if c != nil && c.Kind == game.PendingChoiceTriggerPrompt && c.Chooser == chooser {
			return c
		}
	}
	return nil
}

// hasMiracleGrant reports whether p holds a miracle permission naming
// cardID AS IT SITS IN p's HAND NOW — the instance and its CR 400.7
// epoch. A permission over an object that has since left is not one.
func hasMiracleGrant(p *game.Player, cardID uuid.UUID) bool {
	epoch := -1
	for _, c := range p.Hand.Cards {
		if c.InstanceID == cardID {
			epoch = c.ObjectEpoch
		}
	}
	if epoch < 0 {
		return false
	}
	for _, perm := range p.CastPermissions {
		if perm.Zone != game.ZoneHand || perm.AltCostKey != game.AltCostKeyMiracle {
			continue
		}
		for _, ref := range perm.Cards {
			if ref.ID == cardID && ref.Epoch == epoch {
				return true
			}
		}
	}
	return false
}

// drawMiracleInOwnDrawStep walks to seat's upkeep, stacks the card on
// top, and enters the draw step so the turn-based draw is the first
// card seat draws this turn.
func drawMiracleInOwnDrawStep(t *testing.T, g *game.Game, seat int, name, typeLine, manaCost, oracle string) uuid.UUID {
	t.Helper()
	advanceToUpkeepOf(t, g, seat)
	passPriorityAroundTable(t, g)
	p := g.Seats[seat]
	id := pushMiracleOnTop(g, p, name, typeLine, manaCost, oracle)
	for g.Turn.Step != game.StepDraw {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep to the draw step: %v", err)
		}
	}
	if !p.Hand.Contains(id) {
		t.Fatalf("%s was not drawn in the draw step", name)
	}
	return id
}

// revealAndResolve answers the reveal prompt "yes" and resolves the
// miracle trigger.
func revealAndResolve(t *testing.T, g *game.Game, p *game.Player) {
	t.Helper()
	if miraclePrompt(g, p.ID) == nil {
		t.Fatal("no \"Reveal … for its miracle cost?\" prompt for the first card drawn this turn")
	}
	answerLatestTriggerPrompt(t, g, p.ID, true)
	resolveTopWithoutPassingHolder(t, g)
}

// resolveTopWithoutPassingHolder resolves the miracle trigger. Passing
// priority is the decline once the grant exists (closeMiracleWindow),
// so the passes are made while the trigger is still on the stack and
// the resolving pass is the last one — which is how the table plays it.
func resolveTopWithoutPassingHolder(t *testing.T, g *game.Game) {
	t.Helper()
	for i := 0; i < 8 && !stackFullyEmpty(g); i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	if !stackFullyEmpty(g) {
		t.Fatal("the miracle trigger never resolved")
	}
}

// The whole road, in the owner's own draw step: prompt, reveal, grant,
// a SORCERY cast for {W} in the draw step (CR 608.2g — timing ignored),
// and Terminus doing its job.
func TestMiracleTerminusFirstDrawIsCastForItsMiracleCost(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[1]
	them := g.Seats[2]
	bear := pushPermanentForTest(g, them.ID, "Bear", "", "Creature — Bear")
	term := drawMiracleInOwnDrawStep(t, g, 1, "Terminus", "Sorcery", "{4}{W}{W}", terminusOracle)

	revealAndResolve(t, g, me)
	// CR 702.94a "reveal": the table has seen it.
	for i := range me.Hand.Cards {
		if c := &me.Hand.Cards[i]; c.InstanceID == term && !c.IsKnownTo(them.ID) {
			t.Error("the revealed miracle card is not known to the other seats (CR 701.20)")
		}
	}
	if !hasMiracleGrant(me, term) {
		t.Fatal("the resolved miracle trigger granted no cast")
	}
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{W}"); err != nil {
		t.Fatal(err)
	}
	if err := g.CastSpell(me.ID, term, game.CastSpellParams{
		Strict:          true,
		AlternativeCost: game.AltCostKeyMiracle,
	}); err != nil {
		t.Fatalf("the miracle cast in the draw step: %v", err)
	}
	if len(me.ManaPool) != 0 {
		t.Errorf("mana left after the cast: %d, want 0 — {W} is the whole price", len(me.ManaPool))
	}
	if hasMiracleGrant(me, term) {
		t.Error("the grant outlived the cast (CR 400.7)")
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) {
		t.Error("Terminus left the creature on the battlefield")
	}
	if !them.Library.Contains(bear) {
		t.Error("the creature is not in its owner's library")
	}
}

// "No" is an ordinary card in an ordinary hand: no trigger, no grant,
// and the miracle price refused.
func TestMiracleDeclinedRevealLeavesTheCardInHand(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[1]
	term := drawMiracleInOwnDrawStep(t, g, 1, "Terminus", "Sorcery", "{4}{W}{W}", terminusOracle)
	answerLatestTriggerPrompt(t, g, me.ID, false)
	if !stackFullyEmpty(g) {
		t.Error("a declined reveal put the trigger on the stack")
	}
	if !me.Hand.Contains(term) || hasMiracleGrant(me, term) {
		t.Fatal("a declined reveal must leave the card in hand with no grant")
	}
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{W}"); err != nil {
		t.Fatal(err)
	}
	err := g.CastSpell(me.ID, term, game.CastSpellParams{Strict: true, AlternativeCost: game.AltCostKeyMiracle})
	if !errors.Is(err, game.ErrAltCostNotGranted) {
		t.Fatalf("miracle cast without a reveal: %v, want ErrAltCostNotGranted", err)
	}
}

// The second card drawn in a turn is not a miracle (CR 702.94a).
func TestMiracleSecondDrawOfTheTurnOffersNothing(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[1]
	advanceToUpkeepOf(t, g, 1)
	passPriorityAroundTable(t, g)
	for g.Turn.Step != game.StepPrecombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatal(err)
		}
	}
	// The draw step's card was the first; this is the second.
	term := pushMiracleOnTop(g, me, "Terminus", "Sorcery", "{4}{W}{W}", terminusOracle)
	if err := g.DrawCard(me.ID); err != nil {
		t.Fatalf("DrawCard: %v", err)
	}
	if !me.Hand.Contains(term) {
		t.Fatal("the second draw did not draw Terminus")
	}
	if miraclePrompt(g, me.ID) != nil {
		t.Error("the second card drawn this turn was offered a miracle reveal")
	}
}

// A draw on ANOTHER player's turn counts when it is your first card
// that turn — the reason Terminus is played. The cast is a sorcery on
// an opponent's turn, which only the grant's timing allows; the same
// card hard-cast at the same moment is refused (ForClaim).
func TestMiracleOnAnOpponentsTurnIgnoresTimingForTheMiracleCastOnly(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[1] // seat 0 is active: it is their draw step
	term := pushMiracleOnTop(g, me, "Terminus", "Sorcery", "{4}{W}{W}", terminusOracle)
	if err := g.DrawCard(me.ID); err != nil {
		t.Fatalf("DrawCard: %v", err)
	}
	revealAndResolve(t, g, me)
	if !hasMiracleGrant(me, term) {
		t.Fatal("a first draw on an opponent's turn granted no miracle cast")
	}
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{C}{C}{C}{C}{W}{W}"); err != nil {
		t.Fatal(err)
	}
	// The printed cost is still a sorcery.
	err := g.CastSpell(me.ID, term, game.CastSpellParams{Strict: true})
	if !errors.Is(err, game.ErrSorcerySpeedRequired) {
		t.Fatalf("hard-casting Terminus on an opponent's turn: %v, want ErrSorcerySpeedRequired — the miracle grant opens its own claim only", err)
	}
	if err := g.CastSpell(me.ID, term, game.CastSpellParams{Strict: true, AlternativeCost: game.AltCostKeyMiracle}); err != nil {
		t.Fatalf("the miracle cast on an opponent's turn: %v", err)
	}
}

// #1686: the wire must tell the client's cost picker which offer is
// timing-open RIGHT NOW, so it never shows "Its mana cost" next to
// "Miracle {W}" only to have the printed choice fail at announce with
// ErrSorcerySpeedRequired. A live miracle grant opens the miracle
// claim at instant speed and has nothing to say about the printed one
// (CastPermission.ForClaim) — the printed {4}{W}{W} sorcery is still a
// sorcery, so on another player's turn the view must stamp
// `printed_cost_timing_closed` and must NOT stamp `timing_closed` on
// the miracle offer.
//
// This stays a separate question from CastOffersForLocked's own list
// (still unfiltered — protocol.TestGrantedOfferInASharedZoneIsFilteredByItsLifeComponent
// pins why): the exile strip's informational cast_prices badge reads
// that list without caring whether now is the moment, and the picker
// needs a second, timing-aware signal beside it rather than a
// narrower version of it.
func TestMiracleViewFlagsThePrintedCostAsTimingClosed(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[1] // seat 0 is active: it is their draw step
	term := pushMiracleOnTop(g, me, "Terminus", "Sorcery", "{4}{W}{W}", terminusOracle)
	if err := g.DrawCard(me.ID); err != nil {
		t.Fatalf("DrawCard: %v", err)
	}
	revealAndResolve(t, g, me)
	if !hasMiracleGrant(me, term) {
		t.Fatal("a first draw on an opponent's turn granted no miracle cast")
	}

	view := protocol.ViewOfGameFor(g, me.ID.String())
	var card *protocol.CardView
	for _, seat := range view.Seats {
		if seat.ID != me.ID.String() {
			continue
		}
		for i := range seat.Hand.Cards {
			if seat.Hand.Cards[i].InstanceID == term.String() {
				card = &seat.Hand.Cards[i]
			}
		}
	}
	if card == nil {
		t.Fatal("Terminus is missing from its owner's own hand view")
	}
	if !card.PrintedCostTimingClosed {
		t.Error("the printed sorcery cost is not flagged timing-closed on another player's turn (#1686)")
	}
	if card.AlternativeCostRequired {
		t.Error("alternative_cost_required should stay false — the printed cost IS one of the zone's prices, just not right now")
	}
	found := false
	for _, ac := range card.AlternativeCosts {
		if ac.Key != game.AltCostKeyMiracle {
			continue
		}
		found = true
		if ac.TimingClosed {
			t.Error("the miracle offer itself must not be flagged timing-closed while its grant is live")
		}
	}
	if !found {
		t.Fatal("the miracle offer is missing from the view while the grant is live")
	}
}

// The mirror image, and the one that proves this is additive rather
// than a new noise field on every card: an ORDINARY sorcery — no
// miracle keyword, nothing granted — sitting in its owner's hand
// during their own precombat main phase (CR 307.1's actual sorcery-
// speed window) must carry neither flag. Sorcery speed is not "your
// turn" alone (the draw step isn't it either, which is exactly why a
// sorcery miracle's printed cost stays closed even when the reveal
// happens on the card's own owner's turn) — it is specifically a main
// phase with an empty stack, which this test puts the card in.
func TestMiracleViewLeavesAnOrdinarySorceryUnflaggedInMainPhase(t *testing.T) {
	g := newCatalogGame(t)
	active := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	active.Hand.PushTop(game.Card{
		InstanceID: id,
		Name:       "Ordinary Sorcery",
		TypeLine:   "Sorcery",
		ManaCost:   "{2}{W}",
		OracleID:   "test-ordinary-sorcery-1686",
		Owner:      active.ID,
		Controller: active.ID,
	})
	for g.Turn.Step != game.StepPrecombatMain && g.Turn.Step != game.StepPostcombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}

	view := protocol.ViewOfGameFor(g, active.ID.String())
	var card *protocol.CardView
	for _, seat := range view.Seats {
		if seat.ID != active.ID.String() {
			continue
		}
		for i := range seat.Hand.Cards {
			if seat.Hand.Cards[i].InstanceID == id.String() {
				card = &seat.Hand.Cards[i]
			}
		}
	}
	if card == nil {
		t.Fatal("the ordinary sorcery is missing from its owner's own hand view")
	}
	if card.PrintedCostTimingClosed {
		t.Error("an ordinary sorcery in its owner's main phase must not be flagged timing-closed")
	}
	if card.AlternativeCostRequired {
		t.Error("an ordinary sorcery with no alternative cost must not require one")
	}
}

// CR 702.94b: a card that left the hand before its trigger resolved is
// not the card that was revealed — even when it has come back.
func TestMiracleCardThatLeftTheHandIsNotCast(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[1]
	term := drawMiracleInOwnDrawStep(t, g, 1, "Terminus", "Sorcery", "{4}{W}{W}", terminusOracle)
	answerLatestTriggerPrompt(t, g, me.ID, true)
	// Out and back in response: a new object (CR 400.7).
	g.WithWriteLock(func() {
		if _, err := game.MoveCard(me.Hand, me.Graveyard, term); err != nil {
			t.Fatal(err)
		}
		if _, err := game.MoveCard(me.Graveyard, me.Hand, term); err != nil {
			t.Fatal(err)
		}
	})
	resolveTopWithoutPassingHolder(t, g)
	if hasMiracleGrant(me, term) {
		t.Fatal("the trigger granted a cast to a card that left the hand before it resolved")
	}
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{W}"); err != nil {
		t.Fatal(err)
	}
	err := g.CastSpell(me.ID, term, game.CastSpellParams{Strict: true, AlternativeCost: game.AltCostKeyMiracle})
	if !errors.Is(err, game.ErrAltCostNotGranted) {
		t.Fatalf("miracle cast of a card that left the hand: %v, want ErrAltCostNotGranted", err)
	}
}

// Passing priority is the decline: the window closes, and the card is
// back to a {4}{W}{W} sorcery.
func TestMiracleWindowClosesWhenTheHolderPasses(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[1]
	term := drawMiracleInOwnDrawStep(t, g, 1, "Terminus", "Sorcery", "{4}{W}{W}", terminusOracle)
	revealAndResolve(t, g, me)
	if !hasMiracleGrant(me, term) {
		t.Fatal("no grant after the trigger resolved")
	}
	if err := g.PassPriority(); err != nil {
		t.Fatalf("PassPriority: %v", err)
	}
	if hasMiracleGrant(me, term) {
		t.Fatal("the miracle grant survived its holder passing priority")
	}
}

// #1686: the sandbox's skip-ahead is documented as "pass priority
// until this step ends," so a manual advance must decline a live
// miracle window exactly as an explicit PassPriority does — not leave
// it open until the turn's cleanup sweeps it.
func TestMiracleWindowClosesOnManualStepAdvance(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[1]
	term := drawMiracleInOwnDrawStep(t, g, 1, "Terminus", "Sorcery", "{4}{W}{W}", terminusOracle)
	revealAndResolve(t, g, me)
	if !hasMiracleGrant(me, term) {
		t.Fatal("no grant after the trigger resolved")
	}
	if _, err := g.AdvanceStep(); err != nil {
		t.Fatalf("AdvanceStep: %v", err)
	}
	if hasMiracleGrant(me, term) {
		t.Fatal("the miracle grant survived a manual step advance (#1686)")
	}
}

// The grant is pure data: a game captured with the window open and
// restored still casts the card for its miracle cost, and the undo
// clone does too.
func TestMiracleGrantSurvivesSnapshotAndClone(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[1]
	term := drawMiracleInOwnDrawStep(t, g, 1, "Terminus", "Sorcery", "{4}{W}{W}", terminusOracle)
	revealAndResolve(t, g, me)

	blob, err := json.Marshal(g.CaptureSnapshot())
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var round game.GameSnapshot
	if err := json.Unmarshal(blob, &round); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	restored, err := round.Restore()
	if err != nil {
		t.Fatalf("restore: %v", err)
	}
	clone := g.Clone()
	for name, gg := range map[string]*game.Game{"restored": restored, "clone": clone} {
		p := gg.Seats[1]
		if !hasMiracleGrant(p, term) {
			t.Errorf("%s: the miracle grant was lost", name)
			continue
		}
		if err := gg.AddManaForEffect(p.ID, uuid.Nil, "{W}"); err != nil {
			t.Fatal(err)
		}
		if err := gg.CastSpell(p.ID, term, game.CastSpellParams{Strict: true, AlternativeCost: game.AltCostKeyMiracle}); err != nil {
			t.Errorf("%s: the miracle cast: %v", name, err)
		}
	}
}

// A trigger captured on the stack survives a restore too: its body is
// a registered key (miracle/offer), not a closure.
func TestMiracleTriggerOnTheStackSurvivesARestore(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[1]
	term := drawMiracleInOwnDrawStep(t, g, 1, "Terminus", "Sorcery", "{4}{W}{W}", terminusOracle)
	answerLatestTriggerPrompt(t, g, me.ID, true)
	if stackFullyEmpty(g) {
		t.Fatal("the miracle trigger is not on the stack")
	}
	blob, err := json.Marshal(g.CaptureSnapshot())
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var round game.GameSnapshot
	if err := json.Unmarshal(blob, &round); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	restored, err := round.Restore()
	if err != nil {
		t.Fatalf("restore: %v", err)
	}
	resolveTopWithoutPassingHolder(t, restored)
	if !hasMiracleGrant(restored.Seats[1], term) {
		t.Fatal("the restored miracle trigger granted no cast")
	}
}

// The bot is offered the miracle cast when the grant is live — and
// only the miracle cast: the printed-cost sorcery is not castable in
// the draw step.
func TestMiracleEnumeratorOffersTheMiracleCastOnly(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[1]
	term := drawMiracleInOwnDrawStep(t, g, 1, "Terminus", "Sorcery", "{4}{W}{W}", terminusOracle)

	castsOf := func() (miracle, printed int) {
		for _, m := range legal.EnumerateFor(g, me.ID) {
			if m.Type != "cast_spell" || m.Source != term {
				continue
			}
			var p struct {
				AlternativeCost string `json:"alternative_cost"`
			}
			if err := json.Unmarshal(m.Params, &p); err != nil {
				t.Fatalf("move params: %v", err)
			}
			if p.AlternativeCost == game.AltCostKeyMiracle {
				miracle++
			} else {
				printed++
			}
		}
		return miracle, printed
	}
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{C}{C}{C}{C}{W}{W}"); err != nil {
		t.Fatal(err)
	}
	if m, _ := castsOf(); m != 0 {
		t.Fatalf("the miracle cast was offered before any reveal: %d moves", m)
	}
	answerLatestTriggerPrompt(t, g, me.ID, true)
	resolveTopWithoutPassingHolder(t, g)
	m, p := castsOf()
	if m == 0 {
		t.Error("the enumerator did not offer the miracle cast with the grant live")
	}
	if p != 0 {
		t.Errorf("the enumerator offered %d printed-cost casts of a sorcery in the draw step", p)
	}
}

// Every catalogued miracle card grows the keyword's trigger from its
// Miracle cost, once, watching draws from the hand.
func TestMiracleCardsGrowTheTrigger(t *testing.T) {
	for name, oracle := range map[string]string{
		"Terminus":            terminusOracle,
		"Thunderous Wrath":    thunderousWrathOracle,
		"Devastation Tide":    devastationTideOracle,
		"Reforge the Soul":    reforgeTheSoulOracle,
		"Entreat the Angels":  entreatTheAngelsOracle,
		"Banishing Stroke":    banishingStrokeOracle,
		"Blessings of Nature": blessingsOfNatureOracle,
	} {
		n := 0
		for _, tr := range game.CatalogTriggers(oracle) {
			if tr.Keyword == game.AltCostKeyMiracle && game.TriggerWatchesFromZone(tr, game.ZoneHand) {
				n++
			}
		}
		if n != 1 {
			t.Errorf("%s: %d miracle triggers, want 1", name, n)
		}
		ac := game.AlternativeCostByKey(oracle, game.AltCostKeyMiracle)
		if ac == nil || !ac.RequiresGrant {
			t.Errorf("%s: no Miracle cost gated on the grant", name)
		}
	}
}

// --- the cards ----------------------------------------------------

// castMiracleDrawn draws `name` as seat 1's first card of its turn,
// reveals it, resolves the trigger and casts it for its miracle cost
// with `mana` in the pool.
func castMiracleDrawn(t *testing.T, g *game.Game, name, typeLine, manaCost, oracle, mana string, params game.CastSpellParams) uuid.UUID {
	t.Helper()
	me := g.Seats[1]
	id := drawMiracleInOwnDrawStep(t, g, 1, name, typeLine, manaCost, oracle)
	revealAndResolve(t, g, me)
	if err := g.AddManaForEffect(me.ID, uuid.Nil, mana); err != nil {
		t.Fatal(err)
	}
	params.Strict = true
	params.AlternativeCost = game.AltCostKeyMiracle
	if err := g.CastSpell(me.ID, id, params); err != nil {
		t.Fatalf("%s miracle cast: %v", name, err)
	}
	if len(me.ManaPool) != 0 {
		t.Errorf("%s: %d mana left, want the miracle cost to be the whole price", name, len(me.ManaPool))
	}
	return id
}

func TestThunderousWrathMiracleDealsFive(t *testing.T) {
	g := newCatalogGame(t)
	victim := g.Seats[2]
	before := victim.Life
	castMiracleDrawn(t, g, "Thunderous Wrath", "Instant", "{4}{R}{R}", thunderousWrathOracle, "{R}",
		game.CastSpellParams{Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: victim.ID}}})
	passPriorityAroundTable(t, g)
	if victim.Life != before-5 {
		t.Errorf("life %d → %d, want -5", before, victim.Life)
	}
}

func TestDevastationTideMiracleBouncesNonlandPermanents(t *testing.T) {
	g := newCatalogGame(t)
	them := g.Seats[2]
	bear := pushPermanentForTest(g, them.ID, "Bear", "", "Creature — Bear")
	land := pushPermanentForTest(g, them.ID, "Forest", "", "Basic Land — Forest")
	castMiracleDrawn(t, g, "Devastation Tide", "Sorcery", "{3}{U}{U}", devastationTideOracle, "{C}{U}", game.CastSpellParams{})
	passPriorityAroundTable(t, g)
	if !them.Hand.Contains(bear) {
		t.Error("the creature was not returned to its owner's hand")
	}
	if !g.Battlefield.Contains(land) {
		t.Error("the land was returned; Devastation Tide spares lands")
	}
}

func TestReforgeTheSoulMiracleWheelsEveryone(t *testing.T) {
	g := newCatalogGame(t)
	castMiracleDrawn(t, g, "Reforge the Soul", "Sorcery", "{3}{R}{R}", reforgeTheSoulOracle, "{C}{R}", game.CastSpellParams{})
	passPriorityAroundTable(t, g)
	for _, p := range g.Seats {
		if p.Hand.Size() != 7 {
			t.Errorf("%s: hand %d, want 7", p.Name, p.Hand.Size())
		}
	}
}

func TestEntreatTheAngelsMiraclePaysXOnce(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[1]
	castMiracleDrawn(t, g, "Entreat the Angels", "Sorcery", "{X}{X}{W}{W}{W}", entreatTheAngelsOracle, "{C}{C}{C}{W}{W}",
		game.CastSpellParams{XValue: 3})
	passPriorityAroundTable(t, g)
	angels := 0
	for _, c := range g.Battlefield.Cards {
		if c.Controller == me.ID && c.IsCreature() && game.HasKeyword(&c, "flying") {
			angels++
		}
	}
	if angels != 3 {
		t.Errorf("%d flying Angels, want 3", angels)
	}
}

func TestBanishingStrokeMiracleTucksToTheBottom(t *testing.T) {
	g := newCatalogGame(t)
	them := g.Seats[2]
	bear := pushPermanentForTest(g, them.ID, "Bear", "", "Creature — Bear")
	castMiracleDrawn(t, g, "Banishing Stroke", "Instant", "{5}{W}", banishingStrokeOracle, "{W}",
		game.CastSpellParams{Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}}})
	passPriorityAroundTable(t, g)
	if n := them.Library.Size(); n == 0 || them.Library.Cards[0].InstanceID != bear {
		t.Error("the creature is not on the bottom of its owner's library")
	}
}

func TestBlessingsOfNatureMiracleForG(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[1]
	a := b12Creature(g, me.ID, "A", "Creature — Bear", 2, 2)
	castMiracleDrawn(t, g, "Blessings of Nature", "Sorcery", "{4}{G}", blessingsOfNatureOracle, "{G}",
		game.CastSpellParams{Targets: cardRefs(a), Distribution: map[uuid.UUID]int{a: 4}})
	passPriorityAroundTable(t, g)
	if n := e2Card(t, g, a).Counters[game.CounterPlusOne]; n != 4 {
		t.Errorf("%d +1/+1 counters, want 4", n)
	}
}
