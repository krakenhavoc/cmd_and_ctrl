package cardso

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Oracle en-Vec — Creature — Human Wizard {1}{W}, 1/1 (Tempest, S82 OLD, #5628):
//
//	"{T}: Target opponent chooses any number of creatures they control. During
//	 that player's next turn, the chosen creatures attack if able, and other
//	 creatures can't attack. At the beginning of that turn's end step, destroy
//	 each of the chosen creatures that didn't attack this turn. Activate only
//	 during your turn."
//
// The opponent picks as the ability resolves (any number, none included), each
// creature as the object it is now (CR 400.7). The rest waits as a delayed
// trigger for the beginning of THAT player's next upkeep, which is when the turn
// is theirs: it makes each chosen creature that is still the same object attack
// if able (CR 508.1d), makes every other creature unable to attack through a live
// rule over all creatures (CR 611.2c; with none of the chosen left, no creature
// can attack), and puts one delayed destroy per chosen creature on that turn's
// end step, which skips a creature that attacked. A Wall chosen is destroyed all
// the same, as printed.
//
// No simplification.
func init() {
	const name = "Oracle en-Vec"
	effects.Register(effects.Spec{
		OracleID:     "367faa8c-617a-41a0-901c-f65106e86d56",
		Name:         name,
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:     "{T}: Target opponent chooses any number of creatures they control. During that player's next turn, the chosen creatures attack if able, and other creatures can't attack. At the beginning of that turn's end step, destroy each of the chosen creatures that didn't attack this turn. Activate only during your turn.",
			Purpose:   game.Purpose{Answers: game.AnswerRemove},
			Cost:      effects.TapCost(),
			Condition: effects.DuringYourTurn(),
			Targets:   effects.TargetPlayer("target opponent", effects.Opponent()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				t, ok := ctx.ClauseTarget(0)
				if !ok || t.Kind != game.TargetPlayer {
					return nil
				}
				var cands []uuid.UUID
				for _, c := range g.BattlefieldCardsForEffect() {
					if c.IsCreature() && c.Controller == t.ID {
						cands = append(cands, c.InstanceID)
					}
				}
				if len(cands) == 0 {
					return effects.S82OldB761ScheduleOracleTurn(ctx, t.ID, nil)
				}
				g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
					Chooser:    t.ID,
					FromPlayer: t.ID,
					Source:     ctx.Source(),
					Question:   name + " — choose any number of creatures you control: they attack during your next turn if able, no other creature can, and each that doesn't attack is destroyed",
					Cards:      cands,
					Min:        0,
					Max:        len(cands),
					LiveZone:   game.ZoneBattlefield,
					ThenKey:    oracleEnVecChosen,
					Carry:      []uuid.UUID{t.ID},
					Asked:      game.NewAskedItem(ctx.Item),
				})
				return nil
			},
		}},
	})
}

// oracleEnVecChosen is the opponent's pick, answered. An on-disk key: never
// renamed, never reused.
var oracleEnVecChosen = effects.AskedPick("choose-cards/s82-old-b761-oracle-en-vec", func(ctx *effects.Context, r game.CardsChosen) error {
	if len(r.Carry) == 0 {
		return nil
	}
	return effects.S82OldB761ScheduleOracleTurn(ctx, r.Carry[0], r.Picked)
})
