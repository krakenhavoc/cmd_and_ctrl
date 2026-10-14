package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Extractor Demon — Creature — Demon {4}{B}{B}, 5/5:
//
//	"Flying
//	 Whenever another creature leaves the battlefield, you may have
//	 target player mill two cards.
//	 Unearth {2}{B}"
//
// The trigger watches every battlefield exit (EventLTB), not only
// deaths: a bounce, an exile, a sacrifice and a token dying all count,
// whoever controlled the creature. "Was a creature" is read off the
// exit's last-known types (leftAsType), so a creature card in a
// graveyard is not asked about. The target is chosen as the trigger
// goes on the stack and the "may" is asked first (CR 603.5). Unearth is
// the shared keyword constructor.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "1fb5c014-9366-4f9e-9728-303ff766002b",
		Name:            "Extractor Demon",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			Optional(Targeting(On(game.EventLTB, anotherCreatureLeftTheBattlefield,
				"Extractor Demon — target player mills two cards",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					for _, t := range ctx.LegalTargets() {
						if t.Kind == game.TargetPlayer {
							return MillCards{Player: t.ID, N: 2}.Apply(ctx)
						}
					}
					return nil
				}), TargetPlayer("target player")),
				"Extractor Demon — have target player mill two cards?"),
		},
		Activated: []ActivatedAbility{Unearth("{2}{B}")},
	})
}

// anotherCreatureLeftTheBattlefield is "whenever another creature leaves
// the battlefield": any exit, any controller, the source excluded.
func anotherCreatureLeftTheBattlefield(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.Kind != game.EventLTB || ev.CardID == source.InstanceID {
		return false
	}
	if was, known := ev.WasType("creature"); known {
		// A token that has already ceased to exist is still read
		// here, off the exit's own record.
		return was
	}
	c, ok := g.LookupCardForEffect(ev.CardID)
	return ok && leftAsType(ev, c, "creature")
}
