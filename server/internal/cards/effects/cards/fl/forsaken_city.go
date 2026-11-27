package cardsfl

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Forsaken City — Land (Planeshift, S82 OLD b644, #5615):
//
//	"This land doesn't untap during your untap step.
//	 At the beginning of your upkeep, you may exile a card from your hand. If you
//	 do, untap this land.
//	 {T}: Add one mana of any color."
//
// Basalt Monolith's untap restriction (CR 502.3). The upkeep trigger is Chrome Mox's
// optional exile pick from the hand (a floor of zero is the "you may"), keyed; the
// land untaps only if a card was exiled ("if you do"), and only if it is still the
// same object (CR 400.7).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:              "6bb00a28-8b5a-4049-93b7-3db02de88aeb",
		Name:                  "Forsaken City",
		Completeness:          effects.CompletenessFull,
		UntapStepRestrictions: []game.UntapStepRestriction{effects.DoesntUntapDuringYourUntapStep()},
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{W|U|B|R|G}",
			Label:    "Add one mana of any color",
		}},
		Triggered: []game.TriggeredAbility{
			effects.AtYourUpkeep("Forsaken City — you may exile a card from your hand. If you do, untap this land", forsakenCityUpkeep),
		},
	})
}

// forsakenCityUpkeep queues the optional pick. It queues and does not finish: the
// exile and the untap happen when the controller answers.
func forsakenCityUpkeep(g *game.Game, item *game.StackItem) error {
	p := g.PlayerByIDForEffect(item.Controller)
	if p == nil || p.Hand == nil || p.Hand.Size() == 0 {
		return nil
	}
	ids := make([]uuid.UUID, 0, p.Hand.Size())
	for _, c := range p.Hand.Cards {
		ids = append(ids, c.InstanceID)
	}
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:  item.Controller,
		Source:   item.SourceCardID,
		Question: "Forsaken City — you may exile a card from your hand to untap this land",
		Cards:    ids,
		Min:      0,
		Max:      1,
		LiveZone: game.ZoneHand,
		ThenKey:  forsakenCityExiled,
		Asked:    game.NewAskedItem(item),
	})
	return nil
}

// forsakenCityExiled exiles the chosen card and, if it reached exile, untaps the
// land. An on-disk key.
var forsakenCityExiled = effects.AskedPick("choose-cards/forsaken-city-exile-to-untap", func(ctx *effects.Context, r game.CardsChosen) error {
	if len(r.Picked) == 0 {
		return nil
	}
	if err := effects.NothingIfGone(ctx.Game.ExileCardForEffect(r.Picked[0])); err != nil {
		return err
	}
	if ctx.IsNewSourceObject(ctx.Source()) || !effects.OnBattlefield(ctx.Game, ctx.Source()) {
		return nil
	}
	return effects.UntapTarget{Target: ctx.Source()}.Apply(ctx)
})
