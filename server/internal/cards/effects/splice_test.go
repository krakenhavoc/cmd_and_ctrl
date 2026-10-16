package effects

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// splice_test.go — CR 702.47, splice (#1989, ADR 0202). The host is
// Consuming Vortex ("Return target creature to its owner's hand"), an
// Arcane instant, so every test reads the main text's effect and the
// spliced text's side by side. Every refusal is paired with an
// acceptance on the same board.

const (
	consumingVortexOracle = "863bb8d8-4571-4775-87e3-d6f3ec88c835"
	glacialRayOracle      = "e65a8ab8-84a7-4fe2-b009-ceb87bec6b4e"
	desperateRitualOracle = "3620ab8f-b693-4225-bb96-5ca3cb3fa465"
	vitalSurgeOracle      = "e19c0205-67e3-4e44-995b-751e18fdaafd"
	psychicPuppetryOracle = "8d3c4e50-6709-42f9-97e7-3165ee501684"
	everdreamOracle       = "5b17d3c5-6085-416e-9afd-8ed0639ae7d2"
	overblazeOracle       = "543599e2-8312-43a6-ab26-0be17e690d4f"
	spliceShockOracle     = "a9d288b8-cdc1-4e55-a0c9-d6edfc95e65d"
)

type spliceBoard struct {
	g          *game.Game
	me, opp    *game.Player
	bear, wolf uuid.UUID
}

func newSpliceBoard(t *testing.T) spliceBoard {
	t.Helper()
	g := newCatalogGame(t)
	me, opp := cmcSeats(t, g)
	return spliceBoard{
		g: g, me: me, opp: opp,
		bear: twPut(g, opp.ID, "Bear", "Creature — Bear", []string{"G"}),
		wolf: twPut(g, opp.ID, "Wolf", "Creature — Wolf", []string{"G"}),
	}
}

func (b spliceBoard) hand(name, typeLine, cost, oracle, color string) uuid.UUID {
	return handCardFull(b.me, name, typeLine, cost, oracle, []string{color})
}

func (b spliceBoard) vortex() uuid.UUID {
	return b.hand("Consuming Vortex", "Instant — Arcane", "{1}{U}", consumingVortexOracle, "U")
}

func (b spliceBoard) ray() uuid.UUID {
	return b.hand("Glacial Ray", "Instant — Arcane", "{1}{R}", glacialRayOracle, "R")
}

// cast casts `spell` with exactly `pool`. A refused cast empties the
// pool, so the next attempt starts from the same board.
func (b spliceBoard) cast(spell uuid.UUID, pool string, splices []uuid.UUID, targets ...game.TargetRef) error {
	b.me.ManaPool.EmptyPool()
	if pool != "" {
		if err := b.g.AddManaForEffect(b.me.ID, uuid.Nil, pool); err != nil {
			return err
		}
	}
	err := b.g.CastSpell(b.me.ID, spell, game.CastSpellParams{Strict: true, Splices: splices, Targets: targets})
	if err != nil {
		b.me.ManaPool.EmptyPool()
	}
	return err
}

func spCardRef(id uuid.UUID) game.TargetRef { return game.TargetRef{Kind: game.TargetCard, ID: id} }

func splicedRef(k int, ref game.TargetRef) game.TargetRef {
	ref.Splice = k
	return ref
}

func spInHand(p *game.Player, id uuid.UUID) bool { return p.Hand.Contains(id) }

// The headline: Glacial Ray spliced onto Consuming Vortex. The bear goes
// home, the opponent takes 2, Glacial Ray is still in hand, and the
// spell cost both {1}{U} and the splice cost {1}{R} (CR 702.47a).
func TestGlacialRaySplicedOntoConsumingVortex(t *testing.T) {
	b := newSpliceBoard(t)
	vortex, ray := b.vortex(), b.ray()
	life := b.opp.Life
	if err := b.cast(vortex, "{C}{U}{C}{R}", []uuid.UUID{ray},
		spCardRef(b.bear), splicedRef(1, game.TargetRef{Kind: game.TargetPlayer, ID: b.opp.ID})); err != nil {
		t.Fatalf("splice: %v", err)
	}
	if !spInHand(b.me, ray) {
		t.Fatal("Glacial Ray left the hand: a spliced card stays there (CR 702.47a)")
	}
	if left := len(b.me.ManaPool.Units()); left != 0 {
		t.Errorf("%d mana left, want 0 — {1}{U} plus the splice cost {1}{R}", left)
	}
	passPriorityAroundTable(t, b.g)
	if !spInHand(b.opp, b.bear) {
		t.Error("the main text did not return the bear")
	}
	if got := life - b.opp.Life; got != 2 {
		t.Errorf("the opponent lost %d, want 2 from the spliced text", got)
	}
	if !spInHand(b.me, ray) {
		t.Error("Glacial Ray left the hand as the spell resolved")
	}
}

