package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Oversold Cemetery — Enchantment {1}{B} (EDHREC rank 1698):
//
//	"At the beginning of your upkeep, if you have four or more
//	 creature cards in your graveyard, you may return target creature
//	 card from your graveyard to your hand."
//
// Two mana for a creature back every upkeep once the graveyard is
// stocked. An intervening-if (CR 603.4) on Emeria's shape: the
// creature-card count gates the trigger going on the stack AND is
// checked again at resolution, so four cards exiled in response
// stop the return — as printed, and closing the corner Emeria
// declares. "You may" is the optional-trigger prompt; the target is
// picked in the zone browser and re-checked at resolution. "Creature
// card" off the battlefield is the printed type line
// (b11CreatureCardsInGraveyard).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "ed4cd3a1-688b-4c05-948d-39d3336e00c0",
		Name:         "Oversold Cemetery",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventBeginUpkeep},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return ev.Actor == source.Controller && b11CreatureCardsInGraveyard(g, source.Controller) >= 4
			},
			OptionalPrompt: &game.TriggerOptionalPrompt{Question: "Oversold Cemetery — return a creature card from your graveyard to your hand?"},
			Targets:        TargetCardInGraveyard("target creature card in your graveyard", Creature(), YouOwn()),
			Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				return game.NewTriggeredItem(source, "Oversold Cemetery — return a creature card to your hand",
					func(g *game.Game, item *game.StackItem) error {
						if b11CreatureCardsInGraveyard(g, item.Controller) < 4 {
							return nil
						}
						if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
							return nil
						}
						return ReturnFromGraveyard{Target: item.Targets[0].ID, Dest: game.ZoneHand}.Apply(NewContext(g, item))
					})
			},
		}},
	})
}
