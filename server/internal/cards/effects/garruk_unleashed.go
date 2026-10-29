package effects

import (
	"github.com/google/uuid"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Garruk, Unleashed — #4380, S76. Uses the existing catalog seams.
func init() {
	Register(Spec{
		OracleID: "32ea6b1c-d48e-4663-ad5c-d7f4a85450af", Name: "Garruk, Unleashed", Completeness: CompletenessFull,
		Emblem: &EmblemSpec{Label: "Garruk, Unleashed emblem", Text: "At the beginning of your end step, you may search your library for a creature card, put it onto the battlefield, then shuffle.", Triggered: []game.TriggeredAbility{AtYourEndStep("Garruk, Unleashed emblem — search for a creature", func(g *game.Game, item *game.StackItem) error {
			return SearchLibrary{Player: item.Controller, Predicate: MatchCreature, Dest: game.ZoneBattlefield, Limit: 1, Optional: true, Shuffle: true, Reason: "Garruk, Unleashed emblem — a creature onto the battlefield"}.Apply(NewContext(g, item))
		})}},
		Activated: []ActivatedAbility{
			{Label: "+1: Up to one target creature gets +3/+3 and gains trample until end of turn.", Cost: LoyaltyCost(1), Targets: TargetCreature("up to one target creature").WithCount(0, 1), Purpose: ForTargets(BenefitUntilEOT(0, 3, 3, "trample")), Effect: Do(OnTheTargetCard(func(ctx *Context, id uuid.UUID) error {
				return untilEndOfTurn(ctx, id, nil, "Garruk, Unleashed — +3/+3 and trample", game.ModifyPTMod(3, 3), game.AddKeywordsMod("trample"))
			}))},
			{Label: "−2: Create a 3/3 green Beast creature token. Then if an opponent controls more creatures than you, put a loyalty counter on Garruk.", Cost: LoyaltyCost(-2), Purpose: game.Purpose{Tokens: 1}, Effect: Do(CreateToken{Template: TokenCard("3/3 green Beast"), N: 1}, Instruction(garrukUnleashedOutnumbered))},
			{Label: "\u22127: You get an emblem with \"At the beginning of your end step, you may search your library for a creature card, put it onto the battlefield, then shuffle.\"", Cost: LoyaltyCost(-7), Purpose: game.Purpose{}, Effect: Do(CreateEmblem{})},
		},
	})
}

func garrukUnleashedOutnumbered(ctx *Context) error {
	if !b09SourceStillOnBattlefield(ctx.Game, ctx.Item) {
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
	return AddCounter{Target: ctx.Source(), Kind: game.CounterLoyalty, N: 1}.Apply(ctx)
}
