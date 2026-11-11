package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Umbra Mystic — Creature — Human Wizard {2}{W}, 2/2 (Wilds of Eldraine
// Commander, #5457):
//
//	"Auras attached to permanents you control have umbra armor. (If an
//	 enchanted permanent you control would be destroyed, instead remove
//	 all damage from it and destroy an Aura attached to it.)"
//
// Each Aura attached to a permanent the Mystic's controller controls has
// umbra armor (CR 702.89a), whoever controls the Aura: an opponent's
// Pacifism on your creature saves it once. A permanent with several Auras
// loses one of them, chosen by its controller (CR 616.1), and keeps the
// rest for the next destruction. The Mystic itself is covered when it is
// enchanted, and a destruction that takes the Mystic and another
// enchanted permanent at once is still replaced for both (CR 614.4).
// Sacrifice, exile, bounce and toughness 0 or less are not destruction.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "c0a4f47c-c7f9-4b95-aa1b-30580e4cf030",
		Name:         "Umbra Mystic",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{AurasAttachedToPermanentsYouControlHaveUmbraArmor()},
	})
}
