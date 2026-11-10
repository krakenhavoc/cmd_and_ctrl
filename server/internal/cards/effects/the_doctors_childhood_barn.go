package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// The Doctor's Childhood Barn — Plane — Gallifrey (Doctor Who, #4870):
//
//	"Creatures enter tapped.
//	 Whenever chaos ensues, for each opponent, choose up to one target
//	 nonland permanent that opponent controls. Untap those permanents.
//	 They phase out. They can't phase in for as long as The Doctor's
//	 Childhood Barn remains face up. When a player planeswalks, those
//	 permanents phase in."
//
// A plane (ADR 0216 §6), controlled by the planar controller (CR 901.6). The
// static is Kismet's CR 614 replacement for every creature, whoever controls
// it, a token included. The chaos ability (CR 311.7) has one "up to one"
// clause per opponent, bound to that opponent's permanents (perOpponentClauses).
// It untaps the permanents that are still legal targets as it resolves, then
// phases them out together with what is attached to them (CR 702.26g), locked
// by the plane: the lock is the face-up planar card (game/phasing.go), so
// they do not phase in at their controller's untap step while it is face up
// (the last ruling), and the planeswalk that turns it face down phases them in
// as the table settles after it, which is the printed "when a player
// planeswalks, those permanents phase in". Planeswalked away from before the
// ability resolves, they phase out and come back at their controller's next
// untap step as an ordinary phase-out does.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "2a21b90a-7ea7-4090-82ec-359ba8ba1aad",
		Name:         "The Doctor's Childhood Barn",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{{
			Watches: []game.EventKind{game.EventZoneMove},
			AppliesTo: func(ev *game.ReplacementEvent, g *game.Game, _ *game.Card) bool {
				if ev.Kind != game.RepEventMove || ev.NewZone != game.ZoneBattlefield {
					return false
				}
				entering, ok := g.LookupCardForEffect(ev.CardID)
				return ok && entering.IsCreature()
			},
			Replace: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) error {
				ev.EntersTapped = true
				return nil
			},
			Controller: func(_ *game.ReplacementEvent, _ *game.Game, src *game.Card) uuid.UUID {
				return src.Controller
			},
			Label: "The Doctor's Childhood Barn: creatures enter tapped",
		}},
		Triggered: []game.TriggeredAbility{
			Targeting(WheneverChaosEnsues("The Doctor's Childhood Barn — for each opponent, untap up to one target nonland permanent they control; they phase out and can't phase in while this plane is face up",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					ids := legalTargetCards(item, g)
					for _, id := range ids {
						if err := g.UntapTargetForEffect(id); err != nil {
							if err = nothingIfGone(err); err != nil {
								return err
							}
						}
					}
					if len(ids) == 0 {
						return nil
					}
					if g.PlanarFaceUp != nil && g.PlanarFaceUp.Contains(item.SourceCardID) {
						return g.PhaseOutUntilLeavesForEffect(ctx.Source(), item.SourceCardID, false, ids...)
					}
					return g.PhaseOutForEffect(ctx.Source(), ids...)
				}),
				perOpponentClauses(game.MaxPlayers-1, "nonland permanent", Nonland())),
		},
	})
}
