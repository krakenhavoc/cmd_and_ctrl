package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Najeela, the Blade-Blossom — Legendary Creature — Human Warrior
// {2}{R}, 3/2:
//
//	"Whenever a Warrior attacks, you may have its controller create a 1/1
//	 white Warrior creature token that's tapped and attacking.
//	 {W}{U}{B}{R}{G}: Untap all attacking creatures. They gain trample,
//	 lifelink, and haste until end of turn. After this phase, there is an
//	 additional combat phase. Activate only during combat."
//
// The first ability is any Warrior's attack, whoever controls it (Najeela
// is one herself): Najeela's controller answers the "may", and the token
// is created by the ATTACKER's controller, who also chooses what it
// attacks. The token is put onto the battlefield attacking, never
// declared, so it triggers nothing (CR 508.4). "Warrior" is a word of
// Najeela's text and subtypes are read as they are now, so a changeling
// counts.
//
// The activated ability untaps every attacking creature, not only
// Najeela's side's, and gives each of them the three keywords until end
// of turn; the additional combat comes after the current combat phase
// (no main phase with it). "Activate only during combat" is any combat
// phase, any player's turn.
//
// No simplification.
func init() {
	const tokenLabel = "Najeela, the Blade-Blossom — its controller creates a 1/1 white Warrior token tapped and attacking"
	Register(Spec{
		OracleID:     "09619943-6aec-4080-ace5-a0c6ebb23f1c",
		Name:         "Najeela, the Blade-Blossom",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Optional(On(game.EventAttack, najeelaAWarriorAttacks, tokenLabel, najeelaWarriorToken),
				"Najeela, the Blade-Blossom — have its controller create a 1/1 white Warrior token that's tapped and attacking?"),
		},
		Activated: []ActivatedAbility{{
			Label:     "{W}{U}{B}{R}{G}: Untap all attacking creatures. They gain trample, lifelink, and haste until end of turn. After this phase, there is an additional combat phase. Activate only during combat.",
			Purpose:   game.Purpose{Answers: game.AnswerCombatGrant},
			Cost:      ManaCost("{W}{U}{B}{R}{G}"),
			Condition: duringACombatPhase(),
			Effect:    najeelaUntapAndCombat,
		}},
	})
}

// najeelaAWarriorAttacks is "whenever a Warrior attacks": the declaration
// of an attacker that is a Warrior as it is now.
func najeelaAWarriorAttacks(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.Kind != game.EventAttack {
		return false
	}
	c, ok := g.LookupCardForEffect(ev.CardID)
	return ok && c.IsCreature() && c.HasSubtype(source.TextSubtype("Warrior"))
}

// najeelaWarriorToken puts the token under the attacking Warrior's
// controller (the event's actor), attacking what that Warrior attacks by
// default.
func najeelaWarriorToken(g *game.Game, item *game.StackItem) error {
	if item.Trigger == nil || item.Trigger.Event.Actor == uuid.Nil {
		return nil
	}
	return CreateTokensAttackingYourChoice{
		Controller: item.Trigger.Event.Actor,
		Template:   TokenCard("1/1 white Warrior"),
		N:          1,
		Tapped:     true,
		Prefer:     triggerAttackTarget(item),
		Label:      "Najeela, the Blade-Blossom",
	}.Apply(NewContext(g, item))
}

// najeelaUntapAndCombat untaps every attacking creature, gives each
// trample, lifelink and haste until end of turn, and adds a combat phase
// after this one.
func najeelaUntapAndCombat(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	attackers := AttackingCreatures(g)
	if err := untapEach(ctx, attackers); err != nil {
		return err
	}
	for _, id := range attackers {
		if err := (GrantKeywordUntilEOT{
			Target:   id,
			Keywords: []string{"trample", "lifelink", "haste"},
			Label:    "Najeela, the Blade-Blossom — trample, lifelink and haste",
		}).Apply(ctx.asGroupMember()); err != nil {
			return err
		}
	}
	return ExtraCombatAfterThisPhase().Apply(ctx)
}

// duringACombatPhase is "Activate only during combat": any step of a
// combat phase.
func duringACombatPhase() ActivationCondition {
	return func(g *game.Game, _, _ uuid.UUID) bool {
		return game.PhaseOf(g.Turn.Step) == game.PhaseCombat
	}
}
