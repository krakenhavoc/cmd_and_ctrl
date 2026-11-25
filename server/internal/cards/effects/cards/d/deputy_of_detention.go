package cardsd

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Deputy of Detention — Creature — Vedalken Wizard {1}{W}{U}, 1/3 (Ravnica
// Allegiance):
//
//	"When this creature enters, exile target nonland permanent an opponent
//	 controls and all other nonland permanents that player controls with the same
//	 name as that permanent until this creature leaves the battlefield."
//
// The target and every other nonland permanent its controller has under the
// same name, found as the trigger resolves, are exiled one after another until
// the Deputy leaves (Abdel Adrian's chain of keyed exiles, since an exile can
// stop to ask a question). Tokens exiled this way are gone for good.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0e4150db-ac43-48b4-9791-0d874906acf5",
		Name:         "Deputy of Detention",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WhenThisEnters("Deputy of Detention — exile target nonland permanent an opponent controls and all other nonland permanents that player controls with the same name until this creature leaves the battlefield",
					func(g *game.Game, item *game.StackItem) error {
						ctx := effects.NewContext(g, item)
						victim, ok := effects.B16FirstLegalTargetCard(ctx)
						if !ok {
							return nil
						}
						c, ok := g.LookupCardForEffect(victim)
						if !ok {
							return nil
						}
						ids := []uuid.UUID{victim}
						for _, o := range g.BattlefieldCardsForEffect() {
							if o.InstanceID != victim && !o.IsLand() && o.Controller == c.Controller && o.Name == c.Name {
								ids = append(ids, o.InstanceID)
							}
						}
						return deputyOfDetentionExileNext(ctx, ids)
					}),
				effects.TargetPermanent("target nonland permanent an opponent controls", effects.Nonland(), effects.OpponentControls())),
		},
	})
}

func deputyOfDetentionExileNext(ctx *effects.Context, ids []uuid.UUID) error {
	if len(ids) == 0 {
		return nil
	}
	id, rest := ids[0], ids[1:]
	t := ctx.ItemTail(deputyOfDetentionExiledOne)
	t.Carry = append([]uuid.UUID(nil), rest...)
	return effects.ExileUntil{
		Target:     id,
		ThisLeaves: true,
		Label:      "Deputy of Detention — returns when Deputy of Detention leaves the battlefield",
		ThenTail:   t,
	}.Apply(ctx)
}

var deputyOfDetentionExiledOne game.TailRef

func init() {
	deputyOfDetentionExiledOne = effects.CardTail("then/deputy-of-detention-exiled-one", func(ctx *effects.Context, o game.TailOutcome) error {
		return deputyOfDetentionExileNext(ctx, o.Tail.Carry)
	})
}
