package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Faith's Fetters — Enchantment — Aura for {3}{W}:
//
//	"Enchant permanent
//	 When this Aura enters, you gain 4 life.
//	 Enchanted permanent can't attack or block, and its activated
//	 abilities can't be activated unless they're mana abilities."
//
// The white catch-all: it answers a creature, an equipment, a
// planeswalker, a Sword of Feast and Famine, a Rogue's Passage.
// "Enchant permanent" is what makes it that, and it is the reason
// EnchantPermanent exists beside EnchantCreature — the CR 704.5n
// legality re-check runs the aura's own target spec every turn, so
// an Aura declared "enchant creature" falls off a host that stops
// being one and this one does not.
//
// Three clauses, three pieces of the S24 restriction vocabulary:
//
//   - "can't attack or block" — the same pair Pacifism carries.
//     Inert on a non-creature host, which is the printed behaviour;
//     Fetters on a Sol Ring is a worse Fetters, not a broken one.
//   - "activated abilities can't be activated" — CR 602.5a, checked
//     at announce before any cost is validated. On a planeswalker
//     this is the whole card: loyalty abilities ARE activated
//     abilities (CR 606.1), so a Fettered walker sits there doing
//     nothing while its counters go nowhere.
//   - "unless they're mana abilities" — the carve-out Arrest does
//     not print, and the reason the engine carries two bits instead
//     of one. A Fettered Birds of Paradise still taps for mana; a
//     Fettered Gaea's Cradle is still a land that works. Get this
//     wrong in the permissive direction and Fetters reads as Arrest;
//     get it wrong in the other and it shuts off half the mana base
//     it was pointed at.
//
// The ETB is a triggered ability and uses the stack (#578). It used to
// run from the direct AsEnters hook, which gave nobody a response
// window; now the trigger waits for every player to pass, like every
// other "When ~ enters".
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "2b2d76f5-4c9b-49dc-b202-68095e2d9b29",
		Name:         "Faith's Fetters",
		Completeness: CompletenessFull,
		Targets:      EnchantPermanent(),
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Faith's Fetters — you gain 4 life", Do(GainLife{Amount: 4})),
		},
		Static: []game.StaticAbility{
			RestrictAttached(game.CantAttackOrBlock | game.CantActivate),
		},
	})
}
