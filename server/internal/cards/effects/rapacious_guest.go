package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Rapacious Guest — Creature — Halfling Citizen {2}{B}, 2/2 (EDHREC
// rank 3313):
//
//	"Menace
//	 Whenever one or more creatures you control deal combat damage to
//	 a player, create a Food token.
//	 Whenever you sacrifice a Food, put a +1/+1 counter on this
//	 creature.
//	 When this creature leaves the battlefield, target opponent loses
//	 life equal to its power."
//
// The Food deck's three-drop: connects for a Food, grows when the
// Food is eaten, and drains when it goes. Menace rides
// PrintedKeywords. Three triggers:
//
//   - "One or more creatures you control deal combat damage to a
//     player" is ONE trigger per PLAYER connected with (CR 603.2c,
//     #784). The engine emits one damage event per creature, so a
//     later event naming the same player is declined as a later
//     event of the same batch, keyed by label and damaged player —
//     by label because the other two abilities must not swallow it,
//     by player because a Guest whose creatures hit two opponents
//     makes two Foods. Without the batch half a three-creature alpha
//     strike on one player would make three Foods, the #259
//     direction.
//   - "You sacrifice a Food" is the sacrifice event by the
//     controller, the Food read while it is still on the battlefield
//     (b31YouSacrificedAFood).
//   - The leave trigger fires for any exit — a death, a bounce, an
//     exile — and targets an opponent, picked when it goes on the
//     stack. "Its power" is the power the Guest had as it left: the
//     CR 603.10 last-known characteristics the harvester hands the
//     trigger plus the +1/+1 counters read back off the log
//     (b13LastKnownPower — the counters are cleared on the way out
//     and the LKI carries no counter math), read in Build.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "54accd4a-b471-4ae5-b3b2-a5cec44023b7",
		Name:            "Rapacious Guest",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"menace"},
		Triggered: []game.TriggeredAbility{
			WheneverOneOrMoreCreaturesYouControlDealCombatDamageToAPlayer(nil,
				b31RapaciousGuestFoodLabel, Do(CreateToken{Template: FoodToken(), N: 1})),
			On(game.EventSacrifice, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return b31YouSacrificedAFood(ev, source, g)
			}, "Rapacious Guest — put a +1/+1 counter on it", func(g *game.Game, item *game.StackItem) error {
				if !onBattlefield(g, item.SourceCardID) {
					return nil
				}
				return AddCounter{Target: item.SourceCardID, Kind: game.CounterPlusOne, N: 1}.Apply(NewContext(g, item))
			}),
			{
				Watches: []game.EventKind{game.EventLTB},
				AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
					return ev.CardID == source.InstanceID
				},
				Targets: TargetPlayer("target opponent", Opponent()),
				Key:     "Rapacious Guest — target opponent loses life equal to its power",
				Build: func(_ game.Event, source *game.Card, lki game.Characteristic, g *game.Game) *game.StackItem {
					item := game.NewTriggeredItem(source, "Rapacious Guest — target opponent loses life equal to its power")
					item.Params.Amount = b13LastKnownPower(g, source.InstanceID, lki)
					return item
				},
				Effect: b31ChosenOpponentLosesLife,
			},
		},
	})
}

// b31RapaciousGuestFoodLabel is the stack label of the Guest's
// combat-damage trigger, which the "one or more" guard keys on
// alongside the damaged player.
const b31RapaciousGuestFoodLabel = "Rapacious Guest — create a Food"
