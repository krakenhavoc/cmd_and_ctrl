package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Holy Frazzle-Cannon — Artifact — Equipment, the BACK FACE of Invasion
// of New Capenna (face 1):
//
//	"Whenever equipped creature attacks, put a +1/+1 counter on that
//	 creature and each other creature you control that shares a creature
//	 type with it.
//	 Equip {1}"
//
// Registered under "<oracle_id>#1". The attacker is the one the
// trigger saw; the creatures that share a type with it are read as the
// trigger resolves, a changeling sharing with every creature that has a
// type (CR 702.73a).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     invasionOfNewCapennaOracleID + "#1",
		Name:         "Holy Frazzle-Cannon",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return attachedCreatureAttacked(ev, source)
			}, "Holy Frazzle-Cannon — put a +1/+1 counter on that creature and each other creature you control that shares a creature type with it",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					attacker := item.Trigger.Event.CardID
					base, ok := g.LookupCardForEffect(attacker)
					if !ok {
						return nil
					}
					if err := (AddCounter{Target: attacker, Kind: game.CounterPlusOne, N: 1}).Apply(ctx); err != nil {
						return err
					}
					for _, c := range g.BattlefieldCardsForEffect() {
						c := c
						if c.InstanceID == attacker || c.Controller != item.Controller || !c.IsCreature() ||
							!game.SharesCreatureType(&base, &c) {
							continue
						}
						if err := (AddCounter{Target: c.InstanceID, Kind: game.CounterPlusOne, N: 1}).Apply(ctx.asGroupMember()); err != nil {
							return err
						}
					}
					return nil
				}),
		},
		Activated: []ActivatedAbility{
			EquipAbility("{1}"),
		},
	})
}
