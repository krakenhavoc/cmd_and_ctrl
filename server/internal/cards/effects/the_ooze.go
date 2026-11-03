package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// The Ooze — {2} Legendary Artifact (Teenage Mutant Ninja Turtles, #4457):
//
//	"Whenever a creature you control with a +1/+1 counter on it leaves the
//	 battlefield, create a Mutagen token for each +1/+1 counter on it. (A
//	 Mutagen token is an artifact with "{1}, {T}, Sacrifice this token: Put
//	 a +1/+1 counter on target creature. Activate only as a sorcery.")
//	 {T}: Exile target card from a graveyard. Create a Mutagen token."
//
// The leaves trigger is The Ozolith's: the departed creature's counters
// are its last-known information (CR 603.10a, LastKnownCountersForEffect),
// since the move to a new zone clears them (CR 400.7). The condition is
// read as it leaves, under the control it had then, and the count is read
// again from the same record as the trigger resolves. The tokens do not
// need The Ooze to still be there. The {T} exile is Agatha's Soul
// Cauldron's target, with the token after it; a target gone before
// resolution means the ability does not resolve (CR 608.2b), and no token.
// The Mutagen token is the catalog's CR 111.10v template (MutagenToken, #4676).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "82a6fa6f-13c2-4ea3-bec2-c51b56b2783b",
		Name:         "The Ooze",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventLTB},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				left, ok := g.LookupCardForEffect(ev.CardID)
				if !ok || !leftAsType(ev, left, "creature") || leftUnderControlOf(ev, left) != source.Controller {
					return false
				}
				return g.LastKnownCountersForEffect(ev.CardID)[game.CounterPlusOne] > 0
			},
			Key:    "The Ooze — create a Mutagen token for each +1/+1 counter on it",
			Effect: theOozeMutagenPerCounter,
		}},
		Activated: []ActivatedAbility{{
			Label:   "{T}: Exile target card from a graveyard. Create a Mutagen token.",
			Cost:    TapCost(),
			Targets: TargetCardInGraveyard("target card from a graveyard"),
			Effect: Do(
				Instruction(func(ctx *Context) error {
					for _, t := range ctx.LegalTargets() {
						if t.Kind == game.TargetCard {
							return ExileTarget{Target: t.ID}.Apply(ctx)
						}
					}
					return nil
				}),
				CreateToken{Template: MutagenToken(), N: 1},
			),
		}},
	})
}

// theOozeMutagenPerCounter is the leaves trigger's resolution: one Mutagen
// for each +1/+1 counter the creature had as it left.
func theOozeMutagenPerCounter(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	left, ok := ctx.TriggeringPermanent()
	if !ok {
		return nil
	}
	n := left.Counters[game.CounterPlusOne]
	if n <= 0 {
		return nil
	}
	return CreateToken{Template: MutagenToken(), N: n}.Apply(ctx)
}
