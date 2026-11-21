package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Voltaic Visionary // Volt-Charged Berserker — {1}{R} Creature — Human
// Wizard 3/1 // Creature — Human Berserker 4/3 (Innistrad: Crimson Vow,
// #5495):
//
//	Front: "{T}: This creature deals 2 damage to you. Exile the top card of
//	        your library. You may play that card this turn. Activate only as
//	        a sorcery.
//	        When you play a card exiled with this creature, transform this
//	        creature."
//	Back:  "This creature can't block."
//
// The exile is a LINKED exile (CR 607.2a) stamped with this object and the
// activated row, with an until-end-of-turn play permission for its owner
// (a land too: "play"). The trigger is Prowl's half of the pair
// (PlayedCardExiledWithRow): only a card this Visionary's ability exiled
// counts, whoever plays it, and not one an earlier object exiled. The
// damage comes from the Visionary (CR 120.3): lifelink and prevention
// read it. TransformThis ignores the instruction if the Visionary has
// turned over since the trigger went on the stack (CR 701.27f).
//
// No simplification.
func init() {
	const oracle = "8ffe9e3f-fd9f-4d8b-9467-4cd301e3c8b4"
	const exileLabel = "{T}: This creature deals 2 damage to you. Exile the top card of your library. You may play that card this turn. Activate only as a sorcery."
	effects.Register(effects.Spec{
		OracleID:     oracle,
		Name:         "Voltaic Visionary",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:        exileLabel,
			Cost:         effects.TapCost(),
			SorcerySpeed: true,
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				if err := (effects.DealDamage{Source: ctx.Source(), Target: item.Controller, Amount: 2}).Apply(ctx); err != nil {
					return err
				}
				top := g.LookAtTopOfLibraryForEffect(item.Controller, 1)
				if len(top) == 0 {
					return nil
				}
				return effects.ExileWithPermission{Target: top[0], GrantTo: item.Controller, Linked: true}.Apply(ctx)
			},
		}},
		Triggered: []game.TriggeredAbility{
			effects.OnAny([]game.EventKind{game.EventCast, game.EventZoneMove}, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return game.PlayedCardExiledWithRow(ev,
					game.PermissionCardRef{ID: source.InstanceID, Epoch: source.ObjectEpoch},
					game.AbilitySlotActivated, exileLabel)
			}, "Voltaic Visionary — you played a card exiled with it: transform it", effects.Do(effects.TransformThis{})),
		},
	})
	effects.Register(effects.Spec{
		OracleID:     oracle + "#1",
		Name:         "Volt-Charged Berserker",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{effects.RestrictSelf(game.CantBlock)},
	})
}
