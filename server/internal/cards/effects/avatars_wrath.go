package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Avatar's Wrath — Sorcery {2}{W}{W}:
//
//	"Choose up to one target creature, then airbend all other
//	 creatures. (Exile them. While each one is exiled, its owner may
//	 cast it for {2} rather than its mana cost.)
//	 Until your next turn, your opponents can't cast spells from
//	 anywhere other than their hands.
//	 Exile Avatar's Wrath."
//
// A one-sided board wipe that gives every removed creature back cheap
// (the printed drawback for how one-sided it is), and a turn cycle of
// "no flashback, no cascade, no top-of-library, no exile-cast, no
// second copy in the graveyard" — a strictly narrower rider than
// Mandate of Peace's, and the proof #1316 needed that the rider
// composes with a real target clause and a real board effect on the
// same resolution.
//
// "CHOOSE UP TO ONE TARGET CREATURE, THEN AIRBEND ALL OTHER
// CREATURES." The spared creature is a TARGET (CR 601.2c) and everyone
// else is not — airbend does not target the rest of the board, so
// hexproof and shroud have nothing to say about them. Airbend is
// applied one creature at a time rather than through a batch primitive
// (ExileAllMatching moves cards but has no per-card grant to attach;
// Airbend's exile-with-a-permission is per instance), which is
// rules-correct here because nothing about the clause counts how many
// left or drags anything along with them — unlike Teferi's Protection
// phase-out, which needs one simultaneous call for CR 702.26h.
//
// "UNTIL YOUR NEXT TURN, YOUR OPPONENTS CAN'T CAST SPELLS FROM
// ANYWHERE OTHER THAN THEIR HANDS." #1316's proof: a
// game.CastBanRule{Kind: CastBanOutright, ExceptFromZone: ZoneHand}
// granted to each opponent, stamped with DurationUntilYourNextTurn —
// the exact shape the seam issue named this card for. One grant per
// opponent (RestrictCasting), because #1197's PlayerStatic slice is
// per SEAT and a granted statement about "your opponents" is not one
// statement about the table.
//
// "EXILE AVATAR'S WRATH." The spell moves ITSELF; #489's
// spellMovedItselfLocked is the check that stops the resolution frame
// putting it in the graveyard afterwards — the same line Teferi's
// Protection and Teferi's Reproach end on.
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:     "e3431dae-969c-4896-9f9e-a80e7bec4bdf",
		Name:         "Avatar's Wrath",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("up to one target creature").WithCount(0, 1),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			var spared uuid.UUID
			if len(item.Targets) > 0 && item.Targets[0].Kind == game.TargetCard {
				spared = item.Targets[0].ID
			}
			// "Airbend all other creatures." Snapshotted before any of
			// them moves, exactly as every other mass effect in this
			// catalog reads the board once before mutating it.
			for _, c := range ctx.Game.BattlefieldCardsForEffect() {
				if !c.IsCreature() || c.InstanceID == spared {
					continue
				}
				if err := (Airbend{Target: c.InstanceID}).Apply(ctx); err != nil {
					return err
				}
			}
			// "Until your next turn, your opponents can't cast spells
			// from anywhere other than their hands."
			d := DurationUntilYourNextTurn(ctx, ctx.Controller())
			for _, opp := range ctx.Opponents() {
				if err := (RestrictCasting{
					Player: opp,
					Rule: game.CastBanRule{
						Kind:           game.CastBanOutright,
						ExceptFromZone: game.ZoneHand,
					},
					Label:    "Avatar's Wrath — can't cast spells from anywhere other than their hand",
					Duration: d,
				}).Apply(ctx); err != nil {
					return err
				}
			}
			// "Exile Avatar's Wrath."
			return ctx.Game.ExileCardForEffect(item.SourceCardID)
		},
	})
}
