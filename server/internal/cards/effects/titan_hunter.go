package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Titan Hunter — Creature — Human Warrior {4}{B}, 4/5 (Commander 2020,
// #5513):
//
//	"At the beginning of each player's end step, if no creatures died this
//	 turn, this creature deals 4 damage to that player.
//	 {1}{B}, Sacrifice a creature: You gain 4 life."
//
// The end-step trigger is Fevered Visions' "each player's end step" with the
// active player as "that player", behind an intervening "if" (CR 603.4) read
// off the turn's table-wide death tally (any creature, anyone's, a token
// included): it does not trigger when a creature has died this turn, and it
// does nothing if one dies before it resolves. The Hunter is the damage
// source, as it last existed if it has left (CR 608.2h). The sacrifice
// ability is Bubbling Cauldron's without the tap.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "abd9ed47-e3ef-40f3-9c18-769228e68792",
		Name:         "Titan Hunter",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventBeginEndStep},
			AppliesTo: func(ev game.Event, _ *game.Card, _ game.Characteristic, g *game.Game) bool {
				return ev.Kind == game.EventBeginEndStep && b11CreaturesDiedThisTurn(g) == 0
			},
			Key:    "Titan Hunter — 4 damage to that player if no creatures died this turn",
			Effect: titanHunterDamage,
		}},
		Activated: []ActivatedAbility{{
			Label:   "{1}{B}, Sacrifice a creature: You gain 4 life.",
			Purpose: game.Purpose{Answers: game.AnswerSacOutlet},
			Cost:    Plus(ManaCost("{1}{B}"), SacrificeACreature()),
			Effect:  Do(GainLife{Amount: 4}),
		}},
	})
}

// titanHunterDamage is the end-step trigger's resolution: the "if" again,
// then 4 damage to the player whose end step it is.
func titanHunterDamage(g *game.Game, item *game.StackItem) error {
	if b11CreaturesDiedThisTurn(g) > 0 {
		return nil
	}
	return DealDamage{Source: item.SourceCardID, Target: item.Trigger.Event.Actor, Amount: 4}.Apply(NewContext(g, item))
}
