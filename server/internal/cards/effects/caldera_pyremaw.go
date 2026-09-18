package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Caldera Pyremaw — Creature — Dragon {3}{R}{R}, 3/3 (EDHREC rank
// 4489):
//
//	"Flying
//	 Whenever you cast an instant or sorcery spell, put a +1/+1
//	 counter on this creature. Then this creature deals damage equal
//	 to its power to target opponent."
//
// A five-mana Dragon that turns a spellslinger deck into a burn deck.
// It grows on every cantrip and throws its new power at a player —
// so the third Opt of the turn is six damage, and the Pyremaw does
// not have to attack or even survive combat for any of it.
//
// Two things about the ordering, both printed and both implemented:
//
//   - The counter goes on FIRST, then the damage is measured. A
//     trigger from a spell cast while the Pyremaw is a 3/3 deals 4,
//     not 3.
//   - The power is read AT RESOLUTION, through CurrentPower while the
//     Pyremaw is still on the battlefield, so an anthem or a pump
//     that resolved in between counts. A Pyremaw killed in response
//     still deals its damage, off LAST KNOWN INFORMATION
//     (CR 608.2h) — the trigger is independent of its source once it
//     is on the stack (CR 603.10), and killing the Dragon is not a
//     way to fog the burn.
//
// The trigger fires on CAST, not on resolution, so it happens with
// the spell still on the stack — a countered Fireball has already
// paid for the Pyremaw's counter and damage, which is what "whenever
// you cast" means.
//
// "Target opponent" is a real target picked as the trigger goes on
// the stack; an opponent who left in response is skipped.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "3eb5ae68-7f76-4874-80a4-80058b10fc7a",
		Name:            "Caldera Pyremaw",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventCast},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				if ev.Actor != source.Controller {
					return false
				}
				c, ok := g.LookupCardForEffect(ev.CardID)
				return ok && (c.IsInstant() || c.IsSorcery())
			},
			Targets: TargetPlayer("target opponent", Opponent()),
			Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				return game.NewTriggeredItem(source, "Caldera Pyremaw — a +1/+1 counter, then damage equal to its power to target opponent",
					func(g *game.Game, item *game.StackItem) error {
						ctx := NewContext(g, item)
						if b09SourceStillOnBattlefield(g, item) {
							if err := (AddCounter{Target: item.SourceCardID, Kind: game.CounterPlusOne, N: 1}).Apply(ctx); err != nil {
								return err
							}
						}
						power := b43PowerNowOrLastKnown(g, item.SourceCardID)
						if power <= 0 {
							return nil
						}
						for _, t := range ctx.LegalTargets() {
							if t.Kind == game.TargetPlayer {
								return DealDamage{Source: item.SourceCardID, Target: t.ID, Amount: power}.Apply(ctx)
							}
						}
						return nil
					})
			},
		}},
	})
}
