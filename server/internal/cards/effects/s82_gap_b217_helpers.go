package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_gap_b217_helpers.go — helpers for GAP batch 217 (#5948), Commander
// Legends: Battle for Baldur's Gate green creatures.

// CondGapB217TotalPowerAtLeast is "if creatures you control have total power
// N or greater" as an intervening-if Condition, read for the item's
// controller as it resolves (CR 603.4, CR 208.3). Power counts through
// counters and effects, as B43TotalPowerControlled reads it.
type CondGapB217TotalPowerAtLeast struct{ N int }

func (c CondGapB217TotalPowerAtLeast) Holds(ctx *Context) bool {
	return B43TotalPowerControlled(ctx.Game, ctx.Controller()) >= c.N
}

func (CondGapB217TotalPowerAtLeast) Condition() (game.FootprintStep, bool) {
	return CondStep(game.FootprintCondBoard, 0), true
}

// YouControlTotalPowerAtLeast is the When for the same condition as the
// trigger fires; "you" is the source's controller.
func YouControlTotalPowerAtLeast(n int) When {
	return func(_ game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
		return B43TotalPowerControlled(g, source.Controller) >= n
	}
}

// AtYourEndStepIfTotalPowerAtLeast is "At the beginning of your end step, if
// creatures you control have total power N or greater, <effect>": an
// intervening "if" checked as the step begins and as the ability resolves.
func AtYourEndStepIfTotalPowerAtLeast(n int, label string, effect Effect) game.TriggeredAbility {
	return InterveningIf(CondGapB217TotalPowerAtLeast{N: n}, On(game.EventBeginEndStep, AllOf(ByYou, YouControlTotalPowerAtLeast(n)), label, effect))
}

// GapB217GreatestDragonPower is Skanos Dragonheart's X: the greatest power
// among OTHER Dragons `controller` controls (read live through the layers,
// CR 208.3) and Dragon cards in their graveyard (printed power, there being
// no layers in a graveyard). `dragon` is the creature type word as the source's
// text reads it (CR 612.2). Zero when there are none.
func GapB217GreatestDragonPower(g *game.Game, controller, except uuid.UUID, dragon string) int {
	best := 0
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == controller && c.InstanceID != except && c.IsCreature() && c.HasSubtype(dragon) {
			best = max(best, c.CurrentPower())
		}
	}
	if p := g.PlayerByIDForEffect(controller); p != nil {
		for _, c := range p.Graveyard.Cards {
			if c.HasSubtype(dragon) {
				best = max(best, c.Power)
			}
		}
	}
	return best
}

// GapB217LandsControlledBy counts the lands `player` controls, read through
// the layers (CR 305.7): Owlbear Cub's "a player who controls eight or more
// lands".
func GapB217LandsControlledBy(g *game.Game, player uuid.UUID) int {
	n := 0
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == player && c.HasCardType("land") {
			n++
		}
	}
	return n
}

// CantAttackUnlessYouControl is "This creature can't attack unless you control
// <a permanent matching any of the queries>" (Lovestruck Beast's 1/1 creature,
// #5948, CR 508.1c): a fact about the attacking creature's controller read live
// at declaration, so it refuses every target while unmet.
//
//	Static: []game.StaticAbility{CantAttackUnlessYouControl(game.PermanentQuery{Types: []string{"creature"}}.WithExactPT(1, 1))}
func CantAttackUnlessYouControl(queries ...game.PermanentQuery) game.StaticAbility {
	return selfAttackTargetRestriction(game.AttackTargetRestriction{ControllerMustControl: queries})
}

// GapB217PayZeroForACreatureSpellFromExileOnceEachTurn is Tlincalli Hunter's
// static: "Once each turn, you may pay {0} rather than pay the mana cost for a
// creature spell you cast from exile." Warped Space's offer narrowed to creature
// spells (CR 118.9): free, so X is 0 (CR 107.3b); only where the printed mana cost
// could be paid from exile, so a foretold or plotted card, which is already
// alternatively costed, gets none. The limit is per Tlincalli Hunter and spent as
// the spell becomes cast.
func GapB217PayZeroForACreatureSpellFromExileOnceEachTurn() game.GrantedAlternativeCost {
	return game.GrantedAlternativeCost{
		Offer: game.AlternativeCost{
			Key:   GapB217GrantedAltCostZeroCreatureFromExile,
			Label: "Pay {0} rather than pay this spell's mana cost",
		},
		Spells:       game.PermissionFilter{CreatureOnly: true},
		Zones:        []game.ZoneKind{game.ZoneExile},
		OnceEachTurn: true,
	}
}

// GapB217GrantedAltCostZeroCreatureFromExile is the on-disk key of Tlincalli
// Hunter's granted offer. Never renamed, never reused.
const GapB217GrantedAltCostZeroCreatureFromExile = "granted-zero-creature-from-exile-once-each-turn"
