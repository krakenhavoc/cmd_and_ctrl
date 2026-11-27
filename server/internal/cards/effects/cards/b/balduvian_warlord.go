package cardsb

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Balduvian Warlord — Creature — Human Barbarian {3}{R}, 3/2 (Coldsnap):
//
//	"{T}: Remove target blocking creature from combat. Creatures it was
//	 blocking that hadn't become blocked by another creature this combat
//	 become unblocked, then it blocks an attacking creature of your choice.
//	 Activate only during the declare blockers step."
//
// The first two sentences are RemoveBlockerUnblockingForEffect (CR 506.4,
// CR 509.1h). "Then it blocks" is BlockAttackersForEffect for the one
// attacker the Warlord's controller picks, from the attackers that creature
// could legally block (CR 509.1a-b restrictions still apply, but its own
// summoning sickness and tapped state don't, because it is put onto the
// block, not declared). A blocker that can block nothing stays removed. The
// target may be the Warlord's controller's own blocker or an opponent's.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f73d1525-ab0b-4772-9dfa-66e1c84c3ea4",
		Name:         "Balduvian Warlord",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:     "{T}: Remove target blocking creature from combat. Creatures it was blocking that hadn't become blocked by another creature this combat become unblocked, then it blocks an attacking creature of your choice. Activate only during the declare blockers step.",
			Purpose:   game.Purpose{Answers: game.AnswerRestrict},
			Cost:      effects.TapCost(),
			Condition: effects.DuringStep(game.StepDeclareBlockers),
			Targets:   effects.TargetCreature("target blocking creature", effects.BlockingCreature()),
			Effect:    balduvianWarlordResolve,
		}},
	})
}

func balduvianWarlordResolve(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	blocker, ok := effects.B16FirstLegalTargetCard(ctx)
	if !ok {
		return nil
	}
	g.RemoveBlockerUnblockingForEffect(blocker)
	return effects.ChoosePermanents{
		Question: "Balduvian Warlord — choose an attacking creature for it to block",
		Candidates: func(g *game.Game, _ uuid.UUID) ([]uuid.UUID, int, int) {
			var ids []uuid.UUID
			for _, a := range effects.AttackingCreatures(g) {
				if g.CanBlockByEffectForEffect(blocker, a) {
					ids = append(ids, a)
				}
			}
			return ids, 1, 1
		},
		Keyed: balduvianWarlordChosen,
		Carry: []uuid.UUID{blocker},
	}.Apply(ctx)
}

var balduvianWarlordChosen = effects.RegisterPermanentsPick("prompt-run/balduvian-warlord-block", effects.PermanentsDef{
	ThenRun: func(ctx *effects.Context, picked game.PromptedPicks, r game.RunDone) error {
		chosen := picked.Cards()
		if len(chosen) == 0 || len(r.Carry) == 0 {
			return nil
		}
		ctx.Game.BlockAttackersForEffect(r.Carry[0], chosen[:1])
		return nil
	},
})
