package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Iron Man, Bleeding Edge — Legendary Artifact Creature — Human Hero
// {3}{U}{U}, 3/5 (Marvel Super Heroes Commander):
//
//	"Flying
//	 Whenever you cast an artifact spell, you may copy it, except the copy
//	 isn't legendary. Do this only once each turn. (The copy becomes a
//	 token.)"
//
// Baron Strucker's shape: the "you may" is the prompt asked as the trigger
// resolves (CR 603.5), and "only once each turn" limits the DOING
// (b28DoneThisTurn), so declining does not use the turn up and a second
// artifact spell cast in the same turn is offered but only the first "yes"
// copies. The copy is Double Major's (CopySpell with an "except" clause that
// removes Legendary, CR 707.10a) and keeps the original's targets; as a copy
// of a permanent spell it becomes a token as it resolves (CR 707.10f).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "fc6e170c-820c-48bb-95c7-ee79fba634de",
		Name:            "Iron Man, Bleeding Edge",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			Optional(On(game.EventCast, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				if ev.Actor != source.Controller || b28DoneThisTurn(g, source.InstanceID, ironManBleedingEdgeDone) {
					return false
				}
				spell, ok := g.LookupCardForEffect(ev.CardID)
				return ok && spell.IsArtifact()
			}, ironManBleedingEdgeLabel, ironManBleedingEdgeCopy),
				"Iron Man — copy that artifact spell?"),
		},
	})
}

const (
	ironManBleedingEdgeLabel = "Iron Man, Bleeding Edge — copy that artifact spell, except the copy isn't legendary"
	ironManBleedingEdgeDone  = "Iron Man, Bleeding Edge — copied an artifact spell this turn"
)

func ironManBleedingEdgeCopy(g *game.Game, item *game.StackItem) error {
	if b28DoneThisTurn(g, item.SourceCardID, ironManBleedingEdgeDone) {
		return nil
	}
	ctx := NewContext(g, item)
	b28NoteDoneThisTurn(g, item.SourceCardID, ironManBleedingEdgeDone)
	return CopySpell{
		StackID: ctx.Trigger().Event.CardID,
		Except:  func(v *game.PrintedValues) { v.RemoveSupertype("Legendary") },
	}.Apply(ctx)
}
