package cardsu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Ulamog's Reclaimer — Creature — Eldrazi Processor {4}{U}, 2/5 (Battle for
// Zendikar, #5547):
//
//	"Devoid (This card has no color.)
//	 When this creature enters, you may put a card an opponent owns from exile
//	 into that player's graveyard. If you do, return target instant or sorcery
//	 card from your graveyard to your hand."
//
// The target is chosen as the trigger goes on the stack (CR 603.3d); with no
// instant or sorcery card in the graveyard the trigger is never put on the
// stack. It is re-checked as the return happens (CR 608.2b), after the
// processed card has landed, so a card that left the graveyard is not
// returned.
//
// No simplification.
func init() {
	const name = "Ulamog's Reclaimer"
	effects.Register(effects.Spec{
		OracleID:        "137de343-e3d0-4ccb-b8fc-680024387367",
		Name:            name,
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{game.KeywordDevoid},
		Triggered: []game.TriggeredAbility{
			effects.OldB305WhenThisEntersMayProcess(name, "ulamogs-reclaimer",
				"return target instant or sorcery card from your graveyard to your hand",
				effects.TargetCardInGraveyard("target instant or sorcery card from your graveyard",
					effects.Or(effects.Instant(), effects.Sorcery()), effects.YouOwn()),
				func(ctx *effects.Context) error {
					for _, t := range ctx.LegalTargets() {
						return effects.ReturnFromGraveyard{Target: t.ID, Dest: game.ZoneHand}.Apply(ctx)
					}
					return nil
				}),
		},
	})
}
