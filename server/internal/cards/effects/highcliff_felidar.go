package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Highcliff Felidar — Creature — Cat Beast {5}{W}{W}, 5/5 (Game Night 2019,
// #5515):
//
//	"Vigilance
//	 When this creature enters, for each opponent, choose a creature with
//	 the greatest power among creatures that player controls. Destroy those
//	 creatures."
//
// The Felidar's controller makes every choice as the trigger resolves
// (CR 608.2c), one per opponent in turn order, through Winnowing's
// ChoosePermanents fan-out. Nothing is targeted, so hexproof, shroud and
// protection don't stop the choice. The offer for each opponent is the
// creatures tied for the greatest power among the creatures that player
// controls (GreatestPowerYouControl read for that player), so a pump or an
// anthem counts and a lone 0-power creature is still the greatest. An
// opponent with exactly one such creature is not asked about: that
// creature is collected unasked (Shadowgrange Archfiend's forced pick), and
// an opponent with no creature is skipped.
//
// Once every choice is made, the chosen creatures are destroyed at the
// same time (CR 701.8a), so an indestructible one survives and the others
// still go.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "43296f8b-58d9-446e-a538-1c4921552c41",
		Name:            "Highcliff Felidar",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Highcliff Felidar — for each opponent, destroy a creature with the greatest power among creatures that player controls",
				highcliffFelidarETB),
		},
	})
}

// highcliffFelidarETB asks for each opponent whose greatest power is tied,
// and carries the opponents' lone greatest-power creatures unasked.
func highcliffFelidarETB(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	opponents := opponentsInTurnOrder(g, ctx.Controller())
	var forced []uuid.UUID
	for _, opp := range opponents {
		if ids := highcliffFelidarGreatest(g, opp); len(ids) == 1 {
			forced = append(forced, ids[0])
		}
	}
	return ChoosePermanents{
		Question:   "Highcliff Felidar — choose a creature with the greatest power among creatures that player controls; it is destroyed",
		Of:         opponents,
		Candidates: highcliffFelidarTied,
		Keyed:      highcliffFelidarDestroy,
		Carry:      forced,
	}.Apply(ctx)
}

// highcliffFelidarGreatest is the creatures `of` controls with the greatest
// power among creatures that player controls.
func highcliffFelidarGreatest(g *game.Game, of uuid.UUID) []uuid.UUID {
	return permanentsControlledByMatching(g, of, GreatestPowerYouControl())
}

// highcliffFelidarTied is one opponent's offer: exactly one of their tied
// greatest-power creatures, or nothing (no leg) when there is no tie to
// break.
func highcliffFelidarTied(g *game.Game, of uuid.UUID) ([]uuid.UUID, int, int) {
	ids := highcliffFelidarGreatest(g, of)
	if len(ids) < 2 {
		return nil, 0, 0
	}
	return ids, 1, 1
}

// highcliffFelidarDestroy is "Destroy those creatures": the chosen and the
// unasked ones together, as one simultaneous destruction, keyed (#4290).
var highcliffFelidarDestroy = RegisterPermanentsPick("prompt-run/highcliff-felidar-destroy", PermanentsDef{
	ThenRun: func(ctx *Context, picked game.PromptedPicks, r game.RunDone) error {
		ids := append([]uuid.UUID(nil), r.Carry...)
		ids = append(ids, picked.Cards()...)
		var live []uuid.UUID
		for _, id := range ids {
			if onBattlefield(ctx.Game, id) {
				live = append(live, id)
			}
		}
		if len(live) > 0 {
			ctx.Game.DestroyPermanentsForEffect(live)
		}
		return nil
	},
})
