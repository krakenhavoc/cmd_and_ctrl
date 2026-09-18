package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Switcheroo — Sorcery for {4}{U}:
//
//	"Exchange control of two target creatures."
//
// CR 701.12, and the reason exchange is a primitive rather than two
// GainControls: it is ONE effect. Both halves share a CR 613.7
// timestamp, so a later control-changer beats both or neither and no
// third effect can resolve "between" them. And it is all or nothing
// (CR 701.12b): if either creature has left the battlefield by the
// time this resolves, no control is exchanged at all — the surviving
// one stays where it is rather than being handed to a player who was
// meant to be trading for something.
//
// Neither half states a duration, so both are CR 611.2a indefinite:
// there is no "and they swap back". Each is pinned to its own
// creature, so if one dies later its half of the exchange goes with
// it and the other player keeps what they got, which is right —
// the exchange already happened.
//
// You are not required to be one of the two creatures' controllers,
// and the two may even be controlled by the same player, in which
// case nothing observable happens. That falls out of the primitive
// rather than needing a clause.
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:     "68227969-39cf-42cf-b3b8-cf8a04647d7e",
		Name:         "Switcheroo",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("two target creatures").WithCount(2, 2),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			// CR 608.2b: a target that became illegal in response is
			// skipped, and with only two targets that means the
			// exchange has nothing to do. `ExchangeControlForEffect`
			// refuses it anyway (CR 701.12b), so this is the earlier
			// of two guards, not the only one.
			legal := ctx.LegalTargets()
			if len(legal) != 2 {
				return nil
			}
			return ExchangeControl{
				A:     legal[0].ID,
				B:     legal[1].ID,
				Label: "Switcheroo — exchange control",
			}.Apply(ctx)
		},
	})
}
