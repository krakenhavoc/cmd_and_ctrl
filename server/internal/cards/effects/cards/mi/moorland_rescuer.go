package cardsmi

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Moorland Rescuer — Creature — Human Knight {5}{W}, 4/4 (Midnight Hunt
// Commander, #5497):
//
//	"When this creature dies, return any number of other creature cards
//	 with total power X or less from your graveyard to the battlefield,
//	 where X is this creature's power. Exile this card."
//
// X is the Rescuer's power as it last existed on the battlefield (CR
// 603.10a, PermanentForEffect). The pick is Rip, Spawn Hunter's
// choose_cards prompt with a set rule over what is picked, Protean
// Hulk's "total value" shape. A set rule is a function of the cards alone
// (it must survive a restore), so one rule is registered for each ceiling
// and resolution picks the one for X, clamped to the most power the
// graveyard could ever offer: a ceiling nothing can exceed is no ceiling
// (Smelting Vat's clamp). The chosen cards still in the graveyard enter
// together as one entry (CR 603.6a); then the Rescuer's own card is
// exiled if it is still in the graveyard, and exiled even when nothing was
// returned or the player picked none. A card that left the graveyard in
// response is left alone, and the Rescuer is never among the candidates.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f687a746-4807-41d5-848d-4053ddef6a76",
		Name:         "Moorland Rescuer",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerValue},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisDies("Moorland Rescuer — return any number of other creature cards with total power X or less from your graveyard to the battlefield, then exile this card", moorlandRescuerDies),
		},
	})
}

// moorlandRescuerPicks are the registered picks, index = ceiling, with
// one unbounded pick at the end.
var moorlandRescuerPicks = func() []game.ChooseCardsThen {
	picks := make([]game.ChooseCardsThen, 0, effects.S82MidTotalPowerRuleMax+2)
	for n := 0; n <= effects.S82MidTotalPowerRuleMax; n++ {
		picks = append(picks, effects.AskedPickRule(effects.S82MidTotalPowerKey(n), effects.S82MidTotalPowerAtMost(n), moorlandRescuerPicked))
	}
	return append(picks, effects.AskedPickRule("choose-cards/moorland-rescuer-total-power-unbounded", nil, moorlandRescuerPicked))
}()

func moorlandRescuerDies(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	x := 0
	if info, ok := ctx.SourcePermanent(); ok {
		x = info.Power
	}
	src := ctx.Source()
	var cands []uuid.UUID
	sum := 0
	if x > 0 {
		for _, id := range effects.GraveyardCardIDs(ctx, ctx.Controller(), func(c game.Card) bool { return c.IsCreature() && c.InstanceID != src }) {
			c, ok := g.LookupCardForEffect(id)
			if !ok || c.PowerForComparison() > x {
				continue
			}
			cands = append(cands, id)
			sum += max(c.PowerForComparison(), 0)
		}
	}
	if len(cands) == 0 {
		return moorlandRescuerExileSelf(ctx, src)
	}
	pick := moorlandRescuerPicks[len(moorlandRescuerPicks)-1]
	if x < sum && x <= effects.S82MidTotalPowerRuleMax {
		pick = moorlandRescuerPicks[max(x, 0)]
	}
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:    item.Controller,
		FromPlayer: item.Controller,
		Source:     item.SourceCardID,
		Question:   "Moorland Rescuer — return any number of other creature cards with total power X or less to the battlefield",
		Cards:      cands,
		Min:        0,
		Max:        len(cands),
		LiveZone:   game.ZoneGraveyard,
		ThenKey:    pick,
		Asked:      game.NewAskedItem(item),
	})
	return nil
}

func moorlandRescuerPicked(ctx *effects.Context, r game.CardsChosen) error {
	if err := (effects.ReturnFromGraveyardTogether{Targets: r.Picked}).Apply(ctx); err != nil {
		return err
	}
	return moorlandRescuerExileSelf(ctx, ctx.Source())
}

// moorlandRescuerExileSelf exiles the Rescuer's card if it is still in
// a graveyard.
func moorlandRescuerExileSelf(ctx *effects.Context, id uuid.UUID) error {
	if z := ctx.Game.FindCardZoneForEffect(id); z == nil || z.Kind != game.ZoneGraveyard {
		return nil
	}
	return effects.ExileTarget{Target: id}.Apply(ctx)
}
