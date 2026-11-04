package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Brilliance Unleashed — Sorcery {4}{U}{R} (Teenage Mutant Ninja
// Turtles, #4457):
//
//	"Choose one or both —
//	 • Brilliance Unleashed deals 5 damage to target creature.
//	 • Choose target artifact card in your graveyard. Return it to the
//	   battlefield if it's an artifact creature card. Otherwise, return
//	   it to the battlefield and it's a 3/3 Robot artifact creature with
//	   flying."
//
// The second bullet is ADR 0210's entry overwrite (Supper for Spiders):
// a noncreature artifact card returns with its type line SET by the
// putting effect, so it is a 3/3 Robot artifact creature with flying as
// it lands, before anything sees it enter. The text says "it's" and not
// "in addition to its other types", so the overwrite sets the card types
// to artifact and creature (CR 205.1a): an artifact land comes back
// without the land type. Robot replaces only the creature types, so an
// artifact keeps its artifact types (an Equipment is still one, and as a
// creature it can't equip, CR 301.5c). Supertypes stay (CR 205.4b), and
// the card keeps its own abilities. Whether it is an artifact creature
// card is read in the graveyard, at resolution. It lasts as long as
// that object is on the battlefield (CR 400.7). The card returns under
// the caster's control: it is a card in their own graveyard.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "b047cd3f-9129-41cc-8b91-b0dc970f5068",
		Name:         "Brilliance Unleashed",
		Completeness: CompletenessFull,
		Modes: ChooseN("Choose one or both", 1, 2,
			ModeWithPurpose(ModeDoing("Brilliance Unleashed deals 5 damage to target creature.",
				TargetCreature("target creature"),
				DealFixedDamageToModesTarget(5)), ForTargets(DamageToTarget(0, 5))),
			ModeDoing("Choose target artifact card in your graveyard. Return it to the battlefield if it's an artifact creature card. Otherwise, return it to the battlefield and it's a 3/3 Robot artifact creature with flying.",
				TargetCardInGraveyard("target artifact card in your graveyard", Artifact(), YouOwn()),
				brillianceUnleashedReturn),
		),
	})
}

// brillianceUnleashedReturn is the second bullet.
func brillianceUnleashedReturn(item *game.StackItem, ctx *Context, occurrence int) error {
	t, ok := ModeTarget(ctx, occurrence)
	if !ok || t.Kind != game.TargetCard {
		return nil
	}
	c, ok := ctx.Game.LookupCardForEffect(t.ID)
	if !ok {
		return nil
	}
	opts := game.ZoneEntryOptions{Controller: item.Controller}
	if !(c.IsArtifact() && c.IsCreature()) {
		opts.EntersAs = &game.EntersAs{
			Types:    []string{"Artifact", "Creature"},
			Subtypes: ctx.TextSubtypes("Robot"),
			Keywords: []string{"flying"},
			BasePT:   &game.EntersAsPT{Power: 3, Toughness: 3},
			Label:    "Brilliance Unleashed — a 3/3 Robot artifact creature with flying",
		}
	}
	return putCardsFromZoneTogetherWith(ctx, game.ZoneGraveyard, []uuid.UUID{t.ID}, opts, game.TailRef{}, 0)
}
