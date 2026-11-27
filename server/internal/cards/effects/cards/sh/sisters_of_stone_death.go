package cardssh

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Sisters of Stone Death — Legendary Creature — Gorgon {4}{B}{B}{G}{G} (Ravnica:
// City of Guilds, S82 OLD batch 581, #5600):
//
//	"{G}: Target creature blocks Sisters of Stone Death this turn if able.
//	 {B}{G}: Exile target creature blocking or blocked by Sisters of Stone
//	 Death.
//	 {2}{B}: Put a creature card exiled with Sisters of Stone Death onto the
//	 battlefield under your control."
//
// The first is Vortex Elemental's requirement (CR 509.1c), naming the Sisters
// as the object they are now. The second targets a creature in combat with the
// Sisters, either way round (judged at announce and again at resolution, CR
// 608.2b: one removed from combat in response is illegal) and exiles it linked
// to the Sisters. The third picks among the creature cards exiled with this
// object (a card exiled by an earlier object of the same name is not linked,
// CR 400.7) as it resolves, Tomb of the Dusk Rose's choice, and puts it onto
// the battlefield under the activator's control.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "38e64b15-47c0-4453-a51e-a53c1aebc8d0",
		Name:         "Sisters of Stone Death",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{G}: Target creature blocks Sisters of Stone Death this turn if able.",
				Cost:    effects.ManaCost("{G}"),
				Targets: effects.TargetCreature("target creature"),
				Effect:  effects.S82OldB378TargetBlocksThis,
			},
			{
				Label:   "{B}{G}: Exile target creature blocking or blocked by Sisters of Stone Death.",
				Cost:    effects.ManaCost("{B}{G}"),
				Targets: effects.BlockingOrBlockedBySource(effects.TargetCreature("target creature blocking or blocked by Sisters of Stone Death")),
				Purpose: game.Purpose{Answers: game.AnswerRemove},
				Effect:  effects.ExileLinkedTargetsEffect,
			},
			{
				Label:   "{2}{B}: Put a creature card exiled with Sisters of Stone Death onto the battlefield under your control.",
				Cost:    effects.ManaCost("{2}{B}"),
				Purpose: game.Purpose{Answers: game.AnswerValue},
				Effect:  sistersOfStoneDeathReturn,
			},
		},
	})
}

func sistersOfStoneDeathReturn(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	ref, ok := ctx.SourceRef()
	if !ok {
		return nil
	}
	var cands []uuid.UUID
	for _, c := range g.ExiledWithForEffect(game.PermissionCardRef(ref)) {
		if c.IsCreature() {
			cands = append(cands, c.InstanceID)
		}
	}
	if len(cands) == 0 {
		return nil
	}
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:  ctx.Controller(),
		Source:   ctx.Source(),
		Question: "Sisters of Stone Death — put a creature card exiled with it onto the battlefield under your control",
		Cards:    cands,
		Min:      1,
		Max:      1,
		ThenKey:  sistersOfStoneDeathPicked,
		LiveZone: game.ZoneExile,
		Asked:    game.NewAskedItem(item),
	})
	return nil
}

// sistersOfStoneDeathPicked puts the chosen card onto the battlefield under the
// controller's control. An on-disk key: never renamed, never reused.
var sistersOfStoneDeathPicked = effects.ItemChoice("choose-cards/s82-old-b581-sisters-of-stone-death-return", func(ctx *effects.Context, r game.CardsChosen) error {
	if len(r.Picked) == 0 {
		return nil
	}
	return effects.ReturnFromExile{Target: r.Picked[0], Controller: ctx.Controller()}.Apply(ctx)
})
