package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Rohirrim Chargers — Creature — Human Knight {2}{R}{W}, 4/4:
//
//	"You may exert this creature as it attacks. (It won't untap during
//	 your next untap step.)
//	 Whenever you exert a creature, reveal cards from the top of your
//	 library until you reveal an Equipment card. Put that card onto the
//	 battlefield attached to that creature, then put the rest on the
//	 bottom of your library in a random order."
//
// ADR 0130 §11: exert as it attacks (CR 701.43d) with no linked
// trigger, and a "whenever you exert a creature" payoff that sees this
// creature and any other creature its controller exerts (the Amonkhet
// ruling). "That creature" is the exerted permanent, recorded with its
// object epoch as the trigger is put on the stack (Params.Object): if
// it has left the battlefield, or left and come back as a new object
// (CR 400.7), the Equipment still enters, unattached.
//
// The reveal is the shared reveal-until sentence (revealUntil, as The
// Regalia). The Equipment is attached as it enters: CR 704.5n unattaches
// one that is illegally attached (protection from artifacts), so
// nothing stronger than printed is left on the table. A library with no
// Equipment is revealed whole and goes to the bottom in a random order.
//
// No simplification.
func init() {
	const label = "Rohirrim Chargers — reveal until an Equipment card; put it onto the battlefield attached to the exerted creature"
	payoff := WheneverYouExert(label, rohirrimChargersFetch)
	payoff.Build = whenYouExertBuild(label)
	Register(Spec{
		OracleID:      "8e638c52-e7b9-445b-adcf-77b58e06c2ba",
		Name:          "Rohirrim Chargers",
		Completeness:  CompletenessFull,
		ExertOnAttack: ExertAsItAttacks(),
		Triggered:     []game.TriggeredAbility{payoff},
	})
}

// rohirrimChargersFetch is the payoff's body.
func rohirrimChargersFetch(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	player := item.Controller
	exerted := item.Params.Object
	run, hit := revealUntil(ctx, player, func(c game.Card) bool { return c.HasSubtype("Equipment") },
		"Rohirrim Chargers — revealed until an Equipment card")
	return PutFromLibraryOntoBattlefield{
		Player: player,
		Cards:  run,
		Match:  func(_ *game.Game, _ uuid.UUID, c game.Card) bool { return hit != uuid.Nil && c.InstanceID == hit },
		All:    true,
		Then: func(g *game.Game, res PutFromLibraryResult) error {
			host, ok := g.LookupCardForEffect(exerted.ID)
			if ok && onBattlefield(g, exerted.ID) && host.ObjectEpoch == exerted.Epoch {
				for _, id := range res.Entered {
					if !onBattlefield(g, id) {
						continue
					}
					if err := g.AttachForEffect(id, game.TargetRef{Kind: game.TargetCard, ID: exerted.ID}); err != nil {
						return err
					}
				}
			}
			return PutRestOnBottomInRandomOrder(g, res)
		},
	}.Apply(ctx)
}
