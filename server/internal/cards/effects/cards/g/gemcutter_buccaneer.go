package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Gemcutter Buccaneer — 1/3 Creature — Orc Pirate Artificer for
// {3}{R}:
//
//	"Whenever this creature or another Pirate you control enters,
//	 create a tapped Treasure token.
//	 Treasures you control are Equipment in addition to their other
//	 types and have 'Equipped creature gets +2/+0,' equip Pirate {1},
//	 and equip {3}."
//
// The first half is a Treasure for every Pirate, including its own
// arrival, which is why the trigger doesn't say "another".
//
// The second half is two statics over "Treasures you control" (#2562,
// ADR 0093 amendment 2026-10-10): a layer-4 "Equipment in addition to
// their other types" (CR 205.1b), and a layer-6 grant of three bundles.
// The equip rows are each Treasure's own activated abilities (CR 702.6a,
// 702.6d: either may be activated), so they attach the Treasure. The
// "+2/+0" is a granted layer-7c static, gathered after layer 6
// (game/granted_statics.go), applied to whatever that Treasure is
// attached to. A Treasure keeps its own sacrifice-for-mana ability. When
// the Buccaneer leaves, the Treasures stop being Equipment, and the
// state-based action unattaches them (CR 704.5n).
//
// No simplification.
const (
	gemcutterPumpGrant        = "gemcutter-buccaneer/treasure-pump"
	gemcutterEquipPirateGrant = "gemcutter-buccaneer/treasure-equip-pirate"
	gemcutterEquipGrant       = "gemcutter-buccaneer/treasure-equip"
)

func init() {
	pump := effects.PumpAttached(2, 0)
	pump.Label = "Equipped creature gets +2/+0."
	effects.Register(effects.Spec{
		OracleID:     "68e45c07-96c5-4f87-a816-d9fa4f119740",
		Name:         "Gemcutter Buccaneer",
		Completeness: effects.CompletenessFull,
		Grants: []effects.AbilityGrant{
			{Key: gemcutterPumpGrant, Static: []game.StaticAbility{pump}, Text: "Equipped creature gets +2/+0."},
			{Key: gemcutterEquipPirateGrant, Activated: []effects.ActivatedAbility{
				effects.EquipOnlyAbility("Equip Pirate {1}", "{1}", effects.TargetCreature("target Pirate you control", effects.YouControl(), effects.Subtype("Pirate"))),
			}, Text: "Equip Pirate {1}"},
			{Key: gemcutterEquipGrant, Activated: []effects.ActivatedAbility{effects.EquipAbility("{3}")}, Text: "Equip {3}"},
		},
		Static: []game.StaticAbility{
			effects.AreAlsoEquipment("Treasures you control are Equipment in addition to their other types.", effects.TreasuresYouControl),
			effects.Labelled(effects.GrantAbilities(effects.TreasuresYouControl, gemcutterPumpGrant, gemcutterEquipPirateGrant, gemcutterEquipGrant), "Treasures you control have \"Equipped creature gets +2/+0\", \"Equip Pirate {1}\" and \"Equip {3}\"."),
		},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventETB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				c, ok := effects.EnteredUnderYourControl(ev, source, g, false)
				return ok && (c.InstanceID == source.InstanceID || effects.IsPirate(game.ThroughTextOf(source, c)))
			}, "Gemcutter Buccaneer — create a tapped Treasure", effects.Do(effects.CreateToken{
				Template: effects.TappedTreasureToken(),
				N:        1,
			})),
		},
	})
}
