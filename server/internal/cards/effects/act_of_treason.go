package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Act of Treason — Sorcery for {2}{R}:
//
//	"Gain control of target creature until end of turn. Untap that
//	 creature. It gains haste until end of turn."
//
// The card #756 was opened for, and the proof that the control
// primitive is the right shape: all three clauses are ordinary
// primitives in printed order, and nothing about "give it back at
// cleanup" is written anywhere.
//
// The theft is a layer-2 continuous effect with an until-end-of-turn
// duration (ADR 0063). Control reverts by itself when the cleanup
// sweep drops the entry, because the layer engine reseeds every
// permanent's controller from `Card.BaseController` on the next pass.
// If the creature was already stolen by a Mind Control, this wins on
// timestamp (CR 613.7) and hands it back to the Mind Control's
// controller at cleanup, not to its owner — which is the printed
// interaction.
//
// The other two clauses are why the card is three lines and not one:
//
//   - **Untap.** The creature you steal is usually tapped, from
//     attacking or from paying for something.
//   - **Haste.** A permanent that changes control is summoning-sick
//     under its new controller (CR 302.6) however long it has been on
//     the battlefield, so without this clause the theft would buy an
//     attack from nothing. The grant is its own layer-6
//     until-end-of-turn effect, which is what "It gains haste until
//     end of turn" is.
//
// And CR 506.4 rides along for free: stealing an attacking creature
// removes it from combat, and since S38 that clears the attack
// announcement too, so it can be declared again under its new
// controller in a later combat.
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:     "9d08af23-9f4a-4097-9abc-3b17475ab744",
		Name:         "Act of Treason",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
				return nil
			}
			target := item.Targets[0].ID
			if err := (GainControl{
				Target:   target,
				Duration: DurationUntilEndOfTurn(ctx),
				Label:    "Act of Treason — gain control until end of turn",
			}).Apply(ctx); err != nil {
				return err
			}
			if err := (UntapTarget{Target: target}).Apply(ctx); err != nil {
				return err
			}
			return GrantKeywordUntilEOT{
				Target:   target,
				Keywords: []string{"haste"},
				Label:    "Act of Treason — haste until end of turn",
			}.Apply(ctx)
		},
	})
}
