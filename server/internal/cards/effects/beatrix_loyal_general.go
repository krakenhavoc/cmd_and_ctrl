package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Beatrix, Loyal General — Legendary Creature — Human Soldier
// {4}{W}{W}, 4/4 (Final Fantasy, #4462):
//
//	"Vigilance
//	 At the beginning of combat on your turn, you may attach any number
//	 of Equipment you control to target creature you control."
//
// Ardenn, Intrepid Archaeologist's trigger narrowed to Equipment and to a
// creature you control. The creature is targeted as the trigger goes on
// the stack; as it resolves, if it is still legal (CR 608.2b), the
// controller picks any number of their Equipment, none being the "you
// may", and each is attached to it. Only an Equipment that could legally
// be attached there is offered (CR 701.3b), and one already on it stays
// where it is.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "2f24637d-8d8e-49de-b229-4cdb97595ed7",
		Name:            "Beatrix, Loyal General",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Triggered: []game.TriggeredAbility{
			Targeting(
				AtBeginningOfYourCombat("Beatrix, Loyal General — you may attach any number of Equipment you control to target creature you control",
					beatrixAttach),
				TargetCreature("target creature you control", YouControl())),
		},
	})
}

// beatrixAttach offers the Equipment that could go on the target and
// attaches the ones picked.
func beatrixAttach(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	host, ok := b16FirstLegalTargetCard(ctx)
	if !ok {
		return nil
	}
	var candidates []uuid.UUID
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller != item.Controller || !c.HasSubtype("Equipment") || c.InstanceID == host {
			continue
		}
		if c.AttachedTo.Kind == game.TargetCard && c.AttachedTo.ID == host {
			continue // already there: attaching it again does nothing
		}
		if !g.AttachLegalForEffect(c.InstanceID, host) {
			continue
		}
		candidates = append(candidates, c.InstanceID)
	}
	if len(candidates) == 0 {
		return nil
	}
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:  item.Controller,
		Source:   item.SourceCardID,
		Question: "Beatrix, Loyal General — attach any number of Equipment you control to that creature",
		Cards:    candidates,
		Min:      0,
		Max:      len(candidates),
		ThenKey:  beatrixAttachPick,
		LiveZone: game.ZoneBattlefield,
		Carry:    []uuid.UUID{host},
	})
	return nil
}

// beatrixAttachPick attaches the picked Equipment to the creature (the
// carry). An on-disk key.
var beatrixAttachPick = game.RegisterChooseCardsThen("choose-cards/beatrix-loyal-general-attach", func(g *game.Game, r game.CardsChosen) error {
	if len(r.Carry) == 0 || !g.Battlefield.Contains(r.Carry[0]) {
		return nil
	}
	host := game.TargetRef{Kind: game.TargetCard, ID: r.Carry[0]}
	for _, id := range r.Picked {
		if !g.Battlefield.Contains(id) {
			continue
		}
		if err := g.AttachForEffect(id, host); err != nil {
			return err
		}
	}
	return nil
})
