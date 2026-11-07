package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Magebane Lizard — Creature — Lizard {1}{R}, 1/4 (Outlaws of Thunder
// Junction, #4468):
//
//	"Whenever a player casts a noncreature spell, this creature deals
//	 damage to that player equal to the number of noncreature spells
//	 they've cast this turn."
//
// Every player's noncreature spell triggers it, its controller's
// included. The count is the caster's per-turn cast tally
// (CastTally.Noncreature), which the cast path bumps before EventCast,
// so the spell that triggered the Lizard counts: the first noncreature
// spell of a turn deals 1, the second 2. The number is read as the
// ability resolves, so a noncreature spell cast in response counts too.
// A caster who has left the game by then takes nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "0c38f159-367b-4e4b-94cd-bf1b0721c3a8",
		Name:         "Magebane Lizard",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventCast, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				spell, ok := g.LookupCardForEffect(ev.CardID)
				return ok && !game.ThroughTextOf(source, spell).IsCreature()
			}, "Magebane Lizard — deals damage to that player equal to the number of noncreature spells they've cast this turn",
				magebaneLizardDamage),
		},
	})
}

// magebaneLizardDamage is the trigger's resolution: damage from the
// Lizard to the caster, the caster's noncreature count this turn.
func magebaneLizardDamage(g *game.Game, item *game.StackItem) error {
	if item.Trigger == nil {
		return nil
	}
	caster := item.Trigger.Event.Actor
	if p := g.PlayerByIDForEffect(caster); p == nil || p.Eliminated {
		return nil
	}
	return DealDamage{
		Source: item.SourceCardID,
		Target: caster,
		Amount: g.CastTallyFor(caster).Noncreature,
	}.Apply(NewContext(g, item))
}
