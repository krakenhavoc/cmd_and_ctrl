package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const BreechesBlastmakerOracle = "8514f20a-50c7-4319-84cb-2bf263548234"

// breechesRNGKey pins the keyed coin-flip stream (ADR 0054) so every
// run of breechesPlay draws the same sequence of "heads"/"tails"
// results from Breeches' own (player, source) stream. Without it,
// newCatalogGame leaves Game.rngKey unset, which mints a fresh
// crypto-random key on the first draw — so TestBreechesCopiesOnAWinAndBlastsOnALoss's
// scan over 12 unpinned flips had roughly a 1-in-4096 chance of
// landing all wins or all losses and failing (#1605). Any fixed,
// non-zero key works; this one has no meaning beyond "not all
// zeroes" and is not shared with internal/game's own known-answer
// tests.
func BreechesRNGKey() [32]byte {
	return [32]byte{
		0xb2, 0xee, 0xc4, 0x1e, 0x05, 0xd6, 0x60, 0x5c,
		0x12, 0x87, 0xff, 0x3a, 0x9c, 0x44, 0x70, 0xd1,
		0x91, 0x2d, 0xa8, 0x63, 0x0f, 0xb7, 0x4a, 0x38,
		0x55, 0xc9, 0x6e, 0x21, 0xf4, 0x0a, 0xd3, 0x77,
	}
}

// breechesSourceID is Breeches' fixed InstanceID. The keyed coin stream
// is seeded from (game key, player seat, SOURCE id, turn, counter), so
// pinning only the key left the source as uuid.New() and the whole
// flip sequence different on every run: the 12-burn scan in
// TestBreechesCopiesOnAWinAndBlastsOnALoss then saw all-wins or
// all-losses about 1 run in 2048 (#2545; #1605 pinned the key and
// missed this half).
var BreechesSourceID = uuid.MustParse("5b1e7c0d-3a42-4f6e-9d18-c2a07e4b6f93")

// castSpellWithCost is castCatalogSpell for a spell whose MANA COST
// the test needs — "damage equal to that spell's mana value" is
// unobservable on the zero-cost placeholder cards the harness seeds.
func CastSpellWithCost(t *testing.T, g *game.Game, name, typeLine, oracleID, manaCost string, targets []game.TargetRef) uuid.UUID {
	t.Helper()
	active := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	active.Hand.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: typeLine, OracleID: oracleID,
		ManaCost: manaCost, Owner: active.ID, Controller: active.ID,
	})
	for g.Turn.Step != game.StepPrecombatMain && g.Turn.Step != game.StepPostcombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	if err := g.CastSpell(active.ID, id, game.CastSpellParams{Targets: targets}); err != nil {
		t.Fatalf("CastSpell %s: %v", name, err)
	}
	return id
}

// passUntilPickTarget passes priority until `chooser` is asked to
// choose a target, or the stack empties.
func PassUntilPickTarget(t *testing.T, g *game.Game, chooser uuid.UUID) *game.PendingChoice {
	t.Helper()
	for i := 0; i < 16; i++ {
		if p := LatestPickTarget(g, chooser); p != nil {
			return p
		}
		if StackFullyEmpty(g) {
			return nil
		}
		if err := g.PassPriority(); err != nil {
			return LatestPickTarget(g, chooser)
		}
	}
	return LatestPickTarget(g, chooser)
}
