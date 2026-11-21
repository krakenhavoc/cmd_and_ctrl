package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// The Fifth Doctor — Legendary Creature — Time Lord Doctor {2}{W}{U}, 2/2
// (Doctor Who, #5948):
//
//	"Peaceful Coexistence — At the beginning of your end step, put a +1/+1
//	 counter on each creature you control that didn't attack or enter this
//	 turn. Untap those creatures."
//
// The set of creatures is read once as the ability resolves (CR 611.2c): each
// creature the controller has that did not attack this turn (the turn tally's
// attack record) and did not enter the battlefield this turn (its entry
// record, so a creature that was flickered or reanimated this turn is a new
// object that did enter). Each gets a counter, then those same creatures untap.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "992ddd9f-e0db-4383-909c-6dacca41d188",
		Name:         "The Fifth Doctor",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			AtYourEndStep("The Fifth Doctor — put a +1/+1 counter on each creature you control that didn't attack or enter this turn, then untap those creatures",
				fifthDoctorEndStep),
		},
	})
}

func fifthDoctorEndStep(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	var ids []uuid.UUID
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == item.Controller && c.IsCreature() &&
			!g.AttackedThisTurn(c.InstanceID) && !g.EnteredThisTurn(c.InstanceID) {
			ids = append(ids, c.InstanceID)
		}
	}
	for _, id := range ids {
		if err := (AddCounter{Target: id, Kind: game.CounterPlusOne, N: 1}).Apply(ctx.asGroupMember()); err != nil {
			return err
		}
	}
	for _, id := range ids {
		if err := (UntapTarget{Target: id}).Apply(ctx.asGroupMember()); err != nil {
			return err
		}
	}
	return nil
}