// The splice cost is an additional cost (CR 702.47a, 601.2f): the
// host's own mana does not pay for it.
func TestSpliceCostIsCharged(t *testing.T) {
	b := newSpliceBoard(t)
	vortex, ray := b.vortex(), b.ray()
	targets := []game.TargetRef{spCardRef(b.bear), splicedRef(1, spCardRef(b.wolf))}
	if err := b.cast(vortex, "{C}{U}", []uuid.UUID{ray}, targets...); err == nil {
		t.Fatal("spliced for {1}{U}: the {1}{R} splice cost is unpaid")
	}
	if err := b.cast(vortex, "{C}{U}{C}{R}", []uuid.UUID{ray}, targets...); err != nil {
		t.Fatalf("spliced for {1}{U}{1}{R}: %v", err)
	}
}

// What can't be spliced: a card onto a spell without the quality, a card
// twice, a card onto itself, a card not in the hand, and a text whose
// required target is missing (CR 702.47a-b).
func TestSpliceRefusals(t *testing.T) {
	b := newSpliceBoard(t)
	ray := b.ray()
	shock := b.hand("Shock", "Instant", "{R}", spliceShockOracle, "R")
	opp := game.TargetRef{Kind: game.TargetPlayer, ID: b.opp.ID}
	if err := b.cast(shock, "{R}{C}{R}", []uuid.UUID{ray}, opp, splicedRef(1, opp)); err == nil {
		t.Error("Glacial Ray spliced onto Shock, which is not Arcane")
	}
	vortex := b.vortex()
	for _, tc := range []struct {
		name    string
		splices []uuid.UUID
		targets []game.TargetRef
	}{
		{"the same card twice", []uuid.UUID{ray, ray}, []game.TargetRef{spCardRef(b.bear), splicedRef(1, opp), splicedRef(2, opp)}},
		{"onto itself", []uuid.UUID{vortex}, []game.TargetRef{spCardRef(b.bear), splicedRef(1, spCardRef(b.wolf))}},
		{"a card not in the hand", []uuid.UUID{uuid.New()}, []game.TargetRef{spCardRef(b.bear), splicedRef(1, opp)}},
		{"no target for the spliced text", []uuid.UUID{ray}, []game.TargetRef{spCardRef(b.bear)}},
		{"a target for a text nothing was spliced for", []uuid.UUID{ray}, []game.TargetRef{spCardRef(b.bear), splicedRef(1, opp), splicedRef(2, opp)}},
	} {
		if err := b.cast(vortex, "{C}{U}{C}{R}{C}{R}", tc.splices, tc.targets...); err == nil {
			t.Errorf("%s: accepted", tc.name)
		} else if !errors.Is(err, game.ErrInvalidParam) && !errors.Is(err, game.ErrIllegalTarget) {
			t.Errorf("%s: %v, want a bad announcement", tc.name, err)
		}
	}
	if err := b.cast(vortex, "{C}{U}{C}{R}", []uuid.UUID{ray}, spCardRef(b.bear), splicedRef(1, opp)); err != nil {
		t.Fatalf("the same board, spliced properly: %v", err)
	}
}

