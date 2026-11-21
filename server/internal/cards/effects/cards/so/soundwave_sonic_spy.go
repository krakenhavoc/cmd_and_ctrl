package cardsso

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Soundwave, Sonic Spy // Soundwave, Superior Captain (S82 GAP batch 185,
// #5948; ADR 0152) — a More Than Meets the Eye card.
//
// Front face, Legendary Artifact Creature — Robot {1}{W}{U}{B}, 5/4:
//
//	"More Than Meets the Eye {2}{W}{U}{B}
//	 Whenever one or more creature tokens you control deal combat damage
//	 to a player, exile target instant or sorcery card with mana value
//	 equal to the damage dealt from their graveyard. Copy it. You may
//	 cast the copy without paying its mana cost. If you do, convert
//	 Soundwave."
//
// Back face, Legendary Artifact:
//
//	"Whenever you cast a spell with an odd mana value, convert Soundwave.
//	 If you do, create Ravage, a legendary 3/3 black Robot artifact
//	 creature token with menace and deathtouch.
//	 Whenever you cast a spell with an even mana value, convert Soundwave.
//	 If you do, create Laserbeak, a legendary 2/2 blue Robot artifact
//	 creature token with flying and hexproof."
//
// The damage is the total the creature tokens dealt the player at once
// (CR 510.2, CR 603.2c; ThatMuchDamage, only token sources counted), fixed
// when the trigger is put on the stack, so the target's mana value is
// known when it is chosen (CR 603.3d) and re-checked on resolution
// (ManaValueEqualsTheEventAmount). "Their graveyard" is the graveyard of
// the player who controls the tokens: Soundwave's controller's. The card
// is exiled, copied and the copy offered during the resolution
// (CopyAndCast, CR 707.12). "If you do" is the offer's follow-up: it runs
// only when the copy is really cast, and converts Soundwave unless it has
// turned over since the trigger went on the stack (CR 701.28e; the count
// travels in the offer's FollowUpObject epoch). The back face reads the
// cast spell's mana value as it sits on the stack ({X} is the chosen
// value, CR 202.3e); a spell with mana value 0 is even. Each back-face
// trigger converts only if it still can (the first one to resolve flips
// Soundwave, the other finds it turned over: CR 701.28e), and the token is
// made only when the conversion happened.
//
// No simplification.
const SoundwaveOracleID = "cc18e782-a965-404f-96f1-348f3ed799fd"

func init() {
	effects.Register(effects.Spec{
		OracleID:         SoundwaveOracleID,
		Name:             "Soundwave, Sonic Spy",
		Completeness:     effects.CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{effects.MoreThanMeetsTheEye("{2}{W}{U}{B}")},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventDealDamage},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				if !effects.CombatDamageToPlayerBy(ev, source.Controller, g) {
					return false
				}
				c, ok := g.LookupCardForEffect(ev.Source)
				return ok && effects.IsToken(c)
			},
			Targets: effects.ManaValueEqualsTheEventAmount(effects.TargetCardInGraveyard(
				"target instant or sorcery card with mana value equal to the damage dealt from your graveyard",
				effects.YouOwn(), effects.Or(effects.Instant(), effects.Sorcery()))),
			Key:          "Soundwave, Sonic Spy — exile target instant or sorcery card with mana value equal to the damage dealt from your graveyard, copy it, you may cast the copy without paying its mana cost, and if you do convert Soundwave",
			AtBatchEnd:   effects.ThatMuchDamage(),
			OncePerBatch: true,
			BatchKey:     effects.PerPlayer,
			Effect:       soundwaveExileCopy,
		}},
	})
	effects.Register(effects.Spec{
		OracleID:     SoundwaveOracleID + "#1",
		Name:         "Soundwave, Superior Captain",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WheneverYouCast(soundwaveParity(1), "Soundwave, Superior Captain — convert Soundwave; if you do, create Ravage",
				soundwaveConvertThenToken(effects.TokenCard("3/3 black Legendary Robot artifact named Ravage with menace and deathtouch"))),
			effects.WheneverYouCast(soundwaveParity(0), "Soundwave, Superior Captain — convert Soundwave; if you do, create Laserbeak",
				soundwaveConvertThenToken(effects.TokenCard("2/2 blue Legendary Robot artifact named Laserbeak with flying and hexproof"))),
		},
	})
}

// soundwaveParity matches a spell whose mana value (as it sits on the
// stack) leaves `rem` when divided by two.
func soundwaveParity(rem int) effects.CardPredicate {
	return func(g *game.Game, _ uuid.UUID, c game.Card) bool {
		mv, ok := g.ManaValueForEffect(c)
		return ok && mv%2 == rem
	}
}

// soundwaveConvertThenToken is "convert Soundwave. If you do, create <token>".
func soundwaveConvertThenToken(token game.Card) effects.Effect {
	return func(g *game.Game, item *game.StackItem) error {
		ctx := effects.NewContext(g, item)
		converted, err := effects.ConvertThisReporting(ctx)
		if err != nil || !converted {
			return err
		}
		return effects.CreateToken{Template: token, N: 1}.Apply(ctx)
	}
}

// soundwaveExileCopy exiles the chosen card, then copies what reached exile.
func soundwaveExileCopy(g *game.Game, item *game.StackItem) error {
	ids := effects.LegalTargetIDs(effects.NewContext(g, item))
	if len(ids) == 0 {
		return nil
	}
	return g.ExileCardsKeyedForEffect(ids, effects.NewContext(g, item).ItemTail(soundwaveCopyTail))
}

var soundwaveCopyTail = effects.CardTail("soundwave/copy-exiled", func(ctx *effects.Context, o game.TailOutcome) error {
	g, item := ctx.Game, ctx.Item
	return effects.CopyAndCast{
		Cards:          o.IDs,
		Free:           true,
		FollowUp:       soundwaveConvertIfCast,
		FollowUpObject: game.ObjectRef{ID: item.SourceCardID, Epoch: item.SourceTurnOvers},
	}.Apply(effects.NewContext(g, item))
})

// soundwaveConvertIfCast is "if you do, convert Soundwave": the copy was cast.
var soundwaveConvertIfCast = game.RegisterCastFollowUp("soundwave/convert-if-cast", func(g *game.Game, f game.CastFollowUp) error {
	return effects.ConvertThis{}.Apply(effects.NewContext(g, &game.StackItem{
		Kind:            game.StackItemTriggered,
		Controller:      f.Player,
		Owner:           f.Player,
		SourceCardID:    f.Source,
		SourceTurnOvers: f.Object.Epoch,
	}))
})
