package effects

import (
	"github.com/google/uuid"

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
	Register(Spec{
		OracleID:     "c91e64d4-e687-4ccd-a56e-796cff6e052c",
		Name:         "Baron Strucker, HYDRA Overlord",
		Completeness: CompletenessFull,
		CostModifiers: []game.CostModifier{
			CostsLess(1, "Villain spells you cast cost {1} less to cast.",
				YourSpell(), OtherSpellOfCreatureType("Villain")),
		},
		Triggered: []game.TriggeredAbility{
			Optional(On(game.EventETB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return mshB2AnotherVillainEntered(ev, source, g) && !b28DoneThisTurn(g, source.InstanceID, baronStruckerDone)
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
	if b28DoneThisTurn(g, item.SourceCardID, baronStruckerDone) {
		return nil
	}
	b28NoteDoneThisTurn(g, item.SourceCardID, baronStruckerDone)
	return ConniveEach{Controller: item.Controller, Creatures: []uuid.UUID{item.Trigger.Event.CardID}}.Apply(NewContext(g, item))
}
