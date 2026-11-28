package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b778_helpers.go — S82 OLD batch 778 (Weatherlight, #5629).

// S82OldB778CreaturesCantBeTargetedBySpells is Dense Foliage's static,
// "Creatures can't be the targets of spells": every creature on the
// battlefield, whoever controls it, carries a spell-only targeting bar (the
// bar of #6196, read by the one targeting gate before hexproof). Abilities
// may still target.
func S82OldB778CreaturesCantBeTargetedBySpells(label string) game.StaticAbility {
	bar := game.SourceTargetingBar{
		Label:   label,
		Quality: game.ProtectionQuality{Kind: game.ProtectionQualityEverything, Printed: "everything"},
		Spells:  true,
	}
	return game.StaticAbility{
		Label:     label,
		Layer:     game.Layer6Ability,
		AppliesTo: func(target *game.Card, _ *game.Game, _ *game.Card) bool { return target.IsCreature() },
		Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
			c.TargetingBars = append(c.TargetingBars, bar)
		},
	}
}

// s82OldB778PhaseOutListedBody is the delayed trigger of "it phases out at end
// of combat": the listed creatures that are still on the battlefield phase out
// together. A registered key (ADR 0041 phase 3).
var s82OldB778PhaseOutListedBody = game.SimpleDelayedBody("s82-old-b778/phase-out-listed", func(g *game.Game, item *game.StackItem) error {
	var ids []uuid.UUID
	for _, t := range item.Targets {
		if t.Kind == game.TargetCard && OnBattlefield(g, t.ID) {
			ids = append(ids, t.ID)
		}
	}
	return PhaseOut{Targets: ids}.Apply(NewContext(g, item))
})

// S82OldB778PhaseOutTheAttackerAtEndOfCombat is Teferi's Veil's body: the
// attacking creature the event names phases out at the beginning of this
// combat's end of combat step (CR 511.2).
func S82OldB778PhaseOutTheAttackerAtEndOfCombat(label string) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		if item.Trigger == nil || item.Trigger.Event.CardID == uuid.Nil {
			return nil
		}
		return ScheduleDelayedTrigger{
			At:    game.StepEndCombat,
			Label: label,
			Cards: []uuid.UUID{item.Trigger.Event.CardID},
			Body:  s82OldB778PhaseOutListedBody,
		}.Apply(NewContext(g, item))
	}
}

// S82OldB778ReturnTheDiedCardTail is Angelic Renewal's "if you do, return that
// card to the battlefield": once the enchantment really was sacrificed, the card
// the trigger names comes back from the graveyard it was put into, if it is still
// there as that card (CR 400.7). A registered key (ADR 0041 phase 3).
var S82OldB778ReturnTheDiedCardTail = CardTail("s82-old-b778/return-the-died-card", func(ctx *Context, o game.TailOutcome) error {
	if !OneWentThisWay(o) {
		return nil
	}
	id := ctx.Trigger().Event.CardID
	if id == uuid.Nil || !DiedCardStillInGraveyard(ctx, id) {
		return nil
	}
	return ReturnFromGraveyard{Target: id, Dest: game.ZoneBattlefield}.Apply(ctx)
})
