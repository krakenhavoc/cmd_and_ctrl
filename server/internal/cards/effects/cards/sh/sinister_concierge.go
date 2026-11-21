package cardssh

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Sinister Concierge — {1}{U} Creature — Human Wizard 2/1 (New Capenna
// Commander, #5879):
//
//	"When this creature dies, you may exile it and put three time counters
//	 on it. If you do, exile up to one target creature and put three time
//	 counters on it. Each card exiled this way that doesn't have suspend
//	 gains suspend. (For each card with suspend, its owner removes a time
//	 counter from it at the beginning of their upkeep. When the last is
//	 removed, they may cast it without paying its mana cost. Those creature
//	 spells have haste.)"
//
// The seam is on develop already (#4455, Kang Prime): an exiled card that
// is given suspend is marked Card.SuspendGranted, which puts the two
// exile-zone suspend triggers on it (game.GainSuspendForEffect, CR
// 702.62a). In exile a card's controller is its owner, so the target
// creature counts down on its OWNER's upkeep and its owner is offered the
// free cast with haste, as the reminder text says.
//
// The dies trigger asks "you may" as its own yes/no and announces up to one
// target creature as it goes on the stack (CR 603.3d). On resolution the
// Concierge is exiled from the graveyard first; only if it reached exile
// ("if you do") does it get its counters and suspend and is the target
// exiled with three counters and suspend. A Concierge that left the
// graveyard in response cannot be exiled, so nothing else happens
// (Cavalier of Thorns' shape). A trigger whose one chosen target is no
// longer legal does not resolve at all (CR 608.2b), so the Concierge stays
// in the graveyard; with no target chosen it still exiles itself.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "9b14c20f-c4ee-42ea-99cd-099b4eb25883",
		Name:         "Sinister Concierge",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			func() game.TriggeredAbility {
				t := effects.Targeting(
					effects.WhenThisDies("Sinister Concierge — exile it with three time counters, then exile up to one target creature with three time counters; each gains suspend",
						sinisterConciergeExileSelf),
					effects.UpToOneTargetCreature("up to one target creature"))
				t.OptionalPrompt = &game.TriggerOptionalPrompt{Question: "Sinister Concierge — exile it with three time counters (it gains suspend)?"}
				return t
			}(),
		},
	})
}

// sinisterConciergeExileSelf is "you may exile it": the Concierge, if it is
// still in a graveyard, is exiled, and the rest waits on that exile.
func sinisterConciergeExileSelf(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	z := g.FindCardZoneForEffect(item.SourceCardID)
	if z == nil || z.Kind != game.ZoneGraveyard {
		return nil
	}
	return effects.ExileTarget{
		Target:   item.SourceCardID,
		ThenTail: ctx.ItemTail(sinisterConciergeExiled),
	}.Apply(ctx)
}

// sinisterConciergeExiled is the clause after the Concierge's exile: only
// when it reached exile does it get three time counters (and suspend), and
// is the target creature exiled with three time counters (and suspend).
// An on-disk key.
var sinisterConciergeExiled = effects.CardTail("sinister-concierge/exiled", func(ctx *effects.Context, o game.TailOutcome) error {
	if len(o.IDs) != 1 {
		return nil
	}
	if err := ctx.Game.GainSuspendForEffect(ctx.Controller(), o.IDs[0], 3); err != nil {
		return err
	}
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetCard || t.ID == uuid.Nil {
			continue
		}
		return effects.ExileTarget{
			Target:   t.ID,
			ThenTail: game.KeyedTail{Ref: game.GainSuspendExileTail, Player: ctx.Controller(), N: 3},
		}.Apply(ctx)
	}
	return nil
})
