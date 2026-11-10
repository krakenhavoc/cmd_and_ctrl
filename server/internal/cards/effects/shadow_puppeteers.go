package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Shadow Puppeteers — Creature — Faerie Wizard {6}{U}, 4/4 (Wilds of
// Eldraine Commander, #5368):
//
//	"Flying, ward {2}
//	 When this creature enters, create two 1/1 black Faerie Rogue
//	 creature tokens with flying.
//	 Whenever a creature you control with flying attacks, you may have it
//	 become a red Dragon with base power and toughness 4/4 in addition to
//	 its other colors and types until end of turn."
//
// The attack trigger fires once per flying attacker, read through the
// layers as it is declared (CR 508.2a). Declining does nothing. The
// change is one ScopedEffectFor record on that creature alone: base P/T
// at layer 7b, the colour (layer 5) and the subtype (layer 4) added to
// what it already has (CR 613.1d-e), until end of turn. If the attacker
// has left the battlefield by resolution nothing happens.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "0a570fdb-2104-4592-b158-5b5d88f1d087",
		Name:         "Shadow Puppeteers",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{CreatureTokens: 2},
		Triggered: []game.TriggeredAbility{
			Ward(WardMana("{2}"), "Shadow Puppeteers — ward {2}"),
			WhenThisEnters("Shadow Puppeteers — create two 1/1 black Faerie Rogue creature tokens with flying",
				Do(CreateToken{Template: FaerieRogueToken(), N: 2})),
			Optional(On(game.EventAttack, cmdr2B11YourFlyerAttacked,
				"Shadow Puppeteers — the attacking creature becomes a red Dragon with base power and toughness 4/4 until end of turn",
				cmdr2B11BecomeRedDragon),
				"Shadow Puppeteers — have the attacking creature become a red Dragon with base power and toughness 4/4 until end of turn?"),
		},
	})
}
