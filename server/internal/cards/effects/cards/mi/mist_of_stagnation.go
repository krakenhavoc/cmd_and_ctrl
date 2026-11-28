package cardsmi

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Mist of Stagnation — Enchantment {3}{U}{U} (Judgment):
//
//	"Permanents don't untap during their controllers' untap steps.
//	 At the beginning of each player's upkeep, that player chooses a
//	 permanent for each card in their graveyard, then untaps those
//	 permanents."
//
// Back to Basics' untap-step restriction with no filter, and an every-upkeep
// trigger: the upkeep's player picks up to one permanent per card in their
// own graveyard, counted as the trigger resolves, from any permanent on the
// battlefield, and untaps them. Picking the same permanent twice would untap
// it once, so a distinct pick of up to that many is the whole choice.
//
// No simplification.
func init() {
	const name = "Mist of Stagnation"
	effects.Register(effects.Spec{
		OracleID:              "e78673c4-a969-443f-b1a4-0eb37b8111af",
		Name:                  name,
		Completeness:          effects.CompletenessFull,
		UntapStepRestrictions: []game.UntapStepRestriction{effects.DoesntUntapDuringTheirControllersUntapSteps(nil)},
		Triggered: []game.TriggeredAbility{
			effects.AtEachUpkeep(name+" — that player chooses a permanent for each card in their graveyard, then untaps those permanents",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					player := ctx.Trigger().Event.Actor
					p := g.PlayerByIDForEffect(player)
					if p == nil || p.Graveyard == nil || p.Graveyard.Size() == 0 {
						return nil
					}
					n := p.Graveyard.Size()
					cands := g.BattlefieldCardsForEffect()
					ids := make([]uuid.UUID, 0, len(cands))
					for _, c := range cands {
						ids = append(ids, c.InstanceID)
					}
					if len(ids) == 0 {
						return nil
					}
					if n > len(ids) {
						n = len(ids)
					}
					g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
						Chooser:  player,
						Source:   item.SourceCardID,
						Question: name + " — choose a permanent for each card in your graveyard; untap them",
						Cards:    ids,
						Min:      n,
						Max:      n,
						LiveZone: game.ZoneBattlefield,
						ThenKey:  mistOfStagnationPicked,
						Asked:    game.NewAskedItem(item),
					})
					return nil
				}),
		},
	})
}

// mistOfStagnationPicked untaps the chosen permanents. An on-disk key: never
// renamed, never reused.
var mistOfStagnationPicked = effects.AskedPick("choose-cards/s82-old-b626-mist-of-stagnation-untap", func(ctx *effects.Context, r game.CardsChosen) error {
	return effects.UntapEach(ctx, r.Picked)
})
