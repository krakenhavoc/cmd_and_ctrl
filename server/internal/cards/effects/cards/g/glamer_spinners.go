package cardsg

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Glamer Spinners — Creature — Faerie Wizard {4}{W/U}, 2/4 (Shadowmoor, #5591):
//
//	"Flash
//	 Flying
//	 When this creature enters, attach all Auras enchanting target
//	 permanent to another permanent with the same controller."
//
// Only the first permanent is targeted. As the ability resolves its
// controller chooses the other permanent, among the permanents the target's
// controller controls other than the target; an Aura that cannot legally
// enchant it (its enchant clause, protection) stays where it is
// (CR 701.3b). With no Aura on the target, or no other permanent, nothing
// happens. A target that has left makes the ability do nothing (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "b881abd7-1fe7-4404-b9ba-fb64869126fc",
		Name:            "Glamer Spinners",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash", "flying"},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisEnters("Glamer Spinners — attach all Auras enchanting target permanent to another permanent with the same controller",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					host, ok := effects.B16FirstLegalTargetCard(ctx)
					if !ok {
						return nil
					}
					hostCard, ok := g.LookupCardForEffect(host)
					if !ok {
						return nil
					}
					var auras []uuid.UUID
					for _, id := range g.AttachmentsOf(host) {
						if c, found := g.LookupCardForEffect(id); found && c.HasSubtype("Aura") {
							auras = append(auras, id)
						}
					}
					if len(auras) == 0 {
						return nil
					}
					var cands []uuid.UUID
					for _, c := range g.BattlefieldCardsForEffect() {
						if c.InstanceID != host && c.Controller == hostCard.Controller {
							cands = append(cands, c.InstanceID)
						}
					}
					if len(cands) == 0 {
						return nil
					}
					g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
						Chooser:  item.Controller,
						Source:   item.SourceCardID,
						Question: "Glamer Spinners — choose another permanent with the same controller to attach the Auras to",
						Cards:    cands,
						Min:      1,
						Max:      1,
						Zone:     game.ZoneBattlefield,
						ThenKey:  glamerSpinnersAttach,
						Carry:    auras,
					})
					return nil
				}), effects.TargetPermanent("target permanent")),
		},
	})
}

// glamerSpinnersAttach attaches each carried Aura to the chosen permanent,
// in order; one the new host cannot legally carry stays where it is. An
// on-disk key: never renamed, never reused.
var glamerSpinnersAttach = game.RegisterChooseCardsThen("choose-cards/glamer-spinners-attach", func(g *game.Game, r game.CardsChosen) error {
	if len(r.Picked) != 1 {
		return nil
	}
	for _, aura := range r.Carry {
		if err := g.AttachForEffect(aura, game.TargetRef{Kind: game.TargetCard, ID: r.Picked[0]}); err != nil {
			return err
		}
	}
	return nil
})
