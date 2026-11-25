package cardsp

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Pia's Revolution — Enchantment {2}{R} (Aether Revolt, S82 older sets, #5536):
//
//	"Whenever a nontoken artifact is put into your graveyard from the
//	 battlefield, return that card to your hand unless target opponent has
//	 this enchantment deal 3 damage to them."
//
// The trigger reads the died permanent off last-known information
// (CR 603.10a): it was a nontoken artifact as it left the battlefield and
// it went to its owner's graveyard, which is "your graveyard" (CR 404.2).
// The opponent is the trigger's target (CR 603.3d) and is the one asked as
// it resolves: yes has Pia's Revolution deal 3 damage to that opponent, no
// returns the card to the trigger controller's hand. A card that has left
// the graveyard by then is not returned (CR 608.2b); an opponent who is no
// longer a legal target stops the whole trigger (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "edce515d-828e-49ae-8651-6d14f7470a62",
		Name:         "Pia's Revolution",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.On(game.EventLTB, piasRevolutionArtifactDied,
				"Pia's Revolution — return that card to your hand unless target opponent has this enchantment deal 3 damage to them",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					if item.Trigger == nil {
						return nil
					}
					card := item.Trigger.Event.CardID
					var opp uuid.UUID
					for _, t := range ctx.LegalTargets() {
						if t.Kind == game.TargetPlayer {
							opp = t.ID
						}
					}
					if opp == uuid.Nil || card == uuid.Nil {
						return nil
					}
					return effects.MayChoice{
						Player:   opp,
						Question: "Pia's Revolution — have it deal 3 damage to you? If you don't, its controller returns the artifact card to their hand.",
						YesLabel: "Take 3 damage",
						NoLabel:  "Let them return it",
						Then:     piasRevolutionAnswer,
						Carry:    []uuid.UUID{card},
					}.Apply(ctx)
				}),
				effects.TargetPlayer("target opponent", effects.Opponent())),
		},
	})
}

// piasRevolutionArtifactDied is "a nontoken artifact is put into your
// graveyard from the battlefield".
func piasRevolutionArtifactDied(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.NewZone != game.ZoneGraveyard {
		return false
	}
	c, ok := g.LookupCardForEffect(ev.CardID)
	return ok && !c.IsToken() && c.Owner == source.Controller && effects.LeftAsType(ev, c, "artifact")
}

var piasRevolutionAnswer = effects.MayYesNo("confirm/pias-revolution-opponent-answer",
	func(ctx *effects.Context, r game.ConfirmAnswer) error {
		return effects.DealDamage{Source: ctx.Source(), Target: r.Chooser, Amount: 3}.Apply(ctx)
	},
	func(ctx *effects.Context, r game.ConfirmAnswer) error {
		if len(r.Carry) == 0 {
			return nil
		}
		z := ctx.Game.FindCardZoneForEffect(r.Carry[0])
		if z == nil || z.Kind != game.ZoneGraveyard || z.Owner != ctx.Controller() {
			return nil
		}
		return effects.ReturnFromGraveyard{Target: r.Carry[0], Dest: game.ZoneHand}.Apply(ctx)
	})
