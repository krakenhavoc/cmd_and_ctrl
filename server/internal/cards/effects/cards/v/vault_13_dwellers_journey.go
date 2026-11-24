package cardsv

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Vault 13: Dweller's Journey — Enchantment — Saga {3}{W} (Fallout, #6100):
//
//	"(As this Saga enters and after your draw step, add a lore counter.
//	 Sacrifice after III.)
//	 I — For each player, exile up to one other target enchantment or
//	 creature that player controls until this Saga leaves the battlefield.
//	 II — You gain 2 life and scry 2.
//	 III — Return two cards exiled with this Saga to the battlefield under
//	 their owners' control and put the rest on the bottom of their owners'
//	 libraries."
//
// Chapter I is Battle at the Helvault's per-player clause list: one "up to
// one" clause per seat, the first for you and the rest for the opponents in
// turn order, each bound to its own player so the CR 608.2b re-check asks
// "does THAT player still control it?". "Other" keeps the Saga itself out.
// Each pick is exiled until the Saga leaves (CR 610.3) and linked to it
// (CR 607.2a), so chapter III can find "cards exiled with this Saga".
//
// Chapter III asks you for two of the cards still exiled with this Saga
// (all of them when there are two or fewer), returns those together under
// their owners' control, and then puts every card still exiled with it on
// the bottom of its owner's library, each owner ordering their own
// (CR 401.4). The Saga is sacrificed after chapter III (CR 714.4), and by
// then nothing is left in exile for the "until" return to bring back. If
// the Saga leaves before chapter III, everything chapter I exiled returns.
//
// No simplification.
const vault13Name = "Vault 13: Dweller's Journey"

func init() {
	const exileText = "for each player, exile up to one other target enchantment or creature that player controls until this Saga leaves the battlefield"
	effects.Register(effects.Spec{
		OracleID:     "586fb987-cd77-420f-978c-2ce140bcdcfb",
		Name:         vault13Name,
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Scry: 2},
		Triggered: []game.TriggeredAbility{
			effects.ChapterTriggerTargeting(1, effects.SagaChapterLabel(vault13Name, 1, exileText), vault13Clauses(), vault13Exile),
			effects.ChapterTrigger(2, effects.SagaChapterLabel(vault13Name, 2, "you gain 2 life and scry 2"),
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					if err := (effects.GainLife{Amount: 2}).Apply(ctx); err != nil {
						return err
					}
					return effects.Scry{N: 2}.Apply(ctx)
				}),
			effects.ChapterTrigger(3, effects.SagaChapterLabel(vault13Name, 3, "return two cards exiled with this Saga to the battlefield under their owners' control and put the rest on the bottom of their owners' libraries"),
				vault13Return),
		},
	})
}

// vault13Clauses is "for each player, up to one other target enchantment or
// creature that player controls".
func vault13Clauses() *game.TargetSpec {
	spec := effects.Gap110PerPlayerClauses("other enchantment or creature", effects.Or(effects.Enchantment(), effects.Creature()))
	for i := 0; i < spec.ClauseCount(); i++ {
		effects.Another(spec.Clause(i))
	}
	return spec
}

// vault13Exile exiles every still-legal pick, linked to this Saga, until it
// leaves the battlefield, in announce order.
func vault13Exile(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetCard {
			continue
		}
		if err := (effects.ExileUntil{
			Target:     t.ID,
			ThisLeaves: true,
			Linked:     true,
			Label:      vault13Name + " — returns when the Saga leaves the battlefield",
		}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
}

// vault13Exiled is the cards in exile now that were exiled with this Saga.
func vault13Exiled(ctx *effects.Context) []uuid.UUID {
	src, ok := ctx.SourceRef()
	if !ok {
		return nil
	}
	var ids []uuid.UUID
	for _, c := range ctx.Game.ExiledWithForEffect(game.PermissionCardRef(src)) {
		ids = append(ids, c.InstanceID)
	}
	return ids
}

// vault13Return is chapter III: two cards back (you choose which when there
// are more), then the rest to the bottom.
func vault13Return(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	ids := vault13Exiled(ctx)
	if len(ids) <= 2 {
		return vault13ReturnThese(ctx, ids)
	}
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:  item.Controller,
		Source:   ctx.Source(),
		Question: vault13Name + " — choose two cards exiled with this Saga to return to the battlefield",
		Cards:    ids,
		Min:      2,
		Max:      2,
		LiveZone: game.ZoneExile,
		ThenKey:  vault13Chosen,
		Asked:    game.NewAskedItem(item),
	})
	return nil
}

// vault13Chosen is the answer to chapter III's pick. A prompt that ended
// unanswered (its chooser left the game) returns the first two. An on-disk
// key: never renamed, never reused.
var vault13Chosen = effects.ItemChoice("choose-cards/vault-13-dwellers-journey-return-two", func(ctx *effects.Context, r game.CardsChosen) error {
	picked := r.Picked
	if !r.Answered && len(r.Candidates) >= 2 {
		picked = r.Candidates[:2]
	}
	return vault13ReturnThese(ctx, picked)
})

// vault13ReturnThese returns the cards together under their owners'
// control, then tucks the rest.
func vault13ReturnThese(ctx *effects.Context, ids []uuid.UUID) error {
	return effects.ReturnFromExileTogether{Targets: ids, ThenTail: vault13TuckRest}.Apply(ctx)
}

// vault13TuckRest is "put the rest on the bottom of their owners'
// libraries": every card still exiled with this Saga once the two have
// returned. An on-disk key: never renamed, never reused.
var vault13TuckRest = effects.CardTail("tail/vault-13-dwellers-journey-tuck-rest", func(ctx *effects.Context, _ game.TailOutcome) error {
	rest := vault13Exiled(ctx)
	if len(rest) == 0 {
		return nil
	}
	return effects.TuckCardsToBottomByOwner(ctx, ctx.Item, rest, vault13Name+" — put these on the bottom of your library in any order")
})
