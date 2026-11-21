package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// C.A.M.P. — Artifact — Fortification {3} (Fallout, #5948):
//
//	"Whenever fortified land is tapped for mana, put a +1/+1 counter on
//	 target creature you control. If that creature shares a color with the
//	 mana that land produced, create a Junk token.
//	 Fortify {3}"
//
// The catalog's first Fortification: fortify is equip's twin for a land
// (CR 702.67a, s82GapB60FortifyAbility), and a Fortification attaches to a
// land and falls off a permanent that is not one (CR 301.6, CR 704.5n).
//
// It is a triggered ability, not a mana ability: it adds no mana, so it uses
// the stack (CR 605.1b). It watches EventManaAbilityResolved (Tyvar the
// Bellicose's event, written once the land's mana is in the pool) for the
// fortified land. The colours of the mana the land produced are read at
// resolution off that production's EventManaAdded events, and the creature's
// colours as they are then.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "9bf498ce-6d6d-462f-9ec3-b36a2184aff7",
		Name:         "C.A.M.P.",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.On(game.EventManaAbilityResolved,
					func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
						if source == nil || ev.Source == source.InstanceID || !source.IsAttachedTo(ev.Source) {
							return false
						}
						land, ok := g.LookupCardForEffect(ev.Source)
						return ok && land.IsLand() && land.Tapped
					},
					"C.A.M.P. — put a +1/+1 counter on target creature you control; if it shares a color with the mana that land produced, create a Junk token",
					func(g *game.Game, item *game.StackItem) error {
						ctx := effects.NewContext(g, item)
						t, ok := ctx.ClauseTarget(0)
						if !ok {
							return nil
						}
						if err := (effects.AddCounter{Target: t.ID, Kind: game.CounterPlusOne, N: 1}).Apply(ctx); err != nil {
							return err
						}
						creature, ok := g.LookupCardForEffect(t.ID)
						if !ok || !effects.OnBattlefield(g, t.ID) {
							return nil
						}
						ev := item.Trigger.Event
						for _, produced := range effects.S82GapB60ManaColorsOfLastTap(g, ev.Source, ev.Seq) {
							for _, col := range creature.EffectiveColors() {
								if col == produced {
									return effects.CreateToken{Template: effects.JunkToken(), N: 1}.Apply(ctx)
								}
							}
						}
						return nil
					}),
				effects.TargetCreature("target creature you control", effects.YouControl())),
		},
		Activated: []effects.ActivatedAbility{effects.S82GapB60FortifyAbility("{3}")},
	})
}
