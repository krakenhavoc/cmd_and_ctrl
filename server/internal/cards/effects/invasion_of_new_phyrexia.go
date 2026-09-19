package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Invasion of New Phyrexia — Battle — Siege, defense 6, for {X}{W}{U}:
//
//	"(As a Siege enters, choose an opponent to protect it. You and
//	  others can attack it. When it's defeated, exile it, then cast it
//	  transformed.)
//	 When this Siege enters, create X 2/2 white and blue Knight
//	 creature tokens with vigilance."
//
// The last of S27's four battles (#92, #626) and the only one whose
// back face is a planeswalker. The bracketed reminder text is engine
// behaviour keyed on the card type (server/internal/game/battle.go),
// not card-effect data; what this Spec carries is the X Knights, and
// the defeated trigger that hands over Teferi Akosa of Zhalfir.
//
// The Knights are why the front face was never the problem: X is
// announced at CR 601.2b and the tokens are just CreateToken with
// `N: ctx.X()`. What the two faces share is one oracle ID, and the
// back registers under it plus "#1" (see deluge_of_the_dead.go).
//
// # SANDBOX SIMPLIFICATION, declared: the Knights arrive a beat early
//
// Printed, "when this Siege enters" is a triggered ability: the
// battle enters, the trigger goes on the stack, the table gets a
// response window, and the tokens arrive when it resolves. Here they
// are created as the SPELL resolves, a beat before the battle enters,
// because that is the last moment the announced X is readable — the
// stack item is gone by the time the entry pipeline runs, and an ETB
// trigger built from the permanent has no X to read at all. It is
// Hydroid Krasis's and Goldvein Hydra's trade, made for their reason,
// and it is the weaker direction: the tokens are on the battlefield
// either way, and what is lost is the window between the two. What
// would close it is an entry that remembers what the spell announced,
// which is the same gap Hydroid Krasis's own caveat names.
func init() {
	Register(Spec{
		OracleID:     invasionOfNewPhyrexiaOracleID,
		Name:         "Invasion of New Phyrexia",
		Completeness: CompletenessCaveats,
		XMatters:     true,
		Caveats: []string{
			SiegeTransformedCastCaveat,
			"The X Knights are created as the Siege resolves rather than from an entry trigger, because the announced X is unreadable once the battle is a permanent — so nobody gets a response window between the Siege entering and the tokens arriving.",
		},
		Battle: &BattleSpec{
			Defense: 6,
			Subtype: BattleSubtypeSiege,
		},
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			// X can be 0, and CreateToken with N <= 0 makes nothing,
			// which is the printed outcome rather than a case to
			// special-case here.
			return CreateToken{
				Controller: item.Controller,
				Template:   TokenCard("2/2 white and blue Knight with vigilance"),
				N:          ctx.X(),
			}.Apply(ctx)
		},
		Triggered: []game.TriggeredAbility{
			DefeatedTrigger("Invasion of New Phyrexia — defeated: exile it, then cast Teferi Akosa of Zhalfir", SiegeDefeated()),
		},
	})
}

// invasionOfNewPhyrexiaOracleID is shared with the back face's spec,
// which registers under this ID plus "#1".
const invasionOfNewPhyrexiaOracleID = "480ea052-87d7-4fea-ac97-b6121b78f863"
