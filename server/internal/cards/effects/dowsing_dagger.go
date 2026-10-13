package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Dowsing Dagger // Lost Vale — transforming double-faced card.
//
//	Dowsing Dagger — Artifact — Equipment {2}:
//	"When this Equipment enters, target opponent creates two 0/2 green
//	 Plant creature tokens with defender.
//	 Equipped creature gets +2/+1.
//	 Whenever equipped creature deals combat damage to a player, you may
//	 transform this Equipment.
//	 Equip {2}"
//	Lost Vale — Land:
//	"{T}: Add three mana of any one color."
//
// The enters trigger is the drawback that pays for the card: the TARGET
// OPPONENT is the token's controller, not the Dagger's. It targets (so
// hexproof and shroud on the player apply and it can be responded to),
// and a target that left the game by resolution does nothing (CR
// 608.2b).
//
// The transform is "you may" and is guarded by the engine's CR 701.27f
// rule (TransformThis): a Dagger that already turned over since the
// trigger went on the stack ignores the instruction. Turned over, the
// permanent is a land: it is no longer an Equipment (CR 301.5c), so the
// pump and equip lines go with the front face and it unattaches. The
// Vale takes its own catalog row at "<oracle_id>#1", as every back face
// does. Lost Vale has no summoning-sickness concern (it is not a
// creature) and taps for three mana of one color through the same
// OneColorOfAmount pick Coveted Jewel uses.
//
// No simplification.
const dowsingDaggerOracleID = "df34a6ad-ae1c-4470-8c9e-49815bba1973"

func init() {
	Register(Spec{
		OracleID:     dowsingDaggerOracleID,
		Name:         "Dowsing Dagger",
		Completeness: CompletenessFull,
		Static:       []game.StaticAbility{PumpAttached(2, 1)},
		Triggered: []game.TriggeredAbility{
			Targeting(
				WhenThisEnters("Dowsing Dagger — target opponent creates two 0/2 green Plant creature tokens with defender",
					func(g *game.Game, item *game.StackItem) error {
						ctx := NewContext(g, item)
						targets := ctx.LegalTargets()
						if len(targets) == 0 {
							return nil
						}
						return CreateToken{
							Controller: targets[0].ID,
							Template:   TokenCard("0/2 green Plant with defender"),
							N:          2,
						}.Apply(ctx)
					}),
				TargetPlayer("target opponent", Opponent())),
			Optional(
				On(game.EventDealDamage, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					return attachedCreatureDealtCombatDamageToPlayer(ev, source, g)
				}, "Dowsing Dagger — you may transform this Equipment", func(g *game.Game, item *game.StackItem) error {
					return TransformThis{}.Apply(NewContext(g, item))
				}),
				"Dowsing Dagger — transform it into Lost Vale?"),
		},
		Activated: []ActivatedAbility{
			EquipAbility("{2}"),
		},
	})

	Register(Spec{
		OracleID:     dowsingDaggerOracleID + "#1",
		Name:         "Lost Vale",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: OneColorOfAmount(3),
			Label:    "Add three mana of any one color",
		}},
	})
}
