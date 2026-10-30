package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// S.H.I.E.L.D. Spy Kit — Artifact — Equipment {W} (Marvel Super Heroes,
// #4455):
//
//	"Equipped creature gets +1/+1.
//	 Whenever equipped creature attacks alone, untap it and scry 1.
//	 Equip {1}"
//
// Captain America's Shield's equipped-creature attack watch narrowed to a
// lone attacker (CR 506.5: the engine stamps every attacker of a declaration
// before any attack event is announced, so a lone declaration reads as
// alone). The untap reads the attacker off the event; an attacker that has
// already left is untapped by no one.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "6fd44fb9-3433-4cb0-95a4-4ac8170b2166",
		Name:         "S.H.I.E.L.D. Spy Kit",
		Completeness: CompletenessFull,
		Static:       []game.StaticAbility{PumpAttached(1, 1)},
		Triggered: []game.TriggeredAbility{
			On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return attachedCreatureAttacked(ev, source) && game.AttackedAlone(g)
			}, "S.H.I.E.L.D. Spy Kit — untap the equipped creature and scry 1", mshB7UntapAttackerAndScry),
		},
		Activated: []ActivatedAbility{EquipAbility("{1}")},
	})
}

func mshB7UntapAttackerAndScry(g *game.Game, item *game.StackItem) error {
	if item.Trigger == nil {
		return nil
	}
	ctx := NewContext(g, item)
	if err := (UntapTarget{Target: item.Trigger.Event.CardID}).Apply(ctx); err != nil {
		return err
	}
	return Scry{N: 1}.Apply(ctx)
}
