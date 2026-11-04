package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Spiritcall Enthusiast // Scrollboost — Creature — Cat Cleric {2}{W}, 3/3 //
// Sorcery {1}{W} (preparation card, CR 722; Secrets of Strixhaven, #4456):
//
//	"Whenever one or more tokens you control enter, this creature becomes
//	 prepared. (While it's prepared, you may cast a copy of its spell. Doing
//	 so unprepares it.)"
//
//	Scrollboost — "One or two target creatures each get +2/+2 until end of
//	 turn."
//
// A preparation card: the permanent under the bare oracle ID and its prepare
// spell under "#1". The trigger watches token entries under your control
// once per batch (Caretaker's Talent's reading of the entering permanent,
// OncePerBatch); the Enthusiast itself is a card, not a token. It becomes
// prepared as the trigger resolves (BecomePrepared, which does nothing if it
// already is, or has left). Scrollboost pumps each target that is still
// legal as it resolves (CR 608.2b).
//
// No simplification.
func init() {
	const id = "c77ddcac-ce54-4348-bde2-5d9caf3d5b04"
	Register(Spec{
		OracleID:     id,
		Name:         "Spiritcall Enthusiast",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			OncePerBatch(On(game.EventETB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				c, ok := enteredUnderYourControl(ev, source, g, false)
				return ok && c.IsToken()
			}, "Spiritcall Enthusiast — becomes prepared",
				func(g *game.Game, item *game.StackItem) error {
					return BecomePrepared{Target: item.SourceCardID}.Apply(NewContext(g, item))
				})),
		},
	})
	Register(Spec{
		OracleID:     id + "#1",
		Name:         "Scrollboost",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("one or two target creatures").WithCount(1, 2),
		Purpose:      ForTargets(BenefitUntilEOT(0, 2, 2)),
		OnResolve:    scrollboostResolve,
	})
}

// scrollboostResolve gives each still-legal target +2/+2 until end of turn.
func scrollboostResolve(_ *game.StackItem, ctx *Context) error {
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetCard {
			continue
		}
		if err := (BoostUntilEOT{Target: t.ID, Power: 2, Toughness: 2, Label: "Scrollboost — +2/+2 until end of turn"}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
}
