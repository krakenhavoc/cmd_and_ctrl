package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Stabilizer — {2} Artifact (Scourge):
//
//	"Players can't cycle cards."
//
// #5843. One activation restriction (CR 602.5, CR 101.2) built by
// PlayersCantCycle: the board-wide gate refuses any ability carrying
// the cycling bit, which is every cycling and typecycling ability in
// the catalog (CR 702.29f: an effect that stops players from cycling
// stops typecycling too). It binds every player, Stabilizer's
// controller included, because the clause says "players".
//
// Nothing is stored. The restriction is derived from Stabilizer being on
// the battlefield, so it lifts the moment Stabilizer leaves, and a
// Stabilizer that loses its abilities stops restricting. The same gate
// is read by the activation path, the legal-move enumerator and the
// view's cant_activate stamp, so a refused cycling row is greyed with
// this clause and never offered to a bot.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "856c79ae-feab-4df9-8269-fffee7a44193",
		Name:         "Stabilizer",
		Completeness: CompletenessFull,
		ActivationRestrictions: []game.ActivationRestriction{
			PlayersCantCycle("Stabilizer — players can't cycle cards."),
		},
	})
}
