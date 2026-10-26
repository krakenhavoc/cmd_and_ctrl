package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Last Voyage of the _____ — Enchantment {3}{B}:
//
//	"When this enchantment enters, you may put a name sticker on it, then
//	 it becomes an Aura with enchant creature. Return a creature card from
//	 your graveyard to the battlefield and attach this Aura to it.
//	 Enchanted creature gets +2/+0 for each name sticker on this Aura with
//	 seven or fewer letters.
//	 When this Aura leaves the battlefield, sacrifice enchanted creature."
//
// In printed order, as one Do (each step waits for the one before it):
//
//  1. The name sticker, the engine's keyed questions (ADR 0208 §4).
//  2. "It becomes an Aura with enchant creature": the permanent takes the
//     Aura form of a card that returns as one (game.BecomeAuraForEffect,
//     ADR 0154's 2026-10-23 amendment): Card.ReturnedAura, which the
//     layer-4 static below reads to add the Aura subtype, and which makes
//     Spec.ReturnedAura ("enchant creature") the clause CR 704.5m holds it
//     to. Held until it leaves the battlefield (CR 400.7). An Aura attached
//     to nothing is illegal, so with no creature card to return it goes to
//     the graveyard as a state-based action.
//  3. "Return a creature card from your graveyard to the battlefield": one
//     keyed choose_cards (not a target), mandatory when a creature card is
//     there. The creature enters through the entry pipeline, and the Aura
//     is attached to the permanent that entered (CR 303.4: an Aura attached
//     as it is put on is not cast, so it does not target).
//
// The pump counts the name stickers on the Aura itself whose words have
// seven or fewer letters (CR 123.6a), re-read every recompute.
//
// "When this Aura leaves the battlefield, sacrifice enchanted creature" is
// a leaves trigger judged on the permanent as it last existed (CR 603.10a):
// it is an Aura only if it had become one, and "enchanted creature" is the
// object it was attached to (game.PermanentInfo.AttachedObject), which is
// sacrificed only if it is still that object on the battlefield (CR 400.7).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "54ddf766-b528-488d-9b98-8a101bdb5de9",
		Name:         "Last Voyage of the _____",
		Completeness: CompletenessFull,
		ReturnedAura: EnchantCreature(),
		Static: []game.StaticAbility{
			lastVoyageAuraType(),
			PumpAttachedPer(2, 0, nameStickersWithAtMostLetters(7)),
		},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Last Voyage of the _____ — you may put a name sticker on it, then it becomes an Aura. Return a creature card from your graveyard to the battlefield and attach this Aura to it",
				Do(
					PutASticker{On: StickerOnThis, Kinds: []game.StickerKind{game.StickerName}},
					Instruction(lastVoyageBecomeAnAura),
					Instruction(lastVoyageChooseCreature),
				)),
			On(game.EventLTB, lastVoyageLeftAsAnAura,
				"Last Voyage of the _____ — sacrifice enchanted creature", lastVoyageSacrifice),
		},
	})
}

// lastVoyageAuraType is "it becomes an Aura": the Aura subtype (CR 205.3,
// layer 4) on the permanent that has taken the Aura form.
func lastVoyageAuraType() game.StaticAbility {
	return game.StaticAbility{
		Label: "It's an Aura.",
		Layer: game.Layer4Type,
		AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
			return source != nil && target.InstanceID == source.InstanceID && source.ReturnedAura
		},
		Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
			c.SetSubtypes(append(append([]string(nil), c.Subtypes...), "Aura"))
		},
	}
}

// nameStickersWithAtMostLetters counts the name stickers on the source
// whose word has at most n letters.
func nameStickersWithAtMostLetters(n int) func(g *game.Game, source *game.Card) int {
	return func(_ *game.Game, source *game.Card) int {
		return nameStickerCount(*source, func(text string) bool { return game.StickerLetters(text) <= n })
	}
}

func lastVoyageBecomeAnAura(ctx *Context) error {
	if ctx.Game.AbilitySourceGoneForEffect(ctx.Item) {
		return nil
	}
	ctx.Game.BecomeAuraForEffect(ctx.Source())
	return nil
}

// lastVoyageChooseCreature asks which creature card in the controller's
// graveyard returns. Nothing to ask when there is none.
func lastVoyageChooseCreature(ctx *Context) error {
	g := ctx.Game
	if g.AbilitySourceGoneForEffect(ctx.Item) {
		return nil
	}
	you := ctx.Controller()
	candidates := graveyardCardIDs(ctx, you, func(c game.Card) bool { return c.IsCreature() && !c.IsToken() })
	if len(candidates) == 0 {
		return nil
	}
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:    you,
		FromPlayer: you,
		Source:     ctx.Source(),
		Question:   "Last Voyage of the _____ — return a creature card from your graveyard to the battlefield",
		Cards:      candidates,
		Min:        1,
		Max:        1,
		ThenKey:    lastVoyageReturnThen,
		LiveZone:   game.ZoneGraveyard,
	})
	return nil
}

// lastVoyageReturnThen returns the chosen creature card and attaches the
// Aura to the permanent that entered. It runs from the answered prompt, so
// it is on a restored table too; it captures plain values only.
var lastVoyageReturnThen = game.RegisterChooseCardsThen("choose-cards/last-voyage-of-the-blank-return",
	func(g *game.Game, r game.CardsChosen) error {
		if !r.Answered || len(r.Picked) == 0 {
			return nil
		}
		aura := r.Source
		return g.ReturnFromGraveyardToBattlefieldThenForEffect(r.Picked[0], r.Chooser, false,
			func(g *game.Game, entered uuid.UUID) error {
				if entered == uuid.Nil || !onBattlefield(g, entered) || !onBattlefield(g, aura) {
					return nil
				}
				return g.AttachForEffect(aura, game.TargetRef{Kind: game.TargetCard, ID: entered})
			})
	})

// lastVoyageLeftAsAnAura is "When this Aura leaves the battlefield": the
// source left, and it was an Aura as it left (CR 603.10a).
func lastVoyageLeftAsAnAura(ev game.Event, source *game.Card, lki game.Characteristic, _ *game.Game) bool {
	return ev.CardID == source.InstanceID && keywordSliceContains(lki.Subtypes, "Aura")
}

// lastVoyageSacrifice sacrifices the creature the Aura was attached to,
// while it is still that object on the battlefield.
func lastVoyageSacrifice(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	aura, ok := ctx.TriggeringPermanent()
	if !ok || aura.AttachedObject == nil {
		return nil
	}
	host := *aura.AttachedObject
	c, ok := g.LookupCardForEffect(host.ID)
	if !ok || c.ObjectEpoch != host.Epoch || !onBattlefield(g, host.ID) {
		return nil
	}
	return SacrificePermanent{Target: host.ID}.Apply(ctx)
}
