package cardsj

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Jorn, God of Winter // Kaldring, the Rimestaff — modal double-faced card,
// {2}{G} // {1}{U}{B} (Kaldheim).
//
// Front, Legendary Snow Creature — God, 3/3:
//
//	"Whenever Jorn attacks, untap each snow permanent you control."
//
// Back, Legendary Snow Artifact (registered under "<oracle_id>#1", ADR 0034):
//
//	"{T}: You may play target snow permanent card from your graveyard this
//	 turn. If you do, it enters tapped."
//
// The target is chosen as the ability goes on the stack; as it resolves the
// card gets a permission to be played (land) or cast (any other permanent) from
// the graveyard this turn, which carries "enters tapped" to the permanent it
// makes (CastPermission.EntersTapped). The permission is not used up by
// anything but that play; if the card left the graveyard it gives nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "bd43b485-e42e-4c20-adcb-799b0ce18a59",
		Name:         "Jorn, God of Winter",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisAttacks("Jorn, God of Winter — untap each snow permanent you control",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					for _, id := range effects.PermanentsControlledByMatching(g, item.Controller, effects.Snow()) {
						if err := (effects.UntapTarget{Target: id}).Apply(ctx.AsGroupMember()); err != nil {
							return err
						}
					}
					return nil
				}),
		},
	})
	effects.Register(effects.Spec{
		OracleID:     "bd43b485-e42e-4c20-adcb-799b0ce18a59#1",
		Name:         "Kaldring, the Rimestaff",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label: "{T}: You may play target snow permanent card from your graveyard this turn. If you do, it enters tapped.",
			Cost:  effects.TapCost(),
			Targets: effects.TargetCardInGraveyard("target snow permanent card in your graveyard", effects.YouOwn(), effects.Snow(),
				func(_ *game.Game, _ uuid.UUID, c game.Card) bool { return c.IsPermanent() }),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				ts := ctx.LegalTargets()
				if len(ts) == 0 {
					return nil
				}
				g.GrantCastPermissionOverCardForEffect(ts[0].ID, game.CastPermission{
					Player:       item.Controller,
					Zone:         game.ZoneGraveyard,
					Source:       ctx.Source(),
					EntersTapped: true,
					Label:        "Kaldring, the Rimestaff — you may play that card this turn; it enters tapped",
				})
				return nil
			},
		}},
	})
}
