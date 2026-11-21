package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Shadow the Hedgehog — Legendary Creature — Hedgehog Mercenary
// {B}{B}{R}{R}, 4/2 (Smaug Unchained deck, #4724):
//
//	"Haste
//	 Whenever Shadow the Hedgehog or another creature you control with
//	 flash or haste dies, draw a card.
//	 Chaos Control — Each spell you cast has split second if mana from
//	 an artifact was spent to cast it. (As long as it's on the stack,
//	 players can't cast spells or activate abilities that aren't mana
//	 abilities.)"
//
// Haste rides PrintedKeywords.
//
// The dies trigger is two conditions in one ability. Shadow's own death
// counts whatever he had when he died, since he is named. Another
// creature counts when you controlled it as it died and it had flash or
// haste then: both are read as it last existed on the battlefield
// (CR 603.10a; leftUnderControlOf and leftWithKeyword, The Serpent
// Society's reading), so a creature that had haste only from an effect
// counts, and one that lost it does not. Leaves-the-battlefield
// abilities look back in time, so Shadow dying beside other creatures
// (a board wipe) draws for each of them too.
//
// Chaos Control is a split-second grant (Spec.SplitSecondGrants) built by
// EachSpellYouCastHasSplitSecondIfManaFrom: it reaches every spell
// Shadow's controller casts whose payment included mana from an
// artifact, read off the spell's payment record, where each mana keeps a
// snapshot of what its source was (#1212), so a Treasure that sacrificed
// itself to pay counts. The grant is read live (CR 611.3a), so Shadow
// leaving the battlefield takes split second off a spell already on the
// stack. A copy was not cast and gets nothing. Mana from an artifact the
// engine does not run can be paid on paper, so the card declares
// ReadsYourCasts with ReadsSourceKinds: a waived cast asks the caster
// which of the mana an artifact made, as Coin of Mastery's does.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "0c12abcc-6ada-43b3-8750-9cb6ce2b0785",
		Name:            "Shadow the Hedgehog",
		ReadsManaSpent:  game.ReadsYourCasts | game.ReadsSourceKinds,
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"haste"},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventLTB, shadowOrAFlashOrHasteCreatureYouControlDied,
				"Shadow the Hedgehog — draw a card",
				effects.Do(effects.DrawCards{N: 1})),
		},
		SplitSecondGrants: []game.SplitSecondGrant{
			effects.EachSpellYouCastHasSplitSecondIfManaFrom(game.ManaSourceArtifact,
				"Chaos Control — Each spell you cast has split second if mana from an artifact was spent to cast it."),
		},
	})
}

// shadowOrAFlashOrHasteCreatureYouControlDied is "Whenever Shadow the
// Hedgehog or another creature you control with flash or haste dies".
func shadowOrAFlashOrHasteCreatureYouControlDied(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
	if effects.ThisDied(ev, source, lki, g) {
		return true
	}
	if ev.CardID == source.InstanceID {
		return false
	}
	dead, ok := effects.DiedCreature(ev, g)
	if !ok || effects.LeftUnderControlOf(ev, dead) != source.Controller {
		return false
	}
	return effects.LeftWithKeyword(ev, dead, "flash") || effects.LeftWithKeyword(ev, dead, "haste")
}
