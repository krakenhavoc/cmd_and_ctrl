package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Hallowed Haunting — Enchantment {2}{W}{W} (EDHREC rank 3123):
//
//	"As long as you control seven or more enchantments, creatures
//	 you control have flying and vigilance.
//	 Whenever you cast an enchantment spell, create a white Spirit
//	 Cleric creature token with "This token's power and toughness
//	 are each equal to the number of Spirits you control.""
//
// The enchantress deck's army. The gate is a layer 6 grant over the
// controller's creatures that applies only while they control
// seven or more enchantments, counted per recompute with the
// Haunting itself included (b29CreaturesYouControlWithSevenEnchantments);
// the trigger is Sigil of the Empty Throne's condition
// (enchantmentSpellCastByYou) and makes a Spirit Cleric.
//
// Sandbox simplification, and the reason the card carries a
// caveat (the Simulacrum Synthesizer posture): the Spirit Cleric's
// printed "power and toughness are each equal to the number of
// Spirits you control" is an ability OF THE TOKEN, and a token
// template has no static-ability slot and no oracle ID for the
// catalog to key one on. So the Haunting carries it on the tokens'
// behalf — a layer 7a set over the 0/0 Spirit Cleric tokens its
// controller controls, counting that controller's Spirits, applied
// by the FIRST Haunting the controller controls only, so two
// Hauntings size a token once (b29SpiritClericSizing). The
// observable difference is what happens when the last Haunting
// leaves: the tokens lose their sizing and shrink to 0/0, where
// printed they keep it for good. Weaker than printed, never
// stronger. (They shrink rather than die: the engine's toughness
// state-based action skips a printed 0/0 with no counters — the
// CurrentToughness convention — so a shrunken Cleric lingers as a
// 0/0 body.)
func init() {
	Register(Spec{
		OracleID:     "e310ab58-180e-4840-bc8d-9f9b06f1b478",
		Name:         "Hallowed Haunting",
		Completeness: CompletenessCaveats,
		Caveats:      []string{"The Spirit Cleric tokens get their size from Hallowed Haunting rather than on their own, so if it leaves the battlefield they shrink to 0/0."},
		Static: []game.StaticAbility{
			b16GrantKeywords(b29CreaturesYouControlWithSevenEnchantments, "flying", "vigilance"),
			b29SpiritClericSizing(),
		},
		Triggered: []game.TriggeredAbility{
			On(game.EventCast, enchantmentSpellCastByYou, "Hallowed Haunting — create a Spirit Cleric", Do(CreateToken{Template: TokenCard("0/0 white Spirit Cleric"), N: 1})),
		},
	})
}
