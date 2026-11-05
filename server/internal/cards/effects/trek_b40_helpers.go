package effects

import (
	"slices"
	"strings"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// trek_b40_helpers.go — shared pieces of Star Trek Commander batch 40
// (#4974). Reused rather than rewritten: "whenever you draw your second
// card each turn" is b41DrewYourSecondCardThisTurn, "for each opponent you
// attacked this turn" is OpponentsYouAttackedThisTurn, the differently
// named lands P/T is trekB19DifferentlyNamedLandsPT, the Borg and Klingon
// Warrior tokens are already rows of tokens_table.go.

// trekB40YouAttacked is "whenever you attack": one trigger per attack
// declaration of the source's controller (OncePerBatch over the
// per-creature attack events), with the effect it runs.
func trekB40YouAttacked(label string, effect Effect) game.TriggeredAbility {
	return OncePerBatch(On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
		return attackDeclaredByYou(ev, source.Controller)
	}, label, effect))
}

// trekB40SelfOrAnotherOfSubtypeYouControlEntered is "whenever <this> or
// another <subtype> you control enters" (McCoy, Not a Bricklayer's
// Doctors): the source itself entered, or another permanent of that
// subtype entered under its controller's control, token or not. The
// subtype is a word of the source's text (CR 612.2) and is read through
// the layered view, so a changeling counts.
func trekB40SelfOrAnotherOfSubtypeYouControlEntered(subtype string) When {
	return func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
		if ev.Kind == game.EventETB && ev.CardID == source.InstanceID {
			return true
		}
		c, ok := enteredUnderYourControl(ev, source, g, true)
		return ok && c.HasSubtype(source.TextSubtype(subtype))
	}
}

// trekB40IllusionArtifactExcept is Reg Barclay's copy exception "except it
// isn't legendary, it's an Illusion artifact in addition to its other
// types, and it has haste" (CR 707.9b): all three become part of the
// token's copiable values, so a copy of the token keeps them.
func trekB40IllusionArtifactExcept(t *game.Card) {
	line := saheeliArtifactInAddition(removeLegendaryFromTypeLine(t.TypeLine))
	super, types, sub := game.ParseTypeLine(line)
	if !slices.Contains(sub, "Illusion") {
		sub = append(sub, "Illusion")
	}
	head := strings.Join(append(append([]string{}, super...), types...), " ")
	t.TypeLine = head + " — " + strings.Join(sub, " ")
	TokenCopyGainsHaste(t)
}

// trekB40CreateCopyExiledAtNextEndStep creates one token copy of `copy`
// under `controller`, with `except` applied to its copiable values, and
// exiles it at the beginning of the next end step (a delayed trigger on
// that token alone, Saheeli Rai's shape). A creature that can no longer
// be found is not copied.
func trekB40CreateCopyExiledAtNextEndStep(ctx *Context, controller, copy uuid.UUID, except func(*game.Card), label string) error {
	tmpl, ok := TokenCopyTemplate(ctx.Game, copy)
	if !ok {
		return nil
	}
	except(&tmpl)
	made, err := ctx.Game.CreateTokensForEffect(controller, copiedToken(tmpl), 1, game.TokenEntryOptions{})
	if err != nil || len(made) == 0 {
		return err
	}
	return ScheduleDelayedTrigger{
		At:         game.StepEnd,
		Controller: controller,
		Label:      label,
		Cards:      made,
		Body:       exileListedCardsBody,
	}.Apply(ctx)
}
