package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Liesa, Forgotten Archangel — Legendary Creature — Angel {2}{W}{W}{B},
// 4/5 (EDHREC rank 2104):
//
//	"Flying, lifelink
//	 Whenever another nontoken creature you control dies, return that
//	 card to its owner's hand at the beginning of the next end step.
//	 If a creature an opponent controls would die, exile it instead."
//
// Your creatures come back, theirs never do. Two keywords on
// PrintedKeywords and two abilities:
//
//   - The dies trigger is b15AnotherCreatureDied narrowed to nontoken
//     creatures the controller controls, and its body schedules The
//     Locust God's CR 603.7 delayed trigger for the next end step —
//     any player's, as printed — carrying the dead card's ID. When it
//     fires, the card comes back to hand only if it is still in a
//     graveyard: one reanimated or exiled in the meantime is a
//     different object and is left alone.
//   - The exile is a CR 614 replacement on the battlefield-to-
//     graveyard move of a creature an opponent controls (Cosmic
//     Intervention's shape, made permanent by living on Liesa). It
//     catches destruction, lethal damage, sacrifice and a 0-toughness
//     death alike, since every one of them is the same move; the
//     creature's own dies-triggers do not fire, because it never
//     died (CR 700.4).
//
// An opponent's dying COMMANDER is exiled too, and its owner
// separately decides. CR 903.9a (pinned edition) makes a commander's
// trip from a graveyard or exile a state-based "may" AFTER it
// arrives, so this replacement exiles the commander like any other of
// the opponent's creatures, and the built-in command-zone replacement
// (builtin_replacements.go) meets it in the CR 616 apply loop: the
// commander's OWNER orders the two (CR 616.1 gives that choice to the
// affected object's controller, not to Liesa's), the command-zone
// offer is asked, and "no" leaves this effect to exile the creature.
// Both printed outcomes are reachable and nothing else is — the same
// reasoning that closed Cosmic Intervention's identical stale caveat,
// pinned here by
// TestB19LiesaExilesAnOpponentsDyingCommanderWhenItsOwnerDeclines.
func init() {
	Register(Spec{
		OracleID:        "efcaadbe-24e3-4dfc-b08c-a910f003d427",
		Name:            "Liesa, Forgotten Archangel",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "lifelink"},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventLTB},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return b19AnotherNontokenCreatureYouControlDied(ev, source, g)
			},
			Build: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				dead := ev.CardID
				return game.NewTriggeredItem(source, "Liesa, Forgotten Archangel — return that card to hand at the next end step",
					func(g *game.Game, item *game.StackItem) error {
						return ScheduleDelayedTrigger{
							Label:  "Liesa, Forgotten Archangel — return the creature card to its owner's hand",
							Cards:  []uuid.UUID{dead},
							Effect: b15ReturnListedCardsFromGraveyardToHand,
						}.Apply(NewContext(g, item))
					})
			},
		}},
		Replacements: []game.ReplacementEffect{{
			Watches: []game.EventKind{game.EventZoneMove},
			AppliesTo: func(ev *game.ReplacementEvent, g *game.Game, src *game.Card) bool {
				if ev.Kind != game.RepEventMove || ev.OldZone != game.ZoneBattlefield || ev.NewZone != game.ZoneGraveyard {
					return false
				}
				c, ok := g.LookupCardForEffect(ev.CardID)
				return ok && c.IsCreature() && c.Controller != src.Controller
			},
			Replace: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) error {
				ev.NewZone = game.ZoneExile
				ev.NewZoneOwner = uuid.Nil
				return nil
			},
			Controller: func(_ *game.ReplacementEvent, _ *game.Game, src *game.Card) uuid.UUID {
				return src.Controller
			},
			Label: "Liesa, Forgotten Archangel: exile instead of dying",
		}},
	})
}
