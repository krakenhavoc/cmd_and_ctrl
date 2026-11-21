package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_gap_b64_helpers.go — GAP batch 64 (#5948): the shared pieces of the
// batch's cards.

// gapB64QuestCountersAmongYourPermanents is "the number of quest counters
// among permanents you control": read from inside a layer recompute, so it
// walks g.Battlefield.Cards directly rather than through the read-locked
// helpers.
func gapB64QuestCountersAmongYourPermanents(g *game.Game, source *game.Card) int {
	if g == nil || g.Battlefield == nil || source == nil {
		return 0
	}
	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.Controller == source.Controller && c.Counters["quest"] > 0 {
			n += c.Counters["quest"]
		}
	}
	return n
}

// printedGapB64SurvivalGuideToken is Moira Brown's Wasteland Survival Guide:
// a colorless Equipment artifact token with "Equipped creature gets +1/+1 for
// each quest counter among permanents you control" and equip {1}. Its slug is
// an on-disk identity: never renamed.
func printedGapB64SurvivalGuideToken() TokenTemplate {
	return TokenTemplate{
		Slug: "wasteland-survival-guide",
		Card: game.Card{
			Name:     "Wasteland Survival Guide",
			TypeLine: "Token Artifact — Equipment",
		},
		Text:      "Equipped creature gets +1/+1 for each quest counter among permanents you control.\nEquip {1}",
		Static:    []game.StaticAbility{PumpAttachedPer(1, 1, gapB64QuestCountersAmongYourPermanents)},
		Activated: activatedShapes([]ActivatedAbility{EquipAbility("{1}")}),
	}
}

// GapB64SurvivalGuideToken is that token's card value for CreateToken.
func GapB64SurvivalGuideToken() game.Card { return TokenFromCatalog(printedGapB64SurvivalGuideToken) }

// gapB64AttachmentsOnTheDead lists the Auras (aura true) or the Equipment
// (aura false) that were attached to the permanent `id` as it last existed on
// the battlefield (the departure record, CR 603.10a), as the objects they
// were: an attachment that is still the same object on the battlefield is read
// live, one that has left is read from its own record.
func GapB64AttachmentsOnTheDead(g *game.Game, id uuid.UUID, aura bool) []uuid.UUID {
	info, ok := g.LastKnownPermanentForEffect(id)
	if !ok {
		return nil
	}
	want := "Equipment"
	if aura {
		want = "Aura"
	}
	var out []uuid.UUID
	for _, ref := range info.Attachments {
		if c, live := g.LookupCardForEffect(ref.ID); live && c.ObjectEpoch == ref.Epoch && OnBattlefield(g, ref.ID) {
			if c.HasSubtype(want) {
				out = append(out, ref.ID)
			}
			continue
		}
		if rec, found := g.LastKnownPermanentForEffect(ref.ID); found && rec.Epoch == ref.Epoch &&
			KeywordSliceContains(rec.Characteristic.Subtypes, want) {
			out = append(out, ref.ID)
		}
	}
	return out
}
