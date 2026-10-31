package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Citizen V, Helmut Zemo — Legendary Creature — Human Noble Villain
// {2}{B}, 2/2 (Marvel Super Heroes Commander, #4455):
//
//	"Lifelink
//	 Whenever you gain life, put a +1/+1 counter on each Villain you
//	 control."
//
// Lifelink is the printed keyword, and the lifegain trigger is Ajani's
// Pridemate's WheneverYouGainLife (one trigger per gain event, lifelink
// damage included). The counters go on every Villain the controller
// controls, Citizen V himself among them (he is a Villain): the set is
// snapshotted before any counter lands and each recipient is its own
// placement.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "974fbeb0-77d6-4a29-aa91-226cb5b32233",
		Name:            "Citizen V, Helmut Zemo",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"lifelink"},
		Triggered: []game.TriggeredAbility{
			WheneverYouGainLife("Citizen V, Helmut Zemo — put a +1/+1 counter on each Villain you control",
				Do(Instruction(msc5CounterOnEachVillain))),
		},
	})
}

// msc5CounterOnEachVillain puts a +1/+1 counter on each Villain the
// ability's controller controls, the source included when it is one.
func msc5CounterOnEachVillain(ctx *Context) error {
	var ids []uuid.UUID
	for _, c := range MatchingBattlefield(ctx, And(Subtype(ctx.TextSubtype("Villain")), YouControl())) {
		ids = append(ids, c.InstanceID)
	}
	for _, id := range ids {
		if !onBattlefield(ctx.Game, id) {
			continue
		}
		if err := (AddCounter{Target: id, Kind: game.CounterPlusOne, N: 1}).Apply(ctx.asGroupMember()); err != nil {
			return err
		}
	}
	return nil
}
