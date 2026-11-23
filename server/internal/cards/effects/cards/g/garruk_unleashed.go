package cardsg

import (
	"github.com/google/uuid"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Garruk, Unleashed — #4380, S76. Uses the existing catalog seams.
func init() {
	effects.Register(effects.Spec{
		OracleID: "32ea6b1c-d48e-4663-ad5c-d7f4a85450af", Name: "Garruk, Unleashed", Completeness: effects.CompletenessFull,
		Emblem: &effects.EmblemSpec{Label: "Garruk, Unleashed emblem", Text: "At the beginning of your end step, you may search your library for a creature card, put it onto the battlefield, then shuffle.", Triggered: []game.TriggeredAbility{effects.AtYourEndStep("Garruk, Unleashed emblem — search for a creature", func(g *game.Game, item *game.StackItem) error {
			return effects.SearchLibrary{Player: item.Controller, Predicate: effects.MatchCreature, Dest: game.ZoneBattlefield, Limit: 1, Optional: true, Shuffle: true, Reason: "Garruk, Unleashed emblem — a creature onto the battlefield"}.Apply(effects.NewContext(g, item))
		})}},
		Activated: []effects.ActivatedAbility{
			{Label: "+1: Up to one target creature gets +3/+3 and gains trample until end of turn.", Cost: effects.LoyaltyCost(1), Targets: effects.TargetCreature("up to one target creature").WithCount(0, 1), Purpose: effects.ForTargets(effects.BenefitUntilEOT(0, 3, 3, "trample")), Effect: effects.Do(effects.OnTheTargetCard(func(ctx *effects.Context, id uuid.UUID) error {
				return effects.UntilEndOfTurn(ctx, id, nil, "Garruk, Unleashed — +3/+3 and trample", game.ModifyPTMod(3, 3), game.AddKeywordsMod("trample"))
			}))},
			{Label: "−2: Create a 3/3 green Beast creature token. Then if an opponent controls more creatures than you, put a loyalty counter on Garruk.", Cost: effects.LoyaltyCost(-2), Purpose: game.Purpose{Tokens: 1}, Effect: effects.Do(effects.CreateToken{Template: effects.TokenCard("3/3 green Beast"), N: 1}, effects.Instruction(garrukUnleashedOutnumbered))},
			{Label: "\u22127: You get an emblem with \"At the beginning of your end step, you may search your library for a creature card, put it onto the battlefield, then shuffle.\"", Cost: effects.LoyaltyCost(-7), Purpose: game.Purpose{}, Effect: effects.Do(effects.CreateEmblem{})},
		},
	})
}

func garrukUnleashedOutnumbered(ctx *effects.Context) error {
	if !effects.B09SourceStillOnBattlefield(ctx.Game, ctx.Item) {
		return nil
	}
	ctx.Game.RecomputeLayersIfStaleLocked()
	counts := map[uuid.UUID]int{}
	for _, c := range ctx.Game.BattlefieldCardsForEffect() {
		if c.IsCreature() {
			counts[c.Controller]++
		}
	}
	outnumbered := false
	for _, id := range ctx.Opponents() {
		if counts[id] > counts[ctx.Controller()] {
			outnumbered = true
			break
		}
	}
	if !outnumbered {
		return nil
	}
	return effects.AddCounter{Target: ctx.Source(), Kind: game.CounterLoyalty, N: 1}.Apply(ctx)
}
