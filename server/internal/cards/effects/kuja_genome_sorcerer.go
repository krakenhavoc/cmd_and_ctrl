package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Kuja, Genome Sorcerer // Trance Kuja, Fate Defied — a transform
// double-faced Legendary Creature (oracle 3998a60d…), the commander of
// the Kuja 9's ultimate deck (#3213).
//
//	Kuja, Genome Sorcerer          {2}{B}{R}   3/4
//	Legendary Creature — Human Mutant Wizard
//	  "At the beginning of your end step, create a tapped 0/1 black
//	   Wizard creature token with "Whenever you cast a noncreature
//	   spell, this token deals 1 damage to each opponent." Then if you
//	   control four or more Wizards, transform Kuja."
//
//	Trance Kuja, Fate Defied                    4/6
//	Legendary Creature — Avatar Wizard
//	  "Flare Star — If a Wizard you control would deal damage to a
//	   permanent or player, it deals double that damage instead."
//
// The token is the shared NoncreatureCastWizardToken (ADR 0083), and
// it enters tapped through TokenEntryOptions. The Wizard count reads
// every Wizard the controller has on the battlefield as the ability
// resolves, tokens and Kuja himself included (changelings count too:
// Card.HasSubtype). The transform is TransformThis, so CR 701.27f
// applies: if Kuja has already turned over since the trigger went on
// the stack, the instruction is ignored.
//
// Flare Star is Gratuitous Violence's replacement narrowed to Wizard
// sources, judged by the source's last-known characteristics (CR
// 608.2h). It doubles every damage event from a Wizard you control,
// combat or not, and several doublers multiply (the affected player
// orders them).
//
// No simplification.
const kujaOracleID = "3998a60d-9521-4100-8d41-3e612972a3de"

func init() {
	Register(Spec{
		// Face 0 is the bare oracle ID; the back face registers under
		// the same ID plus "#1" (see game.CatalogKey).
		OracleID:     kujaOracleID,
		Name:         "Kuja, Genome Sorcerer",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			AtYourEndStep("Kuja — create a tapped Wizard token, then transform if you control four or more Wizards", kujaEndStep),
		},
	})
	Register(Spec{
		OracleID:     kujaOracleID + "#1",
		Name:         "Trance Kuja, Fate Defied",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{{
			Watches: []game.EventKind{game.EventDealDamage},
			AppliesTo: func(ev *game.ReplacementEvent, g *game.Game, src *game.Card) bool {
				if ev.Kind != game.RepEventDamage || ev.DamageAmount <= 0 || ev.DamageSource == uuid.Nil {
					return false
				}
				ch, ok := damageSourceCharacteristics(ev, g)
				// "Wizard" is a word of Trance Kuja's text (CR 612.2, #3984).
				return ok && (hasFold(ch.Subtypes, src.TextSubtype("Wizard")) || hasFold(ch.Abilities, game.KeywordChangeling)) && ch.Controller == src.Controller
			},
			Replace: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) error {
				ev.DamageAmount *= 2
				return nil
			},
			Controller: func(_ *game.ReplacementEvent, _ *game.Game, src *game.Card) uuid.UUID {
				return src.Controller
			},
			Label: "Trance Kuja: Flare Star doubles Wizard damage",
		}},
	})
}

// kujaEndStep is the front face's end-step ability.
func kujaEndStep(g *game.Game, item *game.StackItem) error {
	if _, err := g.CreateTokensForEffect(item.Controller, NoncreatureCastWizardToken(), 1, game.TokenEntryOptions{Tapped: true}); err != nil {
		return err
	}
	wizards := 0
	// "Wizards" is a word of the source's text (CR 612.2, #3984).
	wizard := g.TextReaderForEffect(item).TextSubtype("Wizard")
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == item.Controller && c.IsCreature() && c.HasSubtype(wizard) {
			wizards++
		}
	}
	if wizards < 4 {
		return nil
	}
	return TransformThis{}.Apply(NewContext(g, item))
}
