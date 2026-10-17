package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Huntmaster of the Fells // Ravager of the Fells — {2}{R}{G} Creature —
// Human Werewolf 2/2 // Creature — Werewolf 4/4 (#4071, ADR 0132
// amendment 2026-10-16):
//
//	Front: "Whenever this creature enters or transforms into Huntmaster
//	        of the Fells, create a 2/2 green Wolf creature token and you
//	        gain 2 life.
//	        At the beginning of each upkeep, if no spells were cast last
//	        turn, transform this creature."
//	Back:  "Trample
//	        Whenever this creature transforms into Ravager of the Fells,
//	        it deals 2 damage to target opponent or planeswalker and 2
//	        damage to up to one target creature that player or that
//	        planeswalker's controller controls.
//	        At the beginning of each upkeep, if a player cast two or more
//	        spells last turn, transform this creature."
//
// The front's trigger is one ability with two conditions: this permanent
// entering, or turning over to its front face (CR 701.27e, Brutal
// Cathar's shape). The Ravager's second target is tied to its first
// (ControlledByTargetsController): a creature controlled by the opponent
// targeted, or by the targeted planeswalker's controller, and it is
// judged again as the trigger resolves (CR 608.2b). Each half is dealt
// only to a target that is still legal; the Ravager deals the damage
// even if it has left the battlefield.
//
// No simplification.
func init() {
	const oracle = "582328cd-660d-47a4-bb23-e91e80b9a907"
	Register(Spec{
		OracleID:     oracle,
		Name:         "Huntmaster of the Fells",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			OnAny([]game.EventKind{game.EventETB, game.EventTransform}, func(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
				return (ev.Kind == game.EventETB && ev.CardID == source.InstanceID) || ThisTransformedInto(0)(ev, source, lki, g)
			}, "Huntmaster of the Fells — create a 2/2 green Wolf creature token and you gain 2 life",
				Do(CreateToken{Template: TokenCard("2/2 green Wolf"), N: 1}, GainLife{Amount: 2})),
			TransformAtUpkeepIfNoSpellsLastTurn("Huntmaster of the Fells"),
		},
	})
	ravage := On(game.EventTransform, ThisTransformedInto(1),
		"Ravager of the Fells — 2 damage to target opponent or planeswalker and 2 damage to up to one target creature that player or that planeswalker's controller controls",
		ravagerOfTheFellsDamage)
	ravage.Targets = Clauses(
		targetOpponentOrPlaneswalker(),
		ControlledByTargetsController(
			TargetCreature("up to one target creature that player or that planeswalker's controller controls").WithCount(0, 1), 0),
	)
	Register(Spec{
		OracleID:        oracle + "#1",
		Name:            "Ravager of the Fells",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{
			ravage,
			TransformAtUpkeepIfAPlayerCastTwoSpellsLastTurn("Ravager of the Fells"),
		},
	})
}

// ravagerOfTheFellsDamage deals 2 to each of the two targets that is
// still legal, as one damage instruction (CR 120).
func ravagerOfTheFellsDamage(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	var targets []uuid.UUID
	for slot := 0; slot < 2; slot++ {
		if t, ok := ctx.ClauseTarget(slot); ok {
			targets = append(targets, t.ID)
		}
	}
	if len(targets) == 0 {
		return nil
	}
	return g.DealDamageEachThenForEffect(item.SourceCardID, targets, 2, nil)
}
