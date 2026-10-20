package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Rowan Kenrith — Legendary Planeswalker — Rowan {4}{R}{R}, loyalty 4:
//
//	"+2: During target player's next turn, each creature that player
//	     controls attacks if able.
//	 −2: Rowan Kenrith deals 3 damage to each tapped creature target
//	     player controls.
//	 −8: Target player gets an emblem with "Whenever you activate an
//	     ability that isn't a mana ability, copy it. You may choose new
//	     targets for the copy."
//	 Partner with Will Kenrith
//	 Rowan Kenrith can be your commander."
//
// +2 is an attack requirement (CR 508.1d) over the creatures the target
// player controls, read live (a requirement changes no characteristic,
// so CR 611.2c does not lock the set), for the whole of that player's
// next turn. Only the active player attacks, so when the target is not
// the active player the requirement is registered at once, lasting
// until the end of that player's next turn: nothing it binds can attack
// before then. When the target IS the active player (Rowan's controller
// targeting themself, the usual case at sorcery speed) the requirement
// must not bind the rest of this turn, so a delayed trigger at the
// beginning of that player's next upkeep registers it for that turn.
//
// −2 snapshots the tapped creatures the target player controls and
// deals the damage to them simultaneously (CR 608.2).
//
// −8's emblem is The Enigma Jewel's copy trigger: the activation event
// names the ability's stack item, the copy is created rather than
// activated (CR 707.10), so it does not trigger the emblem again, and a
// mana ability never uses the stack and is never copied (CR 605.3b).
// The emblem goes to the TARGET player (CR 114.2), and "you" on it is
// that player.
//
// "Partner with Will Kenrith" is the CR 702.124j entry search;
// "can be your commander" is deck construction, read from the oracle
// text by the deck checker.
//
// No simplification. Guffster's Party deck request (#4184).
func init() {
	Register(Spec{
		OracleID:        "e191976c-6ab1-4a23-aeb8-8a88bd4bd203",
		Name:            "Rowan Kenrith",
		Completeness:    CompletenessFull,
		StartingLoyalty: 4,
		Emblem: &EmblemSpec{
			Label: "Rowan Kenrith emblem",
			Text:  "Whenever you activate an ability that isn't a mana ability, copy it. You may choose new targets for the copy.",
			Triggered: []game.TriggeredAbility{
				On(game.EventActivateAbility, ByYou, "Rowan Kenrith emblem — copy that ability", copyTheTriggeringAbility),
			},
		},
		Triggered: []game.TriggeredAbility{
			PartnerWith("Rowan Kenrith", "Will Kenrith"),
		},
		Activated: []ActivatedAbility{
			{
				Label:   "+2: During target player's next turn, each creature that player controls attacks if able.",
				Cost:    LoyaltyCost(2),
				Targets: TargetPlayer("target player"),
				Effect:  rowanKenrithForceAttacks,
			},
			{
				Label:   "−2: Rowan Kenrith deals 3 damage to each tapped creature target player controls.",
				Cost:    LoyaltyCost(-2),
				Targets: TargetPlayer("target player"),
				Purpose: game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDamage, Amount: 3, Partial: true}},
				Effect:  rowanKenrithDamageTapped,
			},
			{
				Label:   "−8: Target player gets an emblem with \"Whenever you activate an ability that isn't a mana ability, copy it. You may choose new targets for the copy.\"",
				Cost:    LoyaltyCost(-8),
				Targets: TargetPlayer("target player"),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					p, ok := firstLegalPlayerTarget(ctx)
					if !ok {
						return nil
					}
					return CreateEmblem{Player: p}.Apply(ctx)
				},
			},
		},
	})
}

const rowanKenrithAttackLabel = "Rowan Kenrith — each creature that player controls attacks this turn if able"

// rowanKenrithAttacksDuringTheirTurnBody registers the requirement for
// the turn that has just begun: the delayed half of +2, scheduled when
// the target was the active player. The player rides Params.Player.
var rowanKenrithAttacksDuringTheirTurnBody = game.DelayedBody("rowan-kenrith/creatures-attack-this-turn",
	func(g *game.Game, item *game.StackItem, p game.EffectParams) error {
		player := p.Player
		if player == uuid.Nil {
			return nil
		}
		g.RegisterScopedRuleEffectForEffect(item.SourceCardID, game.ScopeYourCreatures, player,
			[]game.Mod{game.AddAttackRequirementMod(uuid.Nil)}, g.UntilEndOfTurnDuration(), rowanKenrithAttackLabel)
		return nil
	})

// rowanKenrithForceAttacks is +2.
func rowanKenrithForceAttacks(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	player, ok := firstLegalPlayerTarget(ctx)
	if !ok {
		return nil
	}
	active := uuid.Nil
	if s := g.Turn.ActiveSeat; s >= 0 && s < len(g.Seats) && g.Seats[s] != nil {
		active = g.Seats[s].ID
	}
	if player != active {
		g.RegisterScopedRuleEffectForEffect(ctx.Source(), game.ScopeYourCreatures, player,
			[]game.Mod{game.AddAttackRequirementMod(uuid.Nil)}, g.UntilEndOfYourNextTurnDuration(player), rowanKenrithAttackLabel)
		return nil
	}
	return ScheduleDelayedTrigger{
		At:     game.StepUpkeep,
		TurnOf: player,
		Label:  rowanKenrithAttackLabel,
		Body:   rowanKenrithAttacksDuringTheirTurnBody,
		Params: game.EffectParams{Player: player},
	}.Apply(ctx)
}

// rowanKenrithDamageTapped is −2: 3 damage to each tapped creature the
// target player controls, all at once.
func rowanKenrithDamageTapped(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	player, ok := firstLegalPlayerTarget(ctx)
	if !ok {
		return nil
	}
	var ids []uuid.UUID
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == player && c.IsCreature() && c.Tapped {
			ids = append(ids, c.InstanceID)
		}
	}
	return g.DamageInstanceForEffect(func() error {
		for _, id := range ids {
			if err := (DealDamage{Source: item.SourceCardID, Target: id, Amount: 3}).Apply(ctx); err != nil {
				return err
			}
		}
		return nil
	})
}
