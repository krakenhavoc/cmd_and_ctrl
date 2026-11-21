package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Carnivorous Cultivator // Enroot — Creature — Elf Warlock {1}{G}, 2/3
// // Sorcery {G} (preparation card, CR 722):
//
//	"Deathtouch
//	 This creature enters prepared.
//	 Whenever this creature deals combat damage to a player, return
//	 target land card from your graveyard to your hand."
//
//	Enroot — "Search your library for a land card, put it into your
//	 graveyard, then shuffle."
//
// The trigger is targeted, so it is removed without a prompt when your
// graveyard holds no land card (CR 603.3d). Enroot's search has no
// "may" and always shuffles, even when no land is found.
//
// No simplification.
func init() {
	const id = "db655c5a-06f6-42bc-8180-d9ca1c1d939a"
	const label = "Carnivorous Cultivator — return target land card from your graveyard to your hand"
	effects.Register(effects.Spec{
		OracleID:        id,
		Name:            "Carnivorous Cultivator",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"deathtouch"},
		Replacements:    []game.ReplacementEffect{effects.SelfEntersPrepared()},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventDealDamage},
			Key:     label,
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return ev.Source == source.InstanceID && effects.CombatDamageToPlayerBy(ev, source.Controller, g)
			},
			Targets: effects.TargetCardInGraveyard("target land card from your graveyard", effects.Land(), effects.YouOwn()),
			Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				return game.NewTriggeredItem(source, label)
			},
			Effect: effects.ReturnTargetGraveyardCardToHand,
		}},
	})
	effects.Register(effects.Spec{
		OracleID:     id + "#1",
		Name:         "Enroot",
		Completeness: effects.CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.SearchLibrary{
				Player:    item.Controller,
				Predicate: game.Card.IsLand,
				Dest:      game.ZoneGraveyard,
				Limit:     1,
				Shuffle:   true,
				Reason:    "Enroot — put a land card into your graveyard",
			}.Apply(ctx)
		},
	})
}
