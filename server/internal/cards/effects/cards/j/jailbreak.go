package cardsj

import (
	"strconv"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Jailbreak — Sorcery {1}{W} (Streets of New Capenna):
//
//	"Return target permanent card in an opponent's graveyard to the
//	 battlefield under their control. When that permanent enters, return
//	 up to one target permanent card with equal or lesser mana value from
//	 your graveyard to the battlefield."
//
// The first half returns the card to its owner (it is in an opponent's
// graveyard, so "their control" is its owner's). The second half is a
// triggered ability created only once the permanent has actually entered
// (CR 603.7a, 603.12): a card that left the graveyard in response, or that
// a replacement kept out, creates no trigger. It is queued as a reflexive
// trigger whose clause carries the returned card's mana value as read as
// it entered; the target is chosen as the trigger goes on the stack
// (CR 603.3d), from your graveyard, and is re-checked on resolution.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "169cf74a-07bf-4841-83f3-904df8a0a39b",
		Name:         "Jailbreak",
		Completeness: effects.CompletenessFull,
		Targets: effects.TargetCardInGraveyard("target permanent card in an opponent's graveyard",
			effects.Permanent(), effects.Not(effects.YouOwn())),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			for _, t := range ctx.LegalTargets() {
				if t.Kind != game.TargetCard {
					continue
				}
				card, ok := ctx.Game.LookupCardForEffect(t.ID)
				if !ok {
					return nil
				}
				mv, _ := ctx.Game.ManaValueForEffect(card)
				if err := (effects.ReturnFromGraveyard{Target: t.ID, Dest: game.ZoneBattlefield}).Apply(ctx); err != nil {
					return err
				}
				if !effects.OnBattlefield(ctx.Game, t.ID) {
					return nil
				}
				return effects.ReflexiveTrigger{
					Label:  "Jailbreak — return up to one target permanent card with equal or lesser mana value from your graveyard to the battlefield",
					Cards:  []uuid.UUID{t.ID},
					Body:   jailbreakReturnBody,
					Params: game.EffectParams{Amount: mv},
				}.Apply(ctx)
			}
			return nil
		},
	})
}

// jailbreakReturnBody is the "When that permanent enters" half. An on-disk
// key.
var jailbreakReturnBody = game.ReflexiveBody("jailbreak/return-equal-or-lesser",
	effects.SimpleBody(effects.ReturnFirstLegalGraveyardTargetToBattlefield),
	func(_ uuid.UUID, p game.EffectParams) *game.TargetSpec {
		return effects.TargetCardInGraveyard("up to one target permanent card with mana value "+strconv.Itoa(p.Amount)+" or less from your graveyard",
			effects.YouOwn(), effects.Permanent(), effects.ManaValueLE(p.Amount)).WithCount(0, 1)
	})
