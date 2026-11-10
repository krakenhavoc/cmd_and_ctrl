package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Starforged Sword — Artifact — Equipment {4} (Bloomburrow, #4467):
//
//	"Gift a tapped Fish (You may promise an opponent a gift as you cast
//	 this spell. If you do, when it enters, they create a tapped 1/1 blue
//	 Fish creature token.)
//	 When this Equipment enters, if the gift was promised, attach this
//	 Equipment to target creature you control.
//	 Equipped creature gets +3/+3 and loses flying.
//	 Equip {3}"
//
// Gift is the shared keyword (gift.go): the promise rides the cast, and
// the Fish is the keyword's own entry trigger (CR 702.174b). The attach
// is Scrapshooter's gated entry trigger, read off the Sword's own record
// of the promise (an intervening if, CR 603.4, that cannot change), with
// Falcon's Wing Harness's AttachSourceToTarget. The statics are Colossus
// Hammer's: a layer 7c +3/+3 and a layer 6 "loses flying" on the
// equipped creature.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "2cdc1d82-30dc-4f0a-af20-33dde90269b1",
		Name:         "Starforged Sword",
		Completeness: CompletenessFull,
		Gift:         GiftATappedFish(),
		Static: []game.StaticAbility{
			PumpAttached(3, 3),
			RemoveFromAttached("flying"),
		},
		Triggered: []game.TriggeredAbility{
			Targeting(On(game.EventETB, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return ev.CardID == source.InstanceID && source.GiftPromised()
			}, "Starforged Sword — attach it to target creature you control", AttachSourceToTarget),
				PermanentYouControl("target creature you control", Creature())),
		},
		Activated: []ActivatedAbility{
			EquipAbility("{3}"),
		},
	})
}
