package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Stone of Erech — Legendary Artifact {1} (EDHREC rank 3545):
//
//	"If a creature an opponent controls would die, exile it instead.
//	 {2}, {T}, Sacrifice Stone of Erech: Exile target player's
//	 graveyard. Draw a card."
//
// A one-mana Rest in Peace for opposing creatures with a Tormod's
// Crypt stapled on. The replacement is Liesa, Forgotten Archangel's:
// a CR 614 rewrite of the battlefield-to-graveyard move of a
// creature an opponent controls, so destruction, lethal damage,
// sacrifice and a 0-toughness death all exile instead and the
// creature's own dies-triggers never fire (CR 700.4). The activated
// ability is Stonespeaker Crystal's narrowed to one target player,
// with the draw after the exile in printed order.
//
// An opponent's dying COMMANDER is exiled too, and its owner
// separately decides — Liesa's reasoning, shared because the
// replacement is Liesa's. CR 903.9a (pinned edition) makes a
// commander's trip from a graveyard or exile a state-based "may"
// AFTER it arrives, so this replacement exiles the commander like any
// other of the opponent's creatures, and the built-in command-zone
// replacement (builtin_replacements.go) meets it in the CR 616 apply
// loop: the commander's OWNER orders the two (CR 616.1 gives that
// choice to the affected object's controller, not to the Stone's),
// the command-zone offer is asked, and "no" leaves this effect to
// exile the creature. Pinned by
// TestB33StoneOfErechExilesAnOpponentsDyingCommanderWhenItsOwnerDeclines.
func init() {
	Register(Spec{
		OracleID:     "73dad679-1edb-41c9-9d43-56dc93c3e9fe",
		Name:         "Stone of Erech",
		Completeness: CompletenessFull,
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
			Label: "Stone of Erech: exile instead of dying",
		}},
		Activated: []ActivatedAbility{{
			Label:   "{2}, {T}, Sacrifice Stone of Erech: Exile target player's graveyard. Draw a card.",
			Cost:    Plus(ManaCost("{2}"), TapCost(), SacrificeThis()),
			Targets: TargetPlayer("target player"),
			Effect:  b30ExileTargetGraveyardsThenDraw,
		}},
	})
}
