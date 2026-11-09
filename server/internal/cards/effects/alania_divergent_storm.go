package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Alania, Divergent Storm — Legendary Creature — Otter Wizard {3}{U}{R}, 3/5 (Bloomburrow):
//
//	"Whenever you cast a spell, if it's the first instant spell, the first
//	 sorcery spell, or the first Otter spell other than Alania you've cast this
//	 turn, you may have target opponent draw a card. If you do, copy that spell.
//	 You may choose new targets for the copy."
//
// The condition is a fact about the cast and is read as the spell is cast (the
// per-turn cast tally already counts it: Instants, Sorceries and
// OtterSpellsExceptAlania equal one on the first of each, CR 603.4); a later
// cast cannot make the spell stop having been the first, so it needs no second
// look. Each kind has its own "first", so the ability can trigger up to three
// times in a turn, and a spell that is both an instant and an Otter (a
// changeling's) triggers once. The target opponent is chosen as the trigger
// goes on the stack; "you may" is asked as it resolves and, on a yes, the
// opponent draws and the spell is copied from last-known information (CR
// 608.2h) with the new-targets choice (CR 707.10c). The copy is not cast, so no
// cast trigger sees it.
//
// No simplification.
var alaniaAnswer = MayYes("may/alania-divergent-storm-draw-and-copy", func(ctx *Context, _ game.ConfirmAnswer) error {
	spell := ctx.Trigger().Event.CardID
	if t, ok := ctx.ClauseTarget(0); ok && ctx.IsTargetLegal(t) {
		if err := (DrawCards{Player: t.ID, N: 1}).Apply(ctx); err != nil {
			return err
		}
	} else {
		return nil
	}
	return CopySpell{
		StackID:          spell,
		Controller:       ctx.Controller(),
		ChooseNewTargets: true,
		FromLastKnown:    true,
	}.Apply(ctx)
})

func init() {
	Register(Spec{
		OracleID:     "63d50d5e-56de-44c4-b906-b2a21a56e3f3",
		Name:         "Alania, Divergent Storm",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(On(game.EventCast, alaniaFirstOfAKind,
				"Alania, Divergent Storm — you may have target opponent draw a card; if you do, copy that spell",
				func(g *game.Game, item *game.StackItem) error {
					return MayChoice{
						Question: "Alania, Divergent Storm — have target opponent draw a card to copy that spell?",
						YesLabel: "Draw and copy",
						NoLabel:  "Decline",
						Then:     alaniaAnswer,
					}.Apply(NewContext(g, item))
				}),
				TargetPlayer("target opponent", Opponent())),
		},
	})
}

// alaniaFirstOfAKind is "you cast a spell, and it's the first instant spell, the
// first sorcery spell, or the first Otter spell other than Alania you've cast
// this turn".
func alaniaFirstOfAKind(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.Kind != game.EventCast || ev.Actor != source.Controller {
		return false
	}
	spell, ok := g.LookupCardForEffect(ev.CardID)
	if !ok {
		return false
	}
	tally := g.CastTallyFor(source.Controller)
	switch {
	case spell.IsInstant() && tally.Instants == 1:
		return true
	case spell.IsSorcery() && tally.Sorceries == 1:
		return true
	case spell.HasSubtype(source.TextSubtype("Otter")) && spell.Name != "Alania, Divergent Storm" && tally.OtterSpellsExceptAlania == 1:
		return true
	}
	return false
}
