package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Mardu Woe-Reaper — {W} Creature — Human Warrior 2/1 (Fate Reforged):
//
//	"Whenever this creature or another Warrior you control enters, you may
//	 exile target creature card from a graveyard. If you do, you gain 1 life."
//
// Carrion Imp's shape on a trigger that also watches other Warriors entering
// under your control. The target is chosen as the trigger goes on the stack
// (CR 603.3d); "you may" is asked as it resolves; the life follows only an
// exile that happened. The Warrior type is a word of this card's text.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "636ed3e3-c31d-403a-a3e7-03e01350ce4f",
		Name:         "Mardu Woe-Reaper",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Optional(effects.Targeting(
				effects.On(game.EventETB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					c, ok := effects.EnteredUnderYourControl(ev, source, g, false)
					if !ok || !c.IsCreature() {
						return false
					}
					return c.InstanceID == source.InstanceID || c.HasSubtype(source.TextSubtype("Warrior"))
				}, "Mardu Woe-Reaper — you may exile target creature card from a graveyard; if you do, you gain 1 life",
					func(g *game.Game, item *game.StackItem) error {
						ctx := effects.NewContext(g, item)
						t, ok := effects.FirstLegalTarget(ctx)
						if !ok {
							return nil
						}
						if err := (effects.ExileTarget{Target: t.ID}).Apply(ctx); err != nil {
							return err
						}
						if z := g.FindCardZoneForEffect(t.ID); z == nil || z.Kind != game.ZoneExile {
							return nil
						}
						return effects.GainLife{Amount: 1}.Apply(ctx)
					}),
				effects.TargetCardInGraveyard("target creature card from a graveyard", effects.Creature())),
				"Mardu Woe-Reaper — exile the creature card?"),
		},
	})
}
