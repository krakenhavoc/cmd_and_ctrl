package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Spawnbed Protector — Creature — Eldrazi {7}, 6/8 (EDHREC rank
// 3041):
//
//	"At the beginning of your end step, return up to one target
//	 Eldrazi creature card from your graveyard to your hand. Create
//	 two 1/1 colorless Eldrazi Scion creature tokens with "Sacrifice
//	 this token: Add {C}.""
//
// The Eldrazi recursion engine. One printed ability, declared as TWO
// TriggeredAbility entries with one label — the Hazel's Brewmaster /
// Witch of the Moors split: the engine drops a targeted trigger
// outright when its legal set is empty (CR 603.3d), which is wrong
// for "up to one … AND create two Scions" — with no Eldrazi creature
// card in the graveyard the Scions must still be made. So the
// targeted entry fires only while the controller's graveyard holds
// an Eldrazi creature card and the untargeted entry only when it
// does not; exactly one applies to any end step. A chosen card that
// leaves the graveyard in response counters the ability by game
// rules, Scions included, as printed. The Scions are
// b28EldraziScionToken — 1/1 colorless with the Spawn's sacrifice
// mana ability.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "256a7f54-0e8e-4d22-a0f7-ff2830e8884e",
		Name:         "Spawnbed Protector",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			{
				Watches: []game.EventKind{game.EventBeginEndStep},
				AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					return ev.Actor == source.Controller && effects.B28GraveyardHasCreatureCardOfSubtype(g, source.Controller, source.TextSubtype("Eldrazi"))
				},
				Targets: effects.TargetCardInGraveyard("up to one target Eldrazi creature card from your graveyard", effects.YouOwn(), effects.Creature(), effects.HasSubtype("Eldrazi")).WithCount(0, 1),
				Key:     b28SpawnbedProtectorLabel,
				Effect:  effects.B28ReturnChosenGraveyardCardToHandThenScions,
			},
			effects.On(game.EventBeginEndStep, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return ev.Actor == source.Controller && !effects.B28GraveyardHasCreatureCardOfSubtype(g, source.Controller, source.TextSubtype("Eldrazi"))
			}, b28SpawnbedProtectorLabel, effects.B28ReturnChosenGraveyardCardToHandThenScions),
		},
	})
}

// b28SpawnbedProtectorLabel is the stack label both declarations
// share — one printed ability, one name on the stack.
const b28SpawnbedProtectorLabel = "Spawnbed Protector — return up to one Eldrazi creature card to your hand, create two Eldrazi Scions"
