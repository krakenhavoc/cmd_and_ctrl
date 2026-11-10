package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Valkyrie's Call — Enchantment {3}{W}{W} (Foundations, #4465):
//
//	"Whenever a nontoken, non-Angel creature you control dies, return that
//	 card to the battlefield under its owner's control with a +1/+1
//	 counter on it. It has flying and is an Angel in addition to its other
//	 types."
//
// Thunderbolts Conspiracy's return-and-retype shape. "You control" and
// "non-Angel" are read off the creature as it last existed on the
// battlefield (CR 603.10a), the type word as the card's own text spells it
// (CR 612.2); a token is refused. The +1/+1 counter is an entry counter
// (CR 614.1c), and the Angel type and flying are layer-4 and layer-6
// effects pinned to the new object (CR 611.2a, 400.7).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "27400927-9aba-4fb3-90c4-490d01969e60",
		Name:         "Valkyrie's Call",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventLTB, stdB15NontokenNonAngelYouControlDied,
				"Valkyrie's Call — return it with a +1/+1 counter, with flying and as an Angel too",
				func(g *game.Game, item *game.StackItem) error {
					return stdB15ReturnDiedWithCounters(g, item, map[string]int{game.CounterPlusOne: 1}, stdB15ValkyrieAngel)
				}),
		},
	})
}

func stdB15NontokenNonAngelYouControlDied(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
	if !ACreatureYouControlDied(ev, source, lki, g) {
		return false
	}
	dead, ok := g.LookupCardForEffect(ev.CardID)
	return ok && !dead.IsToken() && !leftAsSubtype(ev, dead, source.TextSubtype("Angel"))
}

// stdB15ValkyrieAngel gives the returned creature flying and the Angel
// type, for as long as it stays on the battlefield. An on-disk key.
var stdB15ValkyrieAngel = EnteredTail("valkyries-call/angel", func(ctx *Context, entered uuid.UUID, _ game.TailOutcome) error {
	g := ctx.Game
	if entered == uuid.Nil || !onBattlefield(g, entered) {
		return nil
	}
	return ScopedEffectFor{
		Target:   entered,
		Mods:     []game.Mod{game.AddSubtypesMod(ctx.TextSubtypes("Angel")...), game.AddKeywordsMod("flying")},
		Duration: g.PinnedTo(game.IndefiniteDuration(), entered),
		Label:    "Valkyrie's Call — it has flying and is an Angel in addition to its other types",
	}.Apply(ctx)
})
