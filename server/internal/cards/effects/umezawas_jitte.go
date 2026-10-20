package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Umezawa's Jitte banks two charge counters whenever its equipped
// creature deals combat damage, including damage to a blocker. Its
// counter cost and modal choice are made when activated, and only the
// -1/-1 bullet targets. The pump reads the equipped creature when it
// resolves, using the Equipment's last-known attachment if it left.
func init() {
	Register(Spec{
		OracleID:     "1da10d5c-36a8-473f-a5d4-782ad61d8057",
		Name:         "Umezawa's Jitte",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventDealDamage, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return b43AttachedCreatureDealtCombatDamage(ev, source, g)
			}, "Umezawa's Jitte — put two charge counters on it", func(g *game.Game, item *game.StackItem) error {
				if !b09SourceStillOnBattlefield(g, item) {
					return nil
				}
				return AddCounter{Target: item.SourceCardID, Kind: "charge", N: 2}.Apply(NewContext(g, item))
			}),
		},
		Activated: []ActivatedAbility{
			{
				Label:   "Remove a charge counter from Umezawa's Jitte: Choose one — equipped creature gets +2/+2 until end of turn; or target creature gets -1/-1 until end of turn; or you gain 2 life.",
				Cost:    RemoveCountersFromThis("charge", 1),
				Purpose: game.Purpose{Answers: game.AnswerPump},
				Modes: ChooseOne(
					ModeDoing("Equipped creature gets +2/+2 until end of turn.", nil, func(_ *game.StackItem, ctx *Context, _ int) error {
						source, ok := ctx.SourcePermanent()
						if !ok || source.AttachedObject == nil {
							return nil
						}
						host, ok := ctx.Game.PermanentForEffect(*source.AttachedObject)
						if !ok || host.Left {
							return nil
						}
						creature, found := ctx.Game.LookupCardForEffect(source.AttachedObject.ID)
						if !found || !creature.IsCreature() {
							return nil
						}
						return BoostUntilEOT{Target: source.AttachedObject.ID, Power: 2, Toughness: 2, Label: "Umezawa's Jitte — +2/+2"}.Apply(ctx)
					}),
					ModeDoing("Target creature gets -1/-1 until end of turn.", TargetCreature("target creature"), func(_ *game.StackItem, ctx *Context, occ int) error {
						target, ok := ModeTarget(ctx, occ)
						if !ok {
							return nil
						}
						return BoostUntilEOT{Target: target.ID, Power: -1, Toughness: -1, Label: "Umezawa's Jitte — -1/-1"}.Apply(ctx)
					}),
					ModeDoing("You gain 2 life.", nil, func(_ *game.StackItem, ctx *Context, _ int) error { return GainLife{Amount: 2}.Apply(ctx) }),
				),
			},
			EquipAbility("{2}"),
		},
	})
}
