package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Elrond, Lord of Rivendell — Legendary Creature — Elf Noble {2}{U},
// 3/2:
//
//	"Whenever Elrond or another creature you control enters, scry 1.
//	 If this is the second time this ability has resolved this turn,
//	 the Ring tempts you."
//
// The count is Victor, Valgavoth's Seneschal's: the per-turn tally of
// this Elrond's resolutions of this ability, which includes the one
// resolving (Game.ResolvedThisTurn). It is per object (CR 400.7), so
// an Elrond that left and came back starts again. The count is read
// before the scry, and the tempt waits for the scry to be answered.
// Only the second resolution tempts, not the third or later.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "099ae8f2-5c84-43ab-aaa1-ea69f2bba784",
		Name:         "Elrond, Lord of Rivendell",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventETB, elrondOrAnotherCreatureYouControlEntered, elrondLabel, elrondScry),
		},
	})
}

const elrondLabel = "Elrond, Lord of Rivendell — scry 1, then the second time this turn the Ring tempts you"

// elrondOrAnotherCreatureYouControlEntered is "Elrond or another
// creature you control enters".
func elrondOrAnotherCreatureYouControlEntered(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.CardID == source.InstanceID {
		return true
	}
	c, ok := g.LookupCardForEffect(ev.CardID)
	return ok && c.Controller == source.Controller && c.IsCreature()
}

// elrondScry is the ability's body.
func elrondScry(g *game.Game, item *game.StackItem) error {
	second := g.ResolvedThisTurn(item.SourceCardID, elrondLabel) == 2
	var then func(*game.Game) error
	if second {
		then = ringTemptsYouNext(item)
	}
	return Scry{N: 1, Then: then}.Apply(NewContext(g, item))
}
