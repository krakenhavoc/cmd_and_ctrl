package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Mossborn Hydra — Creature — Elemental Hydra {2}{G}, 0/0 (EDHREC
// rank 768):
//
//	"Trample
//	 This creature enters with a +1/+1 counter on it.
//	 Landfall — Whenever a land you control enters, double the number
//	 of +1/+1 counters on this creature."
//
// Three mechanisms, all real:
//
//   - Trample rides PrintedKeywords.
//   - "Enters with a +1/+1 counter" is a self-replacement on its own
//     entry (ev.AddCounterAtETB), so the counter is there before any
//     ETB trigger or state check sees the 0/0 — the first catalog card
//     to use that hook.
//   - Landfall is the Tireless Provisioner trigger; doubling places N
//     more counters where N is the current count, through AddCounter,
//     so Doubling Season and Hardened Scales apply to the doubling as
//     printed.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "84122d81-9634-4d3f-85a5-f6cf4303691a",
		Name:            "Mossborn Hydra",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Replacements: []game.ReplacementEffect{{
			Watches: []game.EventKind{game.EventZoneMove},
			AppliesTo: func(ev *game.ReplacementEvent, _ *game.Game, src *game.Card) bool {
				return ev.Kind == game.RepEventMove && ev.NewZone == game.ZoneBattlefield &&
					src != nil && ev.CardID == src.InstanceID
			},
			Replace: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) error {
				ev.AddCounterAtETB("+1/+1", 1)
				return nil
			},
			Label: "Mossborn Hydra: enters with a +1/+1 counter",
		}},
		Triggered: []game.TriggeredAbility{
			Landfall("Mossborn Hydra — double its +1/+1 counters (landfall)", func(g *game.Game, item *game.StackItem) error {
				c, ok := g.LookupCardForEffect(item.SourceCardID)
				if !ok || c.Counters == nil {
					return nil
				}
				n := c.Counters["+1/+1"]
				if n <= 0 {
					return nil
				}
				return AddCounter{Target: item.SourceCardID, Kind: "+1/+1", N: n}.Apply(NewContext(g, item))
			}),
		},
	})
}
