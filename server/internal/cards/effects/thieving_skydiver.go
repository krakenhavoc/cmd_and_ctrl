package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Thieving Skydiver — Creature — Merfolk Rogue {1}{U}, 2/1 (#3232, the
// Raffine Reanimator deck #3212):
//
//	"Kicker {X}. X can't be 0. (You may pay an additional {X} as you
//	 cast this spell.)
//	 Flying
//	 When this creature enters, if it was kicked, gain control of target
//	 artifact with mana value X or less. If that artifact is an
//	 Equipment, attach it to this creature."
//
// The seam is the kicker: KickerX("{X}", 1) is a kicker whose mana half
// is {X} (the caster's one announced X, CR 107.3a) with the printed
// floor on it, so a kicked cast announces X >= 1, an unkicked one
// announces no X at all (ADR 0073, amendment of 2026-10-13).
//
// The enters trigger is an intervening if (CR 603.4) read off the record
// the resolution carried onto the permanent (CR 400.7d): a Skydiver
// that was reanimated or flickered was not kicked and triggers nothing.
// Its target clause is built as the trigger goes on the stack
// (TargetsFrom, CR 603.3d) from source.CastX() (CR 107.3m), so "mana
// value X or less" is the X announced for the spell; with no artifact
// that cheap the trigger is removed without a prompt. The theft has no
// stated duration (CR 611.2a), so it is indefinite. The Equipment
// sentence is part of the same resolution: a Skydiver that has since
// left the battlefield attaches nothing (CR 701.3b), though the control
// change still happens.
//
// No simplification.
func init() {
	steal := On(game.EventETB, AllOf(Self, ThisKickedAtLeast(1)),
		thievingSkydiverLabel, thievingSkydiverSteal)
	steal.TargetsFrom = func(_ game.TriggerContext, source *game.Card, _ *game.Game) *game.TargetSpec {
		return TargetPermanent("target artifact with mana value X or less",
			Artifact(), ManaValueLE(source.CastX()))
	}
	Register(Spec{
		OracleID:        "f509ac02-f22e-4d1a-9fe1-afa96dd7806d",
		Name:            "Thieving Skydiver",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		OptionalCosts:   []game.AdditionalCost{KickerX("{X}", 1)},
		Triggered:       []game.TriggeredAbility{steal},
	})
}

const thievingSkydiverLabel = "Thieving Skydiver — gain control of target artifact with mana value X or less; attach it if it's an Equipment"

// thievingSkydiverSteal is the trigger's resolution: take the artifact,
// then, if it is an Equipment, attach it to the Skydiver.
func thievingSkydiverSteal(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetCard {
			continue
		}
		if err := (GainControl{
			Target:     t.ID,
			Controller: item.Controller,
			Duration:   game.IndefiniteDuration(),
			Label:      "Thieving Skydiver — gain control (no stated duration)",
		}).Apply(ctx); err != nil {
			return err
		}
		stolen, ok := g.LookupCardForEffect(t.ID)
		if !ok || !stolen.HasSubtype("Equipment") {
			return nil
		}
		// CR 701.3b: a Skydiver that is no longer on the battlefield
		// is not there to be attached to; the theft still stands.
		if info, ok := ctx.SourcePermanent(); !ok || info.Left {
			return nil
		}
		return g.AttachForEffect(t.ID, game.TargetRef{Kind: game.TargetCard, ID: item.SourceCardID})
	}
	return nil
}
