package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Scute Swarm — Creature — Insect {2}{G}, 1/1:
//
//	"Landfall — Whenever a land you control enters, create a 1/1
//	 green Insect creature token. If you control six or more lands,
//	 create a token that's a copy of this creature instead."
//
// The batch-01 triage filed it under "protection / prevention /
// copying", and the copying half is what it really wanted:
// `CreateTokenCopy` (S16.5, and since #762 running the full entry
// pipeline) copies the Swarm, landfall ability and all, which is what
// makes the sixth land the point at which this stops being a Squirrel
// Nest and starts being a win condition.
//
// Three details the printed card decides and this follows:
//
//   - The land count is read as the TRIGGER RESOLVES, not when it was
//     put on the stack. The land that caused the trigger is on the
//     battlefield by then and counts itself, so the sixth land is the
//     first one that makes a copy.
//   - "Instead" is exclusive: at six lands the plain Insect is not
//     made.
//   - A copy is of THIS CREATURE — the Swarm's copiable values, so a
//     +1/+1 counter or an Aura on it is not copied (CR 707.2), and a
//     Swarm that has left the battlefield copies nothing rather than
//     erroring.
//
// The tokens are made one per land, so a Scapeshift or a
// Ramunap Excavator turn makes a trigger per land, each re-reading
// the count.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "aa854d50-444c-49d9-bfb1-5476b33c1c0b",
		Name:         "Scute Swarm",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Landfall("Scute Swarm — create an Insect, or a copy at six lands", scuteSwarmLandfall),
		},
	})
}

// scuteSwarmLandfall is the landfall body: a copy of the Swarm at six
// or more lands, a plain 1/1 green Insect below that.
func scuteSwarmLandfall(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	if b02CountLandsControlledBy(g, item.Controller) >= 6 {
		return CreateTokenCopy{
			Controller: item.Controller,
			Copy:       item.SourceCardID,
			N:          1,
		}.Apply(ctx)
	}
	return CreateToken{
		Controller: item.Controller,
		Template:   TokenCard("1/1 green Insect"),
		N:          1,
	}.Apply(ctx)
}
