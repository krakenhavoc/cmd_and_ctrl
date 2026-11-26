package cardsto

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Trepanation Blade — {3} Artifact — Equipment (Innistrad, S82 OLD b448, #5572):
//
//	"Whenever equipped creature attacks, defending player reveals cards from the
//	 top of their library until they reveal a land card. The creature gets +1/+0
//	 until end of turn for each card revealed this way. That player puts the
//	 revealed cards into their graveyard.
//	 Equip {2}"
//
// The defending player is the player attacked, or the controller of the
// planeswalker or battle attacked (CR 506.3). The reveal runs to the first land
// card (a token is revealed but never stops it), or the whole library if there
// is none, and a library that is empty reveals nothing. The creature that
// attacked, off the trigger's event, gets +1/+0 for every card revealed, the
// land included, and every revealed card goes to its owner's graveyard.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f0031ef3-d5f5-4114-97aa-4beabd11ff8d",
		Name:         "Trepanation Blade",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return source.IsAttachedTo(ev.CardID)
			}, "Trepanation Blade — defending player reveals cards until a land card; the creature gets +1/+0 for each; they go to the graveyard",
				func(g *game.Game, item *game.StackItem) error {
					if item.Trigger == nil {
						return nil
					}
					ctx := effects.NewContext(g, item)
					defender := effects.DefendingPlayerOfAttack(g, item.Trigger.Event)
					if defender == uuid.Nil {
						return nil
					}
					run, _ := effects.RevealUntil(ctx, defender, func(c game.Card) bool { return c.IsLand() },
						"Trepanation Blade — reveal until a land card")
					attacker := item.Trigger.Event.CardID
					if len(run) > 0 && effects.OnBattlefield(g, attacker) {
						if err := (effects.BoostUntilEOT{Target: attacker, Power: len(run), Label: "Trepanation Blade — +1/+0 for each card revealed"}).Apply(ctx); err != nil {
							return err
						}
					}
					return effects.RestIntoGraveyard(g, run)
				}),
		},
		Activated: []effects.ActivatedAbility{effects.EquipAbility("{2}")},
	})
}
