package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	pump := PumpAttached(2, 0)
	pump.Label = "Equipped creature gets +2/+0."
	Register(Spec{
		OracleID:     "68e45c07-96c5-4f87-a816-d9fa4f119740",
		Name:         "Gemcutter Buccaneer",
		Completeness: CompletenessFull,
		Grants: []AbilityGrant{
			{Key: gemcutterPumpGrant, Static: []game.StaticAbility{pump}, Text: "Equipped creature gets +2/+0."},
			{Key: gemcutterEquipPirateGrant, Activated: []ActivatedAbility{
				EquipOnlyAbility("Equip Pirate {1}", "{1}", TargetCreature("target Pirate you control", YouControl(), Subtype("Pirate"))),
			}, Text: "Equip Pirate {1}"},
			{Key: gemcutterEquipGrant, Activated: []ActivatedAbility{EquipAbility("{3}")}, Text: "Equip {3}"},
		},
		Static: []game.StaticAbility{
			AreAlsoEquipment("Treasures you control are Equipment in addition to their other types.", treasuresYouControl),
			GrantAbilities(treasuresYouControl, gemcutterPumpGrant, gemcutterEquipPirateGrant, gemcutterEquipGrant),
		},
		Triggered: []game.TriggeredAbility{
			On(game.EventETB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				c, ok := enteredUnderYourControl(ev, source, g, false)
				return ok && (c.InstanceID == source.InstanceID || isPirate(c))
			}, "Gemcutter Buccaneer — create a tapped Treasure", Do(CreateToken{
				Template: tappedTreasureToken(),
				N:        1,
			})),
		},
	})
}
