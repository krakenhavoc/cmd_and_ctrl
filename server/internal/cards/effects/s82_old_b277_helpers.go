package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// S82OldB277UpToOneCreatureDefendingPlayerControls is the TargetsFrom clause
// of "up to one target creature defending player controls" for a trigger on
// the attacking source itself (Impetuous Devils): TargetOfDefendingPlayer
// with its floor lowered to zero.
func S82OldB277UpToOneCreatureDefendingPlayerControls(tc game.TriggerContext, source *game.Card, g *game.Game) *game.TargetSpec {
	spec := TargetOfDefendingPlayer("creature", "up to one target creature defending player controls", Creature())(tc, source, g)
	if spec != nil {
		spec.Min = 0
	}
	return spec
}

// S82OldB277PlaneswalkerEnteredThisTurn is the intervening "if a
// planeswalker entered the battlefield under your control this turn"
// (Oath of Liliana), read for the source's controller through the per-turn
// entered-card-type tally, so one that has since left still counts.
func S82OldB277PlaneswalkerEnteredThisTurn(g *game.Game, player uuid.UUID) bool {
	return g.EnteredWithCardTypeThisTurn(player, "planeswalker") > 0
}

// S82OldB277SacrificedWhileCastingEmerge is the condition of "When you
// sacrifice this creature while casting a spell with emerge" (Foul
// Emissary): the event is this creature being sacrificed as a cost, and a
// spell its controller is casting (it is already on the stack while its
// cost is paid, CR 601.2h) has had its emerge cost chosen.
func S82OldB277SacrificedWhileCastingEmerge(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.Kind != game.EventSacrifice || ev.CardID != source.InstanceID || ev.Cause != game.MoveCauseCost {
		return false
	}
	if g.Stack == nil {
		return false
	}
	for _, c := range g.Stack.Cards {
		item := g.StackItemForEffect(c.InstanceID)
		if item != nil && item.Kind == game.StackItemSpell && item.Controller == ev.Actor && PaidEmerge(item.AltCost) {
			return true
		}
	}
	return false
}

// S82OldB277CondPlaneswalkerEntered is the Condition form of the same fact,
// for If / InterveningIf: it declares the footprint condition so the row is
// analysed like any other one-sentence trigger.
type S82OldB277CondPlaneswalkerEntered struct{}

func (S82OldB277CondPlaneswalkerEntered) Holds(ctx *Context) bool {
	return S82OldB277PlaneswalkerEnteredThisTurn(ctx.Game, ctx.Controller())
}

func (S82OldB277CondPlaneswalkerEntered) Condition() (game.FootprintStep, bool) {
	return CondStep(game.FootprintCondPlaneswalkerEntered, 0), true
}
