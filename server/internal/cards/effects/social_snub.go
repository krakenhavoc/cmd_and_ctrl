package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Social Snub — Sorcery {1}{W}{B} (Secrets of Strixhaven, #4456):
//
//	"When you cast this spell while you control a creature, you may copy
//	 this spell.
//	 Each player sacrifices a creature of their choice. Each opponent loses
//	 1 life and you gain 1 life."
//
// The cast trigger is the spell's own (WhenYouCastThisSpell, Ancestral
// Communion's shape). "While you control a creature" is part of what
// triggers it (CR 603.4 is about an "if" clause, and this has none), so the
// board is checked once, as the spell is cast, and not again when the
// trigger resolves. The copy is optional, asked as the trigger resolves
// (MayChoice with a keyed answer) and made by the caster (CR 707.10); it
// has no targets to choose, and it is its own spell, so it does all of the
// below too, above the original.
//
// The edict is every player's (the caster's too), in turn order; the life
// loss and gain wait until every chosen creature has been sacrificed
// (EachPlayerSacrifices' keyed "then", Smallpox's reading of a rest-of-the-
// spell clause).
//
// No simplification.
func init() {
	const label = "Social Snub — you may copy this spell"
	copyOnCast := WhenYouCastThisSpell(label,
		func(g *game.Game, item *game.StackItem) error {
			return MayChoice{
				Question: "Social Snub — copy this spell?",
				YesLabel: "Copy",
				NoLabel:  "Decline",
				Then:     socialSnubCopy,
			}.Apply(NewContext(g, item))
		})
	copyOnCast.AppliesTo = func(ev game.Event, source *game.Card, ch game.Characteristic, g *game.Game) bool {
		return Self(ev, source, ch, g) && socB35ControlsACreature(g, ev.Actor)
	}
	Register(Spec{
		OracleID:     "b284913e-3530-4a7e-8c75-efb18788dad0",
		Name:         "Social Snub",
		Completeness: CompletenessFull,
		Triggered:    []game.TriggeredAbility{copyOnCast},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return EachPlayerSacrifices{
				Match: Creature(),
				Label: "a creature",
				Keyed: socialSnubDrain,
			}.Apply(ctx)
		},
	})
}

// socialSnubCopy is the "yes" of the cast trigger: copy the spell it came
// from, which is still on the stack beneath the trigger.
var socialSnubCopy = MayYes("confirm/social-snub-copy", func(ctx *Context, r game.ConfirmAnswer) error {
	return CopySpell{StackID: ctx.Source(), Controller: r.Chooser}.Apply(ctx)
})

// socialSnubDrain is "each opponent loses 1 life and you gain 1 life", once
// every player has sacrificed.
var socialSnubDrain = RegisterEdictThen("prompt-run/social-snub-drain", func(ctx *Context, _ game.PromptedSacrifices) error {
	if err := (EachOpponentLosesLife{N: 1}).Apply(ctx); err != nil {
		return err
	}
	return GainLife{Amount: 1}.Apply(ctx)
})
