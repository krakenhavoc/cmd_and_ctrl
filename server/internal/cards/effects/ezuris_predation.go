package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Ezuri's Predation — Sorcery {5}{G}{G}{G} (EDHREC rank 1192):
//
//	"For each creature your opponents control, create a 4/4 green
//	 Phyrexian Beast creature token. Each of those tokens fights a
//	 different one of those creatures."
//
// Green's board wipe. The opponents' creatures are snapshotted at
// resolution (CR 608.2), one Beast is created per creature, and the
// N-th Beast fights the N-th creature — fight being CR 701.12, each
// deals damage equal to its power to the other, both amounts read
// before either lands (b10Fight). Lethal damage is the SBA's business
// at the next check, so a Beast that traded with a 4-power creature
// and the creature it fought die together. A creature gone by the
// time its Beast would fight it fights nothing, and the Beast stays.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "d0dd425b-fdba-41b4-b9e6-f5161610bd7e",
		Name:         "Ezuri's Predation",
		Completeness: CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			victims := MatchingBattlefield(ctx, And(Creature(), OpponentControls()))
			if len(victims) == 0 {
				return nil
			}
			before := map[uuid.UUID]bool{}
			for _, c := range ctx.Game.BattlefieldCardsForEffect() {
				before[c.InstanceID] = true
			}
			if err := (CreateToken{Controller: item.Controller, Template: b10GreenPhyrexianBeastToken(), N: len(victims)}).Apply(ctx); err != nil {
				return err
			}
			var beasts []uuid.UUID
			for _, c := range ctx.Game.BattlefieldCardsForEffect() {
				if !before[c.InstanceID] && c.Controller == item.Controller && c.Name == "Phyrexian Beast" && IsToken(c) {
					beasts = append(beasts, c.InstanceID)
				}
			}
			for i, v := range victims {
				if i >= len(beasts) {
					break
				}
				if err := b10Fight(ctx, beasts[i], v.InstanceID); err != nil {
					return err
				}
			}
			return nil
		},
	})
}
