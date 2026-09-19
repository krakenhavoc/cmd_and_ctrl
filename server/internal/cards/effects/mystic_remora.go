package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Mystic Remora — Enchantment for {U}:
//
//	"Cumulative upkeep {1} (At the beginning of your upkeep, put an
//	 age counter on this permanent, then sacrifice it unless you pay
//	 its upkeep cost for each age counter on it.)
//	 Whenever an opponent casts a noncreature spell, you may draw a
//	 card unless that player pays {4}."
//
// #567. The card that #74 deferred in S22: its Rhystic half has been
// expressible since Rhystic Study (S19), and the age-counter cost was
// the entire remaining job.
//
// Both halves are triggers and both are ordinary. The first is
// CumulativeUpkeep (cumulative_upkeep.go), which is the whole keyword
// in one constructor; the second is Rhystic Study's clause with a
// noncreature filter and a bigger tax, and the two prompts are the
// same pay_unless kind asked of different players — which is exactly
// why one of them blocks the table and the other does not (ADR 0018
// §6; the upkeep half goes through effects.UpkeepPayUnless and the
// Rhystic half through PayUnless).
//
// Sandbox simplification, inherited from Rhystic Study: the
// controller's "you may draw" is treated as "draw". DrawCards no-ops
// on an empty library and the player loses at the next SBA per
// CR 704.5b, which is what drawing from an empty library does in
// paper — but a Remora controller who would rather not draw cannot
// say so.
func init() {
	Register(Spec{
		OracleID:     "8a52f3c0-2552-4425-b2e3-5496eb2232a7",
		Name:         "Mystic Remora",
		Completeness: CompletenessCaveats,
		Caveats: []string{
			"The draw is mandatory when the opponent declines to pay — you can't choose to skip it, which matters on an empty library.",
		},
		Triggered: []game.TriggeredAbility{
			CumulativeUpkeep("Mystic Remora — cumulative upkeep {1}", "{1}"),
			{
				Watches: []game.EventKind{game.EventCast},
				AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					if ev.Actor == source.Controller {
						return false
					}
					c, ok := g.LookupCardForEffect(ev.CardID)
					return ok && !c.IsCreature()
				},
				Build: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
					caster := ev.Actor
					return game.NewTriggeredItem(source, "Mystic Remora — draw unless caster pays {4}",
						func(g *game.Game, item *game.StackItem) error {
							return PayUnless{
								Chooser:  caster,
								Cost:     "{4}",
								Question: "Mystic Remora — pay {4}?",
								OnDecline: func(ctx *Context) error {
									return DrawCards{Player: ctx.Controller(), N: 1}.Apply(ctx)
								},
							}.Apply(NewContext(g, item))
						})
				},
			},
		},
	})
}
