package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Disciple of Freyalise // Garden of Freyalise — the FRONT face,
// Creature — Elf Druid {3}{G}{G}{G}, 3/3:
//
//	"When this creature enters, you may sacrifice another creature.
//	 If you do, you gain X life and draw X cards, where X is that
//	 creature's power."
//
// The land back (pay 3 life or enter tapped; {T}: Add {G}) is the
// mdfc_lands.go row under "<oracle>#1"; this is face 0, which keeps
// the bare oracle ID (game.CatalogKey).
//
// Ruthless Technomancer's shape exactly, with the payoff swapped from
// Treasures to life and cards: the sacrifice is chosen as the trigger
// RESOLVES (maySacrificeOne, CR 608.2c), so an opponent cannot answer
// the choice and hexproof is no obstacle. The life and cards are the
// sacrificed creature's last-known power (CR 608.2h), and they follow
// only a sacrifice that really happened — a sacrificed commander still
// pays out (CR 701.17a).
func init() {
	Register(Spec{
		OracleID:     "2699005b-a471-429f-a9d8-fbf2077ee2fd",
		Name:         "Disciple of Freyalise",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Disciple of Freyalise — sacrifice a creature, gain life and draw equal to its power", func(g *game.Game, item *game.StackItem) error {
				return maySacrificeOne(g, item, "Disciple of Freyalise — sacrifice another creature to gain that much life and draw that many cards?", Creature(), true, discipleOfFreyalisePayout)
			}),
		},
	})
}

// discipleOfFreyalisePayout is the life gain and the draw once the creature
// has really gone, keyed (#4290).
var discipleOfFreyalisePayout = RegisterSacrificedOne("prompt-run/disciple-of-freyalise-payout", func(ctx *Context, sacrificed uuid.UUID) error {
	// The draw waits for a prompt the life gain asks (#4162, CR 608.2c).
	ctx.RecordStep(sacrificedPowerRecord, nil, departedCreaturePower(ctx.Game, sacrificed))
	return gainThenDrawSacrificedPower.Run(ctx)
})
