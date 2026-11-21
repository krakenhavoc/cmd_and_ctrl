package cardsb

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Baron Strucker, HYDRA Overlord — Legendary Creature — Human Villain
// {2}{B}, 2/2 (Marvel Super Heroes, #4455):
//
//	"Villain spells you cast cost {1} less to cast.
//	 Whenever another Villain you control enters, you may have it connive.
//	 Do this only once each turn."
//
// The discount is a board cost reduction on the controller's own Villain
// spells (a changeling counts). The connive trigger is Ondu Spiritdancer's
// shape: the "you may" is the CR 603.5 prompt asked as the trigger
// resolves, and "only once each turn" limits the DOING (b28DoneThisTurn),
// so declining does not use the turn up and a second Villain entering
// together with the first is offered but only the first "yes" connives.
// The Villain that connives is the one that entered, and ConniveEach
// draws, then discards, then counters it for a nonland card (CR 701.50a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c91e64d4-e687-4ccd-a56e-796cff6e052c",
		Name:         "Baron Strucker, HYDRA Overlord",
		Completeness: effects.CompletenessFull,
		CostModifiers: []game.CostModifier{
			effects.CostsLess(1, "Villain spells you cast cost {1} less to cast.",
				effects.YourSpell(), effects.OtherSpellOfCreatureType("Villain")),
		},
		Triggered: []game.TriggeredAbility{
			effects.Optional(effects.On(game.EventETB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return effects.MshB2AnotherVillainEntered(ev, source, g) && !effects.B28DoneThisTurn(g, source.InstanceID, baronStruckerDone)
			}, baronStruckerLabel, baronStruckerConnive),
				"Baron Strucker — have that Villain connive?"),
		},
	})
}

const (
	baronStruckerLabel = "Baron Strucker, HYDRA Overlord — the Villain may connive"
	baronStruckerDone  = "Baron Strucker, HYDRA Overlord — a Villain connived this turn"
)

func baronStruckerConnive(g *game.Game, item *game.StackItem) error {
	if effects.B28DoneThisTurn(g, item.SourceCardID, baronStruckerDone) {
		return nil
	}
	effects.B28NoteDoneThisTurn(g, item.SourceCardID, baronStruckerDone)
	return effects.ConniveEach{Controller: item.Controller, Creatures: []uuid.UUID{item.Trigger.Event.CardID}}.Apply(effects.NewContext(g, item))
}
