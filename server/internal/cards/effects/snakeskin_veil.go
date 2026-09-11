package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Snakeskin Veil — Instant for {G}:
//
//	"Put a +1/+1 counter on target creature you control. It gains
//	 hexproof until end of turn."
//
// The protection instant whose pump is PERMANENT. That is the whole
// difference from Ranger's Guile and it is a real one in a voltron
// deck: three Veils over three turns leave a creature three points
// bigger forever, where three Guiles leave it exactly as it started.
//
// # Two different durations in two different places
//
// The counter goes on the card (`Card.Counters`) and is read by
// `CurrentPower` / `CurrentToughness` as a layer 7d modification for
// as long as the permanent is on the battlefield. The hexproof goes
// in the turn-scoped registry and is swept at cleanup. Nothing
// coordinates them, which is correct — they are separate effects
// that happen to be printed on one card.
//
// A counter is also why this card survives a Wrath better than its
// cousins: the creature dies, but if it comes back the +1/+1 is
// gone too (CR 400.7, new object), so the advantage is real only
// while the creature lives. The engine gets that for free by storing
// counters on the instance.
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID: "1e6a24be-8281-41c1-a5ba-b68f0ef1d7b8",
		Name:     "Snakeskin Veil",
		Targets:  TargetCreature("target creature you control", YouControl()),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
				return nil
			}
			target := item.Targets[0].ID
			if err := (AddCounter{
				Target: target,
				Kind:   "+1/+1",
				N:      1,
			}).Apply(ctx); err != nil {
				return err
			}
			return GrantKeywordUntilEOT{
				Target:   target,
				Keywords: []string{"hexproof"},
				Label:    "Snakeskin Veil — hexproof",
			}.Apply(ctx)
		},
	})
}
