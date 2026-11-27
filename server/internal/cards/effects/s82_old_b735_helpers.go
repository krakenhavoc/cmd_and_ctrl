package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b735_helpers.go — S82 OLD batch 735 (Portal Second Age, #5625).

// S82OldB735GoblinWarCry is Goblin War Cry's body: "Target opponent chooses a
// creature they control. Other creatures they control can't block this turn."
// The choice is the target opponent's, made at resolution and not targeting;
// one creature is a forced choice, none leaves no creature spared. "Can't
// block" changes no characteristic, so a creature that player gets later this
// turn can't block either (CR 611.2c).
func S82OldB735GoblinWarCry(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	player, ok := FirstLegalPlayerTarget(ctx)
	if !ok {
		return nil
	}
	var candidates []uuid.UUID
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == player && c.IsCreature() {
			candidates = append(candidates, c.InstanceID)
		}
	}
	if len(candidates) <= 1 {
		spared := uuid.Nil
		if len(candidates) == 1 {
			spared = candidates[0]
		}
		g.OtherCreaturesOfPlayerCantBlockForEffect(ctx.Source(), player, spared, DurationUntilEndOfTurn(ctx), s82OldB735WarCryLabel)
		return nil
	}
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:    player,
		FromPlayer: player,
		Source:     ctx.Source(),
		Question:   "Goblin War Cry — choose a creature you control; your other creatures can't block this turn",
		Cards:      candidates,
		Min:        1,
		Max:        1,
		LiveZone:   game.ZoneBattlefield,
		ThenKey:    s82OldB735WarCryPicked,
		Asked:      game.NewAskedItem(item),
	})
	return nil
}

const s82OldB735WarCryLabel = "Goblin War Cry — other creatures that player controls can't block this turn"

// s82OldB735WarCryPicked spares the creature the opponent chose and keeps
// every other creature they control from blocking. An on-disk key: never
// renamed, never reused.
var s82OldB735WarCryPicked = AskedPick("choose-cards/s82-old-b735-goblin-war-cry", func(ctx *Context, r game.CardsChosen) error {
	if len(r.Picked) != 1 {
		return nil
	}
	c, ok := ctx.Game.LookupCardForEffect(r.Picked[0])
	if !ok {
		return nil
	}
	ctx.Game.OtherCreaturesOfPlayerCantBlockForEffect(ctx.Source(), c.Controller, r.Picked[0], DurationUntilEndOfTurn(ctx), s82OldB735WarCryLabel)
	return nil
})

// S82OldB735TapAnyNumberForLife is Harmony of Nature's body: "Tap any number
// of untapped creatures you control. You gain 4 life for each creature tapped
// this way."
func S82OldB735TapAnyNumberForLife(g *game.Game, item *game.StackItem) error {
	candidates := UntappedCreaturesControlledBy(g, item.Controller)
	if len(candidates) == 0 {
		return nil
	}
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:  item.Controller,
		Source:   item.SourceCardID,
		Question: "Harmony of Nature — tap any number of untapped creatures you control",
		Cards:    candidates,
		Min:      0,
		// Max 0 means "all of them", which is what "any number" is.
		Max:      0,
		LiveZone: game.ZoneBattlefield,
		ThenKey:  s82OldB735HarmonyTapped,
		Asked:    game.NewAskedItem(item),
	})
	return nil
}

// s82OldB735HarmonyTapped taps the picks and gains 4 life for each that
// really became tapped. An on-disk key: never renamed, never reused.
var s82OldB735HarmonyTapped = AskedPick("choose-cards/s82-old-b735-harmony-of-nature", func(ctx *Context, r game.CardsChosen) error {
	n := 0
	for _, id := range r.Picked {
		c, ok := ctx.Game.LookupCardForEffect(id)
		if !ok || c.Tapped || !c.IsCreature() || c.Controller != ctx.Controller() {
			continue
		}
		if err := (TapTarget{Target: id}).Apply(ctx.AsGroupMember()); err != nil {
			return err
		}
		n++
	}
	if n == 0 {
		return nil
	}
	return GainLife{Amount: 4 * n}.Apply(ctx)
})

// S82OldB735ReturnNightstalkers is Return of the Nightstalkers' body: "Return
// all Nightstalker permanent cards from your graveyard to the battlefield.
// Then destroy all Swamps you control." The destruction runs once the returns
// (and any entry question) are done, so a returned Swamp is destroyed too.
func S82OldB735ReturnNightstalkers(_ *game.StackItem, ctx *Context) error {
	return ReturnFromGraveyardTogether{
		Targets: GraveyardCardIDs(ctx, ctx.Controller(), func(c game.Card) bool {
			return c.IsPermanent() && c.HasSubtype("Nightstalker")
		}),
		ThenTail: s82OldB735NightstalkersThen,
	}.Apply(ctx)
}

var s82OldB735NightstalkersThen = TogetherTail("together/s82-old-b735-return-of-the-nightstalkers", func(ctx *Context, _ []uuid.UUID, _ int) error {
	return DestroyAllMatching{
		Match: And(Land(), Subtype("Swamp"), ControlledBy(ctx.Controller())),
	}.Apply(ctx)
})

// S82OldB735TapEachTarget is "tap" over the spell's targets (Tidal Surge): each
// target that is still legal is tapped, and one that left is skipped
// (CR 608.2b).
func S82OldB735TapEachTarget(_ *game.StackItem, ctx *Context) error {
	for _, t := range ctx.LegalTargets() {
		if t.Kind == game.TargetCard {
			if err := (TapTarget{Target: t.ID}).Apply(ctx); err != nil {
				return err
			}
		}
	}
	return nil
}
