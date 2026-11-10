package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Season of Gathering — Sorcery {4}{G}{G} (Bloomburrow, #4467):
//
//	"Choose up to five {P} worth of modes. You may choose the same mode
//	 more than once.
//	 {P} — Put a +1/+1 counter on a creature you control. It gains vigilance
//	 and trample until end of turn.
//	 {P}{P} — Choose artifact or enchantment. Destroy all permanents of the
//	 chosen type.
//	 {P}{P}{P} — Draw cards equal to the greatest power among creatures you
//	 control."
//
// A pawprint spell (CR 700.2i, Season of the Burrow's shape). Nothing here
// targets, and the first two choices are made as the spell resolves (the
// ruling), each bullet in printed order, a later one waiting for an earlier
// one's prompt (CR 608.2c, the mode walk): the creature for the counter is a
// choose_cards prompt over the creatures you control; the type is a two-way
// option pick, then every permanent of it is destroyed (anyone's, indestructible
// ones surviving). The draw reads the greatest power as the bullet runs, so a
// counter placed by an earlier bullet counts. With no creature the counter
// bullet and the draw do nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "3ea060d8-449c-4a5e-8883-b45bd7707b5b",
		Name:         "Season of Gathering",
		Completeness: CompletenessFull,
		Modes: ChooseUpToPawprints(5, true,
			Pawprint(1, ModeDoing("Put a +1/+1 counter on a creature you control. It gains vigilance and trample until end of turn.", nil,
				func(_ *game.StackItem, ctx *Context, _ int) error {
					return seasonOfGatheringCounter(ctx)
				})),
			// A sweep of artifacts or of enchantments, one chosen type: the
			// combined class, partial (ADR 0126).
			Pawprint(2, ModeWithPurpose(ModeDoing("Choose artifact or enchantment. Destroy all permanents of the chosen type.", nil,
				func(_ *game.StackItem, ctx *Context, _ int) error {
					return PickOption{
						Question: "Season of Gathering — choose artifact or enchantment",
						Options:  []game.ChoiceOption{{Label: "Artifact"}, {Label: "Enchantment"}},
						ThenKey:  seasonOfGatheringTypeChosen,
					}.Apply(ctx)
				}), game.Purpose{Sweep: game.Sweep{Matches: game.SweepArtifactsAndEnchantments, How: game.SweepDestroy, Partial: true}})),
			Pawprint(3, ModeDoing("Draw cards equal to the greatest power among creatures you control.", nil,
				func(_ *game.StackItem, ctx *Context, _ int) error {
					best := 0
					for _, c := range ctx.Game.BattlefieldCardsForEffect() {
						if c.Controller == ctx.Controller() && c.IsCreature() && c.CurrentPower() > best {
							best = c.CurrentPower()
						}
					}
					return DrawCards{N: best}.Apply(ctx)
				})),
		),
	})
}

// seasonOfGatheringCounter asks which creature you control gets the counter.
func seasonOfGatheringCounter(ctx *Context) error {
	g, you := ctx.Game, ctx.Controller()
	var creatures []uuid.UUID
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == you && c.IsCreature() {
			creatures = append(creatures, c.InstanceID)
		}
	}
	switch len(creatures) {
	case 0:
		return nil
	case 1:
		return seasonOfGatheringBoost(ctx, creatures[0])
	}
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:    you,
		FromPlayer: you,
		Source:     ctx.Source(),
		Question:   "Season of Gathering — put a +1/+1 counter on a creature you control; it gains vigilance and trample until end of turn",
		Cards:      creatures,
		Min:        1,
		Max:        1,
		LiveZone:   game.ZoneBattlefield,
		ThenKey:    seasonOfGatheringCreaturePicked,
		Asked:      game.NewAskedItem(ctx.Item),
	})
	return nil
}

func seasonOfGatheringBoost(ctx *Context, id uuid.UUID) error {
	if !onBattlefield(ctx.Game, id) {
		return nil
	}
	if err := (AddCounter{Target: id, Kind: game.CounterPlusOne, N: 1}).Apply(ctx); err != nil {
		return err
	}
	return GrantKeywordUntilEOT{Target: id, Keywords: []string{"vigilance", "trample"},
		Label: "Season of Gathering — vigilance and trample until end of turn"}.Apply(ctx)
}

// seasonOfGatheringCreaturePicked is the chosen creature. An on-disk key.
var seasonOfGatheringCreaturePicked = askedPick("choose-cards/season-of-gathering-creature", func(ctx *Context, r game.CardsChosen) error {
	for _, id := range r.Picked {
		return seasonOfGatheringBoost(ctx, id)
	}
	return nil
})

// seasonOfGatheringTypeChosen destroys every permanent of the chosen type.
// An on-disk key.
var seasonOfGatheringTypeChosen = OptionPickThen("option-pick/season-of-gathering-type", func(ctx *Context, r game.OptionPicked) error {
	match := Artifact()
	if r.Index == 1 {
		match = Enchantment()
	} else if r.Index != 0 {
		return nil
	}
	return DestroyAllMatching{Match: match}.Apply(ctx)
})
