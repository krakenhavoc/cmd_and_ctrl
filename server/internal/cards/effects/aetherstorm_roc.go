package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Aetherstorm Roc — Creature — Bird {2}{W}{W}, 3/3:
//
//	"Flying
//	 Whenever this creature or another creature you control enters, you
//	 get {E} (an energy counter).
//	 Whenever this creature attacks, you may pay {E}{E}. If you do, put a
//	 +1/+1 counter on it and tap up to one target creature defending
//	 player controls."
//
// ADR 0129 §3 (#1995): "defending player" is the player this Roc is
// attacking (TargetCreatureDefendingPlayerControls), and "up to one"
// allows no target. The energy is paid as the trigger resolves; paid,
// the counter goes on the Roc if it is still the creature that attacked
// (CR 400.7) and the target is tapped.
//
// No simplification.
func init() {
	attack := whenThisAttacksMayPayEnergy("Aetherstorm Roc", 2, "put a +1/+1 counter on it and tap a creature",
		func(g *game.Game, item *game.StackItem) error {
			if err := thisStillHere(plusOneCountersOnThis(1))(g, item); err != nil {
				return err
			}
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
				return nil
			}
			return TapTarget{Target: item.Targets[0].ID}.Apply(NewContext(g, item))
		})
	attack.TargetsFrom = upToOne(TargetCreatureDefendingPlayerControls)
	Register(Spec{
		OracleID:        "06f09e5a-5cfb-437f-9dd1-1682d71f4e7f",
		Name:            "Aetherstorm Roc",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Purpose:         game.Purpose{Energy: 1},
		Triggered: []game.TriggeredAbility{
			On(game.EventETB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				c, ok := enteredUnderYourControl(ev, source, g, false)
				return ok && c.IsCreature()
			}, "Aetherstorm Roc — you get {E}", Do(GetEnergy{N: 1})),
			attack,
		},
	})
}