// "Splice onto instant or sorcery" (Everdream) splices onto any instant.
func TestEverdreamSplicesOntoAnInstant(t *testing.T) {
	b := newSpliceBoard(t)
	dream := b.hand("Everdream", "Instant", "{1}{U}", everdreamOracle, "U")
	shock := b.hand("Shock", "Instant", "{R}", spliceShockOracle, "R")
	hand := b.me.Hand.Size()
	if err := b.cast(shock, "{R}{C}{C}{U}", []uuid.UUID{dream}, game.TargetRef{Kind: game.TargetPlayer, ID: b.opp.ID}); err != nil {
		t.Fatalf("Everdream onto Shock: %v", err)
	}
	passPriorityAroundTable(t, b.g)
	// Shock left the hand (−1) and the spliced text drew one (+1).
	if got := b.me.Hand.Size(); got != hand {
		t.Errorf("hand %d, want %d: Shock cast, one card drawn, Everdream kept", got, hand)
	}
}

// Desperate Ritual spliced onto Desperate Ritual: four mana in, six red
// out — the main text's three, then the spliced text's three.
func TestDesperateRitualSplicedOntoItsTwin(t *testing.T) {
	b := newSpliceBoard(t)
	first := b.hand("Desperate Ritual", "Instant — Arcane", "{1}{R}", desperateRitualOracle, "R")
	second := b.hand("Desperate Ritual", "Instant — Arcane", "{1}{R}", desperateRitualOracle, "R")
	if err := b.cast(first, "{C}{R}{C}{R}", []uuid.UUID{second}); err != nil {
		t.Fatalf("splice: %v", err)
	}
	passPriorityAroundTable(t, b.g)
	if got := redIn(b.me); got != 6 {
		t.Errorf("%d red in the pool, want 6", got)
	}
	if !spInHand(b.me, second) {
		t.Error("the spliced Desperate Ritual left the hand")
	}
}

// The main text happens first (CR 702.47b), and a spliced text waits
// for a prompt the main text opened: Psychic Puppetry asks whether to
// tap or untap, and Vital Surge's life comes only after the answer.
func TestSplicedTextWaitsForTheMainTextsPrompt(t *testing.T) {
	b := newSpliceBoard(t)
	puppetry := b.hand("Psychic Puppetry", "Instant — Arcane", "{1}{U}", psychicPuppetryOracle, "U")
	surge := b.hand("Vital Surge", "Instant — Arcane", "{1}{G}", vitalSurgeOracle, "G")
	life := b.me.Life
	if err := b.cast(puppetry, "{C}{U}{C}{G}", []uuid.UUID{surge}, spCardRef(b.bear)); err != nil {
		t.Fatalf("splice: %v", err)
	}
	passPriorityAroundTable(t, b.g)
	if latestOptionPickFor(b.g, b.me.ID) == nil {
		t.Fatal("Psychic Puppetry did not ask")
	}
	if b.me.Life != life {
		t.Fatalf("life %d before the main text's answer, want %d: the spliced text ran first", b.me.Life, life)
	}
	answerOptionPick(t, b.g, b.me.ID, 0) // tap it
	b.g.SettleResolution()
	if got := b.me.Life - life; got != 3 {
		t.Errorf("gained %d after the answer, want 3", got)
	}
}

// CR 608.2b with spliced text: the spell resolves while any text has a
// legal target, and a text whose every target is gone does nothing.
func TestSplicedSpellResolvesWhileAnyTextHasALegalTarget(t *testing.T) {
	b := newSpliceBoard(t)
	vortex, ray := b.vortex(), b.ray()
	life := b.opp.Life
	if err := b.cast(vortex, "{C}{U}{C}{R}", []uuid.UUID{ray},
		spCardRef(b.bear), splicedRef(1, game.TargetRef{Kind: game.TargetPlayer, ID: b.opp.ID})); err != nil {
		t.Fatalf("splice: %v", err)
	}
	b.g.WithWriteLock(func() { _ = b.g.DestroyPermanentForEffect(b.bear) })
	passPriorityAroundTable(t, b.g)
	if got := life - b.opp.Life; got != 2 {
		t.Errorf("the opponent lost %d, want 2: the spliced text's target is still legal", got)
	}
}

