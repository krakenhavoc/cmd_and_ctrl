package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Solitary Sanctuary — Enchantment {2}{W} (Wilds of Eldraine, #4471):
//
//	"When this enchantment enters, tap target creature an opponent controls
//	 and put a stun counter on it. (If a permanent with a stun counter would
//	 become untapped, remove one from it instead.)
//	 Whenever you tap an untapped creature an opponent controls, put a +1/+1
//	 counter on target creature you control."
//
// The entry trigger is Cryotheory Adept's tap and stun. The second ability
// reads the tap's actor (game.TapTargetForEffectBy, the seam this card
// built): only a creature YOU tapped counts, so an opponent tapping their
// own creature, an attack, or a tap for mana does not, and its own entry
// trigger does (the Sanctuary is already on the battlefield). A creature
// that was already tapped is not tapped again and does not trigger.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "4c0a8eed-f2e2-4821-8011-6e6ad2081936",
		Name:         "Solitary Sanctuary",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(WhenThisEnters("Solitary Sanctuary — tap target creature an opponent controls and put a stun counter on it", tapTargetAndStunIt),
				TargetCreature("target creature an opponent controls", OpponentControls())),
			Targeting(On(game.EventTapCard, std52YouTappedUntappedOpposingCreature,
				"Solitary Sanctuary — put a +1/+1 counter on target creature you control", std52PlusOneCounterOnTarget),
				TargetCreature("target creature you control", YouControl())),
		},
	})
}
