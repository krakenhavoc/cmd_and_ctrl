package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Combat Calligrapher — Creature — Bird Cleric {3}{W}, 3/3:
//
//	"Flying
//	 Inklings can't attack you or planeswalkers you control.
//	 Whenever a player attacks one of your opponents, that attacking
//	 player creates a tapped 2/1 white and black Inkling creature token
//	 with flying that's attacking that opponent."
//
// The restriction is written onto every creature with the Calligrapher
// as its source and binds only an Inkling, judged at declaration
// (socB3SubtypeCantAttackYouOrYourPlaneswalkers): an Inkling may attack
// the Calligrapher's opponents and their planeswalkers, never its
// controller or theirs.
//
// "Attacks one of your opponents" is CR 508.3e: creatures declared
// attacking that player, not a planeswalker or battle of theirs, and one
// trigger per attacked opponent per declaration (OncePerBatchPerPlayer).
// The attacker may be anyone, the Calligrapher's controller included. The
// Inkling is the ATTACKING player's, made tapped and already attacking
// the attacked opponent (CreateTokensAttackingForEffect): it was never
// declared, so it fires no attack trigger of its own (CR 508.3a) and is
// not held to the restriction (CR 508.4c).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "ee7ebbc1-2f9e-4867-a522-424869b6d772",
		Name:            "Combat Calligrapher",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Static: []game.StaticAbility{
			socB3SubtypeCantAttackYouOrYourPlaneswalkers("Inklings can't attack you or planeswalkers you control.", "Inkling"),
		},
		Triggered: []game.TriggeredAbility{
			OncePerBatchPerPlayer(On(game.EventAttack, socB3APlayerAttackedYourOpponent,
				"Combat Calligrapher — the attacking player creates a tapped 2/1 Inkling attacking that opponent",
				combatCalligrapherInkling)),
		},
	})
}

// combatCalligrapherInkling reads the attacking player and the attacked
// opponent off the triggering declaration.
func combatCalligrapherInkling(g *game.Game, item *game.StackItem) error {
	if item.Trigger == nil {
		return nil
	}
	ev := item.Trigger.Event
	attacker := g.PlayerByIDForEffect(ev.Actor)
	if attacker == nil || attacker.Eliminated {
		return nil
	}
	ctx := NewContext(g, item)
	tmpl := TokenCard("2/1 white and black Inkling with flying")
	tmpl.Tapped = true
	return g.CreateTokensAttackingForEffect(ev.Actor, ctx.TextToken(tmpl), 1, ev.Target)
}
