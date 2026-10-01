package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Lonis, Genetics Expert — Legendary Creature — Snake Elf Detective
// {1}{G/U}{G/U}, 1/2:
//
//	"Evolve (Whenever a creature you control enters, if that creature
//	 has greater power or toughness than this creature, put a +1/+1
//	 counter on this creature.)
//	 Whenever one or more +1/+1 counters are put on Lonis, investigate
//	 that many times.
//	 Whenever you sacrifice a Clue, put a +1/+1 counter on another
//	 target creature you control."
//
// Evolve is the engine's keyword trigger (game/evolve.go, #1805).
//
//   - "One or more … are put" is one trigger per placement event, and
//     "that many" is that placement's delta, read off the log as the
//     ability triggers (b33CountersPlacedDelta — Nest of Scarabs'
//     shape) and carried on the item. Investigate is a Clue token
//     (CR 701.16a).
//   - "Whenever you sacrifice a Clue" watches EventSacrifice, which is
//     emitted before the Clue leaves, so its subtypes are read off the
//     battlefield. "Another target creature you control" is a target
//     clause that excludes Lonis itself (Another).
//
// No simplification.
func init() {
	const investigateLabel = "Lonis, Genetics Expert — investigate that many times"
	Register(Spec{
		OracleID:        "52d8e492-b7b7-4ad4-8f1f-5c50993e55e0",
		Name:            "Lonis, Genetics Expert",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{game.KeywordEvolve},
		Triggered: []game.TriggeredAbility{
			{
				Watches: []game.EventKind{game.EventCounterPlaced},
				AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					return ev.Target == source.InstanceID && b33CountersPlacedDelta(ev, game.CounterPlusOne, g) > 0
				},
				Key: investigateLabel,
				Build: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) *game.StackItem {
					item := game.NewTriggeredItem(source, investigateLabel)
					item.Params.Amount = b33CountersPlacedDelta(ev, game.CounterPlusOne, g)
					return item
				},
				Effect: func(g *game.Game, item *game.StackItem) error {
					if item.Params.Amount <= 0 {
						return nil
					}
					return CreateToken{Controller: item.Controller, Template: ClueToken(), N: item.Params.Amount}.Apply(NewContext(g, item))
				},
			},
			Targeting(On(game.EventSacrifice, lonisYouSacrificedAClue,
				"Lonis, Genetics Expert — put a +1/+1 counter on another target creature you control",
				putPlusOneCounterOnEachLegalTarget),
				Another(TargetCreature("another target creature you control", YouControl()))),
		},
	})
}

// lonisYouSacrificedAClue is "you sacrifice a Clue": EventSacrifice
// fires before the zone move, so the Clue is still there to be read.
func lonisYouSacrificedAClue(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.Kind != game.EventSacrifice || ev.Actor != source.Controller || ev.CardID == uuid.Nil {
		return false
	}
	c, ok := g.LookupCardForEffect(ev.CardID)
	return ok && c.HasSubtype("Clue")
}
