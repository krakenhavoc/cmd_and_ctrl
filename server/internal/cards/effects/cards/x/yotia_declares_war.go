package cardsx

import (
	"strconv"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Yotia Declares War — Enchantment — Saga {1}{R} (Dominaria United):
//
//	"Read ahead (Choose a chapter and start with that many lore
//	 counters. Add one after your draw step. Skipped chapters don't
//	 trigger. Sacrifice after III.)
//	 I — Create a 0/2 colorless Thopter artifact creature token with
//	 flying named Ornithopter.
//	 II — Tap any number of untapped artifacts you control. When you do,
//	 this Saga deals that much damage to target creature or planeswalker.
//	 III — Up to one target artifact you control becomes an artifact
//	 creature with base power and toughness 4/4 until end of turn."
//
// Read ahead is the engine's (CR 702.155). Chapter II is two stack items,
// as printed (Teferi Akosa of Zhalfir's -3 shape): the artifacts are
// chosen, any number, none included, and tapping none does nothing; each
// pick is re-read before it is tapped, so an artifact that left, was
// tapped or was stolen in the meantime is not "tapped this way" and adds
// no damage. Tapping one or more starts the reflexive trigger (CR 603.12),
// which picks its target as it goes on the stack, the damage fixed at
// the number tapped. Chapter III is one continuous effect (CR 611.2a):
// the artifact becomes an artifact creature (CR 205.1a) with base 4/4
// (layer 7b) until end of turn.
//
// No simplification.
func init() {
	const name = "Yotia Declares War"
	thopter := effects.ChapterTrigger(1, effects.SagaChapterLabel(name, 1, "create a 0/2 Thopter named Ornithopter"),
		func(g *game.Game, item *game.StackItem) error {
			return effects.CreateToken{
				Controller: item.Controller,
				Template:   effects.TokenCard("0/2 colorless Thopter artifact with flying named Ornithopter"),
				N:          1,
			}.Apply(effects.NewContext(g, item))
		})
	thopter.Purpose = game.Purpose{Tokens: 1, CreatureTokens: 1}
	effects.Register(effects.Spec{
		OracleID:        "4ce785b9-aee0-4ac6-ab20-ee20d0551be4",
		Name:            name,
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{game.KeywordReadAhead},
		Triggered: []game.TriggeredAbility{
			thopter,
			effects.ChapterTrigger(2, effects.SagaChapterLabel(name, 2, "tap any number of untapped artifacts; damage equal to the number tapped"),
				yotiaTapArtifacts),
			effects.ChapterTriggerTargeting(3, effects.SagaChapterLabel(name, 3, "up to one target artifact becomes a 4/4 artifact creature"),
				effects.UpToOneTargetPermanent("up to one target artifact you control", effects.Artifact(), effects.YouControl()), yotiaAnimate),
		},
	})
}

// yotiaTapArtifacts is chapter II's first half: the pick.
func yotiaTapArtifacts(g *game.Game, item *game.StackItem) error {
	var candidates []uuid.UUID
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == item.Controller && c.IsArtifact() && !c.Tapped {
			candidates = append(candidates, c.InstanceID)
		}
	}
	if len(candidates) == 0 {
		return nil
	}
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:  item.Controller,
		Source:   item.SourceCardID,
		Question: "Yotia Declares War — tap any number of untapped artifacts you control",
		Cards:    candidates,
		Min:      0,
		Max:      0, // "any number"
		LiveZone: game.ZoneBattlefield,
		ThenKey:  yotiaTapPick,
		Asked:    game.NewAskedItem(item),
	})
	return nil
}

// yotiaTapPick taps the picks that are still untapped artifacts its
// controller controls and, if at least one was tapped, creates the
// reflexive trigger carrying the count as the damage.
var yotiaTapPick = effects.AskedPick("choose-cards/yotia-declares-war-tap", func(ctx *effects.Context, r game.CardsChosen) error {
	var tapped []uuid.UUID
	for _, id := range r.Picked {
		c, ok := ctx.Game.LookupCardForEffect(id)
		if !ok || c.Tapped || !c.IsArtifact() || c.Controller != ctx.Item.Controller {
			continue
		}
		if err := (effects.TapTarget{Target: id}).Apply(ctx.AsGroupMember()); err != nil {
			return err
		}
		tapped = append(tapped, id)
	}
	if len(tapped) == 0 {
		return nil
	}
	return effects.ReflexiveTrigger{
		Label:  "Yotia Declares War — " + strconv.Itoa(len(tapped)) + " damage to target creature or planeswalker",
		Cards:  tapped,
		Body:   effects.YotiaDamageBody,
		Params: game.EffectParams{Amount: len(tapped)},
	}.Apply(ctx)
})

// yotiaAnimate is chapter III.
func yotiaAnimate(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	t, ok := ctx.ClauseTarget(0)
	if !ok || t.Kind != game.TargetCard || !effects.OnBattlefield(g, t.ID) {
		return nil
	}
	return effects.S82BroB11BecomesArtifactCreature(ctx, t.ID, 4, 4, effects.DurationUntilEndOfTurn(ctx),
		"Yotia Declares War — a 4/4 artifact creature until end of turn")
}
