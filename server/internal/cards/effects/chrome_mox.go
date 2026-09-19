package effects

import (
	"strings"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Chrome Mox — Artifact {0}:
//
//	"Imprint — When this artifact enters, you may exile a nonartifact,
//	 nonland card from your hand.
//	 {T}: Add one mana of any of the exiled card's colors."
//
// A free artifact that costs a card. Both halves are real here, and
// the reason this file exists rather than a caveat is that the lazy
// version of this card — a Mox that taps for any colour because the
// imprint was skipped — is STRONGER than printed, which is the one
// direction a simplification may never go (#259).
//
// # The imprint is a TRIGGER, not an as-enters choice
//
// CR 614.12's "as ~ enters, choose …" happens off the stack and
// cannot be responded to. Chrome Mox prints "When this artifact
// enters", so it is an ordinary ETB trigger (Spec.Triggered watching
// EventETB, via WhenThisEnters): it uses the stack, opponents get a
// window, and a Mox answered in that window imprints nothing. The
// distinction is #578's, and it is observable — a Mox flickered in
// response to its own trigger never exiles the card.
//
// # The pick
//
// "You may exile a nonartifact, nonland card from your hand" is the
// pick-from-hand prompt with a continuation (game.ChooseCardsPrompt
// with Zone: ZoneHand, #552). The "may" is the prompt's floor of
// zero rather than a separate yes/no: one click either way, and
// declining is a real answer the bot enumerator already knows how to
// give. The candidate list is built here rather than with
// handCardsMatching, which skips anything that could not be a
// permanent — and an instant or a sorcery is exactly what a Chrome
// Mox usually eats.
//
// # "The exiled card's colors"
//
// Nothing on a Card records what was imprinted with it, so the link
// is read back off the event log, the way Duplicant's is: the card
// whose most recent move into exile happened while THIS Mox's
// imprint trigger was resolving (b27ExiledWith). The prompt blocks
// priority while it is open, so no resolution, cast, attack or step
// boundary can slip between the trigger's EventResolve and the
// exile.
//
// The colours are the exiled card's, through EffectiveColors, and a
// multicoloured imprint offers a pipe pick the same way Birds of
// Paradise does. Two printed outcomes fall straight out and neither
// is an error: a Mox that imprinted nothing, and one that imprinted
// a colourless card (Kozilek is a legal imprint), both produce ""
// and tap for nothing at all.
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:     "ec3d4466-547c-4e02-b1b5-a156ec4637e9",
		Name:         "Chrome Mox",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters(chromeMoxImprintLabel, chromeMoxImprint),
		},
		ManaAbilities: []ManaAbility{{
			Cost:         ManaAbilityCost{Tap: true},
			ProducedFunc: chromeMoxImprintedColors,
			Label:        "Add one mana of any of the exiled card's colors",
		}},
	})
}

// chromeMoxImprintLabel is the stack label of the imprint trigger.
// The "exiled with this artifact" record keys on it, so it must not
// drift from the label the item is built with.
const chromeMoxImprintLabel = "Chrome Mox — exile a nonartifact, nonland card from your hand"

// chromeMoxImprint queues the pick. It QUEUES and does not finish:
// the exile happens when the controller answers.
func chromeMoxImprint(g *game.Game, item *game.StackItem) error {
	controller, source := item.Controller, item.SourceCardID
	candidates := chromeMoxImprintCandidates(g, controller)
	if len(candidates) == 0 {
		return nil
	}
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:  controller,
		Source:   source,
		Question: chromeMoxImprintLabel,
		Cards:    candidates,
		Min:      0,
		Max:      1,
		Zone:     game.ZoneHand,
		Then:     chromeMoxExileThePick,
	})
	return nil
}

// chromeMoxExileThePick exiles the chosen card, or nothing when the
// controller declined.
func chromeMoxExileThePick(g *game.Game, picked []uuid.UUID) error {
	if len(picked) == 0 {
		return nil
	}
	return g.ExileCardForEffect(picked[0])
}

// chromeMoxImprintCandidates is every nonartifact, nonland card in
// the controller's hand, in hand order. Tokens cannot be in a hand,
// so there is nothing else to exclude.
func chromeMoxImprintCandidates(g *game.Game, player uuid.UUID) []uuid.UUID {
	p := g.PlayerByIDForEffect(player)
	if p == nil || p.Hand == nil {
		return nil
	}
	var out []uuid.UUID
	for _, c := range p.Hand.Cards {
		if c.IsArtifact() || c.IsLand() {
			continue
		}
		out = append(out, c.InstanceID)
	}
	return out
}

// chromeMoxImprintedColors is the mana ability's derived output:
// one slot offering every colour of the card imprinted on THIS Mox.
// Read-only, under g.mu, as ManaAbility.ProducedFunc requires.
func chromeMoxImprintedColors(g *game.Game, _, source uuid.UUID) string {
	seen := map[string]bool{}
	for _, id := range b27ExiledWith(g, source, chromeMoxImprintLabel) {
		c, ok := g.LookupCardForEffect(id)
		if !ok {
			continue
		}
		for _, col := range c.EffectiveColors() {
			seen[strings.ToUpper(col)] = true
		}
	}
	return pipeString(colorsOnly(orderedManaSymbols(seen)))
}
