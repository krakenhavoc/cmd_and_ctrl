package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// kimoyoBeadsLabel is the trigger's stack label, and with it the key
// its "hasn't been chosen" memory is kept under.
const kimoyoBeadsLabel = "Kimoyo Beads — beginning of your end step"

// Kimoyo Beads — Artifact {4}:
//
//	"At the beginning of your end step, choose one that hasn't been
//	 chosen —
//	 • AV Bead — Draw a card.
//	 • Communication Bead — Create two 1/1 white Soldier creature
//	   tokens.
//	 • Prime Bead — You gain 3 life. Exile this artifact, then return it
//	   to the battlefield under its owner's control."
//
// ChooseOneNotChosen (ADR 0097): each bullet once for this object. The
// third bullet is the reset — exiled and returned, the Beads are a new
// object (CR 400.7) with no memory of the modes chosen, so the next
// end step may choose all three again. The bullet names are flavour
// words.
//
// The flicker moves only the Beads that triggered: a Beads that left
// and came back in response is already a new object, and one that is
// no longer on the battlefield has nothing to exile. The life is
// gained either way.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "c7bdbf7a-8054-4e91-bb32-84f47e98553c",
		Name:         "Kimoyo Beads",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			kimoyoBeadsTrigger(),
		},
	})
}

func kimoyoBeadsTrigger() game.TriggeredAbility {
	t := AtYourEndStep(kimoyoBeadsLabel, func(_ *game.Game, _ *game.StackItem) error { return nil })
	t.Modes = ChooseOneNotChosen(
		ModeDoing("AV Bead — Draw a card.", nil,
			func(item *game.StackItem, ctx *Context, _ int) error {
				return DrawCards{Player: item.Controller, N: 1}.Apply(ctx)
			}),
		ModeDoing("Communication Bead — Create two 1/1 white Soldier creature tokens.", nil,
			func(item *game.StackItem, ctx *Context, _ int) error {
				return CreateToken{Controller: item.Controller, Template: TokenCard("1/1 white Soldier"), N: 2}.Apply(ctx)
			}),
		ModeDoing("Prime Bead — You gain 3 life. Exile this artifact, then return it to the battlefield under its owner's control.", nil,
			func(item *game.StackItem, ctx *Context, _ int) error {
				if err := (GainLife{Player: item.Controller, Amount: 3}).Apply(ctx); err != nil {
					return err
				}
				if !sourceIsStillThisPermanent(ctx.Game, item) {
					return nil
				}
				return Flicker{Target: item.SourceCardID}.Apply(ctx)
			}),
	)
	return t
}
