package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Big Game Hunter — {1}{B}{B} Creature — Human Rebel Assassin, 1/1:
//
//	"When this creature enters, destroy target creature with power 4
//	 or greater. It can't be regenerated.
//	 Madness {B}"
//
// The other half of madness, and the half that proves the keyword is
// not just about instants: a creature discarded to a looter is cast
// for {B} at instant speed, which is a 1/1 body and a Murder on an
// opponent's attacker mid-combat. The ETB is a real targeted trigger
// (CR 603.3d — the target is chosen as the ability goes on the
// stack), so a table with nothing that big anywhere simply gets no
// trigger.
//
// "It can't be regenerated" is #667's rider on the shared destroy
// (`destroyTheTargetPermanentNoRegen`), the same body Terminate,
// Mortify and Putrefy use.
//
// Everything about the keyword is the engine's (#657): the card file
// is one trigger and one string. See fiery_temper.go and
// game/madness.go for the declared grant-instead-of-inline-cast
// simplification the two share.
func init() {
	Register(Spec{
		OracleID:     "ab55834f-c935-4773-89c6-bec9712284eb",
		Name:         "Big Game Hunter",
		Completeness: CompletenessFull,
		Madness:      "{B}",
		Triggered: []game.TriggeredAbility{
			Targeting(
				WhenThisEnters("Big Game Hunter — destroy target creature with power 4 or greater",
					func(g *game.Game, item *game.StackItem) error {
						return destroyTheTargetPermanentNoRegen(item, NewContext(g, item))
					}),
				TargetCreature("target creature with power 4 or greater", PowerGE(4))),
		},
	})
}
