package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Defiled Crypt // Cadaver Lab — Enchantment — Room (ADR 0103):
//
//	Defiled Crypt {3}{B}: "Whenever one or more cards leave your
//	 graveyard, create a 2/2 black Horror enchantment creature token.
//	 This ability triggers only once each turn."
//	Cadaver Lab {B}: "When you unlock this door, return target creature
//	 card from your graveyard to your hand."
//
// Defiled Crypt is Teval's graveyard-leave condition (a move out of the
// controller's own graveyard, or a cast from it) with the "one or more"
// batch dedup, plus Curator of Sun's Creation's once-each-turn check
// (the harvester counts a trigger the moment it is queued, as printed).
// Cadaver Lab's target is chosen when the trigger goes on the stack and
// a graveyard with no creature card never puts it there (CR 603.3d).
func init() {
	const crypt = "Defiled Crypt — create a 2/2 black Horror enchantment creature token"
	Register(Room(RoomSpec{
		OracleID:     "f97de425-c7ab-4688-8606-159370d179ee",
		Name:         "Defiled Crypt // Cadaver Lab",
		Completeness: CompletenessFull,
		Left: Door{Triggered: []game.TriggeredAbility{
			OncePerBatch(OnAny([]game.EventKind{game.EventZoneMove, game.EventCast},
				func(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
					return b16CardLeftYourGraveyard(ev, source, g) && curatorFirstThisTurn(crypt)(ev, source, lki, g)
				}, crypt, Do(CreateToken{Template: game.Card{
					Name: "Horror", TypeLine: "Token Enchantment Creature — Horror",
					Power: 2, Toughness: 2, Colors: []string{"B"},
				}, N: 1}))),
		}},
		Right: Door{Triggered: []game.TriggeredAbility{
			Targeting(
				WhenYouUnlockThisDoor(game.DoorRight, "Cadaver Lab — return a creature card from your graveyard to your hand",
					returnFirstLegalGraveyardTargetToHand),
				TargetCardInGraveyard("target creature card in your graveyard", YouOwn(), Creature())),
		}},
	}))
}
