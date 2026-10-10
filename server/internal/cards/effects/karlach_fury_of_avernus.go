package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Karlach, Fury of Avernus — Legendary Creature — Tiefling Barbarian
// {4}{R}, 5/4:
//
//	"Whenever you attack, if it's the first combat phase of the turn,
//	 untap all attacking creatures. They gain first strike until end
//	 of turn. After this phase, there is an additional combat phase.
//	 Choose a Background"
//
// "Whenever you attack" is one trigger per attack declaration
// (OncePerBatch over EventAttack). "If it's the first combat phase of
// the turn" is an intervening if (CR 603.4), read off
// Turn.PhaseOrdinal as the attack is declared and again as the trigger
// resolves, so the added combat's own attack does not add a third
// (ADR 0059 sub-PR 2b, #753). No main phase comes with the combat: end
// of combat goes straight to the next beginning of combat (the
// 2022-06-10 ruling).
//
// Choose a Background is a deck-construction rule (CR 702.124k) that
// internal/deck reads off the oracle text: Karlach may share the
// command zone with a Background as a second commander (#2874).
//
// No simplification.
func init() {
	const label = "Karlach, Fury of Avernus — untap attackers, first strike, additional combat"
	Register(Spec{
		OracleID:     "037355be-71e7-4866-80a6-80352c304970",
		Name:         "Karlach, Fury of Avernus",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			OncePerBatch(On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return attackDeclaredByYou(ev, source.Controller) && IsFirstCombatPhase(g)
			}, label, karlachUntapAndCombat)),
		},
	})
}

// karlachUntapAndCombat is Karlach's trigger at resolution: the
// intervening if again (CR 603.4), then every attacking creature
// untaps and gains first strike, and another combat follows this one.
func karlachUntapAndCombat(g *game.Game, item *game.StackItem) error {
	if !IsFirstCombatPhase(g) {
		return nil
	}
	ctx := NewContext(g, item)
	attackers := AttackingCreatures(g)
	if err := untapEach(ctx, attackers); err != nil {
		return err
	}
	for _, id := range attackers {
		if err := (GrantKeywordUntilEOT{
			Target:   id,
			Keywords: []string{"first strike"},
			Label:    "Karlach, Fury of Avernus — first strike",
		}).Apply(ctx.asGroupMember()); err != nil {
			return err
		}
	}
	return ExtraCombatAfterThisPhase().Apply(ctx)
}
