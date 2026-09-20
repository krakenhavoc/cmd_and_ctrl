package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Righteous Valkyrie — Creature — Angel Cleric {2}{W}, 2/4 (EDHREC
// rank 2030):
//
//	"Flying
//	 Whenever another Angel or Cleric you control enters, you gain
//	 life equal to that creature's toughness.
//	 As long as you have at least 7 life more than your starting
//	 life total, creatures you control get +2/+2."
//
// The Angel deck's lifegain engine. The trigger reads effective
// subtypes, so a changeling counts, and "that creature's toughness"
// is read as the trigger resolves — current toughness, counters and
// anthems included — falling back to the toughness it had when it
// entered if it has since left (CR 608.2h's last-known value).
//
// The +2/+2 anthem was a declared simplification until #1117: it is
// a static whose AppliesTo reads the controller's LIFE TOTAL, and a
// life change was not one of the events that invalidated the layer
// engine's cached resolution (zone moves, counters, attachments, taps
// were) — so the bonus would have switched on and off a beat late,
// and "late off" is a creature hitting for +2 after its controller
// dropped below the line, the direction #259 rules out. It now
// declares the dependency through LifeGatedPump and is live.
//
// "Your starting life total" is the TABLE's (ADR 0075), so the line
// is 47 at a Commander table and 27 at a 20-life one.
func init() {
	Register(Spec{
		OracleID:        "891d2690-7144-4f87-b6ef-96f2469780a9",
		Name:            "Righteous Valkyrie",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Static: []game.StaticAbility{
			LifeGatedPump(YourCreaturesWhileYourLifeAboveStarting(7), 2, 2),
		},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventETB},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				_, ok := b18AnotherAngelOrClericYouControlEntered(ev, source, g)
				return ok
			},
			Build: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) *game.StackItem {
				entered := ev.CardID
				fallback := 0
				if c, ok := g.LookupCardForEffect(entered); ok {
					fallback = c.CurrentToughness()
				}
				return game.NewTriggeredItem(source, "Righteous Valkyrie — gain life equal to its toughness",
					func(g *game.Game, item *game.StackItem) error {
						toughness := fallback
						if z := g.FindCardZoneForEffect(entered); z != nil && z.Kind == game.ZoneBattlefield {
							if c, ok := g.LookupCardForEffect(entered); ok {
								toughness = c.CurrentToughness()
							}
						}
						return GainLife{Player: item.Controller, Amount: toughness}.Apply(NewContext(g, item))
					})
			},
		}},
	})
}