// Every target of every text illegal: countered by the rules.
func TestSplicedSpellWithEveryTargetGoneIsCountered(t *testing.T) {
	b := newSpliceBoard(t)
	vortex, ray := b.vortex(), b.ray()
	if err := b.cast(vortex, "{C}{U}{C}{R}", []uuid.UUID{ray}, spCardRef(b.bear), splicedRef(1, spCardRef(b.wolf))); err != nil {
		t.Fatalf("splice: %v", err)
	}
	b.g.WithWriteLock(func() {
		_ = b.g.DestroyPermanentForEffect(b.bear)
		_ = b.g.DestroyPermanentForEffect(b.wolf)
	})
	passPriorityAroundTable(t, b.g)
	fizzled := false
	b.g.ReadSnapshot(func() {
		for _, ev := range b.g.Events {
			if ev.Kind == game.EventFizzle && ev.CardID == vortex {
				fizzled = true
			}
		}
	})
	if !fizzled {
		t.Error("a spliced spell whose every target is gone was not countered by the rules")
	}
}

// CR 707.10: a copy copies the spliced text and its targets.
func TestACopyOfASplicedSpellCopiesTheSplicedText(t *testing.T) {
	b := newSpliceBoard(t)
	vortex, ray := b.vortex(), b.ray()
	life := b.opp.Life
	if err := b.cast(vortex, "{C}{U}{C}{R}", []uuid.UUID{ray},
		spCardRef(b.bear), splicedRef(1, game.TargetRef{Kind: game.TargetPlayer, ID: b.opp.ID})); err != nil {
		t.Fatalf("splice: %v", err)
	}
	var err error
	b.g.WithWriteLock(func() { err = b.g.CopySpellForEffect(vortex, b.me.ID, false, nil) })
	if err != nil {
		t.Fatalf("copy: %v", err)
	}
	passPriorityAroundTable(t, b.g)
	if got := life - b.opp.Life; got != 4 {
		t.Errorf("the opponent lost %d, want 4: the copy's Glacial Ray too", got)
	}
}

// The record survives undo and a snapshot round trip, and the restored
// game resolves the spliced text.
func TestSplicedTextSurvivesUndoAndASnapshotRoundTrip(t *testing.T) {
	b := newSpliceBoard(t)
	vortex, ray := b.vortex(), b.ray()
	if err := b.cast(vortex, "{C}{U}{C}{R}", []uuid.UUID{ray},
		spCardRef(b.bear), splicedRef(1, game.TargetRef{Kind: game.TargetPlayer, ID: b.opp.ID})); err != nil {
		t.Fatalf("splice: %v", err)
	}
	undo := b.g.Clone()
	b.g.RestoreFrom(undo)
	blob, err := json.Marshal(b.g.CaptureSnapshot())
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
	var opp *game.Player
	for _, p := range restored.Seats {
		if p.ID == b.opp.ID {
			opp = p
		}
	}
	life := opp.Life
	passPriorityAroundTable(t, restored)
	if got := life - opp.Life; got != 2 {
		t.Errorf("restored: the opponent lost %d, want 2", got)
	}
}

// Overblaze spliced: the multiplier is the spell's, pinned to the target.
func TestOverblazeSplicedDoublesItsTarget(t *testing.T) {
	b := newSpliceBoard(t)
	mine := twPut(b.g, b.me.ID, "Ogre", "Creature — Ogre", []string{"R"})
	vortex := b.vortex()
	blaze := b.hand("Overblaze", "Instant — Arcane", "{3}{R}", overblazeOracle, "R")
	if err := b.cast(vortex, "{C}{U}{C}{C}{R}{R}", []uuid.UUID{blaze}, spCardRef(b.bear), splicedRef(1, spCardRef(mine))); err != nil {
		t.Fatalf("splice: %v", err)
	}
	passPriorityAroundTable(t, b.g)
	v := protocol.ViewOfGameFor(b.g, b.me.ID.String())
	if len(v.DamageMultipliers) == 0 {
		t.Fatal("no damage multiplier after Overblaze's spliced text")
	}
}

