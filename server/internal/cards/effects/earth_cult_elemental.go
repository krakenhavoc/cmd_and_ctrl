package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Earth-Cult Elemental — Creature — Elemental {4}{R}{R}, 6/6 (Adventures
// in the Forgotten Realms, #5499):
//
//	"Siege Monster — When this creature enters, roll a d20.
//	 1—9 | Each player sacrifices a permanent of their choice.
//	 10—19 | Each opponent sacrifices a permanent of their choice.
//	 20 | Each opponent sacrifices two permanents of their choice."
//
// Every player asked chooses in APNAP order and everything chosen is
// sacrificed together (CR 101.4, 701.21a); nothing is targeted.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "043d8dee-379f-4706-ad47-0bb562717007",
		Name:         "Earth-Cult Elemental",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Siege Monster — roll a d20", Do(RollDice{Sides: 20, N: 1, Then: earthCultElementalTable})),
		},
	})
}

// earthCultElementalTable is the results table, after the d20.
var earthCultElementalTable = DiceTail("dice/earth-cult-elemental-table", func(ctx *Context, rolls []int) error {
	g := ctx.Game
	if rolls[0] < 10 {
		g.EachPlayerSacrificesForEffect(ctx.Source(), uuid.Nil, nil, "Sacrifice a permanent")
		return nil
	}
	n := 1
	if rolls[0] == 20 {
		n = 2
	}
	var asks []uuid.UUID
	for _, seat := range apnapLivingSeats(g) {
		if seat == ctx.Controller() {
			continue
		}
		for i := 0; i < n; i++ {
			asks = append(asks, seat)
		}
	}
	g.PlayersSacrificeForEffect(ctx.Source(), asks, nil, "Sacrifice "+map[int]string{1: "a permanent", 2: "two permanents"}[n])
	return nil
})
