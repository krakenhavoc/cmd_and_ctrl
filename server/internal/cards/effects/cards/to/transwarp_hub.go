package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Transwarp Hub — {2}{U} Artifact — Spacecraft (Star Trek Commander,
// #4974):
//
//	"When this Spacecraft enters, mill five cards. You may put up to two
//	 artifact cards from among them into your hand.
//	 Station (Tap another creature you control: Put charge counters
//	 equal to its power on this Spacecraft. Station only as a sorcery.)
//	 3+ | {T}: Target creature can't be blocked this turn."
//
// The enter trigger is Silvan Rally's mill-then-take: only the cards
// this mill put into the graveyard are offered, and only those still
// there. It prints no P/T box, so no CR 721.2b line ever makes it a
// creature; the 3+ line is a gated activated ability (CR 721.2a),
// absent below three charge counters. Its {T} is not a creature's, so
// it can be activated the turn the Hub enters.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "4325d4d2-5892-4c20-a994-7396b715e94f",
		Name:         "Transwarp Hub",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Transwarp Hub — mill five cards; you may put up to two artifact cards from among them into your hand",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					return effects.MillToZone{Player: ctx.Controller(), N: 5, ThenTail: ctx.ItemTail(transwarpHubMilled)}.Apply(ctx)
				}),
		},
		Activated: []effects.ActivatedAbility{
			effects.Station(),
			{
				Label:      "3+ | {T}: Target creature can't be blocked this turn.",
				Purpose:    effects.ForTargets(effects.CantBeBlockedUntilEOT(0, 0, 0)),
				Cost:       effects.TapCost(),
				Targets:    effects.TargetCreature("target creature"),
				ActiveWhen: effects.AtChargeCounters(3),
				Effect:     effects.TargetCantBeBlockedThisTurn("Transwarp Hub"),
			},
		},
	})
}

// transwarpHubMilled is the clause after the mill: up to two artifact
// cards from among the milled cards. An on-disk key.
var transwarpHubMilled = effects.CardTail("transwarp-hub/after-mill", func(ctx *effects.Context, o game.TailOutcome) error {
	return effects.HobB9TakeFromMilled(ctx, o.IDs, effects.Artifact(),
		"Transwarp Hub — put up to two artifact cards from among them into your hand", 0, 2)
})
