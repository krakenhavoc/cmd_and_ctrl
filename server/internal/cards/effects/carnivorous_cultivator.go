package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:        id,
		Name:            "Carnivorous Cultivator",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"deathtouch"},
		Replacements:    []game.ReplacementEffect{SelfEntersPrepared()},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventDealDamage},
			Key:     label,
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return ev.Source == source.InstanceID && combatDamageToPlayerBy(ev, source.Controller, g)
			},
			Targets: TargetCardInGraveyard("target land card from your graveyard", Land(), YouOwn()),
			Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				return game.NewTriggeredItem(source, label)
			},
			Effect: returnTargetGraveyardCardToHand,
		}},
	})
	Register(Spec{
		OracleID:     id + "#1",
		Name:         "Enroot",
		Completeness: CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return SearchLibrary{
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
