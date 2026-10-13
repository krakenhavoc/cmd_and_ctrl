package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Ruthless Technomancer — Creature — Human Wizard {3}{B}, 2/4 (EDHREC
// rank 2087):
//
//	"When this creature enters, you may sacrifice another creature
//	 you control. If you do, create a number of Treasure tokens equal
//	 to that creature's power.
//	 {2}{B}, Sacrifice X artifacts: Return target creature card with
//	 power X or less from your graveyard to the battlefield. X can't
//	 be 0."
//
// A Fling into Treasures, and a reanimator that eats them back. The
// ETB is Springbloom Druid's shape: the "you may" is the trigger's
// optional prompt, the creature is chosen as the trigger's target
// (any other creature you control — the picker the engine has for a
// choice among your own permanents), its power is read as it stands
// (counters and anthems included) just before the sacrifice, and
// that many Treasures follow.
//
// "If you do" is the sacrifice's own answer (#993). It used to be a
// live board read on the line after the sacrifice — "is the creature
// still on the battlefield?" — which is sacrificedThisWayLocked's rule
// written out by hand, and right for every outcome but the one that
// matters. A sacrificed COMMANDER is still on the battlefield while its
// owner answers CR 903.9, so the read said "not sacrificed" and the
// Treasures never came, for a sacrifice that landed a beat later. The
// clause hangs off SacrificePermanent.Then now, so the engine answers
// it after the move has settled: a commander that takes the command
// zone was still sacrificed (CR 701.17a — the keyword action is the
// move OFF the battlefield; only where it went was replaced), and a
// sacrifice the window cancelled outright pays nothing.
//
// The reanimation (ADR 0109 §9, #1842): "Sacrifice X artifacts" is the
// variable sacrifice cost (SacrificeX, ADR 0100), whose count IS the
// announced X (CR 107.3a), with "X can't be 0" as its floor (MinX).
// The target's "power X or less" reads that X, before the target is
// chosen (CR 602.2b, 601.2c) and again as the ability resolves
// (CR 608.2b). A card in a graveyard has its printed power.
//
// The sacrifice is chosen as the trigger RESOLVES (CR 608.2c), through
// maySacrificeOne, not as its target: an opponent cannot answer the
// choice, hexproof is no obstacle, and a creature removed in response
// just leaves the others to pick from. The Treasure count is the
// sacrificed creature's last-known power (CR 608.2h).
func init() {
	Register(Spec{
		OracleID:     "4e58ad76-37c7-4531-b207-6890b39a2679",
		Name:         "Ruthless Technomancer",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Ruthless Technomancer — sacrifice another creature, Treasures equal to its power", func(g *game.Game, item *game.StackItem) error {
				controller := item.Controller
				return maySacrificeOne(g, item, "Ruthless Technomancer — sacrifice another creature for Treasures equal to its power?", Creature(), true,
					func(ctx *Context, sacrificed uuid.UUID) error {
						power := departedCreaturePower(ctx.Game, sacrificed)
						if power <= 0 {
							return nil
						}
						return CreateToken{Controller: controller, Template: TreasureToken(), N: power}.Apply(ctx)
					})
			}),
		},
		Activated: []ActivatedAbility{{
			Label: "{2}{B}, Sacrifice X artifacts: Return target creature card with power X or less from your graveyard to the battlefield. X can't be 0.",
			Cost:  Plus(ManaCost("{2}{B}"), SacrificeX("X artifacts", Artifact()), MinX(1)),
			Targets: TargetCardInGraveyard("target creature card with power X or less from your graveyard", Creature(), YouOwn()).
				WithPowerAtMostX(),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				for _, t := range ctx.LegalTargets() {
					if t.Kind == game.TargetCard {
						return ReturnFromGraveyard{Target: t.ID, Dest: game.ZoneBattlefield, Controller: item.Controller}.Apply(ctx)
					}
				}
				return nil
			},
		}},
	})
}
