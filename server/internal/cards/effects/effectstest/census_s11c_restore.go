package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// census_s11c_restore_test.go — #4361 S11c: the last engine route batches
// whose continuation was a closure (zoneRoute.then) now carry it as data,
// and so does the hand-size discard. Each test pauses the move on a CR 616
// ordering prompt over a card named Needle (two probes that replace the
// move with itself), restores the table STRICTLY (the boot path) while the
// prompt is open, answers on the restored table, and checks what the
// continuation did:
//
//   - a surveil's graveyard leg (engine/surveil-routed): the EventSurveil
//     and the card's own "then" after it;
//   - a library search to a graveyard (routeBatchRun.Search): the search
//     finishes (EventSearchLibrary) and runs its ThenTail;
//   - Hinder's leg (engine/library-order-leg-landed on a keyed counter
//     route, engine/countered-to-library): the countered spell lands where
//     its controller put it;
//   - Guile's exile instead of Hinder's counter
//     (engine/counter-exiled-instead): the free cast is offered after the
//     restored answer;
//   - the CR 514.1 hand-size discard (engine/cleanup-discarded): the
//     cleanup step's bookkeeping runs and the turn moves on.

// libraryOrderProbeOracles are two test cards, each "if a card named Needle
// would be put into a library, it is put into that library instead": two
// declared effects, so a Needle's move to a library asks for their order
// (CR 616.1). Catalog rows, so a restored table gathers them again.
var LibraryOrderProbeOracles = []string{"test-4361-library-order-a", "test-4361-library-order-b"}

func init() {
	for i, oracle := range LibraryOrderProbeOracles {
		effects.Register(effects.Spec{
			OracleID: oracle,
			Name:     "Library Order Probe " + string(rune('A'+i)),
			Replacements: []game.ReplacementEffect{{
				Watches: []game.EventKind{game.EventZoneMove},
				AppliesTo: func(ev *game.ReplacementEvent, g *game.Game, _ *game.Card) bool {
					if ev.Kind != game.RepEventMove || ev.NewZone != game.ZoneLibrary {
						return false
					}
					c, ok := g.LookupCardForEffect(ev.CardID)
					return ok && c.Name == "Needle"
				},
				Replace: func(*game.ReplacementEvent, *game.Game, *game.Card) error { return nil },
				Label:   "Library Order Probe " + string(rune('A'+i)),
			}},
		})
	}
}

// --- Hinder ------------------------------------------------------------------

// castNeedleAt casts a Shock named Needle from caster's hand at a player.
func CastNeedleAt(t *testing.T, g *game.Game, caster *game.Player, at uuid.UUID) uuid.UUID {
	t.Helper()
	for g.Turn.Step != game.StepPrecombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	id := uuid.New()
	caster.Hand.PushTop(game.Card{
		InstanceID: id, Name: "Needle", TypeLine: "Instant",
		OracleID: B5ShockOracle, Owner: caster.ID, Controller: caster.ID,
	})
	if err := g.CastSpell(caster.ID, id, game.CastSpellParams{
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: at}},
	}); err != nil {
		t.Fatalf("CastSpell Needle: %v", err)
	}
	return id
}
