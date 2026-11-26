package cardsco

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Crown of Doom — Artifact {3} (Commander 2014):
//
//	"Whenever a creature attacks you or a planeswalker you control, it gets
//	 +2/+0 until end of turn.
//	 {2}: Target player other than this artifact's owner gains control of
//	 it. Activate only during your turn."
//
// The pump is one trigger per attacking creature, aimed at that creature as
// it was declared. The hand-off targets any player but the OWNER
// (TargetSpec.NotSourceOwner), so the artifact can go to a third player or
// back to a non-owner but never home; it is indefinite (CR 611.2a).
func init() {
	hand := effects.TargetPlayer("target player other than this artifact's owner")
	hand.NotSourceOwner = true
	effects.Register(effects.Spec{
		OracleID:     "fed77e0a-a4a6-4409-9cab-bf0fc2cd9f9b",
		Name:         "Crown of Doom",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return ev.Kind == game.EventAttack && ev.Actor != uuid.Nil && ev.CardID != uuid.Nil &&
					effects.AttacksYouOrYourPlaneswalker(g, ev.Target, source.Controller)
			}, "Crown of Doom — the attacking creature gets +2/+0 until end of turn",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					attacker := ctx.Trigger().Event.CardID
					if !effects.OnBattlefield(g, attacker) {
						return nil
					}
					return effects.BoostUntilEOT{Target: attacker, Power: 2, Label: "Crown of Doom — +2/+0"}.Apply(ctx)
				}),
		},
		Activated: []effects.ActivatedAbility{{
			Label:     "{2}: Target player other than this artifact's owner gains control of it. Activate only during your turn.",
			Cost:      effects.ManaCost("{2}"),
			Condition: effects.DuringYourTurn(),
			Targets:   hand,
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				who, ok := effects.FirstLegalPlayerTarget(ctx)
				if !ok || effects.SourceIsNewObject(g, item) {
					return nil
				}
				g.GainControlForEffect(item.SourceCardID, item.SourceCardID, who, game.IndefiniteDuration(),
					"Crown of Doom — target player gains control of it")
				return nil
			},
		}},
	})
}
