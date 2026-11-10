package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Rotisserie Elemental — Creature {R}, 1/1 Elemental (Wilds of Eldraine,
// #4471):
//
//	"Menace
//	 Whenever this creature deals combat damage to a player, put a skewer
//	 counter on this creature. Then you may sacrifice it. If you do,
//	 exile the top X cards of your library, where X is the number of
//	 skewer counters on this creature. You may play those cards this
//	 turn."
//
// The counter goes on first and the sacrifice is asked as the trigger
// resolves, after the counter has landed (the Do walk waits for a prompt
// the counter's window asks). Ogre-Head Helm's shape: the exile hangs off
// the sacrifice actually happening, and X is the skewer counters the
// creature had as it last existed on the battlefield (CR 608.2h,
// trekB5CountersOnThis), so a creature that left or came back as a new
// object exiles nothing. The exile is Memorial Vault's impulse exile with
// the "play" permission (not cast-only), lapsing at end of turn, so a land
// can be played, subject to the usual land-drop and timing rules.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "e33c1744-843f-438f-a8c3-2d9f60120776",
		Name:            "Rotisserie Elemental",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"menace"},
		Triggered: []game.TriggeredAbility{
			WheneverThisDealsCombatDamageToAPlayer(
				"Rotisserie Elemental — put a skewer counter on this creature, then you may sacrifice it to exile cards and play them this turn",
				Do(CounterOnThis{Kind: "skewer", N: 1}, MayChoice{
					Question: "Rotisserie Elemental — sacrifice it to exile the top cards of your library?",
					YesLabel: "Sacrifice",
					NoLabel:  "Decline",
					Then:     rotisserieElementalAnswer,
				})),
		},
	})
}

// rotisserieElementalExile is "exile the top X cards of your library,
// where X is the number of skewer counters on this creature. You may play
// those cards this turn", run once the creature is sacrificed. X is read
// while the creature is still on the battlefield and rides the tail's N.
var rotisserieElementalExile = CardTail("then/rotisserie-elemental-exile", func(ctx *Context, o game.TailOutcome) error {
	if !oneWentThisWay(o) || o.Tail.N <= 0 {
		return nil
	}
	return ExileTopWithPermission{From: ctx.Controller(), GrantTo: ctx.Controller(), N: o.Tail.N}.Apply(ctx)
})

// rotisserieElementalAnswer is the "yes": sacrifice the creature, and if it
// went, the exile of as many cards as it had skewer counters. Keyed.
var rotisserieElementalAnswer = MayYes("confirm/rotisserie-elemental-sacrifice", func(ctx *Context, _ game.ConfirmAnswer) error {
	if !onBattlefield(ctx.Game, ctx.Source()) {
		return nil
	}
	x := trekB5CountersOnThis(ctx.Game, ctx.Source(), "skewer")
	return SacrificePermanent{Target: ctx.Source(), ThenTail: withN(ctx.ItemTail(rotisserieElementalExile), x)}.Apply(ctx)
})
