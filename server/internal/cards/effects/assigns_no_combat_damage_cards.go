package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// assigns_no_combat_damage_cards.go — the shared shapes of the "whenever
// this creature attacks and isn't blocked, you may <do something>. If you
// do, this creature assigns no combat damage this turn" cards (Ophidian,
// Farrel's Zealot, Floral Spuzzem and the rest of #3531). The rule itself
// is game.AssignsNoCombatDamageThisTurnForEffect (CR 510.1a).

// assignsNoCombatDamageThisTurn is "this creature assigns no combat damage
// this turn": the resolving ability's source, pinned as the object it is
// now, so a creature that left and came back is not marked (CR 400.7).
func assignsNoCombatDamageThisTurn(ctx *Context, label string) {
	if ctx.isNewSourceObjectAsThis(ctx.Source()) {
		return
	}
	ctx.Game.AssignsNoCombatDamageThisTurnForEffect(ctx.Source(), ctx.Source(), label)
}

// unblockedAttackerMay builds "whenever this creature attacks and isn't
// blocked, you may <do>. If you do, this creature assigns no combat damage
// this turn". `do` gets the ability's first legal card target (uuid.Nil
// for an ability with none) and runs only on "yes". The caller adds the
// target clause (Targeting, or TargetsFrom for "defending player controls").
func unblockedAttackerMay(card, question string, do func(ctx *Context, target uuid.UUID) error) game.TriggeredAbility {
	label := card + " — " + question
	return WhenAttacksAndIsNotBlockedEffect(label, func(g *game.Game, item *game.StackItem, _ uuid.UUID) error {
		ctx := NewContext(g, item)
		target := firstLegalCardTarget(ctx)
		return MayChoice{
			Question: label + "?",
			OnYes: func(ctx *Context) error {
				if err := do(ctx, target); err != nil {
					return err
				}
				assignsNoCombatDamageThisTurn(ctx, card+" — assigns no combat damage this turn")
				return nil
			},
		}.Apply(ctx)
	})
}

// firstLegalCardTarget is the first still-legal card target of the
// resolving ability, in any zone, or uuid.Nil.
func firstLegalCardTarget(ctx *Context) uuid.UUID {
	for _, t := range ctx.LegalTargets() {
		if t.Kind == game.TargetCard && t.ID != uuid.Nil {
			return t.ID
		}
	}
	return uuid.Nil
}

// sourcePowerNow is "its power" for the resolving ability's source: its
// current power on the battlefield, else its last known power (CR 608.2h).
func sourcePowerNow(ctx *Context) int {
	id := ctx.Source()
	if onBattlefield(ctx.Game, id) {
		return b31CurrentPowerOf(ctx.Game, id)
	}
	return b17LastKnownPowerOffBattlefield(ctx.Game, id)
}

// withTargetsFrom gives a trigger a TargetsFrom clause (chosen as the
// ability goes on the stack, from the board as it is then).
func withTargetsFrom(t game.TriggeredAbility, f func(game.TriggerContext, *game.Card, *game.Game) *game.TargetSpec) game.TriggeredAbility {
	t.TargetsFrom = f
	return t
}

// TargetOfDefendingPlayer is the TargetsFrom clause for "target <thing>
// defending player controls" on a trigger that fires while the source is
// attacking (Floral Spuzzem, Orcish Squatters): the defending player is
// read off the source's own attack, as TargetCreatureDefendingPlayerControls
// does. `pred` narrows the permanent.
func TargetOfDefendingPlayer(mode, label string, pred CardPredicate) func(game.TriggerContext, *game.Card, *game.Game) *game.TargetSpec {
	return func(_ game.TriggerContext, source *game.Card, g *game.Game) *game.TargetSpec {
		defender := g.DefendingPlayerForAttackForEffect(source.AttackingTarget)
		if defender == uuid.Nil {
			return nil
		}
		return &game.TargetSpec{
			Mode: mode, Label: label,
			Zones: []game.ZoneKind{game.ZoneBattlefield},
			CardOK: func(g *game.Game, caster uuid.UUID, c game.Card, _ game.ZoneKind) bool {
				return c.Controller == defender && pred(g, caster, c)
			},
			Min: 1, Max: 1,
		}
	}
}
