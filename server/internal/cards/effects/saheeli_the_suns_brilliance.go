package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Saheeli, the Sun's Brilliance — Legendary Creature — Human Artificer
// {U}{R}, 2/2 (The Lost Caverns of Ixalan, #4470):
//
//	"{U}{R}, {T}: Create a token that's a copy of another target creature
//	 or artifact you control, except it's an artifact in addition to its
//	 other types. It gains haste. Sacrifice it at the beginning of the
//	 next end step."
//
// Chrome Dome's ability with a wider target and Mishra, Eminent One's kind
// of except clause. "Another" is object identity, so Saheeli can't copy
// herself. The except clause is part of the copy (CR 707.9b): Artifact is
// added to the copied type line before the token exists. "It gains haste"
// has no end, so it is an indefinite grant on the token, and the sacrifice
// is a delayed trigger at the next end step (CR 603.7) on the tokens the
// creation actually made. The {T} makes Saheeli subject to summoning
// sickness (CR 302.6).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "5aa76c5d-8046-42ce-b61d-cc9b4806c11a",
		Name:         "Saheeli, the Sun's Brilliance",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{U}{R}, {T}: Create a token that's a copy of another target creature or artifact you control, except it's an artifact in addition to its other types. It gains haste. Sacrifice it at the beginning of the next end step.",
			Cost:    Plus(ManaCost("{U}{R}"), TapCost()),
			Targets: Another(TargetPermanent("another target creature or artifact you control", Or(Creature(), Artifact()), YouControl())),
			Effect:  saheeliSunsBrillianceCopy,
		}},
	})
}

// saheeliSunsBrillianceCopy makes the artifact token copy of the
// still-legal target, then hands the made tokens to the keyed rest.
func saheeliSunsBrillianceCopy(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	copyOf := FirstLegalBattlefieldTarget(ctx)
	if copyOf == uuid.Nil {
		return nil
	}
	return CreateTokenCopiesTogether{
		Controller: item.Controller,
		Copies:     []uuid.UUID{copyOf},
		Except:     tokenCopyIsAlsoAnArtifact,
		Then:       ctx.ItemTail(saheeliSunsBrillianceRest),
	}.Apply(ctx)
}

// tokenCopyIsAlsoAnArtifact is "except it's an artifact in addition to
// its other types" (CR 707.9b).
func tokenCopyIsAlsoAnArtifact(t *game.Card) {
	v := game.PrintedValues{TypeLine: t.TypeLine}
	v.AddCardType("Artifact")
	t.TypeLine = v.TypeLine
}

// saheeliSunsBrillianceRest is "It gains haste. Sacrifice it at the
// beginning of the next end step." as a keyed tail, told the tokens the
// creation made. An on-disk key: never renamed, never reused.
var saheeliSunsBrillianceRest = CardTail("saheeli-the-suns-brilliance/haste-and-sacrifice-copy", func(ctx *Context, o game.TailOutcome) error {
	if len(o.IDs) == 0 {
		return nil
	}
	for _, id := range o.IDs {
		if err := (ScopedEffectFor{Target: id, Mods: []game.Mod{game.AddKeywordsMod("haste")},
			Duration: game.IndefiniteDuration(), Label: "Saheeli, the Sun's Brilliance — the token gains haste"}).Apply(ctx); err != nil {
			return err
		}
	}
	return ScheduleDelayedTrigger{
		At:         game.StepEnd,
		Controller: ctx.Item.Controller,
		Label:      "Saheeli, the Sun's Brilliance — sacrifice the token",
		Cards:      o.IDs,
		Body:       sacrificeListedCardsBody,
	}.Apply(ctx)
})