// The hand view offers the splice on an Arcane spell and not on another.
func TestHandViewOffersSplice(t *testing.T) {
	b := newSpliceBoard(t)
	vortex, ray := b.vortex(), b.ray()
	shock := b.hand("Shock", "Instant", "{R}", spliceShockOracle, "R")
	for i := range b.me.Hand.Cards {
		b.me.Hand.Cards[i].KnownBy = map[uuid.UUID]bool{b.me.ID: true}
	}
	v := protocol.ViewOfGameFor(b.g, b.me.ID.String())
	offers := func(id uuid.UUID) []protocol.SpliceOfferView {
		for _, s := range v.Seats {
			for _, c := range s.Hand.Cards {
				if c.InstanceID == id.String() {
					return c.SpliceOffers
				}
			}
		}
		t.Fatalf("card %s not in the view", id)
		return nil
	}
	got := offers(vortex)
	if len(got) != 1 || got[0].CardID != ray.String() || got[0].ManaCost != "{1}{R}" || got[0].Label != "Splice onto Arcane {1}{R}" {
		t.Fatalf("Consuming Vortex's offers %+v, want Glacial Ray's", got)
	}
	if got[0].LegalTargets == nil || len(got[0].LegalTargets.Players) == 0 {
		t.Errorf("Glacial Ray's offer has no legal targets: %+v", got[0].LegalTargets)
	}
	if got := offers(shock); len(got) != 0 {
		t.Errorf("Shock is not Arcane; offers %+v", got)
	}
	// Glacial Ray is Arcane too: Consuming Vortex splices onto it, and it
	// never splices onto itself.
	if got := offers(ray); len(got) != 1 || got[0].CardID != vortex.String() {
		t.Errorf("Glacial Ray's offers %+v, want Consuming Vortex's alone", got)
	}
	// An opponent sees nothing.
	ov := protocol.ViewOfGameFor(b.g, b.opp.ID.String())
	for _, s := range ov.Seats {
		for _, c := range s.Hand.Cards {
			if len(c.SpliceOffers) > 0 {
				t.Errorf("an opponent's view carries splice offers on %s", c.InstanceID)
			}
		}
	}
}

// Register refuses a splice the engine cannot carry out as printed.
func TestRegisterRefusesAnUnbuildableSplice(t *testing.T) {
	body := func(*game.StackItem, *Context) error { return nil }
	for _, tc := range []struct {
		name string
		spec Spec
	}{
		{"no cost", Spec{Splice: &game.SpliceAbility{Onto: game.SpliceOntoArcane}, OnResolve: body}},
		{"unknown quality", Spec{Splice: &game.SpliceAbility{Onto: "creature", ManaCost: "{1}"}, OnResolve: body}},
		{"no body", Spec{Splice: SpliceOntoArcane("{1}")}},
		{"modes", Spec{Splice: SpliceOntoArcane("{1}"), OnResolve: body, Modes: ChooseOne(Mode("You gain 1 life."), Mode("You lose 1 life."))}},
	} {
		tc.spec.OracleID = "splice-test-" + tc.name
		tc.spec.Name = "Splice Test " + tc.name
		func() {
			defer func() {
				if recover() == nil {
					t.Errorf("%s: Register accepted it", tc.name)
				}
			}()
			Register(tc.spec)
		}()
	}
}

// Candles' Glow spliced onto Consuming Vortex: the shield is the
// spell's, and its life comes as the damage is prevented.
func TestCandlesGlowSplicedShieldsAndGainsLife(t *testing.T) {
	b := newSpliceBoard(t)
	vortex := b.vortex()
	glow := b.hand("Candles' Glow", "Instant — Arcane", "{1}{W}", "7509b934-6fbc-42fd-9040-cdf16bbfff42", "W")
	if err := b.cast(vortex, "{C}{U}{C}{W}", []uuid.UUID{glow},
		spCardRef(b.bear), splicedRef(1, game.TargetRef{Kind: game.TargetPlayer, ID: b.me.ID})); err != nil {
		t.Fatalf("splice: %v", err)
	}
	passPriorityAroundTable(t, b.g)
	life := b.me.Life
	b.g.WithWriteLock(func() { _ = b.g.DealDamageToPlayerForEffect(b.wolf, b.me.ID, 5) })
	b.g.SettleResolution()
	passPriorityAroundTable(t, b.g)
	// 3 of the 5 prevented, 2 through, and 3 life gained.
	if got := b.me.Life - life; got != 1 {
		t.Errorf("life changed by %d, want +1 (−2 through the shield, +3 for what it prevented)", got)
	}
}
