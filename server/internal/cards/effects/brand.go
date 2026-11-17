package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Brand — Instant {R} (Urza's Saga):
//
//	"Gain control of all permanents you own. (This effect lasts
//	 indefinitely.)
//	 Cycling {2} ({2}, Discard this card: Draw a card.)"
//
// The set is every permanent the caster owns as Brand resolves (CR
// 611.2c), the ones they already control included, so a permanent
// they hold back only until end of turn stays theirs once that ends.
// Each one gets a layer 2 control effect with no stated duration (CR
// 613.1b, CR 611.2a), so a later control change still takes it. A
// permanent that enters afterwards is not affected.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "eec7500b-418d-4387-beb2-d53e047c1d22",
		Name:         "Brand",
		Completeness: CompletenessFull,
		Activated:    []ActivatedAbility{Cycling("{2}")},
		OnResolve: func(item *game.StackItem, spell *Context) error {
			ctx := spell.asGroupMember()
			var owned []uuid.UUID
			for _, c := range ctx.Game.BattlefieldCardsForEffect() {
				if c.Owner == item.Controller {
					owned = append(owned, c.InstanceID)
				}
			}
			for _, id := range owned {
				if err := (GainControl{
					Target:   id,
					Duration: game.IndefiniteDuration(),
					Label:    "Brand — gain control of all permanents you own",
				}).Apply(ctx); err != nil {
					return err
				}
			}
			return nil
		},
	})
}
