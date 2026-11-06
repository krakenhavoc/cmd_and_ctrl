package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "4325d4d2-5892-4c20-a994-7396b715e94f",
		Name:         "Transwarp Hub",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Transwarp Hub — mill five cards; you may put up to two artifact cards from among them into your hand",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					return MillToZone{Player: ctx.Controller(), N: 5, ThenTail: ctx.ItemTail(transwarpHubMilled)}.Apply(ctx)
				}),
		},
		Activated: []ActivatedAbility{
			Station(),
			{
				Label:      "3+ | {T}: Target creature can't be blocked this turn.",
				Purpose:    ForTargets(CantBeBlockedUntilEOT(0, 0, 0)),
				Cost:       TapCost(),
				Targets:    TargetCreature("target creature"),
				ActiveWhen: AtChargeCounters(3),
				Effect:     targetCantBeBlockedThisTurn("Transwarp Hub"),
			},
		},
	})
}

// transwarpHubMilled is the clause after the mill: up to two artifact
// cards from among the milled cards. An on-disk key.
var transwarpHubMilled = CardTail("transwarp-hub/after-mill", func(ctx *Context, o game.TailOutcome) error {
	return hobB9TakeFromMilled(ctx, o.IDs, Artifact(),
		"Transwarp Hub — put up to two artifact cards from among them into your hand", 0, 2)
})
