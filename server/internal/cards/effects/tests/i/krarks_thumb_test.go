package testsi

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// krarks_thumb_test.go — Krark's Thumb (#6152): "If you would flip a coin,
// instead flip two coins and ignore one." The flips are random, so the
// tests flip many coins and read the rate: a kept won coin out of two is
// won three times in four, out of four fifteen times in sixteen.

const krarksThumbOracle = "a97c8482-775c-4f11-9872-25b4f9bfcb1a"

func TestKrarksThumbIsRegisteredFull(t *testing.T) {
	spec, ok := effects.Lookup(krarksThumbOracle)
	if !ok || spec.Name != "Krark's Thumb" || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
		t.Fatalf("spec = %+v, %v", spec, ok)
	}
}

// thumbOn puts a Krark's Thumb onto the battlefield under `owner`.
func thumbOn(g *game.Game, owner uuid.UUID) {
	effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Krark's Thumb", TypeLine: "Legendary Artifact", OracleID: krarksThumbOracle,
		Owner: owner, Controller: owner,
	})
}

func thumbFlipEvents(g *game.Game, player uuid.UUID) (heads, won, total int) {
	g.ReadSnapshot(func() {
		for _, ev := range g.EventsThisTurn() {
			if ev.Kind != game.EventFlipCoin || ev.Actor != player {
				continue
			}
			total++
			if ev.Label == "heads" {
				heads++
			}
			if ev.Won {
				won++
			}
		}
	})
	return
}

// calledWins makes one called flip of n coins and returns how many were
// won and how many flip events it emitted.
func calledWins(t *testing.T, thumbs, n int) (won, total int) {
	t.Helper()
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	for i := 0; i < thumbs; i++ {
		thumbOn(g, me.ID)
	}
	effectstest.AdvanceToMain(t, g)
	var id uuid.UUID
	g.WithWriteLock(func() { id = g.FlipCoinForEffect(game.CoinFlipSpec{Flipper: me.ID, Coins: n, Question: "call"}) })
	if err := g.ResolveCoinCall(id, me.ID, "tails"); err != nil {
		t.Fatalf("ResolveCoinCall: %v", err)
	}
	_, won, total = thumbFlipEvents(g, me.ID)
	return won, total
}

// One Thumb: each coin is two and a won one is kept, so a flip that is lost
// on one coin and won on the other is a win (3/4 of coins won). The ignored
// coin emits nothing: one event per coin the effect asked for.
func TestKrarksThumbKeepsAWonCoin(t *testing.T) {
	const n = 400
	plain, total := calledWins(t, 0, n)
	if total != n {
		t.Fatalf("%d flip events without a Thumb, want %d", total, n)
	}
	one, total := calledWins(t, 1, n)
	if total != n {
		t.Fatalf("%d flip events under one Thumb, want %d (the ignored coin emits nothing)", total, n)
	}
	if one < 260 || one > 340 {
		t.Errorf("won %d of %d under one Thumb, want about 300 (3 in 4)", one, n)
	}
	if plain < 160 || plain > 240 {
		t.Errorf("won %d of %d with no Thumb, want about 200 (the plain 50/50)", plain, n)
	}
}

// Two Thumbs compound: four coins for each, one kept (15/16 won).
func TestTwoKrarksThumbsFlipFourCoins(t *testing.T) {
	const n = 400
	two, total := calledWins(t, 2, n)
	if total != n {
		t.Fatalf("%d flip events under two Thumbs, want %d", total, n)
	}
	if two < 350 {
		t.Errorf("won %d of %d under two Thumbs, want about 375 (15 in 16)", two, n)
	}
}

// An uncalled flip keeps the face the effect wants for the flipper: heads
// unless it says otherwise (3/4 heads under one Thumb).
func TestKrarksThumbKeepsTheWantedFaceOnAnUncalledFlip(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	thumbOn(g, me.ID)
	effectstest.AdvanceToMain(t, g)
	var faces []bool
	g.WithWriteLock(func() { faces, _ = g.FlipCoinsForEffect(game.RandomDraw{Player: me.ID}, 400) })
	heads := 0
	for _, h := range faces {
		if h {
			heads++
		}
	}
	if heads < 260 || heads > 340 {
		t.Errorf("%d of 400 heads under one Thumb, want about 300", heads)
	}
	prefer := make([]bool, 400) // tails wanted on every coin
	g.WithWriteLock(func() { faces, _ = g.FlipCoinsPreferringForEffect(game.RandomDraw{Player: me.ID}, prefer) })
	tails := 0
	for _, h := range faces {
		if !h {
			tails++
		}
	}
	if tails < 260 || tails > 340 {
		t.Errorf("%d of 400 tails when tails is wanted, want about 300", tails)
	}
	if _, _, total := thumbFlipEvents(g, me.ID); total != 800 {
		t.Errorf("%d flip events, want 800 (one per coin kept)", total)
	}
}

// An opponent's Thumb changes nothing about my flips.
func TestKrarksThumbOnlyAffectsItsController(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	thumbOn(g, opp.ID)
	effectstest.AdvanceToMain(t, g)
	var id uuid.UUID
	g.WithWriteLock(func() { id = g.FlipCoinForEffect(game.CoinFlipSpec{Flipper: me.ID, Coins: 400, Question: "call"}) })
	if err := g.ResolveCoinCall(id, me.ID, "heads"); err != nil {
		t.Fatalf("ResolveCoinCall: %v", err)
	}
	if _, won, _ := thumbFlipEvents(g, me.ID); won < 160 || won > 240 {
		t.Errorf("won %d of 400 with only an opponent's Thumb, want about 200", won)
	}
}

// Rakdos, the Showstopper under Krark's Thumb: the flipper keeps heads for
// their own creatures and tails for everyone else's, so far more of theirs
// survive and far more of the opponent's die than the plain 50/50.
func TestKrarksThumbUnderRakdosKeepsTheFlippersPreference(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	thumbOn(g, me.ID)
	var mine, theirs []uuid.UUID
	for i := 0; i < 40; i++ {
		mine = append(mine, effectstest.B12Permanent(g, me.ID, "Bear", "Creature — Bear"))
		theirs = append(theirs, effectstest.B12Permanent(g, opp.ID, "Elk", "Creature — Elk"))
	}
	effectstest.CastCatalogSpell(t, g, "Rakdos, the Showstopper", "Legendary Creature — Demon", "03a4b997-4738-41a9-933f-e1f765e3a75a", nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	survived, died := 0, 0
	for _, id := range mine {
		if effectstest.BattlefieldHas(g, id) {
			survived++
		}
	}
	for _, id := range theirs {
		if !effectstest.BattlefieldHas(g, id) {
			died++
		}
	}
	if survived < 23 || died < 23 {
		t.Errorf("%d of 40 of mine survived and %d of 40 of theirs died, want about 30 each (3 in 4)", survived, died)
	}
}
