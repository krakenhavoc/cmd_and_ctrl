package effects

import (
	"strings"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const mournersShieldImprintLabel = "Mourner's Shield — exile target card from a graveyard"

// Mourner's Shield — Artifact {4}:
//
//	"Imprint — When this artifact enters, you may exile target card from a graveyard.
//	 {2}, {T}: Prevent all damage that would be dealt this turn by a source of your choice that shares a color with the exiled card."
//
// ADR 0108 §7 (#1904): the imprint is an ordinary "may" ETB trigger with a
// target, and "the exiled card" is the card that trigger exiled while this
// object has been on the battlefield, read back off the event log as
// Chrome Mox and Duplicant read theirs (b27ExiledWith; the two abilities
// are linked, CR 607.2a). The shield is against a source chosen as the
// ability resolves (CR 609.7a) that shares one of the exiled card's
// colours: the prompt offers only those, and the shield rechecks the
// colour each time the source would deal damage this turn (CR 615.9).
// With nothing exiled, or a colourless card exiled, no source shares a
// colour and nothing is prevented.
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:     "34cadd9d-b8e6-422c-b0ec-a2420cd983ad",
		Name:         "Mourner's Shield",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{{
			Watches:        []game.EventKind{game.EventETB},
			AppliesTo:      b06SelfETB,
			OptionalPrompt: &game.TriggerOptionalPrompt{Question: "Mourner's Shield — exile target card from a graveyard?"},
			Targets:        TargetCardInGraveyard("target card from a graveyard"),
			Key:            mournersShieldImprintLabel,
			Effect:         b27ExileChosenTarget,
		}},
		Activated: []ActivatedAbility{{
			Label: "{2}, {T}: Prevent all damage that would be dealt this turn by a source of your choice that shares a color with the exiled card.",
			Cost:  Plus(ManaCost("{2}"), TapCost()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				colors := mournersShieldColors(g, item)
				if len(colors) == 0 {
					return nil
				}
				return PreventDamageFromChosenSource(ShieldAnything, QueryColors(colors...)).Apply(NewContext(g, item))
			},
		}},
	})
}

// mournersShieldColors is the colours of the card this Mourner's Shield
// imprinted, as that card is in exile now.
func mournersShieldColors(g *game.Game, item *game.StackItem) []string {
	seen := map[string]bool{}
	var out []string
	for _, id := range b27ExiledWith(g, item.SourceCardID, mournersShieldImprintLabel) {
		c, ok := g.LookupCardForEffect(id)
		if !ok {
			continue
		}
		for _, col := range c.EffectiveColors() {
			if col = strings.ToUpper(col); !seen[col] {
				seen[col] = true
				out = append(out, col)
			}
		}
	}
	return out
}
