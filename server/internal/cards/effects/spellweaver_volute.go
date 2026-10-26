package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Spellweaver Volute — Enchantment — Aura {3}{U}{U}:
//
//	"Enchant instant card in a graveyard
//	 Whenever you cast a sorcery spell, copy the enchanted instant card.
//	 You may cast the copy without paying its mana cost. If you do, exile
//	 the enchanted card and attach this Aura to another instant card in a
//	 graveyard."
//
// The graveyard host is ADR 0154's (EnchantCardInGraveyard, #1780); the
// copy and its free cast are CopyAndCast (CR 707.12, #3258). #3157 adds
// the rest:
//
//   - "The enchanted instant card" is read as the trigger resolves
//     (CR 608.2h) through game.EnchantedCardObjectForEffect: the card the
//     Volute is attached to, or, once the Volute has left, the card it
//     was attached to as it last existed. A card exiled in response
//     makes the Volute enchant nothing, and the state-based action takes
//     it, so nothing is copied (first 2007-05-01 response ruling). A
//     Volute removed first still copies its card, even after that card
//     has left the graveyard too (the second): CopyAndCast.LastKnown.
//   - The copy exists in the graveyard (the caster's: the engine makes a
//     card copy in its holder's own graveyard) and is cast, or not,
//     while the trigger is still resolving (CR 608.2g, #2947); uncast, it
//     ceases to exist (CR 704.5e).
//   - "If you do" is the offer's FollowUp, run once the copy is cast
//     through it. It is told the card the copy was made from
//     (CastFollowUp.Object), since the copy spell does not name it.
//   - "Attach this Aura to another instant card in a graveyard" is a
//     choice, not a target, so hexproof means nothing and protection is
//     the attachment's own CR 701.3b test (game.
//     GraveyardCardsAuraCanEnchantForEffect). One candidate is a forced
//     choice; with none, the enchanted card is still exiled, the Volute
//     enchants nothing, and the state-based action puts it into its
//     owner's graveyard (the ruling). A Volute that has left the
//     battlefield, or no longer enchants that card, attaches to nothing.
//
// The new host is asked for BEFORE the exile, and the Aura is moved onto
// it before the old card is exiled. In one uninterrupted resolution
// nothing can tell the two orders apart, and this one keeps the Volute
// legally attached for as long as the question is open, so CR 704.5m
// does not bin it while its controller is choosing.
//
// No simplification beyond the grant model's (the cast is taken after
// the trigger resolves, ADR 0066).
var spellweaverVoluteIfYouDo = game.RegisterCastFollowUp("spellweaver-volute-exile-and-reattach", spellweaverVoluteExileAndReattach)

const spellweaverVoluteLabel = "Spellweaver Volute — copy the enchanted instant card; you may cast the copy without paying its mana cost"

func init() {
	Register(Spec{
		OracleID:     "3d92e5ce-9d30-4203-9f84-6c36117440cf",
		Name:         "Spellweaver Volute",
		Completeness: CompletenessFull,
		Targets:      EnchantCardInGraveyard("instant", Instant()),
		Triggered: []game.TriggeredAbility{
			On(game.EventCast, YouCast(Sorcery()), spellweaverVoluteLabel, spellweaverVoluteResolve),
		},
	})
}

// spellweaverVoluteResolve copies the enchanted card and offers its free
// cast, carrying the original to the "if you do".
func spellweaverVoluteResolve(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	aura, ok := ctx.SourceRef()
	if !ok {
		return nil
	}
	enchanted, ok := g.EnchantedCardObjectForEffect(aura)
	if !ok {
		return nil
	}
	return CopyAndCast{
		Cards:          []uuid.UUID{enchanted.ID},
		Free:           true,
		FollowUp:       spellweaverVoluteIfYouDo,
		FollowUpObject: enchanted,
		LastKnown:      true,
	}.Apply(ctx)
}

// spellweaverVoluteExileAndReattach is "exile the enchanted card and
// attach this Aura to another instant card in a graveyard". f.Source is
// the Volute, f.Object the card the copy was made from, f.Player the
// caster, who controlled the trigger and chooses.
func spellweaverVoluteExileAndReattach(g *game.Game, f game.CastFollowUp) error {
	volute, enchanted, chooser := f.Source, f.Object, f.Player
	var candidates []uuid.UUID
	if spellweaverVoluteStillEnchants(g, volute, enchanted) {
		candidates = g.GraveyardCardsAuraCanEnchantForEffect(volute, enchanted.ID)
	}
	if len(candidates) <= 1 {
		host := uuid.Nil
		if len(candidates) == 1 {
			host = candidates[0]
		}
		return spellweaverVoluteMove(g, volute, enchanted, host)
	}
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:  chooser,
		Source:   volute,
		Question: "Spellweaver Volute — attach it to another instant card in a graveyard",
		Cards:    candidates,
		Min:      1,
		Max:      1,
		Then: func(g *game.Game, picked []uuid.UUID) error {
			host := uuid.Nil
			if len(picked) == 1 {
				host = picked[0]
			}
			return spellweaverVoluteMove(g, volute, enchanted, host)
		},
	})
	return nil
}

// spellweaverVoluteMove attaches the Volute to `host` (none for
// uuid.Nil), while it still enchants the card, then exiles that card if
// it is still that object in a graveyard. Captures plain values only.
func spellweaverVoluteMove(g *game.Game, volute uuid.UUID, enchanted game.ObjectRef, host uuid.UUID) error {
	if host != uuid.Nil && spellweaverVoluteStillEnchants(g, volute, enchanted) {
		if err := g.AttachForEffect(volute, game.TargetRef{Kind: game.TargetCard, ID: host}); err != nil {
			return err
		}
	}
	if !g.CardInGraveyardForEffect(enchanted) {
		return nil
	}
	return g.ExileCardKeyedForEffect(enchanted.ID, game.KeyedTail{})
}

// spellweaverVoluteStillEnchants reports whether the Volute is on the
// battlefield attached to the card the copy was made from: "this Aura"
// is the object whose trigger resolved, and one that left has nothing to
// move.
func spellweaverVoluteStillEnchants(g *game.Game, volute uuid.UUID, enchanted game.ObjectRef) bool {
	if !onBattlefield(g, volute) {
		return false
	}
	ref, ok := g.PermanentRefForEffect(volute)
	if !ok {
		return false
	}
	now, ok := g.EnchantedCardObjectForEffect(ref)
	return ok && now == enchanted
}
